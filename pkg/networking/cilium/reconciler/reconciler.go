package reconciler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
func (r *Reconciler) Reconcile(ctx context.Context, logger logr.Logger, client client.Client, spec *cluster.Spec) (controller.Result, error) {
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

	// To accommodate upgrades of cluster cerated prior to introducing markers, we check for
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

		return controller.Result{}, nil
	}

	if !ciliumCfg.IsManaged() {
		logger.Info("Cilium configured as unmanaged, skipping upgrade")
		return controller.Result{}, nil
	}

	logger.Info("Cilium is already installed, checking if it needs upgrade")
	upgradeInfo := cilium.BuildUpgradePlan(installation, spec)

	if upgradeInfo.VersionUpgradeNeeded() {
		logger.Info("Cilium upgrade needed", "reason", upgradeInfo.Reason())

		if result, err := r.upgrade(ctx, logger, client, installation, spec); err != nil {
			return controller.Result{}, err
		} else if result.Return() {
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
	if err := serverside.ReconcileYaml(ctx, client, upgradeManifest); err != nil {
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
	preFlightCiliumDS, err := getPreflightDaemonSet(ctx, client)
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
