package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// OpenClawInstanceSpec defines the desired state of OpenClawInstance
type OpenClawInstanceSpec struct {
	// Image configuration for the OpenClaw container
	// +optional
	Image ImageSpec `json:"image,omitempty"`

	// Config specifies the OpenClaw configuration
	// +optional
	Config ConfigSpec `json:"config,omitempty"`

	// Workspace configures initial workspace files seeded into the instance.
	// Files are copied once on first boot and never overwritten, so agent
	// modifications survive pod restarts.
	// +optional
	Workspace *WorkspaceSpec `json:"workspace,omitempty"`

	// EnvFrom is a list of sources to populate environment variables from
	// Use this for API keys and other secrets (e.g., ANTHROPIC_API_KEY, OPENAI_API_KEY)
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`

	// Env is a list of environment variables to set in the container
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Resources specifies the compute resources for the OpenClaw container
	// +optional
	Resources ResourcesSpec `json:"resources,omitempty"`

	// Security specifies security-related configuration
	// +optional
	Security SecuritySpec `json:"security,omitempty"`

	// Storage specifies persistent storage configuration
	// +optional
	Storage StorageSpec `json:"storage,omitempty"`

	// Chromium enables the Chromium sidecar for browser automation
	// +optional
	Chromium ChromiumSpec `json:"chromium,omitempty"`

	// Sidecars is a list of additional sidecar containers to inject into the pod.
	// Use this for custom sidecars like database proxies, log forwarders, or service meshes.
	// +optional
	Sidecars []corev1.Container `json:"sidecars,omitempty"`

	// SidecarVolumes is a list of additional volumes to make available to sidecar containers.
	// +optional
	SidecarVolumes []corev1.Volume `json:"sidecarVolumes,omitempty"`

	// Networking specifies network-related configuration
	// +optional
	Networking NetworkingSpec `json:"networking,omitempty"`

	// Probes configures health probes for the OpenClaw container
	// +optional
	Probes ProbesSpec `json:"probes,omitempty"`

	// Observability configures metrics and logging
	// +optional
	Observability ObservabilitySpec `json:"observability,omitempty"`

	// Availability configures high availability settings
	// +optional
	Availability AvailabilitySpec `json:"availability,omitempty"`
}

// ImageSpec defines the container image configuration
type ImageSpec struct {
	// Repository is the container image repository
	// +kubebuilder:default="ghcr.io/openclaw/openclaw"
	// +optional
	Repository string `json:"repository,omitempty"`

	// Tag is the container image tag
	// +kubebuilder:default="latest"
	// +optional
	Tag string `json:"tag,omitempty"`

	// Digest is the container image digest (overrides tag if specified)
	// +optional
	Digest string `json:"digest,omitempty"`

	// PullPolicy specifies when to pull the image
	// +kubebuilder:validation:Enum=Always;IfNotPresent;Never
	// +kubebuilder:default="IfNotPresent"
	// +optional
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`

	// PullSecrets is a list of secret names for pulling from private registries
	// +optional
	PullSecrets []corev1.LocalObjectReference `json:"pullSecrets,omitempty"`
}

// ConfigSpec defines the OpenClaw configuration
type ConfigSpec struct {
	// ConfigMapRef references a ConfigMap containing the openclaw.json configuration
	// +optional
	ConfigMapRef *ConfigMapKeySelector `json:"configMapRef,omitempty"`

	// Raw is inline openclaw.json configuration (used if ConfigMapRef is not set)
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	Raw *RawConfig `json:"raw,omitempty"`
}

// ConfigMapKeySelector selects a key from a ConfigMap
type ConfigMapKeySelector struct {
	// Name of the ConfigMap
	Name string `json:"name"`

	// Key in the ConfigMap to use
	// +kubebuilder:default="openclaw.json"
	// +optional
	Key string `json:"key,omitempty"`
}

// RawConfig holds arbitrary JSON configuration for openclaw.json
// +kubebuilder:pruning:PreserveUnknownFields
type RawConfig struct {
	runtime.RawExtension `json:",inline"`
}

// WorkspaceSpec configures initial workspace files for the instance.
// Files listed in InitialFiles are seeded once (only if they don't already
// exist on the PVC), so agent modifications survive pod restarts.
type WorkspaceSpec struct {
	// InitialFiles maps filenames to their content. Each file is written
	// to the workspace directory only if it does not already exist.
	// +kubebuilder:validation:MaxProperties=50
	// +optional
	InitialFiles map[string]string `json:"initialFiles,omitempty"`

	// InitialDirectories is a list of directories to create (mkdir -p)
	// inside the workspace directory. Nested paths like "tools/scripts" are allowed.
	// +kubebuilder:validation:MaxItems=20
	// +optional
	InitialDirectories []string `json:"initialDirectories,omitempty"`
}

// ResourcesSpec defines compute resource requirements
type ResourcesSpec struct {
	// Requests describes the minimum amount of compute resources required
	// +optional
	Requests ResourceList `json:"requests,omitempty"`

	// Limits describes the maximum amount of compute resources allowed
	// +optional
	Limits ResourceList `json:"limits,omitempty"`
}

// ResourceList defines CPU and memory resources
type ResourceList struct {
	// CPU resource (e.g., "500m", "2")
	// +optional
	CPU string `json:"cpu,omitempty"`

	// Memory resource (e.g., "512Mi", "2Gi")
	// +optional
	Memory string `json:"memory,omitempty"`
}

// SecuritySpec defines security-related configuration
type SecuritySpec struct {
	// PodSecurityContext holds pod-level security attributes
	// +optional
	PodSecurityContext *PodSecurityContextSpec `json:"podSecurityContext,omitempty"`

	// ContainerSecurityContext holds container-level security attributes
	// +optional
	ContainerSecurityContext *ContainerSecurityContextSpec `json:"containerSecurityContext,omitempty"`

	// NetworkPolicy configures network isolation
	// +optional
	NetworkPolicy NetworkPolicySpec `json:"networkPolicy,omitempty"`

	// RBAC configures role-based access control
	// +optional
	RBAC RBACSpec `json:"rbac,omitempty"`
}

// PodSecurityContextSpec defines pod-level security context
type PodSecurityContextSpec struct {
	// RunAsUser is the UID to run the entrypoint of the container process
	// +kubebuilder:default=1000
	// +optional
	RunAsUser *int64 `json:"runAsUser,omitempty"`

	// RunAsGroup is the GID to run the entrypoint of the container process
	// +kubebuilder:default=1000
	// +optional
	RunAsGroup *int64 `json:"runAsGroup,omitempty"`

	// FSGroup is a special supplemental group that applies to all containers
	// +kubebuilder:default=1000
	// +optional
	FSGroup *int64 `json:"fsGroup,omitempty"`

	// RunAsNonRoot indicates that the container must run as a non-root user
	// +kubebuilder:default=true
	// +optional
	RunAsNonRoot *bool `json:"runAsNonRoot,omitempty"`
}

// ContainerSecurityContextSpec defines container-level security context
type ContainerSecurityContextSpec struct {
	// AllowPrivilegeEscalation controls whether a process can gain more privileges
	// +kubebuilder:default=false
	// +optional
	AllowPrivilegeEscalation *bool `json:"allowPrivilegeEscalation,omitempty"`

	// ReadOnlyRootFilesystem mounts the container's root filesystem as read-only
	// Note: OpenClaw requires write access to ~/.openclaw/, so this is false by default
	// +kubebuilder:default=false
	// +optional
	ReadOnlyRootFilesystem *bool `json:"readOnlyRootFilesystem,omitempty"`

	// Capabilities to add/drop
	// +optional
	Capabilities *corev1.Capabilities `json:"capabilities,omitempty"`
}

// NetworkPolicySpec configures network isolation for the OpenClaw instance
type NetworkPolicySpec struct {
	// Enabled enables network policy creation
	// +kubebuilder:default=true
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// AllowedIngressCIDRs is a list of CIDRs allowed to access this instance
	// +optional
	AllowedIngressCIDRs []string `json:"allowedIngressCIDRs,omitempty"`

	// AllowedIngressNamespaces is a list of namespace names allowed to access this instance
	// +optional
	AllowedIngressNamespaces []string `json:"allowedIngressNamespaces,omitempty"`

	// AllowedEgressCIDRs is a list of CIDRs this instance can reach
	// Default allows all egress on port 443 for AI APIs
	// +optional
	AllowedEgressCIDRs []string `json:"allowedEgressCIDRs,omitempty"`

	// AllowDNS allows DNS resolution (port 53)
	// +kubebuilder:default=true
	// +optional
	AllowDNS *bool `json:"allowDNS,omitempty"`

	// AdditionalEgress appends custom egress rules to the default DNS + HTTPS rules.
	// Use this to allow traffic to cluster-internal services on non-standard ports.
	// +optional
	AdditionalEgress []networkingv1.NetworkPolicyEgressRule `json:"additionalEgress,omitempty"`
}

// RBACSpec configures RBAC for the OpenClaw instance
type RBACSpec struct {
	// CreateServiceAccount creates a dedicated ServiceAccount for the instance
	// +kubebuilder:default=true
	// +optional
	CreateServiceAccount *bool `json:"createServiceAccount,omitempty"`

	// ServiceAccountName is the name of an existing ServiceAccount to use
	// Only used if CreateServiceAccount is false
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// AdditionalRules adds custom RBAC rules to the generated Role
	// +optional
	AdditionalRules []RBACRule `json:"additionalRules,omitempty"`
}

// RBACRule represents a RBAC rule
type RBACRule struct {
	// APIGroups is the name of the APIGroup that contains the resources
	APIGroups []string `json:"apiGroups"`
	// Resources is a list of resources this rule applies to
	Resources []string `json:"resources"`
	// Verbs is a list of verbs that apply to the resources
	Verbs []string `json:"verbs"`
}

// StorageSpec defines persistent storage configuration
type StorageSpec struct {
	// Persistence configures the PersistentVolumeClaim
	// +optional
	Persistence PersistenceSpec `json:"persistence,omitempty"`
}

// PersistenceSpec defines PVC configuration
type PersistenceSpec struct {
	// Enabled enables persistent storage
	// +kubebuilder:default=true
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// StorageClass is the name of the StorageClass to use
	// +optional
	StorageClass *string `json:"storageClass,omitempty"`

	// Size is the size of the PVC (e.g., "10Gi")
	// +kubebuilder:default="10Gi"
	// +optional
	Size string `json:"size,omitempty"`

	// AccessModes contains the desired access modes for the PVC
	// +kubebuilder:default={"ReadWriteOnce"}
	// +optional
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`

	// ExistingClaim is the name of an existing PVC to use
	// +optional
	ExistingClaim string `json:"existingClaim,omitempty"`
}

// ChromiumSpec defines the Chromium sidecar configuration
type ChromiumSpec struct {
	// Enabled enables the Chromium sidecar for browser automation
	// +kubebuilder:default=false
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// Image configures the Chromium container image
	// +optional
	Image ChromiumImageSpec `json:"image,omitempty"`

	// Resources specifies compute resources for the Chromium container
	// +optional
	Resources ResourcesSpec `json:"resources,omitempty"`
}

// ChromiumImageSpec defines the Chromium container image
type ChromiumImageSpec struct {
	// Repository is the container image repository
	// +kubebuilder:default="ghcr.io/browserless/chromium"
	// +optional
	Repository string `json:"repository,omitempty"`

	// Tag is the container image tag
	// +kubebuilder:default="latest"
	// +optional
	Tag string `json:"tag,omitempty"`

	// Digest is the container image digest for supply chain security
	// +optional
	Digest string `json:"digest,omitempty"`
}

// NetworkingSpec defines network-related configuration
type NetworkingSpec struct {
	// Service configures the Kubernetes Service
	// +optional
	Service ServiceSpec `json:"service,omitempty"`

	// Ingress configures the Kubernetes Ingress
	// +optional
	Ingress IngressSpec `json:"ingress,omitempty"`
}

// ServiceSpec defines the Service configuration
type ServiceSpec struct {
	// Type is the Kubernetes Service type
	// +kubebuilder:validation:Enum=ClusterIP;LoadBalancer;NodePort
	// +kubebuilder:default="ClusterIP"
	// +optional
	Type corev1.ServiceType `json:"type,omitempty"`

	// Annotations to add to the Service
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// IngressSpec defines the Ingress configuration
type IngressSpec struct {
	// Enabled enables Ingress creation
	// +kubebuilder:default=false
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// ClassName is the name of the IngressClass to use
	// +optional
	ClassName *string `json:"className,omitempty"`

	// Annotations to add to the Ingress
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Hosts is a list of hosts to route traffic for
	// +optional
	Hosts []IngressHost `json:"hosts,omitempty"`

	// TLS configuration
	// +optional
	TLS []IngressTLS `json:"tls,omitempty"`

	// Security configures ingress security settings
	// +optional
	Security IngressSecuritySpec `json:"security,omitempty"`
}

// IngressHost defines a host for the Ingress
type IngressHost struct {
	// Host is the fully qualified domain name
	Host string `json:"host"`

	// Paths is a list of paths to route
	// +optional
	Paths []IngressPath `json:"paths,omitempty"`
}

// IngressPath defines a path for the Ingress
type IngressPath struct {
	// Path is the path to route
	// +kubebuilder:default="/"
	// +optional
	Path string `json:"path,omitempty"`

	// PathType determines how the path should be matched
	// +kubebuilder:validation:Enum=Prefix;Exact;ImplementationSpecific
	// +kubebuilder:default="Prefix"
	// +optional
	PathType string `json:"pathType,omitempty"`
}

// IngressTLS defines TLS configuration for the Ingress
type IngressTLS struct {
	// Hosts are a list of hosts included in the TLS certificate
	Hosts []string `json:"hosts,omitempty"`

	// SecretName is the name of the secret containing the TLS certificate
	SecretName string `json:"secretName,omitempty"`
}

// IngressSecuritySpec defines security settings for the Ingress
type IngressSecuritySpec struct {
	// ForceHTTPS redirects all HTTP traffic to HTTPS
	// +kubebuilder:default=true
	// +optional
	ForceHTTPS *bool `json:"forceHTTPS,omitempty"`

	// EnableHSTS enables HTTP Strict Transport Security
	// +kubebuilder:default=true
	// +optional
	EnableHSTS *bool `json:"enableHSTS,omitempty"`

	// RateLimiting configures rate limiting
	// +optional
	RateLimiting *RateLimitingSpec `json:"rateLimiting,omitempty"`
}

// RateLimitingSpec defines rate limiting configuration
type RateLimitingSpec struct {
	// Enabled enables rate limiting
	// +kubebuilder:default=true
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// RequestsPerSecond is the maximum requests per second
	// +kubebuilder:default=10
	// +optional
	RequestsPerSecond *int32 `json:"requestsPerSecond,omitempty"`
}

// ProbesSpec defines health probe configuration
type ProbesSpec struct {
	// Liveness probe configuration
	// +optional
	Liveness *ProbeSpec `json:"liveness,omitempty"`

	// Readiness probe configuration
	// +optional
	Readiness *ProbeSpec `json:"readiness,omitempty"`

	// Startup probe configuration
	// +optional
	Startup *ProbeSpec `json:"startup,omitempty"`
}

// ProbeSpec defines a health probe
type ProbeSpec struct {
	// Enabled enables the probe
	// +kubebuilder:default=true
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// InitialDelaySeconds is the number of seconds after the container starts before the probe is initiated
	// +optional
	InitialDelaySeconds *int32 `json:"initialDelaySeconds,omitempty"`

	// PeriodSeconds is how often (in seconds) to perform the probe
	// +optional
	PeriodSeconds *int32 `json:"periodSeconds,omitempty"`

	// TimeoutSeconds is the number of seconds after which the probe times out
	// +optional
	TimeoutSeconds *int32 `json:"timeoutSeconds,omitempty"`

	// FailureThreshold is the number of times to retry before giving up
	// +optional
	FailureThreshold *int32 `json:"failureThreshold,omitempty"`
}

// ObservabilitySpec defines observability configuration
type ObservabilitySpec struct {
	// Metrics configures Prometheus metrics
	// +optional
	Metrics MetricsSpec `json:"metrics,omitempty"`

	// Logging configures logging
	// +optional
	Logging LoggingSpec `json:"logging,omitempty"`
}

// MetricsSpec defines metrics configuration
type MetricsSpec struct {
	// Enabled enables metrics endpoint
	// +kubebuilder:default=true
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Port is the port to expose metrics on
	// +kubebuilder:default=9090
	// +optional
	Port *int32 `json:"port,omitempty"`

	// ServiceMonitor configures the Prometheus ServiceMonitor
	// +optional
	ServiceMonitor *ServiceMonitorSpec `json:"serviceMonitor,omitempty"`
}

// ServiceMonitorSpec defines the ServiceMonitor configuration
type ServiceMonitorSpec struct {
	// Enabled enables ServiceMonitor creation
	// +kubebuilder:default=false
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Interval is the scrape interval
	// +kubebuilder:default="30s"
	// +optional
	Interval string `json:"interval,omitempty"`

	// Labels to add to the ServiceMonitor
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}

// LoggingSpec defines logging configuration
type LoggingSpec struct {
	// Level is the log level
	// +kubebuilder:validation:Enum=debug;info;warn;error
	// +kubebuilder:default="info"
	// +optional
	Level string `json:"level,omitempty"`

	// Format is the log format
	// +kubebuilder:validation:Enum=json;text
	// +kubebuilder:default="json"
	// +optional
	Format string `json:"format,omitempty"`
}

// AvailabilitySpec defines high availability settings
type AvailabilitySpec struct {
	// PodDisruptionBudget configures the PDB
	// +optional
	PodDisruptionBudget *PodDisruptionBudgetSpec `json:"podDisruptionBudget,omitempty"`

	// NodeSelector is a selector which must match a node's labels for the pod to be scheduled
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations are tolerations for pod scheduling
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Affinity specifies affinity scheduling rules
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
}

// PodDisruptionBudgetSpec defines PDB configuration
type PodDisruptionBudgetSpec struct {
	// Enabled enables PDB creation
	// +kubebuilder:default=true
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// MaxUnavailable is the maximum number of pods that can be unavailable during disruption
	// +kubebuilder:default=1
	// +optional
	MaxUnavailable *int32 `json:"maxUnavailable,omitempty"`
}

// OpenClawInstanceStatus defines the observed state of OpenClawInstance
type OpenClawInstanceStatus struct {
	// Phase represents the current lifecycle phase of the instance
	// +kubebuilder:validation:Enum=Pending;Provisioning;Running;Degraded;Failed;Terminating
	// +optional
	Phase string `json:"phase,omitempty"`

	// Conditions represent the latest available observations of the instance's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the most recent generation observed by the controller
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// GatewayEndpoint is the endpoint for the OpenClaw gateway
	// +optional
	GatewayEndpoint string `json:"gatewayEndpoint,omitempty"`

	// CanvasEndpoint is the endpoint for the OpenClaw canvas
	// +optional
	CanvasEndpoint string `json:"canvasEndpoint,omitempty"`

	// LastReconcileTime is the timestamp of the last reconciliation
	// +optional
	LastReconcileTime *metav1.Time `json:"lastReconcileTime,omitempty"`

	// ManagedResources tracks the resources created by the operator
	// +optional
	ManagedResources ManagedResourcesStatus `json:"managedResources,omitempty"`
}

// ManagedResourcesStatus tracks resources created by the operator
type ManagedResourcesStatus struct {
	// Deployment is the name of the managed Deployment
	// +optional
	Deployment string `json:"deployment,omitempty"`

	// Service is the name of the managed Service
	// +optional
	Service string `json:"service,omitempty"`

	// ConfigMap is the name of the managed ConfigMap
	// +optional
	ConfigMap string `json:"configMap,omitempty"`

	// PVC is the name of the managed PersistentVolumeClaim
	// +optional
	PVC string `json:"pvc,omitempty"`

	// NetworkPolicy is the name of the managed NetworkPolicy
	// +optional
	NetworkPolicy string `json:"networkPolicy,omitempty"`

	// PodDisruptionBudget is the name of the managed PDB
	// +optional
	PodDisruptionBudget string `json:"podDisruptionBudget,omitempty"`

	// ServiceAccount is the name of the managed ServiceAccount
	// +optional
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// Role is the name of the managed Role
	// +optional
	Role string `json:"role,omitempty"`

	// RoleBinding is the name of the managed RoleBinding
	// +optional
	RoleBinding string `json:"roleBinding,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=='Ready')].status`
// +kubebuilder:printcolumn:name="Gateway",type=string,JSONPath=`.status.gatewayEndpoint`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// OpenClawInstance is the Schema for the openclawinstances API
type OpenClawInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenClawInstanceSpec   `json:"spec,omitempty"`
	Status OpenClawInstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OpenClawInstanceList contains a list of OpenClawInstance
type OpenClawInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenClawInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OpenClawInstance{}, &OpenClawInstanceList{})
}

// Condition types for OpenClawInstance
const (
	// ConditionTypeReady indicates the overall readiness of the instance
	ConditionTypeReady = "Ready"

	// ConditionTypeConfigValid indicates the configuration is valid
	ConditionTypeConfigValid = "ConfigValid"

	// ConditionTypeDeploymentReady indicates the Deployment is ready
	ConditionTypeDeploymentReady = "DeploymentReady"

	// ConditionTypeServiceReady indicates the Service is ready
	ConditionTypeServiceReady = "ServiceReady"

	// ConditionTypeNetworkPolicyReady indicates the NetworkPolicy is ready
	ConditionTypeNetworkPolicyReady = "NetworkPolicyReady"

	// ConditionTypeRBACReady indicates RBAC resources are ready
	ConditionTypeRBACReady = "RBACReady"

	// ConditionTypeStorageReady indicates the PVC is bound
	ConditionTypeStorageReady = "StorageReady"
)

// Phase constants
const (
	PhasePending      = "Pending"
	PhaseProvisioning = "Provisioning"
	PhaseRunning      = "Running"
	PhaseDegraded     = "Degraded"
	PhaseFailed       = "Failed"
	PhaseTerminating  = "Terminating"
)
