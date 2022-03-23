package bootstrapper_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/bootstrapper/mocks"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/types"
)

func TestBootstrapperCreateBootstrapClusterSuccessNoExtraObjects(t *testing.T) {
	kubeconfigFile := "c.kubeconfig"
	clusterName := "cluster-name"
	clusterSpec, wantCluster := given(t, clusterName, kubeconfigFile)

	tests := []struct {
		testName string
		opts     []bootstrapper.BootstrapClusterOption
	}{
		{
			testName: "no options",
			opts:     []bootstrapper.BootstrapClusterOption{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ctx := context.Background()
			b, client := newBootstrapper(t)
			client.EXPECT().CreateBootstrapCluster(ctx, clusterSpec).Return(kubeconfigFile, nil)
			client.EXPECT().GetNamespace(ctx, kubeconfigFile, constants.EksaSystemNamespace)

			got, err := b.CreateBootstrapCluster(ctx, clusterSpec, tt.opts...)
			if err != nil {
				t.Fatalf("Bootstrapper.CreateBootstrapCluster() error = %v, wantErr nil", err)
			}

			if !reflect.DeepEqual(got, wantCluster) {
				t.Fatalf("Bootstrapper.CreateBootstrapCluster() cluster = %#v, want %#v", got, wantCluster)
			}
		})
	}
}

func TestBootstrapperCreateBootstrapClusterSuccessExtraObjects(t *testing.T) {
	kubeconfigFile := "c.kubeconfig"
	clusterName := "cluster-name"
	clusterSpec, wantCluster := given(t, clusterName, kubeconfigFile)
	clusterSpec.VersionsBundle.KubeVersion = "1.20"
	clusterSpec.VersionsBundle.KubeDistro.CoreDNS.Tag = "v1.8.3-eks-1-20-1"

	ctx := context.Background()
	b, client := newBootstrapper(t)
	client.EXPECT().CreateBootstrapCluster(ctx, clusterSpec).Return(kubeconfigFile, nil)
	client.EXPECT().GetNamespace(ctx, kubeconfigFile, constants.EksaSystemNamespace)
	client.EXPECT().ApplyKubeSpecFromBytes(ctx, wantCluster, gomock.Any())

	got, err := b.CreateBootstrapCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("Bootstrapper.CreateBootstrapCluster() error = %v, wantErr nil", err)
	}

	if !reflect.DeepEqual(got, wantCluster) {
		t.Fatalf("Bootstrapper.CreateBootstrapCluster() cluster = %#v, want %#v", got, wantCluster)
	}
}

func TestBootstrapperDeleteBootstrapClusterNoBootstrap(t *testing.T) {
	cluster := &types.Cluster{
		Name:           "cluster-name",
		KubeconfigFile: "c.kubeconfig",
	}

	ctx := context.Background()
	b, client := newBootstrapper(t)
	client.EXPECT().ClusterExists(ctx, cluster.Name).Return(false, nil)
	err := b.DeleteBootstrapCluster(ctx, cluster, false)
	if err != nil {
		t.Fatalf("Bootstrapper.DeleteBootstrapCluster() error = %v, wantErr nil", err)
	}
}

func TestBootstrapperDeleteBootstrapClusterNoKubeconfig(t *testing.T) {
	cluster := &types.Cluster{
		Name:           "cluster-name",
		KubeconfigFile: "",
	}

	ctx := context.Background()
	b, client := newBootstrapper(t)

	client.EXPECT().GetKubeconfig(ctx, cluster.Name).Return("c.kubeconfig", nil)
	client.EXPECT().ClusterExists(ctx, cluster.Name).Return(true, nil)
	client.EXPECT().ValidateClustersCRD(ctx, cluster).Return(nil)
	client.EXPECT().GetClusters(ctx, cluster).Return(nil, nil)
	client.EXPECT().DeleteBootstrapCluster(ctx, cluster).Return(nil)

	err := b.DeleteBootstrapCluster(ctx, cluster, false)
	if err != nil {
		t.Fatalf("Bootstrapper.DeleteBootstrapCluster() error = %v, wantErr nil", err)
	}
}

func TestBootstrapperDeleteBootstrapClusterNoClusterCRD(t *testing.T) {
	cluster := &types.Cluster{
		Name:           "cluster-name",
		KubeconfigFile: "c.kubeconfig",
	}

	ctx := context.Background()
	b, client := newBootstrapper(t)

	client.EXPECT().ClusterExists(ctx, cluster.Name).Return(true, nil)
	client.EXPECT().ValidateClustersCRD(ctx, cluster).Return(errors.New("cluster crd not found"))
	client.EXPECT().DeleteBootstrapCluster(ctx, cluster).Return(nil)

	err := b.DeleteBootstrapCluster(ctx, cluster, false)
	if err != nil {
		t.Fatalf("Bootstrapper.DeleteBootstrapCluster() error = %v, wantErr nil", err)
	}
}

func TestBootstrapperDeleteBootstrapClusterNoManagement(t *testing.T) {
	cluster := &types.Cluster{
		Name:           "cluster-name",
		KubeconfigFile: "c.kubeconfig",
	}

	ctx := context.Background()
	b, client := newBootstrapper(t)

	client.EXPECT().ClusterExists(ctx, cluster.Name).Return(true, nil)
	client.EXPECT().ValidateClustersCRD(ctx, cluster).Return(nil)
	client.EXPECT().DeleteBootstrapCluster(ctx, cluster).Return(nil)
	client.EXPECT().GetClusters(ctx, cluster).Return(nil, nil)

	err := b.DeleteBootstrapCluster(ctx, cluster, false)
	if err != nil {
		t.Fatalf("Bootstrapper.DeleteBootstrapCluster() error = %v, wantErr nil", err)
	}
}

func TestBootstrapperDeleteBootstrapClusterErrorWithManagement(t *testing.T) {
	cluster := &types.Cluster{
		Name:           "cluster-name",
		KubeconfigFile: "c.kubeconfig",
	}

	ctx := context.Background()
	b, client := newBootstrapper(t)

	client.EXPECT().ClusterExists(ctx, cluster.Name).Return(true, nil)
	client.EXPECT().ValidateClustersCRD(ctx, cluster).Return(nil)

	capiClusters := []types.CAPICluster{
		{
			Metadata: types.Metadata{
				Name: "cluster-name",
			},
			Status: types.ClusterStatus{
				Phase: "Provisioned",
			},
		},
	}
	client.EXPECT().GetClusters(ctx, cluster).Return(capiClusters, nil)

	err := b.DeleteBootstrapCluster(ctx, cluster, true)
	if err == nil {
		t.Fatalf("Bootstrapper.DeleteBootstrapCluster() error == nil, wantErr %v", err)
	}
}

func TestBootstrapperDeleteBootstrapClusterCreateOrDelete(t *testing.T) {
	tests := []struct {
		testName     string
		clusterPhase string
	}{
		{
			testName:     "ok to delete if phase is Pending",
			clusterPhase: "Pending",
		},
		{
			testName:     "ok to delete if phase is Provisioning",
			clusterPhase: "Provisioning",
		},
		{
			testName:     "ok to delete if phase is Failed",
			clusterPhase: "Failed",
		},
		{
			testName:     "ok to delete if phase is Failed",
			clusterPhase: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			cluster := &types.Cluster{
				Name:           "cluster-name",
				KubeconfigFile: "c.kubeconfig",
			}

			ctx := context.Background()
			b, client := newBootstrapper(t)

			client.EXPECT().ClusterExists(ctx, cluster.Name).Return(true, nil)
			client.EXPECT().ValidateClustersCRD(ctx, cluster).Return(nil)

			capiClusters := []types.CAPICluster{
				{
					Metadata: types.Metadata{
						Name: "cluster-name",
					},
					Status: types.ClusterStatus{
						Phase: tt.clusterPhase,
					},
				},
			}
			client.EXPECT().GetClusters(ctx, cluster).Return(capiClusters, nil)
			client.EXPECT().DeleteBootstrapCluster(ctx, cluster).Return(nil)

			err := b.DeleteBootstrapCluster(ctx, cluster, false)
			if err != nil {
				t.Fatalf("It shoud be possible to delete a management cluster while in %s phase. Expected error == nil, got %v", tt.clusterPhase, err)
			}
		})
	}
}

func TestBootstrapperDeleteBootstrapClusterUpgrade(t *testing.T) {
	tests := []struct {
		testName     string
		clusterPhase string
	}{
		{
			testName:     "do not delete if phase is Provisioned during an Upgrade",
			clusterPhase: "Provisioned",
		},
		{
			testName:     "do not delete if phase is Pending during an Upgrade",
			clusterPhase: "Pending",
		},
		{
			testName:     "do not delete if phase is Provisioning during an Upgrade",
			clusterPhase: "Provisioning",
		},
		{
			testName:     "do not delete if phase is Failed during an Upgrade",
			clusterPhase: "Failed",
		},
		{
			testName:     "do not delete if phase is Failed during an Upgrade",
			clusterPhase: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			cluster := &types.Cluster{
				Name:           "cluster-name",
				KubeconfigFile: "c.kubeconfig",
			}

			ctx := context.Background()
			b, client := newBootstrapper(t)

			client.EXPECT().ClusterExists(ctx, cluster.Name).Return(true, nil)
			client.EXPECT().ValidateClustersCRD(ctx, cluster).Return(nil)

			capiClusters := []types.CAPICluster{
				{
					Metadata: types.Metadata{
						Name: "cluster-name",
					},
					Status: types.ClusterStatus{
						Phase: tt.clusterPhase,
					},
				},
			}
			client.EXPECT().GetClusters(ctx, cluster).Return(capiClusters, nil)
			client.EXPECT().DeleteBootstrapCluster(ctx, cluster).Return(nil).Times(0)

			err := b.DeleteBootstrapCluster(ctx, cluster, true)
			if err == nil {
				t.Fatalf("upgrade should not delete a management cluster. Expected error == nil, got %v", err)
			}
		})
	}
}

func newBootstrapper(t *testing.T) (*bootstrapper.Bootstrapper, *mocks.MockClusterClient) {
	mockCtrl := gomock.NewController(t)

	client := mocks.NewMockClusterClient(mockCtrl)
	b := bootstrapper.New(client)
	return b, client
}

func given(t *testing.T, clusterName, kubeconfig string) (clusterSpec *cluster.Spec, wantCluster *types.Cluster) {
	return test.NewClusterSpec(func(s *cluster.Spec) {
			s.Cluster.Name = clusterName
			s.VersionsBundle.KubeVersion = "1.19"
			s.VersionsBundle.KubeDistro.CoreDNS.Tag = "v1.8.3-eks-1-20-1"
		}), &types.Cluster{
			Name:           clusterName,
			KubeconfigFile: kubeconfig,
		}
}
