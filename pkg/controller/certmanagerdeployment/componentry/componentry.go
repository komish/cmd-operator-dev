// Package componentry contains metadata and types relevant to the various cert-manager components.
package componentry

import (
	"strings"

	"github.com/komish/certmanager-operator/pkg/controller/certmanagerdeployment/cmdoputils"
	adregv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/kubernetes/pkg/util/labels"
)

var (
	// oneReplica is a value of 1 of type int32 to be used
	// as the appsv1.DeploymentSpec.Replicas struct key which
	// requires a typed int32 pointer.
	oneReplica = int32(1)

	// StandardLabels are the base labels that apply to all CertManagerDeployment-managed resources.
	StandardLabels = map[string]string{
		"app":                          "cert-manager",
		"app.kubernetes.io/managed-by": "operator",
	}

	// SupportedVersions represents the versions of Cert-Manager that are supported by the operator.
	// The value is irrelevant. Only the keys are used for lookup.
	//
	// Changes to these values also need to be mirrored in the CertManageDeploymentSpec generation
	// validation annotations.
	SupportedVersions = map[string]bool{
		"v0.14.3": true,
		"v0.15.0": true,
		"v0.15.1": true,
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
	labels.AddLabelToSelector(&ls, "app.kubernetes.io/component", comp.name)
	labels.AddLabelToSelector(&ls, "app.kubernetes.io/name", comp.name)
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
							Args: []string{
								"--v=2",
								"--cluster-resource-namespace=$(POD_NAMESPACE)",
								"--leader-election-namespace=$(POD_NAMESPACE)",
							},
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
							Image:           "quay.io/jetstack/cert-manager-controller:v0.15.1",
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
	case "v0.15.0":
		{
			container := &comp.deployment.Template.Spec.Containers[0] // we assume one container
			container.Image = "quay.io/jetstack/cert-manager-controller:v0.15.0"
		}
	case "v0.14.3":
		{
			container := &comp.deployment.Template.Spec.Containers[0] // we assume one container
			container.Image = "quay.io/jetstack/cert-manager-controller:v0.14.3"
			container.Args = []string{
				"--v=2",
				"--cluster-resource-namespace=$(POD_NAMESPACE)",
				"--leader-election-namespace=$(POD_NAMESPACE)",
				"--webhook-namespace=$(POD_NAMESPACE)",
				"--webhook-ca-secret=cert-manager-webhook-ca",
				"--webhook-serving-secret=cert-manager-webhook-tls",
				"--webhook-dns-names=cert-manager-webhook,cert-manager-webhook.cert-manager,cert-manager-webhook.cert-manager.svc",
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
							Args: []string{
								"--v=2",
								"--leader-election-namespace=$(POD_NAMESPACE)",
							},
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
							Image:           "quay.io/jetstack/cert-manager-cainjector:v0.15.1",
							ImagePullPolicy: "IfNotPresent",
						},
					},
				},
			},
		},
		webhooks: []WebhookData{},
	}

	switch version {
	case "v0.15.0":
		{
			container := &comp.deployment.Template.Spec.Containers[0]
			container.Image = "quay.io/jetstack/cert-manager-cainjector:v0.15.0"
		}
	case "v0.14.3":
		container := &comp.deployment.Template.Spec.Containers[0]
		container.Image = "quay.io/jetstack/cert-manager-cainjector:v0.14.3"
	}
	return comp
}

// GetComponentForWebhook returns a CertManagerComponent containing
// all the metadata necessary to deploy the subresources needed to run
// the cert-manager webhook.
func GetComponentForWebhook(version string) CertManagerComponent {
	// This is for use with webhook objects.
	var failPolicy adregv1beta1.FailurePolicyType = "Fail"
	var noneSideEffect adregv1beta1.SideEffectClass = "None"

	comp := CertManagerComponent{
		name:               "webhook",
		serviceAccountName: "cert-manager-webhook",
		labels: cmdoputils.MergeMaps(map[string]string{
			"app.kubernetes.io/component": "webhook",
			"app.kubernetes.io/name":      "webhook",
		}, StandardLabels),
		clusterRoles: []RoleData{},
		roles:        []RoleData{roleForWebhook},
		deployment: appsv1.DeploymentSpec{
			Replicas: &oneReplica,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "cert-manager",
							Args: []string{
								"--v=2",
								"--secure-port=10250",
								"--dynamic-serving-ca-secret-namespace=$(POD_NAMESPACE)",
								"--dynamic-serving-ca-secret-name=cert-manager-webhook-ca",
								"--dynamic-serving-dns-names=cert-manager-webhook,cert-manager-webhook.cert-manager,cert-manager-webhook.cert-manager.svc",
							},
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
							Image:           "quay.io/jetstack/cert-manager-webhook:v0.15.1",
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
				mutatingWebhooks: []adregv1beta1.MutatingWebhook{
					{
						Name: "webhook.cert-manager.io",
						ClientConfig: adregv1beta1.WebhookClientConfig{
							Service: &adregv1beta1.ServiceReference{
								Name: "cert-manager-webhook", // pull this from other resources?
								// Namespace is pulled from CR
								Path: cmdoputils.GetStringPointer("/mutate"),
							},
						},
						FailurePolicy: &failPolicy,
						SideEffects:   &noneSideEffect,
						Rules: []adregv1beta1.RuleWithOperations{
							{
								Operations: []adregv1beta1.OperationType{adregv1beta1.Create, adregv1beta1.Update},
								Rule: adregv1beta1.Rule{
									Resources:   []string{"*/*"},
									APIGroups:   []string{"cert-manager.io", "acme.certmanager.io"},
									APIVersions: []string{"v1alpha2", "v1alpha3"},
								},
							},
						},
					},
				},
				validatingWebhooks: []adregv1beta1.ValidatingWebhook{
					{
						Name: "webhook.cert-manager.io",
						ClientConfig: adregv1beta1.WebhookClientConfig{
							Service: &adregv1beta1.ServiceReference{
								Name: "cert-manager-webhook", // pull this from other resources?
								// namespace is set in a context where the CR exists.
								Path: cmdoputils.GetStringPointer("/validate"),
							},
						},
						FailurePolicy: &failPolicy,
						SideEffects:   &noneSideEffect,
						Rules: []adregv1beta1.RuleWithOperations{
							{
								Operations: []adregv1beta1.OperationType{adregv1beta1.Create, adregv1beta1.Update},
								Rule: adregv1beta1.Rule{
									APIGroups:   []string{"cert-manager.io", "acme.certmanager.io"},
									APIVersions: []string{"v1alpha2", "v1alpha3"},
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
	case "v0.15.0":
		{
			container := &comp.deployment.Template.Spec.Containers[0]
			container.Image = "quay.io/jetstack/cert-manager-webhook:v0.15.0"
		}
	case "v0.14.3":
		// make modifications to the current template for v0.14.3.
		container := &comp.deployment.Template.Spec.Containers[0]
		// new image
		container.Image = "quay.io/jetstack/cert-manager-webhook:v0.14.3"
		// new arguments
		container.Args = []string{
			"--v=2",
			"--secure-port=10250",
			"--tls-cert-file=/certs/tls.crt",
			"--tls-private-key-file=/certs/tls.key",
		}
		// livenessProbes remove the delay and period
		container.LivenessProbe = &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/livez",
					Port:   intstr.IntOrString{IntVal: 6080},
					Scheme: corev1.URISchemeHTTP,
				},
			},
		}
		// readinessProbes remove the delay and period
		container.ReadinessProbe = &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/healthz",
					Port:   intstr.IntOrString{IntVal: 6080},
					Scheme: corev1.URISchemeHTTP,
				},
			},
		}
		// this version requires a volume mount
		container.VolumeMounts = []corev1.VolumeMount{
			{
				Name:      "certs",
				MountPath: "/certs",
			},
		}

		comp.deployment.Template.Spec.Volumes = []corev1.Volume{
			{
				Name: "certs",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: "cert-manager-webhook-tls",
					},
				},
			},
		}

		// this version's webhook has a different annotation than default supported
		for _, hook := range comp.webhooks {
			hook.annotations["cert-manager.io/inject-ca-from-secret"] = "cert-manager/cert-manager-webhook-tls"
		}
	}

	return comp
}
