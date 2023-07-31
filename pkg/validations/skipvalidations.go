package validations

import (
	"fmt"
	"strings"
)

// string values of supported validation names that can be skipped.
const (
	PDB             = "pod-disruption"
	VSphereUserPriv = "vsphere-user-privilege"
	EksaVersionSkew = "eksa-version-skew"
)

// ValidSkippableValidationsMap returns a map for all valid skippable validations as keys, defaulting values to false.
// Defaulting to False means these validations won't be skipped unless set to True.
func validSkippableValidationsMap(skippableValidations []string) map[string]bool {
	validationsMap := make(map[string]bool, len(skippableValidations))

	for i := range skippableValidations {
		validationsMap[skippableValidations[i]] = false
	}

	return validationsMap
}

// ValidateSkippableValidation validates if provided validations are supported by EKSA to skip for upgrades.
func ValidateSkippableValidation(skippedValidations []string, skippableValidations []string) (map[string]bool, error) {
	svMap := validSkippableValidationsMap(skippableValidations)

	for i := range skippedValidations {
		validationName := skippedValidations[i]
		_, ok := svMap[validationName]
		if !ok {
			return nil, fmt.Errorf("invalid validation name to be skipped. The supported validations that can be skipped using --skip-validations are %s", strings.Join(skippableValidations[:], ","))
		}
		svMap[validationName] = true
	}

	return svMap, nil
}
