package controllers

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
	"github.com/komish/cmd-operator-dev/cmdoputils"
	"github.com/komish/cmd-operator-dev/controllers/componentry"

	adregv1 "k8s.io/api/admissionregistration/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var _ = Describe(
	"CertManagerDeployment controller", func() {

		const (
			testNamespace = "test-cmd-operator"

			timeout  = time.Second * 20
			duration = time.Second * 20
			interval = time.Millisecond * 750
		)

		var (
			baseCR = operatorsv1alpha1.CertManagerDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: operatorsv1alpha1.CertManagerDeploymentSpec{
					Version: cmdoputils.GetStringPointer(componentry.CertManagerDefaultVersion),
				},
			}

			controllerOverride = runtime.RawExtension{
				Raw:    []byte(`{"enable-certificate-owner-ref":true}`),
				Object: nil,
			}
			controllerOverrideAsOption = "--enable-certificate-owner-ref=true"

			previousSupportedVersion = "v1.0.3"
		)

		By("creating an instance of the CertManagerDeployment kind", func() {
			Context("with an unsupported version in the spec", func() {

				// Test: an attempt to instantiate a CR with an invalid version should be
				// rejected by the API.
				It("should be rejected by the API", func() {
					cr := baseCR.DeepCopy()
					vers := "v0.0.0" // an invalid version
					cr.Spec.Version = &vers

					Expect(k8sClient.Create(context.TODO(), cr)).ToNot(Succeed())
				})
			})

			Context("with the latest supported version", func() {

				// Test: an instantiated CR should reach a Running phase
				It("should reach status phase running with the default version", func() {
					cr := baseCR.DeepCopy()
					// TODO: We should be creating this one for all of these tests here
					// Need to look into how to do things before a series of its in ginkgo.
					Expect(k8sClient.Create(context.TODO(), cr)).To(Succeed())
					Eventually(func() bool {
						var recv operatorsv1alpha1.CertManagerDeployment
						if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "", Name: "cluster"}, &recv); err != nil {
							return false
						}

						if recv.Status.Phase == string(componentry.StatusPhaseRunning) && recv.Status.Version == componentry.CertManagerDefaultVersion {
							return true
						}

						return false
					}, timeout, interval).Should(BeTrue())
				})

				crds := []string{
					"certificates.cert-manager.io",
					"certificaterequests.cert-manager.io",
					"issuers.cert-manager.io",
					"clusterissuers.cert-manager.io",
					"challenges.acme.cert-manager.io}",
					"orders.acme.cert-manager.io",
				}

				// Test: An instantiated CR should create multiple custom resource definitions.
				for _, crd := range crds {
					It(fmt.Sprintf("should create the custom resource definition: %s", crd), func() {

						Eventually(func() bool {
							var recv apiextv1.CustomResourceDefinition

							if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "", Name: crd}, &recv); err != nil {
								return false
							}

							return true
						}, timeout, interval).Should(BeTrue())

					})
				}

				// Test: An instantiated CR should create a namespace.
				It(fmt.Sprintf("should create the expected namespace: %s", componentry.CertManagerBaseName), func() {

					Eventually(func() bool {
						var recv corev1.Namespace

						if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "", Name: componentry.CertManagerBaseName}, &recv); err != nil {
							return false
						}

						return true
					}, timeout, interval).Should(BeTrue())
				})

				// Test: an instantiated CR should create multiple service accounts.
				for _, sa := range []string{"cert-manager-controller", "cert-manager-cainjector", "cert-manager-webhook"} {
					It(fmt.Sprintf("should create the expected service account: %s", sa), func() {
						Eventually(func() bool {
							var recv corev1.ServiceAccount

							if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: componentry.CertManagerBaseName, Name: sa}, &recv); err != nil {
								return false
							}

							return true
						}, timeout, interval).Should(BeTrue())
					})
				}

				// Test: an instantiated CR should create a mutating webhook configuration.
				It(fmt.Sprintf("Should create the expected mutating webhook configuration: %s", " cert-manager-webhook"), func() {
					Eventually(func() bool {
						var recv adregv1.MutatingWebhookConfiguration

						if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "", Name: "cert-manager-webhook"}, &recv); err != nil {
							return false
						}

						return true
					}, timeout, interval).Should(BeTrue())
				})

				// Test: an instantiated CR should create a validating webhook configuration.
				It(fmt.Sprintf("Should create the expected validating webhook configuration: %s", "cert-manager-webhook"), func() {
					Eventually(func() bool {
						var recv adregv1.ValidatingWebhookConfiguration

						if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "", Name: "cert-manager-webhook"}, &recv); err != nil {
							return false
						}

						return true
					}, timeout, interval).Should(BeTrue())
				})

				// Test: an instantiated CR should create multiple roles.
				for _, role := range []string{"cert-manager-cainjector:leaderelection", "cert-manager-controller:leaderelection", "cert-manager-webhook:dynamic-serving"} {
					It(fmt.Sprintf("should create the expected roles: %s", role), func() {
						Eventually(func() bool {
							var recv rbacv1.Role

							if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: componentry.CertManagerBaseName, Name: role}, &recv); err != nil {
								return false
							}

							return true
						}, timeout, interval).Should(BeTrue())

					})
				}

				// Test: an instantiated CR should create multiple cluster roles.
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
					It(fmt.Sprintf("should create the expected cluster roles: %s", clusterRole), func() {
						Eventually(func() bool {
							var recv rbacv1.ClusterRole

							if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "", Name: clusterRole}, &recv); err != nil {
								return false
							}

							return true
						}, timeout, interval).Should(BeTrue())

					})
				}

				// Test: an instantiated CR should create multiple roles.
				for _, rolebinding := range []string{"cert-manager-cainjector:leaderelection", "cert-manager-controller:leaderelection", "cert-manager-webhook:dynamic-serving"} {
					It(fmt.Sprintf("should create the expected rolebindings: %s", rolebinding), func() {
						Eventually(func() bool {
							var recv rbacv1.RoleBinding

							if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: componentry.CertManagerBaseName, Name: rolebinding}, &recv); err != nil {
								return false
							}

							return true
						}, timeout, interval).Should(BeTrue())

					})
				}

				// Test: an instantiated CR should create multiple cluster role bindings.
				for _, clusterRoleBinding := range []string{
					"cert-manager-cainjector",
					"cert-manager-controller-certificates",
					"cert-manager-controller-challenges",
					"cert-manager-controller-clusterissuers",
					"cert-manager-controller-ingress-shim",
					"cert-manager-controller-issuers",
					"cert-manager-controller-orders",
				} {
					It(fmt.Sprintf("should create the expected clusterrolebindings: %s", clusterRoleBinding), func() {
						Eventually(func() bool {
							var recv rbacv1.ClusterRoleBinding

							if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "", Name: clusterRoleBinding}, &recv); err != nil {
								return false
							}

							return true
						}, timeout, interval).Should(BeTrue())

					})
				}

				// Test: an instantiated CR should create multiple deployments.
				for _, deployment := range []string{"cert-manager-controller", "cert-manager-webhook", "cert-manager-cainjector"} {
					It(fmt.Sprintf("should create the expected deployments: %s", deployment), func() {
						Eventually(func() bool {
							var recv appsv1.Deployment

							if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: componentry.CertManagerBaseName, Name: deployment}, &recv); err != nil {
								return false
							}

							return true
						}, timeout, interval)
					})
				}

				// TODO: Re-organize this to create the CR based on the context, and clean up after each context.

				// TODO: Add additional checks other than "does it exist in the API" for each item

				// TODO: Add fixtures to enable the above
			})

			Context("with argument overrides", func() {
				// Test: a CR instantiated with container overrides should correctly include those overrides
				// in the CR spec.
				It("should persist argument overrides for the major cert-manager components", func() {
					// Get a copy of the base template for the CR we use
					modCR := baseCR.DeepCopy()
					// add a command override
					modCR.Spec.DangerZone.ContainerArgOverrides.Controller = controllerOverride

					var existing operatorsv1alpha1.CertManagerDeployment
					k8sClient.Get(context.TODO(), types.NamespacedName{Name: modCR.GetName()}, &existing)

					// TODO: fix this hack - ideally we need to be recreating the resource as we want, not
					// updating the existing resource.
					modCR.ResourceVersion = existing.ResourceVersion

					// Create the CR with the overrides - if we can't do this, it fails outright.
					if err := k8sClient.Update(context.TODO(), modCR); err != nil {
						Fail(fmt.Sprintf("Unable to post an update to the CR: %s", err))
					}

					var recv operatorsv1alpha1.CertManagerDeployment
					Eventually(func() bool {
						Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: modCR.GetName()}, &recv)).To(Succeed())

						// We expect the override that was persisted to match the override that we submitted
						// this implies that the APIserver didn't truncate anything that we were expecting.
						return reflect.DeepEqual(recv.Spec.DangerZone.ContainerArgOverrides.Controller, controllerOverride)
					}, timeout, interval)
				})

				// Test: a CR instantiated with container overrides should correctly set those overrides on the
				// relevant deployment.
				It("should create deployments with the appropriate overrides where valid", func() {
					// get the deployment that would change as a result of the configuration and see if it has changed.
					var recv appsv1.Deployment
					Eventually(func() bool {
						if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: componentry.CertManagerBaseName, Name: "cert-manager-controller"}, &recv); err != nil {
							return false
						}

						args := recv.Spec.Template.Spec.Containers[0].Args

						for _, arg := range args {
							if arg == controllerOverrideAsOption {
								return true
							}
						}

						return false
					}, timeout, interval)
				})
			})
		})

		// Upgrade scenarios
		By("Updating a CertManagerDeployment custom resource with a new, supported version", func() {
			It("Should redeploy the deployments with the appropriate image specified", func() {
				// Get a copy of the base template for the CR we use
				cr := baseCR.DeepCopy()
				// add a command override
				cr.Spec.Version = &previousSupportedVersion

				var existing operatorsv1alpha1.CertManagerDeployment
				k8sClient.Get(context.TODO(), types.NamespacedName{Name: cr.GetName()}, &existing)

				// TODO: fix this hack - ideally we need to be recreating the resource as we want, not
				// updating the existing resource.
				cr.ResourceVersion = existing.ResourceVersion

				// Create the CR with the overrides - if we can't do this, it fails outright.
				if err := k8sClient.Update(context.TODO(), cr); err != nil {
					Fail(fmt.Sprintf("Unable to post an update to the CR: %s", err))
				}

				for _, deployment := range []string{"cert-manager-controller", "cert-manager-webhook", "cert-manager-cainjector"} {
					Eventually(func() bool {
						var recv appsv1.Deployment
						Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: deployment, Namespace: componentry.CertManagerBaseName}, &recv)).To(Succeed())

						deployedImg := recv.Spec.Template.Spec.Containers[0].Image
						tag := strings.Split(deployedImg, ":")[1]

						if tag == previousSupportedVersion {
							return true
						}

						return false
					}, timeout, interval)
				}
			})
		})

		/* unimplemented
		By("Deleting a CertManagerDeployment custom resource", func() {
			// everything in the namespace should get deleted so we don't check for it here
			It("should delete all namespaced resources", func() {
				By("deleting the namespace", func() {
					Skip("Unimplemented")
				})
			})

			// These don't have owner references so the operator should be deleting these.
			It("should delete the cluster-scoped RBAC resources", func() {
				Skip("Unimplemented")
			})

			// These aren't deleted because the user may still have cert-manager CRDs that are in use
			// and we want them to be able to pull the operator-managed infrastructure without taking
			// down those resources in the process.
			It("should leave cert-manager custom resource definitions", func() {
				Skip("Unimplemented")
			})
		})
		*/
	},
)
