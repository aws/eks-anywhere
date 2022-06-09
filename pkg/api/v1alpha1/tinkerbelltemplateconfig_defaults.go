package v1alpha1

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

var cloudInit = `datasource:
  Ec2:
    metadata_urls: [%s]
    strict_id: false
system_info:
  default_user:
    name: tink
    groups: [wheel, adm]
    sudo: ["ALL=(ALL) NOPASSWD:ALL"]
    shell: /bin/bash
manage_etc_hosts: localhost
warnings:
  dsid_missing_source: off
`

func GetDefaultActionsFromBundle(b v1alpha1.VersionsBundle, disk string, tinkerbellIp string) []ActionOpt {
	return []ActionOpt{
		withStreamImageAction(b, disk),
		withNetplanAction(b, disk),
		withDisableCloudInitNetworkCapabilities(b, disk),
		withTinkCloudInitAction(b, disk, tinkerbellIp),
		withDsCloudInitAction(b, disk),
		withKexecAction(b, disk),
	}
}

func withStreamImageAction(b v1alpha1.VersionsBundle, disk string) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "stream-image",
			Image:   b.Tinkerbell.TinkerbellStack.Actions.ImageToDisk.URI,
			Timeout: 360,
			Environment: map[string]string{
				"IMG_URL":    b.EksD.Raw.Ubuntu.URI,
				"DEST_DISK":  disk,
				"COMPRESSED": "true",
			},
		})
	}
}

func withNetplanAction(b v1alpha1.VersionsBundle, disk string) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "write-netplan",
			Image:   b.Tinkerbell.TinkerbellStack.Actions.WriteFile.URI,
			Timeout: 90,
			Environment: map[string]string{
				"DEST_DISK":      fmt.Sprintf("%s2", disk),
				"DEST_PATH":      "/etc/netplan/config.yaml",
				"DIRMODE":        "0755",
				"FS_TYPE":        "ext4",
				"GID":            "0",
				"MODE":           "0644",
				"STATIC_NETPLAN": "true",
				"UID":            "0",
			},
			Pid: "host",
		})
	}
}

func withDisableCloudInitNetworkCapabilities(b v1alpha1.VersionsBundle, disk string) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "disable-cloud-init-network-capabilities",
			Image:   b.Tinkerbell.TinkerbellStack.Actions.WriteFile.URI,
			Timeout: 90,
			Environment: map[string]string{
				"CONTENTS":  "network: {config: disabled}",
				"DEST_DISK": fmt.Sprintf("%s2", disk),
				"DEST_PATH": "/etc/cloud/cloud.cfg.d/99-disable-network-config.cfg",
				"DIRMODE":   "0700",
				"FS_TYPE":   "ext4",
				"GID":       "0",
				"MODE":      "0600",
				"UID":       "0",
			},
		})
	}
}

func withTinkCloudInitAction(b v1alpha1.VersionsBundle, disk string, tinkerbellIp string) ActionOpt {
	metadataString := fmt.Sprintf("\"http://%s:50061\"", tinkerbellIp)

	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "add-tink-cloud-init-config",
			Image:   b.Tinkerbell.TinkerbellStack.Actions.WriteFile.URI,
			Timeout: 90,
			Environment: map[string]string{
				"DEST_DISK": fmt.Sprintf("%s2", disk),
				"FS_TYPE":   "ext4",
				"DEST_PATH": "/etc/cloud/cloud.cfg.d/10_tinkerbell.cfg",
				"CONTENTS":  fmt.Sprintf(cloudInit, metadataString),
				"UID":       "0",
				"GID":       "0",
				"MODE":      "0600",
				"DIRMODE":   "0700",
			},
		})
	}
}

func withDsCloudInitAction(b v1alpha1.VersionsBundle, disk string) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "add-tink-cloud-init-ds-config",
			Image:   b.Tinkerbell.TinkerbellStack.Actions.WriteFile.URI,
			Timeout: 90,
			Environment: map[string]string{
				"DEST_DISK": fmt.Sprintf("%s2", disk),
				"FS_TYPE":   "ext4",
				"DEST_PATH": "/etc/cloud/ds-identify.cfg",
				"CONTENTS":  "datasource: Ec2\n",
				"UID":       "0",
				"GID":       "0",
				"MODE":      "0600",
				"DIRMODE":   "0700",
			},
		})
	}
}

func withKexecAction(b v1alpha1.VersionsBundle, disk string) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "kexec-image",
			Image:   b.Tinkerbell.TinkerbellStack.Actions.Kexec.URI,
			Timeout: 90,
			Pid:     "host",
			Environment: map[string]string{
				"BLOCK_DEVICE": fmt.Sprintf("%s2", disk),
				"FS_TYPE":      "ext4",
			},
		})
	}
}
