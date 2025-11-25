package clusterapi

import (
	_ "embed"
	"fmt"

	etcdbootstrapv1 "github.com/aws/etcdadm-bootstrap-provider/api/v1beta1"
	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"github.com/aws/eks-anywhere/pkg/registrymirror/containerd"
	"github.com/aws/eks-anywhere/pkg/templater"
)

//go:embed config/containerd_config_append.toml
var containerdConfig string

// SetRegistryMirrorInKubeadmControlPlaneForBottlerocket sets up registry mirror configuration in kubeadmControlPlane for bottlerocket.
func SetRegistryMirrorInKubeadmControlPlaneForBottlerocket(kcp *controlplanev1.KubeadmControlPlane, mirrorConfig *v1alpha1.RegistryMirrorConfiguration) {
	if mirrorConfig == nil {
		return
	}

	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.RegistryMirror = registryMirror(mirrorConfig)
	kcp.Spec.KubeadmConfigSpec.JoinConfiguration.RegistryMirror = registryMirror(mirrorConfig)
}

// SetRegistryMirrorInKubeadmControlPlaneForUbuntu sets up registry mirror configuration in kubeadmControlPlane for ubuntu.
func SetRegistryMirrorInKubeadmControlPlaneForUbuntu(kcp *controlplanev1.KubeadmControlPlane, mirrorConfig *v1alpha1.RegistryMirrorConfiguration) error {
	if mirrorConfig == nil {
		return nil
	}

	return addRegistryMirrorInKubeadmConfigSpecFiles(&kcp.Spec.KubeadmConfigSpec, mirrorConfig)
}

// SetRegistryMirrorInKubeadmConfigTemplateForBottlerocket sets up registry mirror configuration in kubeadmConfigTemplate for bottlerocket.
func SetRegistryMirrorInKubeadmConfigTemplateForBottlerocket(kct *bootstrapv1.KubeadmConfigTemplate, mirrorConfig *v1alpha1.RegistryMirrorConfiguration) {
	if mirrorConfig == nil {
		return
	}

	kct.Spec.Template.Spec.JoinConfiguration.RegistryMirror = registryMirror(mirrorConfig)
}

// SetRegistryMirrorInKubeadmConfigTemplateForUbuntu sets up registry mirror configuration in kubeadmConfigTemplate for ubuntu.
func SetRegistryMirrorInKubeadmConfigTemplateForUbuntu(kct *bootstrapv1.KubeadmConfigTemplate, mirrorConfig *v1alpha1.RegistryMirrorConfiguration) error {
	if mirrorConfig == nil {
		return nil
	}

	return addRegistryMirrorInKubeadmConfigSpecFiles(&kct.Spec.Template.Spec, mirrorConfig)
}

// setRegistryMirrorInEtcdCluster sets up registry mirror configuration in etcdadmCluster.
func setRegistryMirrorInEtcdCluster(etcd *etcdv1.EtcdadmCluster, mirrorConfig *v1alpha1.RegistryMirrorConfiguration) {
	if mirrorConfig == nil {
		return
	}

	etcd.Spec.EtcdadmConfigSpec.RegistryMirror = &etcdbootstrapv1.RegistryMirrorConfiguration{
		Endpoint: containerd.ToAPIEndpoint(registrymirror.FromClusterRegistryMirrorConfiguration(mirrorConfig).CoreEKSAMirror()),
		CACert:   mirrorConfig.CACertContent,
	}
}

func registryMirror(mirrorConfig *v1alpha1.RegistryMirrorConfiguration) bootstrapv1.RegistryMirrorConfiguration {
	return bootstrapv1.RegistryMirrorConfiguration{
		Endpoint: containerd.ToAPIEndpoint(registrymirror.FromClusterRegistryMirrorConfiguration(mirrorConfig).CoreEKSAMirror()),
		CACert:   mirrorConfig.CACertContent,
	}
}

type values map[string]interface{}

func registryMirrorConfigContent(registryMirror *registrymirror.RegistryMirror) (string, error) {
	val := values{
		"registryMirrorMap": containerd.ToAPIEndpoints(registryMirror.NamespacedRegistryMap),
		"mirrorBase":        registryMirror.BaseRegistry,
		"registryCACert":    registryMirror.CACertContent,
		"insecureSkip":      registryMirror.InsecureSkipVerify,
	}

	config, err := templater.Execute(containerdConfig, val)
	if err != nil {
		return "", fmt.Errorf("building containerd config file: %v", err)
	}
	return string(config), nil
}

func registryMirrorConfig(registryMirrorConfig *v1alpha1.RegistryMirrorConfiguration) (files []bootstrapv1.File, err error) {
	registryMirror := registrymirror.FromClusterRegistryMirrorConfiguration(registryMirrorConfig)
	registryConfig, err := registryMirrorConfigContent(registryMirror)
	if err != nil {
		return nil, err
	}
	files = []bootstrapv1.File{
		{
			Path:    "/etc/containerd/config_append.toml",
			Owner:   "root:root",
			Content: registryConfig,
		},
	}

	if registryMirrorConfig.CACertContent != "" {
		files = append(files, bootstrapv1.File{
			Path:    fmt.Sprintf("/etc/containerd/certs.d/%s/ca.crt", registryMirror.BaseRegistry),
			Owner:   "root:root",
			Content: registryMirrorConfig.CACertContent,
		})
	}

	return files, nil
}

func addRegistryMirrorInKubeadmConfigSpecFiles(kcs *bootstrapv1.KubeadmConfigSpec, mirrorConfig *v1alpha1.RegistryMirrorConfiguration) error {
	containerdFiles, err := registryMirrorConfig(mirrorConfig)
	if err != nil {
		return fmt.Errorf("setting registry mirror configuration: %v", err)
	}

	kcs.Files = append(kcs.Files, containerdFiles...)

	return nil
}
