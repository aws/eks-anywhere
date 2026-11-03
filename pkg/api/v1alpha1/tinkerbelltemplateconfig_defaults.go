package v1alpha1

import (
	"fmt"
	"net"
	"strings"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell"
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

	// HookOS embeds container images from the bundle.
	// The container images are tagged as below.
	actionImage2Disk = "127.0.0.1/embedded/image2disk"
	actionWriteFile  = "127.0.0.1/embedded/writefile"
	actionReboot     = "127.0.0.1/embedded/reboot"
)

// DefaultActions constructs a set of default actions for the given osFamily.
func DefaultActions(clusterSpec *Cluster, osImageOverride, tinkerbellLocalIP, tinkerbellLBIP string, osFamily OSFamily) []ActionOpt {
	ipv6 := isIPv6Address(tinkerbellLBIP)
	metadataPath := ""
	if ipv6 {
		metadataPath = "/tootles/instanceID/{{ (index .Hardware.Interfaces 0).DHCP.MAC }}"
	}

	// The metadata string will have two URLs:
	// 1. one that will be used initially for bootstrap and will point to tootles running on kind.
	// 2. one that will be used when the workload cluster is up and will point to tootles running on
	//    the workload cluster.
	// Port 7172 is the tootles (metadata service) port in the mono-repo tinkerbell chart.
	metadataURLs := []string{
		fmt.Sprintf("http://%s%s", net.JoinHostPort(tinkerbellLocalIP, "7172"), metadataPath),
		fmt.Sprintf("http://%s%s", net.JoinHostPort(tinkerbellLBIP, "7172"), metadataPath),
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

	actions := []ActionOpt{withStreamImageAction(devicePath, osImageOverride, additionalEnvVar, ipv6)}

	switch osFamily {
	case Bottlerocket:
		partitionPath := fmt.Sprintf(paritionPathFmt, "12")

		actions = append(actions,
			withBottlerocketBootconfigAction(partitionPath),
			withBottlerocketUserDataAction(partitionPath, strings.Join(metadataURLs, ",")),
			// Order matters. This action needs to append to an existing user-data.toml file so
			// must be after withBottlerocketUserDataAction().
			withNetplanAction(partitionPath, osFamily, ipv6),
			withRebootAction(),
		)
	case RedHat:
		var mu []string
		for _, u := range metadataURLs {
			mu = append(mu, fmt.Sprintf("'%s'", u))
		}

		partitionPath := fmt.Sprintf(paritionPathFmt, "1")

		actions = append(actions,
			withNetworkManagerAction(partitionPath, ipv6),
			withDisableCloudInitNetworkCapabilities(partitionPath),
			withTinkCloudInitAction(partitionPath, strings.Join(mu, ",")),
			withDsCloudInitAction(partitionPath),
			withRebootAction(),
		)
	default:
		partitionPath := fmt.Sprintf(paritionPathFmt, "2")
		if ipv6 {
			for i, u := range metadataURLs {
				metadataURLs[i] = fmt.Sprintf("%q", u)
			}
		}

		actions = append(actions,
			withNetplanAction(partitionPath, osFamily, ipv6),
			withDisableCloudInitNetworkCapabilities(partitionPath),
			withTinkCloudInitAction(partitionPath, strings.Join(metadataURLs, ",")),
			withDsCloudInitAction(partitionPath),
			withRebootAction(),
		)
	}

	return actions
}

func withStreamImageAction(disk, imageURL string, additionalEnvVar map[string]string, hostNetwork bool) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		env := map[string]string{
			"DEST_DISK":  disk,
			"IMG_URL":    imageURL,
			"COMPRESSED": "true",
		}

		for k, v := range additionalEnvVar {
			env[k] = v
		}

		action := tinkerbell.Action{
			Name:        "stream image to disk",
			Image:       actionImage2Disk,
			Timeout:     600,
			Environment: env,
		}
		if hostNetwork {
			action.Namespaces = &tinkerbell.ActionNamespaces{Network: "host"}
		}
		*a = append(*a, action)
	}
}

func withNetplanAction(disk string, osFamily OSFamily, ipv6 bool) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		netplanAction := tinkerbell.Action{
			Name:    "write netplan config",
			Image:   actionWriteFile,
			Timeout: 90,
			Environment: map[string]string{
				"DEST_DISK": disk,
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
			// For other OS families (Ubuntu, etc.), use netplan configuration
			netplanAction.Environment["DEST_PATH"] = "/etc/netplan/config.yaml"

			if ipv6 {
				netplanTemplate := `network:
    version: 2
    renderer: networkd
    ethernets:
        id0:
            match:
                macaddress: {{ (index .Hardware.Interfaces 0).DHCP.MAC }}
            dhcp4: false
            dhcp6: false
            accept-ra: true
            link-local: [ipv6]
            nameservers:
                addresses: [{{ range $i, $ns := (index .Hardware.Interfaces 0).DHCP.NameServers }}{{if $i}}, {{end}}{{$ns}}{{end}}]
`
				vlanTemplate := `network:
    version: 2
    renderer: networkd
    ethernets:
        mainif:
            match:
                macaddress: {{ (index .Hardware.Interfaces 0).DHCP.MAC }}
            set-name: mainif
            dhcp4: false
            dhcp6: false
            link-local: [ipv6]

    vlans:
        vlan{{ (index .Hardware.Interfaces 0).DHCP.VLANID }}:
            id: {{ (index .Hardware.Interfaces 0).DHCP.VLANID }}
            link: mainif
            dhcp4: false
            dhcp6: false
            accept-ra: true
            link-local: [ipv6]
            nameservers:
                addresses: [{{ range $i, $ns := (index .Hardware.Interfaces 0).DHCP.NameServers }}{{if $i}}, {{end}}{{$ns}}{{end}}]
`
				netplanAction.Environment["CONTENTS"] = `{{ if (index .Hardware.Interfaces 0).DHCP.VLANID }}` + vlanTemplate + `{{ else }}` + netplanTemplate + `{{ end }}`
				*a = append(*a, netplanAction)
				return
			}

			// Instead of using DHCP to get network information, use the hardware object directly
			netplanTemplate := `network:
    version: 2
    renderer: networkd
    ethernets:
        id0:
            match:
                macaddress: {{ (index .Hardware.Interfaces 0).DHCP.MAC }}
            addresses:
                - {{ (index .Hardware.Interfaces 0).DHCP.IP.Address }}/{{ netmaskToPrefixLength (index .Hardware.Interfaces 0).DHCP.IP.Netmask }}
            nameservers:
                addresses: [{{ range $i, $ns := (index .Hardware.Interfaces 0).DHCP.NameServers }}{{if $i}}, {{end}}{{$ns}}{{end}}]
            routes:
                - to: default
                  via: {{ (index .Hardware.Interfaces 0).DHCP.IP.Gateway }}
`
			// Check if VLAN ID exists in the hardware object and create a VLAN-tagged interface if it does
			vlanTemplate := `network:
    version: 2
    renderer: networkd
    ethernets:
        mainif:
            match:
                macaddress: {{ (index .Hardware.Interfaces 0).DHCP.MAC }}
            set-name: mainif

    vlans:
        vlan{{ (index .Hardware.Interfaces 0).DHCP.VLANID }}:
            id: {{ (index .Hardware.Interfaces 0).DHCP.VLANID }}
            link: mainif
            addresses:
            - {{ (index .Hardware.Interfaces 0).DHCP.IP.Address }}/{{ netmaskToPrefixLength (index .Hardware.Interfaces 0).DHCP.IP.Netmask }}
            nameservers:
                addresses: [{{ range $i, $ns := (index .Hardware.Interfaces 0).DHCP.NameServers }}{{if $i}}, {{end}}{{$ns}}{{end}}]
            {{- if (index .Hardware.Interfaces 0).DHCP.IP.Gateway }}
            routes:
                - to: default
                  via: {{ (index .Hardware.Interfaces 0).DHCP.IP.Gateway }}
            {{- end }}
`
			// Use VLAN template if VLANID is present in the hardware object
			netplanAction.Environment["CONTENTS"] = `{{ if (index .Hardware.Interfaces 0).DHCP.VLANID }}` + vlanTemplate + `{{ else }}` + netplanTemplate + `{{ end }}`
		}
		*a = append(*a, netplanAction)
	}
}

func withDisableCloudInitNetworkCapabilities(disk string) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "disable cloud-init network capabilities",
			Image:   actionWriteFile,
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

func withTinkCloudInitAction(disk, metadataURLs string) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "add cloud-init config",
			Image:   actionWriteFile,
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

func withDsCloudInitAction(disk string) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "add cloud-init ds config",
			Image:   actionWriteFile,
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

func withRebootAction() ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "reboot",
			Image:   actionReboot,
			Timeout: 90,
			Pid:     "host",
			Volumes: []string{"/worker:/worker"},
		})
	}
}

func withBottlerocketBootconfigAction(disk string) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "write Bottlerocket bootconfig",
			Image:   actionWriteFile,
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

func withBottlerocketUserDataAction(disk, metadataURLs string) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		*a = append(*a, tinkerbell.Action{
			Name:    "write Bottlerocket user data",
			Image:   actionWriteFile,
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

func withNetworkManagerAction(disk string, ipv6 bool) ActionOpt {
	return func(a *[]tinkerbell.Action) {
		networkManagerAction := tinkerbell.Action{
			Name:    "write network manager config",
			Image:   actionWriteFile,
			Timeout: 90,
			Environment: map[string]string{
				"DEST_DISK": disk,
				"DEST_PATH": "/etc/NetworkManager/system-connections/static-connection.nmconnection",
				"DIRMODE":   "0755",
				"FS_TYPE":   "ext4",
				"GID":       "0",
				"MODE":      "0600",
				"UID":       "0",
			},
			Pid: "host",
		}

		if ipv6 {
			nmTemplate := `[connection]
id=slaac-connection
type=ethernet
autoconnect=yes
autoconnect-priority=10

[ethernet]
mac-address={{ (index .Hardware.Interfaces 0).DHCP.MAC }}

[ipv4]
method=disabled

[ipv6]
addr-gen-mode=stable-privacy
dns={{ range $i, $ns := (index .Hardware.Interfaces 0).DHCP.NameServers }}{{if $i}};{{end}}{{$ns}}{{end}};
method=auto
`
			nmVlanTemplate := `[connection]
id=slaac-vlan-connection
type=vlan
autoconnect=yes
autoconnect-priority=10

[ethernet]
mac-address={{ (index .Hardware.Interfaces 0).DHCP.MAC }}

[vlan]
flags=1
id={{ (index .Hardware.Interfaces 0).DHCP.VLANID }}
mac-address={{ (index .Hardware.Interfaces 0).DHCP.MAC }}

[ipv4]
method=disabled

[ipv6]
addr-gen-mode=stable-privacy
dns={{ range $i, $ns := (index .Hardware.Interfaces 0).DHCP.NameServers }}{{if $i}};{{end}}{{$ns}}{{end}};
method=auto
`
			networkManagerAction.Environment["CONTENTS"] = `{{ if (index .Hardware.Interfaces 0).DHCP.VLANID }}` + nmVlanTemplate + `{{ else }}` + nmTemplate + `{{ end }}`
			*a = append(*a, networkManagerAction)
			return
		}

		// NetworkManager configuration template for RedHat (no VLAN)
		nmTemplate := `[connection]
id=static-connection
type=ethernet
autoconnect=yes
autoconnect-priority=10

[ethernet]
mac-address={{ (index .Hardware.Interfaces 0).DHCP.MAC }}

[ipv4]
address1={{ (index .Hardware.Interfaces 0).DHCP.IP.Address }}/{{ netmaskToPrefixLength (index .Hardware.Interfaces 0).DHCP.IP.Netmask }}
dns={{ range $i, $ns := (index .Hardware.Interfaces 0).DHCP.NameServers }}{{if $i}};{{end}}{{$ns}}{{end}};
gateway={{ (index .Hardware.Interfaces 0).DHCP.IP.Gateway }}
method=manual

[ipv6]
addr-gen-mode=default
method=ignore
`
		// VLAN connection template for RedHat (single file approach)
		nmVlanTemplate := `[connection]
id=vlan-connection
type=vlan
autoconnect=yes
autoconnect-priority=10

[ethernet]
mac-address={{ (index .Hardware.Interfaces 0).DHCP.MAC }}

[vlan]
flags=1
id={{ (index .Hardware.Interfaces 0).DHCP.VLANID }}
mac-address={{ (index .Hardware.Interfaces 0).DHCP.MAC }}

[ipv4]
address1={{ (index .Hardware.Interfaces 0).DHCP.IP.Address }}/{{ netmaskToPrefixLength (index .Hardware.Interfaces 0).DHCP.IP.Netmask }}
dns={{ range $i, $ns := (index .Hardware.Interfaces 0).DHCP.NameServers }}{{if $i}};{{end}}{{$ns}}{{end}};
gateway={{ (index .Hardware.Interfaces 0).DHCP.IP.Gateway }}
method=manual

[ipv6]
addr-gen-mode=default
method=ignore
`
		// Use VLAN template if VLANID is present in the hardware object
		networkManagerAction.Environment["CONTENTS"] = `{{ if (index .Hardware.Interfaces 0).DHCP.VLANID }}` + nmVlanTemplate + `{{ else }}` + nmTemplate + `{{ end }}`
		*a = append(*a, networkManagerAction)
	}
}

func isIPv6Address(address string) bool {
	ip := net.ParseIP(address)
	return ip != nil && ip.To4() == nil
}
