package reconciler_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/networking/reconciler"
	"github.com/aws/eks-anywhere/pkg/networking/reconciler/mocks"
)

func TestReconcilerReconcileCilium(t *testing.T) {
	ctx := context.Background()
	logger := test.NewNullLogger()
	client := fake.NewClientBuilder().Build()
	spec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Spec.ClusterNetwork.CNIConfig = &v1alpha1.CNIConfig{
			Cilium: &v1alpha1.CiliumConfig{},
		}
	})

	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	ciliumReconciler := mocks.NewMockCiliumReconciler(ctrl)
	ciliumReconciler.EXPECT().Reconcile(ctx, logger, client, spec)

	r := reconciler.New(ciliumReconciler)
	result, err := r.Reconcile(ctx, logger, client, spec)
	g.Expect(result).To(Equal(controller.Result{}))
	g.Expect(err).NotTo(HaveOccurred())
}

func TestReconcilerReconcileUnsupportedCNI(t *testing.T) {
	ctx := context.Background()
	logger := test.NewNullLogger()
	client := fake.NewClientBuilder().Build()
	spec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Spec.ClusterNetwork.CNIConfig = &v1alpha1.CNIConfig{}
	})

	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	ciliumReconciler := mocks.NewMockCiliumReconciler(ctrl)

	r := reconciler.New(ciliumReconciler)
	_, err := r.Reconcile(ctx, logger, client, spec)
	g.Expect(err).To(MatchError(ContainSubstring("unsupported CNI, only Cilium is supported at this time")))
}
