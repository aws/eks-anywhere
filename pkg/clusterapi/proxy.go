package clusterapi

import (
	_ "embed"
	"fmt"

	etcdbootstrapv1 "github.com/aws/etcdadm-bootstrap-provider/api/v1beta1"
	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/templater"
)

//go:embed config/http-proxy.conf
var proxyConfig string

func proxy(cluster *v1alpha1.Cluster) bootstrapv1.ProxyConfiguration {
	return bootstrapv1.ProxyConfiguration{
		HTTPSProxy: cluster.Spec.ProxyConfiguration.HttpsProxy,
		NoProxy:    noProxyList(cluster),
	}
}

// SetProxyConfigInKubeadmControlPlaneForBottlerocket sets up proxy configuration in kubeadmControlPlane for bottlerocket.
func SetProxyConfigInKubeadmControlPlaneForBottlerocket(kcp *controlplanev1.KubeadmControlPlane, cluster *v1alpha1.Cluster) {
	if cluster.Spec.ProxyConfiguration == nil {
		return
	}

	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Proxy = proxy(cluster)
	kcp.Spec.KubeadmConfigSpec.JoinConfiguration.Proxy = proxy(cluster)
}

// SetProxyConfigInKubeadmControlPlaneForUbuntu sets up proxy configuration in kubeadmControlPlane for ubuntu.
func SetProxyConfigInKubeadmControlPlaneForUbuntu(kcp *controlplanev1.KubeadmControlPlane, cluster *v1alpha1.Cluster) error {
	if cluster.Spec.ProxyConfiguration == nil {
		return nil
	}

	return addProxyConfigInKubeadmConfigSpecFiles(&kcp.Spec.KubeadmConfigSpec, cluster)
}

// SetProxyConfigInKubeadmConfigTemplateForBottlerocket sets up proxy configuration in kubeadmConfigTemplate for bottlerocket.
func SetProxyConfigInKubeadmConfigTemplateForBottlerocket(kct *bootstrapv1.KubeadmConfigTemplate, cluster *v1alpha1.Cluster) {
	if cluster.Spec.ProxyConfiguration == nil {
		return
	}

	kct.Spec.Template.Spec.JoinConfiguration.Proxy = proxy(cluster)
}

// SetProxyConfigInKubeadmConfigTemplateForUbuntu sets up proxy configuration in kubeadmConfigTemplate for ubuntu.
func SetProxyConfigInKubeadmConfigTemplateForUbuntu(kct *bootstrapv1.KubeadmConfigTemplate, cluster *v1alpha1.Cluster) error {
	if cluster.Spec.ProxyConfiguration == nil {
		return nil
	}

	return addProxyConfigInKubeadmConfigSpecFiles(&kct.Spec.Template.Spec, cluster)
}

// setProxyConfigInEtcdCluster sets up proxy configuration in etcdadmCluster.
func setProxyConfigInEtcdCluster(etcd *etcdv1.EtcdadmCluster, cluster *v1alpha1.Cluster) {
	if cluster.Spec.ProxyConfiguration == nil {
		return
	}

	etcd.Spec.EtcdadmConfigSpec.Proxy = &etcdbootstrapv1.ProxyConfiguration{
		HTTPProxy:  cluster.Spec.ProxyConfiguration.HttpProxy,
		HTTPSProxy: cluster.Spec.ProxyConfiguration.HttpsProxy,
		NoProxy:    noProxyList(cluster),
	}
}

func NoProxyDefaults() []string {
	return []string{
		"localhost",
		"127.0.0.1",
		".svc",
	}
}

func noProxyList(cluster *v1alpha1.Cluster) []string {
	capacity := len(cluster.Spec.ClusterNetwork.Pods.CidrBlocks) +
		len(cluster.Spec.ClusterNetwork.Services.CidrBlocks) +
		len(cluster.Spec.ProxyConfiguration.NoProxy) + 4

	noProxyList := make([]string, 0, capacity)
	noProxyList = append(noProxyList, cluster.Spec.ClusterNetwork.Pods.CidrBlocks...)
	noProxyList = append(noProxyList, cluster.Spec.ClusterNetwork.Services.CidrBlocks...)
	noProxyList = append(noProxyList, cluster.Spec.ProxyConfiguration.NoProxy...)

	// Add no-proxy defaults
	noProxyList = append(noProxyList, NoProxyDefaults()...)
	noProxyList = append(noProxyList, cluster.Spec.ControlPlaneConfiguration.Endpoint.Host)

	return noProxyList
}

func proxyConfigContent(cluster *v1alpha1.Cluster) (string, error) {
	val := values{
		"httpProxy":  cluster.Spec.ProxyConfiguration.HttpProxy,
		"httpsProxy": cluster.Spec.ProxyConfiguration.HttpsProxy,
		"noProxy":    noProxyList(cluster),
	}

	config, err := templater.Execute(proxyConfig, val)
	if err != nil {
		return "", fmt.Errorf("building http-proxy.conf file: %v", err)
	}
	return string(config), nil
}

func proxyConfigFile(cluster *v1alpha1.Cluster) (bootstrapv1.File, error) {
	proxyConfig, err := proxyConfigContent(cluster)
	if err != nil {
		return bootstrapv1.File{}, err
	}

	return bootstrapv1.File{
		Path:    "/etc/systemd/system/containerd.service.d/http-proxy.conf",
		Owner:   "root:root",
		Content: proxyConfig,
	}, nil
}

func addProxyConfigInKubeadmConfigSpecFiles(kcs *bootstrapv1.KubeadmConfigSpec, cluster *v1alpha1.Cluster) error {
	proxyConfigFile, err := proxyConfigFile(cluster)
	if err != nil {
		return err
	}

	kcs.Files = append(kcs.Files, proxyConfigFile)

	return nil
}
