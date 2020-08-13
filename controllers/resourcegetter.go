package controllers

import operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"

// ResourceGetter facilitates getting various owned resources expected by
// a CertManagerDeployment CR.
type ResourceGetter struct {
	CustomResource operatorsv1alpha1.CertManagerDeployment
}
