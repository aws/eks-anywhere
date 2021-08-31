package cluster_test

import (
	"context"
	_ "embed"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/cluster/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
)

func TestApplyExtraObjects(t *testing.T) {
	tests := []struct {
		testName             string
		clusterSpec          *cluster.Spec
		resourcesFileContent string
	}{
		{
			testName:             "kube 1.20, coreDNS v1.8.3-eks-1-20-1, extra cluster role",
			clusterSpec:          clusterSpec(t, "1.20", "v1.8.3-eks-1-20-1"),
			resourcesFileContent: "testdata/kube1-20.yaml",
		},
		{
			testName:             "kube 1.20, coreDNS v1.8.3, extra cluster role",
			clusterSpec:          clusterSpec(t, "1.20", "v1.8.3"),
			resourcesFileContent: "testdata/kube1-20.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ctx := context.Background()
			c := &types.Cluster{}

			mockCtrl := gomock.NewController(t)
			client := mocks.NewMockClusterClient(mockCtrl)
			client.EXPECT().ApplyKubeSpecFromBytes(ctx, c, gomock.Any()).Do(
				func(ctx context.Context, cluster *types.Cluster, data []byte) error {
					test.AssertContentToFile(t, string(data), tt.resourcesFileContent)
					return nil
				},
			)

			err := cluster.ApplyExtraObjects(ctx, client, c, tt.clusterSpec)
			if err != nil {
				t.Fatalf("cluster.ApplyExtraObjects err = %v, want err = nil", err)
			}
		})
	}
}

func TestApplyExtraObjectsNoObjects(t *testing.T) {
	ctx := context.Background()
	c := &types.Cluster{}
	clusterSpec := clusterSpec(t, "1.19", "v1.8.3")

	mockCtrl := gomock.NewController(t)
	client := mocks.NewMockClusterClient(mockCtrl)

	err := cluster.ApplyExtraObjects(ctx, client, c, clusterSpec)
	if err != nil {
		t.Fatalf("cluster.ApplyExtraObjects err = %v, want err = nil", err)
	}
}

func TestApplyExtraObjectsErroClient(t *testing.T) {
	ctx := context.Background()
	c := &types.Cluster{}

	clusterConfig := clusterSpec(t, "1.20", "v1.8.3")

	mockCtrl := gomock.NewController(t)
	client := mocks.NewMockClusterClient(mockCtrl)
	client.EXPECT().ApplyKubeSpecFromBytes(ctx, c, gomock.Any()).Return(errors.New("error applying kube resources"))

	err := cluster.ApplyExtraObjects(ctx, client, c, clusterConfig)
	if err == nil {
		t.Fatal("cluster.ApplyExtraObjects err = nil, want err not nil")
	}
}

func clusterSpec(t *testing.T, kubeVersion, coreDNSVersion string) *cluster.Spec {
	return test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.KubeVersion = kubeVersion
		s.VersionsBundle.KubeDistro.CoreDNS.Tag = coreDNSVersion
	})
}
