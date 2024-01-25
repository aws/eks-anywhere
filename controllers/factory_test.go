package controllers_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"

	"github.com/aws/eks-anywhere/controllers"
	"github.com/aws/eks-anywhere/controllers/mocks"
)

func TestFactoryBuildAllVSphereReconciler(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	logger := nullLog()
	ctrl := gomock.NewController(t)
	manager := mocks.NewMockManager(ctrl)
	manager.EXPECT().GetClient().AnyTimes()
	manager.EXPECT().GetScheme().AnyTimes()

	f := controllers.NewFactory(logger, manager).
		WithVSphereDatacenterReconciler()

	// testing idempotence
	f.WithVSphereDatacenterReconciler()

	reconcilers, err := f.Build(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(reconcilers.VSphereDatacenterReconciler).NotTo(BeNil())
}

func TestFactoryBuildAllDockerReconciler(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	logger := nullLog()
	ctrl := gomock.NewController(t)
	manager := mocks.NewMockManager(ctrl)
	manager.EXPECT().GetClient().AnyTimes()
	manager.EXPECT().GetScheme().AnyTimes()

	f := controllers.NewFactory(logger, manager).
		WithDockerDatacenterReconciler()

	// testing idempotence
	f.WithDockerDatacenterReconciler()

	reconcilers, err := f.Build(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(reconcilers.DockerDatacenterReconciler).NotTo(BeNil())
}

func TestFactoryBuildAllTinkerbellReconciler(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	logger := nullLog()
	ctrl := gomock.NewController(t)
	manager := mocks.NewMockManager(ctrl)
	manager.EXPECT().GetClient().AnyTimes()
	manager.EXPECT().GetScheme().AnyTimes()

	f := controllers.NewFactory(logger, manager).
		WithTinkerbellDatacenterReconciler()

	// testing idempotence
	f.WithTinkerbellDatacenterReconciler()

	reconcilers, err := f.Build(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(reconcilers.TinkerbellDatacenterReconciler).NotTo(BeNil())
}

func TestFactoryBuildAllCloudStackReconciler(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	logger := nullLog()
	ctrl := gomock.NewController(t)
	manager := mocks.NewMockManager(ctrl)
	manager.EXPECT().GetClient().AnyTimes()
	manager.EXPECT().GetScheme().AnyTimes()

	f := controllers.NewFactory(logger, manager).
		WithCloudStackDatacenterReconciler()

	// testing idempotence
	f.WithCloudStackDatacenterReconciler()

	reconcilers, err := f.Build(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(reconcilers.CloudStackDatacenterReconciler).NotTo(BeNil())
}

func TestFactoryBuildAllNutanixReconciler(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	logger := nullLog()
	ctrl := gomock.NewController(t)
	manager := mocks.NewMockManager(ctrl)
	manager.EXPECT().GetClient().AnyTimes()
	manager.EXPECT().GetScheme().AnyTimes()

	f := controllers.NewFactory(logger, manager).
		WithNutanixDatacenterReconciler().
		WithClusterReconciler([]clusterctlv1.Provider{
			{
				Type:         string(clusterctlv1.InfrastructureProviderType),
				ProviderName: "nutanix",
			},
		})

	// testing idempotence
	f.WithNutanixDatacenterReconciler().
		WithClusterReconciler([]clusterctlv1.Provider{
			{
				Type:         string(clusterctlv1.InfrastructureProviderType),
				ProviderName: "nutanix",
			},
		})

	reconcilers, err := f.Build(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(reconcilers.NutanixDatacenterReconciler).NotTo(BeNil())
}

func TestFactoryBuildClusterReconciler(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	logger := nullLog()
	ctrl := gomock.NewController(t)
	manager := mocks.NewMockManager(ctrl)
	manager.EXPECT().GetClient().AnyTimes()
	manager.EXPECT().GetScheme().AnyTimes()

	providers := []clusterctlv1.Provider{
		{
			Type:         string(clusterctlv1.ControlPlaneProviderType),
			ProviderName: "kubeadm",
		},
		{
			Type:         string(clusterctlv1.InfrastructureProviderType),
			ProviderName: "docker",
		},
		{
			Type:         string(clusterctlv1.InfrastructureProviderType),
			ProviderName: "vsphere",
		},
		{
			Type:         string(clusterctlv1.InfrastructureProviderType),
			ProviderName: "snow",
		},
		{
			Type:         string(clusterctlv1.InfrastructureProviderType),
			ProviderName: "tinkerbell",
		},
		{
			Type:         string(clusterctlv1.InfrastructureProviderType),
			ProviderName: "cloudstack",
		},
		{
			Type:         string(clusterctlv1.InfrastructureProviderType),
			ProviderName: "unknown-provider",
		},
	}

	f := controllers.NewFactory(logger, manager).
		WithClusterReconciler(providers)

	// testing idempotence
	f.WithClusterReconciler(providers)

	reconcilers, err := f.Build(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(reconcilers.ClusterReconciler).NotTo(BeNil())
}

func TestFactoryBuildAllSnowReconciler(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	logger := nullLog()
	ctrl := gomock.NewController(t)
	manager := mocks.NewMockManager(ctrl)
	manager.EXPECT().GetClient().AnyTimes()
	manager.EXPECT().GetScheme().AnyTimes()

	f := controllers.NewFactory(logger, manager).
		WithSnowMachineConfigReconciler()

	// testing idempotence
	f.WithSnowMachineConfigReconciler()

	reconcilers, err := f.Build(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(reconcilers.SnowMachineConfigReconciler).NotTo(BeNil())
}

func TestFactoryClose(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	logger := nullLog()
	ctrl := gomock.NewController(t)
	manager := mocks.NewMockManager(ctrl)
	manager.EXPECT().GetClient().AnyTimes()
	manager.EXPECT().GetScheme().AnyTimes()

	f := controllers.NewFactory(logger, manager)
	_, err := f.Build(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(f.Close(ctx)).To(Succeed())
}

func TestFactoryWithNutanixDatacenterReconciler(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	logger := nullLog()
	ctrl := gomock.NewController(t)
	manager := mocks.NewMockManager(ctrl)
	manager.EXPECT().GetClient().AnyTimes()
	manager.EXPECT().GetScheme().AnyTimes()

	f := controllers.NewFactory(logger, manager).
		WithNutanixDatacenterReconciler()

	// testing idempotence
	f.WithNutanixDatacenterReconciler()

	reconcilers, err := f.Build(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(reconcilers.NutanixDatacenterReconciler).NotTo(BeNil())
}

func TestFactoryWithKubeadmControlPlaneReconciler(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	logger := nullLog()
	ctrl := gomock.NewController(t)
	manager := mocks.NewMockManager(ctrl)
	manager.EXPECT().GetClient().AnyTimes()
	manager.EXPECT().GetScheme().AnyTimes()

	f := controllers.NewFactory(logger, manager).
		WithKubeadmControlPlaneReconciler()

	// testing idempotence
	f.WithKubeadmControlPlaneReconciler()

	reconcilers, err := f.Build(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(reconcilers.KubeadmControlPlaneReconciler).NotTo(BeNil())
}

func TestFactoryWithMachineDeploymentReconciler(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	logger := nullLog()
	ctrl := gomock.NewController(t)
	manager := mocks.NewMockManager(ctrl)
	manager.EXPECT().GetClient().AnyTimes()
	manager.EXPECT().GetScheme().AnyTimes()

	f := controllers.NewFactory(logger, manager).
		WithMachineDeploymentReconciler()

	// testing idempotence
	f.WithMachineDeploymentReconciler()

	reconcilers, err := f.Build(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(reconcilers.MachineDeploymentReconciler).NotTo(BeNil())
}

func TestFactoryWithNodeUpgradeReconciler(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	logger := nullLog()
	ctrl := gomock.NewController(t)
	manager := mocks.NewMockManager(ctrl)
	manager.EXPECT().GetClient().AnyTimes()
	manager.EXPECT().GetScheme().AnyTimes()

	f := controllers.NewFactory(logger, manager).
		WithNodeUpgradeReconciler()

	// testing idempotence
	f.WithNodeUpgradeReconciler()

	reconcilers, err := f.Build(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(reconcilers.NodeUpgradeReconciler).NotTo(BeNil())
}

func TestFactoryWithControlPlaneUpgradeReconciler(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	logger := nullLog()
	ctrl := gomock.NewController(t)
	manager := mocks.NewMockManager(ctrl)
	manager.EXPECT().GetClient().AnyTimes()
	manager.EXPECT().GetScheme().AnyTimes()

	f := controllers.NewFactory(logger, manager).
		WithControlPlaneUpgradeReconciler()

	// testing idempotence
	f.WithControlPlaneUpgradeReconciler()

	reconcilers, err := f.Build(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(reconcilers.ControlPlaneUpgradeReconciler).NotTo(BeNil())
}

func TestFactoryWithMachineDeploymentUpgradeReconciler(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	logger := nullLog()
	ctrl := gomock.NewController(t)
	manager := mocks.NewMockManager(ctrl)
	manager.EXPECT().GetClient().AnyTimes()
	manager.EXPECT().GetScheme().AnyTimes()

	f := controllers.NewFactory(logger, manager).
		WithMachineDeploymentUpgradeReconciler()

	// testing idempotence
	f.WithMachineDeploymentUpgradeReconciler()

	reconcilers, err := f.Build(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(reconcilers.MachineDeploymentUpgradeReconciler).NotTo(BeNil())
}
