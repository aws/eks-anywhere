package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

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
	var diskOffering2 v1alpha1.CloudStackResourceDiskOffering
	g := NewWithT(t)
	g.Expect(diskOffering1.Equal(&diskOffering2)).To(BeFalse())
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

func TestCloudStackMachineConfigDiskOfferingMountValidPath(t *testing.T) {
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
	g.Expect(diskOffering1.ValidatePath()).To(BeTrue())
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
	g.Expect(diskOffering1.ValidatePath()).To(BeFalse())
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
	g.Expect(diskOffering1.ValidatePath()).To(BeFalse())
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
	g.Expect(diskOffering1.Validate()).To(BeTrue())
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
	g.Expect(diskOffering1.Validate()).To(BeFalse())
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
	g.Expect(diskOffering1.Validate()).To(BeFalse())
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
	g.Expect(diskOffering1.Validate()).To(BeFalse())
}
