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

func setDefaultAWSIamPartition(config *AWSIamConfig) {
	if config.Spec.Partition == "" {
		config.Spec.Partition = "aws"
		logger.V(1).Info("AWSIamConfig Partition is empty. Using default partition 'aws'")
	}
}
