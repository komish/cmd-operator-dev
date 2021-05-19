package certmanagerdeployment

import (
	"context"
	"encoding/json"
	"fmt"

	"reflect"
	"strings"

	"github.com/go-logr/logr"
	"github.com/imdario/mergo"
	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
	"github.com/komish/cmd-operator-dev/cmdoputils"
	"github.com/komish/cmd-operator-dev/controllers/componentry"
	certmanagerconfigs "github.com/komish/cmd-operator-dev/controllers/configs"
	"github.com/openshift/library-go/pkg/operator/resource/resourcemerge"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// reconcileDeployments will reconcile the Deployment resources for a given CertManagerDeployment resource.
func (r *CertManagerDeploymentReconciler) reconcileDeployments(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: deployments")
	defer reqLogger.Info("Ending reconciliation: deployments")

	deps := GetDeploymentsFor(*instance)

	for _, dep := range deps {
		if err := controllerutil.SetControllerReference(instance, dep, r.Scheme); err != nil {
			return err
		}
		found := &appsv1.Deployment{}
		err := r.Get(context.TODO(), types.NamespacedName{Namespace: dep.GetNamespace(), Name: dep.GetName()}, found)
		if err != nil && apierrors.IsNotFound(err) {
			reqLogger.Info("Creating Deployment", "Deployment.Namespace", dep.GetNamespace(), "Deployment.Name", dep.GetName())
			r.Eventf(instance,
				createManagedDeployment.etype,
				createManagedDeployment.reason,
				"%s: %s/%s",
				createManagedDeployment.message,
				dep.GetNamespace(), dep.GetName())
			if err := r.Create(context.TODO(), dep); err != nil {
				return err
			}

			// successful create. Move onto next iteration.
			continue
		} else if err != nil {
			return err
		}

		// A deployment exists. Update if necessary.
		genSpecInterface, err := cmdoputils.Interfacer{Data: dep.Spec}.ToJSONInterface()
		if err != nil { // err indicates a marshaling problem
			return err
		}
		foundSpecInterface, err := cmdoputils.Interfacer{Data: found.Spec}.ToJSONInterface()
		if err != nil {
			return err
		}

		genLabelsInterface, err := cmdoputils.Interfacer{Data: dep.Labels}.ToJSONInterface()
		if err != nil {
			return err
		}
		foundLabelsInterface, err := cmdoputils.Interfacer{Data: found.Labels}.ToJSONInterface()
		if err != nil {
			return err
		}

		genAnnotsInterface, err := cmdoputils.Interfacer{Data: dep.Annotations}.ToJSONInterface()
		if err != nil {
			return err
		}
		foundAnnotsInterface, err := cmdoputils.Interfacer{Data: found.Annotations}.ToJSONInterface()
		if err != nil {
			return err
		}

		specsMatch := cmdoputils.ObjectsMatch(genSpecInterface, foundSpecInterface)
		labelsMatch := cmdoputils.ObjectsMatch(genLabelsInterface, foundLabelsInterface)
		annotsMatch := cmdoputils.ObjectsMatch(genAnnotsInterface, foundAnnotsInterface)

		if !(specsMatch && labelsMatch && annotsMatch) {
			reqLogger.Info("Deployment already exists, but needs an update.",
				"Deployment.Name", dep.GetName(),
				"Deployment.Namespace", dep.GetNamespace(),
				"HasExpectedLabels", labelsMatch,
				"HasExpectedAnnotation", annotsMatch,
				"HasExpectedSpec", specsMatch)
			r.Eventf(instance, updatingManagedDeployment.etype, updatingManagedDeployment.reason, "%s: %s/%s", updatingManagedDeployment.message, dep.GetNamespace(), dep.GetName()) // BOOKMARK

			updated := found.DeepCopy()

			if !specsMatch {
				// update our local copy with values to keys as defined in our generated spec.
				err := mergo.Merge(&updated.Spec, dep.Spec, mergo.WithOverride)
				if err != nil {
					// Some problem merging the specs
					return err
				}
			}

			if !labelsMatch {
				// TODO(): should we avoid clobbering and instead just add our labels?
				updated.ObjectMeta.Labels = dep.GetLabels()
			}

			if !annotsMatch {
				// TODO(): should we avoid clobbering and instead just add our annotations?
				updated.ObjectMeta.Annotations = dep.GetAnnotations()
			}

			reqLogger.Info("Updating Deployment.", "Deployment.Name", dep.GetName(), "Deployment.Namespace", dep.GetNamespace())
			if err := r.Update(context.TODO(), updated); err != nil {
				return err
			}

			r.Eventf(instance, updatedManagedDeployment.etype, updatedManagedDeployment.reason, "%s: %s/%s", updatedManagedDeployment.message, dep.GetNamespace(), dep.GetName())
		}
	}

	return nil
}

// DeploymentCustomizations are the values from the CertManagerDeployment that can
// impact the resulting managed deployments.
type DeploymentCustomizations struct {
	// ContainerImage is a container image to be used for a component
	// in the format /registry/container-image:tag
	ContainerImage string
	ContainerArgs  runtime.RawExtension
}

// GetDeployments returns Deployment objects for a given CertManagerDeployment resource.
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

// GetDeploymentsFor returns Deployment objects for a given CertManagerDeployment resource.
func GetDeploymentsFor(cr operatorsv1alpha1.CertManagerDeployment) []*appsv1.Deployment {
	var deploys []*appsv1.Deployment
	for _, componentGetterFunc := range componentry.Components {
		component := componentGetterFunc(
			cmdoputils.CRVersionOrDefaultVersion(
				cr.Spec.Version,
				componentry.CertManagerDefaultVersion),
		)
		deploys = append(deploys, newDeployment(component, cr, getDeploymentCustomizations(cr, component)))
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

// getDeploymentCustomizations will return a DeploymentCustomization object for a given
// CertManagerComponent. This helps derive the resulting DeploymentSpec for the component.
func getDeploymentCustomizations(cr operatorsv1alpha1.CertManagerDeployment, comp componentry.CertManagerComponent) DeploymentCustomizations {
	dc := DeploymentCustomizations{}

	// Check if the image has been overridden
	imageOverrides := cr.Spec.DangerZone.ImageOverrides
	if !reflect.DeepEqual(imageOverrides, map[string]string{}) {
		// imageOverrides is not empty, get the image value for this component.
		dc.ContainerImage = imageOverrides[comp.GetName()]
	}

	// check if the any container arguments are being overridden.
	if argOverrides := cr.Spec.DangerZone.ContainerArgOverrides.GetOverridesFor(comp.GetName()); argOverrides.Raw != nil {
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
	deploy.Spec.Template.Spec.ServiceAccountName = comp.GetServiceAccountName()

	// add the label selectors for the base deployment
	sel := comp.GetBaseLabelSelector()
	sel = metav1.AddLabelToSelector(sel, componentry.InstanceLabelKey, cr.Name)
	deploy.Spec.Selector = sel

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
		certmanagerconfigs.GetEmptyConfigFor(comp.GetName(), cmdoputils.CRVersionOrDefaultVersion(cr.Spec.Version, componentry.CertManagerDefaultVersion)), // the schema
		specialMergeRules, // we have no merge rules
		certmanagerconfigs.GetDefaultConfigFor(comp.GetName(), cmdoputils.CRVersionOrDefaultVersion(cr.Spec.Version, componentry.CertManagerDefaultVersion)), // our default
		userDefinedArgs, // user overridden flags
	)

	if err != nil {
		// run with a default configuratio nif there was an error merging configs
		result = certmanagerconfigs.GetDefaultConfigFor(comp.GetName(), cmdoputils.CRVersionOrDefaultVersion(cr.Spec.Version, componentry.CertManagerDefaultVersion))
	}

	deploy.Spec.Template.Spec.Containers[0].Args = argSliceOf(result, certmanagerconfigs.GetEmptyConfigFor(comp.GetName(), componentry.CertManagerDefaultVersion))

	return deploy
}

// argSliceOf returns a string slice of arguments, built from a CertManagerConfig object.
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
