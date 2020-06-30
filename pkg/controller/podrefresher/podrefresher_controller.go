package podrefresher

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	certManagerIssuerKindAnnotation string = "cert-manager.io/issuer-kind"
)

var log = logf.Log.WithName("controller_pod_refresher")

// Add creates a new Pod Refresher Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePodRefresher{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("pod-refresher-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Secret
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcilePodRefresher implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcilePodRefresher{}

// ReconcilePodRefresher ... TODO
type ReconcilePodRefresher struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile ... TODO!
func (r *ReconcilePodRefresher) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling CertManager TLS Certificates")

	// Fetch secret in the cluster.
	secret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), request.NamespacedName, secret)
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

	// If secret doesn't have cert-manager annotations, stop reconciliing it. This is the failsafe to prevent
	// a bounce on a resource that is not a cert-manager secret.
	ants := secret.GetAnnotations()
	if _, ok := ants[certManagerIssuerKindAnnotation]; !ok {
		reqLogger.Info("Secret is not a cert-manager issued certificate. Disregarding.", "Secret.Name", secret.GetName(), "Secret.Namespace", secret.GetNamespace())
		return reconcile.Result{}, nil
	}

	reqLogger.Info("Secret is a cert-manager issued certificate.", "Secret.Name", secret.GetName(), "Secret.Namespace", secret.GetNamespace())

	// If Secret has been updated, try to find deployments in the same namespace that needs to be bounced.
	// IMPLEMENT
	// If Deployment found, check it has opted-in to being bounced, otherwise stop

	// Update Deployment with annotations to force a bounce.

	return reconcile.Result{}, nil
}
