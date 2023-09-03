package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestCloudStackMachineConfigValidateCreateValidDiskOffering(t *testing.T) {
	c := cloudstackMachineConfig()
	c.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "DiskOffering",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateCreate()).To(Succeed())
}

func TestCloudStackMachineConfigValidateCreateInvalidDiskOfferingBadMountPath(t *testing.T) {
	c := cloudstackMachineConfig()
	c.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "DiskOffering",
		},
		MountPath:  "/",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateCreate()).NotTo(Succeed())
}

func TestCloudStackMachineConfigValidateCreateInvalidDiskOfferingEmptyDevice(t *testing.T) {
	c := cloudstackMachineConfig()
	c.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "DiskOffering",
		},
		MountPath:  "/data",
		Device:     "",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateCreate()).NotTo(Succeed())
}

func TestCloudStackMachineConfigValidateCreateInvalidDiskOfferingEmptyFilesystem(t *testing.T) {
	c := cloudstackMachineConfig()
	c.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "DiskOffering",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateCreate()).NotTo(Succeed())
}

func TestCloudStackMachineConfigValidateCreateInvalidDiskOfferingEmptyLabel(t *testing.T) {
	c := cloudstackMachineConfig()
	c.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "DiskOffering",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateCreate()).NotTo(Succeed())
}

func TestCloudStackMachineConfigValidateCreateValidSymlinks(t *testing.T) {
	c := cloudstackMachineConfig()
	c.Spec.Symlinks = v1alpha1.SymlinkMaps{
		"/var/lib.a": "/_data/var-redirect/log.d",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateCreate()).To(Succeed())
}

func TestCloudStackMachineConfigValidateCreateInvalidSymlinksColon(t *testing.T) {
	c := cloudstackMachineConfig()
	c.Spec.Symlinks = v1alpha1.SymlinkMaps{
		"/var/lib:a": "/_data/var-redirect/log:d",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateCreate()).NotTo(Succeed())
}

func TestCloudStackMachineConfigValidateCreateInvalidSymlinksComma(t *testing.T) {
	c := cloudstackMachineConfig()
	c.Spec.Symlinks = v1alpha1.SymlinkMaps{
		"/var/lib:a": "/_data/var-redirect/log,d",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateCreate()).NotTo(Succeed())
}

func TestCloudStackMachineConfigValidateCreateInvalidSymlinksKeyNotStartWithRoot(t *testing.T) {
	c := cloudstackMachineConfig()
	c.Spec.Symlinks = v1alpha1.SymlinkMaps{
		"var/lib": "/data/var/log",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateCreate()).NotTo(Succeed())
}

func TestCloudStackMachineConfigValidateCreateInvalidSymlinksValueNotStartWithRoot(t *testing.T) {
	c := cloudstackMachineConfig()
	c.Spec.Symlinks = v1alpha1.SymlinkMaps{
		"/var/lib": "data/var/log",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateCreate()).NotTo(Succeed())
}

func TestCloudStackMachineConfigValidateCreateInvalidSymlinksKeyEndWithRoot(t *testing.T) {
	c := cloudstackMachineConfig()
	c.Spec.Symlinks = v1alpha1.SymlinkMaps{
		"/var/lib/": "/data/var/log",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateCreate()).NotTo(Succeed())
}

func TestCloudStackMachineConfigValidateCreateInvalidSymlinksValueEndWithRoot(t *testing.T) {
	c := cloudstackMachineConfig()
	c.Spec.Symlinks = v1alpha1.SymlinkMaps{
		"/var/lib": "/data/var/log/",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateCreate()).NotTo(Succeed())
}

func TestCloudStackMachineConfigValidateCreateInvalidTemplateEmpty(t *testing.T) {
	c := cloudstackMachineConfig()
	c.Spec.Template = v1alpha1.CloudStackResourceIdentifier{}
	g := NewWithT(t)
	g.Expect(c.ValidateCreate()).NotTo(Succeed())
}

func TestCloudStackMachineConfigValidateCreateInvalidComputeOfferingEmpty(t *testing.T) {
	c := cloudstackMachineConfig()
	c.Spec.ComputeOffering = v1alpha1.CloudStackResourceIdentifier{}
	g := NewWithT(t)
	g.Expect(c.ValidateCreate()).NotTo(Succeed())
}

func TestCloudStackMachineConfigValidateCreateInvalidUsers(t *testing.T) {
	c := cloudstackMachineConfig()
	c.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	g := NewWithT(t)
	g.Expect(c.ValidateCreate()).NotTo(Succeed())
}

func TestCPCloudStackMachineValidateUpdateTemplateMutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Template = v1alpha1.CloudStackResourceIdentifier{
		Name: "oldTemplate",
	}
	c := vOld.DeepCopy()

	c.Spec.Template = v1alpha1.CloudStackResourceIdentifier{
		Name: "newTemplate",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestWorkersCPCloudStackMachineValidateUpdateTemplateMutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.Spec.Template = v1alpha1.CloudStackResourceIdentifier{
		Name: "oldTemplate",
	}
	c := vOld.DeepCopy()

	c.Spec.Template = v1alpha1.CloudStackResourceIdentifier{
		Name: "newTemplate",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestCPCloudStackMachineValidateUpdateComputeOfferingMutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.ComputeOffering = v1alpha1.CloudStackResourceIdentifier{
		Name: "oldComputeOffering",
	}
	c := vOld.DeepCopy()

	c.Spec.ComputeOffering = v1alpha1.CloudStackResourceIdentifier{
		Name: "newComputeOffering",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestCPCloudStackMachineValidateUpdateDiskOfferingMutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "oldDiskOffering",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	c := vOld.DeepCopy()

	c.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "newDiskOffering",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestCPCloudStackMachineValidateUpdateDiskOfferingMutableFailInvalidMountPath(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "oldDiskOffering",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	c := vOld.DeepCopy()

	c.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "newDiskOffering",
		},
		MountPath:  "/",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCPCloudStackMachineValidateUpdateDiskOfferingMutableFailEmptyDevice(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "oldDiskOffering",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	c := vOld.DeepCopy()

	c.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "newDiskOffering",
		},
		MountPath:  "/data",
		Device:     "",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCPCloudStackMachineValidateUpdateDiskOfferingMutableFailEmptyFilesystem(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "oldDiskOffering",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	c := vOld.DeepCopy()

	c.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "newDiskOffering",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCPCloudStackMachineValidateUpdateDiskOfferingMutableFailEmptyLabel(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "oldDiskOffering",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	c := vOld.DeepCopy()

	c.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "newDiskOffering",
		},
		MountPath:  "/data",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCPCloudStackMachineValidateUpdateSymlinksMutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Symlinks = v1alpha1.SymlinkMaps{
		"/var/log": "/data/var/log",
	}
	c := vOld.DeepCopy()

	c.Spec.Symlinks = v1alpha1.SymlinkMaps{
		"/var/log": "/data_2/var/log",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestCPCloudStackMachineValidateUpdateSymlinksMutableInvalidComma(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Symlinks = v1alpha1.SymlinkMaps{
		"/var/log": "/data/var/log",
	}
	c := vOld.DeepCopy()

	c.Spec.Symlinks = v1alpha1.SymlinkMaps{
		"/var/log": "/data_2/var/log,d",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCPCloudStackMachineValidateUpdateSymlinksMutableColon(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Symlinks = v1alpha1.SymlinkMaps{
		"/var/log": "/data/var/log",
	}
	c := vOld.DeepCopy()

	c.Spec.Symlinks = v1alpha1.SymlinkMaps{
		"/var/log": "/data_2/var/log:d",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkersCPCloudStackMachineValidateUpdateComputeOfferingMutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.Spec.ComputeOffering = v1alpha1.CloudStackResourceIdentifier{
		Name: "oldComputeOffering",
	}
	c := vOld.DeepCopy()

	c.Spec.ComputeOffering = v1alpha1.CloudStackResourceIdentifier{
		Name: "newComputeOffering",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestWorkersCPCloudStackMachineValidateUpdateDiskOfferingMutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "oldDiskOffering",
		},
		MountPath: "/data",
	}
	c := vOld.DeepCopy()

	c.Spec.DiskOffering = &v1alpha1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
			Name: "newDiskOffering",
		},
		MountPath:  "/data-new",
		Device:     "/dev/vdb",
		Filesystem: "ext4",
		Label:      "data_disk",
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementCloudStackMachineValidateUpdateSshAuthorizedKeyMutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagement("test-cluster")
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	vOld.Spec.Users[0].SshAuthorizedKeys = []string{"rsa-blahdeblahbalh"}
	c := vOld.DeepCopy()

	c.Spec.Users[0].SshAuthorizedKeys[0] = "rsa-laDeLala"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestWorkloadCloudStackMachineValidateUpdateSshAuthorizedKeyMutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	vOld.Spec.Users[0].SshAuthorizedKeys = []string{"rsa-blahdeblahbalh"}
	c := vOld.DeepCopy()

	c.Spec.Users[0].SshAuthorizedKeys[0] = "rsa-laDeLala"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestWorkloadCloudStackMachineValidateUpdateSshUsernameMutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{
		Name:              "Jeff",
		SshAuthorizedKeys: []string{"rsa-blahdeblahbalh"},
	}}
	c := vOld.DeepCopy()

	c.Spec.Users[0].Name = "Andy"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestWorkloadCloudStackMachineValidateUpdateInvalidUsers(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{
		Name:              "Jeff",
		SshAuthorizedKeys: []string{"rsa-blahdeblahbalh"},
	}}
	c := vOld.DeepCopy()

	c.Spec.Users[0].Name = ""
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).ToNot(Succeed())
}

func TestCloudStackMachineValidateUpdateInvalidType(t *testing.T) {
	vOld := &v1alpha1.Cluster{}
	c := &v1alpha1.CloudStackMachineConfig{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(vOld)).NotTo(Succeed())
}

func cloudstackMachineConfig() v1alpha1.CloudStackMachineConfig {
	return v1alpha1.CloudStackMachineConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 2)},
		Spec: v1alpha1.CloudStackMachineConfigSpec{
			Template: v1alpha1.CloudStackResourceIdentifier{
				Name: "template1",
			},
			ComputeOffering: v1alpha1.CloudStackResourceIdentifier{
				Name: "offering1",
			},
			Users: []v1alpha1.UserConfiguration{
				{
					Name:              "capc",
					SshAuthorizedKeys: []string{"ssh-rsa AAAA..."},
				},
			},
		},
		Status: v1alpha1.CloudStackMachineConfigStatus{},
	}
}

func TestCloudStackMachineValidateUpdateAffinityImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Affinity = "pro"
	c := vOld.DeepCopy()

	c.Spec.Affinity = "anti"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).ToNot(Succeed())
}

func TestCloudStackMachineValidateUpdateAffinityGroupIdsImmutable(t *testing.T) {
	vOld := cloudstackMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.AffinityGroupIds = []string{"affinity-group-1"}
	c := vOld.DeepCopy()

	c.Spec.AffinityGroupIds = []string{}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).ToNot(Succeed())

	c.Spec.AffinityGroupIds = []string{"affinity-group-2"}
	g = NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).ToNot(Succeed())
}
