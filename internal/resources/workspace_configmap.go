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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openclawv1alpha1 "github.com/openclawrocks/k8s-operator/api/v1alpha1"
)

// BuildWorkspaceConfigMap creates a ConfigMap containing workspace seed files.
// Returns nil if the instance has no workspace files configured.
func BuildWorkspaceConfigMap(instance *openclawv1alpha1.OpenClawInstance) *corev1.ConfigMap {
	if instance.Spec.Workspace == nil || len(instance.Spec.Workspace.InitialFiles) == 0 {
		return nil
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      WorkspaceConfigMapName(instance),
			Namespace: instance.Namespace,
			Labels:    Labels(instance),
		},
		Data: instance.Spec.Workspace.InitialFiles,
	}
}
