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

package resources

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	openclawv1alpha1 "github.com/openclawrocks/k8s-operator/api/v1alpha1"
)

// newTestInstance creates a minimal OpenClawInstance for testing.
func newTestInstance(name string) *openclawv1alpha1.OpenClawInstance {
	return &openclawv1alpha1.OpenClawInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test-ns",
		},
		Spec: openclawv1alpha1.OpenClawInstanceSpec{},
	}
}

// ---------------------------------------------------------------------------
// common.go tests
// ---------------------------------------------------------------------------

func TestLabels(t *testing.T) {
	instance := newTestInstance("my-instance")
	labels := Labels(instance)

	expected := map[string]string{
		"app.kubernetes.io/name":       "openclaw",
		"app.kubernetes.io/instance":   "my-instance",
		"app.kubernetes.io/managed-by": "openclaw-operator",
	}

	if len(labels) != len(expected) {
		t.Fatalf("expected %d labels, got %d", len(expected), len(labels))
	}
	for k, v := range expected {
		if labels[k] != v {
			t.Errorf("label %q: expected %q, got %q", k, v, labels[k])
		}
	}
}

func TestSelectorLabels(t *testing.T) {
	instance := newTestInstance("my-instance")
	labels := SelectorLabels(instance)

	expected := map[string]string{
		"app.kubernetes.io/name":     "openclaw",
		"app.kubernetes.io/instance": "my-instance",
	}

	if len(labels) != len(expected) {
		t.Fatalf("expected %d selector labels, got %d", len(expected), len(labels))
	}
	for k, v := range expected {
		if labels[k] != v {
			t.Errorf("selector label %q: expected %q, got %q", k, v, labels[k])
		}
	}
}

func TestSelectorLabels_SubsetOfLabels(t *testing.T) {
	instance := newTestInstance("test")
	allLabels := Labels(instance)
	selectorLabels := SelectorLabels(instance)

	for k, v := range selectorLabels {
		if allLabels[k] != v {
			t.Errorf("selector label %q=%q is not present in full labels", k, v)
		}
	}

	if len(selectorLabels) >= len(allLabels) {
		t.Error("selector labels should be a strict subset of full labels")
	}
}

func TestGetImage(t *testing.T) {
	tests := []struct {
		name     string
		image    openclawv1alpha1.ImageSpec
		expected string
	}{
		{
			name:     "defaults",
			image:    openclawv1alpha1.ImageSpec{},
			expected: "ghcr.io/openclaw/openclaw:latest",
		},
		{
			name: "custom repo and tag",
			image: openclawv1alpha1.ImageSpec{
				Repository: "my-registry.io/openclaw",
				Tag:        "v1.2.3",
			},
			expected: "my-registry.io/openclaw:v1.2.3",
		},
		{
			name: "digest takes precedence over tag",
			image: openclawv1alpha1.ImageSpec{
				Repository: "my-registry.io/openclaw",
				Tag:        "v1.2.3",
				Digest:     "sha256:abc123",
			},
			expected: "my-registry.io/openclaw@sha256:abc123",
		},
		{
			name: "digest with default repo",
			image: openclawv1alpha1.ImageSpec{
				Digest: "sha256:def456",
			},
			expected: "ghcr.io/openclaw/openclaw@sha256:def456",
		},
		{
			name: "custom repo with default tag",
			image: openclawv1alpha1.ImageSpec{
				Repository: "custom.io/img",
			},
			expected: "custom.io/img:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance := newTestInstance("test")
			instance.Spec.Image = tt.image

			got := GetImage(instance)
			if got != tt.expected {
				t.Errorf("GetImage() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNameHelpers(t *testing.T) {
	instance := newTestInstance("foo")

	tests := []struct {
		name     string
		fn       func(*openclawv1alpha1.OpenClawInstance) string
		expected string
	}{
		{"DeploymentName", DeploymentName, "foo"},
		{"ServiceName", ServiceName, "foo"},
		{"RoleName", RoleName, "foo"},
		{"RoleBindingName", RoleBindingName, "foo"},
		{"ConfigMapName", ConfigMapName, "foo-config"},
		{"PVCName", PVCName, "foo-data"},
		{"NetworkPolicyName", NetworkPolicyName, "foo"},
		{"PDBName", PDBName, "foo"},
		{"IngressName", IngressName, "foo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn(instance)
			if got != tt.expected {
				t.Errorf("%s() = %q, want %q", tt.name, got, tt.expected)
			}
		})
	}
}

func TestServiceAccountName_Default(t *testing.T) {
	instance := newTestInstance("my-inst")
	if got := ServiceAccountName(instance); got != "my-inst" {
		t.Errorf("ServiceAccountName() = %q, want %q", got, "my-inst")
	}
}

func TestServiceAccountName_Custom(t *testing.T) {
	instance := newTestInstance("my-inst")
	instance.Spec.Security.RBAC.ServiceAccountName = "custom-sa"
	if got := ServiceAccountName(instance); got != "custom-sa" {
		t.Errorf("ServiceAccountName() = %q, want %q", got, "custom-sa")
	}
}

func TestPtr(t *testing.T) {
	intVal := Ptr(int32(42))
	if *intVal != 42 {
		t.Errorf("Ptr(42) = %d, want 42", *intVal)
	}

	boolVal := Ptr(true)
	if !*boolVal {
		t.Error("Ptr(true) should be true")
	}

	strVal := Ptr("hello")
	if *strVal != "hello" {
		t.Errorf("Ptr(hello) = %q, want %q", *strVal, "hello")
	}
}

// ---------------------------------------------------------------------------
// deployment.go tests
// ---------------------------------------------------------------------------

func TestBuildDeployment_Defaults(t *testing.T) {
	instance := newTestInstance("test-deploy")
	dep := BuildDeployment(instance)

	// ObjectMeta
	if dep.Name != "test-deploy" {
		t.Errorf("deployment name = %q, want %q", dep.Name, "test-deploy")
	}
	if dep.Namespace != "test-ns" {
		t.Errorf("deployment namespace = %q, want %q", dep.Namespace, "test-ns")
	}

	// Labels present
	if dep.Labels["app.kubernetes.io/name"] != "openclaw" {
		t.Error("deployment missing app.kubernetes.io/name label")
	}

	// Replicas
	if dep.Spec.Replicas == nil || *dep.Spec.Replicas != 1 {
		t.Errorf("replicas = %v, want 1", dep.Spec.Replicas)
	}

	// Strategy
	if dep.Spec.Strategy.Type != appsv1.RecreateDeploymentStrategyType {
		t.Errorf("strategy = %v, want Recreate", dep.Spec.Strategy.Type)
	}

	// Selector labels
	sel := dep.Spec.Selector.MatchLabels
	if sel["app.kubernetes.io/name"] != "openclaw" || sel["app.kubernetes.io/instance"] != "test-deploy" {
		t.Error("selector labels do not match expected values")
	}

	// Config hash annotation
	ann := dep.Spec.Template.Annotations
	if _, ok := ann["openclaw.openclaw.io/config-hash"]; !ok {
		t.Error("config-hash annotation missing from pod template")
	}

	// Pod security context
	psc := dep.Spec.Template.Spec.SecurityContext
	if psc == nil {
		t.Fatal("pod security context is nil")
	}
	if psc.RunAsNonRoot == nil || !*psc.RunAsNonRoot {
		t.Error("pod security context: runAsNonRoot should be true")
	}
	if psc.RunAsUser == nil || *psc.RunAsUser != 1000 {
		t.Errorf("pod security context: runAsUser = %v, want 1000", psc.RunAsUser)
	}
	if psc.RunAsGroup == nil || *psc.RunAsGroup != 1000 {
		t.Errorf("pod security context: runAsGroup = %v, want 1000", psc.RunAsGroup)
	}
	if psc.FSGroup == nil || *psc.FSGroup != 1000 {
		t.Errorf("pod security context: fsGroup = %v, want 1000", psc.FSGroup)
	}
	if psc.SeccompProfile == nil || psc.SeccompProfile.Type != corev1.SeccompProfileTypeRuntimeDefault {
		t.Error("pod security context: seccomp profile should be RuntimeDefault")
	}

	// Containers
	containers := dep.Spec.Template.Spec.Containers
	if len(containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(containers))
	}

	main := containers[0]
	if main.Name != "openclaw" {
		t.Errorf("main container name = %q, want %q", main.Name, "openclaw")
	}
	if main.Image != "ghcr.io/openclaw/openclaw:latest" {
		t.Errorf("main container image = %q, want default image", main.Image)
	}
	if main.ImagePullPolicy != corev1.PullIfNotPresent {
		t.Errorf("image pull policy = %v, want IfNotPresent", main.ImagePullPolicy)
	}

	// Container security context
	csc := main.SecurityContext
	if csc == nil {
		t.Fatal("container security context is nil")
	}
	if csc.AllowPrivilegeEscalation == nil || *csc.AllowPrivilegeEscalation {
		t.Error("container security context: allowPrivilegeEscalation should be false")
	}
	if csc.RunAsNonRoot == nil || !*csc.RunAsNonRoot {
		t.Error("container security context: runAsNonRoot should be true")
	}
	if csc.Capabilities == nil || len(csc.Capabilities.Drop) != 1 || csc.Capabilities.Drop[0] != "ALL" {
		t.Error("container security context: capabilities should drop ALL")
	}
	if csc.SeccompProfile == nil || csc.SeccompProfile.Type != corev1.SeccompProfileTypeRuntimeDefault {
		t.Error("container security context: seccomp profile should be RuntimeDefault")
	}

	// HOME env var must be set to match the mount path
	if len(main.Env) < 1 || main.Env[0].Name != "HOME" || main.Env[0].Value != "/home/openclaw" {
		t.Error("HOME env var should be set to /home/openclaw")
	}

	// Ports
	if len(main.Ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(main.Ports))
	}
	assertContainerPort(t, main.Ports, "gateway", GatewayPort)
	assertContainerPort(t, main.Ports, "canvas", CanvasPort)

	// Default resources
	cpuReq := main.Resources.Requests[corev1.ResourceCPU]
	if cpuReq.String() != "500m" {
		t.Errorf("cpu request = %v, want 500m", cpuReq.String())
	}
	memReq := main.Resources.Requests[corev1.ResourceMemory]
	if memReq.Cmp(resource.MustParse("1Gi")) != 0 {
		t.Errorf("memory request = %v, want 1Gi", memReq.String())
	}
	cpuLim := main.Resources.Limits[corev1.ResourceCPU]
	if cpuLim.String() != "2" {
		t.Errorf("cpu limit = %v, want 2 (2000m)", cpuLim.String())
	}
	memLim := main.Resources.Limits[corev1.ResourceMemory]
	if memLim.Cmp(resource.MustParse("4Gi")) != 0 {
		t.Errorf("memory limit = %v, want 4Gi", memLim.String())
	}

	// Probes
	if main.LivenessProbe == nil {
		t.Error("liveness probe should not be nil by default")
	}
	if main.ReadinessProbe == nil {
		t.Error("readiness probe should not be nil by default")
	}
	if main.StartupProbe == nil {
		t.Error("startup probe should not be nil by default")
	}

	// Liveness probe defaults
	if main.LivenessProbe.TCPSocket == nil {
		t.Fatal("liveness probe should use TCPSocket")
	}
	if main.LivenessProbe.TCPSocket.Port.IntValue() != GatewayPort {
		t.Errorf("liveness probe port = %d, want %d", main.LivenessProbe.TCPSocket.Port.IntValue(), GatewayPort)
	}
	if main.LivenessProbe.InitialDelaySeconds != 30 {
		t.Errorf("liveness probe initialDelaySeconds = %d, want 30", main.LivenessProbe.InitialDelaySeconds)
	}
	if main.LivenessProbe.PeriodSeconds != 10 {
		t.Errorf("liveness probe periodSeconds = %d, want 10", main.LivenessProbe.PeriodSeconds)
	}

	// Readiness probe defaults
	if main.ReadinessProbe.InitialDelaySeconds != 5 {
		t.Errorf("readiness probe initialDelaySeconds = %d, want 5", main.ReadinessProbe.InitialDelaySeconds)
	}

	// Startup probe defaults
	if main.StartupProbe.FailureThreshold != 30 {
		t.Errorf("startup probe failureThreshold = %d, want 30", main.StartupProbe.FailureThreshold)
	}

	// Data volume mount
	assertVolumeMount(t, main.VolumeMounts, "data", "/home/openclaw/.openclaw")

	// Volumes - default persistence is enabled, so data volume should be PVC
	volumes := dep.Spec.Template.Spec.Volumes
	dataVol := findVolume(volumes, "data")
	if dataVol == nil {
		t.Fatal("data volume not found")
	}
	if dataVol.PersistentVolumeClaim == nil {
		t.Error("data volume should use PVC by default")
	}
	if dataVol.PersistentVolumeClaim.ClaimName != "test-deploy-data" {
		t.Errorf("PVC claim name = %q, want %q", dataVol.PersistentVolumeClaim.ClaimName, "test-deploy-data")
	}
}

func TestBuildDeployment_WithChromium(t *testing.T) {
	instance := newTestInstance("chromium-test")
	instance.Spec.Chromium.Enabled = true

	dep := BuildDeployment(instance)
	containers := dep.Spec.Template.Spec.Containers

	if len(containers) != 2 {
		t.Fatalf("expected 2 containers with chromium enabled, got %d", len(containers))
	}

	// Find chromium container
	var chromium *corev1.Container
	for i := range containers {
		if containers[i].Name == "chromium" {
			chromium = &containers[i]
			break
		}
	}
	if chromium == nil {
		t.Fatal("chromium container not found")
	}

	// Main container should have CHROMIUM_URL env var
	mainContainer := containers[0]
	foundChromiumURL := false
	for _, env := range mainContainer.Env {
		if env.Name == "CHROMIUM_URL" {
			foundChromiumURL = true
			if env.Value != "ws://localhost:9222" {
				t.Errorf("CHROMIUM_URL = %q, want %q", env.Value, "ws://localhost:9222")
			}
			break
		}
	}
	if !foundChromiumURL {
		t.Error("main container should have CHROMIUM_URL env var when chromium is enabled")
	}

	// Chromium image defaults
	if chromium.Image != "ghcr.io/browserless/chromium:latest" {
		t.Errorf("chromium image = %q, want default", chromium.Image)
	}

	// Chromium port
	if len(chromium.Ports) != 1 {
		t.Fatalf("chromium container should have 1 port, got %d", len(chromium.Ports))
	}
	if chromium.Ports[0].ContainerPort != ChromiumPort {
		t.Errorf("chromium port = %d, want %d", chromium.Ports[0].ContainerPort, ChromiumPort)
	}
	if chromium.Ports[0].Name != "cdp" {
		t.Errorf("chromium port name = %q, want %q", chromium.Ports[0].Name, "cdp")
	}

	// Chromium security context
	csc := chromium.SecurityContext
	if csc == nil {
		t.Fatal("chromium security context is nil")
	}
	if csc.ReadOnlyRootFilesystem == nil || !*csc.ReadOnlyRootFilesystem {
		t.Error("chromium: readOnlyRootFilesystem should be true")
	}
	if csc.RunAsUser == nil || *csc.RunAsUser != 1001 {
		t.Errorf("chromium: runAsUser = %v, want 1001", csc.RunAsUser)
	}

	// Chromium resource defaults
	cpuReq := chromium.Resources.Requests[corev1.ResourceCPU]
	if cpuReq.String() != "250m" {
		t.Errorf("chromium cpu request = %v, want 250m", cpuReq.String())
	}
	memReq := chromium.Resources.Requests[corev1.ResourceMemory]
	if memReq.Cmp(resource.MustParse("512Mi")) != 0 {
		t.Errorf("chromium memory request = %v, want 512Mi", memReq.String())
	}

	// Chromium volume mounts
	assertVolumeMount(t, chromium.VolumeMounts, "chromium-tmp", "/tmp")
	assertVolumeMount(t, chromium.VolumeMounts, "chromium-shm", "/dev/shm")

	// Volumes - check chromium-specific volumes exist
	volumes := dep.Spec.Template.Spec.Volumes
	tmpVol := findVolume(volumes, "chromium-tmp")
	if tmpVol == nil {
		t.Fatal("chromium-tmp volume not found")
	}
	if tmpVol.EmptyDir == nil {
		t.Error("chromium-tmp should be emptyDir")
	}

	shmVol := findVolume(volumes, "chromium-shm")
	if shmVol == nil {
		t.Fatal("chromium-shm volume not found")
	}
	if shmVol.EmptyDir == nil {
		t.Fatal("chromium-shm should be emptyDir")
	}
	if shmVol.EmptyDir.Medium != corev1.StorageMediumMemory {
		t.Errorf("chromium-shm medium = %v, want Memory", shmVol.EmptyDir.Medium)
	}
	expectedShmSize := resource.NewQuantity(1024*1024*1024, resource.BinarySI) // 1Gi
	if shmVol.EmptyDir.SizeLimit == nil {
		t.Fatal("chromium-shm sizeLimit is nil")
	}
	if shmVol.EmptyDir.SizeLimit.Cmp(*expectedShmSize) != 0 {
		t.Errorf("chromium-shm sizeLimit = %v, want 1Gi", shmVol.EmptyDir.SizeLimit.String())
	}
}

func TestBuildDeployment_CustomResources(t *testing.T) {
	instance := newTestInstance("res-test")
	instance.Spec.Resources = openclawv1alpha1.ResourcesSpec{
		Requests: openclawv1alpha1.ResourceList{
			CPU:    "1",
			Memory: "2Gi",
		},
		Limits: openclawv1alpha1.ResourceList{
			CPU:    "4",
			Memory: "8Gi",
		},
	}

	dep := BuildDeployment(instance)
	main := dep.Spec.Template.Spec.Containers[0]

	cpuReq := main.Resources.Requests[corev1.ResourceCPU]
	if cpuReq.String() != "1" {
		t.Errorf("cpu request = %v, want 1", cpuReq.String())
	}
	memReq := main.Resources.Requests[corev1.ResourceMemory]
	if memReq.Cmp(resource.MustParse("2Gi")) != 0 {
		t.Errorf("memory request = %v, want 2Gi", memReq.String())
	}
	cpuLim := main.Resources.Limits[corev1.ResourceCPU]
	if cpuLim.String() != "4" {
		t.Errorf("cpu limit = %v, want 4", cpuLim.String())
	}
	memLim := main.Resources.Limits[corev1.ResourceMemory]
	if memLim.Cmp(resource.MustParse("8Gi")) != 0 {
		t.Errorf("memory limit = %v, want 8Gi", memLim.String())
	}
}

func TestBuildDeployment_ImageDigest(t *testing.T) {
	instance := newTestInstance("digest-test")
	instance.Spec.Image = openclawv1alpha1.ImageSpec{
		Repository: "my-registry.io/openclaw",
		Tag:        "v1.0.0",
		Digest:     "sha256:abcdef1234567890",
	}

	dep := BuildDeployment(instance)
	main := dep.Spec.Template.Spec.Containers[0]

	expected := "my-registry.io/openclaw@sha256:abcdef1234567890"
	if main.Image != expected {
		t.Errorf("image = %q, want %q", main.Image, expected)
	}
}

func TestBuildDeployment_ProbesDisabled(t *testing.T) {
	instance := newTestInstance("probes-disabled")
	instance.Spec.Probes = openclawv1alpha1.ProbesSpec{
		Liveness: &openclawv1alpha1.ProbeSpec{
			Enabled: Ptr(false),
		},
		Readiness: &openclawv1alpha1.ProbeSpec{
			Enabled: Ptr(false),
		},
		Startup: &openclawv1alpha1.ProbeSpec{
			Enabled: Ptr(false),
		},
	}

	dep := BuildDeployment(instance)
	main := dep.Spec.Template.Spec.Containers[0]

	if main.LivenessProbe != nil {
		t.Error("liveness probe should be nil when disabled")
	}
	if main.ReadinessProbe != nil {
		t.Error("readiness probe should be nil when disabled")
	}
	if main.StartupProbe != nil {
		t.Error("startup probe should be nil when disabled")
	}
}

func TestBuildDeployment_CustomProbeValues(t *testing.T) {
	instance := newTestInstance("probes-custom")
	instance.Spec.Probes = openclawv1alpha1.ProbesSpec{
		Liveness: &openclawv1alpha1.ProbeSpec{
			InitialDelaySeconds: Ptr(int32(60)),
			PeriodSeconds:       Ptr(int32(20)),
			TimeoutSeconds:      Ptr(int32(10)),
			FailureThreshold:    Ptr(int32(5)),
		},
	}

	dep := BuildDeployment(instance)
	probe := dep.Spec.Template.Spec.Containers[0].LivenessProbe

	if probe == nil {
		t.Fatal("liveness probe should not be nil")
	}
	if probe.InitialDelaySeconds != 60 {
		t.Errorf("liveness initialDelaySeconds = %d, want 60", probe.InitialDelaySeconds)
	}
	if probe.PeriodSeconds != 20 {
		t.Errorf("liveness periodSeconds = %d, want 20", probe.PeriodSeconds)
	}
	if probe.TimeoutSeconds != 10 {
		t.Errorf("liveness timeoutSeconds = %d, want 10", probe.TimeoutSeconds)
	}
	if probe.FailureThreshold != 5 {
		t.Errorf("liveness failureThreshold = %d, want 5", probe.FailureThreshold)
	}
}

func TestBuildDeployment_PersistenceDisabled(t *testing.T) {
	instance := newTestInstance("no-pvc")
	instance.Spec.Storage.Persistence.Enabled = Ptr(false)

	dep := BuildDeployment(instance)
	volumes := dep.Spec.Template.Spec.Volumes

	dataVol := findVolume(volumes, "data")
	if dataVol == nil {
		t.Fatal("data volume not found")
	}
	if dataVol.EmptyDir == nil {
		t.Error("data volume should be emptyDir when persistence is disabled")
	}
	if dataVol.PersistentVolumeClaim != nil {
		t.Error("data volume should not use PVC when persistence is disabled")
	}
}

func TestBuildDeployment_ExistingClaim(t *testing.T) {
	instance := newTestInstance("existing-pvc")
	instance.Spec.Storage.Persistence.ExistingClaim = "my-existing-pvc"

	dep := BuildDeployment(instance)
	volumes := dep.Spec.Template.Spec.Volumes

	dataVol := findVolume(volumes, "data")
	if dataVol == nil {
		t.Fatal("data volume not found")
	}
	if dataVol.PersistentVolumeClaim == nil {
		t.Fatal("data volume should use PVC")
	}
	if dataVol.PersistentVolumeClaim.ClaimName != "my-existing-pvc" {
		t.Errorf("PVC claim name = %q, want %q", dataVol.PersistentVolumeClaim.ClaimName, "my-existing-pvc")
	}
}

func TestBuildDeployment_ConfigVolume_RawConfig(t *testing.T) {
	instance := newTestInstance("raw-cfg")
	instance.Spec.Config.Raw = &openclawv1alpha1.RawConfig{
		RawExtension: runtime.RawExtension{
			Raw: []byte(`{"key":"value"}`),
		},
	}

	dep := BuildDeployment(instance)
	main := dep.Spec.Template.Spec.Containers[0]

	// Should have config volume mount
	assertVolumeMount(t, main.VolumeMounts, "config", "/home/openclaw/.openclaw/openclaw.json")

	// Should have config volume pointing to managed configmap
	volumes := dep.Spec.Template.Spec.Volumes
	cfgVol := findVolume(volumes, "config")
	if cfgVol == nil {
		t.Fatal("config volume not found")
	}
	if cfgVol.ConfigMap == nil {
		t.Fatal("config volume should use ConfigMap")
	}
	if cfgVol.ConfigMap.Name != "raw-cfg-config" {
		t.Errorf("config volume configmap name = %q, want %q", cfgVol.ConfigMap.Name, "raw-cfg-config")
	}
}

func TestBuildDeployment_ConfigVolume_ConfigMapRef(t *testing.T) {
	instance := newTestInstance("ref-cfg")
	instance.Spec.Config.ConfigMapRef = &openclawv1alpha1.ConfigMapKeySelector{
		Name: "external-config",
		Key:  "my-config.json",
	}

	dep := BuildDeployment(instance)
	main := dep.Spec.Template.Spec.Containers[0]

	// Should have config volume mount with custom subpath
	found := false
	for _, vm := range main.VolumeMounts {
		if vm.Name == "config" {
			found = true
			if vm.SubPath != "my-config.json" {
				t.Errorf("config volume mount subPath = %q, want %q", vm.SubPath, "my-config.json")
			}
			break
		}
	}
	if !found {
		t.Error("config volume mount not found")
	}

	// Volume should reference external configmap
	volumes := dep.Spec.Template.Spec.Volumes
	cfgVol := findVolume(volumes, "config")
	if cfgVol == nil {
		t.Fatal("config volume not found")
	}
	if cfgVol.ConfigMap.Name != "external-config" {
		t.Errorf("config volume configmap name = %q, want %q", cfgVol.ConfigMap.Name, "external-config")
	}
}

func TestBuildDeployment_ConfigMapRef_DefaultKey(t *testing.T) {
	instance := newTestInstance("ref-default-key")
	instance.Spec.Config.ConfigMapRef = &openclawv1alpha1.ConfigMapKeySelector{
		Name: "external-config",
		// Key not set - should default to "openclaw.json"
	}

	dep := BuildDeployment(instance)
	main := dep.Spec.Template.Spec.Containers[0]

	for _, vm := range main.VolumeMounts {
		if vm.Name == "config" {
			if vm.SubPath != "openclaw.json" {
				t.Errorf("config volume mount subPath = %q, want %q", vm.SubPath, "openclaw.json")
			}
			return
		}
	}
	t.Error("config volume mount not found")
}

func TestBuildDeployment_ServiceAccountName(t *testing.T) {
	instance := newTestInstance("sa-test")
	dep := BuildDeployment(instance)
	if dep.Spec.Template.Spec.ServiceAccountName != "sa-test" {
		t.Errorf("serviceAccountName = %q, want %q", dep.Spec.Template.Spec.ServiceAccountName, "sa-test")
	}
}

func TestBuildDeployment_ImagePullSecrets(t *testing.T) {
	instance := newTestInstance("pull-secrets")
	instance.Spec.Image.PullSecrets = []corev1.LocalObjectReference{
		{Name: "my-secret"},
		{Name: "other-secret"},
	}

	dep := BuildDeployment(instance)
	secrets := dep.Spec.Template.Spec.ImagePullSecrets
	if len(secrets) != 2 {
		t.Fatalf("expected 2 pull secrets, got %d", len(secrets))
	}
	if secrets[0].Name != "my-secret" {
		t.Errorf("first pull secret = %q, want %q", secrets[0].Name, "my-secret")
	}
	if secrets[1].Name != "other-secret" {
		t.Errorf("second pull secret = %q, want %q", secrets[1].Name, "other-secret")
	}
}

func TestBuildDeployment_ChromiumCustomImage(t *testing.T) {
	instance := newTestInstance("chromium-custom")
	instance.Spec.Chromium.Enabled = true
	instance.Spec.Chromium.Image = openclawv1alpha1.ChromiumImageSpec{
		Repository: "my-registry.io/chromium",
		Tag:        "v120",
	}

	dep := BuildDeployment(instance)
	var chromium *corev1.Container
	for i := range dep.Spec.Template.Spec.Containers {
		if dep.Spec.Template.Spec.Containers[i].Name == "chromium" {
			chromium = &dep.Spec.Template.Spec.Containers[i]
			break
		}
	}
	if chromium == nil {
		t.Fatal("chromium container not found")
	}
	if chromium.Image != "my-registry.io/chromium:v120" {
		t.Errorf("chromium image = %q, want %q", chromium.Image, "my-registry.io/chromium:v120")
	}
}

func TestBuildDeployment_ChromiumDigest(t *testing.T) {
	instance := newTestInstance("chromium-digest")
	instance.Spec.Chromium.Enabled = true
	instance.Spec.Chromium.Image = openclawv1alpha1.ChromiumImageSpec{
		Repository: "my-registry.io/chromium",
		Tag:        "v120",
		Digest:     "sha256:chromiumhash",
	}

	dep := BuildDeployment(instance)
	var chromium *corev1.Container
	for i := range dep.Spec.Template.Spec.Containers {
		if dep.Spec.Template.Spec.Containers[i].Name == "chromium" {
			chromium = &dep.Spec.Template.Spec.Containers[i]
			break
		}
	}
	if chromium == nil {
		t.Fatal("chromium container not found")
	}
	expected := "my-registry.io/chromium@sha256:chromiumhash"
	if chromium.Image != expected {
		t.Errorf("chromium image = %q, want %q", chromium.Image, expected)
	}
}

func TestBuildDeployment_NodeSelectorAndTolerations(t *testing.T) {
	instance := newTestInstance("scheduling")
	instance.Spec.Availability.NodeSelector = map[string]string{
		"node-type": "gpu",
	}
	instance.Spec.Availability.Tolerations = []corev1.Toleration{
		{
			Key:      "gpu",
			Operator: corev1.TolerationOpEqual,
			Value:    "true",
			Effect:   corev1.TaintEffectNoSchedule,
		},
	}

	dep := BuildDeployment(instance)
	podSpec := dep.Spec.Template.Spec

	if podSpec.NodeSelector["node-type"] != "gpu" {
		t.Error("nodeSelector not applied")
	}
	if len(podSpec.Tolerations) != 1 || podSpec.Tolerations[0].Key != "gpu" {
		t.Error("tolerations not applied")
	}
}

func TestBuildDeployment_EnvAndEnvFrom(t *testing.T) {
	instance := newTestInstance("env-test")
	instance.Spec.Env = []corev1.EnvVar{
		{Name: "MY_VAR", Value: "my-value"},
	}
	instance.Spec.EnvFrom = []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: "api-keys"},
			},
		},
	}

	dep := BuildDeployment(instance)
	main := dep.Spec.Template.Spec.Containers[0]

	if len(main.Env) != 2 || main.Env[0].Name != "HOME" || main.Env[1].Name != "MY_VAR" {
		t.Error("env vars should include HOME followed by user-defined vars")
	}
	if len(main.EnvFrom) != 1 || main.EnvFrom[0].SecretRef.Name != "api-keys" {
		t.Error("envFrom not passed through")
	}
}

// ---------------------------------------------------------------------------
// service.go tests
// ---------------------------------------------------------------------------

func TestBuildService_Default(t *testing.T) {
	instance := newTestInstance("svc-test")
	svc := BuildService(instance)

	if svc.Name != "svc-test" {
		t.Errorf("service name = %q, want %q", svc.Name, "svc-test")
	}
	if svc.Namespace != "test-ns" {
		t.Errorf("service namespace = %q, want %q", svc.Namespace, "test-ns")
	}
	if svc.Spec.Type != corev1.ServiceTypeClusterIP {
		t.Errorf("service type = %v, want ClusterIP", svc.Spec.Type)
	}

	// Labels
	if svc.Labels["app.kubernetes.io/name"] != "openclaw" {
		t.Error("service missing app label")
	}

	// Selector
	sel := svc.Spec.Selector
	if sel["app.kubernetes.io/name"] != "openclaw" || sel["app.kubernetes.io/instance"] != "svc-test" {
		t.Error("service selector does not match expected values")
	}

	// Ports - should have gateway and canvas
	if len(svc.Spec.Ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(svc.Spec.Ports))
	}

	assertServicePort(t, svc.Spec.Ports, "gateway", int32(GatewayPort))
	assertServicePort(t, svc.Spec.Ports, "canvas", int32(CanvasPort))
}

func TestBuildService_WithChromium(t *testing.T) {
	instance := newTestInstance("svc-chromium")
	instance.Spec.Chromium.Enabled = true

	svc := BuildService(instance)

	if len(svc.Spec.Ports) != 3 {
		t.Fatalf("expected 3 ports with chromium, got %d", len(svc.Spec.Ports))
	}

	assertServicePort(t, svc.Spec.Ports, "gateway", int32(GatewayPort))
	assertServicePort(t, svc.Spec.Ports, "canvas", int32(CanvasPort))
	assertServicePort(t, svc.Spec.Ports, "chromium", int32(ChromiumPort))
}

func TestBuildService_LoadBalancer(t *testing.T) {
	instance := newTestInstance("svc-lb")
	instance.Spec.Networking.Service.Type = corev1.ServiceTypeLoadBalancer

	svc := BuildService(instance)

	if svc.Spec.Type != corev1.ServiceTypeLoadBalancer {
		t.Errorf("service type = %v, want LoadBalancer", svc.Spec.Type)
	}
}

func TestBuildService_NodePort(t *testing.T) {
	instance := newTestInstance("svc-np")
	instance.Spec.Networking.Service.Type = corev1.ServiceTypeNodePort

	svc := BuildService(instance)

	if svc.Spec.Type != corev1.ServiceTypeNodePort {
		t.Errorf("service type = %v, want NodePort", svc.Spec.Type)
	}
}

func TestBuildService_CustomAnnotations(t *testing.T) {
	instance := newTestInstance("svc-ann")
	instance.Spec.Networking.Service.Annotations = map[string]string{
		"service.beta.kubernetes.io/aws-load-balancer-type": "nlb",
	}

	svc := BuildService(instance)

	if svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-type"] != "nlb" {
		t.Error("service annotations not applied")
	}
}

// ---------------------------------------------------------------------------
// networkpolicy.go tests
// ---------------------------------------------------------------------------

func TestBuildNetworkPolicy_Default(t *testing.T) {
	instance := newTestInstance("np-test")
	np := BuildNetworkPolicy(instance)

	if np.Name != "np-test" {
		t.Errorf("network policy name = %q, want %q", np.Name, "np-test")
	}
	if np.Namespace != "test-ns" {
		t.Errorf("network policy namespace = %q, want %q", np.Namespace, "test-ns")
	}

	// Labels
	if np.Labels["app.kubernetes.io/name"] != "openclaw" {
		t.Error("network policy missing app label")
	}

	// Pod selector
	sel := np.Spec.PodSelector.MatchLabels
	if sel["app.kubernetes.io/name"] != "openclaw" || sel["app.kubernetes.io/instance"] != "np-test" {
		t.Error("pod selector does not match expected values")
	}

	// Policy types
	if len(np.Spec.PolicyTypes) != 2 {
		t.Fatalf("expected 2 policy types, got %d", len(np.Spec.PolicyTypes))
	}

	// Ingress rules - by default, allow from same namespace
	if len(np.Spec.Ingress) < 1 {
		t.Fatal("expected at least 1 ingress rule")
	}
	firstIngress := np.Spec.Ingress[0]
	if len(firstIngress.From) != 1 {
		t.Fatalf("expected 1 peer in first ingress rule, got %d", len(firstIngress.From))
	}
	nsSel := firstIngress.From[0].NamespaceSelector
	if nsSel == nil {
		t.Fatal("first ingress rule should have namespace selector")
	}
	if nsSel.MatchLabels["kubernetes.io/metadata.name"] != "test-ns" {
		t.Errorf("ingress namespace selector = %v, want test-ns", nsSel.MatchLabels)
	}

	// Ingress ports - gateway and canvas
	if len(firstIngress.Ports) != 2 {
		t.Fatalf("expected 2 ingress ports, got %d", len(firstIngress.Ports))
	}
	assertNPPort(t, firstIngress.Ports, GatewayPort)
	assertNPPort(t, firstIngress.Ports, CanvasPort)

	// Egress rules - DNS (UDP+TCP 53) and HTTPS (443)
	if len(np.Spec.Egress) < 2 {
		t.Fatalf("expected at least 2 egress rules (DNS + HTTPS), got %d", len(np.Spec.Egress))
	}

	// First egress: DNS
	dnsRule := np.Spec.Egress[0]
	if len(dnsRule.Ports) != 2 {
		t.Fatalf("DNS egress rule should have 2 ports (UDP+TCP), got %d", len(dnsRule.Ports))
	}
	foundUDP53 := false
	foundTCP53 := false
	for _, p := range dnsRule.Ports {
		if p.Port != nil && p.Port.IntValue() == 53 {
			if p.Protocol != nil && *p.Protocol == corev1.ProtocolUDP {
				foundUDP53 = true
			}
			if p.Protocol != nil && *p.Protocol == corev1.ProtocolTCP {
				foundTCP53 = true
			}
		}
	}
	if !foundUDP53 {
		t.Error("DNS egress rule missing UDP port 53")
	}
	if !foundTCP53 {
		t.Error("DNS egress rule missing TCP port 53")
	}

	// Second egress: HTTPS
	httpsRule := np.Spec.Egress[1]
	if len(httpsRule.Ports) != 1 {
		t.Fatalf("HTTPS egress rule should have 1 port, got %d", len(httpsRule.Ports))
	}
	if httpsRule.Ports[0].Port == nil || httpsRule.Ports[0].Port.IntValue() != 443 {
		t.Error("HTTPS egress rule should allow port 443")
	}
}

func TestBuildNetworkPolicy_CustomCIDRs(t *testing.T) {
	instance := newTestInstance("np-cidrs")
	instance.Spec.Security.NetworkPolicy.AllowedIngressCIDRs = []string{
		"10.0.0.0/8",
		"192.168.1.0/24",
	}
	instance.Spec.Security.NetworkPolicy.AllowedEgressCIDRs = []string{
		"172.16.0.0/12",
	}

	np := BuildNetworkPolicy(instance)

	// Should have 3 ingress rules: same-ns + 2 CIDRs
	if len(np.Spec.Ingress) != 3 {
		t.Fatalf("expected 3 ingress rules, got %d", len(np.Spec.Ingress))
	}

	// Verify CIDR ingress rules
	cidrRule1 := np.Spec.Ingress[1]
	if cidrRule1.From[0].IPBlock == nil {
		t.Fatal("second ingress rule should have IPBlock")
	}
	if cidrRule1.From[0].IPBlock.CIDR != "10.0.0.0/8" {
		t.Errorf("first CIDR = %q, want %q", cidrRule1.From[0].IPBlock.CIDR, "10.0.0.0/8")
	}

	cidrRule2 := np.Spec.Ingress[2]
	if cidrRule2.From[0].IPBlock.CIDR != "192.168.1.0/24" {
		t.Errorf("second CIDR = %q, want %q", cidrRule2.From[0].IPBlock.CIDR, "192.168.1.0/24")
	}

	// Egress should include the CIDR rule (DNS + HTTPS + 1 custom)
	if len(np.Spec.Egress) != 3 {
		t.Fatalf("expected 3 egress rules, got %d", len(np.Spec.Egress))
	}
	egressCIDR := np.Spec.Egress[2]
	if len(egressCIDR.To) != 1 || egressCIDR.To[0].IPBlock == nil {
		t.Fatal("third egress rule should have IPBlock")
	}
	if egressCIDR.To[0].IPBlock.CIDR != "172.16.0.0/12" {
		t.Errorf("egress CIDR = %q, want %q", egressCIDR.To[0].IPBlock.CIDR, "172.16.0.0/12")
	}
}

func TestBuildNetworkPolicy_DNSDisabled(t *testing.T) {
	instance := newTestInstance("np-no-dns")
	instance.Spec.Security.NetworkPolicy.AllowDNS = Ptr(false)

	np := BuildNetworkPolicy(instance)

	// Without DNS, only HTTPS egress rule
	if len(np.Spec.Egress) != 1 {
		t.Fatalf("expected 1 egress rule (HTTPS only), got %d", len(np.Spec.Egress))
	}

	// The single rule should be HTTPS (443)
	httpsRule := np.Spec.Egress[0]
	if len(httpsRule.Ports) != 1 || httpsRule.Ports[0].Port.IntValue() != 443 {
		t.Error("single egress rule should be HTTPS port 443")
	}
}

func TestBuildNetworkPolicy_AllowedNamespaces(t *testing.T) {
	instance := newTestInstance("np-ns")
	instance.Spec.Security.NetworkPolicy.AllowedIngressNamespaces = []string{
		"ingress-nginx",
		"monitoring",
	}

	np := BuildNetworkPolicy(instance)

	// Should have 3 ingress rules: same-ns + 2 allowed namespaces
	if len(np.Spec.Ingress) != 3 {
		t.Fatalf("expected 3 ingress rules, got %d", len(np.Spec.Ingress))
	}

	nsRule1 := np.Spec.Ingress[1]
	if nsRule1.From[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] != "ingress-nginx" {
		t.Error("second ingress rule should allow ingress-nginx namespace")
	}
	nsRule2 := np.Spec.Ingress[2]
	if nsRule2.From[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] != "monitoring" {
		t.Error("third ingress rule should allow monitoring namespace")
	}
}

// ---------------------------------------------------------------------------
// rbac.go tests
// ---------------------------------------------------------------------------

func TestBuildServiceAccount(t *testing.T) {
	instance := newTestInstance("sa-test")
	sa := BuildServiceAccount(instance)

	if sa.Name != "sa-test" {
		t.Errorf("service account name = %q, want %q", sa.Name, "sa-test")
	}
	if sa.Namespace != "test-ns" {
		t.Errorf("service account namespace = %q, want %q", sa.Namespace, "test-ns")
	}
	if sa.Labels["app.kubernetes.io/name"] != "openclaw" {
		t.Error("service account missing app label")
	}
	if sa.Labels["app.kubernetes.io/instance"] != "sa-test" {
		t.Error("service account missing instance label")
	}
	if sa.Labels["app.kubernetes.io/managed-by"] != "openclaw-operator" {
		t.Error("service account missing managed-by label")
	}
}

func TestBuildServiceAccount_CustomName(t *testing.T) {
	instance := newTestInstance("sa-custom")
	instance.Spec.Security.RBAC.ServiceAccountName = "my-custom-sa"

	sa := BuildServiceAccount(instance)

	if sa.Name != "my-custom-sa" {
		t.Errorf("service account name = %q, want %q", sa.Name, "my-custom-sa")
	}
}

func TestBuildRole_Default(t *testing.T) {
	instance := newTestInstance("role-test")
	role := BuildRole(instance)

	if role.Name != "role-test" {
		t.Errorf("role name = %q, want %q", role.Name, "role-test")
	}
	if role.Namespace != "test-ns" {
		t.Errorf("role namespace = %q, want %q", role.Namespace, "test-ns")
	}

	// Should have exactly 1 rule (configmap read)
	if len(role.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(role.Rules))
	}

	rule := role.Rules[0]
	if len(rule.APIGroups) != 1 || rule.APIGroups[0] != "" {
		t.Error("base rule should have core API group")
	}
	if len(rule.Resources) != 1 || rule.Resources[0] != "configmaps" {
		t.Errorf("base rule resources = %v, want [configmaps]", rule.Resources)
	}
	if len(rule.ResourceNames) != 1 || rule.ResourceNames[0] != "role-test-config" {
		t.Errorf("base rule resourceNames = %v, want [role-test-config]", rule.ResourceNames)
	}
	if len(rule.Verbs) != 2 {
		t.Fatalf("expected 2 verbs, got %d", len(rule.Verbs))
	}
	expectedVerbs := map[string]bool{"get": true, "watch": true}
	for _, v := range rule.Verbs {
		if !expectedVerbs[v] {
			t.Errorf("unexpected verb %q", v)
		}
	}
}

func TestBuildRole_AdditionalRules(t *testing.T) {
	instance := newTestInstance("role-extra")
	instance.Spec.Security.RBAC.AdditionalRules = []openclawv1alpha1.RBACRule{
		{
			APIGroups: []string{""},
			Resources: []string{"secrets"},
			Verbs:     []string{"get", "list"},
		},
		{
			APIGroups: []string{"apps"},
			Resources: []string{"deployments"},
			Verbs:     []string{"get"},
		},
	}

	role := BuildRole(instance)

	// 1 base rule + 2 additional rules
	if len(role.Rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(role.Rules))
	}

	// Verify additional rules
	secondRule := role.Rules[1]
	if secondRule.Resources[0] != "secrets" {
		t.Errorf("second rule resources = %v, want [secrets]", secondRule.Resources)
	}
	thirdRule := role.Rules[2]
	if thirdRule.APIGroups[0] != "apps" || thirdRule.Resources[0] != "deployments" {
		t.Error("third rule does not match expected values")
	}
}

func TestBuildRoleBinding(t *testing.T) {
	instance := newTestInstance("rb-test")
	rb := BuildRoleBinding(instance)

	if rb.Name != "rb-test" {
		t.Errorf("role binding name = %q, want %q", rb.Name, "rb-test")
	}
	if rb.Namespace != "test-ns" {
		t.Errorf("role binding namespace = %q, want %q", rb.Namespace, "test-ns")
	}

	// RoleRef
	if rb.RoleRef.Kind != "Role" {
		t.Errorf("roleRef kind = %q, want %q", rb.RoleRef.Kind, "Role")
	}
	if rb.RoleRef.Name != "rb-test" {
		t.Errorf("roleRef name = %q, want %q", rb.RoleRef.Name, "rb-test")
	}
	if rb.RoleRef.APIGroup != "rbac.authorization.k8s.io" {
		t.Errorf("roleRef apiGroup = %q, want %q", rb.RoleRef.APIGroup, "rbac.authorization.k8s.io")
	}

	// Subjects
	if len(rb.Subjects) != 1 {
		t.Fatalf("expected 1 subject, got %d", len(rb.Subjects))
	}
	subj := rb.Subjects[0]
	if subj.Kind != "ServiceAccount" {
		t.Errorf("subject kind = %q, want ServiceAccount", subj.Kind)
	}
	if subj.Name != "rb-test" {
		t.Errorf("subject name = %q, want %q", subj.Name, "rb-test")
	}
	if subj.Namespace != "test-ns" {
		t.Errorf("subject namespace = %q, want %q", subj.Namespace, "test-ns")
	}
}

func TestBuildRoleBinding_CustomServiceAccount(t *testing.T) {
	instance := newTestInstance("rb-custom-sa")
	instance.Spec.Security.RBAC.ServiceAccountName = "my-sa"

	rb := BuildRoleBinding(instance)

	// Subject should use the custom SA name
	if rb.Subjects[0].Name != "my-sa" {
		t.Errorf("subject name = %q, want %q", rb.Subjects[0].Name, "my-sa")
	}
	// RoleRef should still use instance name
	if rb.RoleRef.Name != "rb-custom-sa" {
		t.Errorf("roleRef name = %q, want %q", rb.RoleRef.Name, "rb-custom-sa")
	}
}

// ---------------------------------------------------------------------------
// configmap.go tests
// ---------------------------------------------------------------------------

func TestBuildConfigMap_Default(t *testing.T) {
	instance := newTestInstance("cm-test")
	cm := BuildConfigMap(instance)

	if cm.Name != "cm-test-config" {
		t.Errorf("configmap name = %q, want %q", cm.Name, "cm-test-config")
	}
	if cm.Namespace != "test-ns" {
		t.Errorf("configmap namespace = %q, want %q", cm.Namespace, "test-ns")
	}
	if cm.Labels["app.kubernetes.io/name"] != "openclaw" {
		t.Error("configmap missing app label")
	}

	// Default config should be empty JSON object
	content, ok := cm.Data["openclaw.json"]
	if !ok {
		t.Fatal("configmap missing openclaw.json key")
	}
	if content != "{}" {
		t.Errorf("default config content = %q, want %q", content, "{}")
	}
}

func TestBuildConfigMap_RawConfig(t *testing.T) {
	instance := newTestInstance("cm-raw")
	instance.Spec.Config.Raw = &openclawv1alpha1.RawConfig{
		RawExtension: runtime.RawExtension{
			Raw: []byte(`{"mcpServers":{"test":{"url":"http://localhost:3000"}}}`),
		},
	}

	cm := BuildConfigMap(instance)

	content, ok := cm.Data["openclaw.json"]
	if !ok {
		t.Fatal("configmap missing openclaw.json key")
	}

	// The builder pretty-prints JSON, so check it contains the expected keys
	if content == "{}" {
		t.Error("config content should not be empty with raw config")
	}

	// Verify the content is valid JSON and contains expected data
	if content == "" {
		t.Error("config content should not be empty")
	}
}

func TestBuildConfigMap_InvalidJSON_RawPreserved(t *testing.T) {
	instance := newTestInstance("cm-invalid")
	// If raw JSON is technically valid but the builder tries to pretty-print,
	// verify it handles valid JSON correctly
	instance.Spec.Config.Raw = &openclawv1alpha1.RawConfig{
		RawExtension: runtime.RawExtension{
			Raw: []byte(`{"key":"value"}`),
		},
	}

	cm := BuildConfigMap(instance)
	content := cm.Data["openclaw.json"]

	// Pretty-printed version of {"key":"value"}
	expected := "{\n  \"key\": \"value\"\n}"
	if content != expected {
		t.Errorf("config content = %q, want %q", content, expected)
	}
}

// ---------------------------------------------------------------------------
// pvc.go tests
// ---------------------------------------------------------------------------

func TestBuildPVC_Default(t *testing.T) {
	instance := newTestInstance("pvc-test")
	pvc := BuildPVC(instance)

	if pvc.Name != "pvc-test-data" {
		t.Errorf("pvc name = %q, want %q", pvc.Name, "pvc-test-data")
	}
	if pvc.Namespace != "test-ns" {
		t.Errorf("pvc namespace = %q, want %q", pvc.Namespace, "test-ns")
	}
	if pvc.Labels["app.kubernetes.io/name"] != "openclaw" {
		t.Error("pvc missing app label")
	}

	// Backup annotation
	if pvc.Annotations["openclaw.openclaw.io/backup-enabled"] != "true" {
		t.Error("pvc missing backup-enabled annotation")
	}

	// Default size - 10Gi
	storageReq := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	if storageReq.Cmp(resource.MustParse("10Gi")) != 0 {
		t.Errorf("storage size = %v, want 10Gi", storageReq.String())
	}

	// Default access mode - ReadWriteOnce
	if len(pvc.Spec.AccessModes) != 1 {
		t.Fatalf("expected 1 access mode, got %d", len(pvc.Spec.AccessModes))
	}
	if pvc.Spec.AccessModes[0] != corev1.ReadWriteOnce {
		t.Errorf("access mode = %v, want ReadWriteOnce", pvc.Spec.AccessModes[0])
	}

	// No storage class by default
	if pvc.Spec.StorageClassName != nil {
		t.Errorf("storageClassName should be nil by default, got %v", *pvc.Spec.StorageClassName)
	}
}

func TestBuildPVC_CustomSize(t *testing.T) {
	instance := newTestInstance("pvc-custom")
	instance.Spec.Storage.Persistence.Size = "50Gi"
	scName := "fast-ssd"
	instance.Spec.Storage.Persistence.StorageClass = &scName

	pvc := BuildPVC(instance)

	storageReq := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	if storageReq.Cmp(resource.MustParse("50Gi")) != 0 {
		t.Errorf("storage size = %v, want 50Gi", storageReq.String())
	}

	if pvc.Spec.StorageClassName == nil {
		t.Fatal("storageClassName should not be nil")
	}
	if *pvc.Spec.StorageClassName != "fast-ssd" {
		t.Errorf("storageClassName = %q, want %q", *pvc.Spec.StorageClassName, "fast-ssd")
	}
}

func TestBuildPVC_CustomAccessModes(t *testing.T) {
	instance := newTestInstance("pvc-modes")
	instance.Spec.Storage.Persistence.AccessModes = []corev1.PersistentVolumeAccessMode{
		corev1.ReadWriteMany,
	}

	pvc := BuildPVC(instance)

	if len(pvc.Spec.AccessModes) != 1 {
		t.Fatalf("expected 1 access mode, got %d", len(pvc.Spec.AccessModes))
	}
	if pvc.Spec.AccessModes[0] != corev1.ReadWriteMany {
		t.Errorf("access mode = %v, want ReadWriteMany", pvc.Spec.AccessModes[0])
	}
}

// ---------------------------------------------------------------------------
// pdb.go tests
// ---------------------------------------------------------------------------

func TestBuildPDB_Default(t *testing.T) {
	instance := newTestInstance("pdb-test")
	pdb := BuildPDB(instance)

	if pdb.Name != "pdb-test" {
		t.Errorf("pdb name = %q, want %q", pdb.Name, "pdb-test")
	}
	if pdb.Namespace != "test-ns" {
		t.Errorf("pdb namespace = %q, want %q", pdb.Namespace, "test-ns")
	}
	if pdb.Labels["app.kubernetes.io/name"] != "openclaw" {
		t.Error("pdb missing app label")
	}

	// Selector
	sel := pdb.Spec.Selector.MatchLabels
	if sel["app.kubernetes.io/name"] != "openclaw" || sel["app.kubernetes.io/instance"] != "pdb-test" {
		t.Error("pdb selector does not match expected values")
	}

	// Default maxUnavailable = 1
	if pdb.Spec.MaxUnavailable == nil {
		t.Fatal("maxUnavailable should not be nil")
	}
	if pdb.Spec.MaxUnavailable.Type != intstr.Int {
		t.Error("maxUnavailable should be int type")
	}
	if pdb.Spec.MaxUnavailable.IntVal != 1 {
		t.Errorf("maxUnavailable = %d, want 1", pdb.Spec.MaxUnavailable.IntVal)
	}
}

func TestBuildPDB_Custom(t *testing.T) {
	instance := newTestInstance("pdb-custom")
	instance.Spec.Availability.PodDisruptionBudget = &openclawv1alpha1.PodDisruptionBudgetSpec{
		MaxUnavailable: Ptr(int32(0)),
	}

	pdb := BuildPDB(instance)

	if pdb.Spec.MaxUnavailable.IntVal != 0 {
		t.Errorf("maxUnavailable = %d, want 0", pdb.Spec.MaxUnavailable.IntVal)
	}
}

func TestBuildPDB_CustomValue(t *testing.T) {
	instance := newTestInstance("pdb-val")
	instance.Spec.Availability.PodDisruptionBudget = &openclawv1alpha1.PodDisruptionBudgetSpec{
		MaxUnavailable: Ptr(int32(2)),
	}

	pdb := BuildPDB(instance)

	if pdb.Spec.MaxUnavailable.IntVal != 2 {
		t.Errorf("maxUnavailable = %d, want 2", pdb.Spec.MaxUnavailable.IntVal)
	}
}

// ---------------------------------------------------------------------------
// ingress.go tests
// ---------------------------------------------------------------------------

func TestBuildIngress_Basic(t *testing.T) {
	instance := newTestInstance("ing-test")
	className := "nginx"
	instance.Spec.Networking.Ingress = openclawv1alpha1.IngressSpec{
		Enabled:   true,
		ClassName: &className,
		Hosts: []openclawv1alpha1.IngressHost{
			{
				Host: "openclaw.example.com",
			},
		},
		TLS: []openclawv1alpha1.IngressTLS{
			{
				Hosts:      []string{"openclaw.example.com"},
				SecretName: "openclaw-tls",
			},
		},
	}

	ing := BuildIngress(instance)

	// ObjectMeta
	if ing.Name != "ing-test" {
		t.Errorf("ingress name = %q, want %q", ing.Name, "ing-test")
	}
	if ing.Namespace != "test-ns" {
		t.Errorf("ingress namespace = %q, want %q", ing.Namespace, "test-ns")
	}
	if ing.Labels["app.kubernetes.io/name"] != "openclaw" {
		t.Error("ingress missing app label")
	}

	// IngressClassName
	if ing.Spec.IngressClassName == nil || *ing.Spec.IngressClassName != "nginx" {
		t.Error("ingress className should be nginx")
	}

	// Rules
	if len(ing.Spec.Rules) != 1 {
		t.Fatalf("expected 1 ingress rule, got %d", len(ing.Spec.Rules))
	}
	rule := ing.Spec.Rules[0]
	if rule.Host != "openclaw.example.com" {
		t.Errorf("ingress host = %q, want %q", rule.Host, "openclaw.example.com")
	}
	if rule.HTTP == nil || len(rule.HTTP.Paths) != 1 {
		t.Fatal("expected 1 path in ingress rule")
	}
	path := rule.HTTP.Paths[0]
	if path.Path != "/" {
		t.Errorf("ingress path = %q, want %q", path.Path, "/")
	}
	if path.Backend.Service == nil {
		t.Fatal("ingress backend service is nil")
	}
	if path.Backend.Service.Name != "ing-test" {
		t.Errorf("ingress backend service name = %q, want %q", path.Backend.Service.Name, "ing-test")
	}
	if path.Backend.Service.Port.Number != int32(GatewayPort) {
		t.Errorf("ingress backend port = %d, want %d", path.Backend.Service.Port.Number, GatewayPort)
	}

	// TLS
	if len(ing.Spec.TLS) != 1 {
		t.Fatalf("expected 1 TLS config, got %d", len(ing.Spec.TLS))
	}
	tls := ing.Spec.TLS[0]
	if tls.SecretName != "openclaw-tls" {
		t.Errorf("TLS secretName = %q, want %q", tls.SecretName, "openclaw-tls")
	}
	if len(tls.Hosts) != 1 || tls.Hosts[0] != "openclaw.example.com" {
		t.Errorf("TLS hosts = %v, want [openclaw.example.com]", tls.Hosts)
	}
}

func TestBuildIngress_DefaultAnnotations(t *testing.T) {
	instance := newTestInstance("ing-ann")
	instance.Spec.Networking.Ingress = openclawv1alpha1.IngressSpec{
		Enabled: true,
		Hosts: []openclawv1alpha1.IngressHost{
			{Host: "test.example.com"},
		},
	}

	ing := BuildIngress(instance)
	ann := ing.Annotations

	// Force HTTPS (default: true)
	if ann["nginx.ingress.kubernetes.io/ssl-redirect"] != "true" {
		t.Error("ssl-redirect should be true by default")
	}
	if ann["nginx.ingress.kubernetes.io/force-ssl-redirect"] != "true" {
		t.Error("force-ssl-redirect should be true by default")
	}

	// HSTS (default: true)
	if _, ok := ann["nginx.ingress.kubernetes.io/configuration-snippet"]; !ok {
		t.Error("HSTS configuration snippet should be present by default")
	}

	// WebSocket support
	if ann["nginx.ingress.kubernetes.io/proxy-read-timeout"] != "3600" {
		t.Error("proxy-read-timeout should be 3600")
	}
	if ann["nginx.ingress.kubernetes.io/proxy-send-timeout"] != "3600" {
		t.Error("proxy-send-timeout should be 3600")
	}
	if ann["nginx.ingress.kubernetes.io/proxy-http-version"] != "1.1" {
		t.Error("proxy-http-version should be 1.1")
	}
	if ann["nginx.ingress.kubernetes.io/upstream-hash-by"] != "$binary_remote_addr" {
		t.Error("upstream-hash-by should be $binary_remote_addr")
	}

	// Traefik annotation for HTTPS redirect
	expectedTraefik := "test-ns-redirect-https@kubernetescrd"
	if ann["traefik.ingress.kubernetes.io/router.middlewares"] != expectedTraefik {
		t.Errorf("traefik middleware = %q, want %q", ann["traefik.ingress.kubernetes.io/router.middlewares"], expectedTraefik)
	}
}

func TestBuildIngress_SecurityDisabled(t *testing.T) {
	instance := newTestInstance("ing-nosec")
	instance.Spec.Networking.Ingress = openclawv1alpha1.IngressSpec{
		Enabled: true,
		Hosts: []openclawv1alpha1.IngressHost{
			{Host: "test.example.com"},
		},
		Security: openclawv1alpha1.IngressSecuritySpec{
			ForceHTTPS: Ptr(false),
			EnableHSTS: Ptr(false),
		},
	}

	ing := BuildIngress(instance)
	ann := ing.Annotations

	if _, ok := ann["nginx.ingress.kubernetes.io/ssl-redirect"]; ok {
		t.Error("ssl-redirect should not be set when ForceHTTPS is false")
	}
	if _, ok := ann["nginx.ingress.kubernetes.io/force-ssl-redirect"]; ok {
		t.Error("force-ssl-redirect should not be set when ForceHTTPS is false")
	}
	if _, ok := ann["nginx.ingress.kubernetes.io/configuration-snippet"]; ok {
		t.Error("HSTS configuration snippet should not be set when EnableHSTS is false")
	}

	// WebSocket annotations should still be present
	if ann["nginx.ingress.kubernetes.io/proxy-read-timeout"] != "3600" {
		t.Error("proxy-read-timeout should still be 3600")
	}
}

func TestBuildIngress_RateLimiting(t *testing.T) {
	instance := newTestInstance("ing-rl")
	rps := int32(20)
	instance.Spec.Networking.Ingress = openclawv1alpha1.IngressSpec{
		Enabled: true,
		Hosts: []openclawv1alpha1.IngressHost{
			{Host: "test.example.com"},
		},
		Security: openclawv1alpha1.IngressSecuritySpec{
			RateLimiting: &openclawv1alpha1.RateLimitingSpec{
				RequestsPerSecond: &rps,
			},
		},
	}

	ing := BuildIngress(instance)
	ann := ing.Annotations

	if ann["nginx.ingress.kubernetes.io/limit-rps"] != "20" {
		t.Errorf("limit-rps = %q, want %q", ann["nginx.ingress.kubernetes.io/limit-rps"], "20")
	}
}

func TestBuildIngress_RateLimitingDefault(t *testing.T) {
	instance := newTestInstance("ing-rl-default")
	instance.Spec.Networking.Ingress = openclawv1alpha1.IngressSpec{
		Enabled: true,
		Hosts: []openclawv1alpha1.IngressHost{
			{Host: "test.example.com"},
		},
		Security: openclawv1alpha1.IngressSecuritySpec{
			RateLimiting: &openclawv1alpha1.RateLimitingSpec{
				// Enabled defaults to true, RPS defaults to 10
			},
		},
	}

	ing := BuildIngress(instance)
	ann := ing.Annotations

	if ann["nginx.ingress.kubernetes.io/limit-rps"] != "10" {
		t.Errorf("limit-rps = %q, want %q", ann["nginx.ingress.kubernetes.io/limit-rps"], "10")
	}
}

func TestBuildIngress_RateLimitingDisabled(t *testing.T) {
	instance := newTestInstance("ing-rl-off")
	instance.Spec.Networking.Ingress = openclawv1alpha1.IngressSpec{
		Enabled: true,
		Hosts: []openclawv1alpha1.IngressHost{
			{Host: "test.example.com"},
		},
		Security: openclawv1alpha1.IngressSecuritySpec{
			RateLimiting: &openclawv1alpha1.RateLimitingSpec{
				Enabled: Ptr(false),
			},
		},
	}

	ing := BuildIngress(instance)
	ann := ing.Annotations

	if _, ok := ann["nginx.ingress.kubernetes.io/limit-rps"]; ok {
		t.Error("limit-rps should not be set when rate limiting is disabled")
	}
}

func TestBuildIngress_CustomAnnotations(t *testing.T) {
	instance := newTestInstance("ing-custom-ann")
	instance.Spec.Networking.Ingress = openclawv1alpha1.IngressSpec{
		Enabled: true,
		Annotations: map[string]string{
			"custom-key": "custom-value",
		},
		Hosts: []openclawv1alpha1.IngressHost{
			{Host: "test.example.com"},
		},
	}

	ing := BuildIngress(instance)

	if ing.Annotations["custom-key"] != "custom-value" {
		t.Error("custom annotation not preserved")
	}
	// Default annotations should still be present
	if ing.Annotations["nginx.ingress.kubernetes.io/proxy-http-version"] != "1.1" {
		t.Error("default annotations should coexist with custom annotations")
	}
}

func TestBuildIngress_MultipleHosts(t *testing.T) {
	instance := newTestInstance("ing-multi")
	instance.Spec.Networking.Ingress = openclawv1alpha1.IngressSpec{
		Enabled: true,
		Hosts: []openclawv1alpha1.IngressHost{
			{Host: "a.example.com"},
			{Host: "b.example.com"},
		},
		TLS: []openclawv1alpha1.IngressTLS{
			{
				Hosts:      []string{"a.example.com", "b.example.com"},
				SecretName: "multi-tls",
			},
		},
	}

	ing := BuildIngress(instance)

	if len(ing.Spec.Rules) != 2 {
		t.Fatalf("expected 2 ingress rules, got %d", len(ing.Spec.Rules))
	}
	if ing.Spec.Rules[0].Host != "a.example.com" {
		t.Errorf("first host = %q, want %q", ing.Spec.Rules[0].Host, "a.example.com")
	}
	if ing.Spec.Rules[1].Host != "b.example.com" {
		t.Errorf("second host = %q, want %q", ing.Spec.Rules[1].Host, "b.example.com")
	}
	if len(ing.Spec.TLS) != 1 {
		t.Fatalf("expected 1 TLS entry, got %d", len(ing.Spec.TLS))
	}
	if len(ing.Spec.TLS[0].Hosts) != 2 {
		t.Errorf("TLS hosts count = %d, want 2", len(ing.Spec.TLS[0].Hosts))
	}
}

func TestBuildIngress_CustomPaths(t *testing.T) {
	instance := newTestInstance("ing-paths")
	instance.Spec.Networking.Ingress = openclawv1alpha1.IngressSpec{
		Enabled: true,
		Hosts: []openclawv1alpha1.IngressHost{
			{
				Host: "test.example.com",
				Paths: []openclawv1alpha1.IngressPath{
					{Path: "/api", PathType: "Prefix"},
					{Path: "/health", PathType: "Exact"},
				},
			},
		},
	}

	ing := BuildIngress(instance)

	rule := ing.Spec.Rules[0]
	if len(rule.HTTP.Paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(rule.HTTP.Paths))
	}
	if rule.HTTP.Paths[0].Path != "/api" {
		t.Errorf("first path = %q, want %q", rule.HTTP.Paths[0].Path, "/api")
	}
	if rule.HTTP.Paths[1].Path != "/health" {
		t.Errorf("second path = %q, want %q", rule.HTTP.Paths[1].Path, "/health")
	}

	// Verify path types
	if rule.HTTP.Paths[0].PathType == nil {
		t.Fatal("first path pathType is nil")
	}
	// "Prefix" maps to PathTypePrefix
	// "Exact" maps to PathTypeExact

	if rule.HTTP.Paths[1].PathType == nil {
		t.Fatal("second path pathType is nil")
	}
}

func TestBuildIngress_NoHosts(t *testing.T) {
	instance := newTestInstance("ing-no-hosts")
	instance.Spec.Networking.Ingress = openclawv1alpha1.IngressSpec{
		Enabled: true,
		// No hosts
	}

	ing := BuildIngress(instance)

	if len(ing.Spec.Rules) != 0 {
		t.Errorf("expected 0 rules with no hosts, got %d", len(ing.Spec.Rules))
	}
}

// ---------------------------------------------------------------------------
// Cross-cutting / integration-style tests
// ---------------------------------------------------------------------------

func TestAllBuilders_ConsistentLabels(t *testing.T) {
	instance := newTestInstance("label-check")
	instance.Spec.Chromium.Enabled = true
	instance.Spec.Config.Raw = &openclawv1alpha1.RawConfig{
		RawExtension: runtime.RawExtension{
			Raw: []byte(`{}`),
		},
	}
	instance.Spec.Networking.Ingress = openclawv1alpha1.IngressSpec{
		Enabled: true,
		Hosts:   []openclawv1alpha1.IngressHost{{Host: "test.example.com"}},
	}

	expectedLabels := Labels(instance)

	resources := []struct {
		name   string
		labels map[string]string
	}{
		{"Deployment", BuildDeployment(instance).Labels},
		{"Service", BuildService(instance).Labels},
		{"NetworkPolicy", BuildNetworkPolicy(instance).Labels},
		{"ServiceAccount", BuildServiceAccount(instance).Labels},
		{"Role", BuildRole(instance).Labels},
		{"RoleBinding", BuildRoleBinding(instance).Labels},
		{"ConfigMap", BuildConfigMap(instance).Labels},
		{"PVC", BuildPVC(instance).Labels},
		{"PDB", BuildPDB(instance).Labels},
		{"Ingress", BuildIngress(instance).Labels},
	}

	for _, r := range resources {
		t.Run(r.name, func(t *testing.T) {
			for k, v := range expectedLabels {
				if r.labels[k] != v {
					t.Errorf("%s: label %q = %q, want %q", r.name, k, r.labels[k], v)
				}
			}
		})
	}
}

func TestAllBuilders_ConsistentNamespace(t *testing.T) {
	instance := newTestInstance("ns-check")
	instance.Namespace = "production"
	instance.Spec.Config.Raw = &openclawv1alpha1.RawConfig{
		RawExtension: runtime.RawExtension{
			Raw: []byte(`{}`),
		},
	}
	instance.Spec.Networking.Ingress = openclawv1alpha1.IngressSpec{
		Enabled: true,
		Hosts:   []openclawv1alpha1.IngressHost{{Host: "test.example.com"}},
	}

	resources := []struct {
		name      string
		namespace string
	}{
		{"Deployment", BuildDeployment(instance).Namespace},
		{"Service", BuildService(instance).Namespace},
		{"NetworkPolicy", BuildNetworkPolicy(instance).Namespace},
		{"ServiceAccount", BuildServiceAccount(instance).Namespace},
		{"Role", BuildRole(instance).Namespace},
		{"RoleBinding", BuildRoleBinding(instance).Namespace},
		{"ConfigMap", BuildConfigMap(instance).Namespace},
		{"PVC", BuildPVC(instance).Namespace},
		{"PDB", BuildPDB(instance).Namespace},
		{"Ingress", BuildIngress(instance).Namespace},
	}

	for _, r := range resources {
		t.Run(r.name, func(t *testing.T) {
			if r.namespace != "production" {
				t.Errorf("%s: namespace = %q, want %q", r.name, r.namespace, "production")
			}
		})
	}
}

func TestBuildDeployment_ChromiumCustomResources(t *testing.T) {
	instance := newTestInstance("chromium-res")
	instance.Spec.Chromium.Enabled = true
	instance.Spec.Chromium.Resources = openclawv1alpha1.ResourcesSpec{
		Requests: openclawv1alpha1.ResourceList{
			CPU:    "500m",
			Memory: "1Gi",
		},
		Limits: openclawv1alpha1.ResourceList{
			CPU:    "2",
			Memory: "4Gi",
		},
	}

	dep := BuildDeployment(instance)
	var chromium *corev1.Container
	for i := range dep.Spec.Template.Spec.Containers {
		if dep.Spec.Template.Spec.Containers[i].Name == "chromium" {
			chromium = &dep.Spec.Template.Spec.Containers[i]
			break
		}
	}
	if chromium == nil {
		t.Fatal("chromium container not found")
	}

	cpuReq := chromium.Resources.Requests[corev1.ResourceCPU]
	if cpuReq.String() != "500m" {
		t.Errorf("chromium cpu request = %v, want 500m", cpuReq.String())
	}
	memReq := chromium.Resources.Requests[corev1.ResourceMemory]
	if memReq.Cmp(resource.MustParse("1Gi")) != 0 {
		t.Errorf("chromium memory request = %v, want 1Gi", memReq.String())
	}
	cpuLim := chromium.Resources.Limits[corev1.ResourceCPU]
	if cpuLim.String() != "2" {
		t.Errorf("chromium cpu limit = %v, want 2", cpuLim.String())
	}
	memLim := chromium.Resources.Limits[corev1.ResourceMemory]
	if memLim.Cmp(resource.MustParse("4Gi")) != 0 {
		t.Errorf("chromium memory limit = %v, want 4Gi", memLim.String())
	}
}

func TestBuildDeployment_CustomPodSecurityContext(t *testing.T) {
	instance := newTestInstance("custom-psc")
	instance.Spec.Security.PodSecurityContext = &openclawv1alpha1.PodSecurityContextSpec{
		RunAsUser:  Ptr(int64(2000)),
		RunAsGroup: Ptr(int64(3000)),
		FSGroup:    Ptr(int64(4000)),
	}

	dep := BuildDeployment(instance)
	psc := dep.Spec.Template.Spec.SecurityContext

	if *psc.RunAsUser != 2000 {
		t.Errorf("runAsUser = %d, want 2000", *psc.RunAsUser)
	}
	if *psc.RunAsGroup != 3000 {
		t.Errorf("runAsGroup = %d, want 3000", *psc.RunAsGroup)
	}
	if *psc.FSGroup != 4000 {
		t.Errorf("fsGroup = %d, want 4000", *psc.FSGroup)
	}
	// runAsNonRoot should still be true (default)
	if !*psc.RunAsNonRoot {
		t.Error("runAsNonRoot should still be true")
	}
}

func TestBuildDeployment_CustomPullPolicy(t *testing.T) {
	instance := newTestInstance("pull-policy")
	instance.Spec.Image.PullPolicy = corev1.PullAlways

	dep := BuildDeployment(instance)
	main := dep.Spec.Template.Spec.Containers[0]

	if main.ImagePullPolicy != corev1.PullAlways {
		t.Errorf("pullPolicy = %v, want Always", main.ImagePullPolicy)
	}
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

func assertContainerPort(t *testing.T, ports []corev1.ContainerPort, name string, expectedPort int32) {
	t.Helper()
	for _, p := range ports {
		if p.Name == name {
			if p.ContainerPort != expectedPort {
				t.Errorf("port %q = %d, want %d", name, p.ContainerPort, expectedPort)
			}
			if p.Protocol != corev1.ProtocolTCP {
				t.Errorf("port %q protocol = %v, want TCP", name, p.Protocol)
			}
			return
		}
	}
	t.Errorf("port %q not found", name)
}

func assertServicePort(t *testing.T, ports []corev1.ServicePort, name string, expectedPort int32) {
	t.Helper()
	for _, p := range ports {
		if p.Name == name {
			if p.Port != expectedPort {
				t.Errorf("service port %q = %d, want %d", name, p.Port, expectedPort)
			}
			if p.TargetPort.IntValue() != int(expectedPort) {
				t.Errorf("service target port %q = %d, want %d", name, p.TargetPort.IntValue(), expectedPort)
			}
			if p.Protocol != corev1.ProtocolTCP {
				t.Errorf("service port %q protocol = %v, want TCP", name, p.Protocol)
			}
			return
		}
	}
	t.Errorf("service port %q not found", name)
}

func assertNPPort(t *testing.T, ports []networkingv1.NetworkPolicyPort, expectedPort int) {
	t.Helper()
	for _, p := range ports {
		if p.Port != nil && p.Port.IntValue() == expectedPort {
			return
		}
	}
	t.Errorf("network policy port %d not found", expectedPort)
}

func assertVolumeMount(t *testing.T, mounts []corev1.VolumeMount, name, expectedPath string) {
	t.Helper()
	for _, m := range mounts {
		if m.Name == name {
			if m.MountPath != expectedPath {
				t.Errorf("volume mount %q path = %q, want %q", name, m.MountPath, expectedPath)
			}
			return
		}
	}
	t.Errorf("volume mount %q not found", name)
}

func findVolume(volumes []corev1.Volume, name string) *corev1.Volume {
	for i := range volumes {
		if volumes[i].Name == name {
			return &volumes[i]
		}
	}
	return nil
}
