package snow_test

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

const (
	sshKey = "ssh-rsa ABCDE"
)

func TestGenerateDefaultSSHKeysExists(t *testing.T) {
	g := newConfigManagerTest(t)
	err := g.defaulters.GenerateDefaultSSHKeys(g.ctx, map[string]*v1alpha1.SnowMachineConfig{g.machineConfig.Name: g.machineConfig}, g.clusterName)
	g.Expect(err).To(Succeed())
}

func TestGenerateDefaultSSHKeysError(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.SshKeyName = ""
	g.keyGenerator.EXPECT().GenerateSSHAuthKey(g.writer).Return(sshKey, nil)
	g.aws.EXPECT().EC2KeyNameExists(g.ctx, g.defaultKeyName).Return(false, errors.New("test error"))
	err := g.defaulters.GenerateDefaultSSHKeys(g.ctx, map[string]*v1alpha1.SnowMachineConfig{g.machineConfig.Name: g.machineConfig}, g.clusterName)
	g.Expect(err).NotTo(Succeed())
}

func TestGenerateDefaultSSHKeysGenerated(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.SshKeyName = ""
	secondMachineConfig := g.machineConfig.DeepCopy()
	secondMachineConfig.Name = g.machineConfig.Name + "-2"
	g.keyGenerator.EXPECT().GenerateSSHAuthKey(g.writer).Return(sshKey, nil)
	g.aws.EXPECT().EC2KeyNameExists(g.ctx, g.defaultKeyName).Return(false, nil).Times(2)
	g.aws.EXPECT().EC2KeyNameExists(g.ctx, g.defaultKeyName).Return(true, nil).Times(2)
	g.aws.EXPECT().EC2ImportKeyPair(g.ctx, g.defaultKeyName, []byte(sshKey)).Return(nil).Times(2)
	err := g.defaulters.GenerateDefaultSSHKeys(g.ctx, map[string]*v1alpha1.SnowMachineConfig{
		g.machineConfig.Name:     g.machineConfig,
		secondMachineConfig.Name: secondMachineConfig,
	}, g.clusterName)
	g.Expect(err).To(Succeed())
}

func TestSetDefaultSSHKey(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.SshKeyName = ""
	g.keyGenerator.EXPECT().GenerateSSHAuthKey(g.writer).Return(sshKey, nil)
	g.aws.EXPECT().EC2KeyNameExists(g.ctx, g.defaultKeyName).Return(false, nil).Times(2)
	g.aws.EXPECT().EC2ImportKeyPair(g.ctx, g.defaultKeyName, []byte(sshKey)).Return(nil).Times(2)
	err := g.machineConfigDefaulters.SetupDefaultSSHKey(g.ctx, g.machineConfig, g.clusterName)
	g.Expect(err).To(Succeed())
}

func TestSetDefaultSSHKeyExistsOnAllDevices(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.SshKeyName = ""
	g.keyGenerator.EXPECT().GenerateSSHAuthKey(g.writer).Return(sshKey, nil)
	g.aws.EXPECT().EC2KeyNameExists(g.ctx, g.defaultKeyName).Return(true, nil).Times(2)
	err := g.machineConfigDefaulters.SetupDefaultSSHKey(g.ctx, g.machineConfig, g.clusterName)
	g.Expect(g.machineConfig.Spec.SshKeyName).To(Equal(g.defaultKeyName))
	g.Expect(err).To(Succeed())
}

func TestSetDefaultSSHKeyExistsOnPartialDevices(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.SshKeyName = ""
	g.keyGenerator.EXPECT().GenerateSSHAuthKey(g.writer).Return(sshKey, nil)
	g.aws.EXPECT().EC2KeyNameExists(g.ctx, g.defaultKeyName).Return(true, nil)
	g.aws.EXPECT().EC2KeyNameExists(g.ctx, g.defaultKeyName).Return(false, nil)
	g.aws.EXPECT().EC2ImportKeyPair(g.ctx, g.defaultKeyName, []byte(sshKey)).Return(nil)
	err := g.machineConfigDefaulters.SetupDefaultSSHKey(g.ctx, g.machineConfig, g.clusterName)
	g.Expect(err).To(Succeed())
}

func TestSetDefaultSSHKeyImportKeyError(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.SshKeyName = ""
	g.keyGenerator.EXPECT().GenerateSSHAuthKey(g.writer).Return(sshKey, nil)
	g.aws.EXPECT().EC2KeyNameExists(g.ctx, g.defaultKeyName).Return(false, nil)
	g.aws.EXPECT().EC2ImportKeyPair(g.ctx, g.defaultKeyName, []byte(sshKey)).Return(errors.New("error"))
	err := g.machineConfigDefaulters.SetupDefaultSSHKey(g.ctx, g.machineConfig, g.clusterName)
	g.Expect(err).NotTo(Succeed())
}

func TestSetDefaultSSHKeyClientMapError(t *testing.T) {
	g := newConfigManagerTestClientMapError(t)
	g.machineConfig.Spec.SshKeyName = ""
	err := g.machineConfigDefaulters.SetupDefaultSSHKey(g.ctx, g.machineConfig, g.clusterName)
	g.Expect(err).NotTo(Succeed())
}

func TestSetDefaultSSHKeyDeviceNotFoundInClientMap(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.SshKeyName = ""
	g.machineConfig.Spec.Devices = []string{"device-not-exist"}
	g.keyGenerator.EXPECT().GenerateSSHAuthKey(g.writer).Return(sshKey, nil)
	err := g.machineConfigDefaulters.SetupDefaultSSHKey(g.ctx, g.machineConfig, g.clusterName)
	g.Expect(err).To(MatchError(ContainSubstring("credentials not found for device")))
}
