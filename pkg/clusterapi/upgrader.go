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

// Upgrade checks whether upgrading the CAPI components is necessary and, if so, upgrades them the new versions.
func (u *Upgrader) Upgrade(ctx context.Context, managementCluster *types.Cluster, provider providers.Provider, currentManagementComponents, newManagementComponents *cluster.ManagementComponents, newSpec *cluster.Spec) (*types.ChangeDiff, error) {
	logger.V(1).Info("Checking for CAPI upgrades")
	if !newSpec.Cluster.IsSelfManaged() {
		logger.V(1).Info("Skipping CAPI upgrades, not a self-managed cluster")
		return nil, nil
	}

	capiChangeDiff := capiChangeDiff(currentManagementComponents, newManagementComponents, provider)
	if capiChangeDiff == nil {
		logger.V(1).Info("Nothing to upgrade for CAPI")
		return nil, nil
	}

	logger.V(1).Info("Starting CAPI upgrades")
	if err := u.capiClient.Upgrade(ctx, managementCluster, provider, newManagementComponents, newSpec, capiChangeDiff); err != nil {
		return nil, fmt.Errorf("failed upgrading ClusterAPI from EKS-A version %s to EKS-A version %s: %v", currentManagementComponents.Eksa.Version, newManagementComponents.Eksa.Version, err)
	}

	return capiChangeDiff.toChangeDiff(), nil
}

type CAPIChangeDiff struct {
	CertManager            *types.ComponentChangeDiff
	Core                   *types.ComponentChangeDiff
	ControlPlane           *types.ComponentChangeDiff
	BootstrapProviders     []types.ComponentChangeDiff
	InfrastructureProvider *types.ComponentChangeDiff
}

func (c *CAPIChangeDiff) toChangeDiff() *types.ChangeDiff {
	if c == nil {
		logger.V(1).Info("Nothing to upgrade for CAPI")
		return nil
	}
	r := make([]*types.ComponentChangeDiff, 0, 4+len(c.BootstrapProviders))
	r = append(r, c.CertManager, c.Core, c.ControlPlane, c.InfrastructureProvider)
	for _, bootstrapChangeDiff := range c.BootstrapProviders {
		b := bootstrapChangeDiff
		r = append(r, &b)
	}

	return types.NewChangeDiff(r...)
}

// ChangeDiff generates a version change diff for the CAPI components.
func ChangeDiff(currentManagementComponents, newManagementComponents *cluster.ManagementComponents, provider providers.Provider) *types.ChangeDiff {
	return capiChangeDiff(currentManagementComponents, newManagementComponents, provider).toChangeDiff()
}

func capiChangeDiff(currentManagementComponents, newManagementComponents *cluster.ManagementComponents, provider providers.Provider) *CAPIChangeDiff {
	changeDiff := &CAPIChangeDiff{}
	componentChanged := false

	if currentManagementComponents.CertManager.Version != newManagementComponents.CertManager.Version {
		changeDiff.CertManager = &types.ComponentChangeDiff{
			ComponentName: "cert-manager",
			NewVersion:    newManagementComponents.CertManager.Version,
			OldVersion:    currentManagementComponents.CertManager.Version,
		}
		logger.V(1).Info("Cert-manager change diff", "oldVersion", changeDiff.CertManager.OldVersion, "newVersion", changeDiff.CertManager.NewVersion)
		componentChanged = true
	}

	if currentManagementComponents.ClusterAPI.Version != newManagementComponents.ClusterAPI.Version {
		changeDiff.Core = &types.ComponentChangeDiff{
			ComponentName: "cluster-api",
			NewVersion:    newManagementComponents.ClusterAPI.Version,
			OldVersion:    currentManagementComponents.ClusterAPI.Version,
		}
		logger.V(1).Info("CAPI Core change diff", "oldVersion", changeDiff.Core.OldVersion, "newVersion", changeDiff.Core.NewVersion)
		componentChanged = true
	}

	if currentManagementComponents.ControlPlane.Version != newManagementComponents.ControlPlane.Version {
		changeDiff.ControlPlane = &types.ComponentChangeDiff{
			ComponentName: "kubeadm",
			NewVersion:    newManagementComponents.ControlPlane.Version,
			OldVersion:    currentManagementComponents.ControlPlane.Version,
		}
		logger.V(1).Info("CAPI Control Plane provider change diff", "oldVersion", changeDiff.ControlPlane.OldVersion, "newVersion", changeDiff.ControlPlane.NewVersion)
		componentChanged = true
	}

	if currentManagementComponents.Bootstrap.Version != newManagementComponents.Bootstrap.Version {
		componentChangeDiff := types.ComponentChangeDiff{
			ComponentName: "kubeadm",
			NewVersion:    newManagementComponents.Bootstrap.Version,
			OldVersion:    currentManagementComponents.Bootstrap.Version,
		}
		changeDiff.BootstrapProviders = append(changeDiff.BootstrapProviders, componentChangeDiff)
		logger.V(1).Info("CAPI Kubeadm Bootstrap Provider change diff", "oldVersion", componentChangeDiff.OldVersion, "newVersion", componentChangeDiff.NewVersion)
		componentChanged = true
	}

	if currentManagementComponents.ExternalEtcdBootstrap.Version != newManagementComponents.ExternalEtcdBootstrap.Version {
		componentChangeDiff := types.ComponentChangeDiff{
			ComponentName: "etcdadm-bootstrap",
			NewVersion:    newManagementComponents.ExternalEtcdBootstrap.Version,
			OldVersion:    currentManagementComponents.ExternalEtcdBootstrap.Version,
		}
		changeDiff.BootstrapProviders = append(changeDiff.BootstrapProviders, componentChangeDiff)
		logger.V(1).Info("CAPI Etcdadm Bootstrap Provider change diff", "oldVersion", componentChangeDiff.OldVersion, "newVersion", componentChangeDiff.NewVersion)
		componentChanged = true
	}

	if currentManagementComponents.ExternalEtcdController.Version != newManagementComponents.ExternalEtcdController.Version {
		componentChangeDiff := types.ComponentChangeDiff{
			ComponentName: "etcdadm-controller",
			NewVersion:    newManagementComponents.ExternalEtcdController.Version,
			OldVersion:    currentManagementComponents.ExternalEtcdController.Version,
		}
		changeDiff.BootstrapProviders = append(changeDiff.BootstrapProviders, componentChangeDiff)
		logger.V(1).Info("CAPI Etcdadm Controller Provider change diff", "oldVersion", componentChangeDiff.OldVersion, "newVersion", componentChangeDiff.NewVersion)
		componentChanged = true
	}

	if providerChangeDiff := provider.ChangeDiff(currentManagementComponents, newManagementComponents); providerChangeDiff != nil {
		changeDiff.InfrastructureProvider = providerChangeDiff
		logger.V(1).Info("CAPI Infrastrcture Provider change diff", "provider", providerChangeDiff.ComponentName, "oldVersion", providerChangeDiff.OldVersion, "newVersion", providerChangeDiff.NewVersion)
		componentChanged = true
	}

	if !componentChanged {
		return nil
	}

	return changeDiff
}
