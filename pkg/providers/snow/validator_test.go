package snow

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/aws"
	"github.com/aws/eks-anywhere/pkg/aws/mocks"
)

type configManagerTest struct {
	*WithT
	ctx           context.Context
	aws           *mocks.MockClient
	validator     *Validator
	defaulters    *Defaulters
	machineConfig *v1alpha1.SnowMachineConfig
}

func newConfigManagerTest(t *testing.T) *configManagerTest {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mockaws := mocks.NewMockClient(ctrl)
	awsClients := map[string]aws.Client{
		"device-1": mockaws,
		"device-2": mockaws,
	}
	m := &v1alpha1.SnowMachineConfig{
		ObjectMeta: v1.ObjectMeta{
			Name: "cp-machine",
		},
		Spec: v1alpha1.SnowMachineConfigSpec{
			AMIID:      "ami-1",
			SshKeyName: "default",
		},
	}
	return &configManagerTest{
		WithT:         NewWithT(t),
		ctx:           ctx,
		aws:           mockaws,
		validator:     NewValidator(awsClients),
		defaulters:    NewDefaulters(awsClients, nil),
		machineConfig: m,
	}
}

func TestValidateSshKeyPair(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().KeyPairExists(g.ctx, g.machineConfig.Spec.SshKeyName).Return(true, nil).Times(2)
	err := g.validator.validateSshKeyPair(g.ctx, g.machineConfig)
	g.Expect(err).To(Succeed())
}

func TestValidateSshKeyPairNotExists(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().KeyPairExists(g.ctx, g.machineConfig.Spec.SshKeyName).Return(false, nil)
	err := g.validator.validateSshKeyPair(g.ctx, g.machineConfig)
	g.Expect(err).To(MatchError(ContainSubstring("does not exist")))
}

func TestValidateSshKeyPairError(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().KeyPairExists(g.ctx, g.machineConfig.Spec.SshKeyName).Return(false, errors.New("error"))
	err := g.validator.validateSshKeyPair(g.ctx, g.machineConfig)
	g.Expect(err).NotTo(Succeed())
}

func TestValidateImageExistsOnDevice(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().ImageExists(g.ctx, g.machineConfig.Spec.AMIID).Return(true, nil).Times(2)
	err := g.validator.validateImageExistsOnDevice(g.ctx, g.machineConfig)
	g.Expect(err).To(Succeed())
}

func TestValidateImageExistsOnDeviceNotExists(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().ImageExists(g.ctx, g.machineConfig.Spec.AMIID).Return(false, nil)
	err := g.validator.validateImageExistsOnDevice(g.ctx, g.machineConfig)
	g.Expect(err).To(MatchError(ContainSubstring("does not exist")))
}

func TestValidateImageExistsOnDeviceError(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().ImageExists(g.ctx, g.machineConfig.Spec.AMIID).Return(false, errors.New("error"))
	err := g.validator.validateImageExistsOnDevice(g.ctx, g.machineConfig)
	g.Expect(err).NotTo(Succeed())
}
