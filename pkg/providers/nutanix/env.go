package nutanix

import (
	"fmt"
	"os"

	"github.com/nutanix-cloud-native/prism-go-client/environment/credentials"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

const (
	nutanixEndpointKey = "NUTANIX_ENDPOINT"
)

func setupEnvVars(datacenterConfig *anywherev1.NutanixDatacenterConfig) error {
	if err := os.Setenv(nutanixEndpointKey, datacenterConfig.Spec.Endpoint); err != nil {
		return fmt.Errorf("unable to set %s: %v", nutanixEndpointKey, err)
	}
	return nil
}

// GetCredsFromEnv returns nutanix credentials based on the environment.
func GetCredsFromEnv() credentials.BasicAuthCredential {
	username := os.Getenv(constants.NutanixUsernameKey)
	password := os.Getenv(constants.NutanixPasswordKey)
	return credentials.BasicAuthCredential{
		PrismCentral: credentials.PrismCentralBasicAuth{
			BasicAuth: credentials.BasicAuth{
				Username: username,
				Password: password,
			},
		},
	}
}
