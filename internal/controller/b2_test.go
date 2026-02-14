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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openclawv1alpha1 "github.com/openclawrocks/k8s-operator/api/v1alpha1"
)

var _ = Describe("B2 Helpers", func() {
	Context("getTenantID", func() {
		It("Should return the tenant label value when present", func() {
			instance := &openclawv1alpha1.OpenClawInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "oc-tenant-cus_123",
					Labels: map[string]string{
						LabelTenant: "cus_456",
					},
				},
			}
			Expect(getTenantID(instance)).To(Equal("cus_456"))
		})

		It("Should extract tenant from namespace when label is missing", func() {
			instance := &openclawv1alpha1.OpenClawInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "oc-tenant-cus_789",
				},
			}
			Expect(getTenantID(instance)).To(Equal("cus_789"))
		})

		It("Should return namespace as-is when not in oc-tenant format", func() {
			instance := &openclawv1alpha1.OpenClawInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
			}
			Expect(getTenantID(instance)).To(Equal("default"))
		})
	})

	Context("buildRcloneJob", func() {
		var creds *b2Credentials

		BeforeEach(func() {
			creds = &b2Credentials{
				Bucket:   "test-bucket",
				KeyID:    "key123",
				AppKey:   "secret456",
				Endpoint: "https://s3.us-west-000.backblazeb2.com",
			}
		})

		It("Should build a backup Job with correct args and SecurityContext", func() {
			instance := &openclawv1alpha1.OpenClawInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myinst",
					Namespace: "oc-tenant-t1",
				},
			}
			labels := backupLabels(instance, "backup")
			job := buildRcloneJob("myinst-backup", "oc-tenant-t1", "myinst-data", "backups/t1/myinst/2026-01-01T000000Z", labels, creds, true)

			Expect(job.Name).To(Equal("myinst-backup"))
			Expect(job.Namespace).To(Equal("oc-tenant-t1"))
			Expect(*job.Spec.BackoffLimit).To(Equal(int32(3)))
			Expect(*job.Spec.TTLSecondsAfterFinished).To(Equal(int32(86400)))

			// Verify container
			container := job.Spec.Template.Spec.Containers[0]
			Expect(container.Image).To(Equal(RcloneImage))
			Expect(container.Args[0]).To(Equal("sync"))
			Expect(container.Args[1]).To(Equal("/data/")) // PVC source for backup

			// Verify SecurityContext
			podSC := job.Spec.Template.Spec.SecurityContext
			Expect(*podSC.RunAsUser).To(Equal(int64(1000)))
			Expect(*podSC.RunAsGroup).To(Equal(int64(1000)))
			Expect(*podSC.FSGroup).To(Equal(int64(1000)))

			// Verify PVC volume
			vol := job.Spec.Template.Spec.Volumes[0]
			Expect(vol.PersistentVolumeClaim.ClaimName).To(Equal("myinst-data"))

			// Verify env vars
			var envNames []string
			for _, e := range container.Env {
				envNames = append(envNames, e.Name)
			}
			Expect(envNames).To(ContainElements("B2_ENDPOINT", "B2_KEY_ID", "B2_APP_KEY"))
		})

		It("Should build a restore Job with B2 as source", func() {
			instance := &openclawv1alpha1.OpenClawInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myinst",
					Namespace: "oc-tenant-t1",
				},
			}
			labels := backupLabels(instance, "restore")
			job := buildRcloneJob("myinst-restore", "oc-tenant-t1", "myinst-data", "backups/t1/myinst/2026-01-01T000000Z", labels, creds, false)

			container := job.Spec.Template.Spec.Containers[0]
			Expect(container.Args[0]).To(Equal("sync"))
			// For restore, dest is /data/
			Expect(container.Args[2]).To(Equal("/data/"))

			vol := job.Spec.Template.Spec.Volumes[0]
			Expect(vol.PersistentVolumeClaim.ClaimName).To(Equal("myinst-data"))
		})
	})

	Context("isJobFinished", func() {
		It("Should return false for an active Job", func() {
			job := &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{Name: "test-job", Namespace: "default"},
			}
			finished, _ := isJobFinished(job)
			Expect(finished).To(BeFalse())
		})

		It("Should return true with Complete for a succeeded Job", func() {
			job := &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{Name: "test-job", Namespace: "default"},
				Status: batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{
						{
							Type:   batchv1.JobComplete,
							Status: corev1.ConditionTrue,
						},
					},
				},
			}
			finished, condType := isJobFinished(job)
			Expect(finished).To(BeTrue())
			Expect(condType).To(Equal(batchv1.JobComplete))
		})

		It("Should return true with Failed for a failed Job", func() {
			job := &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{Name: "test-job", Namespace: "default"},
				Status: batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{
						{
							Type:   batchv1.JobFailed,
							Status: corev1.ConditionTrue,
						},
					},
				},
			}
			finished, condType := isJobFinished(job)
			Expect(finished).To(BeTrue())
			Expect(condType).To(Equal(batchv1.JobFailed))
		})
	})

	Context("backupLabels", func() {
		It("Should include tenant, instance, and job-type labels", func() {
			instance := &openclawv1alpha1.OpenClawInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myinst",
					Namespace: "oc-tenant-cus_123",
					Labels: map[string]string{
						LabelTenant: "cus_123",
					},
				},
			}
			labels := backupLabels(instance, "backup")
			Expect(labels[LabelTenant]).To(Equal("cus_123"))
			Expect(labels[LabelInstance]).To(Equal("myinst"))
			Expect(labels["openclaw.rocks/job-type"]).To(Equal("backup"))
			Expect(labels[LabelManagedBy]).To(Equal("openclaw-operator"))
		})
	})
})
