package snow_test

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
)

const (
	sshKey = "ssh-rsa ABCDE"
)

func TestSetDefaultSshKey(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.SshKeyName = ""
	g.keyGenerator.EXPECT().GenerateSSHAuthKey(g.writer).Return(sshKey, nil)
	g.aws.EXPECT().EC2KeyNameExists(g.ctx, "eksa-default").Return(false, nil).Times(2)
	g.aws.EXPECT().EC2ImportKeyPair(g.ctx, "eksa-default", []byte(sshKey)).Return(nil).Times(2)
	err := g.machineConfigDefaulters.SetupDefaultSshKey(g.ctx, g.machineConfig)
	g.Expect(err).To(Succeed())
}

func TestSetDefaultSshKeyExistsOnAllDevices(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.SshKeyName = ""
	g.aws.EXPECT().EC2KeyNameExists(g.ctx, "eksa-default").Return(true, nil).Times(2)
	err := g.machineConfigDefaulters.SetupDefaultSshKey(g.ctx, g.machineConfig)
	g.Expect(g.machineConfig.Spec.SshKeyName).To(Equal("eksa-default"))
	g.Expect(err).To(Succeed())
}

func TestSetDefaultSshKeyExistsOnPartialDevices(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.SshKeyName = ""
	g.aws.EXPECT().EC2KeyNameExists(g.ctx, "eksa-default").Return(true, nil)
	g.aws.EXPECT().EC2KeyNameExists(g.ctx, "eksa-default").Return(false, nil)
	err := g.machineConfigDefaulters.SetupDefaultSshKey(g.ctx, g.machineConfig)
	g.Expect(err).To((MatchError(ContainSubstring("default key [keyName=eksa-default] only exists on some of the devices"))))
}

func TestSetDefaultSshKeySkip(t *testing.T) {
	g := newConfigManagerTest(t)
	err := g.machineConfigDefaulters.SetupDefaultSshKey(g.ctx, g.machineConfig)
	g.Expect(err).To(Succeed())
}

func TestSetDefaultSshKeyImportKeyError(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.SshKeyName = ""
	g.keyGenerator.EXPECT().GenerateSSHAuthKey(g.writer).Return(sshKey, nil)
	g.aws.EXPECT().EC2KeyNameExists(g.ctx, "eksa-default").Return(false, nil).Times(2)
	g.aws.EXPECT().EC2ImportKeyPair(g.ctx, "eksa-default", []byte(sshKey)).Return(errors.New("error"))
	err := g.machineConfigDefaulters.SetupDefaultSshKey(g.ctx, g.machineConfig)
	g.Expect(err).NotTo(Succeed())
}

func TestSetDefaultSshKeyClientMapError(t *testing.T) {
	g := newConfigManagerTestClientMapError(t)
	g.machineConfig.Spec.SshKeyName = ""
	err := g.machineConfigDefaulters.SetupDefaultSshKey(g.ctx, g.machineConfig)
	g.Expect(err).NotTo(Succeed())
}

func TestSetDefaultSshKeyDeviceNotFoundInClientMap(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.SshKeyName = ""
	g.machineConfig.Spec.Devices = []string{"device-not-exist"}
	err := g.machineConfigDefaulters.SetupDefaultSshKey(g.ctx, g.machineConfig)
	g.Expect(err).To(MatchError(ContainSubstring("credentials not found for device")))
}
