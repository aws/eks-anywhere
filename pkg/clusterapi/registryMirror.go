package clusterapi

import (
	_ "embed"
	"fmt"
	"net"

	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/templater"
)

//go:embed config/containerd_config_append.toml
var containerdConfig string

type values map[string]interface{}

func registryMirrorConfigContent(registryAddress, registryCert string) (string, error) {
	val := values{
		"registryMirrorAddress": registryAddress,
		"registryCACert":        registryCert,
	}

	config, err := templater.Execute(containerdConfig, val)
	if err != nil {
		return "", fmt.Errorf("failed building containerd config file: %v", err)
	}
	return string(config), nil
}

func registryMirrorConfig(registryMirrorConfig *v1alpha1.RegistryMirrorConfiguration) (files []bootstrapv1.File, preKubeadmCommands []string, err error) {
	registryAddress := net.JoinHostPort(registryMirrorConfig.Endpoint, registryMirrorConfig.Port)
	registryConfig, err := registryMirrorConfigContent(registryAddress, registryMirrorConfig.CACertContent)
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
