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
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openclawv1alpha1 "github.com/openclawrocks/k8s-operator/api/v1alpha1"
	"github.com/openclawrocks/k8s-operator/internal/resources"
)

const (
	// BackupSecretName is the name of the Secret containing B2 credentials
	BackupSecretName = "b2-backup-credentials"

	// RcloneImage is the pinned rclone container image
	RcloneImage = "rclone/rclone:1.68"

	// AnnotationSkipBackup allows skipping backup on delete
	AnnotationSkipBackup = "openclaw.rocks/skip-backup"

	// LabelTenant is the label key for the tenant ID
	LabelTenant = "openclaw.rocks/tenant"

	// LabelInstance is the label key for the instance ID
	LabelInstance = "openclaw.rocks/instance"

	// LabelManagedBy is the label key for the manager
	LabelManagedBy = "app.kubernetes.io/managed-by"
)

// b2Credentials holds the B2 credential values read from a Secret
type b2Credentials struct {
	Bucket   string
	KeyID    string
	AppKey   string
	Endpoint string
}

// getTenantID extracts the tenant ID from the instance label or falls back to namespace
func getTenantID(instance *openclawv1alpha1.OpenClawInstance) string {
	if tenant, ok := instance.Labels[LabelTenant]; ok && tenant != "" {
		return tenant
	}
	// Fallback: extract from namespace (oc-tenant-{id} → {id})
	ns := instance.Namespace
	if strings.HasPrefix(ns, "oc-tenant-") {
		return strings.TrimPrefix(ns, "oc-tenant-")
	}
	return ns
}

// getB2Credentials reads the B2 backup credentials Secret from the operator namespace
func (r *OpenClawInstanceReconciler) getB2Credentials(ctx context.Context) (*b2Credentials, error) {
	secret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      BackupSecretName,
		Namespace: r.OperatorNamespace,
	}, secret); err != nil {
		return nil, fmt.Errorf("failed to get B2 credentials secret %s/%s: %w", r.OperatorNamespace, BackupSecretName, err)
	}

	get := func(key string) (string, error) {
		v, ok := secret.Data[key]
		if !ok || len(v) == 0 {
			return "", fmt.Errorf("B2 credentials secret missing key %q", key)
		}
		return string(v), nil
	}

	bucket, err := get("B2_BUCKET")
	if err != nil {
		return nil, err
	}
	keyID, err := get("B2_KEY_ID")
	if err != nil {
		return nil, err
	}
	appKey, err := get("B2_APP_KEY")
	if err != nil {
		return nil, err
	}
	endpoint, err := get("B2_ENDPOINT")
	if err != nil {
		return nil, err
	}

	return &b2Credentials{
		Bucket:   bucket,
		KeyID:    keyID,
		AppKey:   appKey,
		Endpoint: endpoint,
	}, nil
}

// buildRcloneJob creates a batch/v1 Job that runs rclone to sync data between a PVC and B2.
// For backup: src=PVC mount, dst=B2 remote path, readOnly=true
// For restore: src=B2 remote path, dst=PVC mount, readOnly=false
func buildRcloneJob(
	name, namespace, pvcName string,
	b2Path string,
	labels map[string]string,
	creds *b2Credentials,
	isBackup bool,
) *batchv1.Job {
	backoffLimit := int32(3)
	ttl := int32(86400) // 24h

	// rclone remote config via env vars
	// :s3: is used because B2 S3-compatible API works with rclone's S3 backend
	remotePath := fmt.Sprintf(":s3:%s/%s", creds.Bucket, b2Path)

	var args []string
	if isBackup {
		// PVC → B2
		args = []string{"sync", "/data/", remotePath, "--s3-provider=Other", "--s3-endpoint=$(B2_ENDPOINT)", "--s3-access-key-id=$(B2_KEY_ID)", "--s3-secret-access-key=$(B2_APP_KEY)", "--transfers=8", "--checkers=16", "-v"}
	} else {
		// B2 → PVC
		args = []string{"sync", remotePath, "/data/", "--s3-provider=Other", "--s3-endpoint=$(B2_ENDPOINT)", "--s3-access-key-id=$(B2_KEY_ID)", "--s3-secret-access-key=$(B2_APP_KEY)", "--transfers=8", "--checkers=16", "-v"}
	}

	// Volume mount: read-only for backup, read-write for restore
	readOnly := isBackup

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{
						{
							Name:    "rclone",
							Image:   RcloneImage,
							Command: []string{"rclone"},
							Args:    args,
							Env: []corev1.EnvVar{
								{Name: "B2_ENDPOINT", Value: creds.Endpoint},
								{Name: "B2_KEY_ID", Value: creds.KeyID},
								{Name: "B2_APP_KEY", Value: creds.AppKey},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/data",
									ReadOnly:  readOnly,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
									ReadOnly:  readOnly,
								},
							},
						},
					},
				},
			},
		},
	}
}

// backupJobName returns a deterministic name for the backup Job
func backupJobName(instance *openclawv1alpha1.OpenClawInstance) string {
	return instance.Name + "-backup"
}

// restoreJobName returns a deterministic name for the restore Job
func restoreJobName(instance *openclawv1alpha1.OpenClawInstance) string {
	return instance.Name + "-restore"
}

// backupLabels returns labels for a backup/restore Job
func backupLabels(instance *openclawv1alpha1.OpenClawInstance, jobType string) map[string]string {
	return map[string]string{
		LabelManagedBy:            "openclaw-operator",
		LabelTenant:               getTenantID(instance),
		LabelInstance:             instance.Name,
		"openclaw.rocks/job-type": jobType,
	}
}

// isJobFinished checks whether the given Job has completed or failed
func isJobFinished(job *batchv1.Job) (bool, batchv1.JobConditionType) {
	for _, c := range job.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			return true, c.Type
		}
	}
	return false, ""
}

// pvcName returns the PVC name for the instance (delegates to resources package)
func pvcNameForInstance(instance *openclawv1alpha1.OpenClawInstance) string {
	return resources.PVCName(instance)
}

// getJob fetches a Job by name and namespace, returns nil if not found
func (r *OpenClawInstanceReconciler) getJob(ctx context.Context, name, namespace string) (*batchv1.Job, error) {
	job := &batchv1.Job{}
	err := r.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, job)
	if err != nil {
		return nil, err
	}
	return job, nil
}
