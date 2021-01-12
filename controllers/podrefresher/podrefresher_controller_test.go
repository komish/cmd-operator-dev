package podrefresher

import (
	"context"
	"fmt"

	"github.com/komish/cmd-operator-dev/tests/fixtures"
	"github.com/komish/cmd-operator-dev/tests/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("PodRefresher Controller", func() {
	var (
		key            = types.NamespacedName{Name: fixtures.TestResourceName, Namespace: fixtures.Namespace}
		baseDeployment = appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
				Labels: map[string]string{
					"job-id": fmt.Sprint(identifier),
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "cmd-operator-tests",
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "cmd-operator-tests",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "web",
								Image: "registry.access.redhat.com/ubi8/ubi-minimal",
								Command: []string{
									"/bin/bash",
									"-c",
									"sleep infinity",
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "tls",
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: fixtures.Secret,
									},
								},
							},
						},
					},
				},
			},
		}

		certificateSecretKey = types.NamespacedName{Name: fixtures.Secret, Namespace: fixtures.Namespace}
		secret               = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      certificateSecretKey.Name,
				Namespace: certificateSecretKey.Namespace,
			},
		}

		baseStatefulset = appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
				Labels: map[string]string{
					"job-id": fmt.Sprint(identifier),
				},
			},
			Spec: appsv1.StatefulSetSpec{
				PodManagementPolicy: appsv1.ParallelPodManagement,
				ServiceName:         fixtures.TestResourceName + "-svc",
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "cmd-operator-tests",
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "cmd-operator-tests",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "web",
								Image: "registry.access.redhat.com/ubi8/ubi-minimal",
								Command: []string{
									"/bin/bash",
									"-c",
									"sleep infinity",
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "tls",
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: fixtures.Secret,
									},
								},
							},
						},
					},
				},
			},
		}

		baseDaemonset = appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
				Labels: map[string]string{
					"job-id": fmt.Sprint(identifier),
				},
			},
			Spec: appsv1.DaemonSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "cmd-operator-tests",
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "cmd-operator-tests",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "web",
								Image: "registry.access.redhat.com/ubi8/ubi-minimal",
								Command: []string{
									"/bin/bash",
									"-c",
									"sleep infinity",
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "tls",
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: fixtures.Secret,
									},
								},
							},
						},
					},
				},
			},
		}
	)
	Context("With a deployment", func() {
		BeforeEach(func() {
			// create the deploymnet used for this test
			deployToCreate := baseDeployment.DeepCopy()
			Expect(k8sClient.Create(context.TODO(), deployToCreate)).To(Succeed())

			Eventually(func() appsv1.Deployment {
				var deploy appsv1.Deployment
				Expect(k8sClient.Get(context.TODO(), key, &deploy)).To(Succeed())
				return deploy
			}, fixtures.Timeout, fixtures.Interval).Should(matchers.BeReady())
		})

		AfterEach(func() {
			// clean up the statefulset used for this test.
			deployToDelete := baseDeployment.DeepCopy()
			Expect(k8sClient.Delete(context.TODO(), deployToDelete)).To(Succeed())
		})

		Specify("the podrefresher should properly handle refresher requests when the resource is using a certificate secret", func() {
			var deploy appsv1.Deployment

			By("not opting in to be refreshed", func() {
				Expect(k8sClient.Get(context.TODO(), key, &deploy)).To(Succeed())
				Expect(deploy.Annotations).ToNot(HaveKey(allowRestartAnnotation))
			})

			By("deleting the secret")
			// TODO: make this less destructive

			Eventually(func() error {
				return k8sClient.Delete(context.TODO(), secret)
			}, fixtures.Timeout, fixtures.Interval).Should(Succeed())

			Expect(k8sClient.Get(context.TODO(), key, &deploy)).To(Succeed()) // get a fresh copy after secret modification
			Expect(deploy.Labels).ToNot(HaveKey(timeRestartedLabel))          // check the resource
			Expect(deploy.Annotations).ToNot(HaveKey(secretResourceVersionAnnotation))
			Expect(deploy.Spec.Template.Labels).ToNot(HaveKey(timeRestartedLabel)) // check the resource template

			By("opting in to be refreshed", func() {

				// TODO: update the statefulset to have the opt-in annotation
				var existing appsv1.Deployment
				Expect(k8sClient.Get(context.TODO(), key, &existing)).To(Succeed())

				deployToBeUpdated := existing.DeepCopy()
				if deployToBeUpdated.Annotations == nil {
					deployToBeUpdated.Annotations = make(map[string]string, 0)
				}
				deployToBeUpdated.Annotations[allowRestartAnnotation] = "true"

				Expect(k8sClient.Update(context.TODO(), deployToBeUpdated)).To(Succeed())

				var deploy appsv1.Deployment
				Expect(k8sClient.Get(context.TODO(), key, &deploy)).To(Succeed())
				Expect(deploy.Annotations).To(HaveKey(allowRestartAnnotation))
			})

			By("deleting the secret")
			Eventually(func() error {
				return k8sClient.Delete(context.TODO(), secret)
			}, fixtures.Timeout, fixtures.Interval).Should(Succeed())

			Expect(k8sClient.Get(context.TODO(), key, &deploy)).To(Succeed()) // get a fresh copy after secret modification
			Expect(deploy.Labels).To(HaveKey(timeRestartedLabel))             // check the resource
			Expect(deploy.Annotations).To(HaveKey(secretResourceVersionAnnotation))
			Expect(deploy.Spec.Template.Labels).To(HaveKey(timeRestartedLabel)) // check the resource template
		})
	})

	Context("With a statefulset", func() {
		BeforeEach(func() {
			// create the statefulset used for this test
			stsToCreate := baseStatefulset.DeepCopy()
			Expect(k8sClient.Create(context.TODO(), stsToCreate)).To(Succeed())

			Eventually(func() appsv1.StatefulSet {
				var sts appsv1.StatefulSet
				Expect(k8sClient.Get(context.TODO(), key, &sts)).To(Succeed())
				return sts
			}, fixtures.Timeout, fixtures.Interval).Should(matchers.BeReadyAndOnTheLatestRevision())
		})

		AfterEach(func() {
			// clean up the statefulset used for this test.
			stsToDelete := baseStatefulset.DeepCopy()
			Expect(k8sClient.Delete(context.TODO(), stsToDelete)).To(Succeed())
		})

		Specify("the podrefresher should properly handle refresher requests when the resource is using a certificate secret", func() {
			var sts appsv1.StatefulSet

			By("not opting in to be refreshed", func() {
				Expect(k8sClient.Get(context.TODO(), key, &sts)).To(Succeed())
				Expect(sts.Annotations).ToNot(HaveKey(allowRestartAnnotation))
			})

			By("deleting the secret")
			// TODO: make this less destructive

			Eventually(func() error {
				return k8sClient.Delete(context.TODO(), secret)
			}, fixtures.Timeout, fixtures.Interval).Should(Succeed())

			Expect(k8sClient.Get(context.TODO(), key, &sts)).To(Succeed()) // get a fresh copy after secret modification
			Expect(sts.Labels).ToNot(HaveKey(timeRestartedLabel))          // check the resource
			Expect(sts.Annotations).ToNot(HaveKey(secretResourceVersionAnnotation))
			Expect(sts.Spec.Template.Labels).ToNot(HaveKey(timeRestartedLabel)) // check the resource template

			By("opting in to be refreshed", func() {

				// TODO: update the statefulset to have the opt-in annotation
				var existing appsv1.StatefulSet
				Expect(k8sClient.Get(context.TODO(), key, &existing)).To(Succeed())

				stsToBeUpdated := existing.DeepCopy()
				if stsToBeUpdated.Annotations == nil {
					stsToBeUpdated.Annotations = make(map[string]string, 0)
				}
				stsToBeUpdated.Annotations[allowRestartAnnotation] = "true"

				Expect(k8sClient.Update(context.TODO(), stsToBeUpdated)).To(Succeed())

				var sts appsv1.StatefulSet
				Expect(k8sClient.Get(context.TODO(), key, &sts)).To(Succeed())
				Expect(sts.Annotations).To(HaveKey(allowRestartAnnotation))

			})

			By("deleting the secret")
			Eventually(func() error {
				return k8sClient.Delete(context.TODO(), secret)
			}, fixtures.Timeout, fixtures.Interval).Should(Succeed())

			Expect(k8sClient.Get(context.TODO(), key, &sts)).To(Succeed()) // get a fresh copy after secret modification
			Expect(sts.Labels).To(HaveKey(timeRestartedLabel))             // check the resource
			Expect(sts.Annotations).To(HaveKey(secretResourceVersionAnnotation))
			Expect(sts.Spec.Template.Labels).To(HaveKey(timeRestartedLabel)) // check the resource template
		})
	})

	Context("With a daemonset", func() {
		BeforeEach(func() {
			// create the daemonset used for this test
			dsetToCreate := baseDaemonset.DeepCopy()
			Expect(k8sClient.Create(context.TODO(), dsetToCreate)).To(Succeed())

			Eventually(func() appsv1.DaemonSet {
				var dset appsv1.DaemonSet
				Expect(k8sClient.Get(context.TODO(), key, &dset)).To(Succeed())
				return dset
			}, fixtures.Timeout, fixtures.Interval).Should(matchers.BeReady())
		})

		AfterEach(func() {
			// clean up the daemonset used for this test.
			dsetToDelete := baseDaemonset.DeepCopy()
			Expect(k8sClient.Delete(context.TODO(), dsetToDelete)).To(Succeed())
		})

		Specify("the podrefresher should properly handle refresher requests when the resource is using a certificate secret", func() {
			var dset appsv1.DaemonSet

			By("not opting in to be refreshed", func() {
				Expect(k8sClient.Get(context.TODO(), key, &dset)).To(Succeed())
				Expect(dset.Annotations).ToNot(HaveKey(allowRestartAnnotation))
			})

			By("deleting the secret")
			// TODO: make this less destructive

			Eventually(func() error {
				return k8sClient.Delete(context.TODO(), secret)
			}, fixtures.Timeout, fixtures.Interval).Should(Succeed())

			Expect(k8sClient.Get(context.TODO(), key, &dset)).To(Succeed()) // get a fresh copy after secret modification
			Expect(dset.Labels).ToNot(HaveKey(timeRestartedLabel))          // check the resource
			Expect(dset.Annotations).ToNot(HaveKey(secretResourceVersionAnnotation))
			Expect(dset.Spec.Template.Labels).ToNot(HaveKey(timeRestartedLabel)) // check the resource template

			By("opting in to be refreshed", func() {
				var existing appsv1.DaemonSet
				Expect(k8sClient.Get(context.TODO(), key, &existing)).To(Succeed())

				dsetToBeUpdated := existing.DeepCopy()
				if dsetToBeUpdated.Annotations == nil {
					dsetToBeUpdated.Annotations = make(map[string]string, 0)
				}
				dsetToBeUpdated.Annotations[allowRestartAnnotation] = "true"

				Expect(k8sClient.Update(context.TODO(), dsetToBeUpdated)).To(Succeed())

				var dset appsv1.DaemonSet
				Expect(k8sClient.Get(context.TODO(), key, &dset)).To(Succeed())
				Expect(dset.Annotations).To(HaveKey(allowRestartAnnotation))
			})

			By("deleting the secret")
			Eventually(func() error {
				return k8sClient.Delete(context.TODO(), secret)
			}, fixtures.Timeout, fixtures.Interval).Should(Succeed())

			Expect(k8sClient.Get(context.TODO(), key, &dset)).To(Succeed()) // get a fresh copy after secret modification
			Expect(dset.Labels).To(HaveKey(timeRestartedLabel))             // check the resource
			Expect(dset.Annotations).To(HaveKey(secretResourceVersionAnnotation))
			Expect(dset.Spec.Template.Labels).To(HaveKey(timeRestartedLabel)) // check the statefulset template
		})
	})
})
