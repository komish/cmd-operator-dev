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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	certManagerIssuerKindAnnotation             string = "cert-manager.io/issuer-kind"
	certManagerDeploymentAllowRestartAnnotation string = "certmanagerdeployment.redhat.io/allow-restart"
	certManagerDeploymentRestartLabel           string = "certmanagerdeployment.redhat.io/time-restarted"
)

var (
	// PodRefresherPredicateFuncs help guide the events we want the podrefresh-controller
	// to activate upon.
	PodRefresherPredicateFuncs = predicate.Funcs{
		UpdateFunc:  predicate.ResourceVersionChangedPredicate{}.Update,
		CreateFunc:  func(e event.CreateEvent) bool { return false },
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		GenericFunc: func(e event.GenericEvent) bool { return false },
	}

	// Eventing helpers
	refresh        = podRefresherEvent{reason: "PodRefresh", message: "Associated pods restarted as a cert-manager secret used by the object has changed."}
	refreshFailure = podRefresherEvent{reason: "PodRefreshFailure", message: "Unable to restart pods associated with object due to an API error."}
)

// PodRefreshReconciler reconciles a Secret object
type PodRefreshReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	record.EventRecorder
}

// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=list;update;watch;
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=list;update;watch;
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=list;update;watch;

// Reconcile watches for secrets and if a secret is a certmanager secret, it checks for deployments, statefulsets,
// and daemonsets that may be using the secret and triggers a re-rollout of those objects.
func (r *PodRefreshReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("secret", req.NamespacedName)

	r.Log.Info("Reconciling CertManager TLS Certificates")

	// Fetch secret in the cluster.
	secret := &corev1.Secret{}
	err := r.Get(context.TODO(), req.NamespacedName, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile req.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the req.
		return reconcile.Result{}, err
	}

	// If secret doesn't have cert-manager annotations, stop reconciliing it. This is the failsafe to prevent
	// a bounce on a resource that is not a cert-manager-related secret.
	ants := secret.GetAnnotations()
	if _, ok := ants[certManagerIssuerKindAnnotation]; !ok {
		r.Log.Info("Secret is not a cert-manager issued certificate. Disregarding.", "Secret.Name", secret.GetName(), "Secret.Namespace", secret.GetNamespace())
		return reconcile.Result{}, nil
	}

	r.Log.Info("Secret is a cert-manager issued certificate. Checking deployments/statefulsets/daemonsets using Secret.", "Secret.Name", secret.GetName(), "Secret.Namespace", secret.GetNamespace())

	// If Secret has been updated, try to find deployments in the same namespace that needs to be bounced.
	r.Log.V(2).Info("Looking for deployments in namespace using certificate", "Secret.Name", secret.GetName(), "Secret.Namespace", secret.GetNamespace())
	deployList := appsv1.DeploymentList{}
	err = r.List(context.TODO(), &deployList, &client.ListOptions{Namespace: secret.GetNamespace()})
	if err != nil {
		r.Log.Error(err, "Error listing deployments", "req.Namespace", secret.GetNamespace())
		return reconcile.Result{}, err
	}

	// If Secret has been updated, try to find daemonsets in the same namespace that needs to be bounced.
	r.Log.V(2).Info("Looking for daemonsets in namespace using certificate", "Secret.Name", secret.GetName(), "Secret.Namespace", secret.GetNamespace())
	dsetList := appsv1.DaemonSetList{}
	err = r.List(context.TODO(), &dsetList, &client.ListOptions{Namespace: secret.GetNamespace()})
	if err != nil {
		r.Log.Error(err, "Error listing daemonsets", "req.Namespace", secret.GetNamespace())
		return reconcile.Result{}, err
	}

	// If Secret has been updated, try to find statefulsets in the same namespace that needs to be bounced.
	r.Log.V(2).Info("Looking for statefulset in namespace using certificate", "Secret.Name", secret.GetName(), "Secret.Namespace", secret.GetNamespace())
	stsList := appsv1.StatefulSetList{}
	err = r.List(context.TODO(), &stsList, &client.ListOptions{Namespace: secret.GetNamespace()})
	if err != nil {
		r.Log.Error(err, "Error listing statefulsets", "req.Namespace", secret.GetNamespace())
		return reconcile.Result{}, err
	}

	// Since we are not sending a requeue if a refresh fails, we log it instead.
	refreshErrors := make([]refreshErrorData, 0)
	var updateFailed bool

	// Check deployments in the relevant namespace
	for _, deploy := range deployList.Items {
		r.Log.Info("Checking deployment for usage of certificate found in secret", "Secret", secret.GetName(), "Deployment", deploy.GetName(), "Namespace", secret.GetNamespace()) //debug make higher verbosity level
		updatedAt := time.Now().Format("2006-1-2.1504")
		if hasAllowRestartAnnotation(deploy.ObjectMeta) && usesSecret(secret, deploy.Spec.Template.Spec) {
			r.Event(&deploy, corev1.EventTypeNormal, refresh.reason, refresh.message)
			r.Log.Info("Deployment makes use of secret and has opted-in. Initiating refresh", "Secret", secret.GetName(), "Deployment", deploy.GetName(), "Namespace", secret.GetNamespace())
			updatedDeploy := deploy.DeepCopy()
			updatedDeploy.ObjectMeta.Labels[certManagerDeploymentRestartLabel] = updatedAt
			updatedDeploy.Spec.Template.ObjectMeta.Labels[certManagerDeploymentRestartLabel] = updatedAt
			err := r.Update(context.TODO(), updatedDeploy)
			if err != nil {
				r.Event(&deploy, corev1.EventTypeWarning, refreshFailure.reason, refreshFailure.message)
				r.Log.Error(err, "Unable to restart deployment.", "Deployment.Name", deploy.GetName())
				refreshErrors = append(refreshErrors, refreshErrorData{kind: deploy.Kind, name: deploy.GetName(), namespace: deploy.GetNamespace(), errorMsg: err.Error()})
				updateFailed = true
			}
		}
	}

	// Check daemonsets in the relevant namespace
	for _, dset := range dsetList.Items {
		r.Log.Info("Checking Daemonset for usage of certificate found in secret", "Secret", secret.GetName(), "Daemonset", dset.GetName(), "Namespace", secret.GetNamespace()) //debug make higher verbosity level
		updatedAt := time.Now().Format("2006-1-2.1504")
		if hasAllowRestartAnnotation(dset.ObjectMeta) && usesSecret(secret, dset.Spec.Template.Spec) {
			r.Log.Info("Daemonset makes use of secret and has opted-in. Initiating refresh", "Secret", secret.GetName(), "Daemonset", dset.GetName(), "Namespace", secret.GetNamespace())
			updatedDset := dset.DeepCopy()
			updatedDset.ObjectMeta.Labels[certManagerDeploymentRestartLabel] = updatedAt
			updatedDset.Spec.Template.ObjectMeta.Labels[certManagerDeploymentRestartLabel] = updatedAt
			err := r.Update(context.TODO(), updatedDset)
			if err != nil {
				r.Log.Error(err, "Unable to restart Daemonset.", "Daemonset.Name", dset.GetName())
				r.Event(&dset, corev1.EventTypeWarning, refreshFailure.reason, refreshFailure.message)
				refreshErrors = append(refreshErrors, refreshErrorData{kind: dset.Kind, name: dset.GetName(), namespace: dset.GetNamespace(), errorMsg: err.Error()})
				updateFailed = true
			}
		}
	}

	// Check statefulsets in the relevant namespace
	for _, sts := range stsList.Items {
		r.Log.Info("Checking Statefulset for usage of certificate found in secret", "Secret", secret.GetName(), "Statefulset", sts.GetName(), "Namespace", secret.GetNamespace()) //debug make higher verbosity level
		updatedAt := time.Now().Format("2006-1-2.1504")
		if hasAllowRestartAnnotation(sts.ObjectMeta) && usesSecret(secret, sts.Spec.Template.Spec) {
			r.Log.Info("Statefulset makes use of secret and has opted-in. Initiating refresh", "Secret", secret.GetName(), "Statefulset", sts.GetName(), "Namespace", secret.GetNamespace())
			updatedsts := sts.DeepCopy()
			updatedsts.ObjectMeta.Labels[certManagerDeploymentRestartLabel] = updatedAt
			updatedsts.Spec.Template.ObjectMeta.Labels[certManagerDeploymentRestartLabel] = updatedAt
			err := r.Update(context.TODO(), updatedsts)
			if err != nil {
				r.Log.Error(err, "Unable to restart Statefulset.", "Statefulset.Name", sts.GetName())
				r.Event(&sts, corev1.EventTypeWarning, refreshFailure.reason, refreshFailure.message)
				refreshErrors = append(refreshErrors, refreshErrorData{kind: sts.Kind, name: sts.GetName(), namespace: sts.GetNamespace(), errorMsg: err.Error()})
				updateFailed = true
			}
		}
	}

	// If updating anything failed
	// TODO(komish): This requeues if _any_ of the refreshes fail, but this would cause a successful deployment to
	// be restarted continuously. Need to requeue but with only the failed deployment.
	if updateFailed {
		r.Log.Info("Resource(s) that opted-in to refreshes have failed to refresh but the request will not be requeued",
			"Secret.Name", secret.GetName(),
			"Secret.Namespace", secret.GetNamespace(),
			"Error Message", refreshErrors)
		// return reconcile.Result{}, err // don't uncomment, see above.
	}

	r.Log.Info("Done Reconciling CertManager TLS Certificates")

	return ctrl.Result{}, nil
}

// SetupWithManager configures a controller owned by the manager mgr.
func (r *PodRefreshReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		WithEventFilter(PodRefresherPredicateFuncs).
		Complete(r)
}

// refreshErrorData represents some metadata about an error encountered while trying to
// refresh a deployment, statefulset, or daemonset.
type refreshErrorData struct {
	kind      string
	name      string
	namespace string
	errorMsg  string
}

type podRefresherEvent struct {
	reason  string
	message string
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
