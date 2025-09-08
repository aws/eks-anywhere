// Package common provides shared functionality for EKS Anywhere providers.
package common

import (
	_ "embed"
	"strings"
)

//go:embed config/admission-plugin-exclusion-rules.json
var admissionPluginExclusionRules string

// GetAdmissionPluginExclusionPolicy returns the admission plugin exclusion rules as a JSON string.
func GetAdmissionPluginExclusionPolicy() (string, error) {
	return strings.TrimSpace(admissionPluginExclusionRules), nil
}
