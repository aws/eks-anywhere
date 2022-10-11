package nutanix

import (
	"fmt"
	"os"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

const (
	nutanixUsernameKey = "NUTANIX_USER"
	nutanixPasswordKey = "NUTANIX_PASSWORD"
	nutanixEndpointKey = "NUTANIX_ENDPOINT"
)

func setupEnvVars(datacenterConfig *anywherev1.NutanixDatacenterConfig) error {
	if err := os.Setenv(nutanixEndpointKey, datacenterConfig.Spec.Endpoint); err != nil {
		return fmt.Errorf("unable to set %s: %v", nutanixEndpointKey, err)
	}
	return nil
}

func getCredsFromEnv() basicAuthCreds {
	username := os.Getenv(nutanixUsernameKey)
	password := os.Getenv(nutanixPasswordKey)
	return basicAuthCreds{
		username: username,
		password: password,
	}
}
