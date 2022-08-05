package snow

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/aws"
)

const (
	defaultAwsSshKeyName = "eksa-default"
)

type AwsClientValidator struct {
	awsClientMap AwsClientMap
}

type Validator interface {
	ValidateEC2SshKeyNameExists(ctx context.Context, m *v1alpha1.SnowMachineConfig) error
	ValidateEC2ImageExistsOnDevice(ctx context.Context, m *v1alpha1.SnowMachineConfig) error
	ValidateMachineDeviceIPs(ctx context.Context, m *v1alpha1.SnowMachineConfig) error
}

type validatorBuilder struct{}

func NewValidatorBuilder() *validatorBuilder {
	return &validatorBuilder{}
}

func (b *validatorBuilder) Build(aws aws.Clients) Validator {
	return NewValidator(aws)
}

func NewValidator(aws aws.Clients) *AwsClientValidator {
	return &AwsClientValidator{
		awsClientMap: NewAwsClientMap(aws),
	}
}

func NewValidatorFromAwsClientMap(awsClientMap AwsClientMap) *AwsClientValidator {
	return &AwsClientValidator{
		awsClientMap: awsClientMap,
	}
}

func (v *AwsClientValidator) ValidateEC2SshKeyNameExists(ctx context.Context, m *v1alpha1.SnowMachineConfig) error {
	if m.Spec.SshKeyName == "" {
		return nil
	}

	for ip, client := range v.awsClientMap {
		keyExists, err := client.EC2KeyNameExists(ctx, m.Spec.SshKeyName)
		if err != nil {
			return fmt.Errorf("describe key pair on snow device [%s]: %v", ip, err)
		}
		if !keyExists {
			return fmt.Errorf("aws key pair [%s] does not exist", m.Spec.SshKeyName)
		}
	}

	return nil
}

func (v *AwsClientValidator) ValidateEC2ImageExistsOnDevice(ctx context.Context, m *v1alpha1.SnowMachineConfig) error {
	if m.Spec.AMIID == "" {
		return nil
	}

	for ip, client := range v.awsClientMap {
		imageExists, err := client.EC2ImageExists(ctx, m.Spec.AMIID)
		if err != nil {
			return fmt.Errorf("describe image on snow device [%s]: %v", ip, err)
		}
		if !imageExists {
			return fmt.Errorf("aws image [%s] does not exist", m.Spec.AMIID)
		}
	}

	return nil
}

func (v *AwsClientValidator) ValidateMachineDeviceIPs(ctx context.Context, m *v1alpha1.SnowMachineConfig) error {
	for _, ip := range m.Spec.Devices {
		if _, ok := v.awsClientMap[ip]; !ok {
			return fmt.Errorf("credentials not found for device [%s]", ip)
		}
	}

	return nil
}
