package clusterapi

import (
	_ "embed"
	"fmt"

	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/templater"
)

//go:embed config/http-proxy.conf
var proxyConfig string

func NoProxyDefaults() []string {
	return []string{
		"localhost",
		"127.0.0.1",
		".svc",
	}
}

func proxyConfigContent(cluster v1alpha1.ClusterSpec) (string, error) {
	capacity := len(cluster.ClusterNetwork.Pods.CidrBlocks) +
		len(cluster.ClusterNetwork.Services.CidrBlocks) +
		len(cluster.ProxyConfiguration.NoProxy) + 4

	noProxyList := make([]string, 0, capacity)
	noProxyList = append(noProxyList, cluster.ClusterNetwork.Pods.CidrBlocks...)
	noProxyList = append(noProxyList, cluster.ClusterNetwork.Services.CidrBlocks...)
	noProxyList = append(noProxyList, cluster.ProxyConfiguration.NoProxy...)

	// Add no-proxy defaults
	noProxyList = append(noProxyList, NoProxyDefaults()...)
	noProxyList = append(noProxyList, cluster.ControlPlaneConfiguration.Endpoint.Host)

	val := values{
		"httpProxy":  cluster.ProxyConfiguration.HttpProxy,
		"httpsProxy": cluster.ProxyConfiguration.HttpsProxy,
		"noProxy":    noProxyList,
	}

	config, err := templater.Execute(proxyConfig, val)
	if err != nil {
		return "", fmt.Errorf("building http-proxy.conf file: %v", err)
	}
	return string(config), nil
}

func proxyConfigFile(cluster v1alpha1.ClusterSpec) (bootstrapv1.File, error) {
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

func SetProxyConfigInKubeadmControlPlane(kcp *controlplanev1.KubeadmControlPlane, cluster v1alpha1.ClusterSpec) error {
	if cluster.ProxyConfiguration == nil {
		return nil
	}

	proxyConfigFile, err := proxyConfigFile(cluster)
	if err != nil {
		return err
	}

	kcp.Spec.KubeadmConfigSpec.Files = append(kcp.Spec.KubeadmConfigSpec.Files, proxyConfigFile)

	return nil
}

func SetProxyConfigInKubeadmConfigTemplate(kct *bootstrapv1.KubeadmConfigTemplate, cluster v1alpha1.ClusterSpec) error {
	if cluster.ProxyConfiguration == nil {
		return nil
	}

	proxyConfigFile, err := proxyConfigFile(cluster)
	if err != nil {
		return err
	}

	kct.Spec.Template.Spec.Files = append(kct.Spec.Template.Spec.Files, proxyConfigFile)

	return nil
}
