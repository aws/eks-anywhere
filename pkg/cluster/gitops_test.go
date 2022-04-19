package cluster_test

import (
	"testing"

	. "github.com/onsi/gomega"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

const (
	owner             = "janedoe"
	repository        = "flux-fleet"
	fluxNamespace     = "test-ns"
	branch            = "test-branch"
	clusterConfigPath = "test-path"
	personal          = false
)

func TestGitOpsToFluxConversionProcessing(t *testing.T) {
	tests := []struct {
		name           string
		wantConfigPath string
		wantFluxSpec   anywherev1.FluxConfigSpec
	}{
		{
			name:           "workload cluster with GitOpsConfig",
			wantConfigPath: "testdata/cluster_gitops_1_21.yaml",
			wantFluxSpec: anywherev1.FluxConfigSpec{
				SystemNamespace:   fluxNamespace,
				ClusterConfigPath: clusterConfigPath,
				Branch:            branch,
				Github: &anywherev1.GithubProviderConfig{
					Owner:      owner,
					Repository: repository,
					Personal:   personal,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			config, err := cluster.ParseConfigFromFile(tt.wantConfigPath)
			if err != nil {
				t.Fatal("cluster.ParseConfigFromFile error != nil, want nil", err)
			}
			g.Expect(config.FluxConfig.Spec).To(Equal(tt.wantFluxSpec))
		})
	}
}
