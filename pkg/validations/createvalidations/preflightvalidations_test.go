package createvalidations_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/internal/test"
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
		s.Cluster.Spec.GitOpsRef = &anywherev1.Ref{
			Name: "gitops",
		}
	})
	version := "v0.19.0-dev+latest"
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

	mgmt := &anywherev1.Cluster{
		ObjectMeta: v1.ObjectMeta{
			Name: "mgmt-cluster",
		},
		Spec: anywherev1.ClusterSpec{
			ManagementCluster: anywherev1.ManagementCluster{
				Name: "mgmt-cluster",
			},
			BundlesRef: &anywherev1.BundlesRef{
				Name:      "bundles-29",
				Namespace: constants.EksaSystemNamespace,
			},
			EksaVersion: &version,
			ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
				KubeletConfiguration: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"staticPodPath": "path",
					},
				},
			},
		},
	}

	tt.c.Opts.Spec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef = &anywherev1.Ref{
		Name: "cpRef",
	}
	tt.c.Opts.Spec.VSphereMachineConfigs = map[string]*anywherev1.VSphereMachineConfig{
		"cpRef": {
			Spec: anywherev1.VSphereMachineConfigSpec{
				OSFamily: anywherev1.Bottlerocket,
			},
		},
	}

	tt.c.Opts.Spec.Cluster.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{
		{
			MachineGroupRef: &anywherev1.Ref{
				Name: "wnRef",
			},
		},
	}

	tt.c.Opts.Spec.VSphereMachineConfigs["wnRef"] = &anywherev1.VSphereMachineConfig{
		Spec: anywherev1.VSphereMachineConfigSpec{
			OSFamily: anywherev1.Bottlerocket,
		},
	}

	tt.k.EXPECT().GetClusters(tt.ctx, tt.c.Opts.WorkloadCluster).Return(nil, nil)
	tt.k.EXPECT().ValidateClustersCRD(tt.ctx, tt.c.Opts.WorkloadCluster).Return(nil)
	tt.k.EXPECT().ValidateEKSAClustersCRD(tt.ctx, tt.c.Opts.WorkloadCluster).Return(nil)
	tt.k.EXPECT().GetEksaCluster(tt.ctx, tt.c.Opts.ManagementCluster, mgmtClusterName).Return(mgmt, nil).MaxTimes(3)

	tt.Expect(validations.ProcessValidationResults(tt.c.PreflightValidations(tt.ctx))).To(Succeed())
}
