package controllers

import (
	"github.com/komish/cmd-operator-dev/controllers/componentry"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetNamespace returns a namespace object for a given CertManagerDeployment
// custom resource.
func (r *ResourceGetter) GetNamespace() *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   componentry.CertManagerDeploymentNamespace,
			Labels: componentry.StandardLabels,
		},
	}
}
