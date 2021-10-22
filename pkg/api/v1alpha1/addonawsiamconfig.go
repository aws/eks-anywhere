package v1alpha1

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	AddOnAWSIamConfigKind = "AddOnAWSIamConfig"
	eksConfigMap          = "EKSConfigMap"
	mountedFile           = "MountedFile"
)

func GetAndValidateAddOnAWSIamConfig(fileName string, refName string, clusterConfig *Cluster) (*AddOnAWSIamConfig, error) {
	config, err := getAddOnAWSIamConfig(fileName)
	if err != nil {
		return nil, err
	}
	err = validateAddOnAWSIamConfig(config, refName, clusterConfig)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func getAddOnAWSIamConfig(fileName string) (*AddOnAWSIamConfig, error) {
	var config AddOnAWSIamConfig
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

func validateAddOnAWSIamConfig(config *AddOnAWSIamConfig, refName string, clusterConfig *Cluster) error {
	if config == nil {
		return nil
	}
	if config.Name != refName {
		return fmt.Errorf("AddOnAWSIamConfig retrieved with name %v does not match name (%v) specified in "+
			"identityProviderRefs", config.Name, refName)
	}
	if config.Namespace != clusterConfig.Namespace {
		return fmt.Errorf("AddOnAWSIamConfig and Cluster objects must have the same namespace specified")
	}
	if config.Spec.AWSRegion == "" {
		return fmt.Errorf("AddOnAWSIamConfig AWSRegion is a required field")
	}
	if config.Spec.ClusterID == "" {
		config.Spec.ClusterID = clusterConfig.ClusterName
		logger.V(1).Info("AddOnAWSIamConfig ClusterID is empty. Using cluster name as default")
	}
	if len(config.Spec.BackendMode) == 0 {
		return fmt.Errorf("AddOnAWSIamConfig BackendMode is a required field")
	}
	for _, backendMode := range config.Spec.BackendMode {
		if backendMode == eksConfigMap && len(config.Spec.MapRoles) == 0 && len(config.Spec.MapUsers) == 0 {
			logger.Info("Warning: AWS IAM Authenticator mapRoles and mapUsers specification is empty. Please be aware this will prevent aws-iam-authenticator from mapping IAM roles to users/groups on the cluster with backendMode EKSConfigMap")
		}
		if backendMode == mountedFile {
			return fmt.Errorf("AddOnAWSIamConfig BackendMode does not support %s backend", mountedFile)
		}
	}

	if len(config.Spec.MapRoles) != 0 {
		err := validateMapRoles(config.Spec.MapRoles)
		if err != nil {
			return err
		}
	}

	if len(config.Spec.MapUsers) != 0 {
		err := validateMapUsers(config.Spec.MapUsers)
		if err != nil {
			return err
		}
	}

	if config.Spec.Partition == "" {
		config.Spec.Partition = "aws"
		logger.V(1).Info("AddOnAWSIamConfig Partition is empty. Using default partition 'aws'")
	}
	return nil
}

func validateMapRoles(mapRoles []MapRoles) error {
	for _, role := range mapRoles {
		if role.RoleARN == "" {
			return fmt.Errorf("AddOnAWSIamConfig MapRoles RoleARN is required")
		}
		if role.Username == "" {
			return fmt.Errorf("AddOnAWSIamConfig MapRoles Username is required")
		}
	}
	return nil
}

func validateMapUsers(mapUsers []MapUsers) error {
	for _, user := range mapUsers {
		if user.UserARN == "" {
			return fmt.Errorf("AddOnAWSIamConfig MapUsers UserARN is required")
		}
		if user.Username == "" {
			return fmt.Errorf("AddOnAWSIamConfig MapUsers Username is required")
		}
	}
	return nil
}
