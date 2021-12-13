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
	*clients
}

func NewUpgrader(capiClient CAPIClient, kubectlClient KubectlClient) *Upgrader {
	return &Upgrader{
		clients: &clients{
			capiClient:    capiClient,
			kubectlClient: kubectlClient,
		},
	}
}

func (u *Upgrader) Upgrade(ctx context.Context, managementCluster *types.Cluster, provider providers.Provider, currentSpec, newSpec *cluster.Spec) (*types.ChangeDiff, error) {
	logger.V(1).Info("Checking for CAPI upgrades")
	if !newSpec.Cluster.IsSelfManaged() {
		logger.V(1).Info("Skipping CAPI upgrades, not a self-managed cluster")
		return nil, nil
	}

	capiChangeDiff := CapiChangeDiff(currentSpec, newSpec, provider)
	if capiChangeDiff == nil {
		logger.V(1).Info("Nothing to upgrade for CAPI")
		return nil, nil
	}

	logger.V(1).Info("Starting CAPI upgrades")
	if err := u.capiClient.Upgrade(ctx, managementCluster, provider, newSpec, capiChangeDiff); err != nil {
		return nil, fmt.Errorf("failed upgrading ClusterAPI from bundles %d to bundles %d: %v", currentSpec.Bundles.Spec.Number, newSpec.Bundles.Spec.Number, err)
	}

	return ToChangeDiff(capiChangeDiff), nil
}

type CAPIChangeDiff struct {
	CertManager            *types.ComponentChangeDiff
	Core                   *types.ComponentChangeDiff
	ControlPlane           *types.ComponentChangeDiff
	BootstrapProviders     []types.ComponentChangeDiff
	InfrastructureProvider *types.ComponentChangeDiff
}

func ToChangeDiff(capiChangeDiff *CAPIChangeDiff) *types.ChangeDiff {
	if capiChangeDiff == nil {
		logger.V(1).Info("Nothing to upgrade for CAPI")
		return nil
	}
	r := make([]*types.ComponentChangeDiff, 0, 4+len(capiChangeDiff.BootstrapProviders))
	r = append(r, capiChangeDiff.CertManager, capiChangeDiff.Core, capiChangeDiff.ControlPlane, capiChangeDiff.InfrastructureProvider)
	for _, bootstrapChangeDiff := range capiChangeDiff.BootstrapProviders {
		b := bootstrapChangeDiff
		r = append(r, &b)
	}

	return types.NewChangeDiff(r...)
}

func CapiChangeDiff(currentSpec, newSpec *cluster.Spec, provider providers.Provider) *CAPIChangeDiff {
	changeDiff := &CAPIChangeDiff{}
	componentChanged := false

	if currentSpec.VersionsBundle.CertManager.Version != newSpec.VersionsBundle.CertManager.Version {
		changeDiff.CertManager = &types.ComponentChangeDiff{
			ComponentName: "cert-manager",
			NewVersion:    newSpec.VersionsBundle.CertManager.Version,
			OldVersion:    currentSpec.VersionsBundle.CertManager.Version,
		}
		logger.V(1).Info("Cert-manager change diff", "oldVersion", changeDiff.CertManager.OldVersion, "newVersion", changeDiff.CertManager.NewVersion)
		componentChanged = true
	}

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
