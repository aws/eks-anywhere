package cloudstack

import (
	"context"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type Defaulter struct {
	cmk ProviderCmkClient // Will be used to generate ZoneId's when using legacy CloudStackDatacenterConfig schema
}

func NewDefaulter(cmk ProviderCmkClient) *Defaulter {
	return &Defaulter{
		cmk: cmk,
	}
}

func (d *Defaulter) SetDefaultsForDatacenterConfig(ctx context.Context, datacenterConfig *anywherev1.CloudStackDatacenterConfig) error {
	datacenterConfig.SetDefaults()

	return nil
}
