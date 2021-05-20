package componentry

import (
	rbacv1 "k8s.io/api/rbac/v1"
)

// RoleData is the relevant metadata that a Role or ClusterRole object
// might have. It could contain a set of Labels that can be used
// to extend the LabelSet that would otherwise exist on the persisted object,
// as well as the set of PolicyRules applicable to the Role/ClusterRole.
type RoleData struct {
	// Name is the name to be used when creating the role or cluster role.
	name string
	// IsAggregate is an n optional value that indicates that the roles is created with
	// aggregate labels and as such does not need to be associated with a role binding.
	isAggregate bool
	// Labels are a set of additional labels to add to a specific role or cluster role
	// object.
	labels map[string]string
	// PolicyRules are the PolicyRule objects for a given role or cluster role, which is
	// in effect the description of permissions.
	policyRules []rbacv1.PolicyRule
}

// GetName returns the name of the role described in the RoleData object.
func (rd *RoleData) GetName() string {
	return rd.name
}

// IsAggregate returns whether or not the role described in the RoleData is an
// aggregate role. Aggregate roles are built for aggregation but are not associated
// with any subjects by default.
func (rd *RoleData) IsAggregate() bool {
	return rd.isAggregate
}

// GetLabels returns the labels in map[string]string format used when creating the
// role described RoleData.
func (rd *RoleData) GetLabels() map[string]string {
	return rd.labels
}

// GetPolicyRules returns a list of PolicyRule objects for the RoleData object.
func (rd *RoleData) GetPolicyRules() []rbacv1.PolicyRule {
	return rd.policyRules
}

var (
	// roleForCAInjectorLeaderElection contains the RoleData necessary to create the Role for
	// the cainjector deployement.
	roleForCAInjectorLeaderElection = RoleData{
		name:   "cert-manager-cainjector:leaderelection",
		labels: map[string]string{},
		policyRules: []rbacv1.PolicyRule{
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
		name:   "cert-manager-controller:leaderelection",
		labels: map[string]string{},
		policyRules: []rbacv1.PolicyRule{
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
		name:   "cert-manager-webhook:dynamic-serving",
		labels: map[string]string{},
		policyRules: []rbacv1.PolicyRule{
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

var (
	// clusterRoleDataForClusterIssuers is the RoleData object
	// for the cert-manager-controller-clusterissuers ClusterRole
	clusterRoleDataForClusterIssuers = RoleData{
		name:   "cert-manager-controller-clusterissuers",
		labels: map[string]string{},
		policyRules: []rbacv1.PolicyRule{
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
		name:   "cert-manager-controller-issuers",
		labels: map[string]string{},
		policyRules: []rbacv1.PolicyRule{
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
		name:   "cert-manager-controller-challenges",
		labels: map[string]string{},
		policyRules: []rbacv1.PolicyRule{
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
		name:        "cert-manager-edit",
		isAggregate: true,
		labels: map[string]string{
			"rbac.authorization.k8s.io/aggregate-to-admin": "true",
			"rbac.authorization.k8s.io/aggregate-to-edit":  "true",
		},
		policyRules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"certificates", "certificaterequests", "issuers"},
				Verbs:     []string{"create", "delete", "deletecollection", "patch", "update"},
			},
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"challenges", "orders"},
				Verbs:     []string{"create", "delete", "deletecollection", "patch", "update"},
			},
		},
	}

	// clusterRoleDataForCertificates is the RoleData object
	// for the cert-manager-controller-certificates ClusterRole
	clusterRoleDataForCertificates = RoleData{
		name:   "cert-manager-controller-certificates",
		labels: map[string]string{},
		policyRules: []rbacv1.PolicyRule{
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
		name:   "cert-manager-controller-orders",
		labels: map[string]string{},
		policyRules: []rbacv1.PolicyRule{
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
		name:   "cert-manager-controller-ingress-shim",
		labels: map[string]string{},
		policyRules: []rbacv1.PolicyRule{
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
		name:        "cert-manager-view",
		isAggregate: true,
		labels: map[string]string{
			"rbac.authorization.k8s.io/aggregate-to-admin": "true",
			"rbac.authorization.k8s.io/aggregate-to-edit":  "true",
			"rbac.authorization.k8s.io/aggregate-to-view":  "true",
		},
		policyRules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"certificates", "certificaterequests", "issuers"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"challenges", "orders"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}

	// clusterRoleDataForCAInjector is the RoleData object
	// for the cert-manager-cainjector ClusterRole
	clusterRoleDataForCAInjector = RoleData{
		name:   "cert-manager-cainjector",
		labels: map[string]string{},
		policyRules: []rbacv1.PolicyRule{
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

	// clusterRoleDataForApprover is the RoleData object
	// for the cert-manager-controller-approve:cert-manager-io
	// ClusterRole.
	clusterRoleDataForApprover = RoleData{
		name:   "cert-manager-controller-approve:cert-manager-io",
		labels: map[string]string{},
		policyRules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"cert-manager.io"},
				ResourceNames: []string{
					"issuers.cert-manager.io/*",
					"clusterissuers.cert-manager.io/*",
				},
				Resources: []string{"signers"},
				Verbs:     []string{"approve"},
			},
		},
	}

	// clusterRoleDataForSubjectAccessReviews is the RoleData object
	// for the cert-manager-webhook:subjectaccessreviews ClusterRole.
	clusterRoleDataForSubjectAccessReviews = RoleData{
		name:   "cert-manager-webhook:subjectaccessreviews",
		labels: map[string]string{},
		policyRules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"authorization.k8s.io"},
				Resources: []string{"subjectaccessreviews"},
				Verbs:     []string{"create"},
			},
		},
	}
)
