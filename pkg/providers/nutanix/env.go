package nutanix

import (
	"fmt"
	"os"

	"github.com/nutanix-cloud-native/prism-go-client/environment/credentials"

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

func getCredsFromEnv() credentials.BasicAuthCredential {
	username := os.Getenv(nutanixUsernameKey)
	password := os.Getenv(nutanixPasswordKey)
	return credentials.BasicAuthCredential{
		PrismCentral: credentials.PrismCentralBasicAuth{
			BasicAuth: credentials.BasicAuth{
				Username: username,
				Password: password,
			},
		},
	}
}
