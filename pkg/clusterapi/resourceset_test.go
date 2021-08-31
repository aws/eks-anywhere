package clusterapi_test

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
)

func TestClusterResourceSetToYaml(t *testing.T) {
	tests := []struct {
		testName           string
		filesWithResources map[string]string
		wantFileContent    string
	}{
		{
			testName:           "no resources",
			filesWithResources: map[string]string{},
			wantFileContent:    "",
		},
		{
			testName: "one resource - cluster role",
			filesWithResources: map[string]string{
				"coredns-role": "testdata/coredns_clusterrole.yaml",
			},
			wantFileContent: "testdata/expected_crs_clusterrole.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			c := clusterapi.NewClusterResourceSet("cluster-name")
			for name, file := range tt.filesWithResources {
				content := test.ReadFile(t, file)
				c.AddResource(name, []byte(content))
			}

			got, err := c.ToYaml()
			if err != nil {
				t.Fatalf("ClusterResourceSet.ToYaml err = %v, want err = nil", err)
			}

			test.AssertContentToFile(t, string(got), tt.wantFileContent)
		})
	}
}
