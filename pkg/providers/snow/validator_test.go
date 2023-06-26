package snow_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers/snow"
	"github.com/aws/eks-anywhere/pkg/providers/snow/mocks"
)

type configManagerTest struct {
	*WithT
	ctx                     context.Context
	aws                     *mocks.MockAwsClient
	keyGenerator            *mocks.MockSshKeyGenerator
	writer                  filewriter.FileWriter
	validator               *snow.Validator
	defaulters              *snow.Defaulters
	machineConfigDefaulters *snow.MachineConfigDefaulters
	machineConfig           *v1alpha1.SnowMachineConfig
	uuid                    uuid.UUID
	clusterName             string
	defaultKeyName          string
}

func newConfigManagerTest(t *testing.T) *configManagerTest {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mockaws := mocks.NewMockAwsClient(ctrl)
	mockKeyGenerator := mocks.NewMockSshKeyGenerator(ctrl)
	awsClients := snow.AwsClientMap{
		"device-1": mockaws,
		"device-2": mockaws,
		"device-3": mockaws,
	}
	mockClientRegistry := mocks.NewMockClientRegistry(ctrl)
	mockClientRegistry.EXPECT().Get(ctx).Return(awsClients, nil).AnyTimes()
	m := &v1alpha1.SnowMachineConfig{
		ObjectMeta: v1.ObjectMeta{
			Name: "cp-machine",
		},
		Spec: v1alpha1.SnowMachineConfigSpec{
			AMIID:      "ami-1",
			SshKeyName: "default",
			Devices: []string{
				"device-1",
				"device-2",
			},
		},
	}
	_, writer := test.NewWriter(t)
	validators := snow.NewValidator(mockClientRegistry)
	uuid := uuid.New()
	defaulters := snow.NewDefaulters(mockClientRegistry, writer, snow.WithKeyGenerator(mockKeyGenerator), snow.WithUUID(uuid))
	return &configManagerTest{
		WithT:                   NewWithT(t),
		ctx:                     ctx,
		aws:                     mockaws,
		keyGenerator:            mockKeyGenerator,
		writer:                  writer,
		validator:               validators,
		defaulters:              defaulters,
		machineConfigDefaulters: snow.NewMachineConfigDefaulters(defaulters),
		machineConfig:           m,
		uuid:                    uuid,
		clusterName:             "test-snow",
		defaultKeyName:          fmt.Sprintf("eksa-default-test-snow-%s", uuid),
	}
}

func newConfigManagerTestClientMapError(t *testing.T) *configManagerTest {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mockKeyGenerator := mocks.NewMockSshKeyGenerator(ctrl)
	mockClientRegistry := mocks.NewMockClientRegistry(ctrl)
	mockClientRegistry.EXPECT().Get(ctx).Return(nil, errors.New("test error"))
	m := &v1alpha1.SnowMachineConfig{
		ObjectMeta: v1.ObjectMeta{
			Name: "cp-machine",
		},
		Spec: v1alpha1.SnowMachineConfigSpec{
			AMIID:      "ami-1",
			SshKeyName: "default",
			Devices: []string{
				"device-1",
				"device-2",
			},
		},
	}
	_, writer := test.NewWriter(t)
	validators := snow.NewValidator(mockClientRegistry)
	defaulters := snow.NewDefaulters(mockClientRegistry, writer, snow.WithKeyGenerator(mockKeyGenerator))
	return &configManagerTest{
		WithT:                   NewWithT(t),
		ctx:                     ctx,
		keyGenerator:            mockKeyGenerator,
		writer:                  writer,
		validator:               validators,
		defaulters:              defaulters,
		machineConfigDefaulters: snow.NewMachineConfigDefaulters(defaulters),
		machineConfig:           m,
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

func TestValidateEC2SshKeyNameEmpty(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.SshKeyName = ""
	err := g.validator.ValidateEC2SshKeyNameExists(g.ctx, g.machineConfig)
	g.Expect(err).To(Succeed())
}

func TestValidateEC2SshKeyNameClientMapError(t *testing.T) {
	g := newConfigManagerTestClientMapError(t)
	err := g.validator.ValidateEC2SshKeyNameExists(g.ctx, g.machineConfig)
	g.Expect(err).NotTo(Succeed())
}

func TestValidateEC2SshKeyNameDeviceNotFoundInClientMapError(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.Devices = []string{"device-not-exist"}
	err := g.validator.ValidateEC2SshKeyNameExists(g.ctx, g.machineConfig)
	g.Expect(err).To(MatchError(ContainSubstring("credentials not found for device")))
}

func TestValidateEC2ImageExistsOnDevice(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().EC2ImageExists(g.ctx, g.machineConfig.Spec.AMIID).Return(true, nil).Times(2)
	err := g.validator.ValidateEC2ImageExistsOnDevice(g.ctx, g.machineConfig)
	g.Expect(err).To(Succeed())
}

func TestValidateEC2ImageExistsOnDeviceAmiIDEmpty(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.AMIID = ""
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

func TestValidateEC2ImageExistsOnDeviceClientMapError(t *testing.T) {
	g := newConfigManagerTestClientMapError(t)
	err := g.validator.ValidateEC2ImageExistsOnDevice(g.ctx, g.machineConfig)
	g.Expect(err).NotTo(Succeed())
}

func TestValidateEC2ImageExistsOnDeviceNotFoundInClientMapError(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.Devices = []string{"device-not-exist"}
	err := g.validator.ValidateEC2ImageExistsOnDevice(g.ctx, g.machineConfig)
	g.Expect(err).To(MatchError(ContainSubstring("credentials not found for device")))
}

func TestValidateDeviceIsUnlocked(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().IsSnowballDeviceUnlocked(g.ctx).Return(true, nil).Times(2)
	err := g.validator.ValidateDeviceIsUnlocked(g.ctx, g.machineConfig)
	g.Expect(err).To(Succeed())
}

func TestValidateDeviceIsUnlockedLocked(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().IsSnowballDeviceUnlocked(g.ctx).Return(false, nil)
	err := g.validator.ValidateDeviceIsUnlocked(g.ctx, g.machineConfig)
	g.Expect(err).To(MatchError(ContainSubstring("Please unlock the device before you proceed")))
}

func TestValidateDeviceIsUnlockedError(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().IsSnowballDeviceUnlocked(g.ctx).Return(false, errors.New("error"))
	err := g.validator.ValidateDeviceIsUnlocked(g.ctx, g.machineConfig)
	g.Expect(err).NotTo(Succeed())
}

func TestValidateDeviceIsUnlockedClientMapError(t *testing.T) {
	g := newConfigManagerTestClientMapError(t)
	err := g.validator.ValidateDeviceIsUnlocked(g.ctx, g.machineConfig)
	g.Expect(err).NotTo(Succeed())
}

func TestValidateDeviceIsUnlockedNotFoundInClientMapError(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.Devices = []string{"device-not-exist"}
	err := g.validator.ValidateDeviceIsUnlocked(g.ctx, g.machineConfig)
	g.Expect(err).To(MatchError(ContainSubstring("credentials not found for device")))
}

func TestValidateDeviceSoftware(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().SnowballDeviceSoftwareVersion(g.ctx).Return("1012", nil).Times(2)
	err := g.validator.ValidateDeviceSoftware(g.ctx, g.machineConfig)
	g.Expect(err).To(Succeed())
}

func TestValidateDeviceSoftwareVersionTooLow(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().SnowballDeviceSoftwareVersion(g.ctx).Return("101", nil)
	err := g.validator.ValidateDeviceSoftware(g.ctx, g.machineConfig)
	g.Expect(err).To(MatchError(ContainSubstring("below the minimum supported version")))
}

func TestValidateDeviceSoftwareVersionError(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().SnowballDeviceSoftwareVersion(g.ctx).Return("", errors.New("error"))
	err := g.validator.ValidateDeviceSoftware(g.ctx, g.machineConfig)
	g.Expect(err).NotTo(Succeed())
}

func TestValidateDeviceSoftwareClientMapError(t *testing.T) {
	g := newConfigManagerTestClientMapError(t)
	err := g.validator.ValidateDeviceSoftware(g.ctx, g.machineConfig)
	g.Expect(err).NotTo(Succeed())
}

func TestValidateDeviceSoftwareNotFoundInClientMapError(t *testing.T) {
	g := newConfigManagerTest(t)
	g.machineConfig.Spec.Devices = []string{"device-not-exist"}
	err := g.validator.ValidateDeviceSoftware(g.ctx, g.machineConfig)
	g.Expect(err).To(MatchError(ContainSubstring("credentials not found for device")))
}

func TestValidateDeviceSoftwareConvertToIntegerError(t *testing.T) {
	g := newConfigManagerTest(t)
	g.aws.EXPECT().SnowballDeviceSoftwareVersion(g.ctx).Return("version", nil)
	err := g.validator.ValidateDeviceSoftware(g.ctx, g.machineConfig)
	g.Expect(err).To(MatchError(ContainSubstring("invalid syntax")))
}
