package certmanagerdeployment

import (
	"context"

	"github.com/go-logr/logr"
	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
	"github.com/komish/cmd-operator-dev/cmdoputils"
	"github.com/komish/cmd-operator-dev/controllers/componentry"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileServiceAccounts will reconcile the Service Account resources for a given CertManagerDeployment resource.
func (r *CertManagerDeploymentReconciler) reconcileServiceAccounts(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: service accounts")
	defer reqLogger.Info("Ending reconciliation: service accounts")
	var err error

	sas := GetServiceAccountsFor(*instance)

	for _, sa := range sas {
		if err := controllerutil.SetControllerReference(instance, sa, r.Scheme); err != nil {
			return err
		}
		found := &corev1.ServiceAccount{}
		err = r.Get(context.TODO(), types.NamespacedName{Name: sa.GetName(), Namespace: sa.GetNamespace()}, found)

		if err != nil && apierrors.IsNotFound(err) {
			reqLogger.Info("Creating service account", "ServiceAccount.Namespace", sa.GetNamespace(), "ServiceAccount.Name", sa.GetName())
			r.Eventf(instance,
				createManagedServiceAccount.etype,
				createManagedServiceAccount.reason,
				"%s: %s/%s",
				createManagedServiceAccount.message,
				sa.GetNamespace(),
				sa.GetName())
			if err := r.Create(context.TODO(), sa); err != nil {
				return err
			}

			// successful create. Move onto next iteration.
			continue
		} else if err != nil {
			return err
		}
	}

	return nil
}

// GetServiceAccounts will return new service account objects for the CR.
func GetServiceAccountsFor(cr operatorsv1alpha1.CertManagerDeployment) []*corev1.ServiceAccount {
	var sas []*corev1.ServiceAccount
	for _, componentGetterFunc := range componentry.Components {
		component := componentGetterFunc(cmdoputils.CRVersionOrDefaultVersion(
			cr.Spec.Version,
			componentry.CertManagerDefaultVersion))
		sa := newServiceAccount(component, cr)
		sas = append(sas, sa)
	}
	return sas
}

// newServiceAccount returns a service account object for a custom resource. These service accounts
// are installed in the global target namespace.
func newServiceAccount(comp componentry.CertManagerComponent, cr operatorsv1alpha1.CertManagerDeployment) *corev1.ServiceAccount {
	automount := true

	return &corev1.ServiceAccount{
		// AutomountServiceAccountToken is not set in v1.2.0 but is set in v1.3.z.
		// The default in k8s is true so this should not cause an issue if explicitly set
		AutomountServiceAccountToken: &automount,
		ObjectMeta: metav1.ObjectMeta{
			Name:      comp.GetServiceAccountName(),
			Namespace: componentry.CertManagerDeploymentNamespace,
			Labels:    comp.GetLabels(),
		},
	}
}
