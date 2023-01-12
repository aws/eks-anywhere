package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestTinkerbellMachineConfig_ValidateCreateSuccess(t *testing.T) {
	machineConfig := createTinkerbellMachineConfig()

	g := NewWithT(t)
	g.Expect(machineConfig.ValidateCreate()).To(Succeed())
}

func TestTinkerbellMachineConfig_ValidateCreateFail(t *testing.T) {
	machineConfig := createTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
		mc.Spec.HardwareSelector = nil
	})

	g := NewWithT(t)
	g.Expect(machineConfig.ValidateCreate()).NotTo(Succeed())
}

func TestTinkerbellMachineConfig_ValidateUpdateSucceed(t *testing.T) {
	machineConfigOld := createTinkerbellMachineConfig()
	machineConfigNew := machineConfigOld.DeepCopy()

	g := NewWithT(t)
	g.Expect(machineConfigNew.ValidateUpdate(machineConfigOld)).To(Succeed())
}

func TestTinkerbellMachineConfig_ValidateUpdateFailOSFamily(t *testing.T) {
	machineConfigOld := createTinkerbellMachineConfig()
	machineConfigNew := createTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
		mc.Spec.OSFamily = Bottlerocket
	})

	g := NewWithT(t)
	g.Expect(machineConfigNew.ValidateUpdate(machineConfigOld)).NotTo(Succeed())
}

func TestTinkerbellMachineConfig_ValidateUpdateFailSshAuthorizedKeys(t *testing.T) {
	machineConfigOld := createTinkerbellMachineConfig()
	machineConfigNew := createTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
		mc.Spec.Users = []UserConfiguration{{
			Name:              "mySshUsername",
			SshAuthorizedKeys: []string{"mySshAuthorizedKey1"},
		}}
	})

	g := NewWithT(t)
	g.Expect(machineConfigNew.ValidateUpdate(machineConfigOld)).NotTo(Succeed())
}

func TestTinkerbellMachineConfig_ValidateUpdateFailUsers(t *testing.T) {
	machineConfigOld := createTinkerbellMachineConfig()
	machineConfigNew := createTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
		mc.Spec.Users = []UserConfiguration{{
			Name:              "mySshUsername1",
			SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
		}}
	})

	g := NewWithT(t)
	g.Expect(machineConfigNew.ValidateUpdate(machineConfigOld)).NotTo(Succeed())
}

func TestTinkerbellMachineConfig_ValidateUpdateFailHardwareSelector(t *testing.T) {
	machineConfigOld := createTinkerbellMachineConfig()
	machineConfigNew := createTinkerbellMachineConfig(func(mc *TinkerbellMachineConfig) {
		mc.Spec.HardwareSelector = map[string]string{
			"type2": "cp2",
		}
	})

	g := NewWithT(t)
	g.Expect(machineConfigNew.ValidateUpdate(machineConfigOld)).NotTo(Succeed())
}
