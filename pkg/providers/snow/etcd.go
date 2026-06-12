package snow

import (
	bootstrapv1beta2 "sigs.k8s.io/cluster-api/api/bootstrap/kubeadm/v1beta2"
	controlplanev1beta2 "sigs.k8s.io/cluster-api/api/controlplane/kubeadm/v1beta2"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// FgEtcdLearner is a Kubeadm feature gate for etcd learner mode.
const FgEtcdLearner = "EtcdLearnerMode"

func addStackedEtcdExtraArgsInKubeadmControlPlane(kcp *controlplanev1beta2.KubeadmControlPlane, externalEtcdConfig *v1alpha1.ExternalEtcdConfiguration) {
	if externalEtcdConfig != nil {
		return
	}

	listenPeerURLs := "https://0.0.0.0:2380"
	listenClientURLs := "https://0.0.0.0:2379"
	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local.ExtraArgs = append(
		kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local.ExtraArgs,
		bootstrapv1beta2.Arg{Name: "listen-peer-urls", Value: &listenPeerURLs},
		bootstrapv1beta2.Arg{Name: "listen-client-urls", Value: &listenClientURLs},
	)
}

func disableEtcdLearnerMode(kcp *controlplanev1beta2.KubeadmControlPlane) {
	if kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.FeatureGates == nil {
		kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.FeatureGates = map[string]bool{}
	}

	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.FeatureGates[FgEtcdLearner] = false
}
