package clusterapi

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
)

type Installer struct {
	*clients
}

func NewInstaller(capiClient CAPIClient, kubectlClient KubectlClient) *Installer {
	return &Installer{
		clients: &clients{
			capiClient:    capiClient,
			kubectlClient: kubectlClient,
		},
	}
}

// EnsureEtcdProvidersInstallation ensures that the CAPI etcd providers are installed in the management cluster.
func (i *Installer) EnsureEtcdProvidersInstallation(ctx context.Context, managementCluster *types.Cluster, provider providers.Provider, managementComponents *cluster.ManagementComponents, currSpec *cluster.Spec) error {
	if !currSpec.Cluster.IsSelfManaged() {
		logger.V(1).Info("Not a management cluster, skipping check for CAPI etcd providers")
		return nil
	}

	var installProviders []string
	etcdBootstrapExists, err := i.kubectlClient.CheckProviderExists(ctx, managementCluster.KubeconfigFile, constants.EtcdAdmBootstrapProviderName, constants.EtcdAdmBootstrapProviderSystemNamespace)
	if err != nil {
		return err
	}
	if !etcdBootstrapExists {
		installProviders = append(installProviders, constants.EtcdAdmBootstrapProviderName)
	}
	etcdControllerExists, err := i.kubectlClient.CheckProviderExists(ctx, managementCluster.KubeconfigFile, constants.EtcdadmControllerProviderName, constants.EtcdAdmControllerSystemNamespace)
	if err != nil {
		return err
	}
	if !etcdControllerExists {
		installProviders = append(installProviders, constants.EtcdadmControllerProviderName)
	}

	if len(installProviders) > 0 {
		return i.capiClient.InstallEtcdadmProviders(ctx, managementComponents, currSpec, managementCluster, provider, installProviders)
	}
	return nil
}
