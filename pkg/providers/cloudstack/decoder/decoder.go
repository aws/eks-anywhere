package decoder

import (
	b64 "encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
	apiv1 "k8s.io/api/core/v1"
)

const (
	EksacloudStackCloudConfigB64SecretKey = "EKSA_CLOUDSTACK_B64ENCODED_SECRET"
	CloudStackCloudConfigB64SecretKey     = "CLOUDSTACK_B64ENCODED_SECRET"
	EksaCloudStackHostPathToMount         = "EKSA_CLOUDSTACK_HOST_PATHS_TO_MOUNT"
	defaultVerifySslValue                 = "true"
	CloudStackGlobalAZ                    = "global"

	APIKeyKey    = "api-key"
	SecretKeyKey = "secret-key"
	APIUrlKey    = "api-url"
	VerifySslKey = "verify-ssl"
)

// ParseCloudStackCredsFromSecrets parses a list of secrets to extract out the api keys, secret keys, and urls.
func ParseCloudStackCredsFromSecrets(secrets []apiv1.Secret) (*CloudStackExecConfig, error) {
	if len(secrets) == 0 {
		return nil, fmt.Errorf("no secrets provided - unable to generate CloudStackExecConfig")
	}
	cloudstackProfiles := make([]CloudStackProfileConfig, 0, len(secrets))
	for _, secret := range secrets {
		apiKey, ok := secret.Data[APIKeyKey]
		if !ok {
			return nil, fmt.Errorf("secret %s is missing required key %s", secret.Name, APIKeyKey)
		}
		secretKey, ok := secret.Data[SecretKeyKey]
		if !ok {
			return nil, fmt.Errorf("secret %s is missing required key %s", secret.Name, SecretKeyKey)
		}
		apiURL, ok := secret.Data[APIUrlKey]
		if !ok {
			return nil, fmt.Errorf("secret %s is missing required key %s", secret.Name, APIUrlKey)
		}
		verifySsl, ok := secret.Data[VerifySslKey]
		if !ok {
			verifySsl = []byte(defaultVerifySslValue)
		}
		cloudstackProfiles = append(
			cloudstackProfiles,
			CloudStackProfileConfig{
				Name:          secret.Name,
				ApiKey:        string(apiKey),
				SecretKey:     string(secretKey),
				ManagementUrl: string(apiURL),
				VerifySsl:     string(verifySsl),
			},
		)
	}

	return &CloudStackExecConfig{
		Profiles: cloudstackProfiles,
	}, nil
}

// ParseCloudStackCredsFromEnv parses the input b64 string into the ini object to extract out the api key, secret key, and url.
func ParseCloudStackCredsFromEnv() (*CloudStackExecConfig, error) {
	cloudStackB64EncodedSecret, ok := os.LookupEnv(EksacloudStackCloudConfigB64SecretKey)
	if !ok {
		return nil, fmt.Errorf("%s is not set or is empty", EksacloudStackCloudConfigB64SecretKey)
	}
	decodedString, err := b64.StdEncoding.DecodeString(cloudStackB64EncodedSecret)
	if err != nil {
		return nil, fmt.Errorf("decoding value for %s with base64: %v", EksacloudStackCloudConfigB64SecretKey, err)
	}
	cfg, err := ini.Load(decodedString)
	if err != nil {
		return nil, fmt.Errorf("extracting values from %s with ini: %v", EksacloudStackCloudConfigB64SecretKey, err)
	}

	var cloudstackProfiles []CloudStackProfileConfig
	sections := cfg.Sections()
	for _, section := range sections {
		if section.Name() == "DEFAULT" {
			continue
		}

		profile, err := parseCloudStackProfileSection(section)
		if err != nil {
			return nil, err
		}
		cloudstackProfiles = append(cloudstackProfiles, *profile)
	}

	if len(cloudstackProfiles) == 0 {
		return nil, fmt.Errorf("no instance found from %s", EksacloudStackCloudConfigB64SecretKey)
	}

	return &CloudStackExecConfig{
		Profiles: cloudstackProfiles,
	}, nil
}

func parseCloudStackProfileSection(section *ini.Section) (*CloudStackProfileConfig, error) {
	apiKey, err := section.GetKey(APIKeyKey)
	if err != nil {
		return nil, fmt.Errorf("extracting value of 'api-key' from %s: %v", section.Name(), err)
	}
	secretKey, err := section.GetKey(SecretKeyKey)
	if err != nil {
		return nil, fmt.Errorf("extracting value of 'secret-key' from %s: %v", EksacloudStackCloudConfigB64SecretKey, err)
	}
	apiURL, err := section.GetKey(APIUrlKey)
	if err != nil {
		return nil, fmt.Errorf("extracting value of 'api-url' from %s: %v", EksacloudStackCloudConfigB64SecretKey, err)
	}
	verifySslValue := defaultVerifySslValue
	if verifySsl, err := section.GetKey(VerifySslKey); err == nil {
		verifySslValue = verifySsl.Value()
		if _, err := strconv.ParseBool(verifySslValue); err != nil {
			return nil, fmt.Errorf("'verify-ssl' has invalid boolean string %s: %v", verifySslValue, err)
		}
	}
	return &CloudStackProfileConfig{
		Name:          strings.ToLower(section.Name()),
		ApiKey:        apiKey.Value(),
		SecretKey:     secretKey.Value(),
		ManagementUrl: apiURL.Value(),
		VerifySsl:     verifySslValue,
	}, nil
}

type CloudStackExecConfig struct {
	Profiles []CloudStackProfileConfig
}

type CloudStackProfileConfig struct {
	Name          string
	ApiKey        string
	SecretKey     string
	ManagementUrl string
	VerifySsl     string
	Timeout       string
}
