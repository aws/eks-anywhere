package snow

import (
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// FgEtcdLearner is a Kubeadm feature gate for etcd learner mode.
const FgEtcdLearner = "EtcdLearnerMode"

func addStackedEtcdExtraArgsInKubeadmControlPlane(kcp *controlplanev1.KubeadmControlPlane, externalEtcdConfig *v1alpha1.ExternalEtcdConfiguration) {
	if externalEtcdConfig != nil {
		return
	}

	stackedEtcdExtraArgs := kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local.ExtraArgs
	stackedEtcdExtraArgs["listen-peer-urls"] = "https://0.0.0.0:2380"
	stackedEtcdExtraArgs["listen-client-urls"] = "https://0.0.0.0:2379"
}

func disableEtcdLearnerMode(kcp *controlplanev1.KubeadmControlPlane) {
	if kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.FeatureGates == nil {
		kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.FeatureGates = map[string]bool{}
	}

	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.FeatureGates[FgEtcdLearner] = false
}
