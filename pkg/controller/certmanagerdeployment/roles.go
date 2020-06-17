package certmanagerdeployment

import (
	redhatv1alpha1 "github.com/komish/certmanager-operator/pkg/apis/redhat/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RoleData is the relevant metadata that a Role or ClusterRole object
// might have. It could contain a set of Labels that can be used
// to extend the LabelSet that would otherwise exist on the persisted object,
// as well as the set of PolicyRules applicable to the Role/ClusterRole.
type RoleData struct {
	// Name is the name to be used when creating the role or cluster role.
	Name string
	// IsAggregate is an n optional value that indicates that the roles is created with
	// aggregate labels and as such does not need to be associated with a role binding.
	IsAggregate bool
	// Labels are a set of additional labels to add to a specific role or cluster role
	// object.
	Labels map[string]string
	// PolicyRules are the PolicyRule objects for a given role or cluster role, which is
	// in effect the description of permissions.
	PolicyRules []rbacv1.PolicyRule
}

var (
	// roleForCAInjectorLeaderElection contains the RoleData necessary to create the Role for
	// the cainjector deployement.
	roleForCAInjectorLeaderElection = RoleData{
		Name:   "cert-manager-cainjector:leaderelection",
		Labels: map[string]string{},
		PolicyRules: []rbacv1.PolicyRule{
			{
				Verbs:         []string{"get", "update", "patch"},
				APIGroups:     []string{""},
				Resources:     []string{"configmaps"},
				ResourceNames: []string{"cert-manager-cainjector-leader-election", "cert-manager-cainjector-leader-election-core"},
			},
			{
				Verbs:     []string{"create"},
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
			},
		},
	}

	// roleForController contains the RoleData necessary to create the Role for
	// the cert-manager controller deployment.
	roleForController = RoleData{
		Name:   "cert-manager-controller:leaderelection",
		Labels: map[string]string{},
		PolicyRules: []rbacv1.PolicyRule{
			{
				Verbs:         []string{"get", "update", "patch"},
				APIGroups:     []string{""},
				Resources:     []string{"configmaps"},
				ResourceNames: []string{"cert-manager-controller"},
			},
			{
				Verbs:     []string{"create"},
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
			},
		},
	}

	// roleForWebhook contains the RoleData necessary to create the Role for
	// the webhook deployment.
	roleForWebhook = RoleData{
		Name:   "cert-manager-webhook:dynamic-serving",
		Labels: map[string]string{},
		PolicyRules: []rbacv1.PolicyRule{
			{
				Verbs:         []string{"get", "list", "watch", "update"},
				APIGroups:     []string{""},
				Resources:     []string{"secrets"},
				ResourceNames: []string{"cert-manager-webhook-ca"},
			},
			{
				Verbs:     []string{"create"},
				APIGroups: []string{""},
				Resources: []string{"secrets"},
			},
		},
	}
)

// GetRoles will return new role objects for each CertManagerComponent associated
// with the CustomResource.
func (r *ResourceGetter) GetRoles() []*rbacv1.Role {
	var roles []*rbacv1.Role
	for _, component := range Components {
		for _, role := range component.Roles {
			roles = append(roles, newRole(component, role, r.CustomResource))
		}
	}

	return roles
}

// newRoles will return a role for a given component and custom resource.
// Roles are namespaced so the standard CertManagerDeploymentNamespace is
// where these are created.
func newRole(comp CertManagerComponent, rd RoleData, cr redhatv1alpha1.CertManagerDeployment) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rd.Name,
			Namespace: CertManagerDeploymentNamespace,
			Labels:    mergeMaps(comp.Labels, rd.Labels),
		},
		Rules: rd.PolicyRules,
	}
}
