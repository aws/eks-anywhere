package aws_test

import (
	"context"
	_ "embed"
	"testing"

	"github.com/golang/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers/aws"
	mocksaws "github.com/aws/eks-anywhere/pkg/providers/aws/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
)

func TestProviderGenerateCAPISpecForUpgradeSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	tests := []struct {
		testName       string
		clusterSpec    *cluster.Spec
		providerConfig v1alpha1.AWSDatacenterConfig
		wantCPFile     string
		wantMDFile     string
	}{
		{
			testName: "no AWS options",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Name = "test-cluster"
				s.Namespace = "test-namespace"
				s.Spec.KubernetesVersion = "1.19"
				s.Spec.ControlPlaneConfiguration.Count = 3
				s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
				s.VersionsBundle.KubeDistro = kubeDistro
			}),
			providerConfig: v1alpha1.AWSDatacenterConfig{},
			wantCPFile:     "testdata/deployment_default_aws_cp_expected.yaml",
			wantMDFile:     "testdata/deployment_default_aws_md_expected.yaml",
		},
		{
			testName: "with AMI",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Name = "test-cluster"
				s.Spec.KubernetesVersion = "1.19"
				s.Spec.ControlPlaneConfiguration.Count = 3
				s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
				s.VersionsBundle.KubeDistro = kubeDistro
			}),
			providerConfig: v1alpha1.AWSDatacenterConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-cluster",
				},
				Spec: v1alpha1.AWSDatacenterConfigSpec{
					AmiID: "anotherAMIID",
				},
			},
			wantCPFile: "testdata/deployment_with_ami_aws_cp_expected.yaml",
			wantMDFile: "testdata/deployment_with_ami_aws_md_expected.yaml",
		},
		{
			testName: "with AMI and Region",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Name = "test-cluster"
				s.Spec.KubernetesVersion = "1.19"
				s.Spec.ControlPlaneConfiguration.Count = 3
				s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
				s.VersionsBundle.KubeDistro = kubeDistro
			}),
			providerConfig: v1alpha1.AWSDatacenterConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-cluster",
				},
				Spec: v1alpha1.AWSDatacenterConfigSpec{
					AmiID:  "anotherAMIID",
					Region: "eu-central-1",
				},
			},
			wantCPFile: "testdata/deployment_with_ami_region_aws_cp_expected.yaml",
			wantMDFile: "testdata/deployment_with_ami_region_aws_md_expected.yaml",
		},
		{
			testName: "with full oidc",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Name = "test-cluster"
				s.Spec.KubernetesVersion = "1.19"
				s.Spec.ControlPlaneConfiguration.Count = 3
				s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
				s.VersionsBundle.KubeDistro = kubeDistro

				s.OIDCConfig = &v1alpha1.OIDCConfig{
					Spec: v1alpha1.OIDCConfigSpec{
						ClientId:     "my-client-id",
						IssuerUrl:    "https://mydomain.com/issuer",
						GroupsClaim:  "claim1",
						GroupsPrefix: "prefix-for-groups",
						RequiredClaims: []v1alpha1.OIDCConfigRequiredClaim{
							{
								Claim: "sub",
								Value: "test",
							},
						},
						UsernameClaim:  "username-claim",
						UsernamePrefix: "username-prefix",
					},
				}
			}),
			providerConfig: v1alpha1.AWSDatacenterConfig{},
			wantCPFile:     "testdata/deployment_full_oidc_aws_cp_expected.yaml",
			wantMDFile:     "testdata/deployment_full_oidc_aws_md_expected.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			_, writer := test.NewWriter(t)
			ctx := context.Background()
			client := mocksaws.NewMockProviderClient(mockCtrl)
			kubectl := mocksaws.NewMockProviderKubectlClient(mockCtrl)
			p := aws.NewProvider(&tt.providerConfig, tt.clusterSpec.ObjectMeta.Name, client, kubectl, writer, test.FakeNow)
			cluster := &types.Cluster{
				Name: "test",
			}
			bootstrapCluster := &types.Cluster{
				Name: "bootstrap-test",
			}
			oriCluster := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					KubernetesVersion: v1alpha1.Kube118,
				},
			}
			kubectl.EXPECT().GetEksaCluster(ctx, cluster).Return(oriCluster, nil)
			kubectl.EXPECT().GetEksaAWSDatacenterConfig(ctx, tt.providerConfig.Name, cluster.KubeconfigFile, tt.clusterSpec.Namespace).Return(&tt.providerConfig, nil)
			cp, md, err := p.GenerateCAPISpecForUpgrade(ctx, bootstrapCluster, cluster, tt.clusterSpec)
			if err != nil {
				t.Fatalf("provider.GenerateCAPISpecForUpgrade() error = %v, wantErr nil", err)
			}

			test.AssertContentToFile(t, string(cp), tt.wantCPFile)
			test.AssertContentToFile(t, string(md), tt.wantMDFile)
		})
	}
}

var kubeDistro = &cluster.KubeDistro{
	Kubernetes: cluster.VersionedRepository{
		Repository: "public.ecr.aws/eks-distro/kubernetes",
		Tag:        "v1.19.6-eks-1-19-2",
	},
	CoreDNS: cluster.VersionedRepository{
		Repository: "public.ecr.aws/eks-distro/coredns",
		Tag:        "v1.8.0-eks-1-19-2",
	},
	Etcd: cluster.VersionedRepository{
		Repository: "public.ecr.aws/eks-distro/etcd-io",
		Tag:        "v3.4.14-eks-1-19-2",
	},
}
