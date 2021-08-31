package eksctl

import (
	"fmt"
	"os"
)

const VersionEnvVar = "EKSCTL_VERSION"

var enabled string

func Enabled() bool {
	return enabled == "true"
}

func ValidateVersion() error {
	if os.Getenv(VersionEnvVar) == "" {
		return fmt.Errorf("unable to retrieve version. Please use the 'eksctl anywhere' command to use EKS-A")
	}
	return nil
}
