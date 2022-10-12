package createvalidations_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/createvalidations"
	"github.com/aws/eks-anywhere/pkg/validations/mocks"
)

type preflightValidationsTest struct {
	*WithT
	ctx context.Context
	k   *mocks.MockKubectlClient
	c   *createvalidations.CreateValidations
}

func newPreflightValidationsTest(t *testing.T) *preflightValidationsTest {
	ctrl := gomock.NewController(t)
	k := mocks.NewMockKubectlClient(ctrl)
	c := &types.Cluster{
		KubeconfigFile: "kubeconfig",
	}
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{
			Name: "gitops",
		}
	})
	opts := &validations.Opts{
		Kubectl:           k,
		Spec:              clusterSpec,
		WorkloadCluster:   c,
		ManagementCluster: c,
	}
	return &preflightValidationsTest{
		WithT: NewWithT(t),
		ctx:   context.Background(),
		k:     k,
		c:     createvalidations.New(opts),
	}
}

func TestPreFlightValidationsGitProvider(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	tt.Expect(tt.c.PreflightValidations(tt.ctx)).To(Succeed())
}

func TestPreFlightValidationsWorkloadCluster(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	tt.c.Opts.Spec.Cluster.SetManagedBy("mgmt-cluster")

	tt.k.EXPECT().GetClusters(tt.ctx, tt.c.Opts.WorkloadCluster).Return(nil, nil)
	tt.k.EXPECT().ValidateClustersCRD(tt.ctx, tt.c.Opts.WorkloadCluster).Return(nil)
	tt.k.EXPECT().ValidateEKSAClustersCRD(tt.ctx, tt.c.Opts.WorkloadCluster).Return(nil)

	tt.Expect(tt.c.PreflightValidations(tt.ctx)).To(Succeed())
}
