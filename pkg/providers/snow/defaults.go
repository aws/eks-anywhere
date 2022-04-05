package snow

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/aws"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
)

type Defaulters struct {
	awsClients aws.Clients
	writer     filewriter.FileWriter
}

func NewDefaulters(aws aws.Clients, writer filewriter.FileWriter) *Defaulters {
	return &Defaulters{
		awsClients: aws,
		writer:     writer,
	}
}

func (d *Defaulters) setDefaultSshKey(ctx context.Context, m *v1alpha1.SnowMachineConfig) error {
	if m.Spec.SshKeyName != "" {
		return nil
	}

	keyName := fmt.Sprintf("%s-%s.pem", defaultAwsSshKeyNamePrefix, m.GetName())

	logger.V(1).Info("SnowMachineConfig SshKey is empty. Creating default key pair", "default key name", keyName)

	// Only need to fetch IP from single device to create ssh key pair. The created key will then be validated in validateSshKeyPair method
	var anyDeviceIP string
	for ip := range d.awsClients {
		anyDeviceIP = ip
		break
	}
	keyVal, err := d.awsClients[anyDeviceIP].CreateEC2KeyPairs(ctx, keyName)
	if err != nil {
		return fmt.Errorf("creating ssh key pair in snow: %v", err)
	}

	path, err := d.writer.Write(keyName, []byte(keyVal), filewriter.PersistentFile, filewriter.Permission0600)
	if err != nil {
		return fmt.Errorf("writing private key: %v", err)
	}

	logger.Info("AWS key pem saved", "sshCommand", fmt.Sprintf("ssh -i %s <username>@<Node-IP-Address>", path))

	m.Spec.SshKeyName = keyName
	return nil
}
