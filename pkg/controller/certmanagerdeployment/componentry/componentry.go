// Package componentry contains metadata and types relevant to the various cert-manager components.
package componentry

import (
	"strings"

	"github.com/komish/certmanager-operator/pkg/controller/certmanagerdeployment/cmdoputils"
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
		"app": "cert-manager",
	}

	// Components are all ComponentGetterFunctions, one per Component, that we need
	// to deploy and manage as a part of a CertManagerDeployment.
	Components = []ComponentGetterFunction{GetComponentForController, GetComponentForCAInjector, GetComponentForWebhook}
)

// ComponentGetterFunction is a function that will return a base CertManagerComponent.
type ComponentGetterFunction func() CertManagerComponent

// CertManagerComponent represents the various components of cert-manager that might
// be installed in a cluster.
type CertManagerComponent struct {
	name               string
	serviceAccountName string
	labels             map[string]string
	clusterRoles       []RoleData
	roles              []RoleData
	deployment         appsv1.DeploymentSpec
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

// GetComponentForController returns a CetManagerComponent containing
// all the metadata necessary to deploy the subresources needed to run
// the cert-manager controller.
func GetComponentForController() CertManagerComponent {
	return CertManagerComponent{
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
							Image:           "quay.io/jetstack/cert-manager-controller:v0.15.0",
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
	}
}

// GetComponentForCAInjector returns a CetManagerComponent containing
// all the metadata necessary to deploy the subresources needed to run
// the cert-manager cainjector.
func GetComponentForCAInjector() CertManagerComponent {
	return CertManagerComponent{
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
							Image:           "quay.io/jetstack/cert-manager-cainjector:v0.15.0",
							ImagePullPolicy: "IfNotPresent",
						},
					},
				},
			},
		},
	}
}

// GetComponentForWebhook returns a CetManagerComponent containing
// all the metadata necessary to deploy the subresources needed to run
// the cert-manager webhook.
func GetComponentForWebhook() CertManagerComponent {
	return CertManagerComponent{
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
							Image:           "quay.io/jetstack/cert-manager-webhook:v0.15.0",
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
							},
						},
					},
				},
			},
		},
	}
}
