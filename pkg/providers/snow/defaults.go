package snow

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/aws"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/common"
)

type Defaulters struct {
	clientRegistry ClientRegistry
	writer         filewriter.FileWriter
	keyGenerator   SshKeyGenerator
	uuid           uuid.UUID
}

type SshKeyGenerator interface {
	GenerateSSHAuthKey(filewriter.FileWriter) (string, error)
}

type DefaultersOpt func(defaulters *Defaulters)

func NewDefaulters(clientRegistry ClientRegistry, writer filewriter.FileWriter, opts ...DefaultersOpt) *Defaulters {
	defaulters := &Defaulters{
		clientRegistry: clientRegistry,
		writer:         writer,
		keyGenerator:   common.SshAuthKeyGenerator{},
		uuid:           uuid.New(), // In the future if we need a cluster wide uuid that is shared, we should move this call to the dependency factory for reuse.
	}
	for _, opt := range opts {
		opt(defaulters)
	}
	return defaulters
}

func WithKeyGenerator(generator SshKeyGenerator) DefaultersOpt {
	return func(defaulters *Defaulters) {
		defaulters.keyGenerator = generator
	}
}

// WithUUID will set uuid generated outside of constructor.
func WithUUID(uuid uuid.UUID) DefaultersOpt {
	return func(defaulters *Defaulters) {
		defaulters.uuid = uuid
	}
}

// GenerateDefaultSSHKeys generates ssh key if it doesn't exist already.
func (d *Defaulters) GenerateDefaultSSHKeys(ctx context.Context, machineConfigs map[string]*v1alpha1.SnowMachineConfig, clusterName string) error {
	md := NewMachineConfigDefaulters(d)

	for _, m := range machineConfigs {
		if m.Spec.SshKeyName == "" {
			if md.keyGenerated {
				// Generate a random string so that other clusters with same name and devices won't conflict in case user doesn't have access to previous key
				m.Spec.SshKeyName = md.defaultSSHKeyName(clusterName)
			} else {
				if err := md.SetupDefaultSSHKey(ctx, m, clusterName); err != nil {
					return err
				}
			}
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

func (md *MachineConfigDefaulters) defaultKeyCount(ctx context.Context, clientMap AwsClientMap, m *v1alpha1.SnowMachineConfig, keyName string) (int, error) {
	var count int

	for _, ip := range m.Spec.Devices {
		client, ok := clientMap[ip]
		if !ok {
			return count, fmt.Errorf("credentials not found for device [%s]", ip)
		}

		keyExists, err := client.EC2KeyNameExists(ctx, keyName)
		if err != nil {
			return count, fmt.Errorf("describing key pair on snow device [deviceIP=%s]: %v", ip, err)
		}
		if keyExists {
			count += 1
		}
	}

	return count, nil
}

// SetupDefaultSSHKey checks the existing keys and sets up a default key.
func (md *MachineConfigDefaulters) SetupDefaultSSHKey(ctx context.Context, m *v1alpha1.SnowMachineConfig, clusterName string) error {
	defaultSSHKeyName := md.defaultSSHKeyName(clusterName)

	clientMap, err := md.defaulters.clientRegistry.Get(ctx)
	if err != nil {
		return err
	}
	keyCount, err := md.defaultKeyCount(ctx, clientMap, m, defaultSSHKeyName)
	if err != nil {
		return err
	}
	if keyCount > 0 && keyCount < len(m.Spec.Devices) {
		return fmt.Errorf("default key [keyName=%s] only exists on some of the devices. Use 'aws ec2 import-key-pair' to import this key to all the devices", defaultSSHKeyName)
	}
	if keyCount == len(m.Spec.Devices) {
		md.keyGenerated = true
		m.Spec.SshKeyName = defaultSSHKeyName
		return nil
	}

	logger.V(1).Info("SnowMachineConfig SshKey is empty. Creating default key pair", "default key name", defaultSSHKeyName)
	key, err := md.defaulters.keyGenerator.GenerateSSHAuthKey(md.defaulters.writer)
	if err != nil {
		return err
	}

	for _, ip := range m.Spec.Devices {
		client, ok := clientMap[ip]
		if !ok {
			return fmt.Errorf("credentials not found for device [%s]", ip)
		}

		err := client.EC2ImportKeyPair(ctx, defaultSSHKeyName, []byte(key))
		if err != nil {
			return fmt.Errorf("importing key pair on snow device [deviceIP=%s]: %v", ip, err)
		}
	}

	md.keyGenerated = true
	m.Spec.SshKeyName = defaultSSHKeyName
	return nil
}

func (md *MachineConfigDefaulters) defaultSSHKeyName(clusterName string) string {
	return fmt.Sprintf("%s-%s-%s", defaultAwsSshKeyName, clusterName, md.defaulters.uuid.String())
}

func SetupEksaCredentialsSecret(c *cluster.Config) error {
	creds, err := aws.EncodeFileFromEnv(eksaSnowCredentialsFileKey)
	if err != nil {
		return fmt.Errorf("setting up snow credentials: %v", err)
	}

	certs, err := aws.EncodeFileFromEnv(eksaSnowCABundlesFileKey)
	if err != nil {
		return fmt.Errorf("setting up snow certificates: %v", err)
	}

	c.SnowCredentialsSecret = EksaCredentialsSecret(c.SnowDatacenter, []byte(creds), []byte(certs))

	return nil
}
