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

package webhook

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	openclawv1alpha1 "github.com/openclawrocks/k8s-operator/api/v1alpha1"
)

// OpenClawInstanceValidator validates OpenClawInstance resources
type OpenClawInstanceValidator struct{}

var _ webhook.CustomValidator = &OpenClawInstanceValidator{}

// SetupWebhookWithManager sets up the webhook with the manager
func SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&openclawv1alpha1.OpenClawInstance{}).
		WithDefaulter(&OpenClawInstanceDefaulter{}).
		WithValidator(&OpenClawInstanceValidator{}).
		Complete()
}

// ValidateCreate implements webhook.CustomValidator
func (v *OpenClawInstanceValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	instance := obj.(*openclawv1alpha1.OpenClawInstance)
	return v.validate(instance)
}

// ValidateUpdate implements webhook.CustomValidator
func (v *OpenClawInstanceValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	instance := newObj.(*openclawv1alpha1.OpenClawInstance)
	oldInstance := oldObj.(*openclawv1alpha1.OpenClawInstance)

	// Check immutable fields
	if oldInstance.Spec.Storage.Persistence.StorageClass != nil &&
		instance.Spec.Storage.Persistence.StorageClass != nil &&
		*oldInstance.Spec.Storage.Persistence.StorageClass != *instance.Spec.Storage.Persistence.StorageClass {
		return nil, fmt.Errorf("storage class is immutable after creation")
	}

	return v.validate(instance)
}

// ValidateDelete implements webhook.CustomValidator
func (v *OpenClawInstanceValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

// validate performs the actual validation logic
func (v *OpenClawInstanceValidator) validate(instance *openclawv1alpha1.OpenClawInstance) (admission.Warnings, error) {
	var warnings admission.Warnings

	// 1. Block running as root (UID 0)
	if instance.Spec.Security.PodSecurityContext != nil &&
		instance.Spec.Security.PodSecurityContext.RunAsUser != nil &&
		*instance.Spec.Security.PodSecurityContext.RunAsUser == 0 {
		return nil, fmt.Errorf("running as root (UID 0) is not allowed for security reasons")
	}

	// 2. Warn if runAsNonRoot is explicitly set to false
	if instance.Spec.Security.PodSecurityContext != nil &&
		instance.Spec.Security.PodSecurityContext.RunAsNonRoot != nil &&
		!*instance.Spec.Security.PodSecurityContext.RunAsNonRoot {
		warnings = append(warnings, "runAsNonRoot is set to false - this allows running as root which is a security risk")
	}

	// 3. Warn if NetworkPolicy is disabled
	if instance.Spec.Security.NetworkPolicy.Enabled != nil &&
		!*instance.Spec.Security.NetworkPolicy.Enabled {
		warnings = append(warnings, "NetworkPolicy is disabled - pods will have unrestricted network access")
	}

	// 4. Warn if Ingress is enabled without TLS
	if instance.Spec.Networking.Ingress.Enabled {
		if len(instance.Spec.Networking.Ingress.TLS) == 0 {
			warnings = append(warnings, "Ingress is enabled without TLS - traffic will not be encrypted")
		}

		// Warn if forceHTTPS is disabled
		if instance.Spec.Networking.Ingress.Security.ForceHTTPS != nil &&
			!*instance.Spec.Networking.Ingress.Security.ForceHTTPS {
			warnings = append(warnings, "Ingress forceHTTPS is disabled - consider enabling for security")
		}
	}

	// 5. Warn if Chromium is enabled without digest pinning
	if instance.Spec.Chromium.Enabled {
		if instance.Spec.Chromium.Image.Digest == "" {
			warnings = append(warnings, "Chromium sidecar is enabled without image digest pinning - consider pinning to a specific digest for supply chain security")
		}
	}

	// 6. Warn if no envFrom is configured (likely missing API keys)
	if len(instance.Spec.EnvFrom) == 0 && len(instance.Spec.Env) == 0 {
		warnings = append(warnings, "No environment variables configured - you likely need to configure API keys via envFrom or env")
	}

	// 7. Warn if privilege escalation is allowed
	if instance.Spec.Security.ContainerSecurityContext != nil &&
		instance.Spec.Security.ContainerSecurityContext.AllowPrivilegeEscalation != nil &&
		*instance.Spec.Security.ContainerSecurityContext.AllowPrivilegeEscalation {
		warnings = append(warnings, "allowPrivilegeEscalation is enabled - this is a security risk")
	}

	// 8. Validate resource limits are set (recommended)
	if instance.Spec.Resources.Limits.CPU == "" || instance.Spec.Resources.Limits.Memory == "" {
		warnings = append(warnings, "Resource limits are not fully configured - consider setting both CPU and memory limits")
	}

	// 9. Warn if using "latest" image tag without a digest pin
	if instance.Spec.Image.Tag == "latest" && instance.Spec.Image.Digest == "" {
		warnings = append(warnings, "Image tag \"latest\" is mutable and not recommended for production - consider pinning to a specific version or digest")
	}

	// 10. Validate workspace spec
	if instance.Spec.Workspace != nil {
		if err := validateWorkspaceSpec(instance.Spec.Workspace); err != nil {
			return nil, err
		}
	}

	return warnings, nil
}

// validateWorkspaceSpec validates workspace file and directory names.
func validateWorkspaceSpec(ws *openclawv1alpha1.WorkspaceSpec) error {
	for name := range ws.InitialFiles {
		if err := validateWorkspaceFilename(name); err != nil {
			return fmt.Errorf("workspace initialFiles key %q: %w", name, err)
		}
	}
	for _, dir := range ws.InitialDirectories {
		if err := validateWorkspaceDirectory(dir); err != nil {
			return fmt.Errorf("workspace initialDirectories entry %q: %w", dir, err)
		}
	}
	return nil
}

// validateWorkspaceFilename checks a single workspace filename.
func validateWorkspaceFilename(name string) error {
	if name == "" {
		return fmt.Errorf("filename must not be empty")
	}
	if len(name) > 253 {
		return fmt.Errorf("filename must be at most 253 characters")
	}
	if strings.Contains(name, "/") {
		return fmt.Errorf("filename must not contain '/'")
	}
	if strings.Contains(name, "\\") {
		return fmt.Errorf("filename must not contain '\\'")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("filename must not contain '..'")
	}
	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("filename must not start with '.'")
	}
	if name == "openclaw.json" {
		return fmt.Errorf("filename 'openclaw.json' is reserved for config")
	}
	return nil
}

// validateWorkspaceDirectory checks a single workspace directory path.
func validateWorkspaceDirectory(dir string) error {
	if dir == "" {
		return fmt.Errorf("directory must not be empty")
	}
	if len(dir) > 253 {
		return fmt.Errorf("directory must be at most 253 characters")
	}
	if strings.Contains(dir, "\\") {
		return fmt.Errorf("directory must not contain '\\'")
	}
	if strings.Contains(dir, "..") {
		return fmt.Errorf("directory must not contain '..'")
	}
	if strings.HasPrefix(dir, "/") {
		return fmt.Errorf("directory must not be an absolute path")
	}
	return nil
}

// OpenClawInstanceDefaulter sets defaults for OpenClawInstance resources
type OpenClawInstanceDefaulter struct{}

var _ webhook.CustomDefaulter = &OpenClawInstanceDefaulter{}

// Default implements webhook.CustomDefaulter
func (d *OpenClawInstanceDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	instance := obj.(*openclawv1alpha1.OpenClawInstance)

	// Default image settings
	if instance.Spec.Image.Repository == "" {
		instance.Spec.Image.Repository = "ghcr.io/openclaw/openclaw"
	}
	if instance.Spec.Image.Tag == "" && instance.Spec.Image.Digest == "" {
		instance.Spec.Image.Tag = "latest"
	}
	if instance.Spec.Image.PullPolicy == "" {
		instance.Spec.Image.PullPolicy = corev1.PullIfNotPresent
	}

	// Default security settings
	if instance.Spec.Security.PodSecurityContext == nil {
		instance.Spec.Security.PodSecurityContext = &openclawv1alpha1.PodSecurityContextSpec{
			RunAsUser:    int64Ptr(1000),
			RunAsGroup:   int64Ptr(1000),
			FSGroup:      int64Ptr(1000),
			RunAsNonRoot: boolPtr(true),
		}
	}
	if instance.Spec.Security.ContainerSecurityContext == nil {
		instance.Spec.Security.ContainerSecurityContext = &openclawv1alpha1.ContainerSecurityContextSpec{
			AllowPrivilegeEscalation: boolPtr(false),
			ReadOnlyRootFilesystem:   boolPtr(false),
		}
	}

	// Default resource limits if not set
	if instance.Spec.Resources.Requests.CPU == "" {
		instance.Spec.Resources.Requests.CPU = "500m"
	}
	if instance.Spec.Resources.Requests.Memory == "" {
		instance.Spec.Resources.Requests.Memory = "1Gi"
	}
	if instance.Spec.Resources.Limits.CPU == "" {
		instance.Spec.Resources.Limits.CPU = "2000m"
	}
	if instance.Spec.Resources.Limits.Memory == "" {
		instance.Spec.Resources.Limits.Memory = "4Gi"
	}

	// Default storage
	if instance.Spec.Storage.Persistence.Enabled == nil {
		instance.Spec.Storage.Persistence.Enabled = boolPtr(true)
	}
	if instance.Spec.Storage.Persistence.Size == "" {
		instance.Spec.Storage.Persistence.Size = "10Gi"
	}

	// Default networking
	if instance.Spec.Networking.Service.Type == "" {
		instance.Spec.Networking.Service.Type = corev1.ServiceTypeClusterIP
	}

	return nil
}

func boolPtr(b bool) *bool {
	return &b
}

func int64Ptr(i int64) *int64 {
	return &i
}
