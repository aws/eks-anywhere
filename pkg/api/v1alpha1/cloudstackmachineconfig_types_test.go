package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestCloudStackMachineConfigISOAttachmentEqual(t *testing.T) {
	isoAttachment1 := v1alpha1.CloudStackISOAttachment{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "cloudstack-script.iso",
		},
		MountPath:              "/data",
		Device:                 "/dev/sr0",
		RunPreKubeadmCommand:   true,
		RunPostKubeadmCommand:  false,
		PreKubeadmCommandArgs:  []string{"--dry-run true", "--env prod"},
		PostKubeadmCommandArgs: []string{"--dry-run true", "--env prod"},
	}
	isoAttachment2 := &v1alpha1.CloudStackISOAttachment{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "cloudstack-script.iso",
		},
		MountPath:              "/data",
		Device:                 "/dev/sr0",
		RunPreKubeadmCommand:   true,
		RunPostKubeadmCommand:  false,
		PreKubeadmCommandArgs:  []string{"--dry-run true", "--env prod"},
		PostKubeadmCommandArgs: []string{"--dry-run true", "--env prod"},
	}
	g := NewWithT(t)
	g.Expect(isoAttachment1.Equal(isoAttachment2)).To(BeTrue())
}

func TestCloudStackMachineConfigISOAttachmentEqualSelf(t *testing.T) {
	isoAttachment := v1alpha1.CloudStackISOAttachment{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "cloudstack-script.iso",
		},
		MountPath:              "/data",
		Device:                 "/dev/sr0",
		RunPreKubeadmCommand:   true,
		RunPostKubeadmCommand:  false,
		PreKubeadmCommandArgs:  []string{"--dry-run true", "--env prod"},
		PostKubeadmCommandArgs: []string{"--dry-run true", "--env prod"},
	}
	g := NewWithT(t)
	g.Expect(isoAttachment.Equal(&isoAttachment)).To(BeTrue())
}

func TestCloudStackMachineConfigISOAttachmentNotEqualNil(t *testing.T) {
	isoAttachment := v1alpha1.CloudStackISOAttachment{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "cloudstack-script.iso",
		},
		MountPath:              "/data",
		Device:                 "/dev/sr0",
		RunPreKubeadmCommand:   true,
		RunPostKubeadmCommand:  false,
		PreKubeadmCommandArgs:  []string{"--dry-run true", "--env prod"},
		PostKubeadmCommandArgs: []string{"--dry-run true", "--env prod"},
	}
	g := NewWithT(t)
	g.Expect(isoAttachment.Equal(nil)).To(BeFalse())
}

func TestCloudStackMachineConfigISOAttachmentNotEqualName(t *testing.T) {
	isoAttachment1 := v1alpha1.CloudStackISOAttachment{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "cloudstack-script.iso",
		},
		MountPath:              "/data",
		Device:                 "/dev/sr0",
		RunPreKubeadmCommand:   true,
		RunPostKubeadmCommand:  false,
		PreKubeadmCommandArgs:  []string{"--dry-run true", "--env prod"},
		PostKubeadmCommandArgs: []string{"--dry-run true", "--env prod"},
	}
	isoAttachment2 := &v1alpha1.CloudStackISOAttachment{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "cloudstack-script-2.iso",
		},
		MountPath:              "/data",
		Device:                 "/dev/sr0",
		RunPreKubeadmCommand:   true,
		RunPostKubeadmCommand:  false,
		PreKubeadmCommandArgs:  []string{"--dry-run true", "--env prod"},
		PostKubeadmCommandArgs: []string{"--dry-run true", "--env prod"},
	}
	g := NewWithT(t)
	g.Expect(isoAttachment1.Equal(isoAttachment2)).To(BeFalse())
}

func TestCloudStackMachineConfigISOAttachmentNotEqualPreKubeadmCommandArgs(t *testing.T) {
	isoAttachment1 := v1alpha1.CloudStackISOAttachment{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "cloudstack-script.iso",
		},
		MountPath:              "/data",
		Device:                 "/dev/sr0",
		RunPreKubeadmCommand:   true,
		RunPostKubeadmCommand:  false,
		PreKubeadmCommandArgs:  []string{"--dry-run true", "--env prod"},
		PostKubeadmCommandArgs: []string{"--dry-run true", "--env prod"},
	}
	isoAttachment2 := &v1alpha1.CloudStackISOAttachment{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "cloudstack-script.iso",
		},
		MountPath:              "/data",
		Device:                 "/dev/sr0",
		RunPreKubeadmCommand:   true,
		RunPostKubeadmCommand:  false,
		PreKubeadmCommandArgs:  []string{"--dry-run true", "--env dev"},
		PostKubeadmCommandArgs: []string{"--dry-run true", "--env prod"},
	}
	g := NewWithT(t)
	g.Expect(isoAttachment1.Equal(isoAttachment2)).To(BeFalse())
}

func TestCloudStackMachineConfigISOAttachmentNotEqualPostKubeadmCommandArgs(t *testing.T) {
	isoAttachment1 := v1alpha1.CloudStackISOAttachment{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "cloudstack-script.iso",
		},
		MountPath:              "/data",
		Device:                 "/dev/sr0",
		RunPreKubeadmCommand:   true,
		RunPostKubeadmCommand:  false,
		PreKubeadmCommandArgs:  []string{"--dry-run true", "--env prod"},
		PostKubeadmCommandArgs: []string{"--dry-run true", "--env prod"},
	}
	isoAttachment2 := &v1alpha1.CloudStackISOAttachment{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "cloudstack-script.iso",
		},
		MountPath:              "/data",
		Device:                 "/dev/sr0",
		RunPreKubeadmCommand:   true,
		RunPostKubeadmCommand:  false,
		PreKubeadmCommandArgs:  []string{"--dry-run true", "--env prod"},
		PostKubeadmCommandArgs: []string{"--dry-run true", "--env dev"},
	}
	g := NewWithT(t)
	g.Expect(isoAttachment1.Equal(isoAttachment2)).To(BeFalse())
}

func TestCloudStackMachineConfigISOAttachmentNotEqualPostArgsLengthNotEqual(t *testing.T) {
	isoAttachment1 := v1alpha1.CloudStackISOAttachment{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "cloudstack-script.iso",
		},
		MountPath:              "/data",
		Device:                 "/dev/sr0",
		RunPreKubeadmCommand:   true,
		RunPostKubeadmCommand:  false,
		PreKubeadmCommandArgs:  []string{"--dry-run true", "--env prod"},
		PostKubeadmCommandArgs: []string{"--dry-run true", "--env prod"},
	}
	isoAttachment2 := &v1alpha1.CloudStackISOAttachment{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "cloudstack-script.iso",
		},
		MountPath:              "/data",
		Device:                 "/dev/sr0",
		RunPreKubeadmCommand:   true,
		RunPostKubeadmCommand:  false,
		PreKubeadmCommandArgs:  []string{"--dry-run true", "--env prod"},
		PostKubeadmCommandArgs: []string{"--dry-run true"},
	}
	g := NewWithT(t)
	g.Expect(isoAttachment1.Equal(isoAttachment2)).To(BeFalse())
}

func TestCloudStackMachineConfigDiskOfferingEqual(t *testing.T) {
	diskOffering1 := v1alpha1.CloudStackResourceDiskOffering{
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

func TestCloudStackMachineConfigDiskOfferingEqualSelf(t *testing.T) {
	diskOffering1 := v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "diskOffering1",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	g.Expect(diskOffering1.Equal(&diskOffering1)).To(BeTrue())
}

func TestCloudStackMachineConfigDiskOfferingNotEqualNil(t *testing.T) {
	diskOffering1 := v1alpha1.CloudStackResourceDiskOffering{
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
	diskOffering1 := v1alpha1.CloudStackResourceDiskOffering{
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
	diskOffering1 := v1alpha1.CloudStackResourceDiskOffering{
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
	diskOffering1 := v1alpha1.CloudStackResourceDiskOffering{
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
	diskOffering1 := v1alpha1.CloudStackResourceDiskOffering{
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
	diskOffering1 := v1alpha1.CloudStackResourceDiskOffering{
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
	diskOffering1 := v1alpha1.CloudStackResourceDiskOffering{
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
	diskOffering1 := v1alpha1.CloudStackResourceDiskOffering{
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
	diskOffering1 := v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{},
	}
	g := NewWithT(t)
	err, fieldName, fieldValue := diskOffering1.Validate()
	g.Expect(err == nil).To(BeTrue())
	g.Expect(fieldName == "").To(BeTrue())
	g.Expect(fieldValue == "").To(BeTrue())
}

func TestCloudStackMachineConfigDiskOfferingInValidMountPathRoot(t *testing.T) {
	diskOffering1 := v1alpha1.CloudStackResourceDiskOffering{
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
	g.Expect(err.Error() == "must start with / and NOT end with /").To(BeTrue())
}

func TestCloudStackMachineConfigDiskOfferingInValidMountPathRelative(t *testing.T) {
	diskOffering1 := v1alpha1.CloudStackResourceDiskOffering{
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
	g.Expect(err.Error() == "must start with / and NOT end with /").To(BeTrue())
}

func TestCloudStackMachineConfigDiskOfferingValid(t *testing.T) {
	diskOffering1 := v1alpha1.CloudStackResourceDiskOffering{
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
	diskOffering1 := v1alpha1.CloudStackResourceDiskOffering{
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
	g.Expect(err.Error() == "must start with / and NOT end with /").To(BeTrue())
}

func TestCloudStackMachineConfigDiskOfferingInValidEmptyFilesystem(t *testing.T) {
	diskOffering1 := v1alpha1.CloudStackResourceDiskOffering{
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
	diskOffering1 := v1alpha1.CloudStackResourceDiskOffering{
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

func TestCloudStackMachineConfigISOAttachmentValid(t *testing.T) {
	isoAttachment := v1alpha1.CloudStackISOAttachment{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "cloudstack-script.iso",
		},
		MountPath:              "/data",
		Device:                 "/dev/sr0",
		RunPreKubeadmCommand:   true,
		RunPostKubeadmCommand:  false,
		PreKubeadmCommandArgs:  []string{"--dry-run true", "--env prod"},
		PostKubeadmCommandArgs: []string{"--dry-run true", "--env prod"},
	}
	g := NewWithT(t)
	err, fieldName, fieldValue := isoAttachment.Validate()
	g.Expect(err == nil).To(BeTrue())
	g.Expect(fieldName == "").To(BeTrue())
	g.Expect(fieldValue == "").To(BeTrue())
}

func TestCloudStackMachineConfigISOAttachmentInValidEmptyDevice(t *testing.T) {
	isoAttachment := v1alpha1.CloudStackISOAttachment{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "cloudstack-script.iso",
		},
		MountPath:              "/data",
		Device:                 "",
		RunPreKubeadmCommand:   true,
		RunPostKubeadmCommand:  false,
		PreKubeadmCommandArgs:  []string{"--dry-run true", "--env prod"},
		PostKubeadmCommandArgs: []string{"--dry-run true", "--env prod"},
	}
	g := NewWithT(t)
	err, fieldName, fieldValue := isoAttachment.Validate()
	g.Expect(err != nil).To(BeTrue())
	g.Expect(fieldName == "device").To(BeTrue())
	g.Expect(fieldValue == "").To(BeTrue())
	g.Expect(err.Error() == "empty device").To(BeTrue())
}

func TestCloudStackMachineConfigISOAttachmentInValidMountPath(t *testing.T) {
	isoAttachment := v1alpha1.CloudStackISOAttachment{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "cloudstack-script.iso",
		},
		MountPath:              "/data/",
		Device:                 "/dev/sr0",
		RunPreKubeadmCommand:   true,
		RunPostKubeadmCommand:  false,
		PreKubeadmCommandArgs:  []string{"--dry-run true", "--env prod"},
		PostKubeadmCommandArgs: []string{"--dry-run true", "--env prod"},
	}
	g := NewWithT(t)
	err, fieldName, fieldValue := isoAttachment.Validate()
	g.Expect(err != nil).To(BeTrue())
	g.Expect(fieldName == "mountPath").To(BeTrue())
	g.Expect(fieldValue == "/data/").To(BeTrue())
	g.Expect(err.Error() == "must start with / and NOT end with /").To(BeTrue())
}

func TestCloudStackMachineConfigISOAttachmentEmptyIDAndNameWithMountPath(t *testing.T) {
	isoAttachment := v1alpha1.CloudStackISOAttachment{
		MountPath:              "/data/",
		Device:                 "/dev/sr0",
		RunPreKubeadmCommand:   true,
		RunPostKubeadmCommand:  false,
		PreKubeadmCommandArgs:  []string{"--dry-run true", "--env prod"},
		PostKubeadmCommandArgs: []string{"--dry-run true", "--env prod"},
	}
	g := NewWithT(t)
	err, fieldName, fieldValue := isoAttachment.Validate()
	g.Expect(err != nil).To(BeTrue())
	g.Expect(fieldName == "id or name").To(BeTrue())
	g.Expect(fieldValue == "").To(BeTrue())
	g.Expect(err.Error() == "empty id/name").To(BeTrue())
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
