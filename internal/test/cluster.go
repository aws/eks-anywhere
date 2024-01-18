package test

import (
	"embed"
	"fmt"
	"net"
	"path/filepath"
	"runtime"
	"testing"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/version"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type ClusterSpecOpt func(*cluster.Spec)

//go:embed testdata
var configFS embed.FS

// DevEksaVersion can be used in tests.
func DevEksaVersion() v1alpha1.EksaVersion {
	return v1alpha1.EksaVersion("v0.0.0-dev")
}

func NewClusterSpec(opts ...ClusterSpecOpt) *cluster.Spec {
	s := &cluster.Spec{}
	version := DevEksaVersion()
	s.Config = &cluster.Config{
		Cluster: &v1alpha1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: "fluxTestCluster",
			},
			Spec: v1alpha1.ClusterSpec{
				KubernetesVersion:             "1.19",
				WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{}},
				EksaVersion:                   &version,
			},
		},
	}
	s.VersionsBundles = map[v1alpha1.KubernetesVersion]*cluster.VersionsBundle{
		v1alpha1.Kube119: {
			VersionsBundle: &releasev1alpha1.VersionsBundle{
				EksD: releasev1alpha1.EksDRelease{
					Name:           "kubernetes-1-19-eks-7",
					EksDReleaseUrl: "embed:///testdata/release.yaml",
					KubeVersion:    "1.19",
				},
			},
			KubeDistro: &cluster.KubeDistro{},
		},
	}
	s.Bundles = &releasev1alpha1.Bundles{
		Spec: releasev1alpha1.BundlesSpec{
			VersionsBundles: []releasev1alpha1.VersionsBundle{
				{
					EksD: releasev1alpha1.EksDRelease{
						Name:           "kubernetes-1-19-eks-7",
						EksDReleaseUrl: "embed:///testdata/release.yaml",
						KubeVersion:    "1.19",
					},
				},
			},
		},
	}
	s.EKSARelease = EKSARelease()

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func NewFullClusterSpec(t *testing.T, clusterConfigFile string) *cluster.Spec {
	b := cluster.NewFileSpecBuilder(
		files.NewReader(files.WithEmbedFS(configFS)),
		version.Info{GitVersion: "v0.0.0-dev"},
		cluster.WithReleasesManifest("embed:///testdata/releases.yaml"),
	)
	s, err := b.Build(clusterConfigFile)
	if err != nil {
		t.Fatalf("can't build cluster spec for tests: %v", err)
	}

	return s
}

// NewClusterSpecForCluster builds a compliant [cluster.Spec] from a Cluster using a test
// Bundles and EKS-D Release.
func NewClusterSpecForCluster(tb testing.TB, c *v1alpha1.Cluster) *cluster.Spec {
	return NewClusterSpecForConfig(tb, &cluster.Config{Cluster: c})
}

// NewClusterSpecForConfig builds a compliant [cluster.Spec] from a [cluster.Config] using a test
// Bundles and EKS-D Release.
func NewClusterSpecForConfig(tb testing.TB, config *cluster.Config) *cluster.Spec {
	spec, err := cluster.NewSpec(
		config,
		Bundles(tb),
		EksdReleases(),
		EKSARelease(),
	)
	if err != nil {
		tb.Fatalf("Failed to build cluster spec: %s", err)
	}

	return spec
}

// Bundles returs a test Bundles. All the paths to referenced manifests are valid and can be read.
func Bundles(tb testing.TB) *releasev1alpha1.Bundles {
	content, err := configFS.ReadFile("testdata/bundles.yaml")
	if err != nil {
		tb.Fatalf("Failed to read embed bundles manifest: %s", err)
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		tb.Fatal("Failed getting path to current file")
	}

	templateValues := map[string]string{
		"TestPath": filepath.Dir(filename),
	}

	bundlesContent, err := templater.Execute(string(content), templateValues)
	if err != nil {
		tb.Fatalf("Failed writing new bundles file: %v", err)
	}

	bundles := &releasev1alpha1.Bundles{}
	if err = yaml.Unmarshal(bundlesContent, bundles); err != nil {
		tb.Fatalf("Failed to unmarshal bundles manifest: %s", err)
	}

	return bundles
}

// EksdReleaseFromTestData returns a test release struct for unit testing from a testdata file.
// See EksdRelease() for a static struct to test with.
func EksdReleaseFromTestData(t *testing.T) *eksdv1alpha1.Release {
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

// RegistryMirrorEndpoint returns the address of the registry mirror configured on the Cluster if any. Just the host and the port.
func RegistryMirrorEndpoint(cluster *v1alpha1.Cluster) string {
	if cluster.Spec.RegistryMirrorConfiguration != nil {
		return net.JoinHostPort(cluster.Spec.RegistryMirrorConfiguration.Endpoint, cluster.Spec.RegistryMirrorConfiguration.Port)
	}
	return ""
}
