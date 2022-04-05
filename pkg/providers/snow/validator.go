package snow

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/aws"
)

const (
	defaultAwsSshKeyNamePrefix = "eksa-default"
)

type Validator struct {
	awsClients aws.Clients
}

func NewValidator(aws aws.Clients) *Validator {
	return &Validator{
		awsClients: aws,
	}
}

func (v *Validator) validateSshKeyPair(ctx context.Context, m *v1alpha1.SnowMachineConfig) error {
	if m.Spec.SshKeyName == "" {
		return nil
	}

	for ip, client := range v.awsClients {
		keyExists, err := client.KeyPairExists(ctx, m.Spec.SshKeyName)
		if err != nil {
			return fmt.Errorf("describe key pair on snow device [%s]: %v", ip, err)
		}
		if !keyExists {
			return fmt.Errorf("aws key pair [%s] does not exist", m.Spec.SshKeyName)
		}
	}

	return nil
}

func (v *Validator) validateImageExistsOnDevice(ctx context.Context, m *v1alpha1.SnowMachineConfig) error {
	for ip, client := range v.awsClients {
		imageExists, err := client.ImageExists(ctx, m.Spec.AMIID)
		if err != nil {
			return fmt.Errorf("describe image on snow device [%s]: %v", ip, err)
		}
		if !imageExists {
			return fmt.Errorf("aws image [%s] does not exist", m.Spec.AMIID)
		}
	}

	return nil
}
