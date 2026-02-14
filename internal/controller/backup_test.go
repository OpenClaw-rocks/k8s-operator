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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	openclawv1alpha1 "github.com/openclawrocks/k8s-operator/api/v1alpha1"
	"github.com/openclawrocks/k8s-operator/internal/resources"
)

var _ = Describe("Backup on Delete", func() {
	const (
		timeout  = time.Second * 30
		interval = time.Millisecond * 250
	)

	Context("When deleting an instance with skip-backup annotation", func() {
		It("Should remove the finalizer immediately", func() {
			// Create a B2 credentials secret (needed by controller)
			b2Secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      BackupSecretName,
					Namespace: "default",
				},
				Data: map[string][]byte{
					"B2_BUCKET":   []byte("test-bucket"),
					"B2_KEY_ID":   []byte("key123"),
					"B2_APP_KEY":  []byte("secret456"),
					"B2_ENDPOINT": []byte("https://s3.example.com"),
				},
			}
			_ = k8sClient.Create(ctx, b2Secret)

			instance := &openclawv1alpha1.OpenClawInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "backup-skip-test",
					Namespace: "default",
					Annotations: map[string]string{
						AnnotationSkipBackup: "true",
					},
				},
				Spec: openclawv1alpha1.OpenClawInstanceSpec{},
			}
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			instanceKey := types.NamespacedName{Name: "backup-skip-test", Namespace: "default"}

			// Wait for instance to be provisioned (finalizer added)
			Eventually(func() bool {
				inst := &openclawv1alpha1.OpenClawInstance{}
				if err := k8sClient.Get(ctx, instanceKey, inst); err != nil {
					return false
				}
				return inst.Status.Phase != "" && inst.Status.Phase != openclawv1alpha1.PhasePending
			}, timeout, interval).Should(BeTrue())

			// Delete the instance
			Expect(k8sClient.Delete(ctx, instance)).Should(Succeed())

			// The instance should be fully deleted (finalizer removed immediately)
			Eventually(func() bool {
				inst := &openclawv1alpha1.OpenClawInstance{}
				err := k8sClient.Get(ctx, instanceKey, inst)
				return err != nil // NotFound means deleted
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When deleting an instance with backup", func() {
		It("Should enter BackingUp phase and scale down StatefulSet", func() {
			instance := &openclawv1alpha1.OpenClawInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "backup-scale-test",
					Namespace: "default",
				},
				Spec: openclawv1alpha1.OpenClawInstanceSpec{},
			}
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			instanceKey := types.NamespacedName{Name: "backup-scale-test", Namespace: "default"}

			// Wait for StatefulSet to be created
			Eventually(func() bool {
				sts := &appsv1.StatefulSet{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      resources.StatefulSetName(instance),
					Namespace: "default",
				}, sts)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Delete the instance
			Expect(k8sClient.Delete(ctx, instance)).Should(Succeed())

			// Verify it enters BackingUp phase
			Eventually(func() string {
				inst := &openclawv1alpha1.OpenClawInstance{}
				if err := k8sClient.Get(ctx, instanceKey, inst); err != nil {
					return ""
				}
				return inst.Status.Phase
			}, timeout, interval).Should(Equal(openclawv1alpha1.PhaseBackingUp))

			// Verify StatefulSet is scaled to 0
			Eventually(func() int32 {
				sts := &appsv1.StatefulSet{}
				if err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      resources.StatefulSetName(instance),
					Namespace: "default",
				}, sts); err != nil {
					return -1
				}
				if sts.Spec.Replicas == nil {
					return 1
				}
				return *sts.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(0)))

			// Clean up: annotate to skip backup so finalizer gets removed
			Eventually(func() error {
				inst := &openclawv1alpha1.OpenClawInstance{}
				if err := k8sClient.Get(ctx, instanceKey, inst); err != nil {
					return err
				}
				if inst.Annotations == nil {
					inst.Annotations = map[string]string{}
				}
				inst.Annotations[AnnotationSkipBackup] = "true"
				return k8sClient.Update(ctx, inst)
			}, timeout, interval).Should(Succeed())

			// Wait for deletion to complete
			Eventually(func() bool {
				inst := &openclawv1alpha1.OpenClawInstance{}
				err := k8sClient.Get(ctx, instanceKey, inst)
				return err != nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})
