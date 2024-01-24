package clusterapi

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
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
func (u *Upgrader) Upgrade(ctx context.Context, managementCluster *types.Cluster, provider providers.Provider, currentManagementComponentsVersionsBundle *releasev1.VersionsBundle, newSpec *cluster.Spec) (*types.ChangeDiff, error) {
	logger.V(1).Info("Checking for CAPI upgrades")
	if !newSpec.Cluster.IsSelfManaged() {
		logger.V(1).Info("Skipping CAPI upgrades, not a self-managed cluster")
		return nil, nil
	}

	newVersionsBundle := newSpec.FirstVersionsBundle()
	capiChangeDiff := capiChangeDiff(currentManagementComponentsVersionsBundle, newVersionsBundle, provider)
	if capiChangeDiff == nil {
		logger.V(1).Info("Nothing to upgrade for CAPI")
		return nil, nil
	}

	logger.V(1).Info("Starting CAPI upgrades")
	if err := u.capiClient.Upgrade(ctx, managementCluster, provider, newSpec, capiChangeDiff); err != nil {
		return nil, fmt.Errorf("failed upgrading ClusterAPI to EKS-A version %s: %v", newVersionsBundle.Eksa.Version, err)
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

// CapiChangeDiff generates a version change diff for the CAPI components.
func CapiChangeDiff(currentVersionsBundle, newVersionsBundle *releasev1.VersionsBundle, provider providers.Provider) *types.ChangeDiff {
	return capiChangeDiff(currentVersionsBundle, newVersionsBundle, provider).toChangeDiff()
}

func capiChangeDiff(currentVersionsBundle, newVersionsBundle *releasev1.VersionsBundle, provider providers.Provider) *CAPIChangeDiff {
	changeDiff := &CAPIChangeDiff{}
	componentChanged := false

	if currentVersionsBundle.CertManager.Version != newVersionsBundle.CertManager.Version {
		changeDiff.CertManager = &types.ComponentChangeDiff{
			ComponentName: "cert-manager",
			NewVersion:    newVersionsBundle.CertManager.Version,
			OldVersion:    currentVersionsBundle.CertManager.Version,
		}
		logger.V(1).Info("Cert-manager change diff", "oldVersion", changeDiff.CertManager.OldVersion, "newVersion", changeDiff.CertManager.NewVersion)
		componentChanged = true
	}

	if currentVersionsBundle.ClusterAPI.Version != newVersionsBundle.ClusterAPI.Version {
		changeDiff.Core = &types.ComponentChangeDiff{
			ComponentName: "cluster-api",
			NewVersion:    newVersionsBundle.ClusterAPI.Version,
			OldVersion:    currentVersionsBundle.ClusterAPI.Version,
		}
		logger.V(1).Info("CAPI Core change diff", "oldVersion", changeDiff.Core.OldVersion, "newVersion", changeDiff.Core.NewVersion)
		componentChanged = true
	}

	if currentVersionsBundle.ControlPlane.Version != newVersionsBundle.ControlPlane.Version {
		changeDiff.ControlPlane = &types.ComponentChangeDiff{
			ComponentName: "kubeadm",
			NewVersion:    newVersionsBundle.ControlPlane.Version,
			OldVersion:    currentVersionsBundle.ControlPlane.Version,
		}
		logger.V(1).Info("CAPI Control Plane provider change diff", "oldVersion", changeDiff.ControlPlane.OldVersion, "newVersion", changeDiff.ControlPlane.NewVersion)
		componentChanged = true
	}

	if currentVersionsBundle.Bootstrap.Version != newVersionsBundle.Bootstrap.Version {
		componentChangeDiff := types.ComponentChangeDiff{
			ComponentName: "kubeadm",
			NewVersion:    newVersionsBundle.Bootstrap.Version,
			OldVersion:    currentVersionsBundle.Bootstrap.Version,
		}
		changeDiff.BootstrapProviders = append(changeDiff.BootstrapProviders, componentChangeDiff)
		logger.V(1).Info("CAPI Kubeadm Bootstrap Provider change diff", "oldVersion", componentChangeDiff.OldVersion, "newVersion", componentChangeDiff.NewVersion)
		componentChanged = true
	}

	if currentVersionsBundle.ExternalEtcdBootstrap.Version != newVersionsBundle.ExternalEtcdBootstrap.Version {
		componentChangeDiff := types.ComponentChangeDiff{
			ComponentName: "etcdadm-bootstrap",
			NewVersion:    newVersionsBundle.ExternalEtcdBootstrap.Version,
			OldVersion:    currentVersionsBundle.ExternalEtcdBootstrap.Version,
		}
		changeDiff.BootstrapProviders = append(changeDiff.BootstrapProviders, componentChangeDiff)
		logger.V(1).Info("CAPI Etcdadm Bootstrap Provider change diff", "oldVersion", componentChangeDiff.OldVersion, "newVersion", componentChangeDiff.NewVersion)
		componentChanged = true
	}

	if currentVersionsBundle.ExternalEtcdController.Version != newVersionsBundle.ExternalEtcdController.Version {
		componentChangeDiff := types.ComponentChangeDiff{
			ComponentName: "etcdadm-controller",
			NewVersion:    newVersionsBundle.ExternalEtcdController.Version,
			OldVersion:    currentVersionsBundle.ExternalEtcdController.Version,
		}
		changeDiff.BootstrapProviders = append(changeDiff.BootstrapProviders, componentChangeDiff)
		logger.V(1).Info("CAPI Etcdadm Controller Provider change diff", "oldVersion", componentChangeDiff.OldVersion, "newVersion", componentChangeDiff.NewVersion)
		componentChanged = true
	}

	if providerChangeDiff := provider.ChangeDiff(currentVersionsBundle, newVersionsBundle); providerChangeDiff != nil {
		changeDiff.InfrastructureProvider = providerChangeDiff
		logger.V(1).Info("CAPI Infrastrcture Provider change diff", "provider", providerChangeDiff.ComponentName, "oldVersion", providerChangeDiff.OldVersion, "newVersion", providerChangeDiff.NewVersion)
		componentChanged = true
	}

	if !componentChanged {
		return nil
	}

	return changeDiff
}
