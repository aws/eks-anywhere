package controllers_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/controllers"
	"github.com/aws/eks-anywhere/controllers/mocks"
)

func TestFactoryBuildAllReconcilers(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	logger := nullLog()
	ctrl := gomock.NewController(t)
	manager := mocks.NewMockManager(ctrl)
	manager.EXPECT().GetClient().AnyTimes()
	manager.EXPECT().GetScheme().AnyTimes()

	f := controllers.NewFactory(logger, manager).
		WithClusterReconciler().
		WithVSphereDatacenterReconciler()

	// testing idempotence
	f.WithClusterReconciler().
		WithVSphereDatacenterReconciler()

	reconcilers, err := f.Build(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(reconcilers.ClusterReconciler).NotTo(BeNil())
	g.Expect(reconcilers.VSphereDatacenterReconciler).NotTo(BeNil())
}
