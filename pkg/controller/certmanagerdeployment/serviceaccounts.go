package certmanagerdeployment

import (
	redhatv1alpha1 "github.com/komish/certmanager-operator/pkg/apis/redhat/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetServiceAccounts will return new service account objects for the CR.
func (r *ResourceGetter) GetServiceAccounts() []*corev1.ServiceAccount {
	var sas []*corev1.ServiceAccount
	for _, component := range Components {
		sa := newServiceAccount(component, r.CustomResource)
		sas = append(sas, sa)
	}
	return sas
}

// newServiceAccount returns a service account object for a custom resource.
// Service accounts are namespaced resources so the global installation namespace
// is used here.
func newServiceAccount(comp CertManagerComponent, cr redhatv1alpha1.CertManagerDeployment) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      comp.ServiceAccountName,
			Namespace: CertManagerDeploymentNamespace,
			Labels:    comp.Labels,
		},
	}
}
