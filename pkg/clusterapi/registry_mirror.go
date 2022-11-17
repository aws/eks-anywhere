package clusterapi

import (
	_ "embed"
	"fmt"

	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/utils/urls"
)

//go:embed config/containerd_config_append.toml
var containerdConfig string

type values map[string]interface{}

func registryMirrorConfigContent(registryAddressMappings map[string]string, registryCert string, insecureSkip bool) (string, error) {
	val := values{
		"registryMirrorAddressMappings": registryAddressMappings,
		"registryCACert":                registryCert,
		"insecureSkip":                  insecureSkip,
	}

	config, err := templater.Execute(containerdConfig, val)
	if err != nil {
		return "", fmt.Errorf("building containerd config file: %v", err)
	}
	return string(config), nil
}

func registryMirrorConfig(registryMirrorConfig *v1alpha1.RegistryMirrorConfiguration) (files []bootstrapv1.File, err error) {
	registryAddressMappings := urls.ToAPIEndpoints(registryMirrorConfig.GetRegistryMirrorAddressMappings())
	registryConfig, err := registryMirrorConfigContent(registryAddressMappings, registryMirrorConfig.CACertContent, registryMirrorConfig.InsecureSkipVerify)
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
			Path:    fmt.Sprintf("/etc/containerd/certs.d/%s/ca.crt", registryAddressMappings[constants.DefaultRegistryMirrorKey]),
			Owner:   "root:root",
			Content: registryMirrorConfig.CACertContent,
		})
	}

	return files, nil
}

func SetRegistryMirrorInKubeadmControlPlane(kcp *controlplanev1.KubeadmControlPlane, mirrorConfig *v1alpha1.RegistryMirrorConfiguration) error {
	if mirrorConfig == nil {
		return nil
	}

	containerdFiles, err := registryMirrorConfig(mirrorConfig)
	if err != nil {
		return fmt.Errorf("setting registry mirror configuration: %v", err)
	}

	kcp.Spec.KubeadmConfigSpec.Files = append(kcp.Spec.KubeadmConfigSpec.Files, containerdFiles...)

	return nil
}

func SetRegistryMirrorInKubeadmConfigTemplate(kct *bootstrapv1.KubeadmConfigTemplate, mirrorConfig *v1alpha1.RegistryMirrorConfiguration) error {
	if mirrorConfig == nil {
		return nil
	}

	containerdFiles, err := registryMirrorConfig(mirrorConfig)
	if err != nil {
		return fmt.Errorf("setting registry mirror configuration: %v", err)
	}

	kct.Spec.Template.Spec.Files = append(kct.Spec.Template.Spec.Files, containerdFiles...)

	return nil
}
