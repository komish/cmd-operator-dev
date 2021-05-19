package certmanagerdeployment

import (
	"context"

	"github.com/go-logr/logr"
	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
	"github.com/komish/cmd-operator-dev/cmdoputils"
	"github.com/komish/cmd-operator-dev/controllers/componentry"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileRoleBindings will reconcile the Role Bindings resources for a given CertManagerDeployment resource.
func (r *CertManagerDeploymentReconciler) reconcileRoleBindings(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: role bindings")
	defer reqLogger.Info("Ending reconciliation: role bindings")

	rbs := GetRoleBindingsFor(*instance)

	for _, rolebinding := range rbs {
		if err := controllerutil.SetControllerReference(instance, rolebinding, r.Scheme); err != nil {
			return err
		}

		found := &rbacv1.RoleBinding{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: rolebinding.Name, Namespace: rolebinding.Namespace}, found)

		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating new rolebinding", "RoleBinding.Name", rolebinding.Name,
				"Rolebinding.Namespace", rolebinding.Namespace,
				"Rolebinding.RoleRef.Kind", rolebinding.RoleRef.Kind)
			r.Eventf(instance,
				createManagedRoleBinding.etype,
				createManagedRoleBinding.reason,
				"%s: %s/%s",
				createManagedRoleBinding.message,
				rolebinding.GetNamespace(),
				rolebinding.GetName())
			if err := r.Create(context.TODO(), rolebinding); err != nil {
				return err
			}

			// successful create. Move onto next iteration.
			continue
		} else if err != nil {
			return err
		}

		// A rolebinding exists. Update if necessary.
		// TODO() RoleRef cannot be updated, need to decide if we want to support changes
		// in which case we need to delete and recreate.
		genSubjectsInterface, err := cmdoputils.Interfacer{Data: rolebinding.Subjects}.ToJSONInterface()
		if err != nil { // errors indicate a marshaling problem
			return err
		}

		foundSubjectsInterface, err := cmdoputils.Interfacer{Data: found.Subjects}.ToJSONInterface()
		if err != nil {
			return err
		}

		genLabelsInterface, err := cmdoputils.Interfacer{Data: rolebinding.Labels}.ToJSONInterface()
		if err != nil {
			return err
		}

		foundLabelsInterface, err := cmdoputils.Interfacer{Data: rolebinding.Labels}.ToJSONInterface()
		if err != nil {
			return err
		}

		subjectsMatch := cmdoputils.ObjectsMatch(genSubjectsInterface, foundSubjectsInterface)
		labelsMatch := cmdoputils.ObjectsMatch(genLabelsInterface, foundLabelsInterface)

		if !(subjectsMatch && labelsMatch) {
			reqLogger.Info("Rolebinding already exists, but needs an update.",
				"RoleBinding.Name", rolebinding.GetName(),
				"RoleBinding.Namespace", rolebinding.GetNamespace(),
				"HasExpectedSubjects", subjectsMatch,
				"HasExpectedLabels", labelsMatch)
			r.Eventf(instance, updatingManagedRoleBinding.etype, updatingManagedRoleBinding.reason, "%s: %s/%s", updatingManagedRoleBinding.message, rolebinding.GetNamespace(), rolebinding.GetName())

			updated := found.DeepCopy()

			if !subjectsMatch {
				updated.Subjects = rolebinding.Subjects
			}

			if !labelsMatch {
				// TODO(): should we avoid clobbering and instead just add our labels?
				updated.ObjectMeta.Labels = rolebinding.GetLabels()
			}

			reqLogger.Info("Updating RoleBinding.",
				"RoleBinding.Name", rolebinding.GetName(),
				"RoleBinding.Namespace", rolebinding.GetNamespace())
			if err := r.Update(context.TODO(), updated); err != nil {
				return err
			}

			r.Eventf(instance,
				updatedManagedRoleBinding.etype,
				updatedManagedRoleBinding.reason,
				"%s: %s/%s",
				updatedManagedRoleBinding.message,
				rolebinding.GetNamespace(),
				rolebinding.GetName())
		}
	}
	return nil
}

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

// GetRoleBindingsFor will return all RoleBindings for the custom resource.
func GetRoleBindingsFor(cr operatorsv1alpha1.CertManagerDeployment) []*rbacv1.RoleBinding {
	var rbs []*rbacv1.RoleBinding
	for _, componentGetterFunc := range componentry.Components {
		component := componentGetterFunc(cmdoputils.CRVersionOrDefaultVersion(
			cr.Spec.Version,
			componentry.CertManagerDefaultVersion))
		for _, role := range component.GetRoles() {
			// need the role and the service account for the CR
			role := newRole(component, role, cr)
			sa := newServiceAccount(component, cr)
			rbs = append(rbs, newRoleBinding(component, cr, role, sa))
		}
	}
	return rbs
}

// newRoleBinding will return a new RoleBinding object for a given CertManagerComponent
func newRoleBinding(comp componentry.CertManagerComponent, cr operatorsv1alpha1.CertManagerDeployment, role *rbacv1.Role, sa *corev1.ServiceAccount) *rbacv1.RoleBinding {
	var rb rbacv1.RoleBinding
	rb = rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      role.GetName(),
			Namespace: componentry.CertManagerDeploymentNamespace,
			Labels:    cmdoputils.MergeMaps(comp.GetLabels(), componentry.StandardLabels),
		},
		Subjects: []rbacv1.Subject{
			{
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
