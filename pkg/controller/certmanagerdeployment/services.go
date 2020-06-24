package certmanagerdeployment

import (
	"reflect"

	redhatv1alpha1 "github.com/komish/certmanager-operator/pkg/apis/redhat/v1alpha1"
	"github.com/komish/certmanager-operator/pkg/controller/certmanagerdeployment/cmdoputils"
	"github.com/komish/certmanager-operator/pkg/controller/certmanagerdeployment/componentry"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/util/labels"
)

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
func newService(comp componentry.CertManagerComponent, cr redhatv1alpha1.CertManagerDeployment) *corev1.Service {
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
	sel = labels.AddLabelToSelector(sel, "app.kubernetes.io/instance", cr.Name)
	// TODO(komish): Probably need to handle this blank-assigned error at some point.
	// Not handled currently because errors are not bubbling up from this function yet.
	svc.Spec.Selector, _ = metav1.LabelSelectorAsMap(sel)

	return svc
}
