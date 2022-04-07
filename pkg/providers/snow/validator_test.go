package snow_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/providers/snow"
	"github.com/aws/eks-anywhere/pkg/providers/snow/mocks"
)

type configManagerTest struct {
	*WithT
	ctx           context.Context
	aws           *mocks.MockAwsClient
	validator     *snow.Validator
	defaulters    *snow.Defaulters
	machineConfig *v1alpha1.SnowMachineConfig
}

func newConfigManagerTest(t *testing.T) *configManagerTest {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mockaws := mocks.NewMockAwsClient(ctrl)
	awsClients := snow.AwsClientMap{
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
		validator:     snow.NewValidatorFromAwsClientMap(awsClients),
		defaulters:    snow.NewDefaultersFromAwsClientMap(awsClients, nil),
		machineConfig: m,
	}
}

func TestValidateEC2SshKeyNameExists(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().EC2KeyNameExists(g.ctx, g.machineConfig.Spec.SshKeyName).Return(true, nil).Times(2)
	err := g.validator.ValidateEC2SshKeyNameExists(g.ctx, g.machineConfig)
	g.Expect(err).To(Succeed())
}

func TestValidateEC2SshKeyNameExistsNotExists(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().EC2KeyNameExists(g.ctx, g.machineConfig.Spec.SshKeyName).Return(false, nil)
	err := g.validator.ValidateEC2SshKeyNameExists(g.ctx, g.machineConfig)
	g.Expect(err).To(MatchError(ContainSubstring("does not exist")))
}

func TestValidateEC2SshKeyNameExistsError(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().EC2KeyNameExists(g.ctx, g.machineConfig.Spec.SshKeyName).Return(false, errors.New("error"))
	err := g.validator.ValidateEC2SshKeyNameExists(g.ctx, g.machineConfig)
	g.Expect(err).NotTo(Succeed())
}

func TestValidateEC2ImageExistsOnDevice(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().EC2ImageExists(g.ctx, g.machineConfig.Spec.AMIID).Return(true, nil).Times(2)
	err := g.validator.ValidateEC2ImageExistsOnDevice(g.ctx, g.machineConfig)
	g.Expect(err).To(Succeed())
}

func TestValidateEC2ImageExistsOnDeviceNotExists(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().EC2ImageExists(g.ctx, g.machineConfig.Spec.AMIID).Return(false, nil)
	err := g.validator.ValidateEC2ImageExistsOnDevice(g.ctx, g.machineConfig)
	g.Expect(err).To(MatchError(ContainSubstring("does not exist")))
}

func TestValidateEC2ImageExistsOnDeviceError(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().EC2ImageExists(g.ctx, g.machineConfig.Spec.AMIID).Return(false, errors.New("error"))
	err := g.validator.ValidateEC2ImageExistsOnDevice(g.ctx, g.machineConfig)
	g.Expect(err).NotTo(Succeed())
}
