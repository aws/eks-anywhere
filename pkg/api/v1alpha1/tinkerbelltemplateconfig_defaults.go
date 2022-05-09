package v1alpha1

import (
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	netplan = `network:
  version: 2
  renderer: networkd
  ethernets:
      eno1:
          dhcp4: true
`
	cloudInit = `datasource:
  Ec2:
    metadata_urls: []
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
)

func GetDefaultActionsFromBundle(b v1alpha1.VersionsBundle) []ActionOpt {
	return []ActionOpt{
		withStreamImageAction(b),
		withNetplanAction(b),
		withTinkCloudInitAction(b),
		withDsCloudInitAction(b),
		withKexecAction(b),
	}
}

func withStreamImageAction(b v1alpha1.VersionsBundle) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "stream-image",
			Image:   b.Tinkerbell.TinkerbellStack.Actions.ImageToDisk.URI,
			Timeout: 360,
			Environment: map[string]string{
				"IMG_URL":    b.EksD.Raw.Ubuntu.URI,
				"DEST_DISK":  "/dev/sda",
				"COMPRESSED": "true",
			},
		})
	}
}

func withNetplanAction(b v1alpha1.VersionsBundle) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "write-netplan",
			Image:   b.Tinkerbell.TinkerbellStack.Actions.WriteFile.URI,
			Timeout: 90,
			Environment: map[string]string{
				"DEST_DISK": "/dev/sda2",
				"FS_TYPE":   "ext4",
				"DEST_PATH": "/etc/netplan/config.yaml",
				"CONTENTS":  netplan,
				"UID":       "0",
				"GID":       "0",
				"MODE":      "0644",
				"DIRMODE":   "0755",
			},
		})
	}
}

func withTinkCloudInitAction(b v1alpha1.VersionsBundle) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "add-tink-cloud-init-config",
			Image:   b.Tinkerbell.TinkerbellStack.Actions.WriteFile.URI,
			Timeout: 90,
			Environment: map[string]string{
				"DEST_DISK": "/dev/sda2",
				"FS_TYPE":   "ext4",
				"DEST_PATH": "/etc/cloud/cloud.cfg.d/10_tinkerbell.cfg",
				"CONTENTS":  cloudInit,
				"UID":       "0",
				"GID":       "0",
				"MODE":      "0600",
				"DIRMODE":   "0700",
			},
		})
	}
}

func withDsCloudInitAction(b v1alpha1.VersionsBundle) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "add-tink-cloud-init-ds-config",
			Image:   b.Tinkerbell.TinkerbellStack.Actions.WriteFile.URI,
			Timeout: 90,
			Environment: map[string]string{
				"DEST_DISK": "/dev/sda2",
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

func withKexecAction(b v1alpha1.VersionsBundle) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "kexec-image",
			Image:   b.Tinkerbell.TinkerbellStack.Actions.Kexec.URI,
			Timeout: 90,
			Pid:     "host",
			Environment: map[string]string{
				"BLOCK_DEVICE": "/dev/sda2",
				"FS_TYPE":      "ext4",
			},
		})
	}
}
