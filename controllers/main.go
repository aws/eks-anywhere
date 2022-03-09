package main

import (
	"context"
	"flag"
	"os"

	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1beta1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/remote"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/aws/eks-anywhere/controllers/controllers"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/features"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

var (
	scheme               = runtime.NewScheme()
	setupLog             = ctrl.Log.WithName("setup")
	metricsAddr          string
	enableLeaderElection bool
	probeAddr            string
	gates                = []string{}
)

const WEBHOOK = "webhook"

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(anywherev1.AddToScheme(scheme))
	utilruntime.Must(releasev1.AddToScheme(scheme))
	utilruntime.Must(clusterv1.AddToScheme(scheme))
	utilruntime.Must(controlplanev1.AddToScheme(scheme))
	utilruntime.Must(vspherev1.AddToScheme(scheme))
	utilruntime.Must(etcdv1.AddToScheme(scheme))
	utilruntime.Must(kubeadmv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	fs.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	fs.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	fs.StringSliceVar(&gates, "feature-gates", []string{}, "A set of key=value pairs that describe feature gates for alpha/experimental features. ")
}

func main() {
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	initFlags(pflag.CommandLine)
	pflag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	features.FeedGates(gates)

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

	// Setup the context that's going to be used in controllers and for the manager.
	ctx := ctrl.SetupSignalHandler()

	setupReconcilers(ctx, mgr)
	setupWebhooks(mgr)
	//+kubebuilder:scaffold:builder
	setupChecks(mgr)

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func setupReconcilers(ctx context.Context, mgr ctrl.Manager) {
	if features.IsActive(features.FullLifecycleAPI()) {
		factory := dependencies.NewFactory()
		deps, err := factory.WithGovc().Build(ctx)
		if err != nil {
			setupLog.Error(err, "unable to build dependencies")
			os.Exit(1)
		}

		tracker, err := remote.NewClusterCacheTracker(
			mgr,
			remote.ClusterCacheTrackerOptions{
				Log:     ctrl.Log.WithName("remote").WithName("ClusterCacheTracker"),
				Indexes: remote.DefaultIndexes,
			},
		)
		if err != nil {
			setupLog.Error(err, "unable to create cluster cache tracker")
			os.Exit(1)
		}

		setupLog.Info("Setting up cluster controller")
		if err := (controllers.NewClusterReconciler(
			mgr.GetClient(),
			ctrl.Log.WithName("controllers").WithName(anywherev1.ClusterKind),
			mgr.GetScheme(),
			deps.Govc,
			tracker,
		)).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", anywherev1.ClusterKind)
			os.Exit(1)
		}

		setupLog.Info("Setting up vspheredatacenter controller")
		if err := (controllers.NewVSphereDatacenterReconciler(
			mgr.GetClient(),
			ctrl.Log.WithName("controllers").WithName(anywherev1.VSphereDatacenterKind),
			mgr.GetScheme(),
			deps.Govc,
		)).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", anywherev1.VSphereDatacenterKind)
			os.Exit(1)
		}

		if err := (controllers.NewVSphereMachineConfigReconciler(
			mgr.GetClient(),
			ctrl.Log.WithName("controllers").WithName(anywherev1.VSphereMachineConfigKind),
			mgr.GetScheme(),
		)).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", anywherev1.VSphereMachineConfigKind)
			os.Exit(1)
		}
	} else {
		setupLog.Info("Setting up legacy cluster controller")
		setupLegacyClusterReconciler(mgr)
	}
}

func setupLegacyClusterReconciler(mgr ctrl.Manager) {
	if err := (controllers.NewClusterReconcilerLegacy(
		mgr.GetClient(),
		ctrl.Log.WithName("controllers").WithName(anywherev1.ClusterKind),
		mgr.GetScheme(),
	)).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create legacy cluster controller", "controller", anywherev1.ClusterKind)
		os.Exit(1)
	}
}

func setupWebhooks(mgr ctrl.Manager) {
	if err := (&anywherev1.Cluster{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", WEBHOOK, anywherev1.ClusterKind)
		os.Exit(1)
	}
	if err := (&anywherev1.VSphereDatacenterConfig{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", WEBHOOK, anywherev1.VSphereDatacenterKind)
		os.Exit(1)
	}
	if err := (&anywherev1.VSphereMachineConfig{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", WEBHOOK, anywherev1.VSphereMachineConfigKind)
		os.Exit(1)
	}

	if err := (&anywherev1.GitOpsConfig{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", WEBHOOK, anywherev1.GitOpsConfigKind)
		os.Exit(1)
	}
	if err := (&anywherev1.OIDCConfig{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", WEBHOOK, anywherev1.OIDCConfigKind)
		os.Exit(1)
	}
	if err := (&anywherev1.AWSIamConfig{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", WEBHOOK, anywherev1.AWSIamConfigKind)
		os.Exit(1)
	}
}

func setupChecks(mgr ctrl.Manager) {
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}
}
