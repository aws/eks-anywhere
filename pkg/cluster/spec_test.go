package cluster_test

import (
	"embed"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/version"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

//go:embed testdata
var testdataFS embed.FS

func TestNewSpecInvalidClusterConfig(t *testing.T) {
	v := version.Info{}
	if _, err := cluster.NewSpecFromClusterConfig("testdata/cluster_invalid_kinds.yaml", v); err == nil {
		t.Fatal("NewSpec() error nil , want err not nil")
	}
}

func TestNewSpecError(t *testing.T) {
	tests := []struct {
		testName          string
		releaseURL        string
		clusterConfigFile string
		cliVersion        string
	}{
		{
			testName:          "InvalidReleasesManifest",
			clusterConfigFile: "testdata/cluster_1_19.yaml",
			releaseURL:        "testdata/cluster_1_19.yaml",
			cliVersion:        "",
		},
		{
			testName:          "CliVersionNotSupported",
			clusterConfigFile: "testdata/cluster_1_19.yaml",
			releaseURL:        "testdata/simple_release.yaml",
			cliVersion:        "v1.0.0",
		},
		{
			testName:          "ReadingBundleUrl",
			clusterConfigFile: "testdata/cluster_1_19.yaml",
			releaseURL:        "testdata/release_empty_bundle.yaml",
			cliVersion:        "v0.0.1",
		},
		{
			testName:          "InvalidCliVersion",
			clusterConfigFile: "testdata/cluster_1_19.yaml",
			releaseURL:        "testdata/simple_release.yaml",
			cliVersion:        "v1.0.X",
		},
		{
			testName:          "InvalidCliVersion",
			clusterConfigFile: "testdata/cluster_1_19.yaml",
			releaseURL:        "testdata/invalid_release_version.yaml",
			cliVersion:        "v0.0.1",
		},
		{
			testName:          "InvalidBundleManifest",
			clusterConfigFile: "testdata/cluster_1_19.yaml",
			releaseURL:        "testdata/release_invalid_bundle.yaml",
			cliVersion:        "v0.0.1",
		},
		{
			testName:          "KubernetesVersionNotSupported",
			clusterConfigFile: "testdata/cluster_1_18.yaml",
			releaseURL:        "testdata/simple_release.yaml",
			cliVersion:        "v0.0.1",
		},
		{
			testName:          "ReadingEkdDRelease",
			clusterConfigFile: "testdata/cluster_1_19.yaml",
			releaseURL:        "testdata/release_bundle_missing_eksd.yaml",
			cliVersion:        "v0.0.1",
		},
		{
			testName:          "InvalidEkdDRelease",
			clusterConfigFile: "testdata/cluster_1_19.yaml",
			releaseURL:        "testdata/release_bundle_invalid_eksd.yaml",
			cliVersion:        "v0.0.1",
		},
		{
			testName:          "EkdDReleaseMissingAssetForImage",
			clusterConfigFile: "testdata/cluster_1_19.yaml",
			releaseURL:        "testdata/release_bundle_eksd_missing_nodedriver.yaml",
			cliVersion:        "v0.0.1",
		},
		{
			testName:          "EkdDReleaseMissingAssetForRepository",
			clusterConfigFile: "testdata/cluster_1_19.yaml",
			releaseURL:        "testdata/release_bundle_eksd_missing_kubeapiserver.yaml",
			cliVersion:        "v0.0.1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			v := version.Info{GitVersion: tt.cliVersion}
			if _, err := cluster.NewSpecFromClusterConfig(tt.clusterConfigFile, v, cluster.WithReleasesManifest(tt.releaseURL)); err == nil {
				t.Fatal("NewSpec() error nil, want err not nil")
			}
		})
	}
}

func TestNewSpecValidEmbedManifest(t *testing.T) {
	v := version.Info{GitVersion: "v0.0.1"}
	_, err := cluster.NewSpecFromClusterConfig(
		"testdata/cluster_1_19.yaml",
		v,
		cluster.WithReleasesManifest("embed:///testdata/simple_release.yaml"),
		cluster.WithEmbedFS(testdataFS),
	)
	if err != nil {
		t.Fatalf("NewSpec() error = %v, want err nil", err)
	}
}

func TestNewSpecValid(t *testing.T) {
	v := version.Info{GitVersion: "v0.0.1"}
	gotSpec, err := cluster.NewSpecFromClusterConfig("testdata/cluster_1_19.yaml", v, cluster.WithReleasesManifest("testdata/simple_release.yaml"))
	if err != nil {
		t.Fatalf("NewSpec() error = %v, want err nil", err)
	}

	validateSpecFromSimpleBundle(t, gotSpec)
}

func TestNewSpecFromClusterConfigTinkerbellValid(t *testing.T) {
	g := NewWithT(t)
	v := version.Info{GitVersion: "v0.0.1"}
	wantConfigs := map[string]*anywherev1.TinkerbellTemplateConfig{
		"tink-test": {
			TypeMeta: metav1.TypeMeta{
				Kind:       "TinkerbellTemplateConfig",
				APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "tink-test",
			},
			Spec: anywherev1.TinkerbellTemplateConfigSpec{
				Template: tinkerbell.Workflow{
					Version:       "0.1",
					Name:          "tink-test",
					GlobalTimeout: 6000,
					Tasks: []tinkerbell.Task{
						{
							Name:       "tink-test",
							WorkerAddr: "{{.device_1}}",
							Volumes: []string{
								"/dev:/dev",
								"/dev/console:/dev/console",
								"/lib/firmware:/lib/firmware:ro",
							},
							Actions: []tinkerbell.Action{
								{
									Name:    "stream-image",
									Image:   "image2disk:v1.0.0",
									Timeout: 360,
									Environment: map[string]string{
										"IMG_URL":    "",
										"DEST_DISK":  "/dev/sda",
										"COMPRESSED": "true",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	gotSpec, err := cluster.NewSpecFromClusterConfig("testdata/cluster_tinkerbell_1_19.yaml", v,
		cluster.WithReleasesManifest("testdata/simple_release.yaml"),
	)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(gotSpec.TinkerbellTemplateConfigs).To(Equal(wantConfigs))
	validateSpecFromSimpleBundle(t, gotSpec)
}

func TestNewSpecWithBundlesOverrideValid(t *testing.T) {
	v := version.Info{GitVersion: "v0.0.1"}
	gotSpec, err := cluster.NewSpecFromClusterConfig("testdata/cluster_1_19.yaml", v,
		cluster.WithReleasesManifest("testdata/invalid_release_version.yaml"),
		cluster.WithOverrideBundlesManifest("testdata/simple_bundle.yaml"),
	)
	if err != nil {
		t.Fatalf("NewSpec() error = %v, want err nil", err)
	}

	validateSpecFromSimpleBundle(t, gotSpec)
}

func validateSpecFromSimpleBundle(t *testing.T, gotSpec *cluster.Spec) {
	validateVersionedRepo(t, gotSpec.VersionsBundle.KubeDistro.Kubernetes, "public.ecr.aws/eks-distro/kubernetes", "v1.19.8-eks-1-19-4")
	validateVersionedRepo(t, gotSpec.VersionsBundle.KubeDistro.CoreDNS, "public.ecr.aws/eks-distro/coredns", "v1.8.0-eks-1-19-4")
	validateVersionedRepo(t, gotSpec.VersionsBundle.KubeDistro.Etcd, "public.ecr.aws/eks-distro/etcd-io", "v3.4.14-eks-1-19-4")
	validateImageURI(t, gotSpec.VersionsBundle.KubeDistro.NodeDriverRegistrar, "public.ecr.aws/eks-distro/kubernetes-csi/node-driver-registrar:v2.1.0-eks-1-19-4")
	validateImageURI(t, gotSpec.VersionsBundle.KubeDistro.LivenessProbe, "public.ecr.aws/eks-distro/kubernetes-csi/livenessprobe:v2.2.0-eks-1-19-4")
	validateImageURI(t, gotSpec.VersionsBundle.KubeDistro.ExternalAttacher, "public.ecr.aws/eks-distro/kubernetes-csi/external-attacher:v3.1.0-eks-1-19-4")
	validateImageURI(t, gotSpec.VersionsBundle.KubeDistro.ExternalProvisioner, "public.ecr.aws/eks-distro/kubernetes-csi/external-provisioner:v2.1.1-eks-1-19-4")
	validateImageURI(t, gotSpec.VersionsBundle.KubeDistro.EtcdImage, "public.ecr.aws/eks-distro/etcd-io/etcd:v3.4.14-eks-1-19-4")
	if gotSpec.VersionsBundle.KubeDistro.EtcdVersion != "3.4.14" {
		t.Errorf("GetNewSpec() = Spec: Invalid etcd version, got %s, want 3.4.14", gotSpec.VersionsBundle.KubeDistro.EtcdVersion)
	}
}

func validateImageURI(t *testing.T, gotImage v1alpha1.Image, wantURI string) {
	if gotImage.URI != wantURI {
		t.Errorf("GetNewSpec() = Spec: Invalid kubernetes URI, got %s, want %s", gotImage.URI, wantURI)
	}
}

func validateVersionedRepo(t *testing.T, gotImage cluster.VersionedRepository, wantRepo, wantTag string) {
	if gotImage.Repository != wantRepo {
		t.Errorf("GetNewSpec() = Spec: Invalid kubernetes repo, got %s, want %s", gotImage.Repository, wantRepo)
	}
	if gotImage.Tag != wantTag {
		t.Errorf("GetNewSpec() = Spec: Invalid kubernetes repo, got %s, want %s", gotImage.Tag, wantTag)
	}
}

func TestSpecLoadManifestError(t *testing.T) {
	s := cluster.NewSpec()
	tests := []struct {
		testName string
		manifest v1alpha1.Manifest
	}{
		{
			testName: "InvalidURI",
			manifest: v1alpha1.Manifest{URI: ":domain.com/"},
		},
		{
			testName: "ErrorReadingFile",
			manifest: v1alpha1.Manifest{URI: "testdata/fake.yaml"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if _, err := s.LoadManifest(tt.manifest); err == nil {
				t.Fatal("spec.LoadManifest() error = nil, want err not nil")
			}
		})
	}
}

func TestSpecLoadManifestSuccess(t *testing.T) {
	filename := "testdata/cluster_1_19.yaml"
	wantFilename := "cluster_1_19.yaml"
	s := cluster.NewSpec()
	manifest := v1alpha1.Manifest{URI: filename}
	m, err := s.LoadManifest(manifest)
	if err != nil {
		t.Fatalf("spec.LoadManifest() error = %v, want err nil", err)
	}

	if m.Filename != wantFilename {
		t.Errorf("spec.LoadManifest() manifest.Filename = %s, want %s", m.Filename, wantFilename)
	}

	test.AssertContentToFile(t, string(m.Content), filename)
}

func TestBundlesRefDefaulter(t *testing.T) {
	tests := []struct {
		name         string
		bundles      *v1alpha1.Bundles
		config, want *cluster.Config
	}{
		{
			name: "no bundles ref",
			bundles: &v1alpha1.Bundles{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bundles-1",
					Namespace: "eksa-system",
				},
			},
			config: &cluster.Config{
				Cluster: &anywherev1.Cluster{},
			},
			want: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					Spec: anywherev1.ClusterSpec{
						BundlesRef: &anywherev1.BundlesRef{},
					},
				},
			},
		},
		{
			name: "with previous bundles ref",
			bundles: &v1alpha1.Bundles{
				ObjectMeta: metav1.ObjectMeta{
					Name: "bundles-1",
				},
			},
			config: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					Spec: anywherev1.ClusterSpec{
						BundlesRef: &anywherev1.BundlesRef{
							Name:       "bundles-2",
							Namespace:  "default",
							APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
						},
					},
				},
			},
			want: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					Spec: anywherev1.ClusterSpec{
						BundlesRef: &anywherev1.BundlesRef{
							Name:       "bundles-2",
							Namespace:  "default",
							APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			defaulter := cluster.BundlesRefDefaulter()
			g.Expect(defaulter(tt.config)).To(Succeed())
			g.Expect(tt.config).To(Equal(tt.want))
		})
	}
}

func TestBuildSpecFromBundles(t *testing.T) {
	g := NewWithT(t)
	clus := &anywherev1.Cluster{
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube123,
		},
	}

	bundles := &v1alpha1.Bundles{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bundles-1",
		},
		Spec: v1alpha1.BundlesSpec{
			VersionsBundles: []v1alpha1.VersionsBundle{
				{
					KubeVersion: "1.23",
				},
			},
		},
	}

	eksdRelease, wantKubeDistro := wantKubeDistroForEksdRelease()

	wantSpec := &cluster.Spec{
		Config: &cluster.Config{
			Cluster: clus,
		},
		VersionsBundle: &cluster.VersionsBundle{
			VersionsBundle: &bundles.Spec.VersionsBundles[0],
			KubeDistro:     wantKubeDistro,
		},
		Bundles: bundles,
	}

	spec, err := cluster.BuildSpecFromBundles(clus, bundles, cluster.WithEksdRelease(eksdRelease))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(spec.Config).To(Equal(wantSpec.Config))
	g.Expect(spec.AWSIamConfig).To(Equal(wantSpec.AWSIamConfig))
	g.Expect(spec.OIDCConfig).To(Equal(wantSpec.OIDCConfig))
	g.Expect(spec.Bundles).To(Equal(wantSpec.Bundles))
	g.Expect(spec.VersionsBundle).To(Equal(wantSpec.VersionsBundle))
}
