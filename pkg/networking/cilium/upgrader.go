package cilium

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/types"
)

// KubernetesClient is a client to interact with the Kubernetes API.
type KubernetesClient interface {
	Apply(ctx context.Context, cluster *types.Cluster, data []byte) error
	Delete(ctx context.Context, cluster *types.Cluster, data []byte) error
	WaitForPreflightDaemonSet(ctx context.Context, cluster *types.Cluster) error
	WaitForPreflightDeployment(ctx context.Context, cluster *types.Cluster) error
	WaitForCiliumDaemonSet(ctx context.Context, cluster *types.Cluster) error
	WaitForCiliumDeployment(ctx context.Context, cluster *types.Cluster) error
	RolloutRestartCiliumDaemonSet(ctx context.Context, cluster *types.Cluster) error
}

// UpgradeTemplater generates a Cilium manifests for upgrade.
type UpgradeTemplater interface {
	GenerateUpgradePreflightManifest(ctx context.Context, spec *cluster.Spec) ([]byte, error)
	GenerateManifest(ctx context.Context, spec *cluster.Spec, opts ...ManifestOpt) ([]byte, error)
}

// Upgrader allows to upgrade a Cilium installation in a EKS-A cluster.
type Upgrader struct {
	templater UpgradeTemplater
	client    KubernetesClient

	// skipUpgrade indicates Cilium upgrades should be skipped.
	skipUpgrade bool
}

// NewUpgrader constructs a new Upgrader.
func NewUpgrader(client KubernetesClient, templater UpgradeTemplater) *Upgrader {
	return &Upgrader{
		templater: templater,
		client:    client,
	}
}

// Upgrade configures a Cilium installation to match the desired state in the cluster Spec.
func (u *Upgrader) Upgrade(ctx context.Context, cluster *types.Cluster, currentSpec, newSpec *cluster.Spec, namespaces []string) (*types.ChangeDiff, error) {
	if u.skipUpgrade {
		logger.V(1).Info("Cilium upgrade skipped")
		return nil, nil
	}

	diff := ciliumChangeDiff(currentSpec, newSpec)
	chartValuesChanged := ciliumHelmChartValuesChanged(currentSpec, newSpec)
	if diff == nil && !chartValuesChanged {
		logger.V(1).Info("Nothing to upgrade for Cilium, skipping")
		return nil, nil
	}

	if diff != nil {
		logger.V(1).Info("Upgrading Cilium", "oldVersion", diff.ComponentReports[0].OldVersion, "newVersion", diff.ComponentReports[0].NewVersion)
	}
	logger.V(4).Info("Generating Cilium upgrade preflight manifest")
	preflight, err := u.templater.GenerateUpgradePreflightManifest(ctx, newSpec)
	if err != nil {
		return nil, err
	}

	logger.V(2).Info("Installing Cilium upgrade preflight manifest")
	if err := u.client.Apply(ctx, cluster, preflight); err != nil {
		return nil, fmt.Errorf("failed applying cilium preflight check: %v", err)
	}

	logger.V(3).Info("Waiting for Cilium upgrade preflight checks to be up")
	if err := u.waitForPreflight(ctx, cluster); err != nil {
		return nil, err
	}

	logger.V(3).Info("Deleting Cilium upgrade preflight")
	if err := u.client.Delete(ctx, cluster, preflight); err != nil {
		return nil, fmt.Errorf("failed deleting cilium preflight check: %v", err)
	}

	versionsBundle := currentSpec.RootVersionsBundle()

	logger.V(3).Info("Generating Cilium upgrade manifest")
	currentKubeVersion, err := getKubeVersionString(currentSpec, versionsBundle)
	if err != nil {
		return nil, err
	}

	previousCiliumVersion, err := semver.New(versionsBundle.Cilium.Version)
	if err != nil {
		return nil, err
	}

	upgradeManifest, err := u.templater.GenerateManifest(ctx, newSpec,
		WithKubeVersion(currentKubeVersion),
		WithUpgradeFromVersion(*previousCiliumVersion),
		WithPolicyAllowedNamespaces(namespaces),
	)
	if err != nil {
		return nil, err
	}

	logger.V(2).Info("Installing new Cilium version")
	if err := u.client.Apply(ctx, cluster, upgradeManifest); err != nil {
		return nil, fmt.Errorf("failed applying cilium upgrade: %v", err)
	}

	logger.V(3).Info("Waiting for upgraded Cilium to be ready")
	if err := u.waitForCilium(ctx, cluster); err != nil {
		return nil, err
	}

	return diff, nil
}

func (u *Upgrader) waitForPreflight(ctx context.Context, cluster *types.Cluster) error {
	if err := u.client.WaitForPreflightDaemonSet(ctx, cluster); err != nil {
		return err
	}

	if err := u.client.WaitForPreflightDeployment(ctx, cluster); err != nil {
		return err
	}

	return nil
}

func (u *Upgrader) waitForCilium(ctx context.Context, cluster *types.Cluster) error {
	if err := u.client.WaitForCiliumDaemonSet(ctx, cluster); err != nil {
		return err
	}

	if err := u.client.WaitForCiliumDeployment(ctx, cluster); err != nil {
		return err
	}

	return nil
}

func ciliumChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ChangeDiff {
	currentVersionsBundle := currentSpec.RootVersionsBundle()
	newVersionsBundle := newSpec.RootVersionsBundle()
	if currentVersionsBundle.Cilium.Version == newVersionsBundle.Cilium.Version {
		return nil
	}

	return &types.ChangeDiff{
		ComponentReports: []types.ComponentChangeDiff{
			{
				ComponentName: "cilium",
				OldVersion:    currentVersionsBundle.Cilium.Version,
				NewVersion:    newVersionsBundle.Cilium.Version,
			},
		},
	}
}

func ChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ChangeDiff {
	return ciliumChangeDiff(currentSpec, newSpec)
}

func ciliumHelmChartValuesChanged(currentSpec, newSpec *cluster.Spec) bool {
	if currentSpec.Cluster.Spec.ClusterNetwork.CNIConfig == nil || currentSpec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium == nil {
		// this is for clusters created using 0.7 and lower versions, they won't have these fields initialized
		// in these cases, a non-default PolicyEnforcementMode in the newSpec will be considered a change
		if newSpec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode != v1alpha1.CiliumPolicyModeDefault {
			return true
		}
	} else {
		if newSpec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode != currentSpec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode {
			return true
		}
		if newSpec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.EgressMasqueradeInterfaces != currentSpec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.EgressMasqueradeInterfaces {
			return true
		}
	}
	// we can add comparisons for more values here as we start accepting them from cluster spec

	return false
}

func (u *Upgrader) RunPostControlPlaneUpgradeSetup(ctx context.Context, cluster *types.Cluster) error {
	// we need to restart cilium pods after control plane vms get upgraded to prevent issue seen in https://github.com/aws/eks-anywhere/issues/1888
	if err := u.client.RolloutRestartCiliumDaemonSet(ctx, cluster); err != nil {
		return fmt.Errorf("restarting cilium daemonset: %v", err)
	}
	return nil
}

// SetSkipUpgrade configures u to skip the upgrade process.
func (u *Upgrader) SetSkipUpgrade(v bool) {
	u.skipUpgrade = v
}
