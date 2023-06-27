package snow

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

const (
	defaultAwsSshKeyName       = "eksa-default"
	snowballMinSoftwareVersion = 102
	minimumVCPU                = 2
)

// Validator includes a client registry that maintains a snow device aws client map,
// and a local imds service that is used to fetch metadata of the host instance.
type Validator struct {
	// clientRegistry maintains a device aws client mapping.
	clientRegistry ClientRegistry

	// imds is a local imds client built with the default aws config. This imds client can only
	// interact with the local instance metata service.
	imds LocalIMDSClient
}

// ValidatorOpt updates an Validator.
type ValidatorOpt func(*Validator)

// WithIMDS returns a ValidatorOpt that sets the imds client.
func WithIMDS(imds LocalIMDSClient) ValidatorOpt {
	return func(c *Validator) {
		c.imds = imds
	}
}

// NewValidator creates a snow validator.
func NewValidator(clientRegistry ClientRegistry, opts ...ValidatorOpt) *Validator {
	v := &Validator{
		clientRegistry: clientRegistry,
	}

	for _, o := range opts {
		o(v)
	}

	return v
}

// ValidateEC2SshKeyNameExists validates the ssh key existence in each device in the device list.
func (v *Validator) ValidateEC2SshKeyNameExists(ctx context.Context, m *v1alpha1.SnowMachineConfig) error {
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
			return fmt.Errorf("aws key pair [%s] does not exist on snow device [deviceIP=%s]", m.Spec.SshKeyName, ip)
		}
	}

	return nil
}

// ValidateEC2ImageExistsOnDevice validates the ami id (if specified) existence in each device in the device list.
func (v *Validator) ValidateEC2ImageExistsOnDevice(ctx context.Context, m *v1alpha1.SnowMachineConfig) error {
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

// ValidateDeviceIsUnlocked verifies if all snow devices in the device list are unlocked.
func (v *Validator) ValidateDeviceIsUnlocked(ctx context.Context, m *v1alpha1.SnowMachineConfig) error {
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

func validateInstanceTypeInDevice(ctx context.Context, client AwsClient, instanceType, deviceIP string) error {
	instanceTypes, err := client.EC2InstanceTypes(ctx)
	if err != nil {
		return fmt.Errorf("fetching supported instance types for device [%s]: %v", deviceIP, err)
	}

	for _, it := range instanceTypes {
		if instanceType != it.Name {
			continue
		}

		if it.DefaultVCPU != nil && *it.DefaultVCPU < minimumVCPU {
			return fmt.Errorf("the instance type [%s] has %d vCPU. Please choose an instance type with at least %d default vCPU", instanceType, *it.DefaultVCPU, minimumVCPU)
		}

		return nil
	}

	return fmt.Errorf("the instance type [%s] is not supported in device [%s]", instanceType, deviceIP)
}

// ValidateInstanceType validates whether the instance type is compatible to run in each device.
func (v *Validator) ValidateInstanceType(ctx context.Context, m *v1alpha1.SnowMachineConfig) error {
	clientMap, err := v.clientRegistry.Get(ctx)
	if err != nil {
		return err
	}

	for _, ip := range m.Spec.Devices {
		client, ok := clientMap[ip]
		if !ok {
			return fmt.Errorf("credentials not found for device [%s]", ip)
		}

		if err := validateInstanceTypeInDevice(ctx, client, m.Spec.InstanceType, ip); err != nil {
			return err
		}
	}

	return nil
}

// ValidateDeviceSoftware validates whether the snow software is compatible to run eks-a in each device.
func (v *Validator) ValidateDeviceSoftware(ctx context.Context, m *v1alpha1.SnowMachineConfig) error {
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

		versionInt, err := strconv.Atoi(version)
		if err != nil {
			return fmt.Errorf("checking software version for device [%s]: %v", ip, err)
		}

		if versionInt < snowballMinSoftwareVersion {
			return fmt.Errorf("the software version installed [%s] on device [%s] is below the minimum supported version [%d]", version, ip, snowballMinSoftwareVersion)
		}
	}

	return nil
}

// ValidateControlPlaneIP checks whether the control plane ip is valid for creating a snow cluster.
func (v *Validator) ValidateControlPlaneIP(ctx context.Context, controlPlaneIP string) error {
	if v.imds == nil || reflect.ValueOf(v.imds).IsNil() {
		return errors.New("imds client is not initialized")
	}

	instanceIP, err := v.imds.EC2InstanceIP(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			// the admin instance is not running inside snow devices or doesn't have a public IP
			return nil
		}
		return fmt.Errorf("fetching host instance ip: %v", err)
	}

	if controlPlaneIP == instanceIP {
		return fmt.Errorf("control plane host ip cannot be same as the admin instance ip")
	}

	return nil
}
