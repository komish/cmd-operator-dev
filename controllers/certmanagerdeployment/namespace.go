package certmanagerdeployment

import (
	"context"

	"github.com/go-logr/logr"
	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
	"github.com/komish/cmd-operator-dev/controllers/componentry"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// GetNamespaceFor returns a namespace object for a given CertManagerDeployment resource.
// This namespace value is standardized, and as such is not dictated by the custom resource.
func GetNamespace() *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   componentry.CertManagerDeploymentNamespace,
			Labels: componentry.StandardLabels,
		},
	}
}

// reconcileNamespace will reconcile the Namespace resources for a given CertManagerDeployment resource.
func (r *CertManagerDeploymentReconciler) reconcileNamespace(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: namespace")
	defer reqLogger.Info("Ending reconciliation: namespace")

	found := &corev1.Namespace{}
	err := r.Get(
		context.TODO(),
		types.NamespacedName{Name: componentry.CertManagerDeploymentNamespace},
		found,
	)

	// Create it if it doesn't exist.
	if err != nil && apierrors.IsNotFound(err) {
		// We didn't find this namespace already, so create it.
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: componentry.CertManagerDeploymentNamespace,
			},
		}
		if err := controllerutil.SetControllerReference(instance, ns, r.Scheme); err != nil {
			return err
		}

		reqLogger.Info("Creating namespace", "Namespace.Name", ns.Name)
		r.Eventf(instance, createManagedNamespace.etype, createManagedNamespace.reason, "%s: %s", createManagedNamespace.message, ns.GetName())
		if err := r.Create(context.TODO(), ns); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}
