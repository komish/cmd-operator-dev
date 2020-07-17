package certmanagerdeployment

import (
	"context"
	e "errors"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
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
	return &ReconcileCertManagerDeployment{client: mgr.GetClient(), scheme: mgr.GetScheme()}
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
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a CertManagerDeployment object and makes changes based on the state read
// and what is in the CertManagerDeployment.Spec
func (r *ReconcileCertManagerDeployment) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	reqLogger.Info(fmt.Sprintf("Supported cert-manager versions: %s", cmdoputils.GetSupportedCertManagerVersions(componentry.SupportedVersions)))
	reqLogger.Info("Reconciling CertManagerDeployment")
	defer reqLogger.Info("Done Reconciling CertManagerDeployment")

	// This is a temporary check to ensure that only a single instance of cert-manager is installed by the operator.
	// This is an incomplete check because it's perfectly fine crossing namespace boundaries (i.e. you can have a
	// certmanagerdeployment in another namespace with the right name and this would not prevent reconciliation
	// from happening.
	// This may need to be replaced down the line, e.g. when reconciliation requests come from objects that are not
	// the CertManagerDeployment.
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
	if err = r.ReconcileStatus(instance, reqLogger.WithValues("Reconciling", "Status")); err != nil {
		reqLogger.Error(err, "Encountered error reconciling CertManagerDeployment status")
		return reconcile.Result{}, err
	}

	if err = reconcileCRDs(r, instance, reqLogger.WithValues("Reconciling", "CustomResourceDefinitions")); err != nil {
		reqLogger.Error(err, "Encountered error reconciling Custom Resource Definitions.")
		return reconcile.Result{}, err
	}

	if err = reconcileNamespace(r, instance, reqLogger.WithValues("Reconciling", "Namespaces")); err != nil {
		// TODO(?): Is the inclusion of the error here redundant? Should we log the actual error in the reconciler
		// and then use a generic error here?
		reqLogger.Error(err, "Encountered error reconciling Namespace")
		return reconcile.Result{}, err
	}

	if err = reconcileServiceAccounts(r, instance, reqLogger.WithValues("Reconciling", "ServiceAccounts")); err != nil {
		reqLogger.Error(err, "Encountered error reconciling ServiceAccounts")
		return reconcile.Result{}, err
	}

	if err = reconcileRoles(r, instance, reqLogger.WithValues("Reconciling", "Roles")); err != nil {
		reqLogger.Error(err, "Encountered error reconciliing Roles")
		return reconcile.Result{}, err
	}

	if err = reconcileRoleBindings(r, instance, reqLogger.WithValues("Reconciling", "RoleBindings")); err != nil {
		reqLogger.Error(err, "Encountered error reconciling RoleBindings")
		return reconcile.Result{}, err
	}

	if err = reconcileClusterRoles(r, instance, reqLogger.WithValues("Reconciling", "ClusterRoles")); err != nil {
		reqLogger.Error(err, "Encountered error reconciling ClusterRoles")
		return reconcile.Result{}, err
	}

	if err = reconcileClusterRoleBindings(r, instance, reqLogger.WithValues("Reconciling", "ClusterRoleBindings")); err != nil {
		reqLogger.Error(err, "Encountered error reconciling ClusterRoleBindings")
		return reconcile.Result{}, err
	}

	if err = reconcileDeployments(r, instance, reqLogger.WithValues("Reconciling", "Deployments")); err != nil {
		reqLogger.Error(err, "Encountered error reconciling Deployments")
		return reconcile.Result{}, err
	}

	if err = reconcileServices(r, instance, reqLogger.WithValues("Reconciling", "Services")); err != nil {
		reqLogger.Error(err, "Encountered error reconciling Services")
		return reconcile.Result{}, err
	}

	if err = reconcileWebhooks(r, instance, reqLogger.WithValues("Reconciling", "Webhooks")); err != nil {
		reqLogger.Error(err, "Encountered error reconciling Webhooks")
		return reconcile.Result{}, err
	}

	// We had no error in reconciliation so we do not requeue.
	return reconcile.Result{}, nil
}

// reconcileNamespace will get the namespace specifications for a given CertManagerDeployment, add an owner reference to
// the generated specification, then check for the existence or otherwise create that specification. Returns an
// error if encountered, or nil, which helps inform whether the reconcile needs to requeue the request.
func reconcileNamespace(r *ReconcileCertManagerDeployment, instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
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
func reconcileRoles(r *ReconcileCertManagerDeployment, instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
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
			err = r.client.Create(context.TODO(), role)
			if err != nil {
				return err
			}
			// role created successfully, don't requeue
			// success case return at end of function.
		} else if err != nil {
			return err
		}
	}

	return err
}

func reconcileServiceAccounts(r *ReconcileCertManagerDeployment, instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
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
func reconcileRoleBindings(r *ReconcileCertManagerDeployment, instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
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
			err = r.client.Create(context.TODO(), rolebinding)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// reconcileClusterRoleBindings will reconcile the ClusterRoleBindings for a given CertManagerDeployment custom resource.
func reconcileClusterRoleBindings(r *ReconcileCertManagerDeployment, instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
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
			err = r.client.Create(context.TODO(), clusterRoleBinding)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// reconcileClusterRoles will reconcile the ClusterRoles for a given CertManagerDeployment custom resource.
func reconcileClusterRoles(r *ReconcileCertManagerDeployment, instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
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
			reqLogger.Info("Creating Cluster Role", "ClusterRole.Namespace", clusterRole.GetNamespace(), "ClusterRole.Name", clusterRole.GetName())
			if err := r.client.Create(context.TODO(), clusterRole); err != nil {
				return err
			}
			// clusterRole was created successfully so we don't requeue.
		} else if err != nil {
			// We had an error when getting the type, and it was not a NotFound error.
			return err
		}
	}

	return nil
}

// reconcileDeployments will reconcile the Deployment resources for a given CertManagerDeployment CustomResource
func reconcileDeployments(r *ReconcileCertManagerDeployment, instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
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
			if err := r.client.Create(context.TODO(), dep); err != nil {
				return err
			}
		} else if err != nil {
			// we had an error, but it was not a NotFound error.
			return err
		}
	}

	return nil
}

// reconcileServices will reconcile the Service resources for a given CertManagerDeployment CustomResource
func reconcileServices(r *ReconcileCertManagerDeployment, instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
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
			if err := r.client.Create(context.TODO(), svc); err != nil {
				return err
			}
		} else if err != nil {
			// we had an error, but it was not a NotFound error.
			return err
		}
	}

	return nil
}

// reconcileWebhooks will reconcile the Webhook resources for a given CertManagerDeployment CustomResource
func reconcileWebhooks(r *ReconcileCertManagerDeployment, instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
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
			if err := r.client.Create(context.TODO(), mwh); err != nil {
				return err
			}
		} else if err != nil {
			// we had an error, but it was not a NotFound error.
			return err
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
			if err := r.client.Create(context.TODO(), vwh); err != nil {
				return err
			}
		} else if err != nil {
			// we had an error, but it was not a NotFound error.
			return err
		}
	}

	return nil
}

// reconcileCRDs will reconcile custom resource definitions for a given CertManagerDeployment CustomResource
// These will not have ownership ownership and will not be removed on removal of the CertManagerDeployment resource.
// TODO(komish): At some point we need to watch CustomResourceDefinitions
func reconcileCRDs(r *ReconcileCertManagerDeployment, instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {

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
			if err := r.client.Create(context.TODO(), crd); err != nil {
				return err
			}
			continue // we created successfully, move on to the next crd.
		} else if err != nil {
			// we had an error, but it was not a NotFound error.
			return err
		}

		// we found an instance of the CRD already. Do the comparison and if they don't match, recreate.
		// TODO(): possible to add a dry run? Make sure we can update all of them before we start updating any of them?
		// Otherwise might need to consider adding a rollback.
		specsMatch := reflect.DeepEqual(crd.Spec, found.Spec)
		lblsAndAnnotsMatch := cmdoputils.LabelsAndAnnotationsMatch(crd, found)
		if !(specsMatch && lblsAndAnnotsMatch) {
			reqLogger.Info("CustomResourceDefinition already exists, but needs an update. Updating.", "CustomResourceDefinition.Name", crd.GetName(), "LabelsAndAnnotationsMatched", lblsAndAnnotsMatch, "SpecsMatch", specsMatch)

			// modify the state of the old object to post to API
			updated := found.DeepCopy()

			if !specsMatch {
				updated.Spec = crd.Spec
			}

			if !lblsAndAnnotsMatch {
				updated.ObjectMeta.Annotations = crd.GetAnnotations()
				updated.ObjectMeta.Labels = crd.GetLabels()
			}

			if err := r.client.Update(context.TODO(), updated); err != nil {
				// some issue performing the update.
				return err
			}
		}

	}

	return nil
}

// ReconcileStatus reconciles the status block of a CertManagerDeployment
func (r *ReconcileCertManagerDeployment) ReconcileStatus(instance *redhatv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {
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
