package framework

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

func TestWithFluxGitConfig(t *testing.T) {
	g := NewWithT(t)
	t.Setenv(JobIdVar, "build:example-1")
	t.Setenv(GitRepoSshUrl, "ssh://git@example.invalid/repository")
	config := &cluster.Config{Cluster: &v1alpha1.Cluster{}}

	WithFluxGitConfig(api.WithFluxConfigName("custom-flux"))(config)

	g.Expect(config.FluxConfig.Spec.Github).To(BeNil())
	g.Expect(config.FluxConfig.Spec.Git.RepositoryUrl).To(Equal("ssh://git@example.invalid/repository"))
	g.Expect(config.FluxConfig.Spec.Branch).To(Equal("build-example-1"))
	g.Expect(config.FluxConfig.Spec.ClusterConfigPath).To(Equal("build-example-1"))
	g.Expect(config.Cluster.Spec.GitOpsRef.Name).To(Equal("custom-flux"))
	g.Expect(config.Cluster.Spec.GitOpsRef.Kind).To(Equal(v1alpha1.FluxConfigKind))
}

func TestWithFluxGitConfigSeparateJobs(t *testing.T) {
	g := NewWithT(t)
	t.Setenv(GitRepoSshUrl, "ssh://git@example.invalid/repository")
	first := &cluster.Config{Cluster: &v1alpha1.Cluster{}}
	second := &cluster.Config{Cluster: &v1alpha1.Cluster{}}
	t.Setenv(JobIdVar, "build:first")
	WithFluxGitConfig()(first)
	t.Setenv(JobIdVar, "build:second")
	WithFluxGitConfig()(second)
	g.Expect(first.FluxConfig.Spec.Git.RepositoryUrl).To(Equal(second.FluxConfig.Spec.Git.RepositoryUrl))
	g.Expect(first.FluxConfig.Spec.Branch).NotTo(Equal(second.FluxConfig.Spec.Branch))
	g.Expect(first.FluxConfig.Spec.ClusterConfigPath).NotTo(Equal(second.FluxConfig.Spec.ClusterConfigPath))
}

func TestWithFluxGitEnvVarCheckDoesNotRequireGithubToken(t *testing.T) {
	g := NewWithT(t)
	value, existed := os.LookupEnv(GithubTokenVar)
	if err := os.Unsetenv(GithubTokenVar); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if existed {
			if err := os.Setenv(GithubTokenVar, value); err != nil {
				t.Fatal(err)
			}
		}
	})
	for _, name := range fluxGitRequiredEnvVars {
		t.Setenv(name, "example")
	}
	capture := &cleanupCaptureT{T: t}
	e := &ClusterE2ETest{T: capture}
	WithFluxGitEnvVarCheck()(e)
	WithFluxGitCleanup()(e)
	g.Expect(capture.cleanups).To(HaveLen(1))
}

type cleanupCaptureT struct {
	T
	cleanups []func()
}

func (t *cleanupCaptureT) Cleanup(f func()) {
	t.cleanups = append(t.cleanups, f)
}
