package curatedpackages

import (
	"fmt"
	"strings"
)

func ValidateKubeVersion(kubeVersion string, clusterName string) error {
	if len(clusterName) > 0 {
		if len(kubeVersion) > 0 {
			return fmt.Errorf("please specify either kube-version or cluster name not both")
		}
		return nil
	}

	if len(kubeVersion) > 0 {
		versionSplit := strings.Split(kubeVersion, ".")
		if len(versionSplit) != 2 {
			return fmt.Errorf("please specify kube-version as <major>.<minor>")
		}
		return nil
	}
	return fmt.Errorf("please specify kube-version or cluster name")
}
