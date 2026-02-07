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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openclawv1alpha1 "github.com/openclawrocks/k8s-operator/api/v1alpha1"
)

// BuildServiceAccount creates a ServiceAccount for the OpenClawInstance
func BuildServiceAccount(instance *openclawv1alpha1.OpenClawInstance) *corev1.ServiceAccount {
	labels := Labels(instance)

	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceAccountName(instance),
			Namespace: instance.Namespace,
			Labels:    labels,
		},
	}
}

// BuildRole creates a Role for the OpenClawInstance
// This implements the principle of least privilege - only granting what's needed
func BuildRole(instance *openclawv1alpha1.OpenClawInstance) *rbacv1.Role {
	labels := Labels(instance)

	// Base rules - minimal permissions needed by OpenClaw
	rules := []rbacv1.PolicyRule{
		// OpenClaw only needs to read its own config
		{
			APIGroups:     []string{""},
			Resources:     []string{"configmaps"},
			ResourceNames: []string{ConfigMapName(instance)},
			Verbs:         []string{"get", "watch"},
		},
	}

	// Add additional rules from spec
	for _, rule := range instance.Spec.Security.RBAC.AdditionalRules {
		rules = append(rules, rbacv1.PolicyRule{
			APIGroups: rule.APIGroups,
			Resources: rule.Resources,
			Verbs:     rule.Verbs,
		})
	}

	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RoleName(instance),
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Rules: rules,
	}
}

// BuildRoleBinding creates a RoleBinding for the OpenClawInstance
func BuildRoleBinding(instance *openclawv1alpha1.OpenClawInstance) *rbacv1.RoleBinding {
	labels := Labels(instance)

	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RoleBindingName(instance),
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     RoleName(instance),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      ServiceAccountName(instance),
				Namespace: instance.Namespace,
			},
		},
	}
}
