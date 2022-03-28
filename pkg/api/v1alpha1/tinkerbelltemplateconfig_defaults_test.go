package v1alpha1

import (
	"reflect"
	"testing"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestWithDefaultActionsFromBundle(t *testing.T) {
	vBundle := givenVersionBundle()
	givenActions := []tinkerbell.Action{}
	wantActions := []tinkerbell.Action{
		{
			Name:    "stream-image",
			Image:   "image2disk:v1.0.0",
			Timeout: 360,
			Environment: map[string]string{
				"IMG_URL":    "http://tinkerbell-example:8080/ubuntu-2004-kube-v1.21.5.gz",
				"DEST_DISK":  "/dev/sda",
				"COMPRESSED": "true",
			},
		},
		{
			Name:    "install-openssl",
			Image:   "cexec:v1.0.0",
			Timeout: 90,
			Environment: map[string]string{
				"BLOCK_DEVICE":        "/dev/sda1",
				"FS_TYPE":             "ext4",
				"CHROOT":              "y",
				"DEFAULT_INTERPRETER": "/bin/sh -c",
				"CMD_LINE":            "apt -y update && apt -y install openssl",
			},
		},
		{
			Name:    "write-netplan",
			Image:   "writefile:v1.0.0",
			Timeout: 90,
			Environment: map[string]string{
				"DEST_DISK": "/dev/sda1",
				"FS_TYPE":   "ext4",
				"DEST_PATH": "/etc/netplan/config.yaml",
				"CONTENTS":  netplan,
				"UID":       "0",
				"GID":       "0",
				"MODE":      "0644",
				"DIRMODE":   "0755",
			},
		},
		{
			Name:    "add-tink-cloud-init-config",
			Image:   "writefile:v1.0.0",
			Timeout: 90,
			Environment: map[string]string{
				"DEST_DISK": "/dev/sda1",
				"FS_TYPE":   "ext4",
				"DEST_PATH": "/etc/cloud/cloud.cfg.d/10_tinkerbell.cfg",
				"CONTENTS":  cloudInit,
				"UID":       "0",
				"GID":       "0",
				"MODE":      "0600",
				"DIRMODE":   "0700",
			},
		},
		{
			Name:    "add-tink-cloud-init-ds-config",
			Image:   "writefile:v1.0.0",
			Timeout: 90,
			Environment: map[string]string{
				"DEST_DISK": "/dev/sda1",
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
			Image:   "kexec:v1.0.0",
			Timeout: 90,
			Pid:     "host",
			Environment: map[string]string{
				"BLOCK_DEVICE": "/dev/sda1",
				"FS_TYPE":      "ext4",
			},
		},
	}

	opts := GetDefaultActionsFromBundle(vBundle)
	for _, opt := range opts {
		opt(&givenActions)
	}

	if !reflect.DeepEqual(givenActions, wantActions) {
		t.Fatalf("Got default actions = %v, want %v", givenActions, wantActions)
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
			},
		},
		Tinkerbell: v1alpha1.TinkerbellBundle{
			Actions: v1alpha1.Actions{
				Cexec: v1alpha1.Image{
					URI: "cexec:v1.0.0",
				},
				Kexec: v1alpha1.Image{
					URI: "kexec:v1.0.0",
				},
				ImageToDisk: v1alpha1.Image{
					URI: "image2disk:v1.0.0",
				},
				WriteFile: v1alpha1.Image{
					URI: "writefile:v1.0.0",
				},
			},
		},
	}
}
