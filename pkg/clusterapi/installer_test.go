package clusterapi_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/clusterapi/mocks"
	"github.com/aws/eks-anywhere/pkg/constants"
	providerMocks "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
)

type installerTest struct {
	*WithT
	ctx                  context.Context
	capiClient           *mocks.MockCAPIClient
	kubectlClient        *mocks.MockKubectlClient
	installer            *clusterapi.Installer
	managementComponents *cluster.ManagementComponents
	currentSpec          *cluster.Spec
	cluster              *types.Cluster
	provider             *providerMocks.MockProvider
}

func newInstallerTest(t *testing.T) installerTest {
	ctrl := gomock.NewController(t)
	capiClient := mocks.NewMockCAPIClient(ctrl)
	kubectlClient := mocks.NewMockKubectlClient(ctrl)

	currentSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Bundles.Spec.Number = 1
		s.Bundles.Spec.VersionsBundles[0].ExternalEtcdBootstrap.Version = "v0.1.0"
		s.Bundles.Spec.VersionsBundles[0].ExternalEtcdController.Version = "v0.1.0"
	})

	return installerTest{
		WithT:                NewWithT(t),
		ctx:                  context.Background(),
		capiClient:           capiClient,
		kubectlClient:        kubectlClient,
		installer:            clusterapi.NewInstaller(capiClient, kubectlClient),
		managementComponents: cluster.ManagementComponentsFromBundles(currentSpec.Bundles),
		currentSpec:          currentSpec,
		provider:             providerMocks.NewMockProvider(ctrl),
		cluster: &types.Cluster{
			Name:           "cluster-name",
			KubeconfigFile: "k.kubeconfig",
		},
	}
}

func TestEnsureEtcdProviderInstallStackedEtcd(t *testing.T) {
	tt := newInstallerTest(t)

	tt.kubectlClient.EXPECT().CheckProviderExists(tt.ctx, tt.cluster.KubeconfigFile, constants.EtcdAdmBootstrapProviderName, constants.EtcdAdmBootstrapProviderSystemNamespace).Return(false, nil)
	tt.kubectlClient.EXPECT().CheckProviderExists(tt.ctx, tt.cluster.KubeconfigFile, constants.EtcdadmControllerProviderName, constants.EtcdAdmControllerSystemNamespace).Return(false, nil)
	tt.capiClient.EXPECT().InstallEtcdadmProviders(tt.ctx, tt.managementComponents, tt.currentSpec, tt.cluster, tt.provider, []string{constants.EtcdAdmBootstrapProviderName, constants.EtcdadmControllerProviderName})

	tt.Expect(tt.installer.EnsureEtcdProvidersInstallation(tt.ctx, tt.cluster, tt.provider, tt.managementComponents, tt.currentSpec))
}

func TestEnsureEtcdProviderInstallExternalEtcd(t *testing.T) {
	tt := newInstallerTest(t)

	tt.kubectlClient.EXPECT().CheckProviderExists(tt.ctx, tt.cluster.KubeconfigFile, constants.EtcdAdmBootstrapProviderName, constants.EtcdAdmBootstrapProviderSystemNamespace).Return(true, nil)
	tt.kubectlClient.EXPECT().CheckProviderExists(tt.ctx, tt.cluster.KubeconfigFile, constants.EtcdadmControllerProviderName, constants.EtcdAdmControllerSystemNamespace).Return(true, nil)

	tt.Expect(tt.installer.EnsureEtcdProvidersInstallation(tt.ctx, tt.cluster, tt.provider, tt.managementComponents, tt.currentSpec))
}
