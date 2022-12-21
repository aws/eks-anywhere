package snow

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

const (
	defaultAwsSshKeyName       = "eksa-default"
	snowballMinSoftwareVersion = "102"
)

type AwsClientValidator struct {
	clientRegistry ClientRegistry
}

func NewValidator(clientRegistry ClientRegistry) *AwsClientValidator {
	return &AwsClientValidator{
		clientRegistry: clientRegistry,
	}
}

func (v *AwsClientValidator) ValidateEC2SshKeyNameExists(ctx context.Context, m *v1alpha1.SnowMachineConfig) error {
	if m.Spec.SshKeyName == "" {
		return nil
	}

	clientMap, err := v.clientRegistry.Get(ctx)
	if err != nil {
		return err
	}

	for _, ip := range m.Spec.Devices {
		client, ok := clientMap[ip]
		if !ok {
			return fmt.Errorf("credentials not found for device [%s]", ip)
		}

		keyExists, err := client.EC2KeyNameExists(ctx, m.Spec.SshKeyName)
		if err != nil {
			return fmt.Errorf("describing key pair on snow device [%s]: %v", ip, err)
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

	clientMap, err := v.clientRegistry.Get(ctx)
	if err != nil {
		return err
	}

	for _, ip := range m.Spec.Devices {
		client, ok := clientMap[ip]
		if !ok {
			return fmt.Errorf("credentials not found for device [%s]", ip)
		}

		imageExists, err := client.EC2ImageExists(ctx, m.Spec.AMIID)
		if err != nil {
			return fmt.Errorf("describing image on snow device [%s]: %v", ip, err)
		}
		if !imageExists {
			return fmt.Errorf("aws image [%s] does not exist", m.Spec.AMIID)
		}
	}

	return nil
}

func (v *AwsClientValidator) ValidateDeviceIsUnlocked(ctx context.Context, m *v1alpha1.SnowMachineConfig) error {
	clientMap, err := v.clientRegistry.Get(ctx)
	if err != nil {
		return err
	}

	for _, ip := range m.Spec.Devices {
		client, ok := clientMap[ip]
		if !ok {
			return fmt.Errorf("credentials not found for device [%s]", ip)
		}

		deviceUnlocked, err := client.IsSnowballDeviceUnlocked(ctx)
		if err != nil {
			return fmt.Errorf("checking unlock status for device [%s]: %v", ip, err)
		}
		if !deviceUnlocked {
			return fmt.Errorf("device [%s] is not unlocked. Please unlock the device before you proceed", ip)
		}
	}

	return nil
}

func (v *AwsClientValidator) ValidateDeviceSoftware(ctx context.Context, m *v1alpha1.SnowMachineConfig) error {
	clientMap, err := v.clientRegistry.Get(ctx)
	if err != nil {
		return err
	}

	for _, ip := range m.Spec.Devices {
		client, ok := clientMap[ip]
		if !ok {
			return fmt.Errorf("credentials not found for device [%s]", ip)
		}

		version, err := client.SnowballDeviceSoftwareVersion(ctx)
		if err != nil {
			return fmt.Errorf("checking software version for device [%s]: %v", ip, err)
		}
		if version < snowballMinSoftwareVersion {
			return fmt.Errorf("the software version installed [%s] on device [%s] is below the minimum supported version [%s]", version, ip, snowballMinSoftwareVersion)
		}
	}

	return nil
}
