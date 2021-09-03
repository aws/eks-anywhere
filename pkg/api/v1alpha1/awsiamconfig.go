package v1alpha1

import (
	"fmt"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const AWSIamConfigKind = "AWSIamConfig"

func GetAndValidateAWSIamConfig(fileName string, refName string) (*AWSIamConfig, error) {
	config, err := getAWSIamConfig(fileName)
	if err != nil {
		return nil, err
	}
	err = validateAWSIamConfig(config, refName)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func getAWSIamConfig(fileName string) (*AWSIamConfig, error) {
	var config AWSIamConfig
	err := ParseClusterConfig(fileName, &config)
	if err != nil {
		return nil, err
	}
	// If the name is empty, we can assume that they didn't configure their AWS IAM configuration, so return nil
	if config.Name == "" {
		return nil, nil
	}
	return &config, nil
}

func validateAWSIamConfig(config *AWSIamConfig, refName string) error {
	if config == nil {
		return nil
	}
	if config.Name != refName {
		return fmt.Errorf("AWSIamConfig retrieved with name %v does not match name (%v) specified in "+
			"identityProviderRefs", config.Name, refName)
	}
	if config.Spec.ClusterID == "" {
		return fmt.Errorf("AWSIamConfig ClusterID is a required field")
	}
	if config.Spec.BackendMode == "" {
		return fmt.Errorf("AWSIamConfig BackendMode is a required field")
	}
	if strings.Contains(config.Spec.BackendMode, "MountedFile") && config.Spec.Data == "" {
		logger.Info("Warning: AWS IAM Authenticator MountedFile data is empty. Please be aware this will prevent aws-iam-authenticator from mapping IAM roles to users/groups on the cluster with MountedFile backend")
	}
	return nil
}
