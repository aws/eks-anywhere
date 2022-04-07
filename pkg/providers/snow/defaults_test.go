package snow_test

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
)

func TestSetDefaultSshKey(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.SshKeyName = ""
	wantKey := "eksa-default-cp-machine.pem"
	wantVal := "pem val"
	g.aws.EXPECT().EC2CreateKeyPair(g.ctx, wantKey).Return(wantVal, nil)
	k, v, err := g.defaulters.GenerateDefaultSshKey(g.ctx, g.machineConfig)
	g.Expect(err).To(Succeed())
	g.Expect(k).To(Equal(wantKey))
	g.Expect(v).To(Equal(wantVal))
}

func TestSetDefaultSshKeySkip(t *testing.T) {
	g := newConfigManagerTest(t)
	k, v, err := g.defaulters.GenerateDefaultSshKey(g.ctx, g.machineConfig)
	g.Expect(err).To(Succeed())
	g.Expect(k).To(Equal(""))
	g.Expect(v).To(Equal(""))
}

func TestSetDefaultSshKeyError(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.SshKeyName = ""
	g.aws.EXPECT().EC2CreateKeyPair(g.ctx, "eksa-default-cp-machine.pem").Return("v", errors.New("error"))
	k, v, err := g.defaulters.GenerateDefaultSshKey(g.ctx, g.machineConfig)
	g.Expect(err).NotTo(Succeed())
	g.Expect(k).To(Equal(""))
	g.Expect(v).To(Equal(""))
}
