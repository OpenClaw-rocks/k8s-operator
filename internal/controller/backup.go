/*
Copyright 2026 OpenClaw.rocks

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	openclawv1alpha1 "github.com/openclawrocks/k8s-operator/api/v1alpha1"
	"github.com/openclawrocks/k8s-operator/internal/resources"
)

// reconcileDeleteWithBackup implements the backup-before-delete state machine:
//  1. Check skip-backup annotation → if set, remove finalizer immediately
//  2. Scale down StatefulSet to 0 → requeue 5s
//  3. Wait for pods to terminate → requeue 5s
//  4. Create/check backup Job
//  5. On success: remove finalizer → K8s GCs all owned resources
func (r *OpenClawInstanceReconciler) reconcileDeleteWithBackup(ctx context.Context, instance *openclawv1alpha1.OpenClawInstance) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling deletion with backup", "instance", instance.Name, "namespace", instance.Namespace)

	// Step 0: Update phase to BackingUp (if not already terminating/backing up)
	if instance.Status.Phase != openclawv1alpha1.PhaseBackingUp && instance.Status.Phase != openclawv1alpha1.PhaseTerminating {
		instance.Status.Phase = openclawv1alpha1.PhaseBackingUp
		if err := r.Status().Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Step 1: Check skip-backup annotation
	if instance.Annotations[AnnotationSkipBackup] == "true" {
		logger.Info("Skip-backup annotation set, removing finalizer immediately")
		instance.Status.Phase = openclawv1alpha1.PhaseTerminating
		if err := r.Status().Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
		return r.removeFinalizer(ctx, instance)
	}

	// Check if persistence is enabled — no PVC means nothing to back up
	persistenceEnabled := instance.Spec.Storage.Persistence.Enabled == nil || *instance.Spec.Storage.Persistence.Enabled
	if !persistenceEnabled {
		logger.Info("Persistence disabled, skipping backup")
		instance.Status.Phase = openclawv1alpha1.PhaseTerminating
		if err := r.Status().Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
		return r.removeFinalizer(ctx, instance)
	}

	// Step 2: Scale down StatefulSet to 0
	sts := &appsv1.StatefulSet{}
	stsKey := client.ObjectKey{Name: resources.StatefulSetName(instance), Namespace: instance.Namespace}
	if err := r.Get(ctx, stsKey, sts); err != nil {
		if apierrors.IsNotFound(err) {
			// StatefulSet doesn't exist — no pods to worry about, proceed to backup
			logger.Info("StatefulSet not found, proceeding to backup")
		} else {
			return ctrl.Result{}, err
		}
	} else {
		// Scale to 0 if needed
		if sts.Spec.Replicas == nil || *sts.Spec.Replicas > 0 {
			logger.Info("Scaling down StatefulSet to 0")
			zero := int32(0)
			sts.Spec.Replicas = &zero
			if err := r.Update(ctx, sts); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}

		// Step 3: Wait for pods to terminate
		podList := &corev1.PodList{}
		if err := r.List(ctx, podList,
			client.InNamespace(instance.Namespace),
			client.MatchingLabels(resources.SelectorLabels(instance)),
		); err != nil {
			return ctrl.Result{}, err
		}
		if len(podList.Items) > 0 {
			logger.Info("Waiting for pods to terminate", "count", len(podList.Items))
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
	}

	// Step 4: Create/check backup Job
	creds, err := r.getB2Credentials(ctx)
	if err != nil {
		logger.Error(err, "Failed to get B2 credentials, cannot backup")
		r.Recorder.Event(instance, corev1.EventTypeWarning, "BackupCredentialsFailed", err.Error())
		// Keep finalizer — manual intervention needed
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	tenantID := getTenantID(instance)
	timestamp := time.Now().UTC().Format("2006-01-02T150405Z")
	b2Path := fmt.Sprintf("backups/%s/%s/%s", tenantID, instance.Name, timestamp)
	jobName := backupJobName(instance)

	// Check for existing Job
	existingJob, err := r.getJob(ctx, jobName, instance.Namespace)
	if err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if apierrors.IsNotFound(err) || existingJob == nil {
		// Create backup Job
		// Use the path from status if already set (idempotent on requeue)
		if instance.Status.LastBackupPath != "" {
			b2Path = instance.Status.LastBackupPath
		} else {
			instance.Status.LastBackupPath = b2Path
			instance.Status.BackupJobName = jobName
			if err := r.Status().Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}

		pvcName := pvcNameForInstance(instance)
		labels := backupLabels(instance, "backup")
		job := buildRcloneJob(jobName, instance.Namespace, pvcName, b2Path, labels, creds, true)

		// Set owner reference so the Job is cleaned up with the instance
		if err := controllerutil.SetControllerReference(instance, job, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		logger.Info("Creating backup Job", "job", jobName, "b2Path", b2Path)
		if err := r.Create(ctx, job); err != nil {
			if apierrors.IsAlreadyExists(err) {
				// Race condition — Job was created between our check and create
				return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
			}
			return ctrl.Result{}, err
		}
		r.Recorder.Event(instance, corev1.EventTypeNormal, "BackupStarted", fmt.Sprintf("Backup Job %s created, target: %s", jobName, b2Path))
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Job exists — check its status
	finished, condType := isJobFinished(existingJob)
	if !finished {
		logger.Info("Backup Job still running", "job", jobName)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	if condType == batchv1.JobFailed {
		logger.Error(nil, "Backup Job failed — keeping finalizer for manual intervention", "job", jobName)
		r.Recorder.Event(instance, corev1.EventTypeWarning, "BackupFailed",
			fmt.Sprintf("Backup Job %s failed. Fix and delete the Job to retry, or annotate %s=true to skip.", jobName, AnnotationSkipBackup))

		meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
			Type:               openclawv1alpha1.ConditionTypeBackupComplete,
			Status:             metav1.ConditionFalse,
			Reason:             "BackupFailed",
			Message:            "Backup Job failed",
			LastTransitionTime: metav1.Now(),
		})
		if err := r.Status().Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Step 5: Backup succeeded — record and remove finalizer
	logger.Info("Backup Job completed successfully", "job", jobName, "b2Path", instance.Status.LastBackupPath)
	r.Recorder.Event(instance, corev1.EventTypeNormal, "BackupComplete",
		fmt.Sprintf("Backup completed to %s", instance.Status.LastBackupPath))

	now := metav1.Now()
	instance.Status.LastBackupTime = &now
	instance.Status.Phase = openclawv1alpha1.PhaseTerminating

	meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
		Type:               openclawv1alpha1.ConditionTypeBackupComplete,
		Status:             metav1.ConditionTrue,
		Reason:             "BackupSucceeded",
		Message:            fmt.Sprintf("Backup completed to %s", instance.Status.LastBackupPath),
		LastTransitionTime: metav1.Now(),
	})
	if err := r.Status().Update(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}

	return r.removeFinalizer(ctx, instance)
}

// removeFinalizer removes the operator finalizer, allowing K8s to GC the resource
func (r *OpenClawInstanceReconciler) removeFinalizer(ctx context.Context, instance *openclawv1alpha1.OpenClawInstance) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	controllerutil.RemoveFinalizer(instance, FinalizerName)
	if err := r.Update(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}
	logger.Info("Finalizer removed, cleanup complete")
	return ctrl.Result{}, nil
}
