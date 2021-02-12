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

package main

import (
	"flag"
	"os"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	corev1 "k8s.io/api/core/v1"

	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
	"github.com/komish/cmd-operator-dev/controllers/certmanagerdeployment"
	"github.com/komish/cmd-operator-dev/controllers/podrefresher"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	// using CustomResourceDefinitions/v1beta1
	utilruntime.Must(apiextv1.AddToScheme(scheme))

	// using CustomResourceDefinitions/v1
	utilruntime.Must(apiextv1.AddToScheme(scheme))

	utilruntime.Must(operatorsv1alpha1.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	controllerNamePodRefresher := "podrefresh-controller"
	controllerNameCertManagerDeployment := "certmanagerdeployment-controller"
	var metricsAddr string
	var enableLeaderElection bool
	var enablePodRefreshController bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&enablePodRefreshController, "enable-pod-refresher", false, "Enables the Pod Refresher Controller.")

	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "d91c88b3.redhat.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&certmanagerdeployment.CertManagerDeploymentReconciler{
		Client:        mgr.GetClient(),
		Log:           ctrl.Log.WithName("controllers").WithName(controllerNameCertManagerDeployment),
		Scheme:        mgr.GetScheme(),
		EventRecorder: mgr.GetEventRecorderFor(controllerNameCertManagerDeployment),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CertManagerDeployment")
		os.Exit(1)
	}

	// The pod refresher controller was enabled via CLI.
	if enablePodRefreshController {
		setupLog.Info("Pod refresh controller is enabled")
		if err = (&podrefresher.PodRefreshReconciler{
			Client:        mgr.GetClient(),
			Log:           ctrl.Log.WithName("controllers").WithName(controllerNamePodRefresher),
			Scheme:        mgr.GetScheme(),
			EventRecorder: mgr.GetEventRecorderFor(controllerNamePodRefresher),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", controllerNameCertManagerDeployment)
			os.Exit(1)
		}
	} else {
		setupLog.Info("Pod refresh controller is disabled")

	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
