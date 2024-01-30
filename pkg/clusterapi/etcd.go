package clusterapi

import (
	etcdbootstrapv1 "github.com/aws/etcdadm-bootstrap-provider/api/v1beta1"
	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/common"
)

// SetUbuntuConfigInEtcdCluster sets up the etcd config in EtcdadmCluster.
func SetUbuntuConfigInEtcdCluster(etcd *etcdv1.EtcdadmCluster, versionsBundle *cluster.VersionsBundle, eksaVersion string) {
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
func setUnstackedEtcdConfigInCluster(cluster *clusterv1.Cluster, unstackedEtcdObject APIObject) {
	cluster.Spec.ManagedExternalEtcdRef = &v1.ObjectReference{
		APIVersion: unstackedEtcdObject.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:       unstackedEtcdObject.GetObjectKind().GroupVersionKind().Kind,
		Name:       unstackedEtcdObject.GetName(),
		Namespace:  constants.EksaSystemNamespace,
	}
}

// SetUnstackedEtcdConfigInKubeadmControlPlaneForBottlerocket sets up unstacked etcd configuration in kubeadmControlPlane for bottlerocket.
func SetUnstackedEtcdConfigInKubeadmControlPlaneForBottlerocket(kcp *controlplanev1.KubeadmControlPlane, externalEtcdConfig *v1alpha1.ExternalEtcdConfiguration) {
	if externalEtcdConfig == nil {
		return
	}

	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External = &bootstrapv1.ExternalEtcd{
		Endpoints: []string{},
		CAFile:    "/var/lib/kubeadm/pki/etcd/ca.crt",
		CertFile:  "/var/lib/kubeadm/pki/server-etcd-client.crt",
		KeyFile:   "/var/lib/kubeadm/pki/apiserver-etcd-client.key",
	}
}

// SetUnstackedEtcdConfigInKubeadmControlPlaneForUbuntu sets up unstacked etcd configuration in kubeadmControlPlane for ubuntu.
func SetUnstackedEtcdConfigInKubeadmControlPlaneForUbuntu(kcp *controlplanev1.KubeadmControlPlane, externalEtcdConfig *v1alpha1.ExternalEtcdConfiguration) {
	if externalEtcdConfig == nil {
		return
	}

	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External = &bootstrapv1.ExternalEtcd{
		Endpoints: []string{},
		CAFile:    "/etc/kubernetes/pki/etcd/ca.crt",
		CertFile:  "/etc/kubernetes/pki/apiserver-etcd-client.crt",
		KeyFile:   "/etc/kubernetes/pki/apiserver-etcd-client.key",
	}
}

// setStackedEtcdConfigInKubeadmControlPlane sets up stacked etcd configuration in kubeadmControlPlane.
func setStackedEtcdConfigInKubeadmControlPlane(kcp *controlplanev1.KubeadmControlPlane, etcd cluster.VersionedRepository) {
	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local = &bootstrapv1.LocalEtcd{
		ImageMeta: bootstrapv1.ImageMeta{
			ImageRepository: etcd.Repository,
			ImageTag:        etcd.Tag,
		},
		ExtraArgs: SecureEtcdTlsCipherSuitesExtraArgs(),
	}
}
