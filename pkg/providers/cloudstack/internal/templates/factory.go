package templates

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type Factory struct {
	client  CloudMonkeyClient
	network string
	domain  string
	zone    string
	account string
}

type CloudMonkeyClient interface {
	SearchTemplate(ctx context.Context, domain string, zone string, account string, template string) (string, error)
	SearchComputeOffering(ctx context.Context, domain string, zone string, account string, computeOffering string) (string, error)
	SearchDiskOffering(ctx context.Context, domain string, zone string, account string, diskOffering string) (string, error)
	ValidateCloudStackSetup(ctx context.Context, deploymentConfig *v1alpha1.CloudStackDeploymentConfig, selfSigned *bool) error
	ValidateCloudStackSetupMachineConfig(ctx context.Context, deploymentConfig *v1alpha1.CloudStackDeploymentConfig, machineConfig *v1alpha1.CloudStackMachineConfig, selfSigned *bool) error
}

func NewFactory(client CloudMonkeyClient, network, domain, zone, account string) *Factory {
	return &Factory{
		client:  client,
		network: network,
		domain:  domain,
		zone:    zone,
		account: account,
	}
}

func (f *Factory) ValidateMachineResources(ctx context.Context, machineConfig *v1alpha1.CloudStackMachineConfig) error {
	_, err := f.client.SearchTemplate(ctx, f.domain, f.zone, f.account, machineConfig.Spec.Template)
	if err != nil {
		return fmt.Errorf("error checking for machine config template: #{err}")
	}
	_, err = f.client.SearchDiskOffering(ctx, f.domain, f.zone, f.account, machineConfig.Spec.DiskOffering)
	if err != nil {
		return fmt.Errorf("error checking for machine config diskOffering: #{err}")
	}
	_, err = f.client.SearchComputeOffering(ctx, f.domain, f.zone, f.account, machineConfig.Spec.ComputeOffering)
	if err != nil {
		return fmt.Errorf("error checking for machine config computeOffering: #{err}")
	}
	return nil
}
