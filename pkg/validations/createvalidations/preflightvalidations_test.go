package createvalidations_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
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
	version := "v0.0.0-dev"
	objects := []client.Object{test.EKSARelease()}
	opts := &validations.Opts{
		Kubectl:           k,
		Spec:              clusterSpec,
		WorkloadCluster:   c,
		ManagementCluster: c,
		CliVersion:        version,
		KubeClient:        test.NewFakeKubeClient(objects...),
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
	tt.Expect(validations.ProcessValidationResults(tt.c.PreflightValidations(tt.ctx))).To(Succeed())
}

func TestPreFlightValidationsWorkloadCluster(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	mgmtClusterName := "mgmt-cluster"
	tt.c.Opts.Spec.Cluster.SetManagedBy(mgmtClusterName)
	tt.c.Opts.Spec.Cluster.Spec.ManagementCluster.Name = mgmtClusterName
	tt.c.Opts.ManagementCluster.Name = mgmtClusterName
	version := test.DevEksaVersion()

	mgmt := &v1alpha1.Cluster{
		ObjectMeta: v1.ObjectMeta{
			Name: "mgmt-cluster",
		},
		Spec: v1alpha1.ClusterSpec{
			ManagementCluster: v1alpha1.ManagementCluster{
				Name: "mgmt-cluster",
			},
			BundlesRef: &anywherev1.BundlesRef{
				Name:      "bundles-29",
				Namespace: constants.EksaSystemNamespace,
			},
			EksaVersion: &version,
		},
	}

	tt.k.EXPECT().GetClusters(tt.ctx, tt.c.Opts.WorkloadCluster).Return(nil, nil)
	tt.k.EXPECT().ValidateClustersCRD(tt.ctx, tt.c.Opts.WorkloadCluster).Return(nil)
	tt.k.EXPECT().ValidateEKSAClustersCRD(tt.ctx, tt.c.Opts.WorkloadCluster).Return(nil)
	tt.k.EXPECT().GetEksaCluster(tt.ctx, tt.c.Opts.ManagementCluster, mgmtClusterName).Return(mgmt, nil).MaxTimes(3)

	tt.Expect(validations.ProcessValidationResults(tt.c.PreflightValidations(tt.ctx))).To(Succeed())
}
