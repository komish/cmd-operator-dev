package certmanagerdeployment

import (
	redhatv1alpha1 "github.com/komish/certmanager-operator/pkg/apis/redhat/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// clusterRoleDataForClusterIssuers is the RoleData object
	// for the cert-manager-controller-clusterissuers ClusterRole
	clusterRoleDataForClusterIssuers = RoleData{
		Name:   "cert-manager-controller-clusterissuers",
		Labels: map[string]string{},
		PolicyRules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"update"},
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"clusterissuers", "clusterissuers/status"},
			},
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"clusterissuers"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch"},
			},
		},
	}

	// clusterRoleDataForIssuers is the RoleData object
	// for the cert-manager-controller-issuers ClusterRole
	clusterRoleDataForIssuers = RoleData{
		Name:   "cert-manager-controller-issuers",
		Labels: map[string]string{},
		PolicyRules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"issuers", "issuers/status"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"issuers"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch"},
			},
		},
	}

	// clusterRoleDataForChallenges is the RoleData Object
	// for the cert-manager-controller-challenges ClusterRole
	clusterRoleDataForChallenges = RoleData{
		Name:   "cert-manager-controller-challenges",
		Labels: map[string]string{},
		PolicyRules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"challenges", "challenges/status"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"challenges"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"issuers", "clusterissuers"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "services"},
				Verbs:     []string{"get", "list", "watch", "create", "delete"},
			},
			{
				APIGroups: []string{"extensions"},
				Resources: []string{"ingresses"},
				Verbs:     []string{"get", "list", "watch", "create", "delete", "update"},
			},
			{
				APIGroups: []string{"route.openshift.io"},
				Resources: []string{"routes/custom-host"},
				Verbs:     []string{"create"},
			},
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"challenges/finalizers"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}

	// clusterRoleDataForEdit is the RoleData Object
	// for the cert-manager-edit ClusterRole
	clusterRoleDataForEdit = RoleData{
		Name:        "cert-manager-edit",
		IsAggregate: true,
		Labels: map[string]string{
			"rbac.authorization.k8s.io/aggregate-to-admin": "true",
			"rbac.authorization.k8s.io/aggregate-to-edit":  "true",
		},
		PolicyRules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"certificates", "certificaterequests", "issuers"},
				Verbs:     []string{"create", "delete", "deletecollection", "patch", "update"},
			},
		},
	}

	// clusterRoleDataForCertificates is the RoleData object
	// for the cert-manager-controller-certificates ClusterRole
	clusterRoleDataForCertificates = RoleData{
		Name:   "cert-manager-controller-certificates",
		Labels: map[string]string{},
		PolicyRules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{
					"certificates",
					"certificates/status",
					"certificaterequests",
					"certificaterequests/status",
				},
				Verbs: []string{"update"},
			},
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{
					"certificates",
					"certificaterequests",
					"clusterissuers",
					"issuers",
				},
				Verbs: []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{
					"certificates/finalizers",
					"certificaterequests/finalizers",
				},
				Verbs: []string{"update"},
			},
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"orders"},
				Verbs:     []string{"create", "delete", "get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch"},
			},
		},
	}

	// clusterRoleDataForOrders is the RoleData object
	// for the cert-manager-controller-orders ClusterRole
	clusterRoleDataForOrders = RoleData{
		Name:   "cert-manager-controller-orders",
		Labels: map[string]string{},
		PolicyRules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"orders", "orders/status"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"orders", "challenges"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"clusterissuers", "issuers"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"challenges"},
				Verbs:     []string{"create", "delete"},
			},
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"orders/finalizers"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch"},
			},
		},
	}

	// clusterRoleDataForIngressShim is the RoleData object
	// for the cert-manager-controller-ingress-shim ClusterRole
	clusterRoleDataForIngressShim = RoleData{
		Name:   "cert-manager-controller-ingress-shim",
		Labels: map[string]string{},
		PolicyRules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"certificates", "certificaterequests"},
				Verbs:     []string{"create", "update", "delete"},
			},
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"certificates", "certificaterequests", "issuers", "clusterissuers"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"extensions"},
				Resources: []string{"ingresses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"extensions"},
				Resources: []string{"ingresses/finalizers"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch"},
			},
		},
	}

	// clusterRoleDataForView is the RoleData object
	// for the cert-manager-controller-view ClusterRole
	clusterRoleDataForView = RoleData{
		Name:        "cert-manager-view",
		IsAggregate: true,
		Labels: map[string]string{
			"rbac.authorization.k8s.io/aggregate-to-admin": "true",
			"rbac.authorization.k8s.io/aggregate-to-edit":  "true",
			"rbac.authorization.k8s.io/aggregate-to-view":  "true",
		},
		PolicyRules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"certificates", "certificaterequests", "issuers"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}

	// clusterRoleDataForCAInjector is the RoleData object
	// for the cert-manager-cainjector ClusterRole
	clusterRoleDataForCAInjector = RoleData{
		Name:   "cert-manager-cainjector",
		Labels: map[string]string{},
		PolicyRules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"certificates"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"get", "create", "update", "patch"},
			},
			{
				APIGroups: []string{"apiregistration.k8s.io"},
				Resources: []string{"apiservices"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{"apiextensions.k8s.io"},
				Resources: []string{"customresourcedefinitions"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{"auditregistration.k8s.io"},
				Resources: []string{"auditsinks"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{"admissionregistration.k8s.io"},
				Resources: []string{"validatingwebhookconfigurations", "mutatingwebhookconfigurations"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
		},
	}
)

// GetClusterRoles will return all ClusterRoles for CertManageComponents.
func (r *ResourceGetter) GetClusterRoles() []*rbacv1.ClusterRole {
	var result []*rbacv1.ClusterRole
	for _, component := range Components {
		for _, clusterRole := range component.ClusterRoles {
			result = append(result, newClusterRole(component, clusterRole, r.CustomResource))
		}
	}
	return result
}

// newClusterRole returns a ClusterRole for a given CertManagerComponent, RoleData, and CertManagerDeployment
// custom resource.
func newClusterRole(comp CertManagerComponent, rd RoleData, cr redhatv1alpha1.CertManagerDeployment) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   rd.Name,
			Labels: mergeMaps(comp.Labels, rd.Labels),
		},
		Rules: rd.PolicyRules,
	}
}
