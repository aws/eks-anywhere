package v1alpha1

import (
	"fmt"
	"strings"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	bottlerocketBootconfig = `kernel {}`

	cloudInit = `datasource:
  Ec2:
    metadata_urls: [%s]
    strict_id: false
manage_etc_hosts: localhost
warnings:
  dsid_missing_source: off
`
)

// GetDefaultActionsFromBundle constructs a set of default actions for the given osFamily using the
// bundle as the source of action images.
func GetDefaultActionsFromBundle(clusterSpec *Cluster, b v1alpha1.VersionsBundle, osImageOverride, tinkerbellLocalIP, tinkerbellLBIP string, osFamily OSFamily) []ActionOpt {
	// The metadata string will have two URLs:
	// 1. one that will be used initially for bootstrap and will point to hegel running on kind.
	// 2. one that will be used when the workload cluster is up and will point to hegel running on
	//    the workload cluster.
	metadataURLs := []string{
		fmt.Sprintf("http://%s:50061", tinkerbellLocalIP),
		fmt.Sprintf("http://%s:50061", tinkerbellLBIP),
	}

	additionalEnvVar := make(map[string]string)

	if clusterSpec.Spec.ProxyConfiguration != nil {
		proxyConfig := clusterSpec.ProxyConfiguration()
		additionalEnvVar["HTTP_PROXY"] = proxyConfig["HTTP_PROXY"]
		additionalEnvVar["HTTPS_PROXY"] = proxyConfig["HTTPS_PROXY"]

		noProxy := fmt.Sprintf("%s,%s", tinkerbellLocalIP, tinkerbellLBIP)
		if proxyConfig["NO_PROXY"] != "" {
			noProxy = fmt.Sprintf("%s,%s", proxyConfig["NO_PROXY"], noProxy)
		}

		additionalEnvVar["NO_PROXY"] = noProxy
	}
	// During workflow reconciliation when the Tinkerbell template is rendered, the Workflow
	// Controller injects a subset of data from the Hardware resource. This lets us use Go template
	// language to render the disks enabling mix'n'match disk types for templates that represent
	// the same kind of machine such as control plane nodes.
	//
	// The devicePath disk index and the storagePartitionPath disk index should match.
	devicePath := "{{ index .Hardware.Disks 0 }}"
	paritionPathFmt := "{{ formatPartition ( index .Hardware.Disks 0 ) %s }}"

	actions := []ActionOpt{withStreamImageAction(b, devicePath, osImageOverride, additionalEnvVar)}

	switch osFamily {
	case Bottlerocket:
		partitionPath := fmt.Sprintf(paritionPathFmt, "12")

		actions = append(actions,
			withBottlerocketBootconfigAction(b, partitionPath),
			withBottlerocketUserDataAction(b, partitionPath, strings.Join(metadataURLs, ",")),
			// Order matters. This action needs to append to an existing user-data.toml file so
			// must be after withBottlerocketUserDataAction().
			withNetplanAction(b, partitionPath, osFamily),
			withRebootAction(b),
		)
	case RedHat:
		var mu []string
		for _, u := range metadataURLs {
			mu = append(mu, fmt.Sprintf("'%s'", u))
		}

		partitionPath := fmt.Sprintf(paritionPathFmt, "1")

		actions = append(actions,
			withNetplanAction(b, partitionPath, osFamily),
			withDisableCloudInitNetworkCapabilities(b, partitionPath),
			withTinkCloudInitAction(b, partitionPath, strings.Join(mu, ",")),
			withDsCloudInitAction(b, partitionPath),
			withRebootAction(b),
		)
	default:
		partitionPath := fmt.Sprintf(paritionPathFmt, "2")

		actions = append(actions,
			withNetplanAction(b, partitionPath, osFamily),
			withDisableCloudInitNetworkCapabilities(b, partitionPath),
			withTinkCloudInitAction(b, partitionPath, strings.Join(metadataURLs, ",")),
			withDsCloudInitAction(b, partitionPath),
			withRebootAction(b),
		)
	}

	return actions
}

func withStreamImageAction(b v1alpha1.VersionsBundle, disk, imageURL string, additionalEnvVar map[string]string) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		env := map[string]string{
			"DEST_DISK":  disk,
			"IMG_URL":    imageURL,
			"COMPRESSED": "true",
		}

		for k, v := range additionalEnvVar {
			env[k] = v
		}

		*a = append(*a, tinkerbell.Action{
			Name:        "stream-image",
			Image:       b.Tinkerbell.TinkerbellStack.Actions.ImageToDisk.URI,
			Timeout:     600,
			Environment: env,
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
			netplanAction.Environment["STATIC_BOTTLEROCKET"] = "true"
			netplanAction.Environment["IFNAME"] = "eno1"
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

func withTinkCloudInitAction(b v1alpha1.VersionsBundle, disk string, metadataURLs string) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "add-tink-cloud-init-config",
			Image:   b.Tinkerbell.TinkerbellStack.Actions.WriteFile.URI,
			Timeout: 90,
			Environment: map[string]string{
				"DEST_DISK": disk,
				"FS_TYPE":   "ext4",
				"DEST_PATH": "/etc/cloud/cloud.cfg.d/10_tinkerbell.cfg",
				"CONTENTS":  fmt.Sprintf(cloudInit, metadataURLs),
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

func withBottlerocketUserDataAction(b v1alpha1.VersionsBundle, disk string, metadataURLs string) ActionOpt {
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
				"HEGEL_URLS": metadataURLs,
				"UID":        "0",
				"GID":        "0",
				"MODE":       "0644",
				"DIRMODE":    "0700",
			},
		})
	}
}
