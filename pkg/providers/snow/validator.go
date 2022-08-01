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

type Validator struct {
	awsClientMap AwsClientMap
}

func NewValidator(aws aws.Clients) *Validator {
	return &Validator{
		awsClientMap: NewAwsClientMap(aws),
	}
}

func NewValidatorFromAwsClientMap(awsClientMap AwsClientMap) *Validator {
	return &Validator{
		awsClientMap: awsClientMap,
	}
}

func (v *Validator) ValidateEC2SshKeyNameExists(ctx context.Context, m *v1alpha1.SnowMachineConfig) error {
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

func (v *Validator) ValidateEC2ImageExistsOnDevice(ctx context.Context, m *v1alpha1.SnowMachineConfig) error {
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

func (v *Validator) ValidateMachineDeviceIPs(ctx context.Context, m *v1alpha1.SnowMachineConfig) error {
	for _, ip := range m.Spec.Devices {
		if _, ok := v.awsClientMap[ip]; !ok {
			return fmt.Errorf("credentials not found for device [%s]", ip)
		}
	}

	return nil
}
