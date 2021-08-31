package v1alpha1

import (
	"fmt"
	"net/url"
)

const OIDCConfigKind = "OIDCConfig"

func GetAndValidateOIDCConfig(fileName string, refName string) (*OIDCConfig, error) {
	config, err := getOIDCConfig(fileName)
	if err != nil {
		return nil, err
	}
	err = validateOIDCConfig(config, refName)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func getOIDCConfig(fileName string) (*OIDCConfig, error) {
	var config OIDCConfig
	err := ParseClusterConfig(fileName, &config)
	if err != nil {
		return nil, err
	}
	// If the name is empty, we can assume that they didn't configure their OIDC configuration, so return nil
	if config.Name == "" {
		return nil, nil
	}
	return &config, nil
}

func validateOIDCConfig(config *OIDCConfig, refName string) error {
	if config == nil {
		return nil
	}
	if config.Name != refName {
		return fmt.Errorf("OIDCConfig retrieved with name %v does not match name (%v) specified in "+
			"identityProviderRefs", config.Name, refName)
	}
	if config.Spec.ClientId == "" {
		return fmt.Errorf("OIDCConfig clientId is required")
	}
	if config.Spec.IssuerUrl == "" {
		return fmt.Errorf("OIDCConfig issuerUrl is required")
	}

	u, err := url.ParseRequestURI(config.Spec.IssuerUrl)
	if err != nil {
		return fmt.Errorf("OIDCConfig issuerUrl is invalid: %v", err)
	}

	if u.Scheme != "https" {
		return fmt.Errorf("OIDCConfig issuerUrl should have HTTPS scheme")
	}

	if len(config.Spec.RequiredClaims) > 1 {
		return fmt.Errorf("only one OIDConfig RequiredClaim is supported at this time")
	}
	return nil
}
