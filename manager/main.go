package main

import (
	"context"
	"flag"
	"os"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta2"
	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1beta1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/aws/eks-anywhere/controllers"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
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
	utilruntime.Must(clusterctlv1.AddToScheme(scheme))
	utilruntime.Must(controlplanev1.AddToScheme(scheme))
	utilruntime.Must(vspherev1.AddToScheme(scheme))
	utilruntime.Must(cloudstackv1.AddToScheme(scheme))
	utilruntime.Must(dockerv1.AddToScheme(scheme))
	utilruntime.Must(etcdv1.AddToScheme(scheme))
	utilruntime.Must(kubeadmv1.AddToScheme(scheme))
	utilruntime.Must(eksdv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func initFlags(fs *pflag.FlagSet) {
	fs.StringVar(&metricsAddr, "metrics-bind-address", "localhost:8080", "The address the metric endpoint binds to.")
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
		setupLog.Info("Reading CAPI providers")
		providers, err := clusterapi.GetProviders(ctx, mgr.GetAPIReader())
		if err != nil {
			setupLog.Error(err, "unable to read installed providers")
			os.Exit(1)
		}

		factory := controllers.NewFactory(ctrl.Log, mgr).
			WithClusterReconciler(providers).
			WithVSphereDatacenterReconciler().
			WithSnowMachineConfigReconciler()

		reconcilers, err := factory.Build(ctx)
		if err != nil {
			setupLog.Error(err, "unable to build reconcilers")
			os.Exit(1)
		}

		setupLog.Info("Setting up cluster controller")
		if err := (reconcilers.ClusterReconciler).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", anywherev1.ClusterKind)
			os.Exit(1)
		}

		setupLog.Info("Setting up vspheredatacenter controller")
		if err := (reconcilers.VSphereDatacenterReconciler).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", anywherev1.VSphereDatacenterKind)
			os.Exit(1)
		}

		setupLog.Info("Setting up snowmachineconfig controller")
		if err := (reconcilers.SnowMachineConfigReconciler).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", anywherev1.SnowMachineConfigKind)
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
	if err := (&anywherev1.CloudStackDatacenterConfig{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", WEBHOOK, anywherev1.CloudStackDatacenterKind)
		os.Exit(1)
	}
	if err := (&anywherev1.CloudStackMachineConfig{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", WEBHOOK, anywherev1.CloudStackMachineConfigKind)
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
	if err := (&anywherev1.FluxConfig{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", WEBHOOK, anywherev1.FluxConfigKind)
		os.Exit(1)
	}
	if err := (&anywherev1.SnowMachineConfig{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", WEBHOOK, anywherev1.SnowMachineConfigKind)
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
