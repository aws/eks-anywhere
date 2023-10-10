package validations

import (
	"errors"
	"fmt"
	"os"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func ValidateClusterNameArg(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("please specify a cluster name")
	}
	err := v1alpha1.ValidateClusterName(args[0])
	if err != nil {
		return args[0], err
	}
	err = v1alpha1.ValidateClusterNameLength(args[0])
	if err != nil {
		return args[0], err
	}
	return args[0], nil
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func FileExistsAndIsNotEmpty(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && info.Size() > 0
}

// ValidateClusterNameFromCommandAndConfig validates if cluster name provided in command matches with cluster name in config file.
func ValidateClusterNameFromCommandAndConfig(args []string, clusterNameConfig string) error {
	if len(args) != 0 {
		clusterNameCli, err := ValidateClusterNameArg(args)
		if err != nil {
			return fmt.Errorf("please provide a valid <cluster-name>")
		}
		if clusterNameCli != clusterNameConfig {
			return fmt.Errorf("please make sure cluster name provided in command matches with cluster name in config file")
		}
	}
	return nil
}
