package clusterapi

import (
	etcdbootstrapv1 "github.com/aws/etcdadm-bootstrap-provider/api/v1beta1"
	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	bootstrapv1beta2 "sigs.k8s.io/cluster-api/api/bootstrap/kubeadm/v1beta2"
	controlplanev1beta2 "sigs.k8s.io/cluster-api/api/controlplane/kubeadm/v1beta2"
	clusterv1beta2 "sigs.k8s.io/cluster-api/api/core/v1beta2"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/common"
)

// SetUbuntuConfigInEtcdCluster sets up the etcd config in EtcdadmCluster.
func SetUbuntuConfigInEtcdCluster(etcd *etcdv1.EtcdadmCluster, versionsBundle *cluster.VersionsBundle, eksaVersion *v1alpha1.EksaVersion) {
	etcd.Spec.EtcdadmConfigSpec.Format = etcdbootstrapv1.Format("cloud-config")
	etcd.Spec.EtcdadmConfigSpec.CloudInitConfig = &etcdbootstrapv1.CloudInitConfig{
		Version:    versionsBundle.KubeDistro.EtcdVersion,
		InstallDir: "/usr/bin",
	}
	etcdURL, _ := common.GetExternalEtcdReleaseURL(eksaVersion, versionsBundle)
	if etcdURL != "" {
		etcd.Spec.EtcdadmConfigSpec.CloudInitConfig.EtcdReleaseURL = etcdURL
	}
}

// SetEtcdConfigInCluster sets up the etcd config in CAPI Cluster.
func setUnstackedEtcdConfigInCluster(cluster *clusterv1beta2.Cluster, unstackedEtcdObject APIObject) {
	cluster.Spec.ManagedExternalEtcdRef = &clusterv1beta2.ContractVersionedObjectReference{
		APIGroup: unstackedEtcdObject.GetObjectKind().GroupVersionKind().Group,
		Kind:     unstackedEtcdObject.GetObjectKind().GroupVersionKind().Kind,
		Name:     unstackedEtcdObject.GetName(),
	}
}

// SetUnstackedEtcdConfigInKubeadmControlPlaneForBottlerocket sets up unstacked etcd configuration in kubeadmControlPlane for bottlerocket.
func SetUnstackedEtcdConfigInKubeadmControlPlaneForBottlerocket(kcp *controlplanev1beta2.KubeadmControlPlane, externalEtcdConfig *v1alpha1.ExternalEtcdConfiguration) {
	if externalEtcdConfig == nil {
		return
	}

	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External = bootstrapv1beta2.ExternalEtcd{
		Endpoints: []string{constants.PlaceholderExternalEtcdEndpoint},
		CAFile:    "/var/lib/kubeadm/pki/etcd/ca.crt",
		CertFile:  "/var/lib/kubeadm/pki/server-etcd-client.crt",
		KeyFile:   "/var/lib/kubeadm/pki/apiserver-etcd-client.key",
	}
}

// SetUnstackedEtcdConfigInKubeadmControlPlaneForUbuntu sets up unstacked etcd configuration in kubeadmControlPlane for ubuntu.
func SetUnstackedEtcdConfigInKubeadmControlPlaneForUbuntu(kcp *controlplanev1beta2.KubeadmControlPlane, externalEtcdConfig *v1alpha1.ExternalEtcdConfiguration) {
	if externalEtcdConfig == nil {
		return
	}

	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External = bootstrapv1beta2.ExternalEtcd{
		Endpoints: []string{constants.PlaceholderExternalEtcdEndpoint},
		CAFile:    "/etc/kubernetes/pki/etcd/ca.crt",
		CertFile:  "/etc/kubernetes/pki/apiserver-etcd-client.crt",
		KeyFile:   "/etc/kubernetes/pki/apiserver-etcd-client.key",
	}
}

// setStackedEtcdConfigInKubeadmControlPlane sets up stacked etcd configuration in kubeadmControlPlane.
func setStackedEtcdConfigInKubeadmControlPlane(kcp *controlplanev1beta2.KubeadmControlPlane, etcd cluster.VersionedRepository) {
	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local = bootstrapv1beta2.LocalEtcd{
		ImageRepository: etcd.Repository,
		ImageTag:        etcd.Tag,
		ExtraArgs:       SecureEtcdTlsCipherSuitesExtraArgs().ToArgs(),
	}
}
