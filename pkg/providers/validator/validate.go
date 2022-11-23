package validator

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/providers"
)

func ValidateSupportedProviderCreate(provider providers.Provider) error {
	if !features.IsActive(features.SnowProvider()) && provider.Name() == constants.SnowProviderName {
		return fmt.Errorf("provider snow is not supported in this release")
	}

	return nil
}
