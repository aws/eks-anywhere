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
			Image:   "public.ecr.aws/eks-anywhere/image2disk:latest",
			Timeout: 360,
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
		},
		{
			Name:    "add-tink-cloud-init-config",
			Image:   "public.ecr.aws/eks-anywhere/writefile:latest",
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
				ImageToDisk: v1alpha1.Image{
					URI: "public.ecr.aws/eks-anywhere/image2disk:latest",
				},
				WriteFile: v1alpha1.Image{
					URI: "public.ecr.aws/eks-anywhere/writefile:latest",
				},
				Kexec: v1alpha1.Image{
					URI: "public.ecr.aws/eks-anywhere/kexec:latest",
				},
			},
		},
	}
}
