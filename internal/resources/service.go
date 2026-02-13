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
	"k8s.io/apimachinery/pkg/util/intstr"

	openclawv1alpha1 "github.com/openclawrocks/k8s-operator/api/v1alpha1"
)

// BuildService creates a Service for the OpenClawInstance
func BuildService(instance *openclawv1alpha1.OpenClawInstance) *corev1.Service {
	labels := Labels(instance)
	selectorLabels := SelectorLabels(instance)

	serviceType := instance.Spec.Networking.Service.Type
	if serviceType == "" {
		serviceType = corev1.ServiceTypeClusterIP
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        ServiceName(instance),
			Namespace:   instance.Namespace,
			Labels:      labels,
			Annotations: instance.Spec.Networking.Service.Annotations,
		},
		Spec: corev1.ServiceSpec{
			Type:            serviceType,
			Selector:        selectorLabels,
			SessionAffinity: corev1.ServiceAffinityNone,
			Ports: []corev1.ServicePort{
				{
					Name:       "gateway",
					Port:       int32(GatewayPort),
					TargetPort: intstr.FromInt(GatewayPort),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "canvas",
					Port:       int32(CanvasPort),
					TargetPort: intstr.FromInt(CanvasPort),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	// Add Chromium port if enabled
	if instance.Spec.Chromium.Enabled {
		service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{
			Name:       "chromium",
			Port:       int32(ChromiumPort),
			TargetPort: intstr.FromInt(ChromiumPort),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	return service
}
