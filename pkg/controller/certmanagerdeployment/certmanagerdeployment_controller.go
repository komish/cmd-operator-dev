package certmanagerdeployment

import (
	"context"
	e "errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/imdario/mergo"
	redhatv1alpha1 "github.com/komish/certmanager-operator/pkg/apis/redhat/v1alpha1"
	"github.com/komish/certmanager-operator/pkg/controller/certmanagerdeployment/cmdoputils"
	"github.com/komish/certmanager-operator/pkg/controller/certmanagerdeployment/componentry"
	adregv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_certmanagerdeployment")

// Add creates a new CertManagerDeployment Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCertManagerDeployment{client: mgr.GetClient(), scheme: mgr.GetScheme(), recorder: mgr.GetEventRecorderFor("certmanagerdeployment-controller")}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("certmanagerdeployment-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource CertManagerDeployment
	err = c.Watch(&source.Kind{Type: &redhatv1alpha1.CertManagerDeployment{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch Service Accounts
	if err = c.Watch(&source.Kind{Type: &corev1.ServiceAccount{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &redhatv1alpha1.CertManagerDeployment{}}); err != nil {
		return err
	}

	// Watch Roles
	if err = c.Watch(&source.Kind{Type: &rbacv1.Role{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &redhatv1alpha1.CertManagerDeployment{}}); err != nil {
		return err
	}

	// Watch RoleBindings
	if err := c.Watch(&source.Kind{Type: &rbacv1.RoleBinding{}},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &redhatv1alpha1.CertManagerDeployment{},
		}); err != nil {
		return err
	}

	// Watch ClusterRoles
	if err := c.Watch(&source.Kind{Type: &rbacv1.ClusterRole{}},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &redhatv1alpha1.CertManagerDeployment{},
		}); err != nil {
		return err
	}

	// Watch ClusterRoleBindings
	if err := c.Watch(&source.Kind{Type: &rbacv1.ClusterRoleBinding{}},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &redhatv1alpha1.CertManagerDeployment{},
		}); err != nil {
		return err
	}

	// Watch Deployments
	if err := c.Watch(&source.Kind{Type: &appsv1.Deployment{}},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &redhatv1alpha1.CertManagerDeployment{},
		}); err != nil {
		return err
	}

	// Watch Services
	if err := c.Watch(&source.Kind{Type: &corev1.Service{}},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &redhatv1alpha1.CertManagerDeployment{},
		}); err != nil {
		return err
	}

	// Watch MutatingWebhookConfigurations
	if err := c.Watch(&source.Kind{Type: &adregv1beta1.MutatingWebhookConfiguration{}},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &redhatv1alpha1.CertManagerDeployment{},
		}); err != nil {
		return err
	}

	// Watch ValidatingWebhookConfigurations
	if err := c.Watch(&source.Kind{Type: &adregv1beta1.ValidatingWebhookConfiguration{}},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &redhatv1alpha1.CertManagerDeployment{},
		}); err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileCertManagerDeployment implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCertManagerDeployment{}

// ReconcileCertManagerDeployment reconciles a CertManagerDeployment object
type ReconcileCertManagerDeployment struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a CertManagerDeployment object and makes changes based on the state read
// and what is in the CertManagerDeployment.Spec
func (r *ReconcileCertManagerDeployment) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Name", request.Name)

	reqLogger.Info(fmt.Sprintf("Supported cert-manager versions: %s", cmdoputils.GetSupportedCertManagerVersions(componentry.SupportedVersions)))
	reqLogger.Info("Reconciling CertManagerDeployment")
	defer reqLogger.Info("Done Reconciling CertManagerDeployment")

	// This is a makeshift check to ensure that only a single instance of cert-manager is installed by the operator.
	// TODO(): Improve this check.
	if request.Name != "production" {
		log.Info("Canceling Reconciliation. Only one CertManagerDeployment named production is allowed", "request name", request.Name)
		return reconcile.Result{}, nil
	}

	// Fetch the CertManagerDeployment instance
	instance := &redhatv1alpha1.CertManagerDeployment{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// If the CR requests a version that is not supported by the operator, halt and don't requeue.
	if !cmdoputils.CertManagerVersionIsSupported(instance, componentry.SupportedVersions) {
		reqLogger.Error(e.New("UnsupportedOperandVersion"),
			"the custom resource has defined an unsupported version of cert-manager",
			"version", *instance.Spec.Version,
		)
		return reconcile.Result{}, nil
	}

	// At this point, we should be reconciling a supported version of cert-manager. Log the version
	// associated with the reconciliation request.
	reqLogger.Info(
		fmt.Sprintf("Requested version of cert-manager: %s",
			cmdoputils.CRVersionOrDefaultVersion(
				instance.Spec.Version,
				componentry.CertManagerDefaultVersion),
		),
	)

	// Reconcile all components.
	if err = r.reconcileStatus(instance, reqLogger.WithValues("Reconciling", "Status")); err != nil {
		reqLogger.Error(err, "Encountered error reconciling CertManagerDeployment status")
		return reconcile.Result{}, err
	}

	if err = r.reconcileCRDs(instance, reqLogger.WithValues("Reconciling", "CustomResourceDefinitions")); err != nil {
		reqLogger.Error(err, "Encountered error reconciling Custom Resource Definitions.")
		return reconcile.Result{}, err
	}

	if err = r.reconcileNamespace(instance, reqLogger.WithValues("Reconciling", "Namespaces")); err != nil {
		// TODO(?): Is the inclusion of the error here redundant? Should we log the actual error in the reconciler
		// and then use a generic error here?
		reqLogger.Error(err, "Encountered error reconciling Namespace")
		return reconcile.Result{}, err
	}

	if err = r.reconcileServiceAccounts(instance, reqLogger.WithValues("Reconciling", "ServiceAccounts")); err != nil {
		reqLogger.Error(err, "Encountered error reconciling ServiceAccounts")
		return reconcile.Result{}, err
	}

	if err = r.reconcileRoles(instance, reqLogger.WithValues("Reconciling", "Roles")); err != nil {
		reqLogger.Error(err, "Encountered error reconciliing Roles")
		return reconcile.Result{}, err
	}

	if err = r.reconcileRoleBindings(instance, reqLogger.WithValues("Reconciling", "RoleBindings")); err != nil {
		reqLogger.Error(err, "Encountered error reconciling RoleBindings")
		return reconcile.Result{}, err
	}

	if err = r.reconcileClusterRoles(instance, reqLogger.WithValues("Reconciling", "ClusterRoles")); err != nil {
		reqLogger.Error(err, "Encountered error reconciling ClusterRoles")
		return reconcile.Result{}, err
	}

	if err = r.reconcileClusterRoleBindings(instance, reqLogger.WithValues("Reconciling", "ClusterRoleBindings")); err != nil {
		reqLogger.Error(err, "Encountered error reconciling ClusterRoleBindings")
		return reconcile.Result{}, err
	}

	if err = r.reconcileDeployments(instance, reqLogger.WithValues("Reconciling", "Deployments")); err != nil {
		reqLogger.Error(err, "Encountered error reconciling Deployments")
		return reconcile.Result{}, err
	}

	if err = r.reconcileServices(instance, reqLogger.WithValues("Reconciling", "Services")); err != nil {
		reqLogger.Error(err, "Encountered error reconciling Services")
		return reconcile.Result{}, err
	}

	if err = r.reconcileWebhooks(instance, reqLogger.WithValues("Reconciling", "Webhooks")); err != nil {
		reqLogger.Error(err, "Encountered error reconciling Webhooks")
		return reconcile.Result{}, err
	}

	// We had no error in reconciliation so we do not requeue.
	return reconcile.Result{}, nil
}

// reconcileNamespace will get the namespace specifications for a given CertManagerDeployment, add an owner reference to
// the generated specification, then check for the existence or otherwise create that specification. Returns an
// error if encountered, or nil, which helps inform whether the reconcile needs to requeue the request.
func (r *ReconcileCertManagerDeployment) reconcileNamespace(instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: namespace")
	defer reqLogger.Info("Ending reconciliation: namespace")

	// Check if a namespace with the same name exists already.
	found := &corev1.Namespace{}
	err := r.client.Get(
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
		if err := controllerutil.SetControllerReference(instance, ns, r.scheme); err != nil {
			// failed to set the controller reference
			return err
		}

		reqLogger.Info("Creating namespace", "Namespace.Name", ns.Name)
		r.recorder.Eventf(instance, createManagedNamespace.etype, createManagedNamespace.reason, "%s: %s", createManagedNamespace.message, ns.GetName())
		err = r.client.Create(context.TODO(), ns)
		if err != nil {
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
func (r *ReconcileCertManagerDeployment) reconcileRoles(instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: roles")
	defer reqLogger.Info("Ending reconciliation: roles")
	var err error

	getter := ResourceGetter{CustomResource: *instance}
	roles := getter.GetRoles()

	for _, role := range roles {
		// add controller reference to the generated roles for this CR.
		if err := controllerutil.SetControllerReference(instance, role, r.scheme); err != nil {
			return err
		}
		found := &rbacv1.Role{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: role.Name, Namespace: role.Namespace}, found)

		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating role", "Role.Namespace", role.Namespace, "Role.Name", role.Name)
			r.recorder.Eventf(instance,
				createManagedRole.etype,
				createManagedRole.reason,
				"%s: %s/%s",
				createManagedRole.message,
				role.GetNamespace(),
				role.GetName())
			err = r.client.Create(context.TODO(), role)
			if err != nil {
				return err
			}
			// role created successfully, don't requeue
			// success case return at end of function.
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
			r.recorder.Eventf(instance, updatingManagedRole.etype, updatingManagedRole.reason, "%s: %s/%s", updatingManagedRole.message, role.GetNamespace(), role.GetName())

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
			if err := r.client.Update(context.TODO(), updated); err != nil {
				// some issue performing the update.
				return err
			}

			r.recorder.Eventf(instance, updatedManagedRole.etype, updatedManagedRole.reason, "%s: %s/%s", updatedManagedRole.message, role.GetNamespace(), role.GetName())
		}
	}

	return err
}

func (r *ReconcileCertManagerDeployment) reconcileServiceAccounts(instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: service accounts")
	defer reqLogger.Info("Ending reconciliation: service accounts")
	var err error

	getter := ResourceGetter{CustomResource: *instance}
	sas := getter.GetServiceAccounts()

	for _, sa := range sas {
		// add controller references to the generated service accounts for this CR.
		if err := controllerutil.SetControllerReference(instance, sa, r.scheme); err != nil {
			return err
		}
		found := &corev1.ServiceAccount{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: sa.Name, Namespace: sa.Namespace}, found)

		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating service account", "ServiceAccount.Namespace", sa.Namespace, "ServiceAccount.Name", sa.Name)
			r.recorder.Eventf(instance,
				createManagedServiceAccount.etype,
				createManagedServiceAccount.reason,
				"%s: %s/%s",
				createManagedServiceAccount.message,
				sa.GetNamespace(),
				sa.GetName())
			err = r.client.Create(context.TODO(), sa)
			if err != nil {
				return err
			}

			// ServiceAccount created successfully - don't requeue
		} else if err != nil {
			return err
		}
	}

	return err
}

// reconcileRoleBindings will reconcile the Clusterroles for a given CertManagerDeployment custom resource.
func (r *ReconcileCertManagerDeployment) reconcileRoleBindings(instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: role bindings")
	defer reqLogger.Info("Ending reconciliation: role bindings")

	getter := ResourceGetter{CustomResource: *instance}
	rbs := getter.GetRoleBindings()

	for _, rolebinding := range rbs {
		if err := controllerutil.SetControllerReference(instance, rolebinding, r.scheme); err != nil {
			return err
		}

		found := &rbacv1.RoleBinding{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: rolebinding.Name, Namespace: rolebinding.Namespace}, found)

		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating new rolebinding", "RoleBinding.Name", rolebinding.Name,
				"Rolebinding.Namespace", rolebinding.Namespace,
				"Rolebinding.RoleRef.Kind", rolebinding.RoleRef.Kind)
			r.recorder.Eventf(instance,
				createManagedRoleBinding.etype,
				createManagedRoleBinding.reason,
				"%s: %s/%s",
				createManagedRoleBinding.message,
				rolebinding.GetNamespace(),
				rolebinding.GetName())
			err = r.client.Create(context.TODO(), rolebinding)
			if err != nil {
				return err
			}
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
			r.recorder.Eventf(instance, updatingManagedRoleBinding.etype, updatingManagedRoleBinding.reason, "%s: %s/%s", updatingManagedRoleBinding.message, rolebinding.GetNamespace(), rolebinding.GetName())

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
			if err := r.client.Update(context.TODO(), updated); err != nil {
				// some issue performing the update.
				return err
			}

			r.recorder.Eventf(instance,
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
func (r *ReconcileCertManagerDeployment) reconcileClusterRoleBindings(instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: cluster role bindings")
	defer reqLogger.Info("Ending reconciliation: cluster role bindings")

	getter := ResourceGetter{CustomResource: *instance}
	crbs := getter.GetClusterRoleBindings()

	for _, clusterRoleBinding := range crbs {
		if err := controllerutil.SetControllerReference(instance, clusterRoleBinding, r.scheme); err != nil {
			return err

		}
		found := &rbacv1.ClusterRoleBinding{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: clusterRoleBinding.Name}, found)

		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating new clusterrolebinding", "ClusterRoleBinding.Name", clusterRoleBinding.Name,
				"ClusterRolebinding.RoleRef.Kind", clusterRoleBinding.RoleRef.Kind)
			r.recorder.Eventf(instance,
				createManagedClusterRoleBinding.etype,
				createManagedClusterRoleBinding.reason,
				"%s: %s",
				createManagedClusterRoleBinding.message,
				clusterRoleBinding.GetName())
			err = r.client.Create(context.TODO(), clusterRoleBinding)
			if err != nil {
				return err
			}
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
			r.recorder.Eventf(instance,
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
			if err := r.client.Update(context.TODO(), updated); err != nil {
				// some issue performing the update.
				return err
			}

			r.recorder.Eventf(instance,
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
func (r *ReconcileCertManagerDeployment) reconcileClusterRoles(instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: cluster roles")
	defer reqLogger.Info("Ending reconciliation: cluster roles")

	// Get Cluster Roles for CR
	getter := ResourceGetter{CustomResource: *instance}
	crls := getter.GetClusterRoles()

	// set controller references on those objects
	for _, clusterRole := range crls {
		if err := controllerutil.SetControllerReference(instance, clusterRole, r.scheme); err != nil {
			return err
		}

		found := &rbacv1.ClusterRole{}
		err := r.client.Get(context.TODO(), types.NamespacedName{
			Name:      clusterRole.GetName(),
			Namespace: clusterRole.Namespace, // this should be empty
		}, found)

		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating Cluster Role", "ClusterRole.Name", clusterRole.GetName())
			r.recorder.Eventf(instance,
				createManagedClusterRole.etype,
				createManagedClusterRole.reason,
				"%s: %s",
				createManagedClusterRole.message,
				clusterRole.GetName())
			if err := r.client.Create(context.TODO(), clusterRole); err != nil {
				return err
			}
			// clusterRole was created successfully so we don't requeue.
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
			r.recorder.Eventf(instance,
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
			if err := r.client.Update(context.TODO(), updated); err != nil {
				// some issue performing the update.
				return err
			}

			r.recorder.Eventf(instance,
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
func (r *ReconcileCertManagerDeployment) reconcileDeployments(instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: deployments")
	defer reqLogger.Info("Ending reconciliation: deployments")

	// Get Cluster Roles for CR
	getter := ResourceGetter{CustomResource: *instance}
	deps := getter.GetDeployments()

	// set controller reference on those objects
	for _, dep := range deps {
		// we failed to set the controller reference so we return
		if err := controllerutil.SetControllerReference(instance, dep, r.scheme); err != nil {
			return err
		}
		found := &appsv1.Deployment{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: dep.GetNamespace(), Name: dep.GetName()}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating Deployment", "Deployment.Namespace", dep.GetNamespace(), "Deployment.Name", dep.GetName())
			r.recorder.Eventf(instance,
				createManagedDeployment.etype,
				createManagedDeployment.reason,
				"%s: %s/%s",
				createManagedDeployment.message,
				dep.GetNamespace(), dep.GetName())
			if err := r.client.Create(context.TODO(), dep); err != nil {
				return err
			}
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
		labelsMatch := cmdoputils.ObjectsMatch(genLabelsInterface, foundLabelsInterface)
		annotsMatch := cmdoputils.ObjectsMatch(genAnnotsInterface, foundAnnotsInterface)

		if !(specsMatch && labelsMatch && annotsMatch) {
			reqLogger.Info("Deployment already exists, but needs an update.",
				"Deployment.Name", dep.GetName(),
				"Deployment.Namespace", dep.GetNamespace(),
				"HasExpectedLabels", labelsMatch,
				"HasExpectedAnnotation", annotsMatch,
				"HasExpectedSpec", specsMatch)
			r.recorder.Eventf(instance, updatingManagedDeployment.etype, updatingManagedDeployment.reason, "%s: %s/%s", updatingManagedDeployment.message, dep.GetNamespace(), dep.GetName()) // BOOKMARK

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
			if err := r.client.Update(context.TODO(), updated); err != nil {
				// some issue performing the update.
				return err
			}

			r.recorder.Eventf(instance, updatedManagedDeployment.etype, updatedManagedDeployment.reason, "%s: %s/%s", updatedManagedDeployment.message, dep.GetNamespace(), dep.GetName())
		}
	}

	return nil
}

// reconcileServices will reconcile the Service resources for a given CertManagerDeployment CustomResource
func (r *ReconcileCertManagerDeployment) reconcileServices(instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: services")
	defer reqLogger.Info("Ending reconciliation: services")

	// Get Cluster Roles for CR
	getter := ResourceGetter{CustomResource: *instance}
	svcs := getter.GetServices()

	// set controller reference on those objects
	for _, svc := range svcs {
		// we failed to set the controller reference so we return
		if err := controllerutil.SetControllerReference(instance, svc, r.scheme); err != nil {
			return err
		}
		found := &corev1.Service{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: svc.GetNamespace(), Name: svc.GetName()}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating Service", "Service.Namespace", svc.GetNamespace(), "Service.Name", svc.GetName())
			r.recorder.Eventf(instance,
				createManagedService.etype,
				createManagedService.reason,
				"%s: %s/%s",
				createManagedService.message,
				svc.GetNamespace(), svc.GetName())
			if err := r.client.Create(context.TODO(), svc); err != nil {
				return err
			}
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
			r.recorder.Eventf(instance, updatingManagedService.etype, updatingManagedService.reason, "%s: %s/%s", updatingManagedService.message, svc.GetNamespace(), svc.GetName())
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
			if err := r.client.Update(context.TODO(), updated); err != nil {
				// some issue performing the update.
				return err
			}

			// successful update!
			r.recorder.Eventf(instance, updatedManagedService.etype, updatedManagedService.reason, "%s: %s/%s", updatedManagedService.message, svc.GetNamespace(), svc.GetName())
		}
	}
	return nil
}

// reconcileWebhooks will reconcile the Webhook resources for a given CertManagerDeployment CustomResource
func (r *ReconcileCertManagerDeployment) reconcileWebhooks(instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
	reqLogger.Info("Starting reconciliation: webhooks")
	defer reqLogger.Info("Ending reconciliation: webhooks")

	// Get Webhooks for CR
	getter := ResourceGetter{CustomResource: *instance}

	mwhs := getter.GetMutatingWebhooks()

	// set controller reference and reconcile MutatingWebhookConfigurations
	for _, mwh := range mwhs {
		if err := controllerutil.SetControllerReference(instance, mwh, r.scheme); err != nil {
			// we failed to set the controller reference so we return
			return err
		}
		found := &adregv1beta1.MutatingWebhookConfiguration{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: mwh.GetName()}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating MutatingWebhookConfiguration", "MutatingWebhookConfiguration.Name", mwh.GetName())
			r.recorder.Eventf(instance,
				createManagedWebhook.etype,
				createManagedWebhook.reason,
				"%s: %s",
				createManagedWebhook.message,
				mwh.GetName())
			if err := r.client.Create(context.TODO(), mwh); err != nil {
				return err
			}
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
			r.recorder.Eventf(instance, updatingManagedWebhook.etype, updatingManagedWebhook.reason, "%s: %s", updatingManagedWebhook.message, mwh.GetName())

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
			if err := r.client.Update(context.TODO(), updated); err != nil {
				// some issue performing the update.
				return err
			}

			r.recorder.Eventf(instance, updatedManagedWebhook.etype, updatedManagedWebhook.reason, "%s: %s", updatedManagedWebhook.message, mwh.GetName())
		}
	}

	vwhs := getter.GetValidatingWebhooks()
	// set controller reference and reconcile ValidatingWebhookConfigurations
	for _, vwh := range vwhs {
		if err := controllerutil.SetControllerReference(instance, vwh, r.scheme); err != nil {
			// we failed to set the controller reference so we return
			return err
		}

		found := &adregv1beta1.ValidatingWebhookConfiguration{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: vwh.GetName()}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating ValidatingWebhookConfiguration", "ValidatingWebhookConfiguration.Name", vwh.GetName())
			r.recorder.Eventf(instance,
				createManagedWebhook.etype,
				createManagedWebhook.reason,
				"%s: %s",
				createManagedWebhook.message,
				vwh.GetName())
			if err := r.client.Create(context.TODO(), vwh); err != nil {
				return err
			}
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
			r.recorder.Eventf(instance, updatingManagedWebhook.etype, updatingManagedWebhook.reason, "%s: %s", updatingManagedWebhook.message, vwh.GetName())

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
			if err := r.client.Update(context.TODO(), updated); err != nil {
				// some issue performing the update.
				return err
			}

			r.recorder.Eventf(instance, updatedManagedWebhook.etype, updatedManagedWebhook.reason, "%s: %s", updatedManagedWebhook.message, vwh.GetName())
		}
	}

	return nil
}

// reconcileCRDs will reconcile custom resource definitions for a given CertManagerDeployment CustomResource
// These will not have ownership ownership and will not be removed on removal of the CertManagerDeployment resource.
// TODO(komish): At some point we need to watch CustomResourceDefinitions
func (r *ReconcileCertManagerDeployment) reconcileCRDs(instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {

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
		found := &apiextv1beta1.CustomResourceDefinition{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: crd.GetName()}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating CustomResourceDefinition", "CustomResourceDefinition.Name", crd.GetName())
			r.recorder.Eventf(instance, createManagedCRD.etype, createManagedCRD.reason, "%s: %s", createManagedCRD.message, crd.GetName())

			if err := r.client.Create(context.TODO(), crd); err != nil {
				return err
			}
			continue // we created successfully, move on to the next crd.
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
			r.recorder.Eventf(instance, updatingManagedCRD.etype, updatingManagedCRD.reason, "%s: %s", updatingManagedCRD.message, crd.GetName())

			// modify the state of the old object to post to API
			updated := found.DeepCopy()

			if !specsMatch {
				// update our local copy with values to keys as defined in our generated spec.
				err := mergo.Merge(&updated.Spec, crd.Spec, mergo.WithOverride)
				if err != nil {
					// Some problem merging the specs
					return err
				}
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
			if err := r.client.Update(context.TODO(), updated); err != nil {
				// some issue performing the update.
				return err
			}

			r.recorder.Eventf(instance, updatedManagedCRD.etype, updatedManagedCRD.reason, "%s: %s", updatedManagedCRD.message, crd.GetName())
		}
	}

	return nil
}

// ReconcileStatus reconciles the status block of a CertManagerDeployment
func (r *ReconcileCertManagerDeployment) reconcileStatus(instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
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
	if err := r.client.Status().Update(context.TODO(), obj); err != nil {
		reqLogger.Info("Error updating CertManagerDeployment's Status", "name", instance.GetName())
		return err
	}

	return nil
}
