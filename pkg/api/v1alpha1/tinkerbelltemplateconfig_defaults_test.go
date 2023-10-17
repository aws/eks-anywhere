package v1alpha1

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestWithDefaultActionsFromBundle(t *testing.T) {
	vBundle := givenVersionBundle()
	tinkerbellLocalIp := "127.0.0.1"
	tinkerbellLBIP := "1.2.3.4"
	metadataString := fmt.Sprintf("http://%s:50061,http://%s:50061", tinkerbellLocalIp, tinkerbellLBIP)
	rhelMetadataString := fmt.Sprintf("'http://%s:50061','http://%s:50061'", tinkerbellLocalIp, tinkerbellLBIP)
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
					Name:    "stream-image",
					Image:   "public.ecr.aws/eks-anywhere/image2disk:latest",
					Timeout: 600,
					Environment: map[string]string{
						"IMG_URL":    "http://tinkerbell-example:8080/bottlerocket-2004-kube-v1.21.5.gz",
						"DEST_DISK":  "{{ index .Hardware.Disks 0 }}",
						"COMPRESSED": "true",
					},
				},
				{
					Name:    "write-bootconfig",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
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
					Name:    "write-user-data",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
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
					Name:    "write-netplan",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
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
					Name:    "reboot-image",
					Image:   "public.ecr.aws/eks-anywhere/reboot:latest",
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
					Name:    "stream-image",
					Image:   "public.ecr.aws/eks-anywhere/image2disk:latest",
					Timeout: 600,
					Environment: map[string]string{
						"IMG_URL":    "http://tinkerbell-example:8080/bottlerocket-2004-kube-v1.21.5.gz",
						"DEST_DISK":  "{{ index .Hardware.Disks 0 }}",
						"COMPRESSED": "true",
					},
				},
				{
					Name:    "write-bootconfig",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
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
					Name:    "write-user-data",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
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
					Name:    "write-netplan",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
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
					Name:    "reboot-image",
					Image:   "public.ecr.aws/eks-anywhere/reboot:latest",
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
					Name:    "stream-image",
					Image:   "public.ecr.aws/eks-anywhere/image2disk:latest",
					Timeout: 600,
					Environment: map[string]string{
						"IMG_URL":    "http://tinkerbell-example:8080/redhat-8.4-kube-v1.21.5.gz",
						"DEST_DISK":  "{{ index .Hardware.Disks 0 }}",
						"COMPRESSED": "true",
					},
				},
				{
					Name:    "write-netplan",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
					Timeout: 90,
					Environment: map[string]string{
						"DEST_DISK":      "{{ formatPartition ( index .Hardware.Disks 0 ) 1 }}",
						"DEST_PATH":      "/etc/netplan/config.yaml",
						"DIRMODE":        "0755",
						"FS_TYPE":        "ext4",
						"GID":            "0",
						"MODE":           "0644",
						"UID":            "0",
						"STATIC_NETPLAN": "true",
					},
					Pid: "host",
				},
				{
					Name:    "disable-cloud-init-network-capabilities",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
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
					Name:    "add-tink-cloud-init-config",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
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
					Name:    "add-tink-cloud-init-ds-config",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
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
					Name:    "reboot-image",
					Image:   "public.ecr.aws/eks-anywhere/reboot:latest",
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
					Name:    "stream-image",
					Image:   "public.ecr.aws/eks-anywhere/image2disk:latest",
					Timeout: 600,
					Environment: map[string]string{
						"IMG_URL":    "http://tinkerbell-example:8080/ubuntu-kube-v1.21.5.gz",
						"DEST_DISK":  "{{ index .Hardware.Disks 0 }}",
						"COMPRESSED": "true",
					},
				},
				{
					Name:    "write-netplan",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
					Timeout: 90,
					Environment: map[string]string{
						"DEST_DISK":      "{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}",
						"DEST_PATH":      "/etc/netplan/config.yaml",
						"DIRMODE":        "0755",
						"FS_TYPE":        "ext4",
						"GID":            "0",
						"MODE":           "0644",
						"UID":            "0",
						"STATIC_NETPLAN": "true",
					},
					Pid: "host",
				},
				{
					Name:    "disable-cloud-init-network-capabilities",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
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
					Name:    "add-tink-cloud-init-config",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
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
					Name:    "add-tink-cloud-init-ds-config",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
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
					Name:    "reboot-image",
					Image:   "public.ecr.aws/eks-anywhere/reboot:latest",
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
					Name:    "stream-image",
					Image:   "public.ecr.aws/eks-anywhere/image2disk:latest",
					Timeout: 600,
					Environment: map[string]string{
						"IMG_URL":    "http://tinkerbell-example:8080/ubuntu-kube-v1.21.5.gz",
						"DEST_DISK":  "{{ index .Hardware.Disks 0 }}",
						"COMPRESSED": "true",
					},
				},
				{
					Name:    "write-netplan",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
					Timeout: 90,
					Environment: map[string]string{
						"DEST_DISK":      "{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}",
						"DEST_PATH":      "/etc/netplan/config.yaml",
						"DIRMODE":        "0755",
						"FS_TYPE":        "ext4",
						"GID":            "0",
						"MODE":           "0644",
						"UID":            "0",
						"STATIC_NETPLAN": "true",
					},
					Pid: "host",
				},
				{
					Name:    "disable-cloud-init-network-capabilities",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
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
					Name:    "add-tink-cloud-init-config",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
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
					Name:    "add-tink-cloud-init-ds-config",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
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
					Name:    "reboot-image",
					Image:   "public.ecr.aws/eks-anywhere/reboot:latest",
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
					Name:    "stream-image",
					Image:   "public.ecr.aws/eks-anywhere/image2disk:latest",
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
					Name:    "write-netplan",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
					Timeout: 90,
					Environment: map[string]string{
						"DEST_DISK":      "{{ formatPartition ( index .Hardware.Disks 0 ) 2 }}",
						"DEST_PATH":      "/etc/netplan/config.yaml",
						"DIRMODE":        "0755",
						"FS_TYPE":        "ext4",
						"GID":            "0",
						"MODE":           "0644",
						"UID":            "0",
						"STATIC_NETPLAN": "true",
					},
					Pid: "host",
				},
				{
					Name:    "disable-cloud-init-network-capabilities",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
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
					Name:    "add-tink-cloud-init-config",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
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
					Name:    "add-tink-cloud-init-ds-config",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
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
					Name:    "reboot-image",
					Image:   "public.ecr.aws/eks-anywhere/reboot:latest",
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
			opts := GetDefaultActionsFromBundle(tt.clusterSpec, vBundle, tt.osImageOverride, tinkerbellLocalIp, tinkerbellLBIP, tt.osFamily)
			for _, opt := range opts {
				opt(&givenActions)
			}
			if diff := cmp.Diff(givenActions, tt.wantActions); diff != "" {
				t.Fatalf("Expected file mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func givenVersionBundle() v1alpha1.VersionsBundle {
	return v1alpha1.VersionsBundle{
		EksD: v1alpha1.EksDRelease{
			Raw: v1alpha1.OSImageBundle{
				Bottlerocket: v1alpha1.Archive{
					URI: "http://tinkerbell-example:8080/bottlerocket-2004-kube-v1.21.5.gz",
				},
			},
		},
		Tinkerbell: v1alpha1.TinkerbellBundle{
			TinkerbellStack: v1alpha1.TinkerbellStackBundle{
				Actions: v1alpha1.ActionsBundle{
					ImageToDisk: v1alpha1.Image{
						URI: "public.ecr.aws/eks-anywhere/image2disk:latest",
					},
					WriteFile: v1alpha1.Image{
						URI: "public.ecr.aws/eks-anywhere/writefile:latest",
					},
					Kexec: v1alpha1.Image{
						URI: "public.ecr.aws/eks-anywhere/kexec:latest",
					},
					Reboot: v1alpha1.Image{
						URI: "public.ecr.aws/eks-anywhere/reboot:latest",
					},
				},
			},
		},
	}
}
