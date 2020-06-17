package certmanagerdeployment

import (
	"strings"

	redhatv1alpha1 "github.com/komish/certmanager-operator/pkg/apis/redhat/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/util/labels"
)

// CertManagerComponent represents the various components of cert-manager that might
// be installed in a cluster.
type CertManagerComponent struct {
	Name               string
	ServiceAccountName string
	Labels             map[string]string
	ClusterRoles       []RoleData
	Roles              []RoleData
	Deployment         appsv1.DeploymentSpec
}

const (
	// CertManagerBaseName is the base name to use for objects that need to
	// include the name in their object names.
	CertManagerBaseName string = "cert-manager"

	// CertManagerDeploymentNamespace is the namespace that is used to deploy
	// namespaced resources (e.g. serviceaccounts, roles) used by the cert-manager
	// controllers.
	CertManagerDeploymentNamespace string = CertManagerBaseName
)

var (
	// standardLabels are the base labels that apply to all CertManagerDeployment-managed
	// resources.
	standardLabels = map[string]string{
		"app": "cert-manager",
	}

	// CAInjector contains all the metadata necessary to deploy the subresources
	// needed to run CAInjector (ServiceAccounts, Roles, etc...)
	CAInjector = CertManagerComponent{
		Name:               "cainjector",
		ServiceAccountName: "cert-manager-cainjector",
		Labels: mergeMaps(map[string]string{
			"app.kubernetes.io/component": "cainjector",
			"app.kubernetes.io/name":      "cainjector",
		}, standardLabels),
		ClusterRoles: []RoleData{clusterRoleDataForCAInjector},
		Roles:        []RoleData{roleForCAInjectorLeaderElection},
		Deployment:   deploySpecForCAInjector,
	}

	// Controller contains all the metadata necessary to deploy the subresources
	// needed to run Controller (ServiceAccounts, Roles, etc...)
	Controller = CertManagerComponent{
		Name:               "controller",
		ServiceAccountName: "cert-manager",
		Labels: mergeMaps(map[string]string{
			"app.kubernetes.io/component": "controller",
			"app.kubernetes.io/name":      "controller",
		}, standardLabels),
		ClusterRoles: []RoleData{
			clusterRoleDataForClusterIssuers,
			clusterRoleDataForIssuers,
			clusterRoleDataForChallenges,
			clusterRoleDataForEdit,
			clusterRoleDataForIngressShim,
			clusterRoleDataForOrders,
			clusterRoleDataForCertificates,
			clusterRoleDataForView,
		},
		Roles:      []RoleData{roleForController},
		Deployment: deploySpecForController,
	}

	// Webhook contains all the metadata necessary to deploy the subresources
	// needed to run Controller (ServiceAccounts, Roles, etc...)
	Webhook = CertManagerComponent{
		Name:               "webhook",
		ServiceAccountName: "cert-manager-webhook",
		Labels: mergeMaps(map[string]string{
			"app.kubernetes.io/component": "webhook",
			"app.kubernetes.io/name":      "webhook",
		}, standardLabels),
		ClusterRoles: []RoleData{},
		Roles:        []RoleData{roleForWebhook},
		Deployment:   deploySpecForWebhook,
	}

	// Components is a slice of all components that might need subresources deployed.
	// Is used to iteratively spin up necessary resources if required.
	Components = []CertManagerComponent{CAInjector, Controller, Webhook}
)

// ResourceGetter facilitates getting various owned resources expected by
// a CertManagerDeployment CR.
type ResourceGetter struct {
	CustomResource redhatv1alpha1.CertManagerDeployment
}

// GetBaseLabelSelector is...TODO(komish)
// TODO(komish): This returns a LabelSelector for a component, but doesn't use the
// component's own pre-defined Labels struct key. Need to check what's using this
// and if we can just derive the LabelSelector from the Labels that are already
// on the component.
func (comp *CertManagerComponent) GetBaseLabelSelector() *metav1.LabelSelector {
	var ls metav1.LabelSelector
	labels.AddLabelToSelector(&ls, "app.kubernetes.io/component", comp.Name)
	labels.AddLabelToSelector(&ls, "app.kubernetes.io/name", comp.Name)
	return &ls
}

// GetResourceName will return a hyphenation of the standard base name and the component name.
func (comp *CertManagerComponent) GetResourceName() string {
	return strings.Join([]string{CertManagerBaseName, comp.Name}, "-")
}

// GetServices will return new services for the CR.
func (r *ResourceGetter) GetServices() {
	// TODO(): Implement Me!
	return
}

//GetCustomResourceDefinitions will return new custom resource definitions for the CR.
func (r *ResourceGetter) GetCustomResourceDefinitions() {
	// TODO(): Implement Me!
	return
}

// GetWebhookConfigurations will return new webhooks for the CR.
func (r *ResourceGetter) GetWebhookConfigurations() {
	// TODO(): Implement Me!
	return
}
