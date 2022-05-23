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

func TestConfigManagerValidateGitOpsConfigSuccess(t *testing.T) {
	tests := []struct {
		testName string
		fileName string
	}{
		{
			testName: "valid 1.19",
			fileName: "testdata/cluster_1_19_gitops.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			c := cluster.NewConfigManager()
			c.RegisterValidations(
				func(c *cluster.Config) error {
					if c.GitOpsConfig != nil {
						return c.GitOpsConfig.Validate()
					}
					return nil
				},
			)

			config, err := cluster.ParseConfigFromFile(tt.fileName)
			g.Expect(c.Validate(config)).To(Succeed())
			if err != nil {
				t.Fatalf("Validate() error = %v, wantErr %v", err, nil)
			}
		})
	}
}

func TestConfigManagerValidateGitOpsConfigError(t *testing.T) {
	tests := []struct {
		testName string
		fileName string
		error    string
	}{
		{
			testName: "file doesn't exist",
			fileName: "testdata/fake_file.yaml",
			error:    "reading cluster config file: open testdata/fake_file.yaml: no such file or directory",
		},
		{
			testName: "not parseable file",
			fileName: "testdata/not_parseable_gitopsconfig.yaml",
		},
		{
			testName: "empty owner",
			fileName: "testdata/cluster_invalid_gitops_unset_gitowner.yaml",
			error:    "'owner' is not set or empty in gitOps.flux; owner is a required field",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			c := cluster.NewConfigManager()
			c.RegisterValidations(
				func(c *cluster.Config) error {
					if c.GitOpsConfig != nil {
						return c.GitOpsConfig.Validate()
					}
					return nil
				},
			)

			config, err := cluster.ParseConfigFromFile(tt.fileName)
			if err != nil {
				g.Expect(err).To(MatchError(ContainSubstring(
					tt.error,
				)))
			} else {
				g.Expect(c.Validate(config)).To(MatchError(ContainSubstring(
					tt.error,
				)))
			}
		})
	}
}
