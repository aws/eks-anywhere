package main

import (
	"flag"
	"os"

	// +kubebuilder:scaffold:imports
	"github.com/spf13/pflag"

	etcdv1alpha3 "github.com/mrajashree/etcdadm-controller/api/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	vspherev3 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1alpha3"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/controllers/remote"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/aws/eks-anywhere/controllers/controllers"
	anywherev1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
	gates    = []string{}
)

const (
	WEBHOOK          = "webhook"
	reconcilerV2Gate = "reconcilerV2"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(anywherev1alpha1.AddToScheme(scheme))
	utilruntime.Must(releasev1alpha1.AddToScheme(scheme))
	utilruntime.Must(clusterv1.AddToScheme(scheme))
	utilruntime.Must(controlplanev1.AddToScheme(scheme))
	utilruntime.Must(vspherev3.AddToScheme(scheme))
	utilruntime.Must(etcdv1alpha3.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func InitFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&gates, "feature-gates", []string{}, "A set of key=value pairs that describe feature gates for alpha/experimental features. ")
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	InitFlags(pflag.CommandLine)
	pflag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "f64ae69e.eks.amazonaws.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if gateEnabled(reconcilerV2Gate) {
		// TODO: figure out if we want our own (maybe ligther implementation) of the remote client
		// For now, using CAPI's, which is wrapped around the "tracker". Which seems to be mostly a cache
		tracker, err := remote.NewClusterCacheTracker(
			ctrl.Log.WithName("remote").WithName("ClusterCacheTracker"),
			mgr,
		)
		if err != nil {
			setupLog.Error(err, "unable to create cluster cache tracker")
			os.Exit(1)
		}

		setupLog.Info("Init reconciler V2")
		if err = controllers.NewClusterReconcilerV2(
			mgr.GetClient(),
			tracker,
			ctrl.Log.WithName("controllers").WithName(anywherev1alpha1.ClusterKind),
			mgr.GetScheme(),
		).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create reconciler v2", "controller", anywherev1alpha1.ClusterKind)
			os.Exit(1)
		}
		setupLog.Info("Reconciler V2 init succesfully")
	} else {
		setupLog.Info("Init reconciler V1")
		if err = (controllers.NewClusterReconciler(
			mgr.GetClient(),
			ctrl.Log.WithName("controllers").WithName(anywherev1alpha1.ClusterKind),
			mgr.GetScheme())).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", anywherev1alpha1.ClusterKind)
			os.Exit(1)
		}
		setupLog.Info("Reconciler V1 init succesfully")
	}

	if err = (&anywherev1alpha1.Cluster{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", WEBHOOK, anywherev1alpha1.ClusterKind)
		os.Exit(1)
	}
	if err = (&anywherev1alpha1.VSphereDatacenterConfig{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", WEBHOOK, anywherev1alpha1.VSphereDatacenterKind)
		os.Exit(1)
	}
	if err = (&anywherev1alpha1.VSphereMachineConfig{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", WEBHOOK, anywherev1alpha1.VSphereMachineConfigKind)
		os.Exit(1)
	}

	if err = (&anywherev1alpha1.GitOpsConfig{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", WEBHOOK, anywherev1alpha1.GitOpsConfigKind)
		os.Exit(1)
	}
	if err = (&anywherev1alpha1.OIDCConfig{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", WEBHOOK, anywherev1alpha1.OIDCConfigKind)
		os.Exit(1)
	}
	if err = (&anywherev1alpha1.AWSIamConfig{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", WEBHOOK, anywherev1alpha1.AWSIamConfigKind)
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func gateEnabled(gate string) bool {
	for _, g := range gates {
		if g == gate {
			return true
		}
	}

	return false
}
