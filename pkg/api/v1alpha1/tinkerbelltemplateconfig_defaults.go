package v1alpha1

import (
	"fmt"
	"strings"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	bottlerocketNetplan = `# Version is required, it will change as we support
# additional settings
version = 1

# "eno1" is the interface name
# Users may turn on dhcp4 and dhcp6 via boolean
[eno1]
dhcp4 = true
# Define this interface as the "primary" interface
# for the system.  This IP is what kubelet will use
# as the node IP.  If none of the interfaces has
# "primary" set, we choose the first interface in
# the file
primary = true
`
	bottlerocketBootconfig = `kernel {
  console = "tty0", "ttyS0,115200n8"
}
`

	cloudInit = `datasource:
  Ec2:
    metadata_urls: [%s]
    strict_id: false
manage_etc_hosts: localhost
warnings:
  dsid_missing_source: off
`
)

func getDiskPart(disk string) string {
	switch {
	case strings.Contains(disk, "nvme"):
		return fmt.Sprintf("%sp", disk)
	default:
		return disk
	}
}

func GetDefaultActionsFromBundle(b v1alpha1.VersionsBundle, disk, osImageOverride, tinkerbellLocalIp, tinkerbellLBIp string, osFamily OSFamily) []ActionOpt {
	var diskPart string

	defaultActions := []ActionOpt{
		withStreamImageAction(b, disk, osImageOverride, osFamily),
	}

	// The metadata string will have two URLs:
	// - one that will be used initially for bootstrap and will point to hegel running on kind
	// - the other will be used when the workload cluster is up and  will point to hegel running on the workload cluster
	metadataUrls := fmt.Sprintf("http://%s:50061,http://%s:50061", tinkerbellLocalIp, tinkerbellLBIp)

	if osFamily == Bottlerocket {
		diskPart = fmt.Sprintf("%s12", getDiskPart(disk))
		defaultActions = append(defaultActions,
			withNetplanAction(b, diskPart, osFamily),
			withBottlerocketBootconfigAction(b, diskPart),
			withBottlerocketUserDataAction(b, diskPart, metadataUrls),
			withRebootAction(b),
		)
	} else {
		diskPart = fmt.Sprintf("%s2", getDiskPart(disk))
		defaultActions = append(defaultActions,
			withNetplanAction(b, diskPart, osFamily),
			withDisableCloudInitNetworkCapabilities(b, diskPart),
			withTinkCloudInitAction(b, diskPart, metadataUrls),
			withDsCloudInitAction(b, diskPart),
		)
		if strings.Contains(disk, "nvme") {
			defaultActions = append(defaultActions, withRebootAction(b))
		} else {
			defaultActions = append(defaultActions, withKexecAction(b, diskPart))
		}
	}

	return defaultActions
}

func withStreamImageAction(b v1alpha1.VersionsBundle, disk, osImageOverride string, osFamily OSFamily) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		var imageUrl string

		switch {
		case osImageOverride != "":
			imageUrl = osImageOverride
		case osFamily == Bottlerocket:
			imageUrl = b.EksD.Raw.Bottlerocket.URI
		default:
			imageUrl = b.EksD.Raw.Ubuntu.URI
		}

		*a = append(*a, tinkerbell.Action{
			Name:    "stream-image",
			Image:   b.Tinkerbell.TinkerbellStack.Actions.ImageToDisk.URI,
			Timeout: 600,
			Environment: map[string]string{
				"DEST_DISK":  disk,
				"IMG_URL":    imageUrl,
				"COMPRESSED": "true",
			},
		})
	}
}

func withNetplanAction(b v1alpha1.VersionsBundle, disk string, osFamily OSFamily) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		netplanAction := tinkerbell.Action{
			Name:    "write-netplan",
			Image:   b.Tinkerbell.TinkerbellStack.Actions.WriteFile.URI,
			Timeout: 90,
			Environment: map[string]string{
				"DEST_DISK": disk,
				"DEST_PATH": "/etc/netplan/config.yaml",
				"DIRMODE":   "0755",
				"FS_TYPE":   "ext4",
				"GID":       "0",
				"MODE":      "0644",
				"UID":       "0",
			},
			Pid: "host",
		}

		if osFamily == Bottlerocket {
			// Bottlerocket needs to write onto the 12th partition as opposed to 2nd for non-Bottlerocket OS
			netplanAction.Environment["DEST_PATH"] = "/net.toml"
			netplanAction.Environment["CONTENTS"] = bottlerocketNetplan
		} else {
			netplanAction.Environment["STATIC_NETPLAN"] = "true"
		}
		*a = append(*a, netplanAction)
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
				"DEST_DISK": disk,
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

func withTinkCloudInitAction(b v1alpha1.VersionsBundle, disk string, metadataUrls string) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "add-tink-cloud-init-config",
			Image:   b.Tinkerbell.TinkerbellStack.Actions.WriteFile.URI,
			Timeout: 90,
			Environment: map[string]string{
				"DEST_DISK": disk,
				"FS_TYPE":   "ext4",
				"DEST_PATH": "/etc/cloud/cloud.cfg.d/10_tinkerbell.cfg",
				"CONTENTS":  fmt.Sprintf(cloudInit, metadataUrls),
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
				"DEST_DISK": disk,
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
				"BLOCK_DEVICE": disk,
				"FS_TYPE":      "ext4",
			},
		})
	}
}

func withRebootAction(b v1alpha1.VersionsBundle) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "reboot-image",
			Image:   b.Tinkerbell.TinkerbellStack.Actions.Reboot.URI,
			Timeout: 90,
			Pid:     "host",
			Volumes: []string{"/worker:/worker"},
		})
	}
}

func withBottlerocketBootconfigAction(b v1alpha1.VersionsBundle, disk string) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "write-bootconfig",
			Image:   b.Tinkerbell.TinkerbellStack.Actions.WriteFile.URI,
			Timeout: 90,
			Pid:     "host",
			Environment: map[string]string{
				"DEST_DISK":           disk,
				"FS_TYPE":             "ext4",
				"DEST_PATH":           "/bootconfig.data",
				"BOOTCONFIG_CONTENTS": bottlerocketBootconfig,
				"UID":                 "0",
				"GID":                 "0",
				"MODE":                "0644",
				"DIRMODE":             "0700",
			},
		})
	}
}

func withBottlerocketUserDataAction(b v1alpha1.VersionsBundle, disk string, metadataUrls string) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "write-user-data",
			Image:   b.Tinkerbell.TinkerbellStack.Actions.WriteFile.URI,
			Timeout: 90,
			Pid:     "host",
			Environment: map[string]string{
				"DEST_DISK":  disk,
				"FS_TYPE":    "ext4",
				"DEST_PATH":  "/user-data.toml",
				"HEGEL_URLS": metadataUrls,
				"UID":        "0",
				"GID":        "0",
				"MODE":       "0644",
				"DIRMODE":    "0700",
			},
		})
	}
}
