package test

import (
	"embed"
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/version"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type ClusterSpecOpt func(*cluster.Spec)

//go:embed testdata
var configFS embed.FS

func NewClusterSpec(opts ...ClusterSpecOpt) *cluster.Spec {
	s := &cluster.Spec{
		Cluster: &v1alpha1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: "fluxAddonTestCluster",
			},
			Spec: v1alpha1.ClusterSpec{
				WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{}},
			},
		},
		VersionsBundle: &cluster.VersionsBundle{
			VersionsBundle: &releasev1alpha1.VersionsBundle{},
			KubeDistro:     &cluster.KubeDistro{},
		},
		Bundles: &releasev1alpha1.Bundles{},
	}

	for _, opt := range opts {
		opt(s)
	}

	s.SetDefaultGitOps()
	return s
}

func NewFullClusterSpec(t *testing.T, clusterConfigFile string) *cluster.Spec {
	s, err := cluster.NewSpec(
		clusterConfigFile,
		version.Info{GitVersion: "v0.0.0-dev"},
		cluster.WithReleasesManifest("embed:///testdata/releases.yaml"),
		cluster.WithEmbedFS(configFS),
	)
	if err != nil {
		t.Fatalf("can't build cluster spec for tests: %v", err)
	}

	return s
}

func Bundles(t *testing.T) *releasev1alpha1.Bundles {
	content, err := configFS.ReadFile("testdata/bundles.yaml")
	if err != nil {
		t.Fatalf("Failed to read embed bundles manifest: %s", err)
	}

	bundles := &releasev1alpha1.Bundles{}
	if err = yaml.Unmarshal(content, bundles); err != nil {
		t.Fatalf("Failed to unmarshal bundles manifest: %s", err)
	}

	return bundles
}

func SetTag(image *releasev1alpha1.Image, tag string) {
	image.URI = fmt.Sprintf("%s:%s", image.Image(), tag)
}
