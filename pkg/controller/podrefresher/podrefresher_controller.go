package podrefresher

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	certManagerIssuerKindAnnotation             string = "cert-manager.io/issuer-kind"
	certManagerDeploymentAllowRestartAnnotation string = "certmanagerdeployment.redhat.io/allow-restart"
	certManagerDeploymentRestartLabel           string = "certmanagerdeployment.redhat.io/time-restarted"
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
	// a bounce on a resource that is not a cert-manager-related secret.
	ants := secret.GetAnnotations()
	if _, ok := ants[certManagerIssuerKindAnnotation]; !ok {
		reqLogger.V(2).Info("Secret is not a cert-manager issued certificate. Disregarding.", "Secret.Name", secret.GetName(), "Secret.Namespace", secret.GetNamespace())
		return reconcile.Result{}, nil
	}

	reqLogger.Info("Secret is a cert-manager issued certificate.", "Secret.Name", secret.GetName(), "Secret.Namespace", secret.GetNamespace())

	// If Secret has been updated, try to find deployments in the same namespace that needs to be bounced.
	deployList := appsv1.DeploymentList{}
	err = r.client.List(context.TODO(), &deployList, &client.ListOptions{Namespace: secret.GetNamespace()})
	if err != nil {
		reqLogger.Error(err, "Error listing deployments", "request.Namespace", secret.GetNamespace())
		return reconcile.Result{}, err
	}

	var updateFailed bool
	for _, deploy := range deployList.Items {
		updatedAt := time.Now().Format("2006-1-2.1504")
		if hasAllowRestartAnnotation(deploy.ObjectMeta) && usesSecret(secret, deploy.Spec.Template.Spec) {
			reqLogger.Info("Deployment opted-in to pod refreshes due to cert-manager secret update", "Deployment.Name", deploy.GetName(), "Secret.Name", secret.GetName())
			updatedDeploy := deploy.DeepCopy() // TODO(): do we need to do this?
			updatedDeploy.ObjectMeta.Labels[certManagerDeploymentRestartLabel] = updatedAt
			updatedDeploy.Spec.Template.ObjectMeta.Labels[certManagerDeploymentRestartLabel] = updatedAt
			err := r.client.Update(context.TODO(), updatedDeploy)
			if err != nil {
				reqLogger.Error(err, "Unable to restart deployment.", "Deployment.Name", deploy.GetName())
				updateFailed = true
			}
		}
	}

	// Updating one of the deployments failed, requeue
	// TODO(komish): This requeues if _any_ of the deployments fail, but this would cause a successful deployment to
	// be restarted continuously. Need to requeue but with only the failed deployment.
	if updateFailed {
		reqLogger.Info("Requeuing due to inability to update a deployment associated with certmanager secret", "Secret.Name", secret.GetName(), "Secret.Namespace", secret.GetNamespace())
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// hasAllowRestartAnnotation returns true if the object meta has opted into restarts
// via inclusion of a restart annotation, and false if the annotation is either there and
// isn't set to true, or isn't set at all.
// TODO() allow for arbitrary annotations to be checked in addition to our default for
// backwards compatibility.
func hasAllowRestartAnnotation(metadata metav1.ObjectMeta) bool {
	annotations := metadata.GetAnnotations()
	val, ok := annotations[certManagerDeploymentAllowRestartAnnotation]
	// It's possible the value of the annotation is false, so check that the annotation
	// is both present and true.
	if val == "true" && ok {
		return ok
	}

	return false
}

// usesSecret returns true if podspec contains a volume that is sourced from
// a secret whose name matches the secret parameter, and false if it does not.
func usesSecret(secret *corev1.Secret, podspec corev1.PodSpec) bool {
	vols := podspec.Volumes

	for _, vol := range vols {
		// VolumeSource.Secret is a pointer, if it's uninitialized it should be nil
		if secretRef := vol.VolumeSource.Secret; secretRef != nil {
			// return true if we get a match
			if secretRef.SecretName == secret.GetName() {
				return true
			}
		}
	}
	// We didn't find a volume from a secret that matched our input secret.
	return false
}
