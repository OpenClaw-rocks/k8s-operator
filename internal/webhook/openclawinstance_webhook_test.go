/*
Copyright 2024 OpenClaw.

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

package webhook

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openclawv1alpha1 "github.com/openclawrocks/k8s-operator/api/v1alpha1"
)

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}

// newTestInstance returns a well-configured OpenClawInstance that passes
// validation with zero warnings and zero errors. Individual tests mutate
// this baseline to trigger specific validation paths.
func newTestInstance() *openclawv1alpha1.OpenClawInstance {
	return &openclawv1alpha1.OpenClawInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: openclawv1alpha1.OpenClawInstanceSpec{
			EnvFrom: []corev1.EnvFromSource{
				{SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: "test-secret"},
				}},
			},
			Resources: openclawv1alpha1.ResourcesSpec{
				Limits: openclawv1alpha1.ResourceList{
					CPU:    "2",
					Memory: "4Gi",
				},
			},
		},
	}
}

// containsWarning returns true if any warning message contains the substring.
func containsWarning(warnings []string, substring string) bool {
	for _, w := range warnings {
		if strings.Contains(w, substring) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// ValidateCreate tests
// ---------------------------------------------------------------------------

func TestValidateCreate_ValidInstance(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error for a valid instance, got: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected 0 warnings for a valid instance, got %d: %v", len(warnings), warnings)
	}
}

func TestValidateCreate_BlocksRootUser(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Security.PodSecurityContext = &openclawv1alpha1.PodSecurityContextSpec{
		RunAsUser: ptr(int64(0)),
	}

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err == nil {
		t.Fatal("expected error when RunAsUser=0, got nil")
	}
	if !strings.Contains(err.Error(), "root") && !strings.Contains(err.Error(), "UID 0") {
		t.Fatalf("error message should mention root or UID 0, got: %v", err)
	}
	// When an error is returned, warnings should be nil.
	if warnings != nil {
		t.Fatalf("expected nil warnings on hard error, got: %v", warnings)
	}
}

func TestValidateCreate_AllowsNonRootUser(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Security.PodSecurityContext = &openclawv1alpha1.PodSecurityContextSpec{
		RunAsUser: ptr(int64(1000)),
	}

	_, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error for RunAsUser=1000, got: %v", err)
	}
}

func TestValidateCreate_WarnsRunAsNonRootFalse(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Security.PodSecurityContext = &openclawv1alpha1.PodSecurityContextSpec{
		RunAsNonRoot: ptr(false),
	}

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !containsWarning(warnings, "runAsNonRoot") {
		t.Fatalf("expected warning about runAsNonRoot, got: %v", warnings)
	}
}

func TestValidateCreate_NoWarnRunAsNonRootTrue(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Security.PodSecurityContext = &openclawv1alpha1.PodSecurityContextSpec{
		RunAsNonRoot: ptr(true),
	}

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if containsWarning(warnings, "runAsNonRoot") {
		t.Fatalf("expected no runAsNonRoot warning when set to true, got: %v", warnings)
	}
}

func TestValidateCreate_WarnsNetworkPolicyDisabled(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Security.NetworkPolicy.Enabled = ptr(false)

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !containsWarning(warnings, "NetworkPolicy") {
		t.Fatalf("expected warning about NetworkPolicy, got: %v", warnings)
	}
}

func TestValidateCreate_NoWarnNetworkPolicyEnabled(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Security.NetworkPolicy.Enabled = ptr(true)

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if containsWarning(warnings, "NetworkPolicy") {
		t.Fatalf("expected no NetworkPolicy warning when enabled, got: %v", warnings)
	}
}

func TestValidateCreate_WarnsIngressWithoutTLS(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Networking.Ingress.Enabled = true
	// TLS is empty by default.

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !containsWarning(warnings, "TLS") {
		t.Fatalf("expected warning about TLS, got: %v", warnings)
	}
}

func TestValidateCreate_NoWarnIngressWithTLS(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Networking.Ingress.Enabled = true
	instance.Spec.Networking.Ingress.TLS = []openclawv1alpha1.IngressTLS{
		{Hosts: []string{"example.com"}, SecretName: "tls-secret"},
	}

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if containsWarning(warnings, "TLS") {
		t.Fatalf("expected no TLS warning when TLS is configured, got: %v", warnings)
	}
}

func TestValidateCreate_NoWarnIngressDisabled(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	// Ingress.Enabled defaults to false. Even without TLS, no warning.

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if containsWarning(warnings, "TLS") {
		t.Fatalf("expected no TLS warning when Ingress is disabled, got: %v", warnings)
	}
}

func TestValidateCreate_WarnsIngressForceHTTPSDisabled(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Networking.Ingress.Enabled = true
	instance.Spec.Networking.Ingress.TLS = []openclawv1alpha1.IngressTLS{
		{Hosts: []string{"example.com"}, SecretName: "tls-secret"},
	}
	instance.Spec.Networking.Ingress.Security.ForceHTTPS = ptr(false)

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !containsWarning(warnings, "forceHTTPS") {
		t.Fatalf("expected warning about forceHTTPS, got: %v", warnings)
	}
}

func TestValidateCreate_NoWarnForceHTTPSWhenIngressDisabled(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	// Ingress disabled; forceHTTPS=false should NOT trigger a warning.
	instance.Spec.Networking.Ingress.Security.ForceHTTPS = ptr(false)

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if containsWarning(warnings, "forceHTTPS") {
		t.Fatalf("expected no forceHTTPS warning when Ingress is disabled, got: %v", warnings)
	}
}

func TestValidateCreate_WarnsChromiumWithoutDigest(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Chromium.Enabled = true
	instance.Spec.Chromium.Image.Digest = "" // no digest

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !containsWarning(warnings, "Chromium") || !containsWarning(warnings, "digest") {
		t.Fatalf("expected warning about Chromium digest pinning, got: %v", warnings)
	}
}

func TestValidateCreate_NoWarnChromiumWithDigest(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Chromium.Enabled = true
	instance.Spec.Chromium.Image.Digest = "sha256:abc123"

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if containsWarning(warnings, "Chromium") {
		t.Fatalf("expected no Chromium warning when digest is set, got: %v", warnings)
	}
}

func TestValidateCreate_NoWarnChromiumDisabled(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	// Chromium.Enabled defaults to false.

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if containsWarning(warnings, "Chromium") {
		t.Fatalf("expected no Chromium warning when disabled, got: %v", warnings)
	}
}

func TestValidateCreate_WarnsNoEnvVars(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.EnvFrom = nil
	instance.Spec.Env = nil

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !containsWarning(warnings, "environment variables") {
		t.Fatalf("expected warning about missing environment variables, got: %v", warnings)
	}
}

func TestValidateCreate_NoWarnWithEnvFrom(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	// newTestInstance already has EnvFrom configured.

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if containsWarning(warnings, "environment variables") {
		t.Fatalf("expected no env warning when envFrom is set, got: %v", warnings)
	}
}

func TestValidateCreate_NoWarnWithEnv(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.EnvFrom = nil
	instance.Spec.Env = []corev1.EnvVar{
		{Name: "ANTHROPIC_API_KEY", Value: "sk-test"},
	}

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if containsWarning(warnings, "environment variables") {
		t.Fatalf("expected no env warning when env is set, got: %v", warnings)
	}
}

func TestValidateCreate_WarnsPrivilegeEscalation(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Security.ContainerSecurityContext = &openclawv1alpha1.ContainerSecurityContextSpec{
		AllowPrivilegeEscalation: ptr(true),
	}

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !containsWarning(warnings, "allowPrivilegeEscalation") {
		t.Fatalf("expected warning about allowPrivilegeEscalation, got: %v", warnings)
	}
}

func TestValidateCreate_NoWarnPrivilegeEscalationFalse(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Security.ContainerSecurityContext = &openclawv1alpha1.ContainerSecurityContextSpec{
		AllowPrivilegeEscalation: ptr(false),
	}

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if containsWarning(warnings, "allowPrivilegeEscalation") {
		t.Fatalf("expected no privilege escalation warning when false, got: %v", warnings)
	}
}

func TestValidateCreate_WarnsNoResourceLimits(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Resources.Limits = openclawv1alpha1.ResourceList{} // empty

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !containsWarning(warnings, "Resource limits") {
		t.Fatalf("expected warning about resource limits, got: %v", warnings)
	}
}

func TestValidateCreate_WarnsPartialResourceLimits_MissingCPU(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Resources.Limits = openclawv1alpha1.ResourceList{
		Memory: "4Gi",
	}

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !containsWarning(warnings, "Resource limits") {
		t.Fatalf("expected warning about resource limits when CPU missing, got: %v", warnings)
	}
}

func TestValidateCreate_WarnsPartialResourceLimits_MissingMemory(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Resources.Limits = openclawv1alpha1.ResourceList{
		CPU: "2",
	}

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !containsWarning(warnings, "Resource limits") {
		t.Fatalf("expected warning about resource limits when memory missing, got: %v", warnings)
	}
}

func TestValidateCreate_MultipleWarnings(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()

	// Trigger multiple warnings at once:
	// 1. runAsNonRoot=false
	instance.Spec.Security.PodSecurityContext = &openclawv1alpha1.PodSecurityContextSpec{
		RunAsNonRoot: ptr(false),
	}
	// 2. NetworkPolicy disabled
	instance.Spec.Security.NetworkPolicy.Enabled = ptr(false)
	// 3. Ingress without TLS
	instance.Spec.Networking.Ingress.Enabled = true
	// 4. forceHTTPS disabled
	instance.Spec.Networking.Ingress.Security.ForceHTTPS = ptr(false)
	// 5. Chromium without digest
	instance.Spec.Chromium.Enabled = true
	// 6. No env vars
	instance.Spec.EnvFrom = nil
	instance.Spec.Env = nil
	// 7. AllowPrivilegeEscalation=true
	instance.Spec.Security.ContainerSecurityContext = &openclawv1alpha1.ContainerSecurityContextSpec{
		AllowPrivilegeEscalation: ptr(true),
	}
	// 8. No resource limits
	instance.Spec.Resources.Limits = openclawv1alpha1.ResourceList{}

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error (only warnings), got: %v", err)
	}

	expectedCount := 8
	if len(warnings) != expectedCount {
		t.Fatalf("expected %d warnings, got %d: %v", expectedCount, len(warnings), warnings)
	}

	// Verify each expected warning substring is present.
	expectedSubstrings := []string{
		"runAsNonRoot",
		"NetworkPolicy",
		"TLS",
		"forceHTTPS",
		"Chromium",
		"environment variables",
		"allowPrivilegeEscalation",
		"Resource limits",
	}
	for _, sub := range expectedSubstrings {
		if !containsWarning(warnings, sub) {
			t.Errorf("expected a warning containing %q, but none found in: %v", sub, warnings)
		}
	}
}

// ---------------------------------------------------------------------------
// ValidateUpdate tests
// ---------------------------------------------------------------------------

func TestValidateUpdate_ImmutableStorageClass(t *testing.T) {
	v := &OpenClawInstanceValidator{}

	oldInstance := newTestInstance()
	oldInstance.Spec.Storage.Persistence.StorageClass = ptr("standard")

	newInstance := newTestInstance()
	newInstance.Spec.Storage.Persistence.StorageClass = ptr("premium-ssd")

	warnings, err := v.ValidateUpdate(context.Background(), oldInstance, newInstance)
	if err == nil {
		t.Fatal("expected error when changing storage class, got nil")
	}
	if !strings.Contains(err.Error(), "immutable") {
		t.Fatalf("error should mention immutability, got: %v", err)
	}
	if warnings != nil {
		t.Fatalf("expected nil warnings on hard error, got: %v", warnings)
	}
}

func TestValidateUpdate_AllowsSameStorageClass(t *testing.T) {
	v := &OpenClawInstanceValidator{}

	oldInstance := newTestInstance()
	oldInstance.Spec.Storage.Persistence.StorageClass = ptr("standard")

	newInstance := newTestInstance()
	newInstance.Spec.Storage.Persistence.StorageClass = ptr("standard")

	warnings, err := v.ValidateUpdate(context.Background(), oldInstance, newInstance)
	if err != nil {
		t.Fatalf("expected no error when storage class is unchanged, got: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected 0 warnings, got %d: %v", len(warnings), warnings)
	}
}

func TestValidateUpdate_AllowsStorageClassWhenOldIsNil(t *testing.T) {
	v := &OpenClawInstanceValidator{}

	oldInstance := newTestInstance()
	// StorageClass is nil by default.

	newInstance := newTestInstance()
	newInstance.Spec.Storage.Persistence.StorageClass = ptr("standard")

	_, err := v.ValidateUpdate(context.Background(), oldInstance, newInstance)
	if err != nil {
		t.Fatalf("expected no error when old storage class is nil, got: %v", err)
	}
}

func TestValidateUpdate_AllowsStorageClassWhenNewIsNil(t *testing.T) {
	v := &OpenClawInstanceValidator{}

	oldInstance := newTestInstance()
	oldInstance.Spec.Storage.Persistence.StorageClass = ptr("standard")

	newInstance := newTestInstance()
	// StorageClass is nil by default.

	_, err := v.ValidateUpdate(context.Background(), oldInstance, newInstance)
	if err != nil {
		t.Fatalf("expected no error when new storage class is nil, got: %v", err)
	}
}

func TestValidateUpdate_AllowsOtherChanges(t *testing.T) {
	v := &OpenClawInstanceValidator{}

	oldInstance := newTestInstance()
	oldInstance.Spec.Image.Tag = "v1.0.0"

	newInstance := newTestInstance()
	newInstance.Spec.Image.Tag = "v2.0.0"

	warnings, err := v.ValidateUpdate(context.Background(), oldInstance, newInstance)
	if err != nil {
		t.Fatalf("expected no error when changing image tag, got: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected 0 warnings, got %d: %v", len(warnings), warnings)
	}
}

func TestValidateUpdate_RunsValidationAfterImmutabilityCheck(t *testing.T) {
	v := &OpenClawInstanceValidator{}

	oldInstance := newTestInstance()
	newInstance := newTestInstance()
	// Trigger a validation warning (root user would be blocked).
	newInstance.Spec.Security.PodSecurityContext = &openclawv1alpha1.PodSecurityContextSpec{
		RunAsUser: ptr(int64(0)),
	}

	_, err := v.ValidateUpdate(context.Background(), oldInstance, newInstance)
	if err == nil {
		t.Fatal("expected error for RunAsUser=0 in update, got nil")
	}
	if !strings.Contains(err.Error(), "root") && !strings.Contains(err.Error(), "UID 0") {
		t.Fatalf("error should mention root/UID 0, got: %v", err)
	}
}

func TestValidateUpdate_ReturnsWarningsFromValidation(t *testing.T) {
	v := &OpenClawInstanceValidator{}

	oldInstance := newTestInstance()
	newInstance := newTestInstance()
	// Remove env vars to trigger a warning during update validation.
	newInstance.Spec.EnvFrom = nil
	newInstance.Spec.Env = nil

	warnings, err := v.ValidateUpdate(context.Background(), oldInstance, newInstance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !containsWarning(warnings, "environment variables") {
		t.Fatalf("expected warning about environment variables from update validation, got: %v", warnings)
	}
}

// ---------------------------------------------------------------------------
// ValidateDelete tests
// ---------------------------------------------------------------------------

func TestValidateDelete_AlwaysAllowed(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()

	warnings, err := v.ValidateDelete(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error on delete, got: %v", err)
	}
	if warnings != nil {
		t.Fatalf("expected nil warnings on delete, got: %v", warnings)
	}
}

func TestValidateDelete_AllowsEvenWithInvalidSpec(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	// Instance that would fail create/update validation.
	instance := newTestInstance()
	instance.Spec.Security.PodSecurityContext = &openclawv1alpha1.PodSecurityContextSpec{
		RunAsUser: ptr(int64(0)),
	}

	warnings, err := v.ValidateDelete(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error on delete even with invalid spec, got: %v", err)
	}
	if warnings != nil {
		t.Fatalf("expected nil warnings on delete, got: %v", warnings)
	}
}

// ---------------------------------------------------------------------------
// Edge case tests
// ---------------------------------------------------------------------------

func TestValidateCreate_NilPodSecurityContext(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Security.PodSecurityContext = nil

	_, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error with nil PodSecurityContext, got: %v", err)
	}
}

func TestValidateCreate_NilContainerSecurityContext(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Security.ContainerSecurityContext = nil

	_, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error with nil ContainerSecurityContext, got: %v", err)
	}
}

func TestValidateCreate_NilNetworkPolicyEnabled(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Security.NetworkPolicy.Enabled = nil

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error with nil NetworkPolicy.Enabled, got: %v", err)
	}
	if containsWarning(warnings, "NetworkPolicy") {
		t.Fatalf("expected no NetworkPolicy warning when Enabled is nil, got: %v", warnings)
	}
}

func TestValidateCreate_NilForceHTTPS(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Networking.Ingress.Enabled = true
	instance.Spec.Networking.Ingress.TLS = []openclawv1alpha1.IngressTLS{
		{Hosts: []string{"example.com"}, SecretName: "tls-secret"},
	}
	instance.Spec.Networking.Ingress.Security.ForceHTTPS = nil

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if containsWarning(warnings, "forceHTTPS") {
		t.Fatalf("expected no forceHTTPS warning when nil, got: %v", warnings)
	}
}

func TestValidateCreate_NilRunAsUser(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Security.PodSecurityContext = &openclawv1alpha1.PodSecurityContextSpec{
		RunAsUser: nil,
	}

	_, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error with nil RunAsUser, got: %v", err)
	}
}

func TestValidateCreate_NilAllowPrivilegeEscalation(t *testing.T) {
	v := &OpenClawInstanceValidator{}
	instance := newTestInstance()
	instance.Spec.Security.ContainerSecurityContext = &openclawv1alpha1.ContainerSecurityContextSpec{
		AllowPrivilegeEscalation: nil,
	}

	warnings, err := v.ValidateCreate(context.Background(), instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if containsWarning(warnings, "allowPrivilegeEscalation") {
		t.Fatalf("expected no privilege escalation warning when nil, got: %v", warnings)
	}
}
