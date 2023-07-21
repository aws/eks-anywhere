package cluster_test

import (
	"embed"
	"testing"

	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/manifests/eksd"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

//go:embed testdata
var testdataFS embed.FS

func TestNewSpecError(t *testing.T) {
	version := test.DevEksaVersion()
	tests := []struct {
		name        string
		config      *cluster.Config
		bundles     *releasev1.Bundles
		eksdRelease *eksdv1.Release
		error       string
	}{
		{
			name: "no VersionsBundle for kube version",
			config: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					Spec: anywherev1.ClusterSpec{
						KubernetesVersion: anywherev1.Kube124,
						EksaVersion:       &version,
					},
				},
			},
			bundles: &releasev1.Bundles{
				Spec: releasev1.BundlesSpec{
					Number: 2,
				},
			},
			eksdRelease: &eksdv1.Release{},
			error:       "kubernetes version 1.24 is not supported by bundles manifest 2",
		},
		{
			name: "invalid eks-d release",
			config: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					Spec: anywherev1.ClusterSpec{
						KubernetesVersion: anywherev1.Kube124,
						EksaVersion:       &version,
					},
				},
			},
			bundles: &releasev1.Bundles{
				Spec: releasev1.BundlesSpec{
					Number: 2,
					VersionsBundles: []releasev1.VersionsBundle{
						{
							KubeVersion: "1.24",
						},
					},
				},
			},
			eksdRelease: &eksdv1.Release{},
			error:       "is no present in eksd release",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(cluster.NewSpec(tt.config, tt.bundles, tt.eksdRelease, test.EKSARelease())).Error().To(
				MatchError(ContainSubstring(tt.error)),
			)
		})
	}
}

func TestNewSpecValid(t *testing.T) {
	g := NewWithT(t)
	version := test.DevEksaVersion()
	config := &cluster.Config{
		Cluster: &anywherev1.Cluster{
			Spec: anywherev1.ClusterSpec{
				KubernetesVersion: anywherev1.Kube124,
				EksaVersion:       &version,
			},
		},
		OIDCConfigs: map[string]*anywherev1.OIDCConfig{
			"myconfig": {},
		},
		AWSIAMConfigs: map[string]*anywherev1.AWSIamConfig{
			"myconfig": {},
		},
	}
	bundles := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			Number: 2,
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1.24",
				},
			},
		},
	}
	eksdRelease := readEksdRelease(t, "testdata/eksd_valid.yaml")

	spec, err := cluster.NewSpec(config, bundles, eksdRelease, test.EKSARelease())
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(spec.AWSIamConfig).NotTo(BeNil())
	g.Expect(spec.OIDCConfig).NotTo(BeNil())
}

func TestSpecDeepCopy(t *testing.T) {
	g := NewWithT(t)
	r := files.NewReader()
	yaml, err := r.ReadFile("testdata/docker_cluster_oidc_awsiam_flux.yaml")
	g.Expect(err).To(Succeed())
	config, err := cluster.ParseConfig(yaml)

	g.Expect(err).To(Succeed())
	bundles := test.Bundles(t)
	eksd := test.EksdRelease()
	spec, err := cluster.NewSpec(config, bundles, eksd, test.EKSARelease())
	g.Expect(err).To(Succeed())

	g.Expect(spec.DeepCopy()).To(Equal(spec))
}

func TestBundlesRefDefaulter(t *testing.T) {
	tests := []struct {
		name         string
		bundles      *releasev1.Bundles
		config, want *cluster.Config
	}{
		{
			name: "no bundles ref",
			bundles: &releasev1.Bundles{
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
			bundles: &releasev1.Bundles{
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

func validateSpecFromSimpleBundle(t *testing.T, gotSpec *cluster.Spec) {
	validateVersionedRepo(t, gotSpec.VersionsBundle.KubeDistro.Kubernetes, "public.ecr.aws/eks-distro/kubernetes", "v1.19.8-eks-1-19-4")
	validateVersionedRepo(t, gotSpec.VersionsBundle.KubeDistro.CoreDNS, "public.ecr.aws/eks-distro/coredns", "v1.8.0-eks-1-19-4")
	validateVersionedRepo(t, gotSpec.VersionsBundle.KubeDistro.Etcd, "public.ecr.aws/eks-distro/etcd-io", "v3.4.14-eks-1-19-4")
	validateImageURI(t, gotSpec.VersionsBundle.KubeDistro.NodeDriverRegistrar, "public.ecr.aws/eks-distro/kubernetes-csi/node-driver-registrar:v2.1.0-eks-1-19-4")
	validateImageURI(t, gotSpec.VersionsBundle.KubeDistro.LivenessProbe, "public.ecr.aws/eks-distro/kubernetes-csi/livenessprobe:v2.2.0-eks-1-19-4")
	validateImageURI(t, gotSpec.VersionsBundle.KubeDistro.ExternalAttacher, "public.ecr.aws/eks-distro/kubernetes-csi/external-attacher:v3.1.0-eks-1-19-4")
	validateImageURI(t, gotSpec.VersionsBundle.KubeDistro.ExternalProvisioner, "public.ecr.aws/eks-distro/kubernetes-csi/external-provisioner:v2.1.1-eks-1-19-4")
	validateImageURI(t, gotSpec.VersionsBundle.KubeDistro.EtcdImage, "public.ecr.aws/eks-distro/etcd-io/etcd:v3.4.14-eks-1-19-4")
	validateImageURI(t, gotSpec.VersionsBundle.KubeDistro.KubeProxy, "public.ecr.aws/eks-distro/kubernetes/kube-proxy:v1.19.8-eks-1-19-4")
	if gotSpec.VersionsBundle.KubeDistro.EtcdVersion != "3.4.14" {
		t.Errorf("GetNewSpec() = Spec: Invalid etcd version, got %s, want 3.4.14", gotSpec.VersionsBundle.KubeDistro.EtcdVersion)
	}
}

func validateImageURI(t *testing.T, gotImage releasev1.Image, wantURI string) {
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

func readEksdRelease(tb testing.TB, url string) *eksdv1.Release {
	tb.Helper()
	r := files.NewReader()
	release, err := eksd.ReadManifest(r, url)
	if err != nil {
		tb.Fatalf("Failed reading eks-d manifest: %s", err)
	}

	return release
}
