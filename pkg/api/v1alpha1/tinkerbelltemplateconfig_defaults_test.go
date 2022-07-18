package v1alpha1

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestWithDefaultActionsFromBundle(t *testing.T) {
	vBundle := givenVersionBundle()
	tinkerbellLocalIp := "127.0.0.1"
	tinkerbellLBIP := "1.2.3.4"
	metadataString := fmt.Sprintf("http://%s:50061,http://%s:50061", tinkerbellLocalIp, tinkerbellLBIP)

	tests := []struct {
		testName    string
		diskType    string
		osFamily    OSFamily
		wantActions []tinkerbell.Action
	}{
		{
			testName: "Ubuntu-sda",
			diskType: "/dev/sda",
			osFamily: Ubuntu,
			wantActions: []tinkerbell.Action{
				{
					Name:    "stream-image",
					Image:   "public.ecr.aws/eks-anywhere/image2disk:latest",
					Timeout: 600,
					Environment: map[string]string{
						"IMG_URL":    "http://tinkerbell-example:8080/ubuntu-2004-kube-v1.21.5.gz",
						"DEST_DISK":  "/dev/sda",
						"COMPRESSED": "true",
					},
				},
				{
					Name:    "write-netplan",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
					Timeout: 90,
					Pid:     "host",
					Environment: map[string]string{
						"DEST_DISK":      "/dev/sda2",
						"FS_TYPE":        "ext4",
						"DEST_PATH":      "/etc/netplan/config.yaml",
						"STATIC_NETPLAN": "true",
						"UID":            "0",
						"GID":            "0",
						"MODE":           "0644",
						"DIRMODE":        "0755",
					},
				},
				{
					Name:    "disable-cloud-init-network-capabilities",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
					Timeout: 90,
					Environment: map[string]string{
						"CONTENTS":  "network: {config: disabled}",
						"DEST_DISK": "/dev/sda2",
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
						"DEST_DISK": "/dev/sda2",
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
						"DEST_DISK": "/dev/sda2",
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
					Name:    "kexec-image",
					Image:   "public.ecr.aws/eks-anywhere/kexec:latest",
					Timeout: 90,
					Pid:     "host",
					Environment: map[string]string{
						"BLOCK_DEVICE": "/dev/sda2",
						"FS_TYPE":      "ext4",
					},
				},
			},
		},
		{
			testName: "Ubuntu-nvme",
			diskType: "/dev/nvme0n1",
			osFamily: Ubuntu,
			wantActions: []tinkerbell.Action{
				{
					Name:    "stream-image",
					Image:   "public.ecr.aws/eks-anywhere/image2disk:latest",
					Timeout: 600,
					Environment: map[string]string{
						"IMG_URL":    "http://tinkerbell-example:8080/ubuntu-2004-kube-v1.21.5.gz",
						"DEST_DISK":  "/dev/nvme0n1",
						"COMPRESSED": "true",
					},
				},
				{
					Name:    "write-netplan",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
					Timeout: 90,
					Pid:     "host",
					Environment: map[string]string{
						"DEST_DISK":      "/dev/nvme0n1p2",
						"FS_TYPE":        "ext4",
						"DEST_PATH":      "/etc/netplan/config.yaml",
						"STATIC_NETPLAN": "true",
						"UID":            "0",
						"GID":            "0",
						"MODE":           "0644",
						"DIRMODE":        "0755",
					},
				},
				{
					Name:    "disable-cloud-init-network-capabilities",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
					Timeout: 90,
					Environment: map[string]string{
						"CONTENTS":  "network: {config: disabled}",
						"DEST_DISK": "/dev/nvme0n1p2",
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
						"DEST_DISK": "/dev/nvme0n1p2",
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
						"DEST_DISK": "/dev/nvme0n1p2",
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
					Volumes: []string{"/worker:/worker"},
					Pid:     "host",
				},
			},
		},
		{
			testName: "Bottlerocket-sda",
			diskType: "/dev/sda",
			osFamily: Bottlerocket,
			wantActions: []tinkerbell.Action{
				{
					Name:    "stream-image",
					Image:   "public.ecr.aws/eks-anywhere/image2disk:latest",
					Timeout: 600,
					Environment: map[string]string{
						"IMG_URL":    "http://tinkerbell-example:8080/bottlerocket-2004-kube-v1.21.5.gz",
						"DEST_DISK":  "/dev/sda",
						"COMPRESSED": "true",
					},
				},
				{
					Name:    "write-netplan",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
					Timeout: 90,
					Pid:     "host",
					Environment: map[string]string{
						"DEST_DISK": "/dev/sda12",
						"FS_TYPE":   "ext4",
						"DEST_PATH": "/net.toml",
						"CONTENTS":  bottlerocketNetplan,
						"UID":       "0",
						"GID":       "0",
						"MODE":      "0644",
						"DIRMODE":   "0755",
					},
				},
				{
					Name:    "write-bootconfig",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
					Timeout: 90,
					Pid:     "host",
					Environment: map[string]string{
						"DEST_DISK":           "/dev/sda12",
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
						"DEST_DISK":  "/dev/sda12",
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
					Name:    "reboot-image",
					Image:   "public.ecr.aws/eks-anywhere/reboot:latest",
					Timeout: 90,
					Volumes: []string{"/worker:/worker"},
					Pid:     "host",
				},
			},
		},
		{
			testName: "Bottlerocket-nvme",
			diskType: "/dev/nvme0n1",
			osFamily: Bottlerocket,
			wantActions: []tinkerbell.Action{
				{
					Name:    "stream-image",
					Image:   "public.ecr.aws/eks-anywhere/image2disk:latest",
					Timeout: 600,
					Environment: map[string]string{
						"IMG_URL":    "http://tinkerbell-example:8080/bottlerocket-2004-kube-v1.21.5.gz",
						"DEST_DISK":  "/dev/nvme0n1",
						"COMPRESSED": "true",
					},
				},
				{
					Name:    "write-netplan",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
					Timeout: 90,
					Pid:     "host",
					Environment: map[string]string{
						"DEST_DISK": "/dev/nvme0n1p12",
						"FS_TYPE":   "ext4",
						"DEST_PATH": "/net.toml",
						"CONTENTS":  bottlerocketNetplan,
						"UID":       "0",
						"GID":       "0",
						"MODE":      "0644",
						"DIRMODE":   "0755",
					},
				},
				{
					Name:    "write-bootconfig",
					Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
					Timeout: 90,
					Pid:     "host",
					Environment: map[string]string{
						"DEST_DISK":           "/dev/nvme0n1p12",
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
						"DEST_DISK":  "/dev/nvme0n1p12",
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
					Name:    "reboot-image",
					Image:   "public.ecr.aws/eks-anywhere/reboot:latest",
					Timeout: 90,
					Volumes: []string{"/worker:/worker"},
					Pid:     "host",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			givenActions := []tinkerbell.Action{}
			opts := GetDefaultActionsFromBundle(vBundle, tt.diskType, "", tinkerbellLocalIp, tinkerbellLBIP, tt.osFamily)
			for _, opt := range opts {
				opt(&givenActions)
			}

			if !reflect.DeepEqual(givenActions, tt.wantActions) {
				t.Fatalf("Got default actions = %+v, want %+v", givenActions, tt.wantActions)
			}
		})
	}
}

func givenVersionBundle() v1alpha1.VersionsBundle {
	return v1alpha1.VersionsBundle{
		EksD: v1alpha1.EksDRelease{
			Raw: v1alpha1.OSImageBundle{
				Ubuntu: v1alpha1.OSImage{
					Archive: v1alpha1.Archive{
						URI: "http://tinkerbell-example:8080/ubuntu-2004-kube-v1.21.5.gz",
					},
				},
				Bottlerocket: v1alpha1.OSImage{
					Archive: v1alpha1.Archive{
						URI: "http://tinkerbell-example:8080/bottlerocket-2004-kube-v1.21.5.gz",
					},
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
