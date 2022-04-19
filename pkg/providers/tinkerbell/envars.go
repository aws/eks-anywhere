package tinkerbell

import (
	"fmt"
	"os"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func setupEnvVars(datacenterConfig *anywherev1.TinkerbellDatacenterConfig) error {
	if err := os.Setenv(tinkerbellIPKey, datacenterConfig.Spec.TinkerbellIP); err != nil {
		return fmt.Errorf("unable to set %s: %v", tinkerbellIPKey, err)
	}

	if err := os.Setenv(tinkerbellCertURLKey, datacenterConfig.Spec.TinkerbellCertURL); err != nil {
		return fmt.Errorf("unable to set %s: %v", tinkerbellCertURLKey, err)
	}

	if err := os.Setenv(tinkerbellGRPCAuthKey, datacenterConfig.Spec.TinkerbellGRPCAuth); err != nil {
		return fmt.Errorf("unable to set %s: %v", tinkerbellGRPCAuthKey, err)
	}

	if err := os.Setenv(tinkerbellPBnJGRPCAuthorityKey, datacenterConfig.Spec.TinkerbellPBnJGRPCAuth); err != nil {
		return fmt.Errorf("unable to set %s: %v", tinkerbellPBnJGRPCAuthorityKey, err)
	}

	if err := os.Setenv(tinkerbellHegelURLKey, datacenterConfig.Spec.TinkerbellHegelURL); err != nil {
		return fmt.Errorf("unable to set %s: %v", tinkerbellHegelURLKey, err)
	}
	return nil
}
