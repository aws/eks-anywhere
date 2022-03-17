package test

import (
	"embed"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/version"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type ClusterSpecOpt func(*cluster.Spec)

//go:embed testdata
var configFS embed.FS

func NewClusterSpec(opts ...ClusterSpecOpt) *cluster.Spec {
	s := cluster.NewSpec()
	s.Cluster = &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fluxAddonTestCluster",
		},
		Spec: v1alpha1.ClusterSpec{
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{}},
		},
	}
	s.VersionsBundle = &cluster.VersionsBundle{
		VersionsBundle: &releasev1alpha1.VersionsBundle{},
		KubeDistro:     &cluster.KubeDistro{},
	}
	s.Bundles = &releasev1alpha1.Bundles{}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func NewFullClusterSpec(t *testing.T, clusterConfigFile string) *cluster.Spec {
	s, err := cluster.NewSpecFromClusterConfig(
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
	content, err := configFS.ReadFile("testdata/bundles_template.yaml")
	if err != nil {
		t.Fatalf("Failed to read embed bundles manifest: %s", err)
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed getting path to current file")
	}

	templateValues := map[string]string{
		"TestPath": filepath.Dir(filename),
	}

	bundlesContent, err := templater.Execute(string(content), templateValues)
	if err != nil {
		t.Fatalf("Failed writing new bundles file: %v", err)
	}

	bundles := &releasev1alpha1.Bundles{}
	if err = yaml.Unmarshal(bundlesContent, bundles); err != nil {
		t.Fatalf("Failed to unmarshal bundles manifest: %s", err)
	}

	return bundles
}

func EksdRelease(t *testing.T) *eksdv1alpha1.Release {
	t.Helper()
	content, err := configFS.ReadFile("testdata/kubernetes-1-21-eks-4.yaml")
	if err != nil {
		t.Fatalf("Failed to read embed eksd release: %s", err)
	}

	eksd := &eksdv1alpha1.Release{}
	if err = yaml.Unmarshal(content, eksd); err != nil {
		t.Fatalf("Failed to unmarshal eksd manifest: %s", err)
	}

	return eksd
}

func SetTag(image *releasev1alpha1.Image, tag string) {
	image.URI = fmt.Sprintf("%s:%s", image.Image(), tag)
}
