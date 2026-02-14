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

package controller

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	openclawv1alpha1 "github.com/openclawrocks/k8s-operator/api/v1alpha1"
	"github.com/openclawrocks/k8s-operator/internal/resources"
)

var _ = Describe("OpenClawInstance Controller", func() {
	const (
		timeout  = time.Second * 30
		interval = time.Millisecond * 250
	)

	Context("When creating OpenClawInstance", func() {
		It("Should create all managed resources", func() {
			By("Creating an OpenClawInstance")
			instance := &openclawv1alpha1.OpenClawInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-instance",
					Namespace: "default",
				},
				Spec: openclawv1alpha1.OpenClawInstanceSpec{
					EnvFrom: []corev1.EnvFromSource{
						{
							SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "test-secret",
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			instanceLookupKey := types.NamespacedName{Name: "test-instance", Namespace: "default"}

			// Wait for the instance to be created
			Eventually(func() bool {
				err := k8sClient.Get(ctx, instanceLookupKey, instance)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying ServiceAccount is created")
			Eventually(func() bool {
				sa := &corev1.ServiceAccount{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      resources.ServiceAccountName(instance),
					Namespace: "default",
				}, sa)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying Role is created")
			Eventually(func() bool {
				role := &rbacv1.Role{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      resources.RoleName(instance),
					Namespace: "default",
				}, role)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying RoleBinding is created")
			Eventually(func() bool {
				roleBinding := &rbacv1.RoleBinding{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      resources.RoleBindingName(instance),
					Namespace: "default",
				}, roleBinding)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying NetworkPolicy is created")
			Eventually(func() bool {
				np := &networkingv1.NetworkPolicy{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      resources.NetworkPolicyName(instance),
					Namespace: "default",
				}, np)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying PodDisruptionBudget is created")
			Eventually(func() bool {
				pdb := &policyv1.PodDisruptionBudget{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      resources.PDBName(instance),
					Namespace: "default",
				}, pdb)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying StatefulSet is created")
			Eventually(func() bool {
				sts := &appsv1.StatefulSet{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      resources.StatefulSetName(instance),
					Namespace: "default",
				}, sts)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying Service is created")
			Eventually(func() bool {
				svc := &corev1.Service{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      resources.ServiceName(instance),
					Namespace: "default",
				}, svc)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying PVC is created")
			Eventually(func() bool {
				pvc := &corev1.PersistentVolumeClaim{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      resources.PVCName(instance),
					Namespace: "default",
				}, pvc)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying instance status is updated")
			Eventually(func() string {
				inst := &openclawv1alpha1.OpenClawInstance{}
				err := k8sClient.Get(ctx, instanceLookupKey, inst)
				if err != nil {
					return ""
				}
				return inst.Status.Phase
			}, timeout, interval).ShouldNot(BeEmpty())

			By("Cleaning up")
			Expect(k8sClient.Delete(ctx, instance)).Should(Succeed())
		})
	})

	Context("When StatefulSet security contexts", func() {
		It("Should enforce non-root execution", func() {
			instance := &openclawv1alpha1.OpenClawInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "security-test",
					Namespace: "default",
				},
				Spec: openclawv1alpha1.OpenClawInstanceSpec{},
			}

			sts := resources.BuildStatefulSet(instance)

			// Verify pod security context
			Expect(sts.Spec.Template.Spec.SecurityContext).NotTo(BeNil())
			Expect(*sts.Spec.Template.Spec.SecurityContext.RunAsNonRoot).To(BeTrue())
			Expect(*sts.Spec.Template.Spec.SecurityContext.RunAsUser).To(Equal(int64(1000)))

			// Verify container security context
			container := sts.Spec.Template.Spec.Containers[0]
			Expect(container.SecurityContext).NotTo(BeNil())
			Expect(*container.SecurityContext.AllowPrivilegeEscalation).To(BeFalse())
			Expect(container.SecurityContext.Capabilities.Drop).To(ContainElement(corev1.Capability("ALL")))
		})
	})

	Context("When NetworkPolicy is configured", func() {
		It("Should create proper ingress and egress rules", func() {
			instance := &openclawv1alpha1.OpenClawInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "netpol-test",
					Namespace: "default",
				},
				Spec: openclawv1alpha1.OpenClawInstanceSpec{
					Security: openclawv1alpha1.SecuritySpec{
						NetworkPolicy: openclawv1alpha1.NetworkPolicySpec{
							AllowedIngressCIDRs: []string{"10.0.0.0/8"},
						},
					},
				},
			}

			np := resources.BuildNetworkPolicy(instance)

			// Verify policy types
			Expect(np.Spec.PolicyTypes).To(ContainElements(
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			))

			// Verify egress rules allow HTTPS (port 443)
			var hasHTTPSEgress bool
			for _, rule := range np.Spec.Egress {
				for _, port := range rule.Ports {
					if port.Port.IntVal == 443 {
						hasHTTPSEgress = true
						break
					}
				}
			}
			Expect(hasHTTPSEgress).To(BeTrue())

			// Verify ingress rules include our CIDR
			var hasCustomCIDR bool
			for _, rule := range np.Spec.Ingress {
				for _, from := range rule.From {
					if from.IPBlock != nil && from.IPBlock.CIDR == "10.0.0.0/8" {
						hasCustomCIDR = true
						break
					}
				}
			}
			Expect(hasCustomCIDR).To(BeTrue())
		})
	})
})
