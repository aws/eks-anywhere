package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestTinkerbellMachineConfigValidateCreateSuccess(t *testing.T) {
	machineConfig := v1alpha1.CreateTinkerbellMachineConfig()

	g := NewWithT(t)
	g.Expect(machineConfig.ValidateCreate()).To(Succeed())
}

func TestTinkerbellMachineConfigValidateCreateFail(t *testing.T) {
	machineConfig := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.HardwareSelector = nil
	})

	g := NewWithT(t)
	g.Expect(machineConfig.ValidateCreate()).NotTo(Succeed())
}

func TestTinkerbellMachineConfigValidateCreateFailNoUsers(t *testing.T) {
	machineConfig := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.Users = []v1alpha1.UserConfiguration{}
	})

	g := NewWithT(t)
	g.Expect(machineConfig.ValidateCreate()).NotTo(Succeed())
}

func TestTinkerbellMachineConfigValidateUpdateSucceed(t *testing.T) {
	machineConfigOld := v1alpha1.CreateTinkerbellMachineConfig()
	machineConfigNew := machineConfigOld.DeepCopy()

	g := NewWithT(t)
	g.Expect(machineConfigNew.ValidateUpdate(machineConfigOld)).To(Succeed())
}

func TestTinkerbellMachineConfigValidateUpdateFailOldMachineConfig(t *testing.T) {
	machineConfigOld := &v1alpha1.TinkerbellDatacenterConfig{}
	machineConfigNew := v1alpha1.CreateTinkerbellMachineConfig()

	g := NewWithT(t)
	g.Expect(machineConfigNew.ValidateUpdate(machineConfigOld)).To(MatchError(ContainSubstring("expected a TinkerbellMachineConfig but got a *v1alpha1.TinkerbellDatacenterConfig")))
}

func TestTinkerbellMachineConfigValidateUpdateFailOSFamily(t *testing.T) {
	machineConfigOld := v1alpha1.CreateTinkerbellMachineConfig()
	machineConfigNew := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.OSFamily = v1alpha1.Bottlerocket
	})

	g := NewWithT(t)
	err := machineConfigNew.ValidateUpdate(machineConfigOld)
	g.Expect(err).NotTo(BeNil())
	g.Expect(HaveField("spec.OSFamily", err))
}

func TestTinkerbellMachineConfigValidateUpdateFailLenSshAuthorizedKeys(t *testing.T) {
	machineConfigOld := v1alpha1.CreateTinkerbellMachineConfig()
	machineConfigNew := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.Users = []v1alpha1.UserConfiguration{{
			Name:              "mySshUsername",
			SshAuthorizedKeys: []string{},
		}}
	})

	g := NewWithT(t)
	err := machineConfigNew.ValidateUpdate(machineConfigOld)
	g.Expect(err).NotTo(BeNil())
	g.Expect(HaveField("Users[0].SshAuthorizedKeys", err))
}

func TestTinkerbellMachineConfigValidateUpdateFailSshAuthorizedKeys(t *testing.T) {
	machineConfigOld := v1alpha1.CreateTinkerbellMachineConfig()
	machineConfigNew := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.Users = []v1alpha1.UserConfiguration{{
			Name:              "mySshUsername",
			SshAuthorizedKeys: []string{"mySshAuthorizedKey1"},
		}}
	})

	g := NewWithT(t)
	err := machineConfigNew.ValidateUpdate(machineConfigOld)
	g.Expect(err).NotTo(BeNil())
	g.Expect(HaveField("Users[0].SshAuthorizedKeys[0]", err))
}

func TestTinkerbellMachineConfigValidateUpdateFailUsersLen(t *testing.T) {
	machineConfigOld := v1alpha1.CreateTinkerbellMachineConfig()
	machineConfigNew := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.Users = []v1alpha1.UserConfiguration{
			{
				Name:              "mySshUsername1",
				SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
			},
			{
				Name:              "mySshUsername2",
				SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
			},
		}
	})

	g := NewWithT(t)
	err := machineConfigNew.ValidateUpdate(machineConfigOld)
	g.Expect(err).NotTo(BeNil())
	g.Expect(HaveField("Users", err))
}

func TestTinkerbellMachineConfigDefaultOSFamily(t *testing.T) {
	mOld := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.OSFamily = ""
	})

	mOld.Default()
	g := NewWithT(t)
	g.Expect(mOld.Spec.OSFamily).To(Equal(v1alpha1.Bottlerocket))
}

func TestTinkerbellMachineConfigValidateUpdateFailUsers(t *testing.T) {
	machineConfigOld := v1alpha1.CreateTinkerbellMachineConfig()
	machineConfigNew := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.Users = []v1alpha1.UserConfiguration{{
			Name:              "mySshUsername1",
			SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
		}}
	})

	g := NewWithT(t)
	err := machineConfigNew.ValidateUpdate(machineConfigOld)
	g.Expect(err).NotTo(BeNil())
	g.Expect(HaveField("Users[0].Name", err))
}

func TestTinkerbellMachineConfigValidateUpdateFailHardwareSelector(t *testing.T) {
	machineConfigOld := v1alpha1.CreateTinkerbellMachineConfig()
	machineConfigNew := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.HardwareSelector = map[string]string{
			"type2": "cp2",
		}
	})

	g := NewWithT(t)
	err := machineConfigNew.ValidateUpdate(machineConfigOld)
	g.Expect(err).NotTo(BeNil())
	g.Expect(HaveField("HardwareSelector", err))
}
