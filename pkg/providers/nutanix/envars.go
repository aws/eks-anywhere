package nutanix

import (
	"fmt"
	"os"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func SetupEnvVars(datacenterConfig *anywherev1.NutanixDatacenterConfig) error {
	fmt.Printf("datacenterConfig: %#v\n", datacenterConfig)
	if err := os.Setenv(nutanixUsernameKey, datacenterConfig.Spec.NutanixUser); err != nil {
		return fmt.Errorf("unable to set %s: %v", nutanixUsernameKey, err)
	}

	if err := os.Setenv(nutanixPasswordKey, datacenterConfig.Spec.NutanixPassword); err != nil {
		return fmt.Errorf("unable to set %s: %v", nutanixPasswordKey, err)
	}

	if err := os.Setenv(nutanixEndpointKey, datacenterConfig.Spec.NutanixEndpoint); err != nil {
		return fmt.Errorf("unable to set %s: %v", nutanixEndpointKey, err)
	}

	return nil
}
