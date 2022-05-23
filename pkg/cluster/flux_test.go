package cluster_test

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/cluster"
)

const (
	EksaGitPrivateKeyTokenEnv = "EKSA_GIT_PRIVATE_KEY"
	EksaGitKnownHostsFileEnv  = "EKSA_GIT_KNOWN_HOSTS"
	RsaAlgorithm              = "rsa"
	EcdsaAlgorithm            = "ecdsa"
	Ed25519Algorithm          = "ed25519"
)

type testContext struct {
	privateKeyFile         string
	isPrivateKeyFileSet    bool
	gitKnownHostsFile      string
	isGitKnownHostsFileSet bool
}

func (tctx *testContext) SaveContext() {
	tctx.privateKeyFile, tctx.isPrivateKeyFileSet = os.LookupEnv(EksaGitPrivateKeyTokenEnv)
	tctx.gitKnownHostsFile, tctx.isGitKnownHostsFileSet = os.LookupEnv(EksaGitKnownHostsFileEnv)
	os.Setenv(EksaGitPrivateKeyTokenEnv, "my/private/key")
	os.Setenv(EksaGitKnownHostsFileEnv, "my/known/hosts")
}

func (tctx *testContext) RestoreContext() {
	if tctx.isPrivateKeyFileSet {
		os.Setenv(EksaGitPrivateKeyTokenEnv, tctx.privateKeyFile)
	} else {
		os.Unsetenv(EksaGitPrivateKeyTokenEnv)
	}
	if tctx.isGitKnownHostsFileSet {
		os.Setenv(EksaGitKnownHostsFileEnv, tctx.gitKnownHostsFile)
	} else {
		os.Unsetenv(EksaGitKnownHostsFileEnv)
	}
}

func setupContext(t *testing.T) {
	var tctx testContext
	tctx.SaveContext()
	t.Cleanup(func() {
		tctx.RestoreContext()
	})
}

func TestValidateFluxConfigSuccess(t *testing.T) {
	tests := []struct {
		testName    string
		fileName    string
		gitProvider bool
	}{
		{
			testName: "valid 1.19 github",
			fileName: "testdata/cluster_1_19_flux_github.yaml",
		},
		{
			testName:    "valid 1.19 git",
			fileName:    "testdata/cluster_1_19_flux_git.yaml",
			gitProvider: true,
		},
		{
			testName:    "valid ssh key algo",
			fileName:    "testdata/cluster_1_19_flux_validgit_sshkey.yaml",
			gitProvider: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			c := cluster.NewConfigManager()
			c.RegisterValidations(
				func(c *cluster.Config) error {
					if c.FluxConfig != nil {
						return c.FluxConfig.Validate()
					}
					return nil
				},
			)

			if tt.gitProvider {
				setupContext(t)
			}

			config, err := cluster.ParseConfigFromFile(tt.fileName)
			g.Expect(c.Validate(config)).To(Succeed())
			if err != nil {
				t.Fatalf("Validate() error = %v, wantErr %v", err, nil)
			}
		})
	}
}

func TestValidateFluxConfigError(t *testing.T) {
	tests := []struct {
		testName    string
		fileName    string
		gitProvider bool
		error       string
	}{
		{
			testName: "file doesn't exist",
			fileName: "testdata/fake_file.yaml",
			error:    "reading cluster config file: open testdata/fake_file.yaml: no such file or directory",
		},
		{
			testName: "not parseable file",
			fileName: "testdata/not_parseable_fluxconfig.yaml",
		},
		{
			testName: "empty owner",
			fileName: "testdata/cluster_invalid_flux_unset_gitowner.yaml",
			error:    "'owner' is not set or empty in githubProviderConfig; owner is a required field",
		},
		{
			testName: "empty repo",
			fileName: "testdata/cluster_invalid_flux_unset_gitrepo.yaml",
			error:    "'repository' is not set or empty in githubProviderConfig; repository is a required field",
		},
		{
			testName: "empty repo url",
			fileName: "testdata/cluster_invalid_flux_unset_gitrepourl.yaml",
			error:    "'repositoryUrl' is not set or empty in gitProviderConfig; repositoryUrl is a required field",
		},
		{
			testName:    "invalid repo url",
			fileName:    "testdata/cluster_invalid_flux_gitrepourl.yaml",
			gitProvider: true,
			error:       fmt.Sprintf("invalid repository url scheme: %s", "http"),
		},
		{
			testName: "invalid sshkey algo",
			fileName: "testdata/cluster_invalid_flux_wrong_gitsshkeyalgo.yaml",
			error:    fmt.Sprintf("'sshKeyAlgorithm' does not have a valid value in gitProviderConfig; sshKeyAlgorithm must be amongst %s, %s, %s", RsaAlgorithm, EcdsaAlgorithm, Ed25519Algorithm),
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			c := cluster.NewConfigManager()
			c.RegisterValidations(
				func(c *cluster.Config) error {
					if c.FluxConfig != nil {
						return c.FluxConfig.Validate()
					}
					return nil
				},
			)

			if tt.gitProvider {
				setupContext(t)
			}

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
