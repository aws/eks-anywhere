package v1alpha1_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestManagementCPVSphereMachineValidateUpdateTemplateMutableManagement(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Template = "oldTemplate"
	c := vOld.DeepCopy()

	c.Spec.Template = "newTemplate"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadCPVSphereMachineValidateUpdateTemplateSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Template = "oldTemplate"
	c := vOld.DeepCopy()

	c.Spec.Template = "newTemplate"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementWorkersVSphereMachineValidateUpdateTemplateSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.Template = "oldTemplate"
	c := vOld.DeepCopy()

	c.Spec.Template = "newTemplate"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadWorkersVSphereMachineValidateUpdateTemplateSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Template = "oldTemplate"
	c := vOld.DeepCopy()

	c.Spec.Template = "newTemplate"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementEtcdVSphereMachineValidateEtcdUpdateTemplateMutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.Spec.Template = "oldTemplate"
	c := vOld.DeepCopy()

	c.Spec.Template = "newTemplate"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadEtcdVSphereMachineValidateUpdateTemplateSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Template = "oldTemplate"
	c := vOld.DeepCopy()

	c.Spec.Template = "newTemplate"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestVSphereMachineValidateUpdateOSFamilyImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.OSFamily = v1alpha1.Ubuntu
	c := vOld.DeepCopy()

	c.Spec.OSFamily = v1alpha1.Bottlerocket
	c.Spec.Users[0].Name = "ec2-user"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(MatchError(ContainSubstring("spec.osFamily: Forbidden: field is immutable")))
}

func TestManagementCPVSphereMachineValidateUpdateMemoryMiBMutableManagement(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.MemoryMiB = 2
	c := vOld.DeepCopy()

	c.Spec.MemoryMiB = 2000000
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadCPVSphereMachineValidateUpdateMemoryMiBSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.MemoryMiB = 2
	c := vOld.DeepCopy()

	c.Spec.MemoryMiB = 2000000
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementWorkersVSphereMachineValidateUpdateMemoryMiBSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.MemoryMiB = 2
	c := vOld.DeepCopy()

	c.Spec.MemoryMiB = 2000000
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadWorkersVSphereMachineValidateUpdateMemoryMiBSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.MemoryMiB = 2
	c := vOld.DeepCopy()

	c.Spec.MemoryMiB = 2000000
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementEtcdVSphereMachineValidateUpdateMemoryMiBImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.Spec.MemoryMiB = 2

	c := vOld.DeepCopy()

	c.Spec.MemoryMiB = 2000000
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadEtcdVSphereMachineValidateUpdateMemoryMiBSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.MemoryMiB = 2
	c := vOld.DeepCopy()

	c.Spec.MemoryMiB = 2000000
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementCPVSphereMachineValidateUpdateNumCPUsMutableManagement(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.NumCPUs = 1
	c := vOld.DeepCopy()

	c.Spec.NumCPUs = 16
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadCPVSphereMachineValidateUpdateNumCPUsSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.NumCPUs = 1
	c := vOld.DeepCopy()

	c.Spec.NumCPUs = 16
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementWorkersVSphereMachineValidateUpdateNumCPUsSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.NumCPUs = 1
	c := vOld.DeepCopy()

	c.Spec.NumCPUs = 16
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadWorkersVSphereMachineValidateUpdateNumCPUsSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.NumCPUs = 1
	c := vOld.DeepCopy()

	c.Spec.NumCPUs = 16
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementEtcdVSphereMachineValidateUpdateNumCPUsMmutableManagement(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.Spec.NumCPUs = 1
	c := vOld.DeepCopy()

	c.Spec.NumCPUs = 16
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadEtcdVSphereMachineValidateUpdateNumCPUsSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.NumCPUs = 1
	c := vOld.DeepCopy()

	c.Spec.NumCPUs = 16
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementCPVSphereMachineValidateUpdateDiskGiBMutableManagement(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.DiskGiB = 1
	c := vOld.DeepCopy()

	c.Spec.DiskGiB = 160
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadCPVSphereMachineValidateUpdateDiskGiBSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.DiskGiB = 1
	c := vOld.DeepCopy()

	c.Spec.DiskGiB = 160
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementWorkersVSphereMachineValidateUpdateDiskGiBSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.DiskGiB = 1
	c := vOld.DeepCopy()

	c.Spec.DiskGiB = 160
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadWorkersVSphereMachineValidateUpdateDiskGiBSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.DiskGiB = 1
	c := vOld.DeepCopy()

	c.Spec.DiskGiB = 160
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementEtcdVSphereMachineValidateUpdateDiskGiBMmutableManagement(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.Spec.DiskGiB = 1

	c := vOld.DeepCopy()

	c.Spec.DiskGiB = 160
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadEtcdVSphereMachineValidateUpdateDiskGiBSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.DiskGiB = 1
	c := vOld.DeepCopy()

	c.Spec.DiskGiB = 160
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementControlPlaneSphereMachineValidateUpdateSshAuthorizedKeyMutableManagement(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	vOld.Spec.Users[0].SshAuthorizedKeys = []string{"rsa-blahdeblahbalh"}
	c := vOld.DeepCopy()

	c.Spec.Users[0].SshAuthorizedKeys[0] = "rsa-laDeLala"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadControlPlaneVSphereMachineValidateUpdateSshAuthorizedKeyMutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	vOld.Spec.Users[0].SshAuthorizedKeys = []string{"rsa-blahdeblahbalh"}
	c := vOld.DeepCopy()

	c.Spec.Users[0].SshAuthorizedKeys[0] = "rsa-laDeLala"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementControlPlaneVSphereMachineValidateUpdateSshUsernameImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Users[0].Name = "Jeff"
	c := vOld.DeepCopy()

	c.Spec.Users[0].Name = "Andy"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadControlPlaneVSphereMachineValidateUpdateSshUsernameMutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Users[0].Name = "Jeff"
	c := vOld.DeepCopy()

	c.Spec.Users[0].Name = "Andy"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
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
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementEtcdSphereMachineValidateUpdateSshAuthorizedKeyMutableManagement(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	vOld.Spec.Users[0].SshAuthorizedKeys = []string{"rsa-blahdeblahbalh"}
	c := vOld.DeepCopy()

	c.Spec.Users[0].SshAuthorizedKeys[0] = "rsa-laDeLala"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
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
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementEtcdVSphereMachineValidateUpdateSshUsernameMutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.Spec.Users[0].Name = "Jeff"
	c := vOld.DeepCopy()

	c.Spec.Users[0].Name = "Andy"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadEtcdVSphereMachineValidateUpdateSshUsernameMutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Users[0].Name = "Jeff"
	c := vOld.DeepCopy()

	c.Spec.Users[0].Name = "Andy"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementWorkerNodeSphereMachineValidateUpdateSshAuthorizedKeyMutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	vOld.Spec.Users[0].SshAuthorizedKeys = []string{"rsa-blahdeblahbalh"}
	c := vOld.DeepCopy()

	c.Spec.Users[0].SshAuthorizedKeys[0] = "rsa-laDeLala"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadWorkerNodeVSphereMachineValidateUpdateSshAuthorizedKeyMutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Users = []v1alpha1.UserConfiguration{{Name: "Jeff"}}
	vOld.Spec.Users[0].SshAuthorizedKeys = []string{"rsa-blahdeblahbalh"}
	c := vOld.DeepCopy()

	c.Spec.Users[0].SshAuthorizedKeys[0] = "rsa-laDeLala"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementWorkerNodeVSphereMachineValidateUpdateSshUsernameMutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.Users[0].Name = "Jeff"
	c := vOld.DeepCopy()

	c.Spec.Users[0].Name = "Andy"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadWorkerNodeVSphereMachineValidateUpdateSshUsernameMutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Users[0].Name = "Jeff"
	c := vOld.DeepCopy()

	c.Spec.Users[0].Name = "Andy"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestVSphereMachineValidateUpdateInvalidType(t *testing.T) {
	vOld := &v1alpha1.Cluster{}
	c := &v1alpha1.VSphereMachineConfig{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, vOld)).Error().To(MatchError(ContainSubstring("expected a VSphereMachineConfig but got a *v1alpha1.Cluster")))
}

func TestVSphereMachineValidateUpdateSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()

	vOld.Spec.NumCPUs = 4
	c := vOld.DeepCopy()
	c.Spec.NumCPUs = 16

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestVSphereMachineValidateUpdateBottleRocketInvalidUserName(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.OSFamily = "bottlerocket"
	c := vOld.DeepCopy()
	c.Spec.Users = []v1alpha1.UserConfiguration{
		{
			Name:              "jeff",
			SshAuthorizedKeys: []string{"ssh AAA..."},
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().ToNot(Succeed())
}

func TestManagementCPVSphereMachineValidateUpdateDatastoreMutableSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Datastore = "OldDataStore"
	c := vOld.DeepCopy()

	c.Spec.Datastore = "NewDataStore"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadCPVSphereMachineValidateUpdateDatastoreSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Datastore = "OldDataStore"
	c := vOld.DeepCopy()

	c.Spec.Datastore = "NewDataStore"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementEtcdVSphereMachineValidateUpdateDatastoreMutableManagement(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.Spec.Datastore = "OldDataStore"
	c := vOld.DeepCopy()

	c.Spec.Datastore = "NewDataStore"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadEtcdVSphereMachineValidateUpdateDatastoreSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Datastore = "OldDataStore"
	c := vOld.DeepCopy()

	c.Spec.Datastore = "NewDataStore"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementWorkersVSphereMachineValidateUpdateDatastoreSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.Datastore = "OldDataStore"
	c := vOld.DeepCopy()

	c.Spec.Datastore = "NewDataStore"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadWorkersVSphereMachineValidateUpdateDatastoreSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Datastore = "OldDataStore"
	c := vOld.DeepCopy()

	c.Spec.Datastore = "NewDataStore"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementCPVSphereMachineValidateUpdateFolderMutableManagement(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.Folder = "/dev/null"
	c := vOld.DeepCopy()

	c.Spec.Folder = "/tmp"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadCPVSphereMachineValidateUpdateFolderSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Folder = "/dev/null"
	c := vOld.DeepCopy()

	c.Spec.Folder = "/tmp"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementEtcdVSphereMachineValidateUpdateFolderMutableManagement(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.Spec.Folder = "/dev/null"
	c := vOld.DeepCopy()

	c.Spec.Folder = "/tmp"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadEtcdVSphereMachineValidateUpdateFolderSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Folder = "/dev/null"
	c := vOld.DeepCopy()

	c.Spec.Folder = "/tmp"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementWorkersVSphereMachineValidateUpdateFolderSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.Folder = "/dev/null"
	c := vOld.DeepCopy()

	c.Spec.Folder = "/tmp"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadWorkersVSphereMachineValidateUpdateFolderSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.Folder = "/dev/null"
	c := vOld.DeepCopy()

	c.Spec.Folder = "/tmp"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementCPVSphereMachineValidateUpdateResourcePoolMutableManagemenmt(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.Spec.ResourcePool = "AbovegroundPool"
	c := vOld.DeepCopy()

	c.Spec.ResourcePool = "IngroundPool"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadCPVSphereMachineValidateUpdateResourcePoolSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetControlPlane()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.ResourcePool = "AbovegroundPool"
	c := vOld.DeepCopy()

	c.Spec.ResourcePool = "IngroundPool"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementEtcdVSphereMachineValidateUpdateResourcePoolMutableManagement(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.Spec.ResourcePool = "AbovegroundPool"
	c := vOld.DeepCopy()

	c.Spec.ResourcePool = "IngroundPool"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadEtcdVSphereMachineValidateUpdateResourcePoolSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetEtcd()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.ResourcePool = "AbovegroundPool"
	c := vOld.DeepCopy()

	c.Spec.ResourcePool = "IngroundPool"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestManagementWorkersVSphereMachineValidateUpdateResourcePoolSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.ResourcePool = "AbovegroundPool"
	c := vOld.DeepCopy()

	c.Spec.ResourcePool = "IngroundPool"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestWorkloadWorkersVSphereMachineValidateUpdateResourcePoolSuccess(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.SetManagedBy("test-cluster")
	vOld.Spec.ResourcePool = "AbovegroundPool"
	c := vOld.DeepCopy()

	c.Spec.ResourcePool = "IngroundPool"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(Succeed())
}

func TestVSphereMachineValidateUpdateStoragePolicyImmutable(t *testing.T) {
	vOld := vsphereMachineConfig()
	vOld.Spec.StoragePolicyName = "Space-Inefficient"
	c := vOld.DeepCopy()

	c.Spec.StoragePolicyName = "Space-Efficient"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(context.TODO(), c, &vOld)).Error().To(MatchError(ContainSubstring("spec.storagePolicyName: Forbidden: field is immutable")))
}

func TestVSphereMachineConfigValidateCreateSuccess(t *testing.T) {
	config := vsphereMachineConfig()

	g := NewWithT(t)
	g.Expect(config.ValidateCreate(context.TODO(), &config)).Error().To(Succeed())
}

func TestVSphereMachineConfigValidateInvalidUserSSHAuthorizedKeys(t *testing.T) {
	config := vsphereMachineConfig()
	config.Spec.Users = []v1alpha1.UserConfiguration{
		{
			Name:              "capv",
			SshAuthorizedKeys: []string{""},
		},
	}
	g := NewWithT(t)
	g.Expect(config.ValidateCreate(context.TODO(), &config)).Error().ToNot(Succeed())
}

func TestVSphereMachineConfigValidateCreateResourcePoolNotSet(t *testing.T) {
	config := vsphereMachineConfig()
	config.Spec.ResourcePool = ""

	g := NewWithT(t)
	g.Expect(config.ValidateCreate(context.TODO(), &config)).Error().To(MatchError(ContainSubstring("resourcePool is not set or is empty")))
}

func TestVSphereMachineConfigValidateCreateTemplateNotSet(t *testing.T) {
	config := vsphereMachineConfig()
	config.Spec.Template = ""

	g := NewWithT(t)
	g.Expect(config.ValidateCreate(context.TODO(), &config)).Error().To(MatchError(ContainSubstring("template field is required")))
}

func TestVSphereMachineConfigSetDefaults(t *testing.T) {
	g := NewWithT(t)

	sOld := vsphereMachineConfig()
	sOld.Spec.OSFamily = ""
	err := sOld.Default(context.TODO(), &sOld)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(sOld.Spec.MemoryMiB).To(Equal(8192))
	g.Expect(sOld.Spec.NumCPUs).To(Equal(2))
	g.Expect(sOld.Spec.OSFamily).To(Equal(v1alpha1.Bottlerocket))
}

func vsphereMachineConfig() v1alpha1.VSphereMachineConfig {
	return v1alpha1.VSphereMachineConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 2)},
		Spec: v1alpha1.VSphereMachineConfigSpec{
			ResourcePool: "my-resourcePool",
			Datastore:    "my-datastore",
			OSFamily:     "ubuntu",
			Template:     "/Datacenter/vm/Templates/bottlerocket-v1.23.12-kubernetes-1-23-eks-7-amd64-d44065e",
			Users: []v1alpha1.UserConfiguration{
				{
					Name:              "capv",
					SshAuthorizedKeys: []string{"ssh AAA..."},
				},
			},
		},
		Status: v1alpha1.VSphereMachineConfigStatus{},
	}
}
func TestVSphereMachineConfigDefaultCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomDefaulter
	config := &v1alpha1.VSphereMachineConfig{}

	// Call Default with the wrong type
	err := config.Default(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a VSphereMachineConfig"))
}

func TestVSphereMachineConfigValidateCreateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.VSphereMachineConfig{}

	// Call ValidateCreate with the wrong type
	warnings, err := config.ValidateCreate(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a VSphereMachineConfig"))
}

func TestVSphereMachineConfigValidateUpdateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.VSphereMachineConfig{}

	// Call ValidateUpdate with the wrong type
	warnings, err := config.ValidateUpdate(context.TODO(), wrongType, &v1alpha1.VSphereMachineConfig{})

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a VSphereMachineConfig"))
}

func TestVSphereMachineConfigValidateDeleteCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.VSphereMachineConfig{}

	// Call ValidateDelete with the wrong type
	warnings, err := config.ValidateDelete(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a VSphereMachineConfig"))
}
