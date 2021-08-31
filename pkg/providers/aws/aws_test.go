package aws_test

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	mockswriter "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/aws"
	mocksaws "github.com/aws/eks-anywhere/pkg/providers/aws/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
)

func TestProviderGenerateDeploymentFileSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	tests := []struct {
		testName       string
		clusterSpec    *cluster.Spec
		providerConfig v1alpha1.AWSDatacenterConfig
		wantFileName   string
	}{
		{
			testName: "no AWS options",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Name = "test-cluster"
				s.Spec.KubernetesVersion = "1.19"
				s.Spec.ControlPlaneConfiguration.Count = 3
				s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
				s.VersionsBundle.KubeDistro = kubeDistro
			}),
			providerConfig: v1alpha1.AWSDatacenterConfig{},
			wantFileName:   "testdata/deployment_default_aws_expected.yaml",
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
			wantFileName: "testdata/deployment_with_ami_aws_expected.yaml",
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
			wantFileName: "testdata/deployment_with_ami_region_aws_expected.yaml",
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
			wantFileName:   "testdata/deployment_full_oidc_aws_expected.yaml",
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
			fileName := fmt.Sprintf("%s-eks-a-cluster.yaml", tt.clusterSpec.ObjectMeta.Name)
			oriCluster := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					KubernetesVersion: v1alpha1.Kube118,
				},
			}
			kubectl.EXPECT().GetEksaCluster(ctx, cluster).Return(oriCluster, nil)
			kubectl.EXPECT().GetEksaAWSDatacenterConfig(ctx, tt.providerConfig.Name, cluster.KubeconfigFile).Return(&tt.providerConfig, nil)
			got, err := p.GenerateDeploymentFileForUpgrade(ctx, bootstrapCluster, cluster, tt.clusterSpec, fileName)
			if err != nil {
				t.Fatalf("provider.GenerateDeploymentFile() error = %v, wantErr nil", err)
			}

			test.AssertFilesEquals(t, got, tt.wantFileName)
		})
	}
}

func TestProviderGenerateDeploymentFileWriterError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	client := mocksaws.NewMockProviderClient(mockCtrl)
	kubectl := mocksaws.NewMockProviderKubectlClient(mockCtrl)
	writer := mockswriter.NewMockFileWriter(mockCtrl)
	writer.EXPECT().Write(test.OfType("string"), gomock.Any()).Return("", errors.New("error write")).Times(1)

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Name = "test-cluster"
		s.Spec.KubernetesVersion = "1.19"
		s.Spec.ControlPlaneConfiguration.Count = 3
		s.Spec.WorkerNodeGroupConfigurations[0].Count = 3
		s.VersionsBundle.KubeDistro = kubeDistro
	})
	providerConfig := v1alpha1.AWSDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-cluster",
		},
		Spec: v1alpha1.AWSDatacenterConfigSpec{
			AmiID:  "anotherAMIID",
			Region: "eu-central-1",
		},
	}

	cluster := &types.Cluster{
		Name: "test",
	}
	p := aws.NewProvider(&providerConfig, "test-cluster", client, kubectl, writer, test.FakeNow)
	fileName := fmt.Sprintf("%s-eks-a-cluster.yaml", clusterSpec.ObjectMeta.Name)
	_, err := p.GenerateDeploymentFileForCreate(ctx, cluster, clusterSpec, fileName)
	if err == nil {
		t.Fatalf("provider.GenerateDeploymentFile() error = nil, want not nil")
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
