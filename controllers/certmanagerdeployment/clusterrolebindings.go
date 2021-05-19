package certmanagerdeployment

import (
	"context"

	"github.com/go-logr/logr"
	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
	"github.com/komish/cmd-operator-dev/cmdoputils"
	"github.com/komish/cmd-operator-dev/controllers/componentry"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileClusterRoleBindings will reconcile the Cluster Role Bindings resources for a given CertManagerDeployment resource.
func (r *CertManagerDeploymentReconciler) reconcileClusterRoleBindings(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: cluster role bindings")
	defer reqLogger.Info("Ending reconciliation: cluster role bindings")

	crbs := GetClusterRoleBindingsFor(*instance)

	for _, clusterRoleBinding := range crbs {
		if err := controllerutil.SetControllerReference(instance, clusterRoleBinding, r.Scheme); err != nil {
			return err

		}
		found := &rbacv1.ClusterRoleBinding{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: clusterRoleBinding.Name}, found)

		if err != nil && apierrors.IsNotFound(err) {
			reqLogger.Info("Creating new clusterrolebinding", "ClusterRoleBinding.Name", clusterRoleBinding.Name,
				"ClusterRolebinding.RoleRef.Kind", clusterRoleBinding.RoleRef.Kind)
			r.Eventf(instance,
				createManagedClusterRoleBinding.etype,
				createManagedClusterRoleBinding.reason,
				"%s: %s",
				createManagedClusterRoleBinding.message,
				clusterRoleBinding.GetName())
			if err := r.Create(context.TODO(), clusterRoleBinding); err != nil {
				return err
			}

			// successful create. Move onto next iteration.
			continue
		} else if err != nil {
			// we had an error, but it was not a NotFound error.
			return err
		}

		// cluster role binding exists, check if it needs an update and update it.
		genSubjectsInterface, err := cmdoputils.Interfacer{Data: clusterRoleBinding.Subjects}.ToJSONInterface()
		if err != nil { // err indicates an issue marshaling
			return err
		}
		foundSubjectsInterface, err := cmdoputils.Interfacer{Data: found.Subjects}.ToJSONInterface()
		if err != nil {
			return err
		}

		genLabelsInterface, err := cmdoputils.Interfacer{Data: clusterRoleBinding.Labels}.ToJSONInterface()
		if err != nil {
			return err
		}
		foundLabelsInterface, err := cmdoputils.Interfacer{Data: clusterRoleBinding.Labels}.ToJSONInterface()
		if err != nil {
			return err
		}

		subjectsMatch := cmdoputils.ObjectsMatch(genSubjectsInterface, foundSubjectsInterface)
		labelsMatch := cmdoputils.ObjectsMatch(genLabelsInterface, foundLabelsInterface)

		if !(subjectsMatch && labelsMatch) {
			reqLogger.Info("ClusterRoleBinding already exists, but needs an update.",
				"ClusterRoleBinding.Name", clusterRoleBinding.GetName(),
				"HasExpectedSubjects", subjectsMatch,
				"HasExpectedLabels", labelsMatch)
			r.Eventf(instance,
				updatingManagedClusterRoleBinding.etype,
				updatingManagedClusterRoleBinding.reason,
				"%s: %s",
				updatingManagedClusterRoleBinding.message,
				clusterRoleBinding.GetName())

			updated := found.DeepCopy()

			if !subjectsMatch {
				updated.Subjects = clusterRoleBinding.Subjects
			}

			if !labelsMatch {
				// TODO(): should we avoid clobbering and instead just add our labels?
				updated.ObjectMeta.Labels = clusterRoleBinding.GetLabels()
			}

			reqLogger.Info("Updating ClusterRoleBinding.",
				"ClusterRoleBinding.Name", clusterRoleBinding.GetName())
			if err := r.Update(context.TODO(), updated); err != nil {
				return err
			}

			r.Eventf(instance,
				updatedManagedClusterRoleBinding.etype,
				updatedManagedClusterRoleBinding.reason,
				"%s: %s",
				updatedManagedClusterRoleBinding.message,
				clusterRoleBinding.GetName())
		}
	}
	return nil
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

// GetClusterRoleBindingsFor will return new ClusterRoleBinding objects for the CR.
func GetClusterRoleBindingsFor(cr operatorsv1alpha1.CertManagerDeployment) []*rbacv1.ClusterRoleBinding {
	var crbs []*rbacv1.ClusterRoleBinding

	for _, compGetterFunc := range componentry.Components {
		component := compGetterFunc(cmdoputils.CRVersionOrDefaultVersion(
			cr.Spec.Version,
			componentry.CertManagerDefaultVersion))
		for _, clusterRole := range component.GetClusterRoles() {
			if !clusterRole.IsAggregate() {
				// only create bindings to non-aggregate cluster roles.
				crole := newClusterRole(component, clusterRole, cr)
				sa := newServiceAccount(component, cr)
				crbs = append(crbs, newClusterRoleBinding(component, cr, crole, sa))
			}
		}
	}

	return crbs
}

// newClusterRoleBinding will return a new ClusterRoleBinding object for a given CertManagerComponent.
func newClusterRoleBinding(comp componentry.CertManagerComponent, cr operatorsv1alpha1.CertManagerDeployment, clusterRole *rbacv1.ClusterRole, sa *corev1.ServiceAccount) *rbacv1.ClusterRoleBinding {
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
