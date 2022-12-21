package v1alpha1

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	AWSIamConfigKind = "AWSIamConfig"
	eksConfigMap     = "EKSConfigMap"
	mountedFile      = "MountedFile"

	DefaultAWSIamConfigPartition = "aws"
)

func GetAndValidateAWSIamConfig(fileName string, refName string, clusterConfig *Cluster) (*AWSIamConfig, error) {
	config, err := getAWSIamConfig(fileName)
	if err != nil {
		return nil, err
	}
	config.SetDefaults()

	if err = validateAWSIamConfig(config); err != nil {
		return nil, err
	}
	if err = validateAWSIamRefName(config, refName); err != nil {
		return nil, err
	}
	if err = validateAWSIamNamespace(config, clusterConfig); err != nil {
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

func validateAWSIamConfig(config *AWSIamConfig) error {
	if config == nil {
		return nil
	}

	if config.Spec.AWSRegion == "" {
		return fmt.Errorf("AWSIamConfig AWSRegion is a required field")
	}
	if len(config.Spec.BackendMode) == 0 {
		return fmt.Errorf("AWSIamConfig BackendMode is a required field")
	}
	for _, backendMode := range config.Spec.BackendMode {
		if backendMode == eksConfigMap && len(config.Spec.MapRoles) == 0 && len(config.Spec.MapUsers) == 0 {
			logger.Info("Warning: AWS IAM Authenticator mapRoles and mapUsers specification is empty. Please be aware this will prevent aws-iam-authenticator from mapping IAM roles to users/groups on the cluster with backendMode EKSConfigMap")
		}
		if backendMode == mountedFile {
			return fmt.Errorf("AWSIamConfig BackendMode does not support %s backend", mountedFile)
		}
	}
	if err := validateMapRoles(config.Spec.MapRoles); err != nil {
		return err
	}
	if err := validateMapUsers(config.Spec.MapUsers); err != nil {
		return err
	}

	return nil
}

func validateMapRoles(mapRoles []MapRoles) error {
	for _, role := range mapRoles {
		if role.RoleARN == "" {
			return fmt.Errorf("AWSIamConfig MapRoles RoleARN is required")
		}
		if role.Username == "" {
			return fmt.Errorf("AWSIamConfig MapRoles Username is required")
		}
	}
	return nil
}

func validateMapUsers(mapUsers []MapUsers) error {
	for _, user := range mapUsers {
		if user.UserARN == "" {
			return fmt.Errorf("AWSIamConfig MapUsers UserARN is required")
		}
		if user.Username == "" {
			return fmt.Errorf("AWSIamConfig MapUsers Username is required")
		}
	}
	return nil
}

func validateAWSIamRefName(config *AWSIamConfig, refName string) error {
	if config == nil {
		return nil
	}

	if config.Name != refName {
		return fmt.Errorf("AWSIamConfig retrieved with name %s does not match name (%s) specified in "+
			"identityProviderRefs", config.Name, refName)
	}

	return nil
}

func validateAWSIamNamespace(config *AWSIamConfig, clusterConfig *Cluster) error {
	if config == nil {
		return nil
	}

	if config.Namespace != clusterConfig.Namespace {
		return fmt.Errorf("AWSIamConfig and Cluster objects must have the same namespace specified")
	}

	return nil
}

func setDefaultAWSIamPartition(config *AWSIamConfig) {
	if config.Spec.Partition == "" {
		config.Spec.Partition = DefaultAWSIamConfigPartition
		logger.V(1).Info("AWSIamConfig Partition is empty. Using default partition 'aws'")
	}
}
