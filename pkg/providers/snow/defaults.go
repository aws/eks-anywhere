package snow

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/aws"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/common"
)

type Defaulters struct {
	awsClientMap AwsClientMap
	writer       filewriter.FileWriter
	keyGenerator SshKeyGenerator
}

type SshKeyGenerator interface {
	GenerateSSHAuthKey(filewriter.FileWriter) (string, error)
}

func NewDefaulters(aws aws.Clients, writer filewriter.FileWriter) *Defaulters {
	return &Defaulters{
		awsClientMap: NewAwsClientMap(aws),
		writer:       writer,
		keyGenerator: common.SshAuthKeyGenerator{},
	}
}

func NewDefaultersFromAwsClientMap(awsClientMap AwsClientMap, writer filewriter.FileWriter, keyGenerator SshKeyGenerator) *Defaulters {
	return &Defaulters{
		awsClientMap: awsClientMap,
		writer:       writer,
		keyGenerator: keyGenerator,
	}
}

func (d *Defaulters) GenerateDefaultSshKeys(ctx context.Context, machineConfigs map[string]*v1alpha1.SnowMachineConfig) error {
	md := NewMachineConfigDefaulters(d)

	for _, m := range machineConfigs {
		if err := md.SetupDefaultSshKey(ctx, m); err != nil {
			return err
		}
	}

	return nil
}

type MachineConfigDefaulters struct {
	keyGenerated bool
	defaulters   *Defaulters
}

func NewMachineConfigDefaulters(d *Defaulters) *MachineConfigDefaulters {
	return &MachineConfigDefaulters{
		defaulters: d,
	}
}

func (md *MachineConfigDefaulters) SetupDefaultSshKey(ctx context.Context, m *v1alpha1.SnowMachineConfig) error {
	if m.Spec.SshKeyName != "" {
		return nil
	}

	if md.keyGenerated {
		m.Spec.SshKeyName = defaultAwsSshKeyName
		return nil
	}

	logger.V(1).Info("SnowMachineConfig SshKey is empty. Creating default key pair", "default key name", defaultAwsSshKeyName)

	key, err := md.defaulters.keyGenerator.GenerateSSHAuthKey(md.defaulters.writer)
	if err != nil {
		return err
	}

	for ip, client := range md.defaulters.awsClientMap {
		err := client.EC2ImportKeyPair(ctx, defaultAwsSshKeyName, []byte(key))
		if err != nil {
			return fmt.Errorf("importing key pair on snow device [deviceIP=%s]: %v", ip, err)
		}
	}

	md.keyGenerated = true

	m.Spec.SshKeyName = defaultAwsSshKeyName

	return nil
}
