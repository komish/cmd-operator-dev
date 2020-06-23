package certmanagerdeployment

import (
	redhatv1alpha1 "github.com/komish/certmanager-operator/pkg/apis/redhat/v1alpha1"
	"github.com/komish/certmanager-operator/pkg/controller/certmanagerdeployment/componentry"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetServiceAccounts will return new service account objects for the CR.
func (r *ResourceGetter) GetServiceAccounts() []*corev1.ServiceAccount {
	var sas []*corev1.ServiceAccount
	for _, componentGetterFunc := range componentry.Components {
		component := componentGetterFunc(*r.CustomResource.Spec.Version)
		sa := newServiceAccount(component, r.CustomResource)
		sas = append(sas, sa)
	}
	return sas
}

// newServiceAccount returns a service account object for a custom resource.
// Service accounts are namespaced resources so the global installation namespace
// is used here.
func newServiceAccount(comp componentry.CertManagerComponent, cr redhatv1alpha1.CertManagerDeployment) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      comp.GetServiceAccountName(),
			Namespace: componentry.CertManagerDeploymentNamespace,
			Labels:    comp.GetLabels(),
		},
	}
}
