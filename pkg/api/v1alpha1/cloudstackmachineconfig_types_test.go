package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestCloudStackMachineConfigDiskOfferingEqual(t *testing.T) {
	diskOffering1 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering1",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	diskOffering2 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering1",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	g.Expect(diskOffering1.Equal(diskOffering2)).To(BeTrue())
}

func TestCloudStackMachineConfigNilDiskOfferingEqual(t *testing.T) {
	var nilDiskOffering *v1alpha1.CloudStackResourceDiskOffering
	emptyDiskOffering := &v1alpha1.CloudStackResourceDiskOffering{
		MountPath:  "",
		Device:     "",
		Filesystem: "",
		Label:      "",
	}
	g := NewWithT(t)
	g.Expect(nilDiskOffering.Equal(emptyDiskOffering)).To(BeTrue())
}

func TestCloudStackMachineConfigEmptyDiskOfferingEqual(t *testing.T) {
	emptyDiskOffering1 := v1alpha1.CloudStackResourceDiskOffering{}
	emptyDiskOffering2 := &v1alpha1.CloudStackResourceDiskOffering{
		MountPath:  "",
		Device:     "",
		Filesystem: "",
		Label:      "",
	}
	g := NewWithT(t)
	g.Expect(emptyDiskOffering1.Equal(emptyDiskOffering2)).To(BeTrue())
}

func TestCloudStackMachineConfigDiskOfferingEqualSelf(t *testing.T) {
	diskOffering1 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering1",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	g.Expect(diskOffering1.Equal(diskOffering1)).To(BeTrue())
}

func TestCloudStackMachineConfigDiskOfferingNotEqualNil(t *testing.T) {
	diskOffering1 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering1",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	g.Expect(diskOffering1.Equal(nil)).To(BeFalse())
}

func TestCloudStackMachineConfigDiskOfferingNotEqualName(t *testing.T) {
	diskOffering1 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering1",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	diskOffering2 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering2",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	g.Expect(diskOffering1.Equal(diskOffering2)).To(BeFalse())
}

func TestCloudStackMachineConfigDiskOfferingNotEqualMountPath(t *testing.T) {
	diskOffering1 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering1",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	diskOffering2 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering2",
		},
		MountPath:  "/data_different",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	g.Expect(diskOffering1.Equal(diskOffering2)).To(BeFalse())
}

func TestCloudStackMachineConfigDiskOfferingNotEqualDevice(t *testing.T) {
	diskOffering1 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering1",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	diskOffering2 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering2",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb_different",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	g.Expect(diskOffering1.Equal(diskOffering2)).To(BeFalse())
}

func TestCloudStackMachineConfigDiskOfferingNotEqualFilesystem(t *testing.T) {
	diskOffering1 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering1",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	diskOffering2 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering2",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "xfs",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	g.Expect(diskOffering1.Equal(diskOffering2)).To(BeFalse())
}

func TestCloudStackMachineConfigDiskOfferingNotEqualLabel(t *testing.T) {
	diskOffering1 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering1",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	diskOffering2 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering2",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk_different",
	}
	g := NewWithT(t)
	g.Expect(diskOffering1.Equal(diskOffering2)).To(BeFalse())
}

func TestCloudStackMachineConfigDiskOfferingValidMountPath(t *testing.T) {
	diskOffering1 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering1",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	err, fieldName, fieldValue := diskOffering1.Validate()
	g.Expect(err == nil).To(BeTrue())
	g.Expect(fieldName == "").To(BeTrue())
	g.Expect(fieldValue == "").To(BeTrue())
}

func TestCloudStackMachineConfigDiskOfferingInValidNoIDAndName(t *testing.T) {
	diskOffering1 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{},
		MountPath:                    "/data",
		Device:                       "/dev/vdb",
		Filesystem:                   "ext4",
		Label:                        "data_disk",
	}
	g := NewWithT(t)
	err, fieldName, fieldValue := diskOffering1.Validate()
	g.Expect(err != nil).To(BeTrue())
	g.Expect(fieldName == "id or name").To(BeTrue())
	g.Expect(fieldValue == "").To(BeTrue())
	g.Expect(err.Error() == "empty id/name").To(BeTrue())
}

func TestCloudStackMachineConfigDiskOfferingValidNoIDAndName(t *testing.T) {
	diskOffering1 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{},
	}
	g := NewWithT(t)
	err, fieldName, fieldValue := diskOffering1.Validate()
	g.Expect(err == nil).To(BeTrue())
	g.Expect(fieldName == "").To(BeTrue())
	g.Expect(fieldValue == "").To(BeTrue())
}

func TestCloudStackMachineConfigDiskOfferingInValidMountPathRoot(t *testing.T) {
	diskOffering1 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering1",
		},
		MountPath:  "/",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	err, fieldName, fieldValue := diskOffering1.Validate()
	g.Expect(err != nil).To(BeTrue())
	g.Expect(fieldName == "mountPath").To(BeTrue())
	g.Expect(fieldValue == "/").To(BeTrue())
	g.Expect(err.Error() == "must be non-empty and start with /").To(BeTrue())
}

func TestCloudStackMachineConfigDiskOfferingInValidMountPathRelative(t *testing.T) {
	diskOffering1 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering1",
		},
		MountPath:  "data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	err, fieldName, fieldValue := diskOffering1.Validate()
	g.Expect(err != nil).To(BeTrue())
	g.Expect(fieldName == "mountPath").To(BeTrue())
	g.Expect(fieldValue == "data").To(BeTrue())
	g.Expect(err.Error() == "must be non-empty and start with /").To(BeTrue())
}

func TestCloudStackMachineConfigDiskOfferingValid(t *testing.T) {
	diskOffering1 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering1",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	err, fieldName, fieldValue := diskOffering1.Validate()
	g.Expect(err == nil).To(BeTrue())
	g.Expect(fieldName == "").To(BeTrue())
	g.Expect(fieldValue == "").To(BeTrue())
}

func TestCloudStackMachineConfigDiskOfferingInValidEmptyDevice(t *testing.T) {
	diskOffering1 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering1",
		},
		MountPath:  "/data",
		Device:     "",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	err, fieldName, fieldValue := diskOffering1.Validate()
	g.Expect(err != nil).To(BeTrue())
	g.Expect(fieldName == "device").To(BeTrue())
	g.Expect(fieldValue == "").To(BeTrue())
	g.Expect(err.Error() == "empty device").To(BeTrue())
}

func TestCloudStackMachineConfigDiskOfferingInValidEmptyFilesystem(t *testing.T) {
	diskOffering1 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering1",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	err, fieldName, fieldValue := diskOffering1.Validate()
	g.Expect(err != nil).To(BeTrue())
	g.Expect(fieldName == "filesystem").To(BeTrue())
	g.Expect(fieldValue == "").To(BeTrue())
	g.Expect(err.Error() == "empty filesystem").To(BeTrue())
}

func TestCloudStackMachineConfigDiskOfferingInValidEmptyLabel(t *testing.T) {
	diskOffering1 := &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering1",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "",
	}
	g := NewWithT(t)
	err, fieldName, fieldValue := diskOffering1.Validate()
	g.Expect(err != nil).To(BeTrue())
	g.Expect(fieldName == "label").To(BeTrue())
	g.Expect(fieldValue == "").To(BeTrue())
	g.Expect(err.Error() == "empty label").To(BeTrue())
}

func TestCloudStackMachineConfigSymlinksValid(t *testing.T) {
	symlinks := v1alpha1.SymlinkMaps{
		"/var/lib.a": "/data/-var/_log",
	}
	err, _, _ := symlinks.Validate()
	g := NewWithT(t)
	g.Expect(err == nil).To(BeTrue())
}

func TestCloudStackMachineConfigSymlinksInValidColon(t *testing.T) {
	symlinks := v1alpha1.SymlinkMaps{
		"/var/lib": "/data/var/log:d",
	}
	err, fieldName, fieldValue := symlinks.Validate()
	g := NewWithT(t)
	g.Expect(err != nil).To(BeTrue())
	g.Expect(fieldName == "symlinks").To(BeTrue())
	g.Expect(fieldValue == "/data/var/log:d").To(BeTrue())
	g.Expect(err.Error() == "has char not in portable file name set").To(BeTrue())
}

func TestCloudStackMachineConfigSymlinksInValidComma(t *testing.T) {
	symlinks := v1alpha1.SymlinkMaps{
		"/var/lib": "/data/var/log,d",
	}
	err, fieldName, fieldValue := symlinks.Validate()
	g := NewWithT(t)
	g.Expect(err != nil).To(BeTrue())
	g.Expect(fieldName == "symlinks").To(BeTrue())
	g.Expect(fieldValue == "/data/var/log,d").To(BeTrue())
	g.Expect(err.Error() == "has char not in portable file name set").To(BeTrue())
}

func TestCloudStackMachineConfigSerialize(t *testing.T) {
	tests := map[string]struct {
		machineConfig interface{}
		expected      string
	}{
		"Serialize machine config": {
			machineConfig: v1alpha1.CloudStackMachineConfig{
				Spec: v1alpha1.CloudStackMachineConfigSpec{
					DiskOffering: &v1alpha1.CloudStackResourceDiskOffering{
						CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
							Name: "diskOffering1",
						},
						MountPath:  "/data",
						Device:     "/dev/sda1",
						Filesystem: "ext4",
						Label:      "data_disk",
					},
				},
			},
			expected: `metadata:
  creationTimestamp: null
spec:
  computeOffering: {}
  diskOffering:
    device: /dev/sda1
    filesystem: ext4
    label: data_disk
    mountPath: /data
    name: diskOffering1
  template: {}
status: {}
`,
		},
		"diskOffering should not appear when it's not defined": {
			machineConfig: v1alpha1.CloudStackMachineConfig{
				Spec: v1alpha1.CloudStackMachineConfigSpec{},
			},
			expected: `metadata:
  creationTimestamp: null
spec:
  computeOffering: {}
  template: {}
status: {}
`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := yaml.Marshal(tc.machineConfig)
			g := NewWithT(t)
			g.Expect(err).To(BeNil())
			g.Expect(string(actual)).To(Equal(tc.expected))
		})
	}
}

func TestCloudStackMachineConfigValidateUsers(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name          string
		machineConfig *v1alpha1.CloudStackMachineConfig
		wantErr       string
	}{
		{
			name: "users valid",
			machineConfig: &v1alpha1.CloudStackMachineConfig{
				Spec: v1alpha1.CloudStackMachineConfigSpec{
					Users: []v1alpha1.UserConfiguration{{
						Name:              "capc",
						SshAuthorizedKeys: []string{"ssh-rsa AAAA..."},
					}},
				},
			},
		},
		{
			name: "users not set",
			machineConfig: &v1alpha1.CloudStackMachineConfig{
				Spec: v1alpha1.CloudStackMachineConfigSpec{},
			},
			wantErr: "users is not set for CloudStackMachineConfig , please provide a user",
		},
		{
			name: "user name empty",
			machineConfig: &v1alpha1.CloudStackMachineConfig{
				Spec: v1alpha1.CloudStackMachineConfigSpec{
					Users: []v1alpha1.UserConfiguration{{
						Name:              "",
						SshAuthorizedKeys: []string{"ssh-rsa AAAA..."},
					}},
				},
			},
			wantErr: "users[0].name is not set or is empty for CloudStackMachineConfig , please provide a username",
		},
		{
			name: "user ssh authorized key empty or not set",
			machineConfig: &v1alpha1.CloudStackMachineConfig{
				Spec: v1alpha1.CloudStackMachineConfigSpec{
					Users: []v1alpha1.UserConfiguration{{
						Name:              "Jeff",
						SshAuthorizedKeys: []string{""},
					}},
				},
			},
			wantErr: "users[0].SshAuthorizedKeys is not set or is empty for CloudStackMachineConfig , please provide a valid ssh authorized key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.machineConfig.ValidateUsers()
			if tt.wantErr == "" {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}

func TestCloudStackMachineConfigSetDefaultUsers(t *testing.T) {
	g := NewWithT(t)
	machineConfig := &v1alpha1.CloudStackMachineConfig{
		Spec: v1alpha1.CloudStackMachineConfigSpec{},
	}
	machineConfig.SetUserDefaults()
	g.Expect(machineConfig.Spec.Users).To(Equal([]v1alpha1.UserConfiguration{
		{
			Name:              v1alpha1.DefaultCloudStackUser,
			SshAuthorizedKeys: []string{""},
		},
	}))
}
