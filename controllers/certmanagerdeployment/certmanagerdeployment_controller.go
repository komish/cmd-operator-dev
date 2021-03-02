/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package certmanagerdeployment

import (
	"context"
	e "errors"
	"fmt"

	"github.com/go-logr/logr"

	adregv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
	"github.com/komish/cmd-operator-dev/cmdoputils"
	"github.com/komish/cmd-operator-dev/controllers/componentry"
	"k8s.io/client-go/tools/record"
)

// CertManagerDeploymentReconciler reconciles a CertManagerDeployment object
type CertManagerDeploymentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	record.EventRecorder
}

// +kubebuilder:rbac:groups=operators.redhat.io,resources=certmanagerdeployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operators.redhat.io,resources=certmanagerdeployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operators.redhat.io,resources=certmanagerdeployments/finalizers,verbs=update;
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;create;update;patch;watch;
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces;serviceaccounts;services,verbs=get;list;watch;create;update;patch;delete;
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings;clusterroles;clusterrolebindings,verbs=get;list;create;update;patch;watch;bind;escalate;
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations;validatingwebhookconfigurations,verbs=get;list;create;update;patch;watch;
// +kubebuilder:rbac:groups=core,resources=namespaces/finalizers;serviceaccounts/finalizers;services/finalizers,verbs=update;
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations/finalizers;validatingwebhookconfigurations/finalizers,verbs=update;
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles/finalizers;rolebindings/finalizers;clusterroles/finalizers;clusterrolebindings/finalizers,verbs=update;
// +kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update;

// Reconcile compares the desired state of CertManagerDeployment custom resources and works to get
// the existing state to match the desired state.
func (r *CertManagerDeploymentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("certmanagerdeployment", req.NamespacedName)

	r.Log.Info("Reconciling CertManagerDeployment")
	instanceKey := req.NamespacedName

	// ensure only a single instance named "cluster"
	if req.Name != "cluster" {
		r.Log.Info("Canceling Reconciliation. Only one CertManagerDeployment named cluster is allowed", "request name", req.Name)
		return ctrl.Result{}, nil
	}

	// fetch the instance for this request
	instance := &operatorsv1alpha1.CertManagerDeployment{}
	err := r.Get(context.TODO(), instanceKey, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// halt of the requested resource's version is unsupported
	if !cmdoputils.CertManagerVersionIsSupported(instance, componentry.SupportedVersions) {
		r.Log.Error(e.New("UnsupportedOperandVersion"),
			"the custom resource has defined an unsupported version of cert-manager",
			"version", *instance.Spec.Version,
		)
		return ctrl.Result{}, nil
	}

	// continue reconciling with supported version
	r.Log.Info(
		fmt.Sprintf("Requested version of cert-manager: %s",
			cmdoputils.CRVersionOrDefaultVersion(
				instance.Spec.Version,
				componentry.CertManagerDefaultVersion),
		),
	)

	// reconcile all components
	if err = r.reconcileStatus(instance, r.Log.WithValues("Reconciling", "Status")); err != nil {
		r.Log.Error(err, "Encountered error reconciling CertManagerDeployment status")
		return ctrl.Result{}, err
	}

	if err = r.reconcileCRDs(instance, r.Log.WithValues("Reconciling", "CustomResourceDefinitions")); err != nil {
		r.Log.Error(err, "Encountered error reconciling Custom Resource Definitions.")
		return ctrl.Result{}, err
	}

	if err = r.reconcileNamespace(instance, r.Log.WithValues("Reconciling", "Namespaces")); err != nil {
		r.Log.Error(err, "Encountered error reconciling Namespace")
		return ctrl.Result{}, err
	}

	if err = r.reconcileServiceAccounts(instance, r.Log.WithValues("Reconciling", "ServiceAccounts")); err != nil {
		r.Log.Error(err, "Encountered error reconciling ServiceAccounts")
		return ctrl.Result{}, err
	}

	if err = r.reconcileRoles(instance, r.Log.WithValues("Reconciling", "Roles")); err != nil {
		r.Log.Error(err, "Encountered error reconciliing Roles")
		return ctrl.Result{}, err
	}

	if err = r.reconcileRoleBindings(instance, r.Log.WithValues("Reconciling", "RoleBindings")); err != nil {
		r.Log.Error(err, "Encountered error reconciling RoleBindings")
		return ctrl.Result{}, err
	}

	if err = r.reconcileClusterRoles(instance, r.Log.WithValues("Reconciling", "ClusterRoles")); err != nil {
		r.Log.Error(err, "Encountered error reconciling ClusterRoles")
		return ctrl.Result{}, err
	}

	if err = r.reconcileClusterRoleBindings(instance, r.Log.WithValues("Reconciling", "ClusterRoleBindings")); err != nil {
		r.Log.Error(err, "Encountered error reconciling ClusterRoleBindings")
		return ctrl.Result{}, err
	}

	if err = r.reconcileDeployments(instance, r.Log.WithValues("Reconciling", "Deployments")); err != nil {
		r.Log.Error(err, "Encountered error reconciling Deployments")
		return ctrl.Result{}, err
	}

	if err = r.reconcileServices(instance, r.Log.WithValues("Reconciling", "Services")); err != nil {
		r.Log.Error(err, "Encountered error reconciling Services")
		return ctrl.Result{}, err
	}

	if err = r.reconcileWebhooks(instance, r.Log.WithValues("Reconciling", "Webhooks")); err != nil {
		r.Log.Error(err, "Encountered error reconciling Webhooks")
		return ctrl.Result{}, err
	}

	// We had no error in reconciliation so we do not requeue.
	return ctrl.Result{}, nil
}

// SetupWithManager configures a controller owned by the manager mgr.
func (r *CertManagerDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorsv1alpha1.CertManagerDeployment{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&rbacv1.ClusterRole{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&adregv1.MutatingWebhookConfiguration{}).
		Owns(&adregv1.ValidatingWebhookConfiguration{}).
		Complete(r)
}
