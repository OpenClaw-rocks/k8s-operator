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
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openclawv1alpha1 "github.com/openclawrocks/k8s-operator/api/v1alpha1"
)

// BuildConfigMap creates a ConfigMap for the OpenClawInstance configuration
func BuildConfigMap(instance *openclawv1alpha1.OpenClawInstance) *corev1.ConfigMap {
	labels := Labels(instance)

	// Generate openclaw.json content from raw config
	configContent := "{}"
	if instance.Spec.Config.Raw != nil && len(instance.Spec.Config.Raw.Raw) > 0 {
		configContent = string(instance.Spec.Config.Raw.Raw)
	}

	// Try to pretty-print the JSON
	var parsed interface{}
	if err := json.Unmarshal([]byte(configContent), &parsed); err == nil {
		if pretty, err := json.MarshalIndent(parsed, "", "  "); err == nil {
			configContent = string(pretty)
		}
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ConfigMapName(instance),
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"openclaw.json": configContent,
		},
	}
}
