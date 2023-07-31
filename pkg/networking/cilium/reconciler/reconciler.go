package reconciler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networking/cilium"
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/utils/oci"
)

const (
	defaultRequeueTime = time.Second * 10
)

var (
	serviceKind    = corev1.SchemeGroupVersion.WithKind("Service").GroupKind()
	daemonSetKind  = appsv1.SchemeGroupVersion.WithKind("DaemonSet").GroupKind()
	deploymentKind = appsv1.SchemeGroupVersion.WithKind("Deployment").GroupKind()
)

type Templater interface {
	GenerateUpgradePreflightManifest(ctx context.Context, spec *cluster.Spec) ([]byte, error)
	GenerateManifest(ctx context.Context, spec *cluster.Spec, opts ...cilium.ManifestOpt) ([]byte, error)
}

// Reconciler allows to reconcile a Cilium CNI.
type Reconciler struct {
	templater Templater
}

func New(templater Templater) *Reconciler {
	return &Reconciler{
		templater: templater,
	}
}

// Reconcile takes the Cilium CNI in a cluster to the desired state defined in a cluster Spec.
// It uses a controller.Result to indicate when requeues are needed. client is connected to the
// target Kubernetes cluster, not the management cluster.
// nolint:gocyclo
// TODO: reduce cyclomatic complexity - https://github.com/aws/eks-anywhere-internal/issues/1461
func (r *Reconciler) Reconcile(ctx context.Context, logger logr.Logger, client client.Client, spec *cluster.Spec) (res controller.Result, reterr error) {
	installation, err := cilium.GetInstallation(ctx, client)
	if err != nil {
		return controller.Result{}, err
	}

	// We use a marker to detect if EKS-A Cilium has ever been installed. If it has never been
	// installed and isn't currently installed we always attempt to install it regardless of whether
	// the user is skipping EKS-A Cilium management. This satsifies criteria for successful cluster
	// creation.
	//
	// If EKS-A Cilium was previously installed, as denoted by the marker, we only want to
	// manage it if its still installed and the user still wants us to manage the installation (as
	// denoted by the API skip flag).
	//
	// In the event a user uninstalls EKS-A Cilium, updates the cluster spec to skip EKS-A Cilium
	// management, then tries to upgrade, we will attempt to install EKS-A Cilium. This is because
	// reconciliation has no operational context (create vs upgrade) and can only observe that no
	// installation is present and there is no marker indicating it was ever present which is
	// equivilent to a typical create scenario where we must install a CNI to satisfy cluster
	// create success criteria.

	// To accommodate upgrades of cluster created prior to introducing markers, we check for
	// an existing installation and try to mark the cluster as having already had EKS-A
	// Cilium installed.
	if !ciliumWasInstalled(ctx, spec.Cluster) && installation.Installed() {
		logger.Info(fmt.Sprintf(
			"Cilium installed but missing %v annotation; applying annotation",
			EKSACiliumInstalledAnnotation,
		))
		markCiliumInstalled(ctx, spec.Cluster)
	}

	ciliumCfg := spec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium

	if !installation.Installed() &&
		(ciliumCfg.IsManaged() || !ciliumWasInstalled(ctx, spec.Cluster)) {
		if err := r.install(ctx, logger, client, spec); err != nil {
			return controller.Result{}, err
		}

		logger.Info(fmt.Sprintf(
			"Applying %v annotation to Cluster object",
			EKSACiliumInstalledAnnotation,
		))

		markCiliumInstalled(ctx, spec.Cluster)
		conditions.MarkTrue(spec.Cluster, anywherev1.DefaultCNIConfiguredCondition)
		return controller.Result{}, nil
	}

	if !ciliumCfg.IsManaged() {
		logger.Info("Cilium configured as unmanaged, skipping upgrade")
		conditions.MarkFalse(spec.Cluster, anywherev1.DefaultCNIConfiguredCondition, anywherev1.SkipUpgradesForDefaultCNIConfiguredReason, clusterv1.ConditionSeverityWarning, "Configured to skip default Cilium CNI upgrades")
		return controller.Result{}, nil
	}

	logger.Info("Cilium is already installed, checking if it needs upgrade")
	upgradeInfo := cilium.BuildUpgradePlan(installation, spec)

	if upgradeInfo.VersionUpgradeNeeded() {
		logger.Info("Cilium upgrade needed", "reason", upgradeInfo.Reason())
		if result, err := r.upgrade(ctx, logger, client, installation, spec); err != nil {
			return controller.Result{}, err
		} else if result.Return() {
			conditions.MarkFalse(spec.Cluster, anywherev1.DefaultCNIConfiguredCondition, anywherev1.DefaultCNIUpgradeInProgressReason, clusterv1.ConditionSeverityInfo, "Cilium version upgrade needed")
			return result, nil
		}

	} else if upgradeInfo.ConfigUpdateNeeded() {
		logger.Info("Cilium config update needed", "reason", upgradeInfo.Reason())
		if err := r.updateConfig(ctx, client, spec); err != nil {
			return controller.Result{}, err
		}
	} else {
		logger.Info("Cilium is already up to date")
	}

	// Upgrade process has run its course, and so we can now mark that the default cni has been configured.
	conditions.MarkTrue(spec.Cluster, anywherev1.DefaultCNIConfiguredCondition)

	return r.deletePreflightIfExists(ctx, client, spec)
}

func (r *Reconciler) install(ctx context.Context, log logr.Logger, client client.Client, spec *cluster.Spec) error {
	log.Info("Installing Cilium")
	if err := r.applyFullManifest(ctx, client, spec); err != nil {
		return errors.Wrap(err, "installing Cilium")
	}

	return nil
}

func (r *Reconciler) upgrade(ctx context.Context, logger logr.Logger, client client.Client, installation *cilium.Installation, spec *cluster.Spec) (controller.Result, error) {
	if err := cilium.CheckDaemonSetReady(installation.DaemonSet); err != nil {
		logger.Info("Cilium DS is not ready, requeueing", "reason", err.Error())
		return controller.Result{Result: &ctrl.Result{
			RequeueAfter: defaultRequeueTime,
		}}, nil
	}

	preflightInstallation, err := getPreflightInstallation(ctx, client)
	if err != nil {
		return controller.Result{}, err
	}

	if !preflightInstallation.installed() {
		logger.Info("Installing Cilium upgrade preflight manifest")
		if err = r.installPreflight(ctx, client, spec); err != nil {
			return controller.Result{}, err
		}

		preflightInstallation, err = getPreflightInstallation(ctx, client)
		if err != nil {
			return controller.Result{}, err
		}

		if !preflightInstallation.installed() {
			logger.Info("Cilium preflight is not available yet, requeueing")
			return controller.Result{Result: &ctrl.Result{
				RequeueAfter: defaultRequeueTime,
			}}, nil
		}
	}

	if err = cilium.CheckPreflightDaemonSetReady(installation.DaemonSet, preflightInstallation.daemonSet); err != nil {
		logger.Info("Cilium preflight daemon set is not ready, requeueing", "reason", err.Error())
		return controller.Result{Result: &ctrl.Result{
			RequeueAfter: defaultRequeueTime,
		}}, nil
	}

	if err = cilium.CheckDeploymentReady(preflightInstallation.deployment); err != nil {
		logger.Info("Cilium preflight deployment is not ready, requeueing", "reason", err.Error())
		return controller.Result{Result: &ctrl.Result{
			RequeueAfter: defaultRequeueTime,
		}}, nil
	}

	logger.Info("Generating Cilium upgrade manifest")
	dsImage := installation.DaemonSet.Spec.Template.Spec.Containers[0].Image
	_, dsImageTag := oci.Split(dsImage)
	previousCiliumVersion, err := semver.New(dsImageTag)
	if err != nil {
		return controller.Result{}, errors.Wrapf(err, "installed cilium DS has an invalid version tag: %s", dsImage)
	}

	upgradeManifest, err := r.templater.GenerateManifest(ctx, spec,
		cilium.WithUpgradeFromVersion(*previousCiliumVersion),
	)
	if err != nil {
		return controller.Result{}, err
	}

	logger.Info("Applying Cilium upgrade manifest")

	// When upgrading from Cilium v1.11.x --> Cilium v1.12.x, the port number for a few components changed but the port names remained the same.
	// This caused "duplicate value" errors when upgrading to the new Cilium version using server-side apply because the merge key for these fields is the port number, not the name.
	// To alleviate this issue, we will use the client.Update strategy to update the yaml to be compatible with the new version of Cilium.
	// We are only doing this for the Cilium Service, DaemonSet, and Deployment since those are the only objects affected.
	// The rest of the objects in the Cilium upgrade manifest will continue to be applied using server-side apply.
	manifestObjs, err := r.reconcileSpecialCases(ctx, client, upgradeManifest)
	if err != nil {
		return controller.Result{}, err
	}

	if err := serverside.ReconcileObjects(ctx, client, manifestObjs); err != nil {
		return controller.Result{}, err
	}

	return controller.Result{}, nil
}

func (r *Reconciler) updateConfig(ctx context.Context, client client.Client, spec *cluster.Spec) error {
	if err := r.applyFullManifest(ctx, client, spec); err != nil {
		return errors.Wrap(err, "updating cilium config")
	}

	return nil
}

func (r *Reconciler) applyFullManifest(ctx context.Context, client client.Client, spec *cluster.Spec) error {
	upgradeManifest, err := r.templater.GenerateManifest(ctx, spec)
	if err != nil {
		return err
	}

	return serverside.ReconcileYaml(ctx, client, upgradeManifest)
}

func (r *Reconciler) deletePreflightIfExists(ctx context.Context, client client.Client, spec *cluster.Spec) (controller.Result, error) {
	preFlightCiliumDS, err := getDaemonSet(ctx, client, cilium.PreflightDaemonSetName)
	if err != nil {
		return controller.Result{}, err
	}

	if preFlightCiliumDS != nil {
		preflight, err := r.templater.GenerateUpgradePreflightManifest(ctx, spec)
		if err != nil {
			return controller.Result{}, err
		}

		logger.Info("Deleting Preflight Cilium objects")
		if err := clientutil.DeleteYaml(ctx, client, preflight); err != nil {
			return controller.Result{}, err
		}
	}

	return controller.Result{}, nil
}

func (r *Reconciler) installPreflight(ctx context.Context, client client.Client, spec *cluster.Spec) error {
	preflight, err := r.templater.GenerateUpgradePreflightManifest(ctx, spec)
	if err != nil {
		return err
	}

	if err = serverside.ReconcileYaml(ctx, client, preflight); err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) reconcileSpecialCases(ctx context.Context, c client.Client, yaml []byte) ([]client.Object, error) {
	objs, err := clientutil.YamlToClientObjects(yaml)
	if err != nil {
		return nil, err
	}

	index := 0
	for _, o := range objs {
		if (o.GetObjectKind().GroupVersionKind().GroupKind() == serviceKind && o.GetName() == cilium.ServiceName) ||
			(o.GetObjectKind().GroupVersionKind().GroupKind() == daemonSetKind && o.GetName() == cilium.DaemonSetName) ||
			(o.GetObjectKind().GroupVersionKind().GroupKind() == deploymentKind && o.GetName() == cilium.DeploymentName) {
			if err := serverside.UpdateObject(ctx, c, o); err != nil {
				return nil, err
			}
		} else {
			objs[index] = o
			index++
		}
	}
	return objs[:index], nil
}
