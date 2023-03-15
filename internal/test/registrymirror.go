package test

import (
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// RegistryMirrorInsecureSkipVerifyEnabled returns a test RegistryMirrorConfiguration with InsecureSkipVerify enabled.
func RegistryMirrorInsecureSkipVerifyEnabled() *anywherev1.RegistryMirrorConfiguration {
	return &anywherev1.RegistryMirrorConfiguration{
		Endpoint:           "0.0.0.0",
		Port:               "5000",
		InsecureSkipVerify: true,
	}
}

// RegistryMirrorInsecureSkipVerifyEnabledAndCACert returns a test RegistryMirrorConfiguration with a CACert specified and InsecureSkipVerify enabled.
func RegistryMirrorInsecureSkipVerifyEnabledAndCACert() *anywherev1.RegistryMirrorConfiguration {
	return &anywherev1.RegistryMirrorConfiguration{
		Endpoint:           "0.0.0.0",
		Port:               "5000",
		InsecureSkipVerify: true,
		CACertContent:      CACertContent(),
	}
}

// CACertContent returns a test string representing a cacert contents.
func CACertContent() string {
	return `-----BEGIN CERTIFICATE-----
8wHMSjm0Mzf0VtaqLcoNXEYv3rWB08wydabTAxSAlFMmDJbFyXmI1tYgeps0n/Mt
7dvC9zcJkTibFw8YdV5TTlo3aZYYaUiAsFOLhPB41JA4hOtCgrN38Uj3R7pniAlq
u55B9FjSOOMaooBo+vzdhG/AZD9A2qohos4C7azLjWnTunqjEh0PC0QLZ6oE76jw
xTAY4N4C2s/wIybZxaJ8iQ39OzDpyN2Ym40Q58GVOHt16XCjFVVorVcZsI3y2B9Q
9iV7x+Ulu9jtvZ1K54Uspx6dq4K3
-----END CERTIFICATE-----
`
}

// RegistryMirrorConfigFilesInsecureSkipVerify returns cluster-api bootstrap files that configure containerd
// to use a registry mirror with the insecure_skip_verify flag enabled.
func RegistryMirrorConfigFilesInsecureSkipVerify() []bootstrapv1.File {
	return []bootstrapv1.File{
		{
			Content: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."public.ecr.aws"]
    endpoint = ["https://0.0.0.0:5000"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs."0.0.0.0:5000".tls]
    insecure_skip_verify = true
`,
			Owner: "root:root",
			Path:  "/etc/containerd/config_append.toml",
		},
	}
}

// RegistryMirrorConfigFilesInsecureSkipVerifyAndCACert returns cluster-api bootstrap files that configure containerd
// to use a registry mirror with a cacert file and insecure_skip_verify flag enabled.
func RegistryMirrorConfigFilesInsecureSkipVerifyAndCACert() []bootstrapv1.File {
	return []bootstrapv1.File{
		{
			Content: CACertContent(),
			Owner:   "root:root",
			Path:    "/etc/containerd/certs.d/0.0.0.0:5000/ca.crt",
		},
		{
			Content: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."public.ecr.aws"]
    endpoint = ["https://0.0.0.0:5000"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs."0.0.0.0:5000".tls]
    ca_file = "/etc/containerd/certs.d/0.0.0.0:5000/ca.crt"
    insecure_skip_verify = true
`,
			Owner: "root:root",
			Path:  "/etc/containerd/config_append.toml",
		},
	}
}

// RegistryMirrorPreKubeadmCommands returns a list of commands to writes a config_append.toml file
// to configure the registry mirror and restart containerd.
func RegistryMirrorPreKubeadmCommands() []string {
	return []string{
		"cat /etc/containerd/config_append.toml >> /etc/containerd/config.toml",
		"systemctl daemon-reload",
		"systemctl restart containerd",
	}
}

// RegistryMirrorSudoPreKubeadmCommands returns a list of commands that writes a config_append.toml file
// to configure the registry mirror and restart containerd with sudo permissions.
func RegistryMirrorSudoPreKubeadmCommands() []string {
	return []string{
		"cat /etc/containerd/config_append.toml >> /etc/containerd/config.toml",
		"sudo systemctl daemon-reload",
		"sudo systemctl restart containerd",
	}
}
