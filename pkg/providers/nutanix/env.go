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
	if nutanixUsername, ok := os.LookupEnv(constants.EksaNutanixUsernameKey); ok && len(nutanixUsername) > 0 {
		if err := os.Setenv(constants.NutanixUsernameKey, nutanixUsername); err != nil {
			return fmt.Errorf("unable to set %s: %v", constants.EksaNutanixUsernameKey, err)
		}
	} else {
		return fmt.Errorf("%s is not set or is empty", constants.EksaNutanixUsernameKey)
	}

	if nutanixPassword, ok := os.LookupEnv(constants.EksaNutanixPasswordKey); ok && len(nutanixPassword) > 0 {
		if err := os.Setenv(constants.NutanixPasswordKey, nutanixPassword); err != nil {
			return fmt.Errorf("unable to set %s: %v", constants.EksaNutanixPasswordKey, err)
		}
	} else {
		return fmt.Errorf("%s is not set or is empty", constants.EksaNutanixPasswordKey)
	}

	if err := os.Setenv(nutanixEndpointKey, datacenterConfig.Spec.Endpoint); err != nil {
		return fmt.Errorf("unable to set %s: %v", nutanixEndpointKey, err)
	}
	return nil
}

// GetCredsFromEnv returns nutanix credentials based on the environment.
func GetCredsFromEnv() credentials.BasicAuthCredential {
	username := os.Getenv(constants.EksaNutanixUsernameKey)
	password := os.Getenv(constants.EksaNutanixPasswordKey)
	return credentials.BasicAuthCredential{
		PrismCentral: credentials.PrismCentralBasicAuth{
			BasicAuth: credentials.BasicAuth{
				Username: username,
				Password: password,
			},
		},
	}
}
