package test

import (
	"time"

	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
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
					KubeVersion: "1.22",
					EksD: releasev1.EksDRelease{
						Name:           "test",
						EksDReleaseUrl: "testdata/release.yaml",
						KubeVersion:    "1.22",
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
