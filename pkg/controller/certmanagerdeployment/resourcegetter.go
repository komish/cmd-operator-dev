package certmanagerdeployment

import (
	redhatv1alpha1 "github.com/komish/certmanager-operator/pkg/apis/redhat/v1alpha1"
)

// ResourceGetter facilitates getting various owned resources expected by
// a CertManagerDeployment CR.
type ResourceGetter struct {
	CustomResource redhatv1alpha1.CertManagerDeployment
}

//GetCustomResourceDefinitions will return new custom resource definitions for the CR.
func (r *ResourceGetter) GetCustomResourceDefinitions() {
	// TODO(): Implement Me!
	return
}

// GetWebhookConfigurations will return new webhooks for the CR.
func (r *ResourceGetter) GetWebhookConfigurations() {
	// TODO(): Implement Me!
	return
}
