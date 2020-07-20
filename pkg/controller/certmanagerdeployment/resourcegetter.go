package certmanagerdeployment

import (
	redhatv1alpha1 "github.com/komish/certmanager-operator/pkg/apis/redhat/v1alpha1"
)

// ResourceGetter facilitates getting various owned resources expected by
// a CertManagerDeployment CR.
type ResourceGetter struct {
	CustomResource redhatv1alpha1.CertManagerDeployment
}
