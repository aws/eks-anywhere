package v1alpha1

import (
	"fmt"
	"net/url"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const OIDCConfigKind = "OIDCConfig"

func GetAndValidateOIDCConfig(fileName string, refName string, clusterConfig *Cluster) (*OIDCConfig, error) {
	config, err := getOIDCConfig(fileName)
	if err != nil {
		return nil, err
	}
	if errs := validateOIDCConfig(config); len(errs) != 0 {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind(OIDCConfigKind).GroupKind(), config.Name, errs)
	}
	if err = validateOIDCRefName(config, refName); err != nil {
		return nil, err
	}
	if err = validateOIDCNamespace(config, clusterConfig); err != nil {
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

func validateOIDCConfig(config *OIDCConfig) field.ErrorList {
	var errs field.ErrorList

	if config == nil {
		return nil
	}

	if config.Spec.ClientId == "" {
		errs = append(errs, field.Invalid(field.NewPath("spec", "clientId"), config.Spec.ClientId, "OIDCConfig clientId is required"))
	}
	if len(config.Spec.RequiredClaims) > 1 {
		errs = append(errs, field.Invalid(field.NewPath("spec", "requiredClaims"), config.Spec.RequiredClaims, "only one OIDConfig requiredClaim is supported at this time"))
	}
	if config.Spec.IssuerUrl == "" {
		errs = append(errs, field.Invalid(field.NewPath("spec", "issuerUrl"), config.Spec.IssuerUrl, "OIDCConfig issuerUrl is required"))
		return errs
	}

	u, err := url.ParseRequestURI(config.Spec.IssuerUrl)
	if err != nil {
		errs = append(errs, field.Invalid(field.NewPath("spec", "issuerUrl"), config.Spec.IssuerUrl, fmt.Sprintf("OIDCConfig issuerUrl is invalid: %v", err)))
		return errs
	}

	if u.Scheme != "https" {
		errs = append(errs, field.Invalid(field.NewPath("spec", "issuerUrl"), config.Spec.IssuerUrl, "OIDCConfig issuerUrl should have HTTPS scheme"))
	}

	return errs
}

func validateOIDCRefName(config *OIDCConfig, refName string) error {
	if config == nil {
		return nil
	}

	if config.Name != refName {
		return fmt.Errorf("OIDCConfig retrieved with name %v does not match name (%s) specified in "+
			"identityProviderRefs", config.Name, refName)
	}

	return nil
}

func validateOIDCNamespace(config *OIDCConfig, clusterConfig *Cluster) error {
	if config == nil {
		return nil
	}

	if config.Namespace != clusterConfig.Namespace {
		return fmt.Errorf("OIDCConfig and Cluster objects must have the same namespace specified")
	}

	return nil
}
