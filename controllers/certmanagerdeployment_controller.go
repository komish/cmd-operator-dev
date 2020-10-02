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

package controllers

import (
	"context"
	"encoding/json"
	e "errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/imdario/mergo"

	adregv1 "k8s.io/api/admissionregistration/v1"
	adregv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

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
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;create;update;patch;watch;
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces;serviceaccounts;services,verbs=get;list;watch;create;update;patch;delete;
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings;clusterroles;clusterrolebindings,verbs=get;list;create;update;patch;watch;
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations;validatingwebhookconfigurations,verbs=get;list;create;update;patch;watch;

// Reconcile compares the desired state of CertManagerDeployment custom resources and works to get
// the existing state to match the desired state.]
func (r *CertManagerDeploymentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("certmanagerdeployment", req.NamespacedName)

	r.Log.Info("Reconciling CertManagerDeployment")

	// This is a makeshift check to ensure that only a single instance of cert-manager is installed by the operator.
	// TODO(): Improve this check.
	if req.Name != "production" {
		r.Log.Info("Canceling Reconciliation. Only one CertManagerDeployment named production is allowed", "request name", req.Name)
		return ctrl.Result{}, nil
	}

	// Fetch the CertManagerDeployment instance
	instance := &operatorsv1alpha1.CertManagerDeployment{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// If the CR requests a version that is not supported by the operator, halt and don't requeue.
	if !cmdoputils.CertManagerVersionIsSupported(instance, componentry.SupportedVersions) {
		r.Log.Error(e.New("UnsupportedOperandVersion"),
			"the custom resource has defined an unsupported version of cert-manager",
			"version", *instance.Spec.Version,
		)
		return ctrl.Result{}, nil
	}

	// At this point, we should be reconciling a supported version of cert-manager. Log the version
	// associated with the reconciliation request.
	r.Log.Info(
		fmt.Sprintf("Requested version of cert-manager: %s",
			cmdoputils.CRVersionOrDefaultVersion(
				instance.Spec.Version,
				componentry.CertManagerDefaultVersion),
		),
	)

	// Reconcile all components.
	if err = r.reconcileStatus(instance, r.Log.WithValues("Reconciling", "Status")); err != nil {
		r.Log.Error(err, "Encountered error reconciling CertManagerDeployment status")
		return ctrl.Result{}, err
	}

	if err = r.reconcileCRDs(instance, r.Log.WithValues("Reconciling", "CustomResourceDefinitions")); err != nil {
		r.Log.Error(err, "Encountered error reconciling Custom Resource Definitions.")
		return ctrl.Result{}, err
	}

	if err = r.reconcileNamespace(instance, r.Log.WithValues("Reconciling", "Namespaces")); err != nil {
		// TODO(?): Is the inclusion of the error here redundant? Should we log the actual error in the reconciler
		// and then use a generic error here?
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
		Owns(&adregv1beta1.MutatingWebhookConfiguration{}).
		Owns(&adregv1beta1.ValidatingWebhookConfiguration{}).
		Complete(r)
}

// reconcileNamespace will get the namespace specifications for a given CertManagerDeployment, add an owner reference to
// the generated specification, then check for the existence or otherwise create that specification. Returns an
// error if encountered, or nil, which helps inform whether the reconcile needs to requeue the request.
func (r *CertManagerDeploymentReconciler) reconcileNamespace(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: namespace")
	defer reqLogger.Info("Ending reconciliation: namespace")

	// Check if a namespace with the same name exists already.
	found := &corev1.Namespace{}
	err := r.Get(
		context.TODO(),
		types.NamespacedName{Name: componentry.CertManagerDeploymentNamespace},
		found,
	)

	// Create it if it doesn't exist.
	if err != nil && errors.IsNotFound(err) {
		// We didn't find this namespace already, so create it.
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: componentry.CertManagerDeploymentNamespace,
			},
		}
		if err := controllerutil.SetControllerReference(instance, ns, r.Scheme); err != nil {
			// failed to set the controller reference
			return err
		}

		reqLogger.Info("Creating namespace", "Namespace.Name", ns.Name)
		r.Eventf(instance, createManagedNamespace.etype, createManagedNamespace.reason, "%s: %s", createManagedNamespace.message, ns.GetName())
		if err := r.Create(context.TODO(), ns); err != nil {
			return err
		}
	} else if err != nil {
		// some error occured when trying to see if a namespace with
		// this name already existsed.
		return err
	}

	return nil
}

// reconcileRoles will get all role specifications for a given CertManagerDeployment, add an owner reference to
// the generated specification, then check for the existence or otherwise create that specification. Returns an
// error if encountered, or nil. This helps inform whether the reconciliation loop should requeue.
func (r *CertManagerDeploymentReconciler) reconcileRoles(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: roles")
	defer reqLogger.Info("Ending reconciliation: roles")
	var err error

	getter := ResourceGetter{CustomResource: *instance}
	roles := getter.GetRoles()

	for _, role := range roles {
		// add controller reference to the generated roles for this CR.
		if err := controllerutil.SetControllerReference(instance, role, r.Scheme); err != nil {
			return err
		}
		found := &rbacv1.Role{}
		err = r.Get(context.TODO(), types.NamespacedName{Name: role.GetName(), Namespace: role.GetNamespace()}, found)

		if err != nil && errors.IsNotFound(err) {
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

		// A role exists. Update if necessary.
		genRulesInterface, err := cmdoputils.Interfacer{Data: role.Rules}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundRulesInterface, err := cmdoputils.Interfacer{Data: found.Rules}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}

		genLabelsInterface, err := cmdoputils.Interfacer{Data: role.Labels}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundLabelsInterface, err := cmdoputils.Interfacer{Data: found.Labels}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
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

			// modify the state of the old object to post to API
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
				// some issue performing the update.
				return err
			}

			r.Eventf(instance, updatedManagedRole.etype, updatedManagedRole.reason, "%s: %s/%s", updatedManagedRole.message, role.GetNamespace(), role.GetName())
		}
	}

	return err
}

func (r *CertManagerDeploymentReconciler) reconcileServiceAccounts(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: service accounts")
	defer reqLogger.Info("Ending reconciliation: service accounts")
	var err error

	getter := ResourceGetter{CustomResource: *instance}
	sas := getter.GetServiceAccounts()

	for _, sa := range sas {
		// add controller references to the generated service accounts for this CR.
		if err := controllerutil.SetControllerReference(instance, sa, r.Scheme); err != nil {
			return err
		}
		found := &corev1.ServiceAccount{}
		err = r.Get(context.TODO(), types.NamespacedName{Name: sa.GetName(), Namespace: sa.GetNamespace()}, found)

		if err != nil && errors.IsNotFound(err) {
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

// reconcileRoleBindings will reconcile the Clusterroles for a given CertManagerDeployment custom resource.
func (r *CertManagerDeploymentReconciler) reconcileRoleBindings(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: role bindings")
	defer reqLogger.Info("Ending reconciliation: role bindings")

	getter := ResourceGetter{CustomResource: *instance}
	rbs := getter.GetRoleBindings()

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
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundSubjectsInterface, err := cmdoputils.Interfacer{Data: found.Subjects}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}

		genLabelsInterface, err := cmdoputils.Interfacer{Data: rolebinding.Labels}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundLabelsInterface, err := cmdoputils.Interfacer{Data: rolebinding.Labels}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
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

			// modify the state of the old object to post to API
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
				// some issue performing the update.
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

// reconcileClusterRoleBindings will reconcile the ClusterRoleBindings for a given CertManagerDeployment custom resource.
func (r *CertManagerDeploymentReconciler) reconcileClusterRoleBindings(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: cluster role bindings")
	defer reqLogger.Info("Ending reconciliation: cluster role bindings")

	getter := ResourceGetter{CustomResource: *instance}
	crbs := getter.GetClusterRoleBindings()

	for _, clusterRoleBinding := range crbs {
		if err := controllerutil.SetControllerReference(instance, clusterRoleBinding, r.Scheme); err != nil {
			return err

		}
		found := &rbacv1.ClusterRoleBinding{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: clusterRoleBinding.Name}, found)

		if err != nil && errors.IsNotFound(err) {
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
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundSubjectsInterface, err := cmdoputils.Interfacer{Data: found.Subjects}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}

		genLabelsInterface, err := cmdoputils.Interfacer{Data: clusterRoleBinding.Labels}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundLabelsInterface, err := cmdoputils.Interfacer{Data: clusterRoleBinding.Labels}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
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

			// modify the state of the old object to post to API
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
				// some issue performing the update.
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

// reconcileClusterRoles will reconcile the ClusterRoles for a given CertManagerDeployment custom resource.
func (r *CertManagerDeploymentReconciler) reconcileClusterRoles(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: cluster roles")
	defer reqLogger.Info("Ending reconciliation: cluster roles")

	// Get Cluster Roles for CR
	getter := ResourceGetter{CustomResource: *instance}
	crls := getter.GetClusterRoles()

	// set controller references on those objects
	for _, clusterRole := range crls {
		if err := controllerutil.SetControllerReference(instance, clusterRole, r.Scheme); err != nil {
			return err
		}

		found := &rbacv1.ClusterRole{}
		err := r.Get(context.TODO(), types.NamespacedName{
			Name:      clusterRole.GetName(),
			Namespace: clusterRole.Namespace, // this should be empty
		}, found)

		if err != nil && errors.IsNotFound(err) {
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
			// We had an error when getting the type, and it was not a NotFound error.
			return err
		}

		// cluster role exists, check if it needs an update and update it.
		genRulesInterface, err := cmdoputils.Interfacer{Data: clusterRole.Rules}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundRulesInterface, err := cmdoputils.Interfacer{Data: found.Rules}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}

		genLabelsInterface, err := cmdoputils.Interfacer{Data: clusterRole.Labels}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundLabelsInterface, err := cmdoputils.Interfacer{Data: clusterRole.Labels}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
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

			// modify the state of the old object to post to API
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
				// some issue performing the update.
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

// reconcileDeployments will reconcile the Deployment resources for a given CertManagerDeployment CustomResource
func (r *CertManagerDeploymentReconciler) reconcileDeployments(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: deployments")
	defer reqLogger.Info("Ending reconciliation: deployments")

	// Get Cluster Roles for CR
	getter := ResourceGetter{CustomResource: *instance}
	deps := getter.GetDeployments()

	// set controller reference on those objects
	for _, dep := range deps {
		// we failed to set the controller reference so we return
		if err := controllerutil.SetControllerReference(instance, dep, r.Scheme); err != nil {
			return err
		}
		found := &appsv1.Deployment{}
		err := r.Get(context.TODO(), types.NamespacedName{Namespace: dep.GetNamespace(), Name: dep.GetName()}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating Deployment", "Deployment.Namespace", dep.GetNamespace(), "Deployment.Name", dep.GetName())
			r.Eventf(instance,
				createManagedDeployment.etype,
				createManagedDeployment.reason,
				"%s: %s/%s",
				createManagedDeployment.message,
				dep.GetNamespace(), dep.GetName())
			if err := r.Create(context.TODO(), dep); err != nil {
				return err
			}

			// successful create. Move onto next iteration.
			continue
		} else if err != nil {
			// we had an error, but it was not a NotFound error.
			return err
		}

		// A deployment exists. Update if necessary.
		genSpecInterface, err := cmdoputils.Interfacer{Data: dep.Spec}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundSpecInterface, err := cmdoputils.Interfacer{Data: found.Spec}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}

		genLabelsInterface, err := cmdoputils.Interfacer{Data: dep.Labels}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundLabelsInterface, err := cmdoputils.Interfacer{Data: found.Labels}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}

		genAnnotsInterface, err := cmdoputils.Interfacer{Data: dep.Annotations}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundAnnotsInterface, err := cmdoputils.Interfacer{Data: found.Annotations}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}

		specsMatch := cmdoputils.ObjectsMatch(genSpecInterface, foundSpecInterface)
		d0, _ := json.MarshalIndent(genSpecInterface, "", "    ")
		d1, _ := json.MarshalIndent(genSpecInterface, "", "    ")
		fmt.Println(string(d0), string(d1))
		labelsMatch := cmdoputils.ObjectsMatch(genLabelsInterface, foundLabelsInterface)
		annotsMatch := cmdoputils.ObjectsMatch(genAnnotsInterface, foundAnnotsInterface)

		if !(specsMatch && labelsMatch && annotsMatch) {
			reqLogger.Info("Deployment already exists, but needs an update.",
				"Deployment.Name", dep.GetName(),
				"Deployment.Namespace", dep.GetNamespace(),
				"HasExpectedLabels", labelsMatch,
				"HasExpectedAnnotation", annotsMatch,
				"HasExpectedSpec", specsMatch)
			r.Eventf(instance, updatingManagedDeployment.etype, updatingManagedDeployment.reason, "%s: %s/%s", updatingManagedDeployment.message, dep.GetNamespace(), dep.GetName()) // BOOKMARK

			// modify the state of the old object to post to API
			updated := found.DeepCopy()

			if !specsMatch {
				// update our local copy with values to keys as defined in our generated spec.
				err := mergo.Merge(&updated.Spec, dep.Spec, mergo.WithOverride)
				if err != nil {
					// Some problem merging the specs
					return err
				}
			}

			if !labelsMatch {
				// TODO(): should we avoid clobbering and instead just add our labels?
				updated.ObjectMeta.Labels = dep.GetLabels()
			}

			if !annotsMatch {
				// TODO(): should we avoid clobbering and instead just add our annotations?
				updated.ObjectMeta.Annotations = dep.GetAnnotations()
			}

			reqLogger.Info("Updating Deployment.", "Deployment.Name", dep.GetName(), "Deployment.Namespace", dep.GetNamespace())
			if err := r.Update(context.TODO(), updated); err != nil {
				// some issue performing the update.
				return err
			}

			r.Eventf(instance, updatedManagedDeployment.etype, updatedManagedDeployment.reason, "%s: %s/%s", updatedManagedDeployment.message, dep.GetNamespace(), dep.GetName())
		}
	}

	return nil
}

// reconcileServices will reconcile the Service resources for a given CertManagerDeployment CustomResource
func (r *CertManagerDeploymentReconciler) reconcileServices(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: services")
	defer reqLogger.Info("Ending reconciliation: services")

	// Get Cluster Roles for CR
	getter := ResourceGetter{CustomResource: *instance}
	svcs := getter.GetServices()

	// set controller reference on those objects
	for _, svc := range svcs {
		// we failed to set the controller reference so we return
		if err := controllerutil.SetControllerReference(instance, svc, r.Scheme); err != nil {
			return err
		}
		found := &corev1.Service{}
		err := r.Get(context.TODO(), types.NamespacedName{Namespace: svc.GetNamespace(), Name: svc.GetName()}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating Service", "Service.Namespace", svc.GetNamespace(), "Service.Name", svc.GetName())
			r.Eventf(instance,
				createManagedService.etype,
				createManagedService.reason,
				"%s: %s/%s",
				createManagedService.message,
				svc.GetNamespace(), svc.GetName())
			if err := r.Create(context.TODO(), svc); err != nil {
				return err
			}

			// successful create. Move onto next iteration.
			continue
		} else if err != nil {
			// we had an error, but it was not a NotFound error.
			return err
		}

		// the service exists. If it needs updating, update it.
		// TODO(komish): move update logic to its own function?
		genSpecInterface, err := cmdoputils.Interfacer{Data: svc.Spec}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundSpecInterface, err := cmdoputils.Interfacer{Data: found.Spec}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}

		genLabelsInterface, err := cmdoputils.Interfacer{Data: svc.Labels}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundLabelsInterface, err := cmdoputils.Interfacer{Data: found.Labels}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}

		genAnnotsInterface, err := cmdoputils.Interfacer{Data: svc.Annotations}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundAnnotsInterface, err := cmdoputils.Interfacer{Data: found.Annotations}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}

		specsMatch := cmdoputils.ObjectsMatch(genSpecInterface, foundSpecInterface)
		labelsMatch := cmdoputils.ObjectsMatch(genLabelsInterface, foundLabelsInterface)
		annotsMatch := cmdoputils.ObjectsMatch(genAnnotsInterface, foundAnnotsInterface)
		if !(specsMatch && labelsMatch && annotsMatch) {
			reqLogger.Info("Service already exists, but needs an update.",
				"Service.Name", svc.GetName(),
				"HasExpectedLabels", labelsMatch,
				"HasExpectedAnnotation", annotsMatch,
				"HasExpectedSpec", specsMatch)
			r.Eventf(instance, updatingManagedService.etype, updatingManagedService.reason, "%s: %s/%s", updatingManagedService.message, svc.GetNamespace(), svc.GetName())
			// TODO(komish): Uncomment when you've determined how to get the resource version after an update as you would in client-go
			// originalResourceVersion := found.GetResourceVersion()

			// modify the state of the old object to post to API
			updated := found.DeepCopy()

			if !specsMatch {
				// update our local copy with values to keys as defined in our generated spec.
				err := mergo.Merge(&updated.Spec, svc.Spec, mergo.WithOverride)
				if err != nil {
					// Some problem merging the specs
					return err
				}
			}

			if !labelsMatch {
				// TODO(): should we avoid clobbering and instead just add our labels?
				updated.ObjectMeta.Labels = svc.GetLabels()
			}

			if !annotsMatch {
				// TODO(): should we avoid clobbering and instead just add our annotations?
				updated.ObjectMeta.Annotations = svc.GetAnnotations()
			}

			reqLogger.Info("Updating Service.", "Service.Name", svc.GetName(), "Service.Namespace", svc.GetNamespace())
			if err := r.Update(context.TODO(), updated); err != nil {
				// some issue performing the update.
				return err
			}

			// successful update!
			r.Eventf(instance, updatedManagedService.etype, updatedManagedService.reason, "%s: %s/%s", updatedManagedService.message, svc.GetNamespace(), svc.GetName())
		}
	}
	return nil
}

// reconcileWebhooks will reconcile the Webhook resources for a given CertManagerDeployment CustomResource
func (r *CertManagerDeploymentReconciler) reconcileWebhooks(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: webhooks")
	defer reqLogger.Info("Ending reconciliation: webhooks")

	// Get Webhooks for CR
	getter := ResourceGetter{CustomResource: *instance}

	mwhs := getter.GetMutatingWebhooks()

	// set controller reference and reconcile MutatingWebhookConfigurations
	for _, mwh := range mwhs {
		if err := controllerutil.SetControllerReference(instance, mwh, r.Scheme); err != nil {
			// we failed to set the controller reference so we return
			return err
		}
		found := &adregv1.MutatingWebhookConfiguration{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: mwh.GetName()}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating MutatingWebhookConfiguration", "MutatingWebhookConfiguration.Name", mwh.GetName())
			r.Eventf(instance,
				createManagedWebhook.etype,
				createManagedWebhook.reason,
				"%s: %s",
				createManagedWebhook.message,
				mwh.GetName())
			if err := r.Create(context.TODO(), mwh); err != nil {
				return err
			}

			// successful create. Move onto next iteration.
			continue
		} else if err != nil {
			// we had an error, but it was not a NotFound error.
			return err
		}

		// Update the mutating webhook if needed
		genLabelsInterface, err := cmdoputils.Interfacer{Data: mwh.Labels}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundLabelsInterface, err := cmdoputils.Interfacer{Data: found.Labels}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}

		genAnnotsInterface, err := cmdoputils.Interfacer{Data: mwh.Annotations}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundAnnotsInterface, err := cmdoputils.Interfacer{Data: found.Annotations}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}

		genWebhooksInterface, err := cmdoputils.Interfacer{Data: mwh.Webhooks}.ToJSONInterface()
		if err != nil {
			return err
		}

		foundWebhooksInterface, err := cmdoputils.Interfacer{Data: found.Webhooks}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}

		// Check for equality
		labelsMatch := cmdoputils.ObjectsMatch(genLabelsInterface, foundLabelsInterface)
		annotsMatch := cmdoputils.ObjectsMatch(genAnnotsInterface, foundAnnotsInterface)
		webhooksMatch := cmdoputils.ObjectsMatch(genWebhooksInterface, foundWebhooksInterface)

		// If not equal, update.
		if !(webhooksMatch && labelsMatch && annotsMatch) {
			reqLogger.Info("MutatingWebhookConfiguration already exists, but needs an update. Updating.",
				"MutatingWebhookConfiguration.Name", mwh.GetName(),
				"HasExpectedWebhooks", webhooksMatch,
				"HasExpectedLabels", labelsMatch,
				"HasExpectedAnnotations", annotsMatch)
			r.Eventf(instance, updatingManagedWebhook.etype, updatingManagedWebhook.reason, "%s: %s", updatingManagedWebhook.message, mwh.GetName())

			// modify the state of the old object to post to API
			updated := found.DeepCopy()

			if !webhooksMatch {
				// TODO(): should we avoid clobbering and instead just add our annotations?
				updated.Webhooks = mwh.Webhooks
			}

			if !labelsMatch {
				// TODO(): should we avoid clobbering and instead just add our labels?
				updated.ObjectMeta.Labels = mwh.GetLabels()
			}

			if !annotsMatch {
				// TODO(): should we avoid clobbering and instead just add our annotations?
				updated.ObjectMeta.Annotations = mwh.GetAnnotations()
			}

			reqLogger.Info("Updating MutatingWebhookConfiguration", "MutatingWebhookConfiguration.Name", mwh.GetName())
			if err := r.Update(context.TODO(), updated); err != nil {
				// some issue performing the update.
				return err
			}

			r.Eventf(instance, updatedManagedWebhook.etype, updatedManagedWebhook.reason, "%s: %s", updatedManagedWebhook.message, mwh.GetName())
		}
	}

	vwhs := getter.GetValidatingWebhooks()
	// set controller reference and reconcile ValidatingWebhookConfigurations
	for _, vwh := range vwhs {
		if err := controllerutil.SetControllerReference(instance, vwh, r.Scheme); err != nil {
			// we failed to set the controller reference so we return
			return err
		}

		found := &adregv1.ValidatingWebhookConfiguration{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: vwh.GetName()}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating ValidatingWebhookConfiguration", "ValidatingWebhookConfiguration.Name", vwh.GetName())
			r.Eventf(instance,
				createManagedWebhook.etype,
				createManagedWebhook.reason,
				"%s: %s",
				createManagedWebhook.message,
				vwh.GetName())
			if err := r.Create(context.TODO(), vwh); err != nil {
				return err
			}

			// successful create. Move onto next iteration.
			continue
		} else if err != nil {
			// we had an error, but it was not a NotFound error.
			return err
		}

		// Update the validating webhook if needed
		genLabelsInterface, err := cmdoputils.Interfacer{Data: vwh.Labels}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundLabelsInterface, err := cmdoputils.Interfacer{Data: found.Labels}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}

		genAnnotsInterface, err := cmdoputils.Interfacer{Data: vwh.Annotations}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundAnnotsInterface, err := cmdoputils.Interfacer{Data: found.Annotations}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}

		genWebhooksInterface, err := cmdoputils.Interfacer{Data: vwh.Webhooks}.ToJSONInterface()
		if err != nil {
			return err
		}

		foundWebhooksInterface, err := cmdoputils.Interfacer{Data: found.Webhooks}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}

		// Check for equality
		labelsMatch := cmdoputils.ObjectsMatch(genLabelsInterface, foundLabelsInterface)
		annotsMatch := cmdoputils.ObjectsMatch(genAnnotsInterface, foundAnnotsInterface)
		webhooksMatch := cmdoputils.ObjectsMatch(genWebhooksInterface, foundWebhooksInterface)

		// If not equal, update.
		if !(webhooksMatch && labelsMatch && annotsMatch) {
			reqLogger.Info("ValidatingWebhookConfiguration already exists, but needs an update. Updating.",
				"ValidatingWebhookConfiguration.Name", vwh.GetName(),
				"HasExpectedWebhooks", webhooksMatch,
				"HasExpectedLabels", labelsMatch,
				"HasExpectedAnnotations", annotsMatch)
			r.Eventf(instance, updatingManagedWebhook.etype, updatingManagedWebhook.reason, "%s: %s", updatingManagedWebhook.message, vwh.GetName())

			// modify the state of the old object to post to API
			updated := found.DeepCopy()

			if !webhooksMatch {
				// TODO(): should we avoid clobbering and instead just add our annotations?
				updated.Webhooks = vwh.Webhooks
			}

			if !labelsMatch {
				// TODO(): should we avoid clobbering and instead just add our labels?
				updated.ObjectMeta.Labels = vwh.GetLabels()
			}

			if !annotsMatch {
				// TODO(): should we avoid clobbering and instead just add our annotations?
				updated.ObjectMeta.Annotations = vwh.GetAnnotations()
			}

			reqLogger.Info("Updating ValidatingWebhookConfiguration", "ValidatingWebhookConfiguration.Name", vwh.GetName())
			if err := r.Update(context.TODO(), updated); err != nil {
				// some issue performing the update.
				return err
			}

			r.Eventf(instance, updatedManagedWebhook.etype, updatedManagedWebhook.reason, "%s: %s", updatedManagedWebhook.message, vwh.GetName())
		}
	}

	return nil
}

// reconcileCRDs will reconcile custom resource definitions for a given CertManagerDeployment CustomResource
// These will not have ownership ownership and will not be removed on removal of the CertManagerDeployment resource.
// TODO(komish): At some point we need to watch CustomResourceDefinitions
func (r *CertManagerDeploymentReconciler) reconcileCRDs(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {

	reqLogger.Info("Starting reconciliation: CRDs")
	defer reqLogger.Info("Ending reconciliation: CRDs")

	// Get Webhooks for CR
	getter := ResourceGetter{CustomResource: *instance}
	crds, err := getter.GetCRDs()
	if err != nil {
		reqLogger.Error(err, "Failed to get CRDs")
		// Something happened when trying to get CRDs for this reconciliation
		return err
	}

	for _, crd := range crds {
		found := &apiextv1.CustomResourceDefinition{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: crd.GetName()}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating CustomResourceDefinition", "CustomResourceDefinition.Name", crd.GetName())
			r.Eventf(instance, createManagedCRD.etype, createManagedCRD.reason, "%s: %s", createManagedCRD.message, crd.GetName())

			if err := r.Create(context.TODO(), crd); err != nil {
				return err
			}

			// successful create. Move onto next iteration.
			continue
		} else if err != nil {
			// we had an error, but it was not a NotFound error.
			return err
		}

		// TODO(): possible to add a dry run? Make sure we can update all of them before we start updating any of them?
		// Otherwise might need to consider adding a rollback.

		// If needed, update CRD.
		genSpecInterface, err := cmdoputils.Interfacer{Data: crd.Spec}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundSpecInterface, err := cmdoputils.Interfacer{Data: found.Spec}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}

		genLabelsInterface, err := cmdoputils.Interfacer{Data: crd.Labels}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundLabelsInterface, err := cmdoputils.Interfacer{Data: found.Labels}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}

		genAnnotsInterface, err := cmdoputils.Interfacer{Data: crd.Annotations}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}
		foundAnnotsInterface, err := cmdoputils.Interfacer{Data: found.Annotations}.ToJSONInterface()
		if err != nil {
			// couldn't convert to an interface, likely means some kind of marshaling problem
			return err
		}

		// Check for equality
		specsMatch := cmdoputils.ObjectsMatch(genSpecInterface, foundSpecInterface)
		labelsMatch := cmdoputils.ObjectsMatch(genLabelsInterface, foundLabelsInterface)
		annotsMatch := cmdoputils.ObjectsMatch(genAnnotsInterface, foundAnnotsInterface)

		// If not equal, update.
		if !(specsMatch && labelsMatch && annotsMatch) {
			reqLogger.Info("CustomResourceDefinition already exists, but needs an update. Updating.",
				"CustomResourceDefinition.Name", crd.GetName(),
				"HasExpectedLabels", labelsMatch,
				"HasExpectedAnnotations", annotsMatch,
				"HasExpectedSpec", specsMatch)
			r.Eventf(instance, updatingManagedCRD.etype, updatingManagedCRD.reason, "%s: %s", updatingManagedCRD.message, crd.GetName())

			// modify the state of the old object to post to API
			updated := found.DeepCopy()

			if !specsMatch {
				// // update our local copy with values to keys as defined in our generated spec.
				// err := mergo.Merge(&updated.Spec, crd.Spec, mergo.WithOverride)
				// if err != nil {
				// 	// Some problem merging the specs
				// 	return err
				// }
				updated.Spec = crd.Spec
			}

			if !labelsMatch {
				// TODO(): should we avoid clobbering and instead just add our labels?
				updated.ObjectMeta.Labels = crd.GetLabels()
			}

			if !annotsMatch {
				// TODO(): should we avoid clobbering and instead just add our annotations?
				updated.ObjectMeta.Annotations = crd.GetAnnotations()
			}

			reqLogger.Info("Updating CustomResourceDefinition.", "CustomResourceDefinition.Name", crd.GetName())
			if err := r.Update(context.TODO(), updated); err != nil {
				// some issue performing the update.
				return err
			}

			r.Eventf(instance, updatedManagedCRD.etype, updatedManagedCRD.reason, "%s: %s", updatedManagedCRD.message, crd.GetName())
		}
	}

	return nil
}

// ReconcileStatus reconciles the status block of a CertManagerDeployment
func (r *CertManagerDeploymentReconciler) reconcileStatus(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: status")
	defer reqLogger.Info("Ending reconciliation: status")

	// get an empty status, copy the object we're working with,
	// and get a getter to query for expected states of resources.
	status := getUninitializedCertManagerDeploymentStatus()
	obj := instance.DeepCopy()
	getter := ResourceGetter{CustomResource: *instance}

	r.reconcileStatusVersion(status, cmdoputils.CRVersionOrDefaultVersion(instance.Spec.Version, componentry.CertManagerDefaultVersion))
	r.reconcileStatusDeploymentsHealthy(status, getter, reqLogger)
	r.reconcileStatusCRDsHealthy(status, getter, reqLogger)
	r.reconcileStatusPhase(status)

	// Update the object with new status
	obj.Status = *status
	reqLogger.V(2).Info("Updating Status for object", "CertManagerDeployment.Name", instance.GetName())
	if err := r.Status().Update(context.TODO(), obj); err != nil {
		reqLogger.Info("Error updating CertManagerDeployment's Status", "name", instance.GetName())
		return err
	}

	return nil
}
