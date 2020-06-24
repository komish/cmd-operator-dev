package certmanagerdeployment

import (
	redhatv1alpha1 "github.com/komish/certmanager-operator/pkg/apis/redhat/v1alpha1"
	"github.com/komish/certmanager-operator/pkg/controller/certmanagerdeployment/cmdoputils"
	"github.com/komish/certmanager-operator/pkg/controller/certmanagerdeployment/componentry"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetRoleBindings will return all RoleBindings for the custom resource.
func (r *ResourceGetter) GetRoleBindings() []*rbacv1.RoleBinding {
	var rbs []*rbacv1.RoleBinding
	for _, componentGetterFunc := range componentry.Components {
		component := componentGetterFunc(cmdoputils.CRVersionOrDefaultVersion(
			r.CustomResource.Spec.Version,
			componentry.CertManagerDefaultVersion))
		for _, role := range component.GetRoles() {
			// need the role and the service account for the CR
			role := newRole(component, role, r.CustomResource)
			sa := newServiceAccount(component, r.CustomResource)
			rbs = append(rbs, newRoleBinding(component, r.CustomResource, role, sa))
		}
	}
	return rbs
}

// newRoleBinding will return a new RoleBinding object for a given CertManagerComponent
func newRoleBinding(comp componentry.CertManagerComponent, cr redhatv1alpha1.CertManagerDeployment, role *rbacv1.Role, sa *corev1.ServiceAccount) *rbacv1.RoleBinding {
	var rb rbacv1.RoleBinding
	rb = rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      role.GetName(),
			Namespace: componentry.CertManagerDeploymentNamespace,
			Labels:    cmdoputils.MergeMaps(comp.GetLabels(), componentry.StandardLabels),
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
func newRole(comp componentry.CertManagerComponent, rd componentry.RoleData, cr redhatv1alpha1.CertManagerDeployment) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rd.GetName(),
			Namespace: componentry.CertManagerDeploymentNamespace,
			Labels:    cmdoputils.MergeMaps(comp.GetLabels(), rd.GetLabels()),
		},
		Rules: rd.GetPolicyRules(),
	}
}

// GetClusterRoles will return all ClusterRoles for CertManageComponents.
func (r *ResourceGetter) GetClusterRoles() []*rbacv1.ClusterRole {
	var result []*rbacv1.ClusterRole
	for _, componentGetterFunc := range componentry.Components {
		component := componentGetterFunc(cmdoputils.CRVersionOrDefaultVersion(
			r.CustomResource.Spec.Version,
			componentry.CertManagerDefaultVersion))
		for _, clusterRole := range component.GetClusterRoles() {
			result = append(result, newClusterRole(component, clusterRole, r.CustomResource))
		}
	}
	return result
}

// newClusterRole returns a ClusterRole for a given CertManagerComponent, RoleData, and CertManagerDeployment
// custom resource.
func newClusterRole(comp componentry.CertManagerComponent, rd componentry.RoleData, cr redhatv1alpha1.CertManagerDeployment) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   rd.GetName(),
			Labels: cmdoputils.MergeMaps(comp.GetLabels(), rd.GetLabels()),
		},
		Rules: rd.GetPolicyRules(),
	}
}

// GetClusterRoleBindings will return new ClusterRoleBinding objects for the CR.
func (r *ResourceGetter) GetClusterRoleBindings() []*rbacv1.ClusterRoleBinding {
	var crbs []*rbacv1.ClusterRoleBinding
	for _, compGetterFunc := range componentry.Components {
		component := compGetterFunc(cmdoputils.CRVersionOrDefaultVersion(
			r.CustomResource.Spec.Version,
			componentry.CertManagerDefaultVersion))
		for _, clusterRole := range component.GetClusterRoles() {
			if !clusterRole.IsAggregate() {
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
func newClusterRoleBinding(comp componentry.CertManagerComponent, cr redhatv1alpha1.CertManagerDeployment, clusterRole *rbacv1.ClusterRole, sa *corev1.ServiceAccount) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterRole.Name,
			Namespace: componentry.CertManagerDeploymentNamespace,
			Labels:    cmdoputils.MergeMaps(comp.GetLabels(), componentry.StandardLabels),
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
