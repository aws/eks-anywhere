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
			ProviderName: "vsphere",
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
