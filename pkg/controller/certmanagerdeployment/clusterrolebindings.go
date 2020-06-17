package certmanagerdeployment

import (
	redhatv1alpha1 "github.com/komish/certmanager-operator/pkg/apis/redhat/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetClusterRoleBindings will return new ClusterRoleBinding objects for the CR.
func (r *ResourceGetter) GetClusterRoleBindings() []*rbacv1.ClusterRoleBinding {
	var crbs []*rbacv1.ClusterRoleBinding
	for _, component := range Components {
		for _, clusterRole := range component.ClusterRoles {
			if !clusterRole.IsAggregate {
				// only create bindings to non-aggregate cluster roles.
				crole := newClusterRole(component, clusterRole, r.CustomResource)
				sa := newServiceAccount(component, r.CustomResource)
				crbs = append(crbs, newClusterRoleBinding(component, r.CustomResource, crole, sa))
			}
		}
	}
	return crbs
}

// newClusterRoleBinding will return a new ClusterRoleBinding object for a given CertManagerComponent.
func newClusterRoleBinding(comp CertManagerComponent, cr redhatv1alpha1.CertManagerDeployment, clusterRole *rbacv1.ClusterRole, sa *corev1.ServiceAccount) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterRole.Name,
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
			Kind:     "ClusterRole",
			Name:     clusterRole.GetName(),
		},
	}
}
