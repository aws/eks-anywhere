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

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/cluster/mocks"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestBuildSpecForCluster(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	oidcConfig := &anywherev1.OIDCConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "myconfig",
		},
	}
	awsIamConfig := &anywherev1.AWSIamConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "myconfig",
		},
	}
	fluxConfig := &anywherev1.FluxConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "myconfig",
		},
	}
	config := &cluster.Config{
		Cluster: &anywherev1.Cluster{
			Spec: anywherev1.ClusterSpec{
				KubernetesVersion: anywherev1.Kube124,
				IdentityProviderRefs: []anywherev1.Ref{
					{
						Kind: anywherev1.OIDCConfigKind,
						Name: "myconfig",
					},
					{
						Kind: anywherev1.AWSIamConfigKind,
						Name: "myconfig",
					},
				},
				GitOpsRef: &anywherev1.Ref{
					Kind: anywherev1.FluxConfigKind,
					Name: "myconfig",
				},
			},
		},
		OIDCConfigs: map[string]*anywherev1.OIDCConfig{
			"myconfig": oidcConfig,
		},
		AWSIAMConfigs: map[string]*anywherev1.AWSIamConfig{
			"myconfig": awsIamConfig,
		},
		FluxConfig: fluxConfig,
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

	bundlesFetch := func(_ context.Context, _, _ string) (*releasev1.Bundles, error) {
		return bundles, nil
	}
	eksdFetch := func(_ context.Context, _, _ string) (*eksdv1.Release, error) {
		return eksdRelease, nil
	}
	gitOpsFetch := func(_ context.Context, name, namespace string) (*anywherev1.GitOpsConfig, error) {
		return nil, nil
	}
	fluxConfigFetch := func(_ context.Context, _, _ string) (*anywherev1.FluxConfig, error) {
		return fluxConfig, nil
	}
	oidcFetch := func(_ context.Context, _, _ string) (*anywherev1.OIDCConfig, error) {
		return oidcConfig, nil
	}
	awsIamConfigFetch := func(_ context.Context, _, _ string) (*anywherev1.AWSIamConfig, error) {
		return awsIamConfig, nil
	}

	spec, err := cluster.BuildSpecForCluster(ctx, config.Cluster, bundlesFetch, eksdFetch, gitOpsFetch, fluxConfigFetch, oidcFetch, awsIamConfigFetch)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(spec.Config).To(Equal(config))
	g.Expect(spec.OIDCConfig).To(Equal(oidcConfig))
	g.Expect(spec.AWSIamConfig).To(Equal(awsIamConfig))
	g.Expect(spec.Bundles).To(Equal(bundles))
}

func TestGetBundlesForCluster(t *testing.T) {
	testCases := []struct {
		testName                string
		cluster                 *anywherev1.Cluster
		wantName, wantNamespace string
	}{
		{
			testName: "no bundles ref",
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eksa-cluster",
					Namespace: "eksa",
				},
				Spec: anywherev1.ClusterSpec{},
			},
			wantName:      "eksa-cluster",
			wantNamespace: "eksa",
		},
		{
			testName: "bundles ref",
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eksa-cluster",
					Namespace: "eksa",
				},
				Spec: anywherev1.ClusterSpec{
					BundlesRef: &anywherev1.BundlesRef{
						Name:       "bundles-1",
						Namespace:  "eksa-system",
						APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
					},
				},
			},
			wantName:      "bundles-1",
			wantNamespace: "eksa-system",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			wantBundles := &releasev1.Bundles{}
			mockFetch := func(ctx context.Context, name, namespace string) (*releasev1.Bundles, error) {
				g.Expect(name).To(Equal(tt.wantName))
				g.Expect(namespace).To(Equal(tt.wantNamespace))

				return wantBundles, nil
			}

			gotBundles, err := cluster.GetBundlesForCluster(context.Background(), tt.cluster, mockFetch)
			g.Expect(err).To(BeNil())
			g.Expect(gotBundles).To(Equal(wantBundles))
		})
	}
}

func TestGetFluxConfigForClusterIsNil(t *testing.T) {
	g := NewWithT(t)
	c := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "eksa-cluster",
			Namespace: "eksa",
		},
		Spec: anywherev1.ClusterSpec{
			GitOpsRef: nil,
		},
	}
	var wantFlux *anywherev1.FluxConfig
	mockFetch := func(ctx context.Context, name, namespace string) (*anywherev1.FluxConfig, error) {
		g.Expect(name).To(Equal(c.Name))
		g.Expect(namespace).To(Equal(c.Namespace))

		return wantFlux, nil
	}

	gotFlux, err := cluster.GetFluxConfigForCluster(context.Background(), c, mockFetch)
	g.Expect(err).To(BeNil())
	g.Expect(gotFlux).To(Equal(wantFlux))
}

func TestGetFluxConfigForCluster(t *testing.T) {
	g := NewWithT(t)
	c := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "eksa-cluster",
			Namespace: "eksa",
		},
		Spec: anywherev1.ClusterSpec{
			GitOpsRef: &anywherev1.Ref{
				Kind: anywherev1.FluxConfigKind,
				Name: "eksa-cluster",
			},
		},
	}
	wantFlux := &anywherev1.FluxConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       anywherev1.FluxConfigSpec{},
		Status:     anywherev1.FluxConfigStatus{},
	}
	mockFetch := func(ctx context.Context, name, namespace string) (*anywherev1.FluxConfig, error) {
		g.Expect(name).To(Equal(c.Name))
		g.Expect(namespace).To(Equal(c.Namespace))

		return wantFlux, nil
	}

	gotFlux, err := cluster.GetFluxConfigForCluster(context.Background(), c, mockFetch)
	g.Expect(err).To(BeNil())
	g.Expect(gotFlux).To(Equal(wantFlux))
}

func TestGetAWSIamConfigForCluster(t *testing.T) {
	g := NewWithT(t)
	c := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "eksa-cluster",
			Namespace: "eksa",
		},
		Spec: anywherev1.ClusterSpec{
			IdentityProviderRefs: []anywherev1.Ref{
				{
					Kind: anywherev1.AWSIamConfigKind,
					Name: "eksa-cluster",
				},
			},
		},
	}
	wantIamConfig := &anywherev1.AWSIamConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       anywherev1.AWSIamConfigSpec{},
		Status:     anywherev1.AWSIamConfigStatus{},
	}

	mockFetch := func(ctx context.Context, name, namespace string) (*anywherev1.AWSIamConfig, error) {
		g.Expect(name).To(Equal(c.Name))
		g.Expect(namespace).To(Equal(c.Namespace))

		return wantIamConfig, nil
	}

	gotIamConfig, err := cluster.GetAWSIamConfigForCluster(context.Background(), c, mockFetch)
	g.Expect(err).To(BeNil())
	g.Expect(gotIamConfig).To(Equal(wantIamConfig))
}

func TestGetAWSIamConfigForClusterIsNil(t *testing.T) {
	g := NewWithT(t)
	c := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "eksa-cluster",
			Namespace: "eksa",
		},
		Spec: anywherev1.ClusterSpec{
			IdentityProviderRefs: nil,
		},
	}
	var wantIamConfig *anywherev1.AWSIamConfig
	mockFetch := func(ctx context.Context, name, namespace string) (*anywherev1.AWSIamConfig, error) {
		g.Expect(name).To(Equal(c.Name))
		g.Expect(namespace).To(Equal(c.Namespace))

		return wantIamConfig, nil
	}

	gotIamConfig, err := cluster.GetAWSIamConfigForCluster(context.Background(), c, mockFetch)
	g.Expect(err).To(BeNil())
	g.Expect(gotIamConfig).To(Equal(wantIamConfig))
}

type buildSpecTest struct {
	*WithT
	ctx         context.Context
	ctrl        *gomock.Controller
	client      *mocks.MockClient
	cluster     *anywherev1.Cluster
	bundles     *releasev1.Bundles
	eksdRelease *eksdv1.Release
	kubeDistro  *cluster.KubeDistro
}

func newBuildSpecTest(t *testing.T) *buildSpecTest {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)
	cluster := &anywherev1.Cluster{
		Spec: anywherev1.ClusterSpec{
			BundlesRef: &anywherev1.BundlesRef{
				Name:      "bundles-1",
				Namespace: "my-namespace",
			},
			KubernetesVersion: anywherev1.Kube123,
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
			},
		},
	}
	eksdRelease, kubeDistro := wantKubeDistroForEksdRelease()

	return &buildSpecTest{
		WithT:       NewWithT(t),
		ctx:         context.Background(),
		ctrl:        ctrl,
		client:      client,
		cluster:     cluster,
		bundles:     bundles,
		eksdRelease: eksdRelease,
		kubeDistro:  kubeDistro,
	}
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
			o.Status = tt.eksdRelease.Status
			return nil
		},
	)
}

func TestBuildSpec(t *testing.T) {
	tt := newBuildSpecTest(t)
	tt.expectGetBundles()
	tt.expectGetEksd()

	wantSpec := &cluster.Spec{
		Config: &cluster.Config{
			Cluster:       tt.cluster,
			OIDCConfigs:   map[string]*anywherev1.OIDCConfig{},
			AWSIAMConfigs: map[string]*anywherev1.AWSIamConfig{},
		},
		VersionsBundle: &cluster.VersionsBundle{
			VersionsBundle: &tt.bundles.Spec.VersionsBundles[0],
			KubeDistro:     tt.kubeDistro,
		},
		Bundles: tt.bundles,
	}

	spec, err := cluster.BuildSpec(tt.ctx, tt.client, tt.cluster)
	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(spec.Config).To(Equal(wantSpec.Config))
	tt.Expect(spec.AWSIamConfig).To(Equal(wantSpec.AWSIamConfig))
	tt.Expect(spec.OIDCConfig).To(Equal(wantSpec.OIDCConfig))
	tt.Expect(spec.Bundles).To(Equal(wantSpec.Bundles))
	tt.Expect(spec.VersionsBundle).To(Equal(wantSpec.VersionsBundle))
}

func TestBuildSpecGetBundlesError(t *testing.T) {
	tt := newBuildSpecTest(t)
	tt.client.EXPECT().Get(tt.ctx, "bundles-1", "my-namespace", &releasev1.Bundles{}).Return(errors.New("client error"))

	_, err := cluster.BuildSpec(tt.ctx, tt.client, tt.cluster)
	tt.Expect(err).To(MatchError(ContainSubstring("client error")))
}

func TestBuildSpecGetEksdError(t *testing.T) {
	tt := newBuildSpecTest(t)
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

func TestBuildSpecUnsupportedKubernetesVersionError(t *testing.T) {
	tt := newBuildSpecTest(t)
	tt.bundles.Spec.VersionsBundles = []releasev1.VersionsBundle{}
	tt.bundles.Spec.Number = 2
	tt.expectGetBundles()

	_, err := cluster.BuildSpec(tt.ctx, tt.client, tt.cluster)
	tt.Expect(err).To(MatchError(ContainSubstring("kubernetes version 1.23 is not supported by bundles manifest 2")))
}

func TestBuildSpecInitError(t *testing.T) {
	tt := newBuildSpecTest(t)
	tt.eksdRelease.Status.Components = []eksdv1.Component{}
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
		Status: eksdv1.ReleaseStatus{
			Components: []eksdv1.Component{
				{
					Name:   "etcd",
					GitTag: "v3.4.14",
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
	}

	return eksdRelease, kubeDistro
}
