package certmanagerdeployment

import (
	redhatv1alpha1 "github.com/komish/certmanager-operator/pkg/apis/redhat/v1alpha1"
	"github.com/komish/certmanager-operator/pkg/controller/certmanagerdeployment/cmdoputils"
	"github.com/komish/certmanager-operator/pkg/controller/certmanagerdeployment/componentry"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetClusterRoles will return all ClusterRoles for CertManageComponents.
func (r *ResourceGetter) GetClusterRoles() []*rbacv1.ClusterRole {
	var result []*rbacv1.ClusterRole
	for _, componentGetterFunc := range componentry.Components {
		component := componentGetterFunc()
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
