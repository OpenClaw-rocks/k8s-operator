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
	openclawv1alpha1 "github.com/openclawrocks/k8s-operator/api/v1alpha1"
)

const (
	// GatewayPort is the port for the OpenClaw gateway WebSocket server
	GatewayPort = 18789

	// CanvasPort is the port for the OpenClaw canvas HTTP server
	CanvasPort = 18793

	// ChromiumPort is the port for Chrome DevTools Protocol
	ChromiumPort = 9222

	// AppName is the application name used in labels
	AppName = "openclaw"

	// ComponentLabel is the component label key
	ComponentLabel = "app.kubernetes.io/component"
)

// Labels returns the standard labels for an OpenClawInstance
func Labels(instance *openclawv1alpha1.OpenClawInstance) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       AppName,
		"app.kubernetes.io/instance":   instance.Name,
		"app.kubernetes.io/managed-by": "openclaw-operator",
	}
}

// SelectorLabels returns the labels used for selecting pods
func SelectorLabels(instance *openclawv1alpha1.OpenClawInstance) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     AppName,
		"app.kubernetes.io/instance": instance.Name,
	}
}

// DeploymentName returns the name of the Deployment
func DeploymentName(instance *openclawv1alpha1.OpenClawInstance) string {
	return instance.Name
}

// ServiceName returns the name of the Service
func ServiceName(instance *openclawv1alpha1.OpenClawInstance) string {
	return instance.Name
}

// ServiceAccountName returns the name of the ServiceAccount
func ServiceAccountName(instance *openclawv1alpha1.OpenClawInstance) string {
	if instance.Spec.Security.RBAC.ServiceAccountName != "" {
		return instance.Spec.Security.RBAC.ServiceAccountName
	}
	return instance.Name
}

// RoleName returns the name of the Role
func RoleName(instance *openclawv1alpha1.OpenClawInstance) string {
	return instance.Name
}

// RoleBindingName returns the name of the RoleBinding
func RoleBindingName(instance *openclawv1alpha1.OpenClawInstance) string {
	return instance.Name
}

// ConfigMapName returns the name of the ConfigMap
func ConfigMapName(instance *openclawv1alpha1.OpenClawInstance) string {
	return instance.Name + "-config"
}

// PVCName returns the name of the PVC
func PVCName(instance *openclawv1alpha1.OpenClawInstance) string {
	return instance.Name + "-data"
}

// NetworkPolicyName returns the name of the NetworkPolicy
func NetworkPolicyName(instance *openclawv1alpha1.OpenClawInstance) string {
	return instance.Name
}

// PDBName returns the name of the PodDisruptionBudget
func PDBName(instance *openclawv1alpha1.OpenClawInstance) string {
	return instance.Name
}

// IngressName returns the name of the Ingress
func IngressName(instance *openclawv1alpha1.OpenClawInstance) string {
	return instance.Name
}

// GetImageRepository returns the image repository with defaults
func GetImageRepository(instance *openclawv1alpha1.OpenClawInstance) string {
	if instance.Spec.Image.Repository != "" {
		return instance.Spec.Image.Repository
	}
	return "ghcr.io/openclaw/openclaw"
}

// GetImageTag returns the image tag with defaults
func GetImageTag(instance *openclawv1alpha1.OpenClawInstance) string {
	if instance.Spec.Image.Tag != "" {
		return instance.Spec.Image.Tag
	}
	return "latest"
}

// GetImage returns the full image reference
func GetImage(instance *openclawv1alpha1.OpenClawInstance) string {
	repo := GetImageRepository(instance)
	if instance.Spec.Image.Digest != "" {
		return repo + "@" + instance.Spec.Image.Digest
	}
	return repo + ":" + GetImageTag(instance)
}

// Ptr returns a pointer to the given value
func Ptr[T any](v T) *T {
	return &v
}
