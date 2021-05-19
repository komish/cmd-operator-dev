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

// reconcileRoles will reconcile the Role resources for a given CertManagerDeployment resource.
func (r *CertManagerDeploymentReconciler) reconcileRoles(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: roles")
	defer reqLogger.Info("Ending reconciliation: roles")
	var err error

	getter := ResourceGetter{CustomResource: *instance}
	roles := getter.GetRoles()

	for _, role := range roles {
		if err := controllerutil.SetControllerReference(instance, role, r.Scheme); err != nil {
			return err
		}
		found := &rbacv1.Role{}
		err = r.Get(context.TODO(), types.NamespacedName{Name: role.GetName(), Namespace: role.GetNamespace()}, found)

		if err != nil && apierrors.IsNotFound(err) {
			reqLogger.Info("Creating role", "Role.Namespace", role.Namespace, "Role.Name", role.Name)
			r.Eventf(instance,
				createManagedRole.etype,
				createManagedRole.reason,
				"%s: %s/%s",
				createManagedRole.message,
				role.GetNamespace(),
				role.GetName())
			if err = r.Create(context.TODO(), role); err != nil {
				return err
			}

			// successful create. Move onto next iteration.
			continue
		} else if err != nil {
			return err
		}

		// A role exists. Determine if it needs updating and do so.
		genRulesInterface, err := cmdoputils.Interfacer{Data: role.Rules}.ToJSONInterface()
		if err != nil { // errors indicate a marshaling problem.
			return err
		}
		foundRulesInterface, err := cmdoputils.Interfacer{Data: found.Rules}.ToJSONInterface()
		if err != nil {
			return err
		}

		genLabelsInterface, err := cmdoputils.Interfacer{Data: role.Labels}.ToJSONInterface()
		if err != nil {
			return err
		}
		foundLabelsInterface, err := cmdoputils.Interfacer{Data: found.Labels}.ToJSONInterface()
		if err != nil {
			return err
		}

		rulesMatch := cmdoputils.ObjectsMatch(genRulesInterface, foundRulesInterface)
		labelsMatch := cmdoputils.ObjectsMatch(genLabelsInterface, foundLabelsInterface)

		if !(rulesMatch && labelsMatch) {
			reqLogger.Info("Role already exists, but needs an update.",
				"Role.Name", role.GetName(),
				"Role.Namespace", role.GetNamespace(),
				"HasExpectedRules", rulesMatch,
				"HasExpectedLabels", labelsMatch)
			r.Eventf(instance,
				updatingManagedRole.etype,
				updatingManagedRole.reason,
				"%s: %s/%s",
				updatingManagedRole.message,
				role.GetNamespace(),
				role.GetName())

			updated := found.DeepCopy()

			if !rulesMatch {
				updated.Rules = role.Rules
			}

			if !labelsMatch {
				// TODO(): should we avoid clobbering and instead just add our labels?
				updated.ObjectMeta.Labels = role.GetLabels()
			}

			reqLogger.Info("Updating Role.", "Role.Name", role.GetName(), "Role.Namespace", role.GetNamespace())
			if err := r.Update(context.TODO(), updated); err != nil {
				return err
			}

			r.Eventf(instance, updatedManagedRole.etype, updatedManagedRole.reason, "%s: %s/%s", updatedManagedRole.message, role.GetNamespace(), role.GetName())
		}
	}

	return err
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

// GetRolesFor will return new role objects for each CertManagerComponent associated
// with the CustomResource.
func GetRolesFor(cr operatorsv1alpha1.CertManagerDeployment) []*rbacv1.Role {
	var roles []*rbacv1.Role
	for _, componentGetterFunc := range componentry.Components {
		component := componentGetterFunc(cmdoputils.CRVersionOrDefaultVersion(
			cr.Spec.Version,
			componentry.CertManagerDefaultVersion))
		for _, role := range component.GetRoles() {
			roles = append(roles, newRole(component, role, cr))
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
