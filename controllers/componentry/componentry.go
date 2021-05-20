// Package componentry contains metadata and types relevant to the various cert-manager components.
package componentry

import (
	"fmt"
	"strings"

	"github.com/komish/cmd-operator-dev/cmdoputils"
	adregv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// oneReplica is a value of 1 of type int32 to be used
	// as the appsv1.DeploymentSpec.Replicas struct key which
	// requires a typed int32 pointer.
	oneReplica = int32(1)

	// tenSeconds is a value of 10 of type int32 to be used for
	// webhook timeouts which requires an int32 pointer.
	tenSeconds = int32(10)

	// StandardLabels are the base labels that apply to all CertManagerDeployment-managed resources.
	StandardLabels = map[string]string{
		"app":                          "cert-manager",
		"app.kubernetes.io/managed-by": "operator",
	}

	// InstanceLabelKey is a basic label key used to associate an owner object's name to a resource using a label.
	InstanceLabelKey = "app.kubernetes.io/instance"

	// StandardListOptions is a standardized set of filters to use to list components that should be managed
	// by this operator. This is for use with cluster-scoped resource lists.
	StandardListOptions = []client.ListOption{
		client.MatchingLabels(StandardLabels),
	}
	// StandardListOptionsWithNamespace is a standardized set of filters to use to list components that should be managed
	// by this operator. This is for use with namespace-scoped resource lists.
	StandardListOptionsWithNamespace = []client.ListOption{
		client.InNamespace(CertManagerDeploymentNamespace),
		client.MatchingLabels(StandardLabels),
	}

	// SupportedVersions represents the versions of Cert-Manager that are supported by the operator.
	// The value is irrelevant. Only the keys are used for lookup.
	//
	// Changes to these values also need to be mirrored in the CertManageDeploymentSpec generation
	// validation annotations.
	SupportedVersions = map[string]bool{
		"v1.2.0": true,
		"v1.3.0": true,
		"v1.3.1": true, // latest
	}

	// Components are all ComponentGetterFunctions, one per Component, that we need
	// to deploy and manage as a part of a CertManagerDeployment.
	Components = []ComponentGetterFunction{GetComponentForController, GetComponentForCAInjector, GetComponentForWebhook}
)

// ComponentGetterFunction is a function that will return a base CertManagerComponent.
type ComponentGetterFunction func(string) CertManagerComponent

// CertManagerComponent represents the various components of cert-manager that might
// be installed in a cluster.
type CertManagerComponent struct {
	name               string
	serviceAccountName string
	labels             map[string]string
	clusterRoles       []RoleData
	roles              []RoleData
	deployment         appsv1.DeploymentSpec
	service            corev1.ServiceSpec
	webhooks           []WebhookData
}

// GetName returns CertManagerComponent name in  lower case.
func (comp *CertManagerComponent) GetName() string {
	return comp.name
}

// GetServiceAccountName returns the service account name used for the CertManagerComponent.
func (comp *CertManagerComponent) GetServiceAccountName() string {
	return comp.serviceAccountName
}

// GetLabels returns a base set of labels expected to be used by the CertManagerComponent.
func (comp *CertManagerComponent) GetLabels() map[string]string {
	return comp.labels
}

// GetLabelsWithInstanceName will return the default labels for a given component as well the app.kubernetes.io/instance label
// set with the provided name as the value.
func (comp *CertManagerComponent) GetLabelsWithInstanceName(name string) map[string]string {
	lbls := comp.GetLabels()
	lbls[InstanceLabelKey] = name
	return lbls
}

// GetClusterRoles returns all cluster roles that need to be created for the CertManagerComponent.
func (comp *CertManagerComponent) GetClusterRoles() []RoleData {
	return comp.clusterRoles
}

// GetRoles returns all the roles that need to be created for the CertManagerComponent.
func (comp *CertManagerComponent) GetRoles() []RoleData {
	return comp.roles
}

// GetDeployment returns the deployment spec that needs to be created for the CertManageComponent.
func (comp *CertManagerComponent) GetDeployment() appsv1.DeploymentSpec {
	return comp.deployment
}

// GetService returns the service spec that needs to be created for the CertManageComponent.
func (comp *CertManagerComponent) GetService() corev1.ServiceSpec {
	return comp.service
}

// GetWebhooks returns the webhooks that need to be created for the CertManagerComponent
func (comp *CertManagerComponent) GetWebhooks() []WebhookData {
	return comp.webhooks
}

// GetBaseLabelSelector returns label selectors using metadatda available on the
// CertManagerComponent as values.
// TODO(komish): This returns a LabelSelector for a component, but doesn't use the
// component's own pre-defined Labels struct key. Need to check what's using this
// and if we can just derive the LabelSelector from the Labels that are already
// on the component.
func (comp *CertManagerComponent) GetBaseLabelSelector() *metav1.LabelSelector {
	var ls metav1.LabelSelector
	metav1.AddLabelToSelector(&ls, "app.kubernetes.io/component", comp.name)
	metav1.AddLabelToSelector(&ls, "app.kubernetes.io/name", comp.name)
	return &ls
}

// GetResourceName will return a hyphenation of the standard base name and the component name.
func (comp *CertManagerComponent) GetResourceName() string {
	return strings.Join([]string{CertManagerBaseName, comp.name}, "-")
}

// getContainers returns the CertManageComponent.deployment.containers
func (comp *CertManagerComponent) getContainers() []corev1.Container {
	return comp.deployment.Template.Spec.Containers
}

// GetComponentForController returns a CertManagerComponent containing
// all the metadata necessary to deploy the subresources needed to run
// the cert-manager controller.
func GetComponentForController(version string) CertManagerComponent {
	// Component and all resources for the latest supported version of cert-manager controller
	// that this operator supports.
	var comp = CertManagerComponent{
		name:               "controller",
		serviceAccountName: "cert-manager",
		labels: cmdoputils.MergeMaps(map[string]string{
			"app.kubernetes.io/component": "controller",
			"app.kubernetes.io/name":      "controller",
		}, StandardLabels),
		clusterRoles: []RoleData{
			clusterRoleDataForClusterIssuers,
			clusterRoleDataForIssuers,
			clusterRoleDataForChallenges,
			clusterRoleDataForEdit,
			clusterRoleDataForIngressShim,
			clusterRoleDataForOrders,
			clusterRoleDataForCertificates,
			clusterRoleDataForView,
			clusterRoleDataForApprover,
		},
		roles: []RoleData{roleForController},
		deployment: appsv1.DeploymentSpec{
			Replicas: &oneReplica,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"prometheus.io/path":   "/metrics",
						"prometheus.io/port":   "9402",
						"prometheus.io/scrape": "true",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "cert-manager",
							Args: []string{}, // container args are generated
							Env: []corev1.EnvVar{
								{
									Name: "POD_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
							Image:           "quay.io/jetstack/cert-manager-controller:v1.3.1",
							ImagePullPolicy: "IfNotPresent",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 9402,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
		service: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       9402,
					TargetPort: intstr.FromInt(9402),
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
		webhooks: []WebhookData{},
	}

	// handle other supported versions
	switch version {
	case "v1.3.0":
		{
			// Image changes for this version.
			container := &comp.deployment.Template.Spec.Containers[0] // we assume one container
			container.Image = fmt.Sprintf("quay.io/jetstack/cert-manager-controller:%s", version)
		}
	case "v1.2.0":
		{
			// Image changes for this version.
			container := &comp.deployment.Template.Spec.Containers[0] // we assume one container
			container.Image = fmt.Sprintf("quay.io/jetstack/cert-manager-controller:%s", version)

			// The cert-manager-edit role is different for v1.2.0 so we
			// replace the latest representation with our new representation
			clusterRoleDataForEditv1_2_0 := RoleData{
				name:        "cert-manager-edit",
				isAggregate: true,
				labels: map[string]string{
					"rbac.authorization.k8s.io/aggregate-to-admin": "true",
					"rbac.authorization.k8s.io/aggregate-to-edit":  "true",
				},
				policyRules: []rbacv1.PolicyRule{
					{
						APIGroups: []string{"cert-manager.io"},
						Resources: []string{"certificates", "certificaterequests", "issuers"},
						Verbs:     []string{"create", "delete", "deletecollection", "patch", "update"},
					},
					{
						APIGroups: []string{"acme.cert-manager.io"},
						Resources: []string{"challenges", "orders"},
						Verbs:     []string{"get", "list", "watch"},
					},
				},
			}

			comp.clusterRoles = []RoleData{
				clusterRoleDataForClusterIssuers,
				clusterRoleDataForIssuers,
				clusterRoleDataForChallenges,
				clusterRoleDataForEditv1_2_0,
				clusterRoleDataForIngressShim,
				clusterRoleDataForOrders,
				clusterRoleDataForCertificates,
				clusterRoleDataForView,
				clusterRoleDataForApprover,
			}
		}
	}

	return comp
}

// GetComponentForCAInjector returns a CetManagerComponent containing
// all the metadata necessary to deploy the subresources needed to run
// the cert-manager cainjector.
func GetComponentForCAInjector(version string) CertManagerComponent {
	comp := CertManagerComponent{
		name:               "cainjector",
		serviceAccountName: "cert-manager-cainjector",
		labels: cmdoputils.MergeMaps(map[string]string{
			"app.kubernetes.io/component": "cainjector",
			"app.kubernetes.io/name":      "cainjector",
		}, StandardLabels),
		clusterRoles: []RoleData{clusterRoleDataForCAInjector},
		roles:        []RoleData{roleForCAInjectorLeaderElection},
		deployment: appsv1.DeploymentSpec{
			Replicas: &oneReplica,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "cert-manager",
							Args: []string{}, // container args are generated
							Env: []corev1.EnvVar{
								{
									Name: "POD_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
							Image:           "quay.io/jetstack/cert-manager-cainjector:v1.3.1",
							ImagePullPolicy: "IfNotPresent",
						},
					},
				},
			},
		},
		webhooks: []WebhookData{},
	}

	switch version {
	case "v1.3.0":
		{
			container := &comp.deployment.Template.Spec.Containers[0]
			container.Image = fmt.Sprintf("quay.io/jetstack/cert-manager-cainjector:%s", version)
		}
	case "v1.2.0":
		{
			container := &comp.deployment.Template.Spec.Containers[0]
			container.Image = fmt.Sprintf("quay.io/jetstack/cert-manager-cainjector:%s", version)
		}
	}
	return comp
}

// GetComponentForWebhook returns a CertManagerComponent containing
// all the metadata necessary to deploy the subresources needed to run
// the cert-manager webhook.
func GetComponentForWebhook(version string) CertManagerComponent {
	// This is for use with webhook objects.
	var failPolicy adregv1.FailurePolicyType = "Fail"
	var noneSideEffect adregv1.SideEffectClass = "None"

	comp := CertManagerComponent{
		name:               "webhook",
		serviceAccountName: "cert-manager-webhook",
		labels: cmdoputils.MergeMaps(map[string]string{
			"app.kubernetes.io/component": "webhook",
			"app.kubernetes.io/name":      "webhook",
		}, StandardLabels),
		clusterRoles: []RoleData{clusterRoleDataForSubjectAccessReviews},
		roles:        []RoleData{roleForWebhook},
		deployment: appsv1.DeploymentSpec{
			Replicas: &oneReplica,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "cert-manager",
							Args: []string{}, // container args are generated
							Env: []corev1.EnvVar{
								{
									Name: "POD_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
							Image:           "quay.io/jetstack/cert-manager-webhook:v1.3.1",
							ImagePullPolicy: "IfNotPresent",
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/livez",
										Port:   intstr.IntOrString{IntVal: 6080},
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 60,
								PeriodSeconds:       10,
								SuccessThreshold:    1,
								TimeoutSeconds:      1,
								FailureThreshold:    3,
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 10250,
									Name:          "https",
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.IntOrString{IntVal: 6080},
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       5,
								SuccessThreshold:    1,
								TimeoutSeconds:      1,
								FailureThreshold:    3,
							},
						},
					},
				},
			},
		},
		service: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       443,
					TargetPort: intstr.FromInt(10250),
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
		webhooks: []WebhookData{
			{
				name: "cert-manager-webhook",
				annotations: map[string]string{
					"cert-manager.io/inject-ca-from-secret": "cert-manager/cert-manager-webhook-ca",
				},
				mutatingWebhooks: []adregv1.MutatingWebhook{
					{
						Name: "webhook.cert-manager.io",
						AdmissionReviewVersions: []string{
							"v1",
							"v1beta1",
						},
						ClientConfig: adregv1.WebhookClientConfig{
							Service: &adregv1.ServiceReference{
								Name: "cert-manager-webhook", // pull this from other resources?
								// Namespace is pulled from CR
								Path: cmdoputils.GetStringPointer("/mutate"),
							},
						},
						FailurePolicy:  &failPolicy,
						SideEffects:    &noneSideEffect,
						TimeoutSeconds: &tenSeconds,
						Rules: []adregv1.RuleWithOperations{
							{
								Operations: []adregv1.OperationType{adregv1.Create, adregv1.Update},
								Rule: adregv1.Rule{
									Resources:   []string{"*/*"},
									APIGroups:   []string{"cert-manager.io", "acme.certmanager.io"},
									APIVersions: []string{"*"},
								},
							},
						},
					},
				},
				validatingWebhooks: []adregv1.ValidatingWebhook{
					{
						Name: "webhook.cert-manager.io",
						AdmissionReviewVersions: []string{
							"v1",
							"v1beta1",
						},
						ClientConfig: adregv1.WebhookClientConfig{
							Service: &adregv1.ServiceReference{
								Name: "cert-manager-webhook", // pull this from other resources?
								// namespace is set in a context where the CR exists.
								Path: cmdoputils.GetStringPointer("/validate"),
							},
						},
						FailurePolicy:  &failPolicy,
						SideEffects:    &noneSideEffect,
						TimeoutSeconds: &tenSeconds,
						Rules: []adregv1.RuleWithOperations{
							{
								Operations: []adregv1.OperationType{adregv1.Create, adregv1.Update},
								Rule: adregv1.Rule{
									APIGroups:   []string{"cert-manager.io", "acme.certmanager.io"},
									APIVersions: []string{"*"},
									Resources:   []string{"*/*"},
								},
							},
						},
						NamespaceSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "cert-manager.io/disable-validation",
									Operator: metav1.LabelSelectorOpNotIn,
									Values:   []string{"true"},
								},
								{
									Key:      "name",
									Operator: metav1.LabelSelectorOpNotIn,
									Values:   []string{"cert-manager"},
								},
							},
						},
					},
				},
			},
		},
	}

	switch version {
	case "v1.3.0":
		{
			container := &comp.deployment.Template.Spec.Containers[0]
			container.Image = fmt.Sprintf("quay.io/jetstack/cert-manager-webhook:%s", version)
		}
	case "v1.2.0":
		{
			container := &comp.deployment.Template.Spec.Containers[0]
			container.Image = fmt.Sprintf("quay.io/jetstack/cert-manager-webhook:%s", version)
		}
	}

	return comp
}
