package clusterapi

import (
	_ "embed"
	"fmt"
	"net"

	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/templater"
)

//go:embed config/containerd_config_append.toml
var containerdConfig string

type values map[string]interface{}

func registryMirrorConfigContent(registryAddress, registryCert string, insecureSkip bool) (string, error) {
	val := values{
		"registryMirrorAddress": registryAddress,
		"registryCACert":        registryCert,
		"insecureSkip":          insecureSkip,
	}

	config, err := templater.Execute(containerdConfig, val)
	if err != nil {
		return "", fmt.Errorf("failed building containerd config file: %v", err)
	}
	return string(config), nil
}

func registryMirrorConfig(registryMirrorConfig *v1alpha1.RegistryMirrorConfiguration) (files []bootstrapv1.File, preKubeadmCommands []string, err error) {
	registryAddress := net.JoinHostPort(registryMirrorConfig.Endpoint, registryMirrorConfig.Port)
	registryConfig, err := registryMirrorConfigContent(registryAddress, registryMirrorConfig.CACertContent, registryMirrorConfig.InsecureSkipVerify)
	if err != nil {
		return nil, nil, err
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
			Path:    fmt.Sprintf("/etc/containerd/certs.d/%s/ca.crt", registryAddress),
			Owner:   "root:root",
			Content: registryMirrorConfig.CACertContent,
		})
	}

	preKubeadmCommands = []string{
		"cat /etc/containerd/config_append.toml >> /etc/containerd/config.toml",
		"sudo systemctl daemon-reload",
		"sudo systemctl restart containerd",
	}
	return files, preKubeadmCommands, nil
}

func SetRegistryMirrorInKubeadmControlPlane(kcp *controlplanev1.KubeadmControlPlane, mirrorConfig *v1alpha1.RegistryMirrorConfiguration) error {
	if mirrorConfig == nil {
		return nil
	}

	containerdFiles, containerdCommands, err := registryMirrorConfig(mirrorConfig)
	if err != nil {
		return fmt.Errorf("setting registry mirror configuration: %v", err)
	}

	kcp.Spec.KubeadmConfigSpec.Files = append(kcp.Spec.KubeadmConfigSpec.Files, containerdFiles...)
	kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands = append(kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands, containerdCommands...)

	return nil
}

func SetRegistryMirrorInKubeadmConfigTemplate(kct *bootstrapv1.KubeadmConfigTemplate, mirrorConfig *v1alpha1.RegistryMirrorConfiguration) error {
	if mirrorConfig == nil {
		return nil
	}

	containerdFiles, containerdCommands, err := registryMirrorConfig(mirrorConfig)
	if err != nil {
		return fmt.Errorf("setting registry mirror configuration: %v", err)
	}

	kct.Spec.Template.Spec.Files = append(kct.Spec.Template.Spec.Files, containerdFiles...)
	kct.Spec.Template.Spec.PreKubeadmCommands = append(kct.Spec.Template.Spec.PreKubeadmCommands, containerdCommands...)

	return nil
}
