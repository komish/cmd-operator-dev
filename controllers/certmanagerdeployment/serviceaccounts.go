package certmanagerdeployment

import (
	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
	"github.com/komish/cmd-operator-dev/cmdoputils"
	"github.com/komish/cmd-operator-dev/controllers/componentry"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetServiceAccounts will return new service account objects for the CR.
func (r *ResourceGetter) GetServiceAccounts() []*corev1.ServiceAccount {
	var sas []*corev1.ServiceAccount
	for _, componentGetterFunc := range componentry.Components {
		component := componentGetterFunc(cmdoputils.CRVersionOrDefaultVersion(
			r.CustomResource.Spec.Version,
			componentry.CertManagerDefaultVersion))
		sa := newServiceAccount(component, r.CustomResource)
		sas = append(sas, sa)
	}
	return sas
}

// newServiceAccount returns a service account object for a custom resource.
// Service accounts are namespaced resources so the global installation namespace
// is used here.
func newServiceAccount(comp componentry.CertManagerComponent, cr operatorsv1alpha1.CertManagerDeployment) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      comp.GetServiceAccountName(),
			Namespace: componentry.CertManagerDeploymentNamespace,
			Labels:    comp.GetLabels(),
		},
	}
}
