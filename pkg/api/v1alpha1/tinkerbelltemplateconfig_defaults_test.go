package v1alpha1

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell"
)

func TestWithDefaultActionsFromBundle(t *testing.T) {
	tinkerbellLocalIp := "127.0.0.1"
	tinkerbellLBIP := "1.2.3.4"
	metadataString := fmt.Sprintf("http://%s:7172,http://%s:7172", tinkerbellLocalIp, tinkerbellLBIP)
	rhelMetadataString := fmt.Sprintf("'http://%s:7172','http://%s:7172'", tinkerbellLocalIp, tinkerbellLBIP)
	cloudInit := `datasource:
  Ec2:
    metadata_urls: [%s]
    strict_id: false
manage_etc_hosts: localhost
warnings:
  dsid_missing_source: off
`
	tests := []struct {
		testName        string
		osFamily        OSFamily
		osImageOverride string
		clusterSpec     *Cluster
		wantActions     []tinkerbell.Action
	}{
		{
			testName:        "Bottlerocket-sda",
			osFamily:        Bottlerocket,
			osImageOverride: "http://tinkerbell-example:8080/bottlerocket-2004-kube-v1.21.5.gz",
			clusterSpec:     &Cluster{},
			wantActions: []tinkerbell.Action{
				{
					Name:    "stream image to disk",
					Image:   "127.0.0.1/embedded/image2disk",
					Timeout: 600,
					Environment: map[string]string{
						"IMG_URL":    "http://tinkerbell-example:8080/bottlerocket-2004-kube-v1.21.5.gz",
						"DEST_DISK":  "{{ index .Hardware.Disks 0 }}",
						"COMPRESSED": "true",
					},
				},
				{
					Name:    "write Bottlerocket bootconfig",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Pid:     "host",
					Environment: map[string]string{
						"DEST_DISK":           "{{ formatPartition ( index .Hardware.Disks 0 ) 12 }}",
						"FS_TYPE":             "ext4",
						"DEST_PATH":           "/bootconfig.data",
						"BOOTCONFIG_CONTENTS": bottlerocketBootconfig,
						"UID":                 "0",
						"GID":                 "0",
						"MODE":                "0644",
						"DIRMODE":             "0700",
					},
				},
				{
					Name:    "write Bottlerocket user data",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Pid:     "host",
					Environment: map[string]string{
						"DEST_DISK":  "{{ formatPartition ( index .Hardware.Disks 0 ) 12 }}",
						"FS_TYPE":    "ext4",
						"DEST_PATH":  "/user-data.toml",
						"HEGEL_URLS": metadataString,
						"UID":        "0",
						"GID":        "0",
						"MODE":       "0644",
						"DIRMODE":    "0700",
					},
				},
				{
					Name:    "write netplan config",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Pid:     "host",
					Environment: map[string]string{
						"DEST_DISK":           "{{ formatPartition ( index .Hardware.Disks 0 ) 12 }}",
						"FS_TYPE":             "ext4",
						"DEST_PATH":           "/net.toml",
						"STATIC_BOTTLEROCKET": "true",
						"IFNAME":              "eno1",
						"UID":                 "0",
						"GID":                 "0",
						"MODE":                "0644",
						"DIRMODE":             "0755",
					},
				},
				{
					Name:    "reboot",
					Image:   "127.0.0.1/embedded/reboot",
					Timeout: 90,
					Volumes: []string{"/worker:/worker"},
					Pid:     "host",
				},
			},
		},
		{
			testName:        "Bottlerocket-nvme",
			osFamily:        Bottlerocket,
			osImageOverride: "http://tinkerbell-example:8080/bottlerocket-2004-kube-v1.21.5.gz",
			clusterSpec:     &Cluster{},
			wantActions: []tinkerbell.Action{
				{
					Name:    "stream image to disk",
					Image:   "127.0.0.1/embedded/image2disk",
					Timeout: 600,
					Environment: map[string]string{
						"IMG_URL":    "http://tinkerbell-example:8080/bottlerocket-2004-kube-v1.21.5.gz",
						"DEST_DISK":  "{{ index .Hardware.Disks 0 }}",
						"COMPRESSED": "true",
					},
				},
				{
					Name:    "write Bottlerocket bootconfig",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Pid:     "host",
					Environment: map[string]string{
						"DEST_DISK":           "{{ formatPartition ( index .Hardware.Disks 0 ) 12 }}",
						"FS_TYPE":             "ext4",
						"DEST_PATH":           "/bootconfig.data",
						"BOOTCONFIG_CONTENTS": bottlerocketBootconfig,
						"UID":                 "0",
						"GID":                 "0",
						"MODE":                "0644",
						"DIRMODE":             "0700",
					},
				},
				{
					Name:    "write Bottlerocket user data",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Pid:     "host",
					Environment: map[string]string{
						"DEST_DISK":  "{{ formatPartition ( index .Hardware.Disks 0 ) 12 }}",
						"FS_TYPE":    "ext4",
						"DEST_PATH":  "/user-data.toml",
						"HEGEL_URLS": metadataString,
						"UID":        "0",
						"GID":        "0",
						"MODE":       "0644",
						"DIRMODE":    "0700",
					},
				},
				{
					Name:    "write netplan config",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Pid:     "host",
					Environment: map[string]string{
						"DEST_DISK":           "{{ formatPartition ( index .Hardware.Disks 0 ) 12 }}",
						"FS_TYPE":             "ext4",
						"DEST_PATH":           "/net.toml",
						"STATIC_BOTTLEROCKET": "true",
						"IFNAME":              "eno1",
						"UID":                 "0",
						"GID":                 "0",
						"MODE":                "0644",
						"DIRMODE":             "0755",
					},
				},
				{
					Name:    "reboot",
					Image:   "127.0.0.1/embedded/reboot",
					Timeout: 90,
					Volumes: []string{"/worker:/worker"},
					Pid:     "host",
				},
			},
		},
		{
			testName:        "RedHat-sda",
			osFamily:        RedHat,
			clusterSpec:     &Cluster{},
			osImageOverride: "http://tinkerbell-example:8080/redhat-8.4-kube-v1.21.5.gz",
			wantActions: []tinkerbell.Action{
				{
					Name:    "stream image to disk",
					Image:   "127.0.0.1/embedded/image2disk",
					Timeout: 600,
					Environment: map[string]string{
						"IMG_URL":    "http://tinkerbell-example:8080/redhat-8.4-kube-v1.21.5.gz",
						"DEST_DISK":  "{{ index .Hardware.Disks 0 }}",
						"COMPRESSED": "true",
					},
				},
				{
					Name:    "write network manager config",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Environment: map[string]string{
						"DEST_DISK": "{{ formatPartition ( index .Hardware.Disks 0 ) 1 }}",
						"DEST_PATH": "/etc/NetworkManager/system-connections/static-connection.nmconnection",
						"DIRMODE":   "0755",
						"FS_TYPE":   "ext4",
						"GID":       "0",
						"MODE":      "0600",
						"UID":       "0",
						"CONTENTS": `{{ if (index .Hardware.Interfaces 0).DHCP.VLANID }}[connection]
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
{{ else }}[connection]
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
{{ end }}`,
					},
					Pid: "host",
				},
				{
					Name:    "disable cloud-init network capabilities",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Environment: map[string]string{
						"CONTENTS":  "network: {config: disabled}",
						"DEST_DISK": "{{ formatPartition ( index .Hardware.Disks 0 ) 1 }}",
						"DEST_PATH": "/etc/cloud/cloud.cfg.d/99-disable-network-config.cfg",
						"DIRMODE":   "0700",
						"FS_TYPE":   "ext4",
						"GID":       "0",
						"MODE":      "0600",
						"UID":       "0",
					},
				},
				{
					Name:    "add cloud-init config",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Environment: map[string]string{
						"DEST_DISK": "{{ formatPartition ( index .Hardware.Disks 0 ) 1 }}",
						"FS_TYPE":   "ext4",
						"DEST_PATH": "/etc/cloud/cloud.cfg.d/10_tinkerbell.cfg",
						"CONTENTS":  fmt.Sprintf(cloudInit, rhelMetadataString),
						"UID":       "0",
						"GID":       "0",
						"MODE":      "0600",
						"DIRMODE":   "0700",
					},
				},
				{
					Name:    "add cloud-init ds config",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Environment: map[string]string{
						"DEST_DISK": "{{ formatPartition ( index .Hardware.Disks 0 ) 1 }}",
						"FS_TYPE":   "ext4",
						"DEST_PATH": "/etc/cloud/ds-identify.cfg",
						"CONTENTS":  "datasource: Ec2\n",
						"UID":       "0",
						"GID":       "0",
						"MODE":      "0600",
						"DIRMODE":   "0700",
					},
				},
				{
					Name:    "reboot",
					Image:   "127.0.0.1/embedded/reboot",
					Timeout: 90,
					Pid:     "host",
					Volumes: []string{"/worker:/worker"},
				},
			},
		},
		{
			testName:        "Ubuntu-sda",
			osFamily:        Ubuntu,
			clusterSpec:     &Cluster{},
			osImageOverride: "http://tinkerbell-example:8080/ubuntu-kube-v1.21.5.gz",
			wantActions: []tinkerbell.Action{
				{
					Name:    "stream image to disk",
					Image:   "127.0.0.1/embedded/image2disk",
					Timeout: 600,
					Environment: map[string]string{
						"IMG_URL":    "http://tinkerbell-example:8080/ubuntu-kube-v1.21.5.gz",
						"DEST_DISK":  "{{ index .Hardware.Disks 0 }}",
						"COMPRESSED": "true",
					},
				},
				{
					Name:    "write netplan config",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Environment: map[string]string{
						"DEST_DISK": "{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}",
						"DEST_PATH": "/etc/netplan/config.yaml",
						"DIRMODE":   "0755",
						"FS_TYPE":   "ext4",
						"GID":       "0",
						"MODE":      "0644",
						"UID":       "0",
						"CONTENTS": `{{ if (index .Hardware.Interfaces 0).DHCP.VLANID }}network:
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
{{ else }}network:
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
{{ end }}`,
					},
					Pid: "host",
				},
				{
					Name:    "disable cloud-init network capabilities",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Environment: map[string]string{
						"CONTENTS":  "network: {config: disabled}",
						"DEST_DISK": "{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}",
						"DEST_PATH": "/etc/cloud/cloud.cfg.d/99-disable-network-config.cfg",
						"DIRMODE":   "0700",
						"FS_TYPE":   "ext4",
						"GID":       "0",
						"MODE":      "0600",
						"UID":       "0",
					},
				},
				{
					Name:    "add cloud-init config",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Environment: map[string]string{
						"DEST_DISK": "{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}",
						"FS_TYPE":   "ext4",
						"DEST_PATH": "/etc/cloud/cloud.cfg.d/10_tinkerbell.cfg",
						"CONTENTS":  fmt.Sprintf(cloudInit, metadataString),
						"UID":       "0",
						"GID":       "0",
						"MODE":      "0600",
						"DIRMODE":   "0700",
					},
				},
				{
					Name:    "add cloud-init ds config",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Environment: map[string]string{
						"DEST_DISK": "{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}",
						"FS_TYPE":   "ext4",
						"DEST_PATH": "/etc/cloud/ds-identify.cfg",
						"CONTENTS":  "datasource: Ec2\n",
						"UID":       "0",
						"GID":       "0",
						"MODE":      "0600",
						"DIRMODE":   "0700",
					},
				},
				{
					Name:    "reboot",
					Image:   "127.0.0.1/embedded/reboot",
					Timeout: 90,
					Pid:     "host",
					Volumes: []string{"/worker:/worker"},
				},
			},
		},
		{
			testName:        "Ubuntu-nvme",
			osFamily:        Ubuntu,
			clusterSpec:     &Cluster{},
			osImageOverride: "http://tinkerbell-example:8080/ubuntu-kube-v1.21.5.gz",
			wantActions: []tinkerbell.Action{
				{
					Name:    "stream image to disk",
					Image:   "127.0.0.1/embedded/image2disk",
					Timeout: 600,
					Environment: map[string]string{
						"IMG_URL":    "http://tinkerbell-example:8080/ubuntu-kube-v1.21.5.gz",
						"DEST_DISK":  "{{ index .Hardware.Disks 0 }}",
						"COMPRESSED": "true",
					},
				},
				{
					Name:    "write netplan config",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Environment: map[string]string{
						"DEST_DISK": "{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}",
						"DEST_PATH": "/etc/netplan/config.yaml",
						"DIRMODE":   "0755",
						"FS_TYPE":   "ext4",
						"GID":       "0",
						"MODE":      "0644",
						"UID":       "0",
						"CONTENTS": `{{ if (index .Hardware.Interfaces 0).DHCP.VLANID }}network:
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
{{ else }}network:
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
{{ end }}`,
					},
					Pid: "host",
				},
				{
					Name:    "disable cloud-init network capabilities",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Environment: map[string]string{
						"CONTENTS":  "network: {config: disabled}",
						"DEST_DISK": "{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}",
						"DEST_PATH": "/etc/cloud/cloud.cfg.d/99-disable-network-config.cfg",
						"DIRMODE":   "0700",
						"FS_TYPE":   "ext4",
						"GID":       "0",
						"MODE":      "0600",
						"UID":       "0",
					},
				},
				{
					Name:    "add cloud-init config",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Environment: map[string]string{
						"DEST_DISK": "{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}",
						"FS_TYPE":   "ext4",
						"DEST_PATH": "/etc/cloud/cloud.cfg.d/10_tinkerbell.cfg",
						"CONTENTS":  fmt.Sprintf(cloudInit, metadataString),
						"UID":       "0",
						"GID":       "0",
						"MODE":      "0600",
						"DIRMODE":   "0700",
					},
				},
				{
					Name:    "add cloud-init ds config",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Environment: map[string]string{
						"DEST_DISK": "{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}",
						"FS_TYPE":   "ext4",
						"DEST_PATH": "/etc/cloud/ds-identify.cfg",
						"CONTENTS":  "datasource: Ec2\n",
						"UID":       "0",
						"GID":       "0",
						"MODE":      "0600",
						"DIRMODE":   "0700",
					},
				},
				{
					Name:    "reboot",
					Image:   "127.0.0.1/embedded/reboot",
					Timeout: 90,
					Pid:     "host",
					Volumes: []string{"/worker:/worker"},
				},
			},
		},
		{
			testName: "Ubuntu-sda-with-proxy",
			osFamily: Ubuntu,
			clusterSpec: &Cluster{
				Spec: ClusterSpec{
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Endpoint: &Endpoint{
							Host: "1.2.3.4",
						},
					},
					ProxyConfiguration: &ProxyConfiguration{
						HttpProxy:  "2.3.4.5:3128",
						HttpsProxy: "2.3.4.5:3128",
					},
				},
			},
			osImageOverride: "http://tinkerbell-example:8080/ubuntu-kube-v1.21.5.gz",
			wantActions: []tinkerbell.Action{
				{
					Name:    "stream image to disk",
					Image:   "127.0.0.1/embedded/image2disk",
					Timeout: 600,
					Environment: map[string]string{
						"IMG_URL":     "http://tinkerbell-example:8080/ubuntu-kube-v1.21.5.gz",
						"DEST_DISK":   "{{ index .Hardware.Disks 0 }}",
						"COMPRESSED":  "true",
						"HTTPS_PROXY": "2.3.4.5:3128",
						"HTTP_PROXY":  "2.3.4.5:3128",
						"NO_PROXY":    "1.2.3.4,127.0.0.1,1.2.3.4",
					},
				},
				{
					Name:    "write netplan config",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Environment: map[string]string{
						"DEST_DISK": "{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}",
						"DEST_PATH": "/etc/netplan/config.yaml",
						"DIRMODE":   "0755",
						"FS_TYPE":   "ext4",
						"GID":       "0",
						"MODE":      "0644",
						"UID":       "0",
						"CONTENTS": `{{ if (index .Hardware.Interfaces 0).DHCP.VLANID }}network:
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
{{ else }}network:
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
{{ end }}`,
					},
					Pid: "host",
				},
				{
					Name:    "disable cloud-init network capabilities",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Environment: map[string]string{
						"CONTENTS":  "network: {config: disabled}",
						"DEST_DISK": "{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}",
						"DEST_PATH": "/etc/cloud/cloud.cfg.d/99-disable-network-config.cfg",
						"DIRMODE":   "0700",
						"FS_TYPE":   "ext4",
						"GID":       "0",
						"MODE":      "0600",
						"UID":       "0",
					},
				},
				{
					Name:    "add cloud-init config",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Environment: map[string]string{
						"DEST_DISK": "{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}",
						"FS_TYPE":   "ext4",
						"DEST_PATH": "/etc/cloud/cloud.cfg.d/10_tinkerbell.cfg",
						"CONTENTS":  fmt.Sprintf(cloudInit, metadataString),
						"UID":       "0",
						"GID":       "0",
						"MODE":      "0600",
						"DIRMODE":   "0700",
					},
				},
				{
					Name:    "add cloud-init ds config",
					Image:   "127.0.0.1/embedded/writefile",
					Timeout: 90,
					Environment: map[string]string{
						"DEST_DISK": "{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}",
						"FS_TYPE":   "ext4",
						"DEST_PATH": "/etc/cloud/ds-identify.cfg",
						"CONTENTS":  "datasource: Ec2\n",
						"UID":       "0",
						"GID":       "0",
						"MODE":      "0600",
						"DIRMODE":   "0700",
					},
				},
				{
					Name:    "reboot",
					Image:   "127.0.0.1/embedded/reboot",
					Timeout: 90,
					Pid:     "host",
					Volumes: []string{"/worker:/worker"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			givenActions := []tinkerbell.Action{}
			opts := DefaultActions(tt.clusterSpec, tt.osImageOverride, tinkerbellLocalIp, tinkerbellLBIP, tt.osFamily)
			for _, opt := range opts {
				opt(&givenActions)
			}
			if diff := cmp.Diff(givenActions, tt.wantActions); diff != "" {
				t.Fatalf("Expected file mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDefaultActionsIPv6SLAAC(t *testing.T) {
	var actions []tinkerbell.Action
	for _, opt := range DefaultActions(
		&Cluster{},
		"https://example.com/ubuntu-1.32.gz",
		"2001:db8::10",
		"2001:db8::20",
		Ubuntu,
	) {
		opt(&actions)
	}

	var netplan, metadata string
	var streamNetwork string
	for _, action := range actions {
		switch action.Name {
		case "stream image to disk":
			if action.Namespaces != nil {
				streamNetwork = action.Namespaces.Network
			}
		case "write netplan config":
			netplan = action.Environment["CONTENTS"]
		case "add cloud-init config":
			metadata = action.Environment["CONTENTS"]
		}
	}

	if !strings.Contains(netplan, "accept-ra: true") || strings.Contains(netplan, ".DHCP.IP.") {
		t.Fatalf("expected addressless SLAAC netplan, got:\n%s", netplan)
	}
	if streamNetwork != "host" {
		t.Fatalf("expected image download to use host networking, got %q", streamNetwork)
	}
	template := TinkerbellTemplateConfig{
		Spec: TinkerbellTemplateConfigSpec{
			Template: tinkerbell.Workflow{
				Tasks: []tinkerbell.Task{{Actions: actions}},
			},
		},
	}
	rendered, err := template.ToTemplateString()
	if err != nil {
		t.Fatal(err)
	}
	var workflow struct {
		Tasks []struct {
			Actions []struct {
				Network    string `json:"network"`
				Namespaces struct {
					Network string `json:"network"`
				} `json:"namespaces"`
			} `json:"actions"`
		} `json:"tasks"`
	}
	if err := yaml.Unmarshal([]byte(rendered), &workflow); err != nil {
		t.Fatal(err)
	}
	streamAction := workflow.Tasks[0].Actions[0]
	if streamAction.Network != "" || streamAction.Namespaces.Network != "host" {
		t.Fatalf("expected namespaces.network: host in the Tinkerbell template, got:\n%s", rendered)
	}
	if !strings.Contains(metadata, "http://[2001:db8::10]:7172/tootles/instanceID/") ||
		!strings.Contains(metadata, "{{ (index .Hardware.Interfaces 0).DHCP.MAC }}") {
		t.Fatalf("expected bracketed instance-ID metadata URL, got:\n%s", metadata)
	}
}

func TestDefaultActionsCloudInitMetadata(t *testing.T) {
	tests := []struct {
		name      string
		osFamily  OSFamily
		localIP   string
		loadBalIP string
		wantURLs  []string
	}{
		{
			name:      "Ubuntu IPv4",
			osFamily:  Ubuntu,
			localIP:   "192.0.2.10",
			loadBalIP: "192.0.2.20",
			wantURLs:  []string{"http://192.0.2.10:7172", "http://192.0.2.20:7172"},
		},
		{
			name:      "Ubuntu IPv6",
			osFamily:  Ubuntu,
			localIP:   "2001:db8::10",
			loadBalIP: "2001:db8::20",
			wantURLs: []string{
				"http://[2001:db8::10]:7172/tootles/instanceID/{{ (index .Hardware.Interfaces 0).DHCP.MAC }}",
				"http://[2001:db8::20]:7172/tootles/instanceID/{{ (index .Hardware.Interfaces 0).DHCP.MAC }}",
			},
		},
		{
			name:      "RHEL IPv4",
			osFamily:  RedHat,
			localIP:   "192.0.2.10",
			loadBalIP: "192.0.2.20",
			wantURLs:  []string{"http://192.0.2.10:7172", "http://192.0.2.20:7172"},
		},
		{
			name:      "RHEL IPv6",
			osFamily:  RedHat,
			localIP:   "2001:db8::10",
			loadBalIP: "2001:db8::20",
			wantURLs: []string{
				"http://[2001:db8::10]:7172/tootles/instanceID/{{ (index .Hardware.Interfaces 0).DHCP.MAC }}",
				"http://[2001:db8::20]:7172/tootles/instanceID/{{ (index .Hardware.Interfaces 0).DHCP.MAC }}",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actions []tinkerbell.Action
			for _, opt := range DefaultActions(&Cluster{}, "https://example.com/os.gz", tt.localIP, tt.loadBalIP, tt.osFamily) {
				opt(&actions)
			}
			var metadata string
			for _, action := range actions {
				if action.Name == "add cloud-init config" {
					metadata = action.Environment["CONTENTS"]
				}
			}
			var config struct {
				Datasource struct {
					Ec2 struct {
						MetadataURLs []string `json:"metadata_urls"`
					} `json:"Ec2"`
				} `json:"datasource"`
			}
			if err := yaml.Unmarshal([]byte(metadata), &config); err != nil {
				t.Fatalf("invalid cloud-init metadata YAML: %v\n%s", err, metadata)
			}
			if diff := cmp.Diff(tt.wantURLs, config.Datasource.Ec2.MetadataURLs); diff != "" {
				t.Fatalf("metadata URLs mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
