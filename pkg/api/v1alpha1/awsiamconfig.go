package v1alpha1

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	AWSIamConfigKind = "AWSIamConfig"
	eksConfigMap     = "EKSConfigMap"
	mountedFile      = "MountedFile"
)

func GetAndValidateAWSIamConfig(fileName string, refName string, clusterConfig *Cluster) (*AWSIamConfig, error) {
	config, err := getAWSIamConfig(fileName)
	if err != nil {
		return nil, err
	}
	err = validateAWSIamConfig(config, refName, clusterConfig)
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

func validateAWSIamConfig(config *AWSIamConfig, refName string, clusterConfig *Cluster) error {
	if config == nil {
		return nil
	}
	if config.Name != refName {
		return fmt.Errorf("AWSIamConfig retrieved with name %v does not match name (%v) specified in "+
			"identityProviderRefs", config.Name, refName)
	}
	if config.Namespace != clusterConfig.Namespace {
		return fmt.Errorf("AWSIamConfig and Cluster objects must have the same namespace specified")
	}
	if config.Spec.AWSRegion == "" {
		return fmt.Errorf("AWSIamConfig AWSRegion is a required field")
	}
	if config.Spec.ClusterID == "" {
		config.Spec.ClusterID = clusterConfig.ClusterName
		logger.V(1).Info("AWSIamConfig ClusterID is empty. Using cluster name as default")
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
	if config.Spec.Partition == "" {
		config.Spec.Partition = "aws"
		logger.V(1).Info("AWSIamConfig Partition is empty. Using default partition 'aws'")
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
