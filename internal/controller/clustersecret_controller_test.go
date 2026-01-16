// SPDX-FileCopyrightText: 2020 The Kubernetes Authors
// SPDX-FileCopyrightText: 2024 Kalle Fagerberg
// SPDX-FileCopyrightText: 2024 Nicolas Kowenski
// SPDX-License-Identifier: MIT

package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clustersecretiov2 "github.com/zakkg3/ClusterSecret/api/v2"
)

var _ = Describe("ClusterSecret Controller", func() {
	const (
		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	ctx := context.Background()

	Context("When creating a ClusterSecret with namespace selector", func() {
		var (
			csecName   string
			testNs1    *corev1.Namespace
			testNs2    *corev1.Namespace
			excludedNs *corev1.Namespace
		)

		BeforeEach(func() {
			csecName = "test-csec-" + randString(5)

			// Create test namespaces with different labels
			testNs1 = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-ns1-" + randString(5),
					Labels: map[string]string{"env": "prod", "team": "backend"},
				},
			}
			Expect(k8sClient.Create(ctx, testNs1)).To(Succeed())

			testNs2 = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-ns2-" + randString(5),
					Labels: map[string]string{"env": "prod", "team": "frontend"},
				},
			}
			Expect(k8sClient.Create(ctx, testNs2)).To(Succeed())

			excludedNs = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "excluded-ns-" + randString(5),
					Labels: map[string]string{"env": "dev"},
				},
			}
			Expect(k8sClient.Create(ctx, excludedNs)).To(Succeed())
		})

		AfterEach(func() {
			// Cleanup ClusterSecret
			csec := &clustersecretiov2.ClusterSecret{}
			if err := k8sClient.Get(ctx, types.NamespacedName{Name: csecName}, csec); err == nil {
				Expect(k8sClient.Delete(ctx, csec)).To(Succeed())
			}

			// Cleanup namespaces
			for _, ns := range []*corev1.Namespace{testNs1, testNs2, excludedNs} {
				if ns != nil {
					_ = k8sClient.Delete(ctx, ns)
				}
			}
		})

		It("should create secrets in matching namespaces only", func() {
			By("Creating a ClusterSecret that matches env=prod")
			csec := &clustersecretiov2.ClusterSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name: csecName,
				},
				Spec: clustersecretiov2.ClusterSecretSpec{
					NamespaceSelectorTerms: []clustersecretiov2.NamespaceSelectorTerm{{
						MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{{
							Key:      "env",
							Operator: clustersecretiov2.NamespaceSelectorOpIn,
							Values:   []string{"prod"},
						}},
					}},
					Template: clustersecretiov2.SecretTemplate{
						Data: map[string][]byte{
							"api-key": []byte("secret-value"),
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, csec)).To(Succeed())

			By("Verifying secret is created in matching namespaces")
			Eventually(func() error {
				secret := &corev1.Secret{}
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      csecName,
					Namespace: testNs1.Name,
				}, secret)
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				secret := &corev1.Secret{}
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      csecName,
					Namespace: testNs2.Name,
				}, secret)
			}, timeout, interval).Should(Succeed())

			By("Verifying secret is NOT created in non-matching namespace")
			Consistently(func() bool {
				secret := &corev1.Secret{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      csecName,
					Namespace: excludedNs.Name,
				}, secret)
				return errors.IsNotFound(err)
			}, time.Second*2, interval).Should(BeTrue())

			By("Verifying secret data is correct")
			secret := &corev1.Secret{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      csecName,
				Namespace: testNs1.Name,
			}, secret)).To(Succeed())
			Expect(secret.Data).To(HaveKeyWithValue("api-key", []byte("secret-value")))
		})

		It("should set owner reference on created secrets", func() {
			By("Creating a ClusterSecret")
			csec := &clustersecretiov2.ClusterSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name: csecName,
				},
				Spec: clustersecretiov2.ClusterSecretSpec{
					NamespaceSelectorTerms: []clustersecretiov2.NamespaceSelectorTerm{{
						MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{{
							Key:      "env",
							Operator: clustersecretiov2.NamespaceSelectorOpIn,
							Values:   []string{"prod"},
						}},
					}},
					Template: clustersecretiov2.SecretTemplate{
						Data: map[string][]byte{"key": []byte("value")},
					},
				},
			}
			Expect(k8sClient.Create(ctx, csec)).To(Succeed())

			By("Verifying owner reference is set")
			Eventually(func() bool {
				secret := &corev1.Secret{}
				if err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      csecName,
					Namespace: testNs1.Name,
				}, secret); err != nil {
					return false
				}
				for _, ref := range secret.OwnerReferences {
					if ref.Kind == "ClusterSecret" && ref.Name == csecName {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})

		It("should update status with matching namespaces", func() {
			By("Creating a ClusterSecret")
			csec := &clustersecretiov2.ClusterSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name: csecName,
				},
				Spec: clustersecretiov2.ClusterSecretSpec{
					NamespaceSelectorTerms: []clustersecretiov2.NamespaceSelectorTerm{{
						MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{{
							Key:      "env",
							Operator: clustersecretiov2.NamespaceSelectorOpIn,
							Values:   []string{"prod"},
						}},
					}},
					Template: clustersecretiov2.SecretTemplate{
						Data: map[string][]byte{"key": []byte("value")},
					},
				},
			}
			Expect(k8sClient.Create(ctx, csec)).To(Succeed())

			By("Verifying status is updated")
			Eventually(func() int32 {
				updated := &clustersecretiov2.ClusterSecret{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: csecName}, updated); err != nil {
					return -1
				}
				return updated.Status.MatchingNamespacesCount
			}, timeout, interval).Should(BeNumerically(">=", 2))
		})
	})

	Context("When using dataFrom to reference another secret", func() {
		var (
			csecName     string
			sourceNs     *corev1.Namespace
			targetNs     *corev1.Namespace
			sourceSecret *corev1.Secret
		)

		BeforeEach(func() {
			csecName = "test-datafrom-" + randString(5)

			sourceNs = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "source-ns-" + randString(5),
				},
			}
			Expect(k8sClient.Create(ctx, sourceNs)).To(Succeed())

			targetNs = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "target-ns-" + randString(5),
					Labels: map[string]string{"sync": "true"},
				},
			}
			Expect(k8sClient.Create(ctx, targetNs)).To(Succeed())

			sourceSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "source-secret",
					Namespace: sourceNs.Name,
				},
				Data: map[string][]byte{
					"username": []byte("admin"),
					"password": []byte("secret123"),
				},
			}
			Expect(k8sClient.Create(ctx, sourceSecret)).To(Succeed())
		})

		AfterEach(func() {
			csec := &clustersecretiov2.ClusterSecret{}
			if err := k8sClient.Get(ctx, types.NamespacedName{Name: csecName}, csec); err == nil {
				Expect(k8sClient.Delete(ctx, csec)).To(Succeed())
			}
			for _, ns := range []*corev1.Namespace{sourceNs, targetNs} {
				if ns != nil {
					_ = k8sClient.Delete(ctx, ns)
				}
			}
		})

		It("should copy data from referenced secret", func() {
			By("Creating ClusterSecret with dataFrom")
			csec := &clustersecretiov2.ClusterSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name: csecName,
				},
				Spec: clustersecretiov2.ClusterSecretSpec{
					NamespaceSelectorTerms: []clustersecretiov2.NamespaceSelectorTerm{{
						MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{{
							Key:      "sync",
							Operator: clustersecretiov2.NamespaceSelectorOpExists,
						}},
					}},
					DataFrom: []clustersecretiov2.DataFrom{{
						SecretRef: &clustersecretiov2.SecretReference{
							Name:      sourceSecret.Name,
							Namespace: sourceNs.Name,
						},
					}},
					Template: clustersecretiov2.SecretTemplate{},
				},
			}
			Expect(k8sClient.Create(ctx, csec)).To(Succeed())

			By("Verifying target secret has data from source")
			Eventually(func() map[string][]byte {
				secret := &corev1.Secret{}
				if err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      csecName,
					Namespace: targetNs.Name,
				}, secret); err != nil {
					return nil
				}
				return secret.Data
			}, timeout, interval).Should(And(
				HaveKeyWithValue("username", []byte("admin")),
				HaveKeyWithValue("password", []byte("secret123")),
			))
		})
	})

	Context("When using dataValueFrom to reference a specific key", func() {
		var (
			csecName     string
			sourceNs     *corev1.Namespace
			targetNs     *corev1.Namespace
			sourceSecret *corev1.Secret
		)

		BeforeEach(func() {
			csecName = "test-valuefrom-" + randString(5)

			sourceNs = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "source-ns-" + randString(5),
				},
			}
			Expect(k8sClient.Create(ctx, sourceNs)).To(Succeed())

			targetNs = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "target-ns-" + randString(5),
					Labels: map[string]string{"sync": "true"},
				},
			}
			Expect(k8sClient.Create(ctx, targetNs)).To(Succeed())

			sourceSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "source-secret",
					Namespace: sourceNs.Name,
				},
				Data: map[string][]byte{
					"db-password": []byte("dbpass123"),
					"db-host":     []byte("localhost"),
				},
			}
			Expect(k8sClient.Create(ctx, sourceSecret)).To(Succeed())
		})

		AfterEach(func() {
			csec := &clustersecretiov2.ClusterSecret{}
			if err := k8sClient.Get(ctx, types.NamespacedName{Name: csecName}, csec); err == nil {
				Expect(k8sClient.Delete(ctx, csec)).To(Succeed())
			}
			for _, ns := range []*corev1.Namespace{sourceNs, targetNs} {
				if ns != nil {
					_ = k8sClient.Delete(ctx, ns)
				}
			}
		})

		It("should copy specific key from referenced secret", func() {
			By("Creating ClusterSecret with dataValueFrom")
			csec := &clustersecretiov2.ClusterSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name: csecName,
				},
				Spec: clustersecretiov2.ClusterSecretSpec{
					NamespaceSelectorTerms: []clustersecretiov2.NamespaceSelectorTerm{{
						MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{{
							Key:      "sync",
							Operator: clustersecretiov2.NamespaceSelectorOpExists,
						}},
					}},
					DataValueFrom: map[string]clustersecretiov2.DataValueFrom{
						"password": {
							SecretKeyRef: &clustersecretiov2.SecretKeyReference{
								Name:      sourceSecret.Name,
								Namespace: sourceNs.Name,
								Key:       "db-password",
							},
						},
					},
					Template: clustersecretiov2.SecretTemplate{},
				},
			}
			Expect(k8sClient.Create(ctx, csec)).To(Succeed())

			By("Verifying target secret has specific key renamed")
			Eventually(func() map[string][]byte {
				secret := &corev1.Secret{}
				if err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      csecName,
					Namespace: targetNs.Name,
				}, secret); err != nil {
					return nil
				}
				return secret.Data
			}, timeout, interval).Should(And(
				HaveKeyWithValue("password", []byte("dbpass123")),
				Not(HaveKey("db-host")), // Should not include other keys
			))
		})
	})

	Context("When updating a ClusterSecret", func() {
		var (
			csecName string
			testNs   *corev1.Namespace
		)

		BeforeEach(func() {
			csecName = "test-update-" + randString(5)

			testNs = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "update-ns-" + randString(5),
					Labels: map[string]string{"test": "true"},
				},
			}
			Expect(k8sClient.Create(ctx, testNs)).To(Succeed())
		})

		AfterEach(func() {
			csec := &clustersecretiov2.ClusterSecret{}
			if err := k8sClient.Get(ctx, types.NamespacedName{Name: csecName}, csec); err == nil {
				Expect(k8sClient.Delete(ctx, csec)).To(Succeed())
			}
			if testNs != nil {
				_ = k8sClient.Delete(ctx, testNs)
			}
		})

		It("should update secrets when data changes", func() {
			By("Creating initial ClusterSecret")
			csec := &clustersecretiov2.ClusterSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name: csecName,
				},
				Spec: clustersecretiov2.ClusterSecretSpec{
					NamespaceSelectorTerms: []clustersecretiov2.NamespaceSelectorTerm{{
						MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{{
							Key:      "test",
							Operator: clustersecretiov2.NamespaceSelectorOpExists,
						}},
					}},
					Template: clustersecretiov2.SecretTemplate{
						Data: map[string][]byte{"key": []byte("original")},
					},
				},
			}
			Expect(k8sClient.Create(ctx, csec)).To(Succeed())

			By("Waiting for initial secret to be created")
			Eventually(func() []byte {
				secret := &corev1.Secret{}
				if err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      csecName,
					Namespace: testNs.Name,
				}, secret); err != nil {
					return nil
				}
				return secret.Data["key"]
			}, timeout, interval).Should(Equal([]byte("original")))

			By("Updating ClusterSecret data")
			Eventually(func() error {
				updated := &clustersecretiov2.ClusterSecret{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: csecName}, updated); err != nil {
					return err
				}
				updated.Spec.Template.Data = map[string][]byte{"key": []byte("updated")}
				return k8sClient.Update(ctx, updated)
			}, timeout, interval).Should(Succeed())

			By("Verifying secret is updated")
			Eventually(func() []byte {
				secret := &corev1.Secret{}
				if err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      csecName,
					Namespace: testNs.Name,
				}, secret); err != nil {
					return nil
				}
				return secret.Data["key"]
			}, timeout, interval).Should(Equal([]byte("updated")))
		})
	})

	Context("When deleting a ClusterSecret", func() {
		var (
			csecName string
			testNs   *corev1.Namespace
		)

		BeforeEach(func() {
			csecName = "test-delete-" + randString(5)

			testNs = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "delete-ns-" + randString(5),
					Labels: map[string]string{"cleanup": "true"},
				},
			}
			Expect(k8sClient.Create(ctx, testNs)).To(Succeed())
		})

		AfterEach(func() {
			// Clean up secret manually since envtest doesn't run garbage collector
			secret := &corev1.Secret{}
			if err := k8sClient.Get(ctx, types.NamespacedName{
				Name:      csecName,
				Namespace: testNs.Name,
			}, secret); err == nil {
				_ = k8sClient.Delete(ctx, secret)
			}
			if testNs != nil {
				_ = k8sClient.Delete(ctx, testNs)
			}
		})

		It("should have owner reference that enables garbage collection", func() {
			// Note: envtest doesn't run the garbage collector, so we verify
			// the owner reference is set correctly instead of actual deletion.
			// In a real cluster, Kubernetes GC will delete secrets when
			// ClusterSecret is deleted due to the owner reference.

			By("Creating ClusterSecret")
			csec := &clustersecretiov2.ClusterSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name: csecName,
				},
				Spec: clustersecretiov2.ClusterSecretSpec{
					NamespaceSelectorTerms: []clustersecretiov2.NamespaceSelectorTerm{{
						MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{{
							Key:      "cleanup",
							Operator: clustersecretiov2.NamespaceSelectorOpExists,
						}},
					}},
					Template: clustersecretiov2.SecretTemplate{
						Data: map[string][]byte{"key": []byte("value")},
					},
				},
			}
			Expect(k8sClient.Create(ctx, csec)).To(Succeed())

			By("Waiting for secret to be created")
			var secret *corev1.Secret
			Eventually(func() error {
				secret = &corev1.Secret{}
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      csecName,
					Namespace: testNs.Name,
				}, secret)
			}, timeout, interval).Should(Succeed())

			By("Verifying owner reference is set with controller=true for GC")
			var foundOwnerRef bool
			for _, ref := range secret.OwnerReferences {
				if ref.Kind == "ClusterSecret" && ref.Name == csecName {
					Expect(ref.Controller).NotTo(BeNil())
					Expect(*ref.Controller).To(BeTrue(), "Controller should be true for GC")
					foundOwnerRef = true
					break
				}
			}
			Expect(foundOwnerRef).To(BeTrue(), "Owner reference should be set")

			By("Deleting ClusterSecret (GC would delete secrets in real cluster)")
			Expect(k8sClient.Delete(ctx, csec)).To(Succeed())
		})
	})

	Context("When using regex namespace selector", func() {
		var (
			csecName   string
			matchingNs *corev1.Namespace
			otherNs    *corev1.Namespace
		)

		BeforeEach(func() {
			csecName = "test-regex-" + randString(5)

			matchingNs = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "app-production-" + randString(5),
					Labels: map[string]string{"name": "app-production"},
				},
			}
			Expect(k8sClient.Create(ctx, matchingNs)).To(Succeed())

			otherNs = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "app-staging-" + randString(5),
					Labels: map[string]string{"name": "app-staging"},
				},
			}
			Expect(k8sClient.Create(ctx, otherNs)).To(Succeed())
		})

		AfterEach(func() {
			csec := &clustersecretiov2.ClusterSecret{}
			if err := k8sClient.Get(ctx, types.NamespacedName{Name: csecName}, csec); err == nil {
				Expect(k8sClient.Delete(ctx, csec)).To(Succeed())
			}
			for _, ns := range []*corev1.Namespace{matchingNs, otherNs} {
				if ns != nil {
					_ = k8sClient.Delete(ctx, ns)
				}
			}
		})

		It("should match namespaces using regex pattern", func() {
			By("Creating ClusterSecret with InRegex selector")
			csec := &clustersecretiov2.ClusterSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name: csecName,
				},
				Spec: clustersecretiov2.ClusterSecretSpec{
					NamespaceSelectorTerms: []clustersecretiov2.NamespaceSelectorTerm{{
						MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{{
							Key:      "name",
							Operator: clustersecretiov2.NamespaceSelectorOpInRegex,
							Values:   []string{".*production.*"},
						}},
					}},
					Template: clustersecretiov2.SecretTemplate{
						Data: map[string][]byte{"key": []byte("value")},
					},
				},
			}
			Expect(k8sClient.Create(ctx, csec)).To(Succeed())

			By("Verifying secret is created in matching namespace")
			Eventually(func() error {
				secret := &corev1.Secret{}
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      csecName,
					Namespace: matchingNs.Name,
				}, secret)
			}, timeout, interval).Should(Succeed())

			By("Verifying secret is NOT created in non-matching namespace")
			Consistently(func() bool {
				secret := &corev1.Secret{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      csecName,
					Namespace: otherNs.Name,
				}, secret)
				return errors.IsNotFound(err)
			}, time.Second*2, interval).Should(BeTrue())
		})
	})
})

// randString generates a random string of specified length for unique test names
func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}

// Helper to list secrets owned by a ClusterSecret
func listOwnedSecrets(ctx context.Context, c client.Client, csecName string) ([]corev1.Secret, error) {
	var secrets corev1.SecretList
	if err := c.List(ctx, &secrets); err != nil {
		return nil, err
	}

	var owned []corev1.Secret
	for _, sec := range secrets.Items {
		for _, ref := range sec.OwnerReferences {
			if ref.Kind == "ClusterSecret" && ref.Name == csecName {
				owned = append(owned, sec)
				break
			}
		}
	}
	return owned, nil
}
