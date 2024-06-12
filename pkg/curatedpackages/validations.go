package curatedpackages

import (
	"fmt"
	"strings"
)

func ValidateKubeVersion(kubeVersion string, clusterName string) error {
	if len(kubeVersion) > 0 {
		versionSplit := strings.Split(kubeVersion, ".")
		if len(versionSplit) != 2 {
			return fmt.Errorf("please specify kube-version as <major>.<minor>")
		}
		return nil
	}

	if len(clusterName) > 0 {
		// no-op since either cluster name or kube-version is needed.
		return nil
	}
	return fmt.Errorf("please specify kube-version or cluster name")
}
