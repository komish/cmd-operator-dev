package certmanagerdeployment

import (
	"context"
	"fmt"
	"strings"

	"github.com/komish/cmd-operator-dev/cmdoputils"
	"github.com/komish/cmd-operator-dev/tests/fixtures"
	adregv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
	"github.com/komish/cmd-operator-dev/controllers/componentry"
)

var (
	key                        types.NamespacedName
	baseCR                     operatorsv1alpha1.CertManagerDeployment
	controllerOverride         runtime.RawExtension
	crds                       []string
	controllerOverrideAsOption = "--enable-certificate-owner-ref=true"
)

var _ = Describe(
	"CertManagerDeployment controller", func() {

		BeforeEach(func() {
			// initialize bases with fresh state
			key = types.NamespacedName{
				Name: "cluster",
			}

			baseCR = operatorsv1alpha1.CertManagerDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: key.Name,
				},
				Spec: operatorsv1alpha1.CertManagerDeploymentSpec{
					Version: cmdoputils.GetStringPointer(componentry.CertManagerDefaultVersion),
				},
			}

			controllerOverride = runtime.RawExtension{
				Raw:    []byte(`{"enable-certificate-owner-ref":true}`),
				Object: nil,
			}

			crds = []string{
				"certificates.cert-manager.io",
				"certificaterequests.cert-manager.io",
				"issuers.cert-manager.io",
				"clusterissuers.cert-manager.io",
				"challenges.acme.cert-manager.io",
				"orders.acme.cert-manager.io",
			}
		})

		BeforeEach(func() {
			// we clean up after every spec, wait until the namespace is gone
			// before we move forward. If it exists unrelated to our tests,
			// this should stop the test just in case we're in an environment with
			// an existing cluster unrelated to our testing
			Eventually(func() bool {
				var ns corev1.Namespace
				err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: "cert-manager"}, &ns)

				return apierrors.IsNotFound(err)
			}, fixtures.Timeout, fixtures.Interval).Should(BeTrue())
		})

		AfterEach(func() {
			// clean up cr
			crToDelete := baseCR.DeepCopy()
			err := k8sClient.Delete(context.TODO(), crToDelete)
			if err != nil {
				Expect(apierrors.IsNotFound(err)).To(BeTrue(), "resource IsNotFound is acceptable during clean up steps")
			} else {
				Expect(err).To(BeNil())
			}
		})

		Context("creating an cert manager deployment", func() {
			Context("with the most recent supported \"y-1\" version", func() {
				It("should properly persist and be upgradeable", func() {
					crToCreate := baseCR.DeepCopy()
					crToCreate.Spec.Version = &fixtures.PreviousSupportedVersion
					Expect(k8sClient.Create(context.TODO(), crToCreate)).To(Succeed())

					By("rolling out successfully")
					Eventually(func() bool {
						var recv operatorsv1alpha1.CertManagerDeployment
						if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "", Name: "cluster"}, &recv); err != nil {
							return false
						}

						// TODO: write a matcher
						if recv.Status.Phase == string(componentry.StatusPhaseRunning) && recv.Status.Version == fixtures.PreviousSupportedVersion {
							return true
						}

						return false
					}, fixtures.Timeout, fixtures.Interval).Should(BeTrue())

					for _, deployment := range []string{"cert-manager-controller", "cert-manager-webhook", "cert-manager-cainjector"} {
						By(fmt.Sprintf("configuring the deployment to use the y-1 image for component: %s", deployment))
						Eventually(func() string {
							var recv appsv1.Deployment
							Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: deployment, Namespace: componentry.CertManagerBaseName}, &recv)).To(Succeed())

							deployedImg := recv.Spec.Template.Spec.Containers[0].Image
							tag := strings.Split(deployedImg, ":")[1]

							return tag
						}, fixtures.Timeout, fixtures.Interval).Should(Equal(fixtures.PreviousSupportedVersion))
					}

					By("successfully updating to the latest version")
					var existingCR operatorsv1alpha1.CertManagerDeployment
					Expect(k8sClient.Get(context.TODO(), key, &existingCR)).To(Succeed())
					Expect(*existingCR.Spec.Version).To(Equal(fixtures.PreviousSupportedVersion))

					// unset the version so the controller installs the latest
					crToUpdate := existingCR.DeepCopy()
					crToUpdate.Spec.Version = cmdoputils.GetStringPointer(componentry.CertManagerDefaultVersion)
					Eventually(func() error {
						return k8sClient.Update(context.TODO(), crToUpdate)
					}).Should(Succeed())

					By("reaching a status phase \"running\" after an update")
					Eventually(func() bool {
						var recv operatorsv1alpha1.CertManagerDeployment
						if err := k8sClient.Get(context.TODO(), key, &recv); err != nil {
							return false
						}

						return recv.Status.Phase == string(componentry.StatusPhaseRunning)
					}, fixtures.Timeout, fixtures.Interval).Should(BeTrue())
				})
			})

			Context("with an argument override", func() {
				It("should persist a valid argument override for the given cert-manager component", func() {
					// add command override to the base
					modifiedCRToCreate := baseCR.DeepCopy()
					modifiedCRToCreate.Spec.DangerZone.ContainerArgOverrides.Controller = controllerOverride
					Expect(k8sClient.Create(context.TODO(), modifiedCRToCreate)).To(Succeed())

					// check the deployment for the persisted option
					By("persisting the option on the CertManagerDeployment itself")
					var createdCMD operatorsv1alpha1.CertManagerDeployment
					Eventually(func() runtime.RawExtension {
						Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: modifiedCRToCreate.GetName()}, &createdCMD)).To(Succeed())
						return createdCMD.Spec.DangerZone.ContainerArgOverrides.Controller
					}, fixtures.Timeout, fixtures.Interval).Should(Equal(controllerOverride), "ensure argument overrides persist to related deployments")

					By("persisting the option on the component deployment itself")
					Eventually(func() []string {
						var createdDep appsv1.Deployment
						result := make([]string, 0)
						if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: componentry.CertManagerBaseName, Name: "cert-manager-controller"}, &createdDep); err != nil {
							return result
						}

						return createdDep.Spec.Template.Spec.Containers[0].Args
					}, fixtures.Timeout, fixtures.Interval).Should(ContainElement(controllerOverrideAsOption))
				})
			})

			Context("with an invalid spec", func() {
				It("should be rejected by the API", func() {
					By("having an unsupported version in the spec", func() {
						cr := baseCR.DeepCopy()
						vers := "v0.0.0" // an invalid version
						cr.Spec.Version = &vers

						Expect(k8sClient.Create(context.TODO(), cr)).ToNot(Succeed())
					})
				})
			})

			Context("With a valid spec", func() {
				It("should be accepted by the API and deploy required resources", func() {
					cr := baseCR.DeepCopy()
					Expect(k8sClient.Create(context.TODO(), cr)).To(Succeed())

					// crd created
					for _, crd := range crds {
						By(fmt.Sprintf("creating the custom resource definition: %s", crd))
						Eventually(func() bool {
							var recv apiextv1.CustomResourceDefinition

							if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "", Name: crd}, &recv); err != nil {
								return false
							}

							return true
						}, fixtures.Timeout, fixtures.Interval).Should(BeTrue())
					}

					// namespace
					By(fmt.Sprintf("creating the expected namespace: %s", componentry.CertManagerBaseName), func() {
						Eventually(func() bool {
							var recv corev1.Namespace
							if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "", Name: componentry.CertManagerBaseName}, &recv); err != nil {
								return false
							}
							return true
						}, fixtures.Timeout, fixtures.Interval).Should(BeTrue())
					})

					// service account
					for _, sa := range []string{"cert-manager", "cert-manager-cainjector", "cert-manager-webhook"} {
						By(fmt.Sprintf("creating the expected service account: %s", sa), func() {
							Eventually(func() bool {
								var recv corev1.ServiceAccount
								if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: componentry.CertManagerBaseName, Name: sa}, &recv); err != nil {
									return false
								}
								return true
							}, fixtures.Timeout, fixtures.Interval).Should(BeTrue())
						})
					}

					// mutating webhook configuration
					By("creating the expected mutating webhook configuration: cert-manager-webhook", func() {
						Eventually(func() bool {
							var recv adregv1.MutatingWebhookConfiguration

							if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "", Name: "cert-manager-webhook"}, &recv); err != nil {
								return false
							}

							return true
						}, fixtures.Timeout, fixtures.Interval).Should(BeTrue())
					})

					// validating webhook configuration
					By(fmt.Sprintf("creating the expected validating webhook configuration: %s", "cert-manager-webhook"), func() {
						Eventually(func() bool {
							var recv adregv1.ValidatingWebhookConfiguration

							if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "", Name: "cert-manager-webhook"}, &recv); err != nil {
								return false
							}

							return true
						}, fixtures.Timeout, fixtures.Interval).Should(BeTrue())
					})

					// roles
					for _, role := range []string{"cert-manager-cainjector:leaderelection", "cert-manager-controller:leaderelection", "cert-manager-webhook:dynamic-serving"} {
						By(fmt.Sprintf("creating the expected role: %s", role), func() {
							Eventually(func() bool {
								var recv rbacv1.Role

								if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: componentry.CertManagerBaseName, Name: role}, &recv); err != nil {
									return false
								}

								return true
							}, fixtures.Timeout, fixtures.Interval).Should(BeTrue())

						})
					}

					// cluster roles
					for _, clusterRole := range []string{
						"cert-manager-cainjector",
						"cert-manager-controller-certificates",
						"cert-manager-controller-challenges",
						"cert-manager-controller-clusterissuers",
						"cert-manager-controller-ingress-shim",
						"cert-manager-controller-issuers",
						"cert-manager-controller-orders",
						"cert-manager-edit",
						"cert-manager-view"} {
						By(fmt.Sprintf("creating the expected cluster role: %s", clusterRole), func() {
							Eventually(func() bool {
								var recv rbacv1.ClusterRole

								if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "", Name: clusterRole}, &recv); err != nil {
									return false
								}

								return true
							}, fixtures.Timeout, fixtures.Interval).Should(BeTrue())

						})
					}

					// role bindings
					for _, rolebinding := range []string{"cert-manager-cainjector:leaderelection", "cert-manager-controller:leaderelection", "cert-manager-webhook:dynamic-serving"} {
						By(fmt.Sprintf("creating the expected rolebinding: %s", rolebinding), func() {
							Eventually(func() bool {
								var recv rbacv1.RoleBinding

								if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: componentry.CertManagerBaseName, Name: rolebinding}, &recv); err != nil {
									return false
								}

								return true
							}, fixtures.Timeout, fixtures.Interval).Should(BeTrue())

						})
					}

					// cluster role bindings
					for _, clusterRoleBinding := range []string{
						"cert-manager-cainjector",
						"cert-manager-controller-certificates",
						"cert-manager-controller-challenges",
						"cert-manager-controller-clusterissuers",
						"cert-manager-controller-ingress-shim",
						"cert-manager-controller-issuers",
						"cert-manager-controller-orders",
					} {
						By(fmt.Sprintf("creating the expected clusterrolebinding: %s", clusterRoleBinding), func() {
							Eventually(func() bool {
								var recv rbacv1.ClusterRoleBinding

								if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "", Name: clusterRoleBinding}, &recv); err != nil {
									return false
								}

								return true
							}, fixtures.Timeout, fixtures.Interval).Should(BeTrue())

						})
					}

					// deployments
					for _, deployment := range []string{"cert-manager-controller", "cert-manager-webhook", "cert-manager-cainjector"} {
						By(fmt.Sprintf("creating the expected deployment: %s", deployment), func() {
							Eventually(func() bool {
								var recv appsv1.Deployment

								if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: componentry.CertManagerBaseName, Name: deployment}, &recv); err != nil {
									return false
								}

								return true
							}, fixtures.Timeout, fixtures.Interval)
						})
					}

					// status
					By("reaching a status phase running")
					Eventually(func() bool {
						var recv operatorsv1alpha1.CertManagerDeployment
						if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "", Name: "cluster"}, &recv); err != nil {
							return false
						}

						// TODO write a matcher for this
						if recv.Status.Phase == string(componentry.StatusPhaseRunning) && recv.Status.Version == componentry.CertManagerDefaultVersion {
							return true
						}

						return false
					}, fixtures.Timeout, fixtures.Interval).Should(BeTrue())

					// TODO: Re-organize this to create the CR based on the context, and clean up after each context.
					// TODO: Add additional checks other than "does it exist in the API" for each item
				})
			})
		})
	})
