package clusterapi

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
)

type Upgrader struct {
	capiClient CAPIClient
}

type CAPIClient interface {
	Upgrade(ctx context.Context, managementCluster *types.Cluster, provider providers.Provider, newSpec *cluster.Spec, changeDiff *CAPIChangeDiff) error
}

func NewUpgrader(client CAPIClient) *Upgrader {
	return &Upgrader{
		capiClient: client,
	}
}

func (u *Upgrader) Upgrade(ctx context.Context, managementCluster *types.Cluster, provider providers.Provider, currentSpec, newSpec *cluster.Spec) error {
	logger.V(1).Info("Checking for CAPI upgrades")
	if !newSpec.Cluster.IsSelfManaged() {
		logger.V(1).Info("Skipping CAPI upgrades, not a self-managed cluster")
		return nil
	}

	changeDiff := u.capiChangeDiff(currentSpec, newSpec, provider)
	if changeDiff == nil {
		logger.V(1).Info("Nothing to upgrade for CAPI")
		return nil
	}

	logger.V(1).Info("Starting CAPI upgrades")
	if err := u.capiClient.Upgrade(ctx, managementCluster, provider, newSpec, changeDiff); err != nil {
		return fmt.Errorf("failed upgrading ClusterAPI from bundles %d to bundles %d: %v", currentSpec.Bundles.Spec.Number, newSpec.Bundles.Spec.Number, err)
	}

	return nil
}

type CAPIChangeDiff struct {
	Core                   *types.ComponentChangeDiff
	ControlPlane           *types.ComponentChangeDiff
	BootstrapProviders     []types.ComponentChangeDiff
	InfrastructureProvider *types.ComponentChangeDiff
}

func (u *Upgrader) capiChangeDiff(currentSpec, newSpec *cluster.Spec, provider providers.Provider) *CAPIChangeDiff {
	changeDiff := &CAPIChangeDiff{}
	componentChanged := false

	if currentSpec.VersionsBundle.ClusterAPI.Version != newSpec.VersionsBundle.ClusterAPI.Version {
		changeDiff.Core = &types.ComponentChangeDiff{
			ComponentName: "cluster-api",
			NewVersion:    newSpec.VersionsBundle.ClusterAPI.Version,
			OldVersion:    currentSpec.VersionsBundle.ClusterAPI.Version,
		}
		logger.V(1).Info("CAPI Core change diff", "oldVersion", changeDiff.Core.OldVersion, "newVersion", changeDiff.Core.NewVersion)
		componentChanged = true
	}

	if currentSpec.VersionsBundle.ControlPlane.Version != newSpec.VersionsBundle.ControlPlane.Version {
		changeDiff.ControlPlane = &types.ComponentChangeDiff{
			ComponentName: "kubeadm",
			NewVersion:    newSpec.VersionsBundle.ControlPlane.Version,
			OldVersion:    currentSpec.VersionsBundle.ControlPlane.Version,
		}
		logger.V(1).Info("CAPI Control Plane provider change diff", "oldVersion", changeDiff.ControlPlane.OldVersion, "newVersion", changeDiff.ControlPlane.NewVersion)
		componentChanged = true
	}

	if currentSpec.VersionsBundle.Bootstrap.Version != newSpec.VersionsBundle.Bootstrap.Version {
		componentChangeDiff := types.ComponentChangeDiff{
			ComponentName: "kubeadm",
			NewVersion:    newSpec.VersionsBundle.Bootstrap.Version,
			OldVersion:    currentSpec.VersionsBundle.Bootstrap.Version,
		}
		changeDiff.BootstrapProviders = append(changeDiff.BootstrapProviders, componentChangeDiff)
		logger.V(1).Info("CAPI Kubeadm Bootstrap Provider change diff", "oldVersion", componentChangeDiff.OldVersion, "newVersion", componentChangeDiff.NewVersion)
		componentChanged = true
	}

	if newSpec.Spec.ExternalEtcdConfiguration != nil {
		if currentSpec.VersionsBundle.ExternalEtcdBootstrap.Version != newSpec.VersionsBundle.ExternalEtcdBootstrap.Version {
			componentChangeDiff := types.ComponentChangeDiff{
				ComponentName: "etcdadm-bootstrap",
				NewVersion:    newSpec.VersionsBundle.ExternalEtcdBootstrap.Version,
				OldVersion:    currentSpec.VersionsBundle.ExternalEtcdBootstrap.Version,
			}
			changeDiff.BootstrapProviders = append(changeDiff.BootstrapProviders, componentChangeDiff)
			logger.V(1).Info("CAPI Etcdadm Bootstrap Provider change diff", "oldVersion", componentChangeDiff.OldVersion, "newVersion", componentChangeDiff.NewVersion)
			componentChanged = true
		}

		if currentSpec.VersionsBundle.ExternalEtcdController.Version != newSpec.VersionsBundle.ExternalEtcdController.Version {
			componentChangeDiff := types.ComponentChangeDiff{
				ComponentName: "etcdadm-controller",
				NewVersion:    newSpec.VersionsBundle.ExternalEtcdController.Version,
				OldVersion:    currentSpec.VersionsBundle.ExternalEtcdController.Version,
			}
			changeDiff.BootstrapProviders = append(changeDiff.BootstrapProviders, componentChangeDiff)
			logger.V(1).Info("CAPI Etcdadm Controller Provider change diff", "oldVersion", componentChangeDiff.OldVersion, "newVersion", componentChangeDiff.NewVersion)
			componentChanged = true
		}
	}

	if providerChangeDiff := provider.ChangeDiff(currentSpec, newSpec); providerChangeDiff != nil {
		changeDiff.InfrastructureProvider = providerChangeDiff
		logger.V(1).Info("CAPI Infrastrcture Provider change diff", "provider", providerChangeDiff.ComponentName, "oldVersion", providerChangeDiff.OldVersion, "newVersion", providerChangeDiff.NewVersion)
		componentChanged = true
	}

	if !componentChanged {
		return nil
	}

	return changeDiff
}
