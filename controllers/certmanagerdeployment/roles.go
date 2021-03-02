package certmanagerdeployment

import (
	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
	"github.com/komish/cmd-operator-dev/cmdoputils"
	"github.com/komish/cmd-operator-dev/controllers/componentry"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetRoles will return new role objects for each CertManagerComponent associated
// with the CustomResource.
func (r *ResourceGetter) GetRoles() []*rbacv1.Role {
	var roles []*rbacv1.Role
	for _, componentGetterFunc := range componentry.Components {
		component := componentGetterFunc(cmdoputils.CRVersionOrDefaultVersion(
			r.CustomResource.Spec.Version,
			componentry.CertManagerDefaultVersion))
		for _, role := range component.GetRoles() {
			roles = append(roles, newRole(component, role, r.CustomResource))
		}
	}

	return roles
}

// newRoles will return a role for a given component and custom resource.
// Roles are namespaced so the standard CertManagerDeploymentNamespace is
// where these are created.
func newRole(comp componentry.CertManagerComponent, rd componentry.RoleData, cr operatorsv1alpha1.CertManagerDeployment) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rd.GetName(),
			Namespace: componentry.CertManagerDeploymentNamespace,
			Labels:    cmdoputils.MergeMaps(comp.GetLabels(), rd.GetLabels()),
		},
		Rules: rd.GetPolicyRules(),
	}
}
