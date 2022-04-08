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
	g.aws.EXPECT().EC2ImportKeyPair(g.ctx, "eksa-default", []byte(sshKey)).Return(nil).Times(2)
	err := g.machineConfigDefaulters.SetupDefaultSshKey(g.ctx, g.machineConfig)
	g.Expect(err).To(Succeed())
}

func TestSetDefaultSshKeySkip(t *testing.T) {
	g := newConfigManagerTest(t)
	err := g.machineConfigDefaulters.SetupDefaultSshKey(g.ctx, g.machineConfig)
	g.Expect(err).To(Succeed())
}

func TestSetDefaultSshKeyError(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.SshKeyName = ""
	g.keyGenerator.EXPECT().GenerateSSHAuthKey(g.writer).Return(sshKey, nil)
	g.aws.EXPECT().EC2ImportKeyPair(g.ctx, "eksa-default", []byte(sshKey)).Return(errors.New("error"))
	err := g.machineConfigDefaulters.SetupDefaultSshKey(g.ctx, g.machineConfig)
	g.Expect(err).NotTo(Succeed())
}
