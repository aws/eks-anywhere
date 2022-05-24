package v1alpha1

import (
	"fmt"
	"net/url"
)

const OIDCConfigKind = "OIDCConfig"

func validateOIDCConfig(config *OIDCConfig) error {
	if config == nil {
		return nil
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
