package v1alpha1_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestTinkerbellMachineConfigValidateCreateSuccess(t *testing.T) {
	ctx := context.Background()
	machineConfig := v1alpha1.CreateTinkerbellMachineConfig()

	g := NewWithT(t)
	g.Expect(machineConfig.ValidateCreate(ctx, machineConfig)).Error().To(Succeed())
}

func TestTinkerbellMachineConfigValidateCreateFail(t *testing.T) {
	ctx := context.Background()
	machineConfig := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.HardwareSelector = nil
	})

	g := NewWithT(t)
	g.Expect(machineConfig.ValidateCreate(ctx, machineConfig)).Error().To(MatchError(ContainSubstring("TinkerbellMachineConfig: missing spec.hardwareSelector: tinkerbellmachineconfig")))
}

func TestTinkerbellMachineConfigValidateCreateFailNoUsers(t *testing.T) {
	ctx := context.Background()
	machineConfig := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.Users = []v1alpha1.UserConfiguration{}
	})

	g := NewWithT(t)
	g.Expect(machineConfig.ValidateCreate(ctx, machineConfig)).Error().To(MatchError(ContainSubstring("TinkerbellMachineConfig: missing spec.Users: tinkerbellmachineconfig")))
}

func TestTinkerbellMachineConfigValidateCreateFailNoSSHkeys(t *testing.T) {
	ctx := context.Background()
	machineConfig := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.Users = []v1alpha1.UserConfiguration{
			{
				Name: "test",
			},
		}
	})

	g := NewWithT(t)
	g.Expect(machineConfig.ValidateCreate(ctx, machineConfig)).Error().To(MatchError(ContainSubstring("Please specify a ssh authorized key")))
}

func TestTinkerbellMachineConfigValidateCreateFailEmptySSHkeys(t *testing.T) {
	ctx := context.Background()
	machineConfig := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.Users = []v1alpha1.UserConfiguration{
			{
				Name:              "test",
				SshAuthorizedKeys: []string{},
			},
		}
	})

	g := NewWithT(t)
	g.Expect(machineConfig.ValidateCreate(ctx, machineConfig)).Error().To(MatchError(ContainSubstring("Please specify a ssh authorized key")))
}

func TestTinkerbellMachineConfigValidateUpdateSucceed(t *testing.T) {
	ctx := context.Background()
	machineConfigOld := v1alpha1.CreateTinkerbellMachineConfig()
	machineConfigNew := machineConfigOld.DeepCopy()

	g := NewWithT(t)
	g.Expect(machineConfigNew.ValidateUpdate(ctx, machineConfigNew, machineConfigOld)).Error().To(Succeed())
}

func TestTinkerbellMachineConfigValidateUpdateFailOldMachineConfig(t *testing.T) {
	ctx := context.Background()
	machineConfigOld := &v1alpha1.TinkerbellDatacenterConfig{}
	machineConfigNew := v1alpha1.CreateTinkerbellMachineConfig()

	g := NewWithT(t)
	g.Expect(machineConfigNew.ValidateUpdate(ctx, machineConfigNew, machineConfigOld)).Error().To(MatchError(ContainSubstring("expected a TinkerbellMachineConfig but got a *v1alpha1.TinkerbellDatacenterConfig")))
}

func TestTinkerbellMachineConfigValidateUpdateFailOSFamily(t *testing.T) {
	ctx := context.Background()
	machineConfigOld := v1alpha1.CreateTinkerbellMachineConfig()
	machineConfigNew := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.OSFamily = v1alpha1.Bottlerocket
	})

	g := NewWithT(t)
	_, err := machineConfigNew.ValidateUpdate(ctx, machineConfigNew, machineConfigOld)
	g.Expect(err).NotTo(BeNil())
	g.Expect(HaveField("spec.OSFamily", err))
}

func TestTinkerbellMachineConfigValidateUpdateFailLenSshAuthorizedKeys(t *testing.T) {
	ctx := context.Background()
	machineConfigOld := v1alpha1.CreateTinkerbellMachineConfig()
	machineConfigNew := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.Users = []v1alpha1.UserConfiguration{{
			Name:              "mySshUsername",
			SshAuthorizedKeys: []string{},
		}}
	})

	g := NewWithT(t)
	_, err := machineConfigNew.ValidateUpdate(ctx, machineConfigNew, machineConfigOld)
	g.Expect(err).NotTo(BeNil())
	g.Expect(HaveField("Users[0].SshAuthorizedKeys", err))
}

func TestTinkerbellMachineConfigValidateUpdateFailSshAuthorizedKeys(t *testing.T) {
	ctx := context.Background()
	machineConfigOld := v1alpha1.CreateTinkerbellMachineConfig()
	machineConfigNew := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.Users = []v1alpha1.UserConfiguration{{
			Name:              "mySshUsername",
			SshAuthorizedKeys: []string{"mySshAuthorizedKey1"},
		}}
	})

	g := NewWithT(t)
	_, err := machineConfigNew.ValidateUpdate(ctx, machineConfigNew, machineConfigOld)
	g.Expect(err).NotTo(BeNil())
	g.Expect(HaveField("Users[0].SshAuthorizedKeys[0]", err))
}

func TestTinkerbellMachineConfigValidateUpdateFailUsersLen(t *testing.T) {
	ctx := context.Background()
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
	_, err := machineConfigNew.ValidateUpdate(ctx, machineConfigNew, machineConfigOld)
	g.Expect(err).NotTo(BeNil())
	g.Expect(HaveField("Users", err))
}

func TestTinkerbellMachineConfigDefaultOSFamily(t *testing.T) {
	ctx := context.Background()
	mOld := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.OSFamily = ""
	})

	err := mOld.Default(ctx, mOld)
	g := NewWithT(t)
	g.Expect(err).To(BeNil())
	g.Expect(mOld.Spec.OSFamily).To(Equal(v1alpha1.Bottlerocket))
}

func TestTinkerbellMachineConfigMutateSSHKey(t *testing.T) {
	ctx := context.Background()
	sshKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGuWn+GtgUe/g85l4SqSsGCV56CXZzqktKX/hYAl7MwO"
	mOld := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.Users = []v1alpha1.UserConfiguration{
			{
				Name:              "user",
				SshAuthorizedKeys: []string{fmt.Sprintf("%s abc@xyz.com", sshKey)},
			},
		}
	})

	err := mOld.Default(ctx, mOld)
	g := NewWithT(t)
	g.Expect(err).To(BeNil())
	g.Expect(mOld.Spec.Users[0].SshAuthorizedKeys[0]).To(Equal(sshKey))
}

func TestTinkerbellMachineConfigMutateSSHKeyNotMutated(t *testing.T) {
	ctx := context.Background()
	sshKey := "ssh incorrect Key abc@xyz.com"
	mOld := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.Users = []v1alpha1.UserConfiguration{
			{
				Name:              "user",
				SshAuthorizedKeys: []string{sshKey},
			},
		}
	})

	err := mOld.Default(ctx, mOld)
	g := NewWithT(t)
	g.Expect(err).To(BeNil())
	g.Expect(mOld.Spec.Users[0].SshAuthorizedKeys[0]).To(Equal(sshKey))
}

func TestTinkerbellMachineConfigValidateUpdateFailUsers(t *testing.T) {
	ctx := context.Background()
	machineConfigOld := v1alpha1.CreateTinkerbellMachineConfig()
	machineConfigNew := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.Users = []v1alpha1.UserConfiguration{{
			Name:              "mySshUsername1",
			SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
		}}
	})

	g := NewWithT(t)
	_, err := machineConfigNew.ValidateUpdate(ctx, machineConfigNew, machineConfigOld)
	g.Expect(err).NotTo(BeNil())
	g.Expect(HaveField("Users[0].Name", err))
}

func TestTinkerbellMachineConfigValidateUpdateFailHardwareSelector(t *testing.T) {
	ctx := context.Background()
	machineConfigOld := v1alpha1.CreateTinkerbellMachineConfig()
	machineConfigNew := v1alpha1.CreateTinkerbellMachineConfig(func(mc *v1alpha1.TinkerbellMachineConfig) {
		mc.Spec.HardwareSelector = map[string]string{
			"type2": "cp2",
		}
	})

	g := NewWithT(t)
	_, err := machineConfigNew.ValidateUpdate(ctx, machineConfigNew, machineConfigOld)
	g.Expect(err).NotTo(BeNil())
	g.Expect(HaveField("HardwareSelector", err))
}

func TestTinkerbellMachineConfigDefaultCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomDefaulter
	config := &v1alpha1.TinkerbellMachineConfig{}

	// Call Default with the wrong type
	err := config.Default(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a TinkerbellMachineConfig"))
}

func TestTinkerbellMachineConfigValidateCreateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.TinkerbellMachineConfig{}

	// Call ValidateCreate with the wrong type
	warnings, err := config.ValidateCreate(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a TinkerbellMachineConfig"))
}

func TestTinkerbellMachineConfigValidateUpdateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.TinkerbellMachineConfig{}

	// Call ValidateUpdate with the wrong type
	warnings, err := config.ValidateUpdate(context.TODO(), wrongType, &v1alpha1.TinkerbellMachineConfig{})

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a TinkerbellMachineConfig"))
}

func TestTinkerbellMachineConfigValidateDeleteCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.TinkerbellMachineConfig{}

	// Call ValidateDelete with the wrong type
	warnings, err := config.ValidateDelete(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a TinkerbellMachineConfig"))
}
