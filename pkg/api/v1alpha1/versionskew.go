package v1alpha1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/version"
)

// SupportedMinorVersionIncrement represents the minor version skew for kubernetes version upgrades.
const SupportedMinorVersionIncrement = 1

// ValidateVersionSkew validates Kubernetes version skew between valid non-nil versions.
func ValidateVersionSkew(oldVersion, newVersion *version.Version) error {
	if newVersion.LessThan(oldVersion) {
		return fmt.Errorf("kubernetes version downgrade is not supported (%s) -> (%s)", oldVersion, newVersion)
	}

	newVersionMinor := newVersion.Minor()
	oldVersionMinor := oldVersion.Minor()

	minorVersionDifference := int(newVersionMinor) - int(oldVersionMinor)

	if minorVersionDifference < 0 || minorVersionDifference > SupportedMinorVersionIncrement {
		return fmt.Errorf("only +%d minor version skew is supported, minor version skew detected %v", SupportedMinorVersionIncrement, minorVersionDifference)
	}

	return nil
}
