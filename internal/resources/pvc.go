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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openclawv1alpha1 "github.com/openclawrocks/k8s-operator/api/v1alpha1"
)

// BuildPVC creates a PersistentVolumeClaim for the OpenClawInstance
func BuildPVC(instance *openclawv1alpha1.OpenClawInstance) *corev1.PersistentVolumeClaim {
	labels := Labels(instance)

	// Get storage size with default
	size := instance.Spec.Storage.Persistence.Size
	if size == "" {
		size = "10Gi"
	}

	// Get access modes with default
	accessModes := instance.Spec.Storage.Persistence.AccessModes
	if len(accessModes) == 0 {
		accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PVCName(instance),
			Namespace: instance.Namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"openclaw.openclaw.io/backup-enabled": "true",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: accessModes,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(size),
				},
			},
		},
	}

	// Set storage class if specified
	if instance.Spec.Storage.Persistence.StorageClass != nil {
		pvc.Spec.StorageClassName = instance.Spec.Storage.Persistence.StorageClass
	}

	return pvc
}
