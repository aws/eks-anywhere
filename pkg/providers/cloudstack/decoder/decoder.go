package decoder

import (
	b64 "encoding/base64"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/ini.v1"
)

const (
	EksacloudStackCloudConfigB64SecretKey = "EKSA_CLOUDSTACK_B64ENCODED_SECRET"
	CloudStackCloudConfigB64SecretKey     = "CLOUDSTACK_B64ENCODED_SECRET"
	EksaCloudStackHostPathToMount         = "EKSA_CLOUDSTACK_HOST_PATHS_TO_MOUNT"
	CloudStackGlobalAZ                    = "Global"
)

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

	cloudstackProfiles := []CloudStackProfileConfig{}
	sections := cfg.Sections()
	for _, section := range sections {
		if section.Name() == "DEFAULT" {
			continue
		}

		apiKey, err := section.GetKey("api-key")
		if err != nil {
			return nil, fmt.Errorf("failed to extract value of 'api-key' from %s: %v", section.Name(), err)
		}
		secretKey, err := section.GetKey("secret-key")
		if err != nil {
			return nil, fmt.Errorf("failed to extract value of 'secret-key' from %s: %v", EksacloudStackCloudConfigB64SecretKey, err)
		}
		apiUrl, err := section.GetKey("api-url")
		if err != nil {
			return nil, fmt.Errorf("failed to extract value of 'api-url' from %s: %v", EksacloudStackCloudConfigB64SecretKey, err)
		}
		verifySslValue := "true"
		if verifySsl, err := section.GetKey("verify-ssl"); err == nil {
			verifySslValue = verifySsl.Value()
			if _, err := strconv.ParseBool(verifySslValue); err != nil {
				return nil, fmt.Errorf("'verify-ssl' has invalid boolean string %s: %v", verifySslValue, err)
			}
		}
		cloudstackProfiles = append(cloudstackProfiles, CloudStackProfileConfig{
			Name:          section.Name(),
			ApiKey:        apiKey.Value(),
			SecretKey:     secretKey.Value(),
			ManagementUrl: apiUrl.Value(),
			VerifySsl:     verifySslValue,
		})
	}

	if len(cloudstackProfiles) == 0 {
		return nil, fmt.Errorf("no instance found from %s", EksacloudStackCloudConfigB64SecretKey)
	}

	return &CloudStackExecConfig{
		Profiles: cloudstackProfiles,
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
