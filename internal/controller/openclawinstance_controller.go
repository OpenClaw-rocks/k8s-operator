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
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	openclawv1alpha1 "github.com/openclawrocks/k8s-operator/api/v1alpha1"
	"github.com/openclawrocks/k8s-operator/internal/resources"
)

const (
	// FinalizerName is the finalizer used by this controller
	FinalizerName = "openclaw.openclaw.rocks/finalizer"

	// RequeueAfter is the default requeue interval
	RequeueAfter = 5 * time.Minute
)

// OpenClawInstanceReconciler reconciles a OpenClawInstance object
type OpenClawInstanceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=openclaw.openclaw.rocks,resources=openclawinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openclaw.openclaw.rocks,resources=openclawinstances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=openclaw.openclaw.rocks,resources=openclawinstances/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop
func (r *OpenClawInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reconcileStart := time.Now()
	defer func() {
		reconcileDuration.WithLabelValues(req.Name, req.Namespace).Observe(time.Since(reconcileStart).Seconds())
	}()

	logger := log.FromContext(ctx)
	logger.Info("Reconciling OpenClawInstance")

	// Fetch the OpenClawInstance
	instance := &openclawv1alpha1.OpenClawInstance{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("OpenClawInstance not found, likely deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get OpenClawInstance")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !instance.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, instance)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(instance, FinalizerName) {
		logger.Info("Adding finalizer")
		controllerutil.AddFinalizer(instance, FinalizerName)
		if err := r.Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Set initial phase if not set
	if instance.Status.Phase == "" {
		instance.Status.Phase = openclawv1alpha1.PhasePending
		if err := r.Status().Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Update phase to Provisioning
	if instance.Status.Phase == openclawv1alpha1.PhasePending {
		instance.Status.Phase = openclawv1alpha1.PhaseProvisioning
		if err := r.Status().Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Reconcile all resources
	if err := r.reconcileResources(ctx, instance); err != nil {
		logger.Error(err, "Failed to reconcile resources")
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReconcileFailed", err.Error())

		// Update status to Failed
		meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
			Type:               openclawv1alpha1.ConditionTypeReady,
			Status:             metav1.ConditionFalse,
			Reason:             "ReconcileFailed",
			Message:            err.Error(),
			LastTransitionTime: metav1.Now(),
		})
		instance.Status.Phase = openclawv1alpha1.PhaseFailed
		reconcileTotal.WithLabelValues(instance.Name, instance.Namespace, "error").Inc()
		updatePhaseMetric(instance.Name, instance.Namespace, instance.Status.Phase)
		if statusErr := r.Status().Update(ctx, instance); statusErr != nil {
			logger.Error(statusErr, "Failed to update status")
		}

		// Use shorter requeue for transient errors, longer for persistent ones
		requeueAfter := 30 * time.Second
		if instance.Status.Phase == openclawv1alpha1.PhaseFailed {
			// If already in failed state, back off more
			requeueAfter = 2 * time.Minute
		}
		return ctrl.Result{RequeueAfter: requeueAfter}, err
	}

	// Update status to Running
	instance.Status.Phase = openclawv1alpha1.PhaseRunning
	instance.Status.ObservedGeneration = instance.Generation
	instance.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}

	meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
		Type:               openclawv1alpha1.ConditionTypeReady,
		Status:             metav1.ConditionTrue,
		Reason:             "ReconcileSucceeded",
		Message:            "All resources reconciled successfully",
		LastTransitionTime: metav1.Now(),
	})

	if err := r.Status().Update(ctx, instance); err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	reconcileTotal.WithLabelValues(instance.Name, instance.Namespace, "success").Inc()
	updatePhaseMetric(instance.Name, instance.Namespace, instance.Status.Phase)

	r.Recorder.Event(instance, corev1.EventTypeNormal, "ReconcileSucceeded", "All resources reconciled successfully")
	logger.Info("Reconciliation completed successfully")

	return ctrl.Result{RequeueAfter: RequeueAfter}, nil
}

func updatePhaseMetric(name, namespace, currentPhase string) {
	phases := []string{"Pending", "Provisioning", "Running", "Degraded", "Failed", "Terminating"}
	for _, phase := range phases {
		val := float64(0)
		if phase == currentPhase {
			val = 1
		}
		instancePhase.WithLabelValues(name, namespace, phase).Set(val)
	}
}

// reconcileResources reconciles all managed resources
func (r *OpenClawInstanceReconciler) reconcileResources(ctx context.Context, instance *openclawv1alpha1.OpenClawInstance) error {
	logger := log.FromContext(ctx)

	// 1. Reconcile RBAC (ServiceAccount, Role, RoleBinding)
	if err := r.reconcileRBAC(ctx, instance); err != nil {
		return fmt.Errorf("failed to reconcile RBAC: %w", err)
	}
	logger.V(1).Info("RBAC reconciled")

	// 2. Reconcile NetworkPolicy
	if err := r.reconcileNetworkPolicy(ctx, instance); err != nil {
		return fmt.Errorf("failed to reconcile NetworkPolicy: %w", err)
	}
	logger.V(1).Info("NetworkPolicy reconciled")

	// 3. Reconcile ConfigMap (if using raw config)
	if err := r.reconcileConfigMap(ctx, instance); err != nil {
		return fmt.Errorf("failed to reconcile ConfigMap: %w", err)
	}
	logger.V(1).Info("ConfigMap reconciled")

	// 4. Reconcile PVC
	if err := r.reconcilePVC(ctx, instance); err != nil {
		return fmt.Errorf("failed to reconcile PVC: %w", err)
	}
	logger.V(1).Info("PVC reconciled")

	// 5. Reconcile PodDisruptionBudget
	if err := r.reconcilePDB(ctx, instance); err != nil {
		return fmt.Errorf("failed to reconcile PodDisruptionBudget: %w", err)
	}
	logger.V(1).Info("PodDisruptionBudget reconciled")

	// 6. Reconcile Deployment
	if err := r.reconcileDeployment(ctx, instance); err != nil {
		return fmt.Errorf("failed to reconcile Deployment: %w", err)
	}
	logger.V(1).Info("Deployment reconciled")

	// 7. Reconcile Service
	if err := r.reconcileService(ctx, instance); err != nil {
		return fmt.Errorf("failed to reconcile Service: %w", err)
	}
	logger.V(1).Info("Service reconciled")

	// 8. Reconcile Ingress (if enabled)
	if err := r.reconcileIngress(ctx, instance); err != nil {
		return fmt.Errorf("failed to reconcile Ingress: %w", err)
	}
	logger.V(1).Info("Ingress reconciled")

	// 9. Reconcile ServiceMonitor (if enabled)
	if err := r.reconcileServiceMonitor(ctx, instance); err != nil {
		return fmt.Errorf("failed to reconcile ServiceMonitor: %w", err)
	}
	logger.V(1).Info("ServiceMonitor reconciled")

	return nil
}

// reconcileRBAC reconciles ServiceAccount, Role, and RoleBinding
func (r *OpenClawInstanceReconciler) reconcileRBAC(ctx context.Context, instance *openclawv1alpha1.OpenClawInstance) error {
	// Check if we should create a ServiceAccount
	createSA := instance.Spec.Security.RBAC.CreateServiceAccount == nil || *instance.Spec.Security.RBAC.CreateServiceAccount

	if createSA {
		// Reconcile ServiceAccount
		sa := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      resources.ServiceAccountName(instance),
				Namespace: instance.Namespace,
			},
		}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, sa, func() error {
			desired := resources.BuildServiceAccount(instance)
			sa.Labels = desired.Labels
			sa.Annotations = desired.Annotations
			return controllerutil.SetControllerReference(instance, sa, r.Scheme)
		}); err != nil {
			return err
		}
		instance.Status.ManagedResources.ServiceAccount = sa.Name

		// Reconcile Role
		role := &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      resources.RoleName(instance),
				Namespace: instance.Namespace,
			},
		}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, role, func() error {
			desired := resources.BuildRole(instance)
			role.Labels = desired.Labels
			role.Rules = desired.Rules
			return controllerutil.SetControllerReference(instance, role, r.Scheme)
		}); err != nil {
			return err
		}
		instance.Status.ManagedResources.Role = role.Name

		// Reconcile RoleBinding
		roleBinding := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      resources.RoleBindingName(instance),
				Namespace: instance.Namespace,
			},
		}
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, roleBinding, func() error {
			desired := resources.BuildRoleBinding(instance)
			roleBinding.Labels = desired.Labels
			roleBinding.RoleRef = desired.RoleRef
			roleBinding.Subjects = desired.Subjects
			return controllerutil.SetControllerReference(instance, roleBinding, r.Scheme)
		}); err != nil {
			return err
		}
		instance.Status.ManagedResources.RoleBinding = roleBinding.Name
	}

	meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
		Type:               openclawv1alpha1.ConditionTypeRBACReady,
		Status:             metav1.ConditionTrue,
		Reason:             "RBACCreated",
		Message:            "RBAC resources created successfully",
		LastTransitionTime: metav1.Now(),
	})

	return nil
}

// reconcileNetworkPolicy reconciles the NetworkPolicy
func (r *OpenClawInstanceReconciler) reconcileNetworkPolicy(ctx context.Context, instance *openclawv1alpha1.OpenClawInstance) error {
	// Check if NetworkPolicy is enabled
	enabled := instance.Spec.Security.NetworkPolicy.Enabled == nil || *instance.Spec.Security.NetworkPolicy.Enabled

	if !enabled {
		// Delete existing NetworkPolicy if it exists
		np := &networkingv1.NetworkPolicy{}
		np.Name = resources.NetworkPolicyName(instance)
		np.Namespace = instance.Namespace
		if err := r.Delete(ctx, np); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		instance.Status.ManagedResources.NetworkPolicy = ""
		return nil
	}

	np := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resources.NetworkPolicyName(instance),
			Namespace: instance.Namespace,
		},
	}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, np, func() error {
		desired := resources.BuildNetworkPolicy(instance)
		np.Labels = desired.Labels
		np.Spec = desired.Spec
		return controllerutil.SetControllerReference(instance, np, r.Scheme)
	}); err != nil {
		return err
	}
	instance.Status.ManagedResources.NetworkPolicy = np.Name

	meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
		Type:               openclawv1alpha1.ConditionTypeNetworkPolicyReady,
		Status:             metav1.ConditionTrue,
		Reason:             "NetworkPolicyCreated",
		Message:            "NetworkPolicy created successfully",
		LastTransitionTime: metav1.Now(),
	})

	return nil
}

// reconcileConfigMap reconciles the ConfigMap for openclaw.json
func (r *OpenClawInstanceReconciler) reconcileConfigMap(ctx context.Context, instance *openclawv1alpha1.OpenClawInstance) error {
	// Only create ConfigMap if using raw config (not referencing external ConfigMap)
	if instance.Spec.Config.ConfigMapRef != nil {
		// Using external ConfigMap, nothing to create
		instance.Status.ManagedResources.ConfigMap = ""
		return nil
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resources.ConfigMapName(instance),
			Namespace: instance.Namespace,
		},
	}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, cm, func() error {
		desired := resources.BuildConfigMap(instance)
		cm.Labels = desired.Labels
		cm.Data = desired.Data
		return controllerutil.SetControllerReference(instance, cm, r.Scheme)
	}); err != nil {
		return err
	}
	instance.Status.ManagedResources.ConfigMap = cm.Name

	meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
		Type:               openclawv1alpha1.ConditionTypeConfigValid,
		Status:             metav1.ConditionTrue,
		Reason:             "ConfigMapCreated",
		Message:            "ConfigMap created successfully",
		LastTransitionTime: metav1.Now(),
	})

	return nil
}

// reconcilePVC reconciles the PersistentVolumeClaim
func (r *OpenClawInstanceReconciler) reconcilePVC(ctx context.Context, instance *openclawv1alpha1.OpenClawInstance) error {
	// Check if persistence is enabled
	enabled := instance.Spec.Storage.Persistence.Enabled == nil || *instance.Spec.Storage.Persistence.Enabled

	if !enabled {
		instance.Status.ManagedResources.PVC = ""
		return nil
	}

	// Check if using existing claim
	if instance.Spec.Storage.Persistence.ExistingClaim != "" {
		instance.Status.ManagedResources.PVC = instance.Spec.Storage.Persistence.ExistingClaim
		return nil
	}

	pvc := resources.BuildPVC(instance)
	if err := controllerutil.SetControllerReference(instance, pvc, r.Scheme); err != nil {
		return err
	}

	// PVCs are immutable after creation, so we only create if not exists
	existing := &corev1.PersistentVolumeClaim{}
	if err := r.Get(ctx, client.ObjectKeyFromObject(pvc), existing); err != nil {
		if apierrors.IsNotFound(err) {
			if createErr := r.Create(ctx, pvc); createErr != nil {
				return createErr
			}
		} else {
			return err
		}
	}

	instance.Status.ManagedResources.PVC = pvc.Name

	meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
		Type:               openclawv1alpha1.ConditionTypeStorageReady,
		Status:             metav1.ConditionTrue,
		Reason:             "PVCCreated",
		Message:            "PersistentVolumeClaim created successfully",
		LastTransitionTime: metav1.Now(),
	})

	return nil
}

// reconcilePDB reconciles the PodDisruptionBudget
func (r *OpenClawInstanceReconciler) reconcilePDB(ctx context.Context, instance *openclawv1alpha1.OpenClawInstance) error {
	// Check if PDB is enabled
	enabled := instance.Spec.Availability.PodDisruptionBudget == nil ||
		instance.Spec.Availability.PodDisruptionBudget.Enabled == nil ||
		*instance.Spec.Availability.PodDisruptionBudget.Enabled

	if !enabled {
		// Delete existing PDB if it exists
		pdb := &policyv1.PodDisruptionBudget{}
		pdb.Name = resources.PDBName(instance)
		pdb.Namespace = instance.Namespace
		if err := r.Delete(ctx, pdb); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		instance.Status.ManagedResources.PodDisruptionBudget = ""
		return nil
	}

	pdb := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resources.PDBName(instance),
			Namespace: instance.Namespace,
		},
	}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, pdb, func() error {
		desired := resources.BuildPDB(instance)
		pdb.Labels = desired.Labels
		pdb.Spec = desired.Spec
		return controllerutil.SetControllerReference(instance, pdb, r.Scheme)
	}); err != nil {
		return err
	}
	instance.Status.ManagedResources.PodDisruptionBudget = pdb.Name

	return nil
}

// reconcileDeployment reconciles the Deployment
func (r *OpenClawInstanceReconciler) reconcileDeployment(ctx context.Context, instance *openclawv1alpha1.OpenClawInstance) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resources.DeploymentName(instance),
			Namespace: instance.Namespace,
		},
	}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		desired := resources.BuildDeployment(instance)
		deployment.Labels = desired.Labels
		deployment.Spec = desired.Spec
		return controllerutil.SetControllerReference(instance, deployment, r.Scheme)
	}); err != nil {
		return err
	}
	instance.Status.ManagedResources.Deployment = deployment.Name

	// Check deployment status (deployment already fetched by CreateOrUpdate)
	ready := deployment.Status.ReadyReplicas > 0
	status := metav1.ConditionFalse
	reason := "DeploymentNotReady"
	message := "Deployment is not ready yet"
	if ready {
		status = metav1.ConditionTrue
		reason = "DeploymentReady"
		message = "Deployment is ready"
	}

	meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
		Type:               openclawv1alpha1.ConditionTypeDeploymentReady,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})

	return nil
}

// reconcileService reconciles the Service
func (r *OpenClawInstanceReconciler) reconcileService(ctx context.Context, instance *openclawv1alpha1.OpenClawInstance) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resources.ServiceName(instance),
			Namespace: instance.Namespace,
		},
	}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		desired := resources.BuildService(instance)
		service.Labels = desired.Labels
		service.Annotations = desired.Annotations
		// Preserve ClusterIP — it is assigned by the API server and immutable
		clusterIP := service.Spec.ClusterIP
		clusterIPs := service.Spec.ClusterIPs
		service.Spec = desired.Spec
		service.Spec.ClusterIP = clusterIP
		service.Spec.ClusterIPs = clusterIPs
		return controllerutil.SetControllerReference(instance, service, r.Scheme)
	}); err != nil {
		return err
	}
	instance.Status.ManagedResources.Service = service.Name

	// Update endpoint in status
	instance.Status.GatewayEndpoint = fmt.Sprintf("%s.%s.svc:%d", service.Name, service.Namespace, resources.GatewayPort)
	instance.Status.CanvasEndpoint = fmt.Sprintf("%s.%s.svc:%d", service.Name, service.Namespace, resources.CanvasPort)

	meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
		Type:               openclawv1alpha1.ConditionTypeServiceReady,
		Status:             metav1.ConditionTrue,
		Reason:             "ServiceCreated",
		Message:            "Service created successfully",
		LastTransitionTime: metav1.Now(),
	})

	return nil
}

// reconcileIngress reconciles the Ingress
func (r *OpenClawInstanceReconciler) reconcileIngress(ctx context.Context, instance *openclawv1alpha1.OpenClawInstance) error {
	if !instance.Spec.Networking.Ingress.Enabled {
		// Delete existing Ingress if it exists
		ing := &networkingv1.Ingress{}
		ing.Name = resources.IngressName(instance)
		ing.Namespace = instance.Namespace
		if err := r.Delete(ctx, ing); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		return nil
	}

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resources.IngressName(instance),
			Namespace: instance.Namespace,
		},
	}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ingress, func() error {
		desired := resources.BuildIngress(instance)
		ingress.Labels = desired.Labels
		ingress.Annotations = desired.Annotations
		ingress.Spec = desired.Spec
		return controllerutil.SetControllerReference(instance, ingress, r.Scheme)
	}); err != nil {
		return err
	}

	return nil
}

// reconcileServiceMonitor reconciles the ServiceMonitor for Prometheus
func (r *OpenClawInstanceReconciler) reconcileServiceMonitor(ctx context.Context, instance *openclawv1alpha1.OpenClawInstance) error {
	// Check if ServiceMonitor is enabled
	if instance.Spec.Observability.Metrics.ServiceMonitor == nil ||
		instance.Spec.Observability.Metrics.ServiceMonitor.Enabled == nil ||
		!*instance.Spec.Observability.Metrics.ServiceMonitor.Enabled {
		return nil
	}

	sm := &unstructured.Unstructured{}
	sm.SetGroupVersionKind(resources.ServiceMonitorGVK())
	sm.SetName(resources.ServiceMonitorName(instance))
	sm.SetNamespace(instance.Namespace)

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, sm, func() error {
		desired := resources.BuildServiceMonitor(instance)

		// Copy spec from desired into existing
		if spec, ok := desired.Object["spec"]; ok {
			sm.Object["spec"] = spec
		}
		sm.SetLabels(desired.GetLabels())

		// Set owner reference
		ownerRef := metav1.OwnerReference{
			APIVersion: instance.APIVersion,
			Kind:       instance.Kind,
			Name:       instance.Name,
			UID:        instance.UID,
			Controller: resources.Ptr(true),
		}
		sm.SetOwnerReferences([]metav1.OwnerReference{ownerRef})
		return nil
	})
	return err
}

// reconcileDelete handles cleanup when the instance is being deleted
//
//nolint:unparam // result is always zero-value but signature kept for consistency with Reconcile return pattern
func (r *OpenClawInstanceReconciler) reconcileDelete(ctx context.Context, instance *openclawv1alpha1.OpenClawInstance) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling deletion", "instance", instance.Name, "namespace", instance.Namespace)

	// Update phase
	instance.Status.Phase = openclawv1alpha1.PhaseTerminating
	if err := r.Status().Update(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}

	// Resources are cleaned up automatically via owner references
	// Just remove the finalizer
	controllerutil.RemoveFinalizer(instance, FinalizerName)
	if err := r.Update(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Finalizer removed, cleanup complete")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *OpenClawInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openclawv1alpha1.OpenClawInstance{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Owns(&networkingv1.Ingress{}).
		Owns(&policyv1.PodDisruptionBudget{}).
		Complete(r)
}
