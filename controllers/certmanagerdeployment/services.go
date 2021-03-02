package certmanagerdeployment

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/imdario/mergo"
	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
	"github.com/komish/cmd-operator-dev/cmdoputils"
	"github.com/komish/cmd-operator-dev/controllers/componentry"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileServices will reconcile the Service resources for a given CertManagerDeployment resource.
func (r *CertManagerDeploymentReconciler) reconcileServices(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: services")
	defer reqLogger.Info("Ending reconciliation: services")

	getter := ResourceGetter{CustomResource: *instance}
	svcs := getter.GetServices()

	// set controller reference on those objects
	for _, svc := range svcs {
		if err := controllerutil.SetControllerReference(instance, svc, r.Scheme); err != nil {
			return err
		}
		found := &corev1.Service{}
		err := r.Get(context.TODO(), types.NamespacedName{Namespace: svc.GetNamespace(), Name: svc.GetName()}, found)
		if err != nil && apierrors.IsNotFound(err) {
			reqLogger.Info("Creating Service", "Service.Namespace", svc.GetNamespace(), "Service.Name", svc.GetName())
			r.Eventf(instance,
				createManagedService.etype,
				createManagedService.reason,
				"%s: %s/%s",
				createManagedService.message,
				svc.GetNamespace(), svc.GetName())
			if err := r.Create(context.TODO(), svc); err != nil {
				return err
			}

			// successful create. Move onto next iteration.
			continue
		} else if err != nil {
			return err
		}

		// the service exists. If it needs updating, update it.
		genSpecInterface, err := cmdoputils.Interfacer{Data: svc.Spec}.ToJSONInterface()
		if err != nil { // err indicates a marshaling problem
			return err
		}
		foundSpecInterface, err := cmdoputils.Interfacer{Data: found.Spec}.ToJSONInterface()
		if err != nil {
			return err
		}

		genLabelsInterface, err := cmdoputils.Interfacer{Data: svc.Labels}.ToJSONInterface()
		if err != nil {
			return err
		}
		foundLabelsInterface, err := cmdoputils.Interfacer{Data: found.Labels}.ToJSONInterface()
		if err != nil {
			return err
		}

		genAnnotsInterface, err := cmdoputils.Interfacer{Data: svc.Annotations}.ToJSONInterface()
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
			reqLogger.Info("Service already exists, but needs an update.",
				"Service.Name", svc.GetName(),
				"HasExpectedLabels", labelsMatch,
				"HasExpectedAnnotation", annotsMatch,
				"HasExpectedSpec", specsMatch)
			r.Eventf(instance, updatingManagedService.etype, updatingManagedService.reason, "%s: %s/%s", updatingManagedService.message, svc.GetNamespace(), svc.GetName())

			// modify the state of the old object to post to API
			updated := found.DeepCopy()

			if !specsMatch {
				// update our local copy with values to keys as defined in our generated spec.
				err := mergo.Merge(&updated.Spec, svc.Spec, mergo.WithOverride)
				if err != nil {
					// Some problem merging the specs
					return err
				}
			}

			if !labelsMatch {
				// TODO(): should we avoid clobbering and instead just add our labels?
				updated.ObjectMeta.Labels = svc.GetLabels()
			}

			if !annotsMatch {
				// TODO(): should we avoid clobbering and instead just add our annotations?
				updated.ObjectMeta.Annotations = svc.GetAnnotations()
			}

			reqLogger.Info("Updating Service.", "Service.Name", svc.GetName(), "Service.Namespace", svc.GetNamespace())
			if err := r.Update(context.TODO(), updated); err != nil {
				return err
			}

			r.Eventf(instance, updatedManagedService.etype, updatedManagedService.reason, "%s: %s/%s", updatedManagedService.message, svc.GetNamespace(), svc.GetName())
		}
	}
	return nil
}

// GetServices will return new services for the CR.
func (r *ResourceGetter) GetServices() []*corev1.Service {
	var svcs []*corev1.Service
	for _, componentGetterFunc := range componentry.Components {
		component := componentGetterFunc(cmdoputils.CRVersionOrDefaultVersion(
			r.CustomResource.Spec.Version,
			componentry.CertManagerDefaultVersion))
		// Not all components have services. If a component has an uninitialized
		// corev1.ServiceSpec, then we skip it here.
		if !reflect.DeepEqual(component.GetService(), corev1.ServiceSpec{}) {
			svcs = append(svcs, newService(component, r.CustomResource))
		}
	}

	return svcs
}

// newService returns a service object for a custom resource.
func newService(comp componentry.CertManagerComponent, cr operatorsv1alpha1.CertManagerDeployment) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      comp.GetResourceName(),
			Namespace: componentry.CertManagerDeploymentNamespace,
			Labels:    cmdoputils.MergeMaps(componentry.StandardLabels, comp.GetLabels()),
		},
		Spec: comp.GetService(),
	}

	// add the label selectors for the base deployment
	sel := comp.GetBaseLabelSelector()
	sel = metav1.AddLabelToSelector(sel, componentry.InstanceLabelKey, cr.Name)
	svc.Spec.Selector, _ = metav1.LabelSelectorAsMap(sel)

	return svc
}
