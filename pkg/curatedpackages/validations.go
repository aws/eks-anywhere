package curatedpackages

import (
	"fmt"
	"strings"
)

func ValidateKubeVersion(kubeVersion string, source BundleSource) error {
	if source != Registry {
		return nil
	}
	versionSplit := strings.Split(kubeVersion, ".")
	if len(versionSplit) != 2 {
		return fmt.Errorf("please specify kubeversion as <major>.<minor>")
	}
	return nil
}
