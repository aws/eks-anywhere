package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestManagementCPVSphereMachineValidateUpdateTemplateImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Template = "oldTemplate"
	c := vOld.DeepCopy()

	c.Spec.Template = "newTemplate"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadCPVSphereMachineValidateUpdateTemplateSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Template = "oldTemplate"
	c := vOld.DeepCopy()

	c.Spec.Template = "newTemplate"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementWorkersVSphereMachineValidateUpdateTemplateSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.Template = "oldTemplate"
	c := vOld.DeepCopy()

	c.Spec.Template = "newTemplate"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestWorkloadWorkersVSphereMachineValidateUpdateTemplateSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Template = "oldTemplate"
	c := vOld.DeepCopy()

	c.Spec.Template = "newTemplate"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementEtcdVSphereMachineValidateUpdateTemplateImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.Spec.Template = "oldTemplate"
	c := vOld.DeepCopy()

	c.Spec.Template = "newTemplate"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadEtcdVSphereMachineValidateUpdateTemplateSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Template = "oldTemplate"
	c := vOld.DeepCopy()

	c.Spec.Template = "newTemplate"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestVSphereMachineValidateUpdateOSFamilyImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.OSFamily = v1alpha1.Ubuntu
	c := vOld.DeepCopy()

	c.Spec.OSFamily = v1alpha1.Bottlerocket
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestManagementCPVSphereMachineValidateUpdateMemoryMiBImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.MemoryMiB = 2
	c := vOld.DeepCopy()

	c.Spec.MemoryMiB = 2000000
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadCPVSphereMachineValidateUpdateMemoryMiBSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.MemoryMiB = 2
	c := vOld.DeepCopy()

	c.Spec.MemoryMiB = 2000000
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementWorkersVSphereMachineValidateUpdateMemoryMiBSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.MemoryMiB = 2
	c := vOld.DeepCopy()

	c.Spec.MemoryMiB = 2000000
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestWorkloadWorkersVSphereMachineValidateUpdateMemoryMiBSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.MemoryMiB = 2
	c := vOld.DeepCopy()

	c.Spec.MemoryMiB = 2000000
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementEtcdVSphereMachineValidateUpdateMemoryMiBImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.Spec.MemoryMiB = 2

	c := vOld.DeepCopy()

	c.Spec.MemoryMiB = 2000000
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadEtcdVSphereMachineValidateUpdateMemoryMiBSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.MemoryMiB = 2
	c := vOld.DeepCopy()

	c.Spec.MemoryMiB = 2000000
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementCPVSphereMachineValidateUpdateNumCPUsImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.NumCPUs = 1
	c := vOld.DeepCopy()

	c.Spec.NumCPUs = 16
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadCPVSphereMachineValidateUpdateNumCPUsSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.NumCPUs = 1
	c := vOld.DeepCopy()

	c.Spec.NumCPUs = 16
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementWorkersVSphereMachineValidateUpdateNumCPUsSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.NumCPUs = 1
	c := vOld.DeepCopy()

	c.Spec.NumCPUs = 16
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestWorkloadWorkersVSphereMachineValidateUpdateNumCPUsSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.NumCPUs = 1
	c := vOld.DeepCopy()

	c.Spec.NumCPUs = 16
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementEtcdVSphereMachineValidateUpdateNumCPUsImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.Spec.NumCPUs = 1
	c := vOld.DeepCopy()

	c.Spec.NumCPUs = 16
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadEtcdVSphereMachineValidateUpdateNumCPUsSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.NumCPUs = 1
	c := vOld.DeepCopy()

	c.Spec.NumCPUs = 16
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementCPVSphereMachineValidateUpdateDiskGiBImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.DiskGiB = 1
	c := vOld.DeepCopy()

	c.Spec.DiskGiB = 160
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadCPVSphereMachineValidateUpdateDiskGiBSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.DiskGiB = 1
	c := vOld.DeepCopy()

	c.Spec.DiskGiB = 160
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementWorkersVSphereMachineValidateUpdateDiskGiBSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.DiskGiB = 1
	c := vOld.DeepCopy()

	c.Spec.DiskGiB = 160
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestWorkloadWorkersVSphereMachineValidateUpdateDiskGiBSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.DiskGiB = 1
	c := vOld.DeepCopy()

	c.Spec.DiskGiB = 160
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementEtcdVSphereMachineValidateUpdateDiskGiBImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.Spec.DiskGiB = 1

	c := vOld.DeepCopy()

	c.Spec.DiskGiB = 160
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadEtcdVSphereMachineValidateUpdateDiskGiBSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.DiskGiB = 1
	c := vOld.DeepCopy()

	c.Spec.DiskGiB = 160
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementControlPlaneSphereMachineValidateUpdateSshAuthorizedKeyImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	vOld.Spec.Users[0].SshAuthorizedKeys = []string{"rsa-blahdeblahbalh"}
	c := vOld.DeepCopy()

	c.Spec.Users[0].SshAuthorizedKeys[0] = "rsa-laDeLala"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadControlPlaneVSphereMachineValidateUpdateSshAuthorizedKeyImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	vOld.Spec.Users[0].SshAuthorizedKeys = []string{"rsa-blahdeblahbalh"}
	c := vOld.DeepCopy()

	c.Spec.Users[0].SshAuthorizedKeys[0] = "rsa-laDeLala"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementControlPlaneVSphereMachineValidateUpdateSshUsernameImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	c := vOld.DeepCopy()

	c.Spec.Users[0].Name = "Andy"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadControlPlaneVSphereMachineValidateUpdateSshUsernameImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	c := vOld.DeepCopy()

	c.Spec.Users[0].Name = "Andy"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestVSphereMachineValidateUpdateWithPausedAnnotation(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.DiskGiB = 1
	vOld.Spec.Template = "oldTemplate"
	c := vOld.DeepCopy()

	c.Spec.DiskGiB = 160
	c.Spec.Template = "newTemplate"

	vOld.PauseReconcile()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementEtcdSphereMachineValidateUpdateSshAuthorizedKeyImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	vOld.Spec.Users[0].SshAuthorizedKeys = []string{"rsa-blahdeblahbalh"}
	c := vOld.DeepCopy()

	c.Spec.Users[0].SshAuthorizedKeys[0] = "rsa-laDeLala"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadEtcdVSphereMachineValidateUpdateSshAuthorizedKeyImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	vOld.Spec.Users[0].SshAuthorizedKeys = []string{"rsa-blahdeblahbalh"}
	c := vOld.DeepCopy()

	c.Spec.Users[0].SshAuthorizedKeys[0] = "rsa-laDeLala"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementEtcdVSphereMachineValidateUpdateSshUsernameImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	c := vOld.DeepCopy()

	c.Spec.Users[0].Name = "Andy"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadEtcdVSphereMachineValidateUpdateSshUsernameImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	c := vOld.DeepCopy()

	c.Spec.Users[0].Name = "Andy"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementWorkerNodeSphereMachineValidateUpdateSshAuthorizedKeyImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	vOld.Spec.Users[0].SshAuthorizedKeys = []string{"rsa-blahdeblahbalh"}
	c := vOld.DeepCopy()

	c.Spec.Users[0].SshAuthorizedKeys[0] = "rsa-laDeLala"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestWorkloadWorkerNodeVSphereMachineValidateUpdateSshAuthorizedKeyImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	vOld.Spec.Users[0].SshAuthorizedKeys = []string{"rsa-blahdeblahbalh"}
	c := vOld.DeepCopy()

	c.Spec.Users[0].SshAuthorizedKeys[0] = "rsa-laDeLala"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementWorkerNodeVSphereMachineValidateUpdateSshUsernameImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	c := vOld.DeepCopy()

	c.Spec.Users[0].Name = "Andy"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestWorkloadWorkerNodeVSphereMachineValidateUpdateSshUsernameImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	c := vOld.DeepCopy()

	c.Spec.Users[0].Name = "Andy"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestVSphereMachineValidateUpdateInvalidType(t *testing.T) {
	vOld := &v1alpha1.Cluster{}
	c := &v1alpha1.VSphereMachineConfig{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(vOld)).NotTo(Succeed())
}

func TestVSphereMachineValidateUpdateSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()

	vOld.Spec.NumCPUs = 4
	c := vOld.DeepCopy()
	c.Spec.NumCPUs = 16

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementCPVSphereMachineValidateUpdateDatastoreImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Datastore = "OldDataStore"
	c := vOld.DeepCopy()

	c.Spec.Datastore = "NewDataStore"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadCPVSphereMachineValidateUpdateDatastoreSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Datastore = "OldDataStore"
	c := vOld.DeepCopy()

	c.Spec.Datastore = "NewDataStore"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementEtcdVSphereMachineValidateUpdateDatastoreImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.Spec.Datastore = "OldDataStore"
	c := vOld.DeepCopy()

	c.Spec.Datastore = "NewDataStore"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadEtcdVSphereMachineValidateUpdateDatastoreSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Datastore = "OldDataStore"
	c := vOld.DeepCopy()

	c.Spec.Datastore = "NewDataStore"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementWorkersVSphereMachineValidateUpdateDatastoreSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.Datastore = "OldDataStore"
	c := vOld.DeepCopy()

	c.Spec.Datastore = "NewDataStore"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestWorkloadWorkersVSphereMachineValidateUpdateDatastoreSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Datastore = "OldDataStore"
	c := vOld.DeepCopy()

	c.Spec.Datastore = "NewDataStore"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementCPVSphereMachineValidateUpdateFolderImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Folder = "/dev/null"
	c := vOld.DeepCopy()

	c.Spec.Folder = "/tmp"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadCPVSphereMachineValidateUpdateFolderSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Folder = "/dev/null"
	c := vOld.DeepCopy()

	c.Spec.Folder = "/tmp"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementEtcdVSphereMachineValidateUpdateFolderImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.Spec.Folder = "/dev/null"
	c := vOld.DeepCopy()

	c.Spec.Folder = "/tmp"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadEtcdVSphereMachineValidateUpdateFolderSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Folder = "/dev/null"
	c := vOld.DeepCopy()

	c.Spec.Folder = "/tmp"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementWorkersVSphereMachineValidateUpdateFolderSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.Folder = "/dev/null"
	c := vOld.DeepCopy()

	c.Spec.Folder = "/tmp"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestWorkloadWorkersVSphereMachineValidateUpdateFolderSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Folder = "/dev/null"
	c := vOld.DeepCopy()

	c.Spec.Folder = "/tmp"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementCPVSphereMachineValidateUpdateResourcePoolImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.ResourcePool = "AbovegroundPool"
	c := vOld.DeepCopy()

	c.Spec.ResourcePool = "IngroundPool"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadCPVSphereMachineValidateUpdateResourcePoolSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.ResourcePool = "AbovegroundPool"
	c := vOld.DeepCopy()

	c.Spec.ResourcePool = "IngroundPool"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementEtcdVSphereMachineValidateUpdateResourcePoolImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.Spec.ResourcePool = "AbovegroundPool"
	c := vOld.DeepCopy()

	c.Spec.ResourcePool = "IngroundPool"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestWorkloadEtcdVSphereMachineValidateUpdateResourcePoolSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.ResourcePool = "AbovegroundPool"
	c := vOld.DeepCopy()

	c.Spec.ResourcePool = "IngroundPool"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestManagementWorkersVSphereMachineValidateUpdateResourcePoolSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.ResourcePool = "AbovegroundPool"
	c := vOld.DeepCopy()

	c.Spec.ResourcePool = "IngroundPool"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestWorkloadWorkersVSphereMachineValidateUpdateResourcePoolSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.ResourcePool = "AbovegroundPool"
	c := vOld.DeepCopy()

	c.Spec.ResourcePool = "IngroundPool"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestVSphereMachineValidateUpdateStoragePolicyImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.StoragePolicyName = "Space-Inefficient"
	c := vOld.DeepCopy()

	c.Spec.StoragePolicyName = "Space-Efficient"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestVSphereMachineConfigValidateCreateSuccess(t *testing.T) {
	config := vsphereMachineConfig()

	g := NewWithT(t)
	g.Expect(config.ValidateCreate()).To(Succeed())
}

func TestVSphereMachineConfigValidateCreateResourcePoolNotSet(t *testing.T) {
	config := vsphereMachineConfig()
	config.Spec.ResourcePool = ""

	g := NewWithT(t)
	g.Expect(config.ValidateCreate()).To(MatchError(ContainSubstring("resourcePool is not set or is empty")))
}

func TestVSphereMachineConfigSetDefaults(t *testing.T) {
	g := NewWithT(t)

	sOld := vsphereMachineConfig()
	sOld.Spec.OSFamily = ""
	sOld.Default()

	g.Expect(sOld.Spec.MemoryMiB).To(Equal(8192))
	g.Expect(sOld.Spec.NumCPUs).To(Equal(2))
	g.Expect(sOld.Spec.OSFamily).To(Equal(v1alpha1.Bottlerocket))
	g.Expect(sOld.Spec.Users).To(Equal([]v1alpha1.UserConfiguration{{Name: "ec2-user", SshAuthorizedKeys: []string{""}}}))
}

func vsphereMachineConfig() v1alpha1.VSphereMachineConfig {
	return v1alpha1.VSphereMachineConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 2)},
		Spec: v1alpha1.VSphereMachineConfigSpec{
			ResourcePool: "my-resourcePool",
			Datastore:    "my-datastore",
			OSFamily:     "ubuntu",
		},
		Status: v1alpha1.VSphereMachineConfigStatus{},
	}
}
