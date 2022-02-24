package decoder

import (
	b64 "encoding/base64"
	"fmt"
	"os"

	"gopkg.in/ini.v1"
)

const eksacloudStackCloudConfigB64SecretKey = "EKSA_CLOUDSTACK_B64ENCODED_SECRET"

// ParseCloudStackSecret parses the input b64 string into the ini object to extract out the api key, secret key, and url
func ParseCloudStackSecret() (*CloudStackExecConfig, error) {
	cloudStackB64EncodedSecret, ok := os.LookupEnv(eksacloudStackCloudConfigB64SecretKey)
	if !ok {
		return nil, fmt.Errorf("%s is not set or is empty", eksacloudStackCloudConfigB64SecretKey)
	}
	decodedString, err := b64.StdEncoding.DecodeString(cloudStackB64EncodedSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode value for %s with base64: %v", eksacloudStackCloudConfigB64SecretKey, err)
	}
	cfg, err := ini.Load(decodedString)
	if err != nil {
		return nil, fmt.Errorf("failed to extract values from %s with ini: %v", eksacloudStackCloudConfigB64SecretKey, err)
	}
	section, err := cfg.GetSection("Global")
	if err != nil {
		return nil, fmt.Errorf("failed to extract section 'Global' from %s: %v", eksacloudStackCloudConfigB64SecretKey, err)
	}
	apiKey, err := section.GetKey("api-key")
	if err != nil {
		return nil, fmt.Errorf("failed to extract value of 'api-key' from %s: %v", eksacloudStackCloudConfigB64SecretKey, err)
	}
	secretKey, err := section.GetKey("secret-key")
	if err != nil {
		return nil, fmt.Errorf("failed to extract value of 'secret-key' from %s: %v", eksacloudStackCloudConfigB64SecretKey, err)
	}
	apiUrl, err := section.GetKey("api-url")
	if err != nil {
		return nil, fmt.Errorf("failed to extract value of 'api-url' from %s: %v", eksacloudStackCloudConfigB64SecretKey, err)
	}
	return &CloudStackExecConfig{
		CloudStackApiKey:        apiKey.Value(),
		CloudStackSecretKey:     secretKey.Value(),
		CloudStackManagementUrl: apiUrl.Value(),
	}, nil
}

type CloudStackExecConfig struct {
	CloudStackApiKey        string
	CloudStackSecretKey     string
	CloudStackManagementUrl string
	CloudMonkeyVerifyCert   string
}
