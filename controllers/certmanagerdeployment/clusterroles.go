package certmanagerdeployment

import (
	"context"

	"github.com/go-logr/logr"
	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
	"github.com/komish/cmd-operator-dev/cmdoputils"
	"github.com/komish/cmd-operator-dev/controllers/componentry"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileClusterRoles will reconcile the Cluster Role resources for a given CertManagerDeployment resource.
func (r *CertManagerDeploymentReconciler) reconcileClusterRoles(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: cluster roles")
	defer reqLogger.Info("Ending reconciliation: cluster roles")

	getter := ResourceGetter{CustomResource: *instance}
	crls := getter.GetClusterRoles()

	for _, clusterRole := range crls {
		if err := controllerutil.SetControllerReference(instance, clusterRole, r.Scheme); err != nil {
			return err
		}

		found := &rbacv1.ClusterRole{}
		err := r.Get(context.TODO(), types.NamespacedName{
			Name:      clusterRole.GetName(),
			Namespace: clusterRole.Namespace, // this should be empty
		}, found)

		if err != nil && apierrors.IsNotFound(err) {
			reqLogger.Info("Creating Cluster Role", "ClusterRole.Name", clusterRole.GetName())
			r.Eventf(instance,
				createManagedClusterRole.etype,
				createManagedClusterRole.reason,
				"%s: %s",
				createManagedClusterRole.message,
				clusterRole.GetName())
			if err := r.Create(context.TODO(), clusterRole); err != nil {
				return err
			}

			// successful create. Move onto next iteration.
			continue
		} else if err != nil {
			return err
		}

		// cluster role exists, check if it needs an update and update it.
		genRulesInterface, err := cmdoputils.Interfacer{Data: clusterRole.Rules}.ToJSONInterface()
		if err != nil { // err indicates a marshaling error
			return err
		}
		foundRulesInterface, err := cmdoputils.Interfacer{Data: found.Rules}.ToJSONInterface()
		if err != nil {
			return err
		}

		genLabelsInterface, err := cmdoputils.Interfacer{Data: clusterRole.Labels}.ToJSONInterface()
		if err != nil {
			return err
		}
		foundLabelsInterface, err := cmdoputils.Interfacer{Data: clusterRole.Labels}.ToJSONInterface()
		if err != nil {
			return err
		}

		rulesMatch := cmdoputils.ObjectsMatch(genRulesInterface, foundRulesInterface)
		labelsMatch := cmdoputils.ObjectsMatch(genLabelsInterface, foundLabelsInterface)

		if !(rulesMatch && labelsMatch) {
			reqLogger.Info("ClusterRoleBinding already exists, but needs an update.",
				"ClusterRoleBinding.Name", clusterRole.GetName(),
				"HasExpectedSubjects", rulesMatch,
				"HasExpectedLabels", labelsMatch)
			r.Eventf(instance,
				updatingManagedClusterRole.etype,
				updatingManagedClusterRole.reason,
				"%s: %s",
				updatingManagedClusterRole.message,
				clusterRole.GetName())

			updated := found.DeepCopy()

			if !rulesMatch {
				updated.Rules = clusterRole.Rules
			}

			if !labelsMatch {
				// TODO(): should we avoid clobbering and instead just add our labels?
				updated.ObjectMeta.Labels = clusterRole.GetLabels()
			}

			reqLogger.Info("Updating ClusterRole.",
				"ClusterRole.Name", clusterRole.GetName())
			if err := r.Update(context.TODO(), updated); err != nil {
				return err
			}

			r.Eventf(instance,
				updatedManagedClusterRoleBinding.etype,
				updatedManagedClusterRoleBinding.reason,
				"%s: %s",
				updatedManagedClusterRoleBinding.message,
				clusterRole.GetName())
		}
	}

	return nil
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

// GetClusterRolesFor will return all ClusterRoles for CertManageComponents.
func GetClusterRolesFor(cr operatorsv1alpha1.CertManagerDeployment) []*rbacv1.ClusterRole {
	var result []*rbacv1.ClusterRole
	for _, componentGetterFunc := range componentry.Components {
		component := componentGetterFunc(cmdoputils.CRVersionOrDefaultVersion(
			cr.Spec.Version,
			componentry.CertManagerDefaultVersion))
		for _, clusterRole := range component.GetClusterRoles() {
			result = append(result, newClusterRole(component, clusterRole, cr))
		}
	}
	return result
}

// newClusterRole returns a ClusterRole for a given CertManagerComponent, RoleData, and CertManagerDeployment
// custom resource.
func newClusterRole(comp componentry.CertManagerComponent, rd componentry.RoleData, cr operatorsv1alpha1.CertManagerDeployment) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   rd.GetName(),
			Labels: cmdoputils.MergeMaps(comp.GetLabels(), rd.GetLabels()),
		},
		Rules: rd.GetPolicyRules(),
	}
}
