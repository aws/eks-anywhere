package upgradevalidations

import (
	"fmt"
	"strings"
)

// string values of supported validation names that can be skipped.
const (
	PDB = "pod-disruption"
)

// SkippableValidations represents all the validations we offer for users to skip.
var SkippableValidations = []string{
	PDB,
}

// ValidSkippableValidationsMap returns a map for all valid skippable validations as keys, defaulting values to false. Defaulting to False means these validations won't be skipped unless set to True.
func validSkippableValidationsMap() map[string]bool {
	validationsMap := make(map[string]bool, len(SkippableValidations))

	for i := range SkippableValidations {
		validationsMap[SkippableValidations[i]] = false
	}

	return validationsMap
}

// ValidateSkippableUpgradeValidation validates if provided validations are supported by EKSA to skip for upgrades.
func ValidateSkippableUpgradeValidation(skippedValidations []string) (map[string]bool, error) {
	svMap := validSkippableValidationsMap()

	for i := range skippedValidations {
		validationName := skippedValidations[i]
		_, ok := svMap[validationName]
		if !ok {
			return nil, fmt.Errorf("invalid validation name to be skipped. The supported upgrade validations that can be skipped using --skip-validations are %s", strings.Join(SkippableValidations[:], ","))
		}
		svMap[validationName] = true
	}

	return svMap, nil
}
