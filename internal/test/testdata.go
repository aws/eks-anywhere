package test

import (
	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// Namespace returns a test namespace struct for unit testing.
func Namespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// EksdRelease returns a test release struct for unit testing.
func EksdRelease() *eksdv1.Release {
	return &eksdv1.Release{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Release",
			APIVersion: "distro.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "eksa-system",
		},
		Spec: eksdv1.ReleaseSpec{
			Number: 1,
		},
		Status: eksdv1.ReleaseStatus{
			Components: []eksdv1.Component{
				{
					Assets: []eksdv1.Asset{
						{
							Name:  "etcd-image",
							Image: &eksdv1.AssetImage{},
						},
						{
							Name:  "node-driver-registrar-image",
							Image: &eksdv1.AssetImage{},
						},
						{
							Name:  "livenessprobe-image",
							Image: &eksdv1.AssetImage{},
						},
						{
							Name:  "external-attacher-image",
							Image: &eksdv1.AssetImage{},
						},
						{
							Name:  "external-provisioner-image",
							Image: &eksdv1.AssetImage{},
						},
						{
							Name:  "pause-image",
							Image: &eksdv1.AssetImage{},
						},
						{
							Name:  "aws-iam-authenticator-image",
							Image: &eksdv1.AssetImage{},
						},
						{
							Name:  "coredns-image",
							Image: &eksdv1.AssetImage{},
						},
						{
							Name: "kube-apiserver-image",
							Image: &eksdv1.AssetImage{
								URI: "public.ecr.aws/eks-distro/kubernetes/kube-apiserver:v1.19.8",
							},
						},
					},
				},
			},
		},
	}
}

// Bundle returns a test bundle struct for unit testing.
func Bundle() *releasev1.Bundles {
	return &releasev1.Bundles{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Bundles",
			APIVersion: releasev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bundles-1",
			Namespace: "default",
		},
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1.20",
					EksD: releasev1.EksDRelease{
						Name:           "test",
						EksDReleaseUrl: "testdata/release.yaml",
						KubeVersion:    "1.20",
					},
					CertManager:            releasev1.CertManagerBundle{},
					ClusterAPI:             releasev1.CoreClusterAPI{},
					Bootstrap:              releasev1.KubeadmBootstrapBundle{},
					ControlPlane:           releasev1.KubeadmControlPlaneBundle{},
					VSphere:                releasev1.VSphereBundle{},
					Docker:                 releasev1.DockerBundle{},
					Eksa:                   releasev1.EksaBundle{},
					Cilium:                 releasev1.CiliumBundle{},
					Kindnetd:               releasev1.KindnetdBundle{},
					Flux:                   releasev1.FluxBundle{},
					BottleRocketBootstrap:  releasev1.BottlerocketBootstrapBundle{},
					BottleRocketAdmin:      releasev1.BottlerocketAdminBundle{},
					ExternalEtcdBootstrap:  releasev1.EtcdadmBootstrapBundle{},
					ExternalEtcdController: releasev1.EtcdadmControllerBundle{},
					Tinkerbell:             releasev1.TinkerbellBundle{},
				},
			},
		},
	}
}
