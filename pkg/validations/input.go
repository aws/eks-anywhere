package validations

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

func ValidateClusterNameArg(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("please specify a cluster name")
	}
	err := ClusterName(args[0])
	if err != nil {
		return args[0], err
	}
	err = ClusterNameLength(args[0])
	if err != nil {
		return args[0], err
	}
	return args[0], nil
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func KubeConfigExists(dir, clusterName string, kubeConfigFileOverride string, kubeconfigPattern string) bool {
	kubeConfigFile := kubeConfigFileOverride
	if kubeConfigFile == "" {
		kubeConfigFile = filepath.Join(dir, fmt.Sprintf(kubeconfigPattern, clusterName))
	}

	if info, err := os.Stat(kubeConfigFile); err == nil && info.Size() > 0 {
		return true
	}
	return false
}

func ClusterName(clusterName string) error {
	// this regex will not work for AWS provider as CFN has restrictions with UPPERCASE chars;
	// if you are using AWS provider please use only lowercase chars
	allowedClusterNameRegex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]+$`)
	if !allowedClusterNameRegex.MatchString(clusterName) {
		return fmt.Errorf("%v is not a valid cluster name, cluster names must start with lowercase/uppercase letters and can include numbers and dashes. For instance 'testCluster-123' is a valid name but '123testCluster' is not. ", clusterName)
	}
	return nil
}

func ClusterNameLength(clusterName string) error {
	// vSphere has the maximum length for clusters to be 80 chars
	if len(clusterName) > 80 {
		return fmt.Errorf("number of characters in %v should be less than 81", clusterName)
	}
	return nil
}
