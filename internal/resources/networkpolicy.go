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
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	openclawv1alpha1 "github.com/openclawrocks/k8s-operator/api/v1alpha1"
)

// BuildNetworkPolicy creates a NetworkPolicy for the OpenClawInstance
// This implements a default-deny with selective allowlist approach
func BuildNetworkPolicy(instance *openclawv1alpha1.OpenClawInstance) *networkingv1.NetworkPolicy {
	labels := Labels(instance)
	selectorLabels := SelectorLabels(instance)

	np := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      NetworkPolicyName(instance),
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: selectorLabels,
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
			Ingress: buildIngressRules(instance),
			Egress:  buildEgressRules(instance),
		},
	}

	return np
}

// buildIngressRules creates the ingress rules for the NetworkPolicy
func buildIngressRules(instance *openclawv1alpha1.OpenClawInstance) []networkingv1.NetworkPolicyIngressRule {
	rules := []networkingv1.NetworkPolicyIngressRule{}

	// Allow from same namespace by default
	rules = append(rules, networkingv1.NetworkPolicyIngressRule{
		From: []networkingv1.NetworkPolicyPeer{
			{
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"kubernetes.io/metadata.name": instance.Namespace,
					},
				},
			},
		},
		Ports: []networkingv1.NetworkPolicyPort{
			{
				Protocol: Ptr(corev1.ProtocolTCP),
				Port:     Ptr(intstr.FromInt(GatewayPort)),
			},
			{
				Protocol: Ptr(corev1.ProtocolTCP),
				Port:     Ptr(intstr.FromInt(CanvasPort)),
			},
		},
	})

	// Allow from specified namespaces
	for _, ns := range instance.Spec.Security.NetworkPolicy.AllowedIngressNamespaces {
		rules = append(rules, networkingv1.NetworkPolicyIngressRule{
			From: []networkingv1.NetworkPolicyPeer{
				{
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"kubernetes.io/metadata.name": ns,
						},
					},
				},
			},
			Ports: []networkingv1.NetworkPolicyPort{
				{
					Protocol: Ptr(corev1.ProtocolTCP),
					Port:     Ptr(intstr.FromInt(GatewayPort)),
				},
				{
					Protocol: Ptr(corev1.ProtocolTCP),
					Port:     Ptr(intstr.FromInt(CanvasPort)),
				},
			},
		})
	}

	// Allow from specified CIDRs
	for _, cidr := range instance.Spec.Security.NetworkPolicy.AllowedIngressCIDRs {
		rules = append(rules, networkingv1.NetworkPolicyIngressRule{
			From: []networkingv1.NetworkPolicyPeer{
				{
					IPBlock: &networkingv1.IPBlock{
						CIDR: cidr,
					},
				},
			},
			Ports: []networkingv1.NetworkPolicyPort{
				{
					Protocol: Ptr(corev1.ProtocolTCP),
					Port:     Ptr(intstr.FromInt(GatewayPort)),
				},
				{
					Protocol: Ptr(corev1.ProtocolTCP),
					Port:     Ptr(intstr.FromInt(CanvasPort)),
				},
			},
		})
	}

	return rules
}

// buildEgressRules creates the egress rules for the NetworkPolicy
func buildEgressRules(instance *openclawv1alpha1.OpenClawInstance) []networkingv1.NetworkPolicyEgressRule {
	rules := []networkingv1.NetworkPolicyEgressRule{}

	// Allow DNS if enabled (default: true)
	allowDNS := instance.Spec.Security.NetworkPolicy.AllowDNS == nil || *instance.Spec.Security.NetworkPolicy.AllowDNS
	if allowDNS {
		rules = append(rules, networkingv1.NetworkPolicyEgressRule{
			To: []networkingv1.NetworkPolicyPeer{},
			Ports: []networkingv1.NetworkPolicyPort{
				{
					Protocol: Ptr(corev1.ProtocolUDP),
					Port:     Ptr(intstr.FromInt(53)),
				},
				{
					Protocol: Ptr(corev1.ProtocolTCP),
					Port:     Ptr(intstr.FromInt(53)),
				},
			},
		})
	}

	// Allow HTTPS egress for AI APIs (port 443)
	// This is essential for OpenClaw to communicate with AI providers
	rules = append(rules, networkingv1.NetworkPolicyEgressRule{
		To: []networkingv1.NetworkPolicyPeer{},
		Ports: []networkingv1.NetworkPolicyPort{
			{
				Protocol: Ptr(corev1.ProtocolTCP),
				Port:     Ptr(intstr.FromInt(443)),
			},
		},
	})

	// Allow additional egress CIDRs if specified
	for _, cidr := range instance.Spec.Security.NetworkPolicy.AllowedEgressCIDRs {
		rules = append(rules, networkingv1.NetworkPolicyEgressRule{
			To: []networkingv1.NetworkPolicyPeer{
				{
					IPBlock: &networkingv1.IPBlock{
						CIDR: cidr,
					},
				},
			},
		})
	}

	// Append user-defined additional egress rules
	rules = append(rules, instance.Spec.Security.NetworkPolicy.AdditionalEgress...)

	return rules
}
