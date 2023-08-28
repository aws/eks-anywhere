package v1alpha1

import (
	"fmt"
	"strings"
)

// EKSAReleaseKind is the Kind of EKSARelease.
const EKSAReleaseKind = "EKSARelease"

// Generates the naming convention of EKSARelease from a version.
func GenerateEKSAReleaseName(version string) string {
	version = strings.ReplaceAll(version, "+", "-plus-")
	return fmt.Sprintf("eksa-%s", strings.ReplaceAll(version, ".", "-"))
}
