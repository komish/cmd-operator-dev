package certmanagerdeployment

import (
	"reflect"

	redhatv1alpha1 "github.com/komish/certmanager-operator/pkg/apis/redhat/v1alpha1"
	"github.com/komish/certmanager-operator/pkg/controller/certmanagerdeployment/cmdoputils"
	"github.com/komish/certmanager-operator/pkg/controller/certmanagerdeployment/componentry"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/util/labels"
)

// DeploymentCustomizations are the values from the CustomResource that will
// impact the deployment for a given CertManagerComponent.
type DeploymentCustomizations struct {
	// ContainerImage is a container image to be used for a component
	// in the format /registry/container-image:tag
	ContainerImage  string
	ImagePullPolicy corev1.PullPolicy
	ContainerArgs   *[]string
}

// GetDeployments returns Deployment objects for a given CertManagerDeployment
// custom resource.
func (r *ResourceGetter) GetDeployments() []*appsv1.Deployment {
	var deploys []*appsv1.Deployment
	for _, componentGetterFunc := range componentry.Components {
		component := componentGetterFunc(
			cmdoputils.CRVersionOrDefaultVersion(
				r.CustomResource.Spec.Version,
				componentry.CertManagerDefaultVersion),
		)
		deploys = append(deploys, newDeployment(component, r.CustomResource, r.GetDeploymentCustomizations(component)))
	}

	return deploys
}

// GetDeploymentCustomizations will return a DeploymentCustomization object for a given
// CertManagerComponent. This helps derive the resulting DeploymentSpec for the component.
func (r *ResourceGetter) GetDeploymentCustomizations(comp componentry.CertManagerComponent) DeploymentCustomizations {
	dc := DeploymentCustomizations{}

	// Check if the image has been overridden
	imageOverrides := r.CustomResource.Spec.DangerZone.ImageOverrides
	if !reflect.DeepEqual(imageOverrides, map[string]string{}) {
		// imageOverrides is not empty, get the image value for this component.
		dc.ContainerImage = imageOverrides[comp.GetName()]
	}

	// check if pull policy has been overridden.
	pullPolicyOverride := r.CustomResource.Spec.ImagePullPolicy
	var emptyPullPolicy corev1.PullPolicy
	if !reflect.DeepEqual(pullPolicyOverride, emptyPullPolicy) {
		dc.ImagePullPolicy = pullPolicyOverride
	}

	// check if the container arguments are being overridden.
	argOverrides := r.CustomResource.Spec.DangerZone.ContainerArgOverrides
	if !reflect.DeepEqual(imageOverrides, map[string][]string{}) {
		args := argOverrides[comp.GetName()]
		dc.ContainerArgs = &args
	}

	return dc
}

// newDeployment returns a Deployment object for a given CertManagerComponent
// and CertManagerDeployment CustomResource
func newDeployment(comp componentry.CertManagerComponent, cr redhatv1alpha1.CertManagerDeployment, cstm DeploymentCustomizations) *appsv1.Deployment {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      comp.GetResourceName(),
			Namespace: componentry.CertManagerDeploymentNamespace,
			Labels:    cmdoputils.MergeMaps(componentry.StandardLabels, comp.GetLabels()),
		},
		Spec: comp.GetDeployment(),
	}

	// Add the service account entry to the base deployment
	setServiceAccount(deploy, comp.GetServiceAccountName())

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

	// If the CR containers customized arguments: override our deployment.
	// if cstm.ContainerArgs != nil {
	// 	deploy.Spec.Template.Spec.Containers[0].Args = *cstm.ContainerArgs
	// }

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
