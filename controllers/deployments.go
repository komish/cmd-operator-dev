package controllers

import (
	"encoding/json"
	"fmt"

	"reflect"
	"strings"

	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
	"github.com/komish/cmd-operator-dev/cmdoputils"
	"github.com/komish/cmd-operator-dev/controllers/componentry"
	certmanagerconfigsv1 "github.com/komish/cmd-operator-dev/controllers/configs/v1"
	"github.com/openshift/library-go/pkg/operator/resource/resourcemerge"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeploymentCustomizations are the values from the CustomResource that will
// impact the deployment for a given CertManagerComponent.
type DeploymentCustomizations struct {
	// ContainerImage is a container image to be used for a component
	// in the format /registry/container-image:tag
	ContainerImage string
	ContainerArgs  runtime.RawExtension
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

	// check if the any container arguments are being overridden.
	if argOverrides := r.CustomResource.Spec.DangerZone.ContainerArgOverrides.GetOverridesFor(comp.GetName()); argOverrides.Raw != nil {
		dc.ContainerArgs = *argOverrides
	} else {
		// if argOverrides.Raw is nil, that implies the user did not set the override for this component.
		// If we pass a nil value to this, we end up setting our arguments to null which sets the container
		// args to null, causing no args to get set (not even defaults).
		// Instead, set it to an empty byte slice.
		dc.ContainerArgs = runtime.RawExtension{Raw: []byte{}}
	}

	return dc
}

// newDeployment returns a Deployment object for a given CertManagerComponent
// and CertManagerDeployment CustomResource
func newDeployment(comp componentry.CertManagerComponent, cr operatorsv1alpha1.CertManagerDeployment, cstm DeploymentCustomizations) *appsv1.Deployment {
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
	sel = metav1.AddLabelToSelector(sel, componentry.InstanceLabelKey, cr.Name)
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

	// we don't have any custom merge rules to consider
	specialMergeRules := map[string]resourcemerge.MergeFunc{}
	// we have to lay out the flag overriding to be in the right format, we don't expect the
	// user to add the flags key
	f := overrideConfig{Flags: cstm.ContainerArgs}
	userDefinedArgs, _ := json.Marshal(f)

	// TODO: handling this error requires some refactor, but we probably need to do it.
	result, err := resourcemerge.MergePrunedProcessConfig(
		certmanagerconfigsv1.GetEmptyConfigFor(comp.GetName()), // the schema
		specialMergeRules, // we have no merge rules
		certmanagerconfigsv1.DefaultConfigsFor[comp.GetName()], // our default
		userDefinedArgs, // user overridden flags
	)

	if err != nil {
		// we don't log this at the moment, but we should
		// for now we just run a default configuration
		result = certmanagerconfigsv1.DefaultConfigsFor[comp.GetName()]
	}

	deploy.Spec.Template.Spec.Containers[0].Args = argSliceOf(result, certmanagerconfigsv1.GetEmptyConfigFor(comp.GetName()))

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

// argSliceOf returns a string slice of arguments, built from a CertManager*Config object.
func argSliceOf(data []byte, schema runtime.Object) []string {
	// TODO handle errors in this function
	var args []string

	// get the object as a map so we can pull the flag data
	var objectMap map[string]interface{}
	json.Unmarshal(data, &objectMap) // unhandled error

	f, _ := json.Marshal(objectMap["flags"]) //unhandled error
	// get the flags as a map
	var flagMap map[string]interface{}
	json.Unmarshal(f, &flagMap) //unhandled error

	// TODO: consider enforcing a more-correct argument structure
	// by allowing the config types to reference keys as custom types
	// and then doing some type assertions against map[string]interface{}
	// when they get marshaled as JSON objects.
	for k, v := range flagMap {
		switch z := v.(type) {
		case string, bool, float64:
			args = append(args, fmt.Sprintf("--%s=%v", k, z))
		case []interface{}:
			// marshaling turns json arrays into []interface, but the config types only accept []strings
			s := make([]string, len(z))
			for i, v := range z {
				s[i] = fmt.Sprintf("%s", v)
			}
			val := strings.Join(s, ",")
			args = append(args, fmt.Sprintf("--%s=%v", k, val))
		default:
			// TODO implement some kind of logger here
		}
	}

	return args
}

type overrideConfig struct {
	Flags runtime.RawExtension `json:"flags"`
}
