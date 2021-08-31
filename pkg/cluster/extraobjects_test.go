package cluster_test

import (
	_ "embed"
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

func TestBuildExtraObjects(t *testing.T) {
	tests := []struct {
		testName             string
		clusterSpec          *cluster.Spec
		resourcesFileContent map[string]string
	}{
		{
			testName:    "kube 1.19, no extra objects",
			clusterSpec: clusterSpec(t, "1.19", "1.8.3-eks-1-20-1"),
		},
		{
			testName:    "kube 1.20, coreDNS v1.8.3-eks-1-20-1, extra cluster role",
			clusterSpec: clusterSpec(t, "1.20", "v1.8.3-eks-1-20-1"),
			resourcesFileContent: map[string]string{
				"core-dns-clusterrole": "objects/coredns_clusterrole.yaml",
			},
		},
		{
			testName:    "kube 1.20, coreDNS v1.8.3, extra cluster role",
			clusterSpec: clusterSpec(t, "1.20", "v1.8.3"),
			resourcesFileContent: map[string]string{
				"core-dns-clusterrole": "objects/coredns_clusterrole.yaml",
			},
		},
		{
			testName:    "kube 1.21, coreDNS v1.8.3-eks-1-21-3, extra cluster role",
			clusterSpec: clusterSpec(t, "1.21", "v1.8.3-eks-1-21-3"),
			resourcesFileContent: map[string]string{
				"core-dns-clusterrole": "objects/coredns_clusterrole.yaml",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := cluster.BuildExtraObjects(tt.clusterSpec)
			for name, content := range got {
				file, ok := tt.resourcesFileContent[name]
				if !ok {
					t.Errorf("BuildExtraObjects resource %s was not expected", name)
				}

				test.AssertContentToFile(t, string(content), file)
			}

			for name := range tt.resourcesFileContent {
				_, ok := got[name]
				if !ok {
					t.Errorf("BuildExtraObjects resource %s not expected but not present in response", name)
				}
			}
		})
	}
}
