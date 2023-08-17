package cluster_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/version"
)

func TestFileSpecBuilderBuildError(t *testing.T) {
	tests := []struct {
		testName          string
		releaseURL        string
		clusterConfigFile string
		cliVersion        string
	}{
		{
			testName:          "Reading cluster config",
			clusterConfigFile: "",
			releaseURL:        "testdata/simple_release.yaml",
			cliVersion:        "v1.0.0",
		},
		{
			testName:          "Cli version not supported",
			clusterConfigFile: "testdata/cluster_1_19.yaml",
			releaseURL:        "testdata/simple_release.yaml",
			cliVersion:        "v1.0.0",
		},
		{
			testName:          "Kubernetes version not supported",
			clusterConfigFile: "testdata/cluster_1_18.yaml",
			releaseURL:        "testdata/simple_release.yaml",
			cliVersion:        "v0.0.1",
		},
		{
			testName:          "Reading EkdD Release",
			clusterConfigFile: "testdata/cluster_1_19.yaml",
			releaseURL:        "testdata/release_bundle_missing_eksd.yaml",
			cliVersion:        "v0.0.1",
		},
		{
			testName:          "Worker EkdD Release",
			clusterConfigFile: "testdata/cluster_worker_k8s.yaml",
			releaseURL:        "testdata/simple_release.yaml",
			cliVersion:        "v0.0.1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)

			v := version.Info{GitVersion: tt.cliVersion}
			reader := files.NewReader()
			b := cluster.NewFileSpecBuilder(reader, v, cluster.WithReleasesManifest(tt.releaseURL))

			g.Expect(b.Build(tt.clusterConfigFile)).Error().NotTo(Succeed())
		})
	}
}

func TestFileSpecBuilderBuildSuccess(t *testing.T) {
	g := NewWithT(t)

	v := version.Info{GitVersion: "v0.0.1"}
	reader := files.NewReader()
	b := cluster.NewFileSpecBuilder(reader, v, cluster.WithReleasesManifest("testdata/simple_release.yaml"))

	gotSpec, err := b.Build("testdata/cluster_1_19.yaml")

	g.Expect(err).NotTo(HaveOccurred())
	validateSpecFromSimpleBundle(t, gotSpec)
}

func TestNewSpecWithBundlesOverrideValid(t *testing.T) {
	g := NewWithT(t)

	v := version.Info{GitVersion: "v0.0.1"}
	reader := files.NewReader()
	b := cluster.NewFileSpecBuilder(reader, v,
		cluster.WithReleasesManifest("testdata/simple_release.yaml"),
		cluster.WithOverrideBundlesManifest("testdata/simple_bundle.yaml"),
	)

	gotSpec, err := b.Build("testdata/cluster_1_19.yaml")

	g.Expect(err).NotTo(HaveOccurred())
	validateSpecFromSimpleBundle(t, gotSpec)
}
