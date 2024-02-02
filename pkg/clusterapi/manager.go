package clusterapi

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
)

type Manager struct {
	*Installer
	*Upgrader
}

type clients struct {
	capiClient    CAPIClient
	kubectlClient KubectlClient
}

func NewManager(capiClient CAPIClient, kubectlClient KubectlClient) *Manager {
	return &Manager{
		Installer: NewInstaller(capiClient, kubectlClient),
		Upgrader:  NewUpgrader(capiClient, kubectlClient),
	}
}

type CAPIClient interface {
	Upgrade(ctx context.Context, managementCluster *types.Cluster, provider providers.Provider, managementComponents *cluster.ManagementComponents, newSpec *cluster.Spec, changeDiff *CAPIChangeDiff) error
	InstallEtcdadmProviders(ctx context.Context, managementComponents *cluster.ManagementComponents, clusterSpec *cluster.Spec, cluster *types.Cluster, provider providers.Provider, installProviders []string) error
}

type KubectlClient interface {
	CheckProviderExists(ctx context.Context, kubeconfigFile, name, namespace string) (bool, error)
}
