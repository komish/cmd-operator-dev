package certmanagerdeployment

import (
	"reflect"

	redhatv1alpha1 "github.com/komish/certmanager-operator/pkg/apis/redhat/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/kubernetes/pkg/util/labels"
)

// DeploymentCustomizations are the values from the CustomResource that will
// impact the deployment for a given CertManagerComponent.
type DeploymentCustomizations struct {
	// ContainerImage is a container image to be used for a component
	// in the format /registry/container-image:tag
	ContainerImage  string
	ImagePullPolicy corev1.PullPolicy
}

var (
	// oneReplica is a value of 1 of type int32 to be used
	// as the appsv1.DeploymentSpec.Replicas struct key which
	// requires a typed int32 pointer.
	oneReplica = int32(1)

	// DeploySpecForController is a deployment base template for
	// the Controller CertManagerComponent
	deploySpecForController = appsv1.DeploymentSpec{
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
	}

	// DeploySpecForCAInjector is a deployment base template for
	// the CAInjector CertManagerComponent
	deploySpecForCAInjector = appsv1.DeploymentSpec{
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
	}

	// DeploySpecForWebhook is a deployment base template for
	// the Webhook CertManagerComponent
	deploySpecForWebhook = appsv1.DeploymentSpec{
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
	}
)

// GetDeployments returns Deployment objects for a given CertManagerDeployment
// custom resource.
func (r *ResourceGetter) GetDeployments() []*appsv1.Deployment {
	return []*appsv1.Deployment{
		newDeployment(CAInjector, r.CustomResource, r.GetDeploymentCustomizations(CAInjector)),
		newDeployment(Controller, r.CustomResource, r.GetDeploymentCustomizations(Controller)),
		newDeployment(Webhook, r.CustomResource, r.GetDeploymentCustomizations(Webhook)),
	}
}

// GetDeploymentCustomizations will return a DeploymentCustomization object for a given
// CertManagerComponent. This helps derive the resulting DeploymentSpec for the component.
func (r *ResourceGetter) GetDeploymentCustomizations(comp CertManagerComponent) DeploymentCustomizations {
	dc := DeploymentCustomizations{}

	// Check if the image has been overridden
	imageOverrides := r.CustomResource.Spec.DangerZone.ImageOverrides
	if !reflect.DeepEqual(imageOverrides, map[string]string{}) {
		// imageOverrides is not empty, get the image value for this component.
		dc.ContainerImage = imageOverrides[comp.Name]
	}

	// check if pull policy has been overridden
	pullPolicyOverride := r.CustomResource.Spec.DangerZone.ImagePullPolicy
	var emptyPullPolicy corev1.PullPolicy
	if !reflect.DeepEqual(pullPolicyOverride, emptyPullPolicy) {
		dc.ImagePullPolicy = pullPolicyOverride
	}

	return dc
}

// newDeployment returns a Deployment object for a given CertManagerComponent
// and CertManagerDeployment CustomResource
func newDeployment(comp CertManagerComponent, cr redhatv1alpha1.CertManagerDeployment, cstm DeploymentCustomizations) *appsv1.Deployment {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      comp.GetResourceName(),
			Namespace: CertManagerDeploymentNamespace,
			Labels:    mergeMaps(standardLabels, comp.Labels),
		},
		Spec: comp.Deployment,
	}

	// Add the service account entry to the base deployment
	setServiceAccount(deploy, comp.ServiceAccountName)

	// add the label selectors for the base deployment
	sel := comp.GetBaseLabelSelector()
	sel = labels.AddLabelToSelector(sel, "app.kubernetes.io/instance", cr.Name)
	deploy.Spec.Selector = sel

	// TODO(): Should probably handle the below blank-assigned error in some way.
	selmap, _ := metav1.LabelSelectorAsMap(sel)
	deploy.Spec.Template.ObjectMeta.Labels = selmap

	// If the CR contains a customized container image for the component, override our deployment
	if cstm.ContainerImage != "" {
		// TODO(): I'm assuming a single container image per deployment for the components because
		// That is what's true today. If this changes, this will need to be updated.
		deploy.Spec.Template.Spec.Containers[0].Image = cstm.ContainerImage
	}

	// If the CR contains a customized image pull policy: override our deployment
	if cstm.ImagePullPolicy != "" {
		deploy.Spec.Template.Spec.Containers[0].ImagePullPolicy = cstm.ImagePullPolicy
	}

	return deploy
}

// setServiceAccount adds the service account to a given DeploymentSpec Object, or
// overwrites the value for the service acccount if one already exists.
func setServiceAccount(deploy *appsv1.Deployment, sa string) *appsv1.Deployment {
	deploy.Spec.Template.Spec.ServiceAccountName = sa
	return deploy
}

// setContainerImage will update a container object's image.
func setContainerImage(container *corev1.Container, image string) *corev1.Container {
	container.Image = image
	return container
}

// setContainerImage will update a container object's image.
func setImagePullPolicy(container *corev1.Container, policy corev1.PullPolicy) *corev1.Container {
	container.ImagePullPolicy = policy
	return container
}
