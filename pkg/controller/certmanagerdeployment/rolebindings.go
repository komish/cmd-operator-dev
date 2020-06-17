package certmanagerdeployment

import (
	redhatv1alpha1 "github.com/komish/certmanager-operator/pkg/apis/redhat/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetRoleBindings will return all RoleBindings for the custom resource.
func (r *ResourceGetter) GetRoleBindings() []*rbacv1.RoleBinding {
	var rbs []*rbacv1.RoleBinding
	for _, component := range Components {
		for _, role := range component.Roles {
			// need the role and the service account for the CR
			role := newRole(component, role, r.CustomResource)
			sa := newServiceAccount(component, r.CustomResource)
			rbs = append(rbs, newRoleBinding(component, r.CustomResource, role, sa))
		}
	}
	return rbs
}

// newRoleBinding will return a new RoleBinding object for a given CertManagerComponent
func newRoleBinding(comp CertManagerComponent, cr redhatv1alpha1.CertManagerDeployment, role *rbacv1.Role, sa *corev1.ServiceAccount) *rbacv1.RoleBinding {
	var rb rbacv1.RoleBinding
	rb = rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      role.GetName(),
			Namespace: CertManagerDeploymentNamespace,
			Labels:    mergeMaps(comp.Labels, standardLabels),
		},
		Subjects: []rbacv1.Subject{
			{
				// TODO(?): replace hard-coded kind based on the object that's used here. Couldn't find
				// a way to get the right string.
				Kind:      "ServiceAccount",
				Namespace: sa.GetNamespace(),
				Name:      sa.GetName(),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.SchemeGroupVersion.Group,
			Kind:     "Role",
			Name:     role.GetName(),
		},
	}

	return &rb
}
