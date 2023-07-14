package cluster_test

import (
	"context"
	"errors"
	"testing"

	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/cluster/mocks"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type buildSpecTest struct {
	*WithT
	ctx         context.Context
	ctrl        *gomock.Controller
	client      *mocks.MockClient
	cluster     *anywherev1.Cluster
	bundles     *releasev1.Bundles
	eksdRelease *eksdv1.Release
	kubeDistro  *cluster.KubeDistro
	eksaRelease *releasev1.EKSARelease
}

func newBuildSpecTest(t *testing.T) *buildSpecTest {
	kube122 := anywherev1.KubernetesVersion("1.22")
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)
	version := test.DevEksaVersion()
	cluster := &anywherev1.Cluster{
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube123,
			EksaVersion:       &version,
			WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Name:              "md-0",
					KubernetesVersion: &kube122,
					Count:             ptr.Int(1),
				},
			},
		},
	}
	bundles := &releasev1.Bundles{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bundles-1",
		},
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1.23",
					EksD: releasev1.EksDRelease{
						Name: "eksd-123",
					},
				},
				{
					KubeVersion: "1.22",
					EksD: releasev1.EksDRelease{
						Name: "eksd-122",
					},
				},
			},
		},
	}
	eksdRelease, kubeDistro := wantKubeDistroForEksdRelease()

	eksaRelease := &releasev1.EKSARelease{
		ObjectMeta: metav1.ObjectMeta{
			Name: "eksa-v0-0-0-dev",
		},
		Spec: releasev1.EKSAReleaseSpec{
			BundlesRef: releasev1.BundlesRef{
				Name:      "bundles-1",
				Namespace: "my-namespace",
			},
		},
	}

	return &buildSpecTest{
		WithT:       NewWithT(t),
		ctx:         context.Background(),
		ctrl:        ctrl,
		client:      client,
		cluster:     cluster,
		bundles:     bundles,
		eksdRelease: eksdRelease,
		kubeDistro:  kubeDistro,
		eksaRelease: eksaRelease,
	}
}

func (tt *buildSpecTest) expectGetEKSARelease() {
	tt.client.EXPECT().Get(tt.ctx, "eksa-v0-0-0-dev", "eksa-system", &releasev1.EKSARelease{}).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			o := obj.(*releasev1.EKSARelease)
			o.ObjectMeta = tt.eksaRelease.ObjectMeta
			o.Spec = tt.eksaRelease.Spec
			return nil
		},
	)
}

func (tt *buildSpecTest) expectGetBundles() {
	tt.client.EXPECT().Get(tt.ctx, "bundles-1", "my-namespace", &releasev1.Bundles{}).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			o := obj.(*releasev1.Bundles)
			o.ObjectMeta = tt.bundles.ObjectMeta
			o.Spec = tt.bundles.Spec
			return nil
		},
	)
}

func (tt *buildSpecTest) expectGetEksd() {
	tt.client.EXPECT().Get(tt.ctx, "eksd-123", "eksa-system", &eksdv1.Release{}).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			o := obj.(*eksdv1.Release)
			o.ObjectMeta = tt.eksdRelease.ObjectMeta
			o.Spec = tt.eksdRelease.Spec
			o.Status = tt.eksdRelease.Status
			return nil
		},
	)
	tt.client.EXPECT().Get(tt.ctx, "eksd-122", "eksa-system", &eksdv1.Release{}).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			o := obj.(*eksdv1.Release)
			o.ObjectMeta = tt.eksdRelease.ObjectMeta
			o.Spec = eksdv1.ReleaseSpec{
				Channel: "1-22",
			}
			o.Status = tt.eksdRelease.Status
			return nil
		},
	)
}

func TestBuildSpec(t *testing.T) {
	tt := newBuildSpecTest(t)
	tt.expectGetEKSARelease()
	tt.expectGetBundles()
	tt.expectGetEksd()

	_, kubeDistro := wantKubeDistroForEksdRelease()
	kubeDistro.EKSD.Channel = "1-22"

	wantSpec := &cluster.Spec{
		Config: &cluster.Config{
			Cluster:       tt.cluster,
			OIDCConfigs:   map[string]*anywherev1.OIDCConfig{},
			AWSIAMConfigs: map[string]*anywherev1.AWSIamConfig{},
		},
		Bundles: tt.bundles,
		VersionsBundles: map[anywherev1.KubernetesVersion]*cluster.VersionsBundle{
			anywherev1.Kube123: {
				VersionsBundle: &tt.bundles.Spec.VersionsBundles[0],
				KubeDistro:     tt.kubeDistro,
			},
			anywherev1.Kube122: {
				VersionsBundle: &tt.bundles.Spec.VersionsBundles[1],
				KubeDistro:     kubeDistro,
			},
		},
	}

	spec, err := cluster.BuildSpec(tt.ctx, tt.client, tt.cluster)
	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(spec.Config).To(Equal(wantSpec.Config))
	tt.Expect(spec.AWSIamConfig).To(Equal(wantSpec.AWSIamConfig))
	tt.Expect(spec.OIDCConfig).To(Equal(wantSpec.OIDCConfig))
	tt.Expect(spec.Bundles).To(Equal(wantSpec.Bundles))
	tt.Expect(spec.VersionsBundles).To(Equal(wantSpec.VersionsBundles))
}

func TestBuildSpecGetEKSAReleaseError(t *testing.T) {
	tt := newBuildSpecTest(t)
	tt.cluster.Spec.BundlesRef = nil
	tt.client.EXPECT().Get(tt.ctx, "eksa-v0-0-0-dev", "eksa-system", &releasev1.EKSARelease{}).Return(errors.New("client error"))

	_, err := cluster.BuildSpec(tt.ctx, tt.client, tt.cluster)
	tt.Expect(err).To(MatchError(ContainSubstring("error getting EKSARelease")))
}

func TestBuildSpecNilEksaVersion(t *testing.T) {
	tt := newBuildSpecTest(t)
	tt.cluster.Spec.BundlesRef = nil
	tt.cluster.Spec.EksaVersion = nil

	_, err := cluster.BuildSpec(tt.ctx, tt.client, tt.cluster)
	tt.Expect(err).To(MatchError(ContainSubstring("either cluster's EksaVersion or BundlesRef need to be set")))
}

func TestBuildSpecGetBundlesError(t *testing.T) {
	tt := newBuildSpecTest(t)
	tt.expectGetEKSARelease()
	tt.client.EXPECT().Get(tt.ctx, "bundles-1", "my-namespace", &releasev1.Bundles{}).Return(errors.New("client error"))

	_, err := cluster.BuildSpec(tt.ctx, tt.client, tt.cluster)
	tt.Expect(err).To(MatchError(ContainSubstring("client error")))
}

func TestBuildSpecGetEksdError(t *testing.T) {
	tt := newBuildSpecTest(t)
	tt.expectGetEKSARelease()
	tt.expectGetBundles()
	tt.client.EXPECT().Get(tt.ctx, "eksd-123", "eksa-system", &eksdv1.Release{}).Return(errors.New("client error"))

	_, err := cluster.BuildSpec(tt.ctx, tt.client, tt.cluster)
	tt.Expect(err).To(MatchError(ContainSubstring("client error")))
}

func TestBuildSpecBuildConfigError(t *testing.T) {
	tt := newBuildSpecTest(t)
	tt.cluster.Namespace = "default"
	tt.cluster.Spec.GitOpsRef = &anywherev1.Ref{
		Name: "my-flux",
		Kind: anywherev1.FluxConfigKind,
	}
	tt.client.EXPECT().Get(tt.ctx, "my-flux", "default", &anywherev1.FluxConfig{}).Return(errors.New("client error"))

	_, err := cluster.BuildSpec(tt.ctx, tt.client, tt.cluster)
	tt.Expect(err).To(MatchError(ContainSubstring("client error")))
}

func TestBuildSpecBuildConfigErrorEksd(t *testing.T) {
	tt := newBuildSpecTest(t)
	tt.expectGetEKSARelease()
	tt.expectGetBundles()
	tt.cluster.Spec.KubernetesVersion = "1.18"

	_, err := cluster.BuildSpec(tt.ctx, tt.client, tt.cluster)
	tt.Expect(err).To(MatchError(ContainSubstring("kubernetes version 1.18 is not supported by bundles manifest")))
}

func TestBuildSpecBuildConfigErrorEksdWorkerNodes(t *testing.T) {
	tt := newBuildSpecTest(t)
	kube118 := anywherev1.KubernetesVersion("1.18")
	tt.cluster.Spec.WorkerNodeGroupConfigurations[0].KubernetesVersion = &kube118
	tt.expectGetEKSARelease()
	tt.expectGetBundles()
	tt.client.EXPECT().Get(tt.ctx, "eksd-123", "eksa-system", &eksdv1.Release{}).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			o := obj.(*eksdv1.Release)
			o.ObjectMeta = tt.eksdRelease.ObjectMeta
			o.Status = tt.eksdRelease.Status
			return nil
		},
	)

	_, err := cluster.BuildSpec(tt.ctx, tt.client, tt.cluster)
	tt.Expect(err).To(MatchError(ContainSubstring("kubernetes version 1.18 is not supported by bundles manifest")))
}

func TestBuildSpecUnsupportedKubernetesVersionError(t *testing.T) {
	tt := newBuildSpecTest(t)
	tt.bundles.Spec.VersionsBundles = []releasev1.VersionsBundle{}
	tt.bundles.Spec.Number = 2
	tt.expectGetEKSARelease()
	tt.expectGetBundles()

	_, err := cluster.BuildSpec(tt.ctx, tt.client, tt.cluster)
	tt.Expect(err).To(MatchError(ContainSubstring("kubernetes version 1.23 is not supported by bundles manifest 2")))
}

func TestBuildSpecInitError(t *testing.T) {
	tt := newBuildSpecTest(t)
	tt.eksdRelease.Status.Components = []eksdv1.Component{}
	tt.expectGetEKSARelease()
	tt.expectGetBundles()
	tt.expectGetEksd()

	_, err := cluster.BuildSpec(tt.ctx, tt.client, tt.cluster)
	tt.Expect(err).To(MatchError(ContainSubstring("is no present in eksd release")))
}

func wantKubeDistroForEksdRelease() (*eksdv1.Release, *cluster.KubeDistro) {
	eksdRelease := &eksdv1.Release{
		ObjectMeta: metav1.ObjectMeta{
			Name: "eksd-123",
		},
		Spec: eksdv1.ReleaseSpec{
			Channel: "1-23",
		},
		Status: eksdv1.ReleaseStatus{
			Components: []eksdv1.Component{
				{
					Name:   "etcd",
					GitTag: "v3.4.14",
					Assets: []eksdv1.Asset{
						{
							Arch: []string{"amd64"},
							Archive: &eksdv1.AssetArchive{
								URI: "https://distro.eks.amazonaws.com/kubernetes-1-19/releases/4/artifacts/etcd/v3.4.14/etcd-linux-amd64-v3.4.14.tar.gz",
							},
						},
					},
				},
				{
					Name: "comp-1",
					Assets: []eksdv1.Asset{
						{
							Name: "external-provisioner-image",
							Image: &eksdv1.AssetImage{
								URI: "public.ecr.aws/eks-distro/kubernetes-csi/external-provisioner:v2.1.1",
							},
						},
						{
							Name: "node-driver-registrar-image",
							Image: &eksdv1.AssetImage{
								URI: "public.ecr.aws/eks-distro/kubernetes-csi/node-driver-registrar:v2.1.0",
							},
						},
						{
							Name: "livenessprobe-image",
							Image: &eksdv1.AssetImage{
								URI: "public.ecr.aws/eks-distro/kubernetes-csi/livenessprobe:v2.2.0",
							},
						},
						{
							Name: "external-attacher-image",
							Image: &eksdv1.AssetImage{
								URI: "public.ecr.aws/eks-distro/kubernetes-csi/external-attacher:v3.1.0",
							},
						},
						{
							Name: "pause-image",
							Image: &eksdv1.AssetImage{
								URI: "public.ecr.aws/eks-distro/kubernetes/pause:v1.19.8",
							},
						},
						{
							Name: "coredns-image",
							Image: &eksdv1.AssetImage{
								URI: "public.ecr.aws/eks-distro/coredns/coredns:v1.8.0",
							},
						},
						{
							Name: "etcd-image",
							Image: &eksdv1.AssetImage{
								URI: "public.ecr.aws/eks-distro/etcd-io/etcd:v3.4.14",
							},
						},
						{
							Name: "aws-iam-authenticator-image",
							Image: &eksdv1.AssetImage{
								URI: "public.ecr.aws/eks-distro/kubernetes-sigs/aws-iam-authenticator:v0.5.2",
							},
						},
						{
							Name: "kube-apiserver-image",
							Image: &eksdv1.AssetImage{
								URI: "public.ecr.aws/eks-distro/kubernetes/kube-apiserver:v1.19.8",
							},
						},
						{
							Name: "kube-proxy-image",
							Image: &eksdv1.AssetImage{
								URI: "public.ecr.aws/eks-distro/kubernetes/kube-proxy:v1.19.8",
							},
						},
					},
				},
			},
		},
	}

	kubeDistro := &cluster.KubeDistro{
		EKSD: cluster.EKSD{
			Channel: "1-23",
		},
		Kubernetes: cluster.VersionedRepository{
			Repository: "public.ecr.aws/eks-distro/kubernetes",
			Tag:        "v1.19.8",
		},
		CoreDNS: cluster.VersionedRepository{
			Repository: "public.ecr.aws/eks-distro/coredns",
			Tag:        "v1.8.0",
		},
		Etcd: cluster.VersionedRepository{
			Repository: "public.ecr.aws/eks-distro/etcd-io",
			Tag:        "v3.4.14",
		},
		NodeDriverRegistrar: releasev1.Image{
			URI: "public.ecr.aws/eks-distro/kubernetes-csi/node-driver-registrar:v2.1.0",
		},
		LivenessProbe: releasev1.Image{
			URI: "public.ecr.aws/eks-distro/kubernetes-csi/livenessprobe:v2.2.0",
		},
		ExternalAttacher: releasev1.Image{
			URI: "public.ecr.aws/eks-distro/kubernetes-csi/external-attacher:v3.1.0",
		},
		ExternalProvisioner: releasev1.Image{
			URI: "public.ecr.aws/eks-distro/kubernetes-csi/external-provisioner:v2.1.1",
		},
		Pause: releasev1.Image{
			URI: "public.ecr.aws/eks-distro/kubernetes/pause:v1.19.8",
		},
		EtcdImage: releasev1.Image{
			URI: "public.ecr.aws/eks-distro/etcd-io/etcd:v3.4.14",
		},
		AwsIamAuthImage: releasev1.Image{
			URI: "public.ecr.aws/eks-distro/kubernetes-sigs/aws-iam-authenticator:v0.5.2",
		},
		KubeProxy: releasev1.Image{
			URI: "public.ecr.aws/eks-distro/kubernetes/kube-proxy:v1.19.8",
		},
		EtcdVersion: "3.4.14",
		EtcdURL:     "https://distro.eks.amazonaws.com/kubernetes-1-19/releases/4/artifacts/etcd/v3.4.14/etcd-linux-amd64-v3.4.14.tar.gz",
	}

	return eksdRelease, kubeDistro
}

func TestManagementComponentsFromBundles(t *testing.T) {
	g := NewWithT(t)
	bundles := test.Bundle()
	got := cluster.ManagementComponentsFromBundles(bundles)
	want := &cluster.ManagementComponents{
		EksD:                   bundles.Spec.VersionsBundles[0].EksD,
		CertManager:            bundles.Spec.VersionsBundles[0].CertManager,
		ClusterAPI:             bundles.Spec.VersionsBundles[0].ClusterAPI,
		Bootstrap:              bundles.Spec.VersionsBundles[0].Bootstrap,
		ControlPlane:           bundles.Spec.VersionsBundles[0].ControlPlane,
		VSphere:                bundles.Spec.VersionsBundles[0].VSphere,
		CloudStack:             bundles.Spec.VersionsBundles[0].CloudStack,
		Docker:                 bundles.Spec.VersionsBundles[0].Docker,
		Eksa:                   bundles.Spec.VersionsBundles[0].Eksa,
		Flux:                   bundles.Spec.VersionsBundles[0].Flux,
		ExternalEtcdBootstrap:  bundles.Spec.VersionsBundles[0].ExternalEtcdBootstrap,
		ExternalEtcdController: bundles.Spec.VersionsBundles[0].ExternalEtcdController,
		Tinkerbell:             bundles.Spec.VersionsBundles[0].Tinkerbell,
		Snow:                   bundles.Spec.VersionsBundles[0].Snow,
		Nutanix:                bundles.Spec.VersionsBundles[0].Nutanix,
	}

	g.Expect(got).To(Equal(want))
}

func TestGetManagementComponents(t *testing.T) {
	tests := []struct {
		name        string
		clusterSpec *cluster.Spec
		want        *cluster.ManagementComponents
		wantErr     string
	}{
		{
			name: "no management components version annotation",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Annotations = nil
				s.Bundles = &releasev1.Bundles{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bundles-1",
						Namespace: "my-namespace",
					},
					Spec: releasev1.BundlesSpec{
						VersionsBundles: []releasev1.VersionsBundle{
							{
								Eksa: releasev1.EksaBundle{
									Version: "v0.0.0-dev",
								},
							},
						},
					},
				}
				s.EKSARelease = &releasev1.EKSARelease{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "eksa-v0-0-0-dev",
						Namespace: constants.EksaSystemNamespace,
					},
					Spec: releasev1.EKSAReleaseSpec{
						BundlesRef: releasev1.BundlesRef{
							Name:      "bundles-1",
							Namespace: "my-namespace",
						},
					},
				}
			}),
			want: &cluster.ManagementComponents{
				Eksa: releasev1.EksaBundle{
					Version: "v0.0.0-dev",
				},
			},
			wantErr: "",
		},
		{
			name: "no management components version annotation and eksa version",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Annotations = nil
				s.Cluster.Spec.EksaVersion = nil
			}),
			want:    nil,
			wantErr: "either management components version or cluster's EksaVersion need to be set",
		},
		{
			name: "with management components version annotation",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				eksaVersion := test.DevEksaVersion()
				s.Cluster.Spec.EksaVersion = &eksaVersion
				s.Cluster.SetManagementComponentsVersion("v0.0.1-dev")
				s.Bundles = &releasev1.Bundles{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bundles-2",
						Namespace: "my-namespace",
					},
					Spec: releasev1.BundlesSpec{
						VersionsBundles: []releasev1.VersionsBundle{
							{
								Eksa: releasev1.EksaBundle{
									Version: "v0.0.1-dev",
								},
							},
						},
					},
				}
				s.EKSARelease = &releasev1.EKSARelease{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "eksa-v0-0-1-dev",
						Namespace: constants.EksaSystemNamespace,
					},
					Spec: releasev1.EKSAReleaseSpec{
						BundlesRef: releasev1.BundlesRef{
							Name:      "bundles-2",
							Namespace: "my-namespace",
						},
					},
				}
			}),
			want: &cluster.ManagementComponents{
				Eksa: releasev1.EksaBundle{
					Version: "v0.0.1-dev",
				},
			},
			wantErr: "",
		},
		{
			name: "eksarelease not found",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				eksaVersion := test.DevEksaVersion()
				s.Cluster.Spec.EksaVersion = &eksaVersion
				s.Cluster.SetManagementComponentsVersion("v0.0.1-dev")
				s.EKSARelease = &releasev1.EKSARelease{}
			}),
			want:    nil,
			wantErr: "\"eksa-v0-0-1-dev\" not found",
		},
		{
			name: "bundles not found",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				eksaVersion := test.DevEksaVersion()
				s.Cluster.Spec.EksaVersion = &eksaVersion
				s.Cluster.SetManagementComponentsVersion("v0.0.1-dev")
				s.EKSARelease = &releasev1.EKSARelease{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "eksa-v0-0-1-dev",
						Namespace: constants.EksaSystemNamespace,
					},
					Spec: releasev1.EKSAReleaseSpec{
						BundlesRef: releasev1.BundlesRef{
							Name:      "bundles-2",
							Namespace: "my-namespace",
						},
					},
				}
			}),
			want:    nil,
			wantErr: " \"bundles-2\" not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()

			client := test.NewFakeKubeClient(tt.clusterSpec.Cluster, tt.clusterSpec.Bundles, tt.clusterSpec.EKSARelease)

			vb, err := cluster.GetManagementComponents(ctx, client, tt.clusterSpec.Cluster)
			if tt.wantErr != "" {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(vb).To(Equal(tt.want))
			}
		})
	}
}
