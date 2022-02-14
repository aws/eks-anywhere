package templates

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type Factory struct {
	client  CloudMonkeyClient
	network v1alpha1.CloudStackResourceRef
	domain  string
	zone    v1alpha1.CloudStackResourceRef
	account string
}

type CloudMonkeyClient interface{}

func NewFactory(client CloudMonkeyClient, network v1alpha1.CloudStackResourceRef, domain string, zone v1alpha1.CloudStackResourceRef, account string) *Factory {
	return &Factory{
		client:  client,
		network: network,
		domain:  domain,
		zone:    zone,
		account: account,
	}
}

func (f *Factory) ValidateMachineResources(ctx context.Context, machineConfig *v1alpha1.CloudStackMachineConfig) error {
	return nil
}
