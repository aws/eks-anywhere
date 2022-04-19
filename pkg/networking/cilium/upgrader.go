package cilium

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
)

type upgraderClient interface {
	Apply(ctx context.Context, cluster *types.Cluster, data []byte) error
	Delete(ctx context.Context, cluster *types.Cluster, data []byte) error
	WaitForPreflightDaemonSet(ctx context.Context, cluster *types.Cluster) error
	WaitForPreflightDeployment(ctx context.Context, cluster *types.Cluster) error
	WaitForCiliumDaemonSet(ctx context.Context, cluster *types.Cluster) error
	WaitForCiliumDeployment(ctx context.Context, cluster *types.Cluster) error
}

type Upgrader struct {
	templater *Templater
	client    upgraderClient
}

func NewUpgrader(client Client, helm Helm) *Upgrader {
	return &Upgrader{
		templater: NewTemplater(helm),
		client:    newRetrier(client),
	}
}

func (u *Upgrader) Upgrade(ctx context.Context, cluster *types.Cluster, currentSpec, newSpec *cluster.Spec, namespaces []string) (*types.ChangeDiff, error) {
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

	logger.V(3).Info("Generating Cilium upgrade manifest")
	upgradeManifest, err := u.templater.GenerateUpgradeManifest(ctx, currentSpec, newSpec)
	if err != nil {
		return nil, err
	}

	if chartValuesChanged {
		if newSpec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode == v1alpha1.CiliumPolicyModeAlways {
			logger.V(3).Info("Installing NetworkPolicy resources for policy enforcement mode 'always'")
			networkPolicyManifest, err := u.templater.GenerateNetworkPolicyManifest(newSpec, namespaces)
			if err != nil {
				return nil, err
			}
			upgradeManifest = templater.AppendYamlResources(upgradeManifest, networkPolicyManifest)
		}
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
	if currentSpec.VersionsBundle.Cilium.Version == newSpec.VersionsBundle.Cilium.Version {
		return nil
	}

	return &types.ChangeDiff{
		ComponentReports: []types.ComponentChangeDiff{
			{
				ComponentName: "cilium",
				OldVersion:    currentSpec.VersionsBundle.Cilium.Version,
				NewVersion:    newSpec.VersionsBundle.Cilium.Version,
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
	}
	// we can add comparisons for more values here as we start accepting them from cluster spec
	return false
}
