package test

import (
	"time"

	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
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

// EKSARelease returns a test eksaRelease struct for unit testing.
func EKSARelease() *releasev1.EKSARelease {
	return &releasev1.EKSARelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       releasev1.EKSAReleaseKind,
			APIVersion: releasev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "eksa-v0-0-0-dev",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: releasev1.EKSAReleaseSpec{
			ReleaseDate:       "",
			Version:           "v0.0.0-dev",
			GitCommit:         "",
			BundleManifestURL: "",
			BundlesRef: releasev1.BundlesRef{
				Name:       "bundles-1",
				Namespace:  "default",
				APIVersion: releasev1.GroupVersion.String(),
			},
		},
	}
}

// EksdReleases returns a test release slice for unit testing.
func EksdReleases() []eksdv1.Release {
	return []eksdv1.Release{
		*EksdRelease("1-19"),
		*EksdRelease("1-22"),
		*EksdRelease("1-24"),
	}
}

// VersionsBundlesMap returns a test VersionsBundle map for unit testing.
func VersionsBundlesMap() map[anywherev1.KubernetesVersion]*cluster.VersionsBundle {
	return map[anywherev1.KubernetesVersion]*cluster.VersionsBundle{
		anywherev1.Kube118: VersionBundle(),
		anywherev1.Kube119: VersionBundle(),
		anywherev1.Kube120: VersionBundle(),
		anywherev1.Kube121: VersionBundle(),
		anywherev1.Kube122: VersionBundle(),
		anywherev1.Kube123: VersionBundle(),
		anywherev1.Kube124: VersionBundle(),
	}
}

// VersionBundle returns a test VersionsBundle struct for unit testing.
func VersionBundle() *cluster.VersionsBundle {
	return &cluster.VersionsBundle{
		VersionsBundle: &releasev1.VersionsBundle{
			Eksa: releasev1.EksaBundle{
				DiagnosticCollector: releasev1.Image{
					URI: "public.ecr.aws/eks-anywhere/diagnostic-collector:v0.9.1-eks-a-10",
				},
			},
		},
		KubeDistro: &cluster.KubeDistro{
			AwsIamAuthImage: releasev1.Image{
				URI: "public.ecr.aws/eks-distro/kubernetes-sigs/aws-iam-authenticator:v0.5.2-eks-1-18-11",
			},
		},
	}
}

// EksdRelease returns a test release struct for unit testing.
func EksdRelease(channel string) *eksdv1.Release {
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
			Number:  1,
			Channel: channel,
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
						{
							Name: "kube-proxy-image",
							Image: &eksdv1.AssetImage{
								URI: "public.ecr.aws/eks-distro/kubernetes/kube-proxy:v1.19.8-eks-1-19-18",
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
					KubeVersion: "1.19",
					EksD: releasev1.EksDRelease{
						Name:           "test",
						EksDReleaseUrl: "embed:///testdata/release.yaml",
						KubeVersion:    "1.19",
					},
					CertManager:                releasev1.CertManagerBundle{},
					ClusterAPI:                 releasev1.CoreClusterAPI{},
					Bootstrap:                  releasev1.KubeadmBootstrapBundle{},
					ControlPlane:               releasev1.KubeadmControlPlaneBundle{},
					VSphere:                    releasev1.VSphereBundle{},
					Docker:                     releasev1.DockerBundle{},
					Eksa:                       releasev1.EksaBundle{},
					Cilium:                     releasev1.CiliumBundle{},
					Kindnetd:                   releasev1.KindnetdBundle{},
					Flux:                       releasev1.FluxBundle{},
					BottleRocketHostContainers: releasev1.BottlerocketHostContainersBundle{},
					ExternalEtcdBootstrap:      releasev1.EtcdadmBootstrapBundle{},
					ExternalEtcdController:     releasev1.EtcdadmControllerBundle{},
					Tinkerbell:                 releasev1.TinkerbellBundle{},
				},
				{
					KubeVersion: "1.22",
					EksD: releasev1.EksDRelease{
						Name:           "test",
						EksDReleaseUrl: "embed:///testdata/release.yaml",
						KubeVersion:    "1.22",
						Raw: releasev1.OSImageBundle{
							Bottlerocket: releasev1.Archive{
								URI: "http://tinkerbell-example:8080/bottlerocket-2004-kube-v1.22.5.gz",
							},
						},
					},
					CertManager:                releasev1.CertManagerBundle{},
					ClusterAPI:                 releasev1.CoreClusterAPI{},
					Bootstrap:                  releasev1.KubeadmBootstrapBundle{},
					ControlPlane:               releasev1.KubeadmControlPlaneBundle{},
					VSphere:                    releasev1.VSphereBundle{},
					Docker:                     releasev1.DockerBundle{},
					Eksa:                       releasev1.EksaBundle{},
					Cilium:                     releasev1.CiliumBundle{},
					Kindnetd:                   releasev1.KindnetdBundle{},
					Flux:                       releasev1.FluxBundle{},
					BottleRocketHostContainers: releasev1.BottlerocketHostContainersBundle{},
					ExternalEtcdBootstrap:      releasev1.EtcdadmBootstrapBundle{},
					ExternalEtcdController:     releasev1.EtcdadmControllerBundle{},
					Tinkerbell:                 releasev1.TinkerbellBundle{},
				},
				{
					KubeVersion: "1.24",
					EksD: releasev1.EksDRelease{
						Name:           "test",
						EksDReleaseUrl: "embed:///testdata/release.yaml",
						KubeVersion:    "1.22",
					},
					CertManager:  releasev1.CertManagerBundle{},
					ClusterAPI:   releasev1.CoreClusterAPI{},
					Bootstrap:    releasev1.KubeadmBootstrapBundle{},
					ControlPlane: releasev1.KubeadmControlPlaneBundle{},
					VSphere:      releasev1.VSphereBundle{},
					Docker:       releasev1.DockerBundle{},
					Eksa: releasev1.EksaBundle{
						DiagnosticCollector: releasev1.Image{
							URI: "public.ecr.aws/eks-anywhere/diagnostic-collector:v0.9.1-eks-a-10",
						},
					},
					Cilium:                     releasev1.CiliumBundle{},
					Kindnetd:                   releasev1.KindnetdBundle{},
					Flux:                       releasev1.FluxBundle{},
					BottleRocketHostContainers: releasev1.BottlerocketHostContainersBundle{},
					ExternalEtcdBootstrap:      releasev1.EtcdadmBootstrapBundle{},
					ExternalEtcdController:     releasev1.EtcdadmControllerBundle{},
					Tinkerbell:                 releasev1.TinkerbellBundle{},
				},
			},
		},
	}
}

// CAPIClusterOpt represents an function where a capi cluster is passed as an argument.
type CAPIClusterOpt func(*clusterv1.Cluster)

// CAPICluster returns a capi cluster which can be configured by passing in opts arguments.
func CAPICluster(opts ...CAPIClusterOpt) *clusterv1.Cluster {
	c := &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: clusterv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.EksaSystemNamespace,
		},
		Status: clusterv1.ClusterStatus{
			Conditions: clusterv1.Conditions{
				{
					Type:               clusterapi.ControlPlaneReadyCondition,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(time.Now()),
				},
			},
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// KubeadmControlPlaneOpt represents an function where a kubeadmcontrolplane is passed as an argument.
type KubeadmControlPlaneOpt func(kcp *controlplanev1.KubeadmControlPlane)

// KubeadmControlPlane returns a kubeadm controlplane which can be configured by passing in opts arguments.
func KubeadmControlPlane(opts ...KubeadmControlPlaneOpt) *controlplanev1.KubeadmControlPlane {
	kcp := &controlplanev1.KubeadmControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmControlPlane",
			APIVersion: controlplanev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.EksaSystemNamespace,
		},
	}

	for _, opt := range opts {
		opt(kcp)
	}

	return kcp
}

// MachineDeploymentOpt represents an function where a kubeadmcontrolplane is passed as an argument.
type MachineDeploymentOpt func(md *clusterv1.MachineDeployment)

// MachineDeployment returns a machinedeployment which can be configured by passing in opts arguments.
func MachineDeployment(opts ...MachineDeploymentOpt) *clusterv1.MachineDeployment {
	md := &clusterv1.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MachineDeployment",
			APIVersion: clusterv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.EksaSystemNamespace,
		},
	}

	for _, opt := range opts {
		opt(md)
	}

	return md
}
