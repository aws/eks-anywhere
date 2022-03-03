package decoder

import (
	b64 "encoding/base64"
	"fmt"
	"os"

	"gopkg.in/ini.v1"
)

const EksacloudStackCloudConfigB64SecretKey = "EKSA_CLOUDSTACK_B64ENCODED_SECRET"

// ParseCloudStackSecret parses the input b64 string into the ini object to extract out the api key, secret key, and url
func ParseCloudStackSecret() (*CloudStackExecConfig, error) {
	cloudStackB64EncodedSecret, ok := os.LookupEnv(EksacloudStackCloudConfigB64SecretKey)
	if !ok {
		return nil, fmt.Errorf("%s is not set or is empty", EksacloudStackCloudConfigB64SecretKey)
	}
	decodedString, err := b64.StdEncoding.DecodeString(cloudStackB64EncodedSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode value for %s with base64: %v", EksacloudStackCloudConfigB64SecretKey, err)
	}
	cfg, err := ini.Load(decodedString)
	if err != nil {
		return nil, fmt.Errorf("failed to extract values from %s with ini: %v", EksacloudStackCloudConfigB64SecretKey, err)
	}
	section, err := cfg.GetSection("Global")
	if err != nil {
		return nil, fmt.Errorf("failed to extract section 'Global' from %s: %v", EksacloudStackCloudConfigB64SecretKey, err)
	}
	apiKey, err := section.GetKey("api-key")
	if err != nil {
		return nil, fmt.Errorf("failed to extract value of 'api-key' from %s: %v", EksacloudStackCloudConfigB64SecretKey, err)
	}
	secretKey, err := section.GetKey("secret-key")
	if err != nil {
		return nil, fmt.Errorf("failed to extract value of 'secret-key' from %s: %v", EksacloudStackCloudConfigB64SecretKey, err)
	}
	apiUrl, err := section.GetKey("api-url")
	if err != nil {
		return nil, fmt.Errorf("failed to extract value of 'api-url' from %s: %v", EksacloudStackCloudConfigB64SecretKey, err)
	}
	verifySsl, err := section.GetKey("verify-ssl")
	if err != nil {
		return nil, fmt.Errorf("failed to extract value of 'verify-ssl' from %s: %v", EksacloudStackCloudConfigB64SecretKey, err)
	}
	return &CloudStackExecConfig{
		ApiKey:        apiKey.Value(),
		SecretKey:     secretKey.Value(),
		ManagementUrl: apiUrl.Value(),
		VerifySsl:     verifySsl.Value(),
	}, nil
}

type CloudStackExecConfig struct {
	ApiKey        string
	SecretKey     string
	ManagementUrl string
	VerifySsl     string
}
