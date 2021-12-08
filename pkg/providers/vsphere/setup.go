package vsphere

import (
	"context"
	"fmt"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func (v *vsphereProvider) setup(ctx context.Context, datacenterConfig *anywherev1.VSphereDatacenterConfig) error {
	if err := setupEnvVars(datacenterConfig); err != nil {
		return err
	}

	if err := v.setupGovc(ctx, datacenterConfig); err != nil {
		return err
	}

	return nil
}

func (v *vsphereProvider) setupGovc(ctx context.Context, datacenterConfig *anywherev1.VSphereDatacenterConfig) error {
	if datacenterConfig.Spec.Thumbprint == "" {
		return nil
	}

	if err := v.providerGovcClient.ConfigureCertThumbprint(ctx, datacenterConfig.Spec.Server, datacenterConfig.Spec.Thumbprint); err != nil {
		return fmt.Errorf("failed setting up govc: %v", err)
	}

	return nil
}
