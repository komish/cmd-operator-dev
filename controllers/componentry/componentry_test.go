package componentry

import (
	"sort"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/komish/cmd-operator-dev/tests/fixtures"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Componentry", func() {
	Context("Supported Versions", func() {
		It("Should always contain the default (latest) version", func() {
			_, ok := SupportedVersions[CertManagerDefaultVersion]
			Expect(ok).To(BeTrue())
		})

		It("Should always contain the previous supported version", func() {
			_, ok := SupportedVersions[fixtures.PreviousSupportedVersion]
			Expect(ok).To(BeTrue())
		})
	})

	Context("The component list", func() {
		It("Contains component getter functions for the three major components", func() {
			result := []string{}
			expected := []string{"cainjector", "controller", "webhook"} // make sure this is sorted!

			for i := 0; i <= 2; i++ {
				componentGetterFunc := Components[i]
				comp := componentGetterFunc(CertManagerDefaultVersion)
				result = append(result, comp.GetName())
			}

			sort.Strings(result)
			Expect(result).To(Equal(expected))
		})
	})

	Context("The controller component getter function", func() {
		It("Properly persists version changes", func() {
			By("changing the container image tag when passed a supported container version other than the default", func() {
				controller := GetComponentForController(fixtures.PreviousSupportedVersion)
				containers := controller.getContainers()
				container := containers[0]
				image := container.Image

				Expect(image).To(Equal("quay.io/jetstack/cert-manager-controller:" + fixtures.PreviousSupportedVersion))
			})
		})
	})

	Context("The cainjector component getter function", func() {
		It("Properly persists version changes", func() {
			By("changing the container image tag when passed a supported container version other than the default", func() {
				cainjector := GetComponentForCAInjector(fixtures.PreviousSupportedVersion)
				containers := cainjector.getContainers()
				container := containers[0]
				image := container.Image

				Expect(image).To(Equal("quay.io/jetstack/cert-manager-cainjector:" + fixtures.PreviousSupportedVersion))
			})
		})
	})

	Context("The webhook component getter function", func() {
		It("Properly persists version changes", func() {
			By("changing the container image tag when passed a supported container version other than the default", func() {
				webhook := GetComponentForWebhook(fixtures.PreviousSupportedVersion)
				containers := webhook.getContainers()
				container := containers[0]
				image := container.Image

				Expect(image).To(Equal("quay.io/jetstack/cert-manager-webhook:" + fixtures.PreviousSupportedVersion))
			})
		})
	})
})

var _ = Describe("CertManagerComponent", func() {
	Context("The CertManagerComponent specification", func() {
		testInstanceName := "foo-instance"
		testName := "foo"
		testServiceAccountName := "foo-service-account"
		testLabelKey := "foo"
		testLabelVal := "bar"
		testDeploymentReplicas := int32(99)
		testServiceClusterIP := "127.0.0.1"
		testRoleData := RoleData{
			name: "foo",
		}
		testClusterRoleData := RoleData{
			name: "fooCluster",
		}
		testWebhookData := WebhookData{}

		component := CertManagerComponent{
			name:               testName,
			serviceAccountName: testServiceAccountName,
			labels: map[string]string{
				testLabelKey: testLabelVal,
			},
			clusterRoles: []RoleData{testClusterRoleData},
			roles:        []RoleData{testRoleData},
			deployment: appsv1.DeploymentSpec{
				Replicas: &testDeploymentReplicas,
			},
			service:  corev1.ServiceSpec{ClusterIP: testServiceClusterIP},
			webhooks: []WebhookData{testWebhookData}}

		It("Should expose struct key \"name\" via its getter method", func() {
			name := component.GetName()
			Expect(name).To(BeIdenticalTo(testName))
		})
		It("Should expose struct key \"serviceAccountName\" via its getter method", func() {
			serviceAccount := component.GetServiceAccountName()
			Expect(serviceAccount).To(BeIdenticalTo(testServiceAccountName))
		})
		It("Should expose struct key \"labels\" via its getter method", func() {
			labels := component.GetLabels()
			Expect(labels[testLabelKey]).To(BeIdenticalTo(testLabelVal))
		})
		It("Should expose struct key \"clusterRoles\" via its getter method", func() {
			clusterRoles := component.GetClusterRoles()
			Expect(clusterRoles[0].GetName()).To(Equal(testClusterRoleData.name))
		})
		It("Should expose struct key \"roles\" via its getter method", func() {
			roles := component.GetRoles()
			Expect(roles[0].GetName()).To(BeIdenticalTo(testRoleData.name))
		})
		It("Should expose struct key \"deployment\" via its getter method", func() {
			deployments := component.GetDeployment()
			Expect(deployments.Replicas).To(Equal(&testDeploymentReplicas))
		})
		It("Should expose struct key \"service\" via its getter method", func() {
			service := component.GetService()
			Expect(service.ClusterIP).To(BeIdenticalTo(testServiceClusterIP))
		})
		It("Should expose struct key \"webhooks\" via its getter method", func() {
			webhook := component.GetWebhooks()
			Expect(webhook[0].name).To(BeIdenticalTo(testWebhookData.name))
		})
		It("Should provide access to a uniformed base set of label selectors", func() {
			componentSelectorKey := "app.kubernetes.io/component"
			nameSelectorKey := "app.kubernetes.io/name"
			selectors := component.GetBaseLabelSelector()
			Expect(selectors.MatchLabels[componentSelectorKey]).To(Equal(component.GetName()))
			Expect(selectors.MatchLabels[nameSelectorKey]).To(BeIdenticalTo(component.GetName()))
		})

		It("Should provide a name formatted as a kubernetes resource", func() {
			resourceName := component.GetResourceName()
			Expect(resourceName).To(BeIdenticalTo(CertManagerBaseName + "-" + component.GetName()))
		})

		It("Should provide access to the labels with the additional instance key", func() {
			instanceLabels := component.GetLabelsWithInstanceName(testInstanceName)
			key := "app.kubernetes.io/instance"
			Expect(instanceLabels[key]).To(BeIdenticalTo(testInstanceName))
		})
	})

	Context("The CertManagerComponent structure", func() {
		testContainer := corev1.Container{
			Name: "foo-container",
		}

		component := CertManagerComponent{
			deployment: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{testContainer},
					},
				},
			},
		}

		It("Should allow you to get containers for a component's deployment", func() {
			containers := component.getContainers()
			Expect(containers[0].Name).To(BeIdenticalTo(testContainer.Name))
		})
	})
})
