package snow

import (
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func addStackedEtcdExtraArgsInKubeadmControlPlane(kcp *controlplanev1.KubeadmControlPlane, externalEtcdConfig *v1alpha1.ExternalEtcdConfiguration) {
	if externalEtcdConfig != nil {
		return
	}

	stackedEtcdExtraArgs := kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local.ExtraArgs
	stackedEtcdExtraArgs["listen-peer-urls"] = "https://0.0.0.0:2380"
	stackedEtcdExtraArgs["listen-client-urls"] = "https://0.0.0.0:2379"
}
