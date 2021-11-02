package api

import (
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AWSIamConfigOpt func(c *v1alpha1.AWSIamConfig)

func NewAWSIamConfig(name string, opts ...AWSIamConfigOpt) *v1alpha1.AWSIamConfig {
	config := &v1alpha1.AWSIamConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeBuilder.GroupVersion.String(),
			Kind:       v1alpha1.AWSIamConfigKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.AWSIamConfigSpec{},
	}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

func WithAWSIamAWSRegion(awsRegion string) AWSIamConfigOpt {
	return func(c *v1alpha1.AWSIamConfig) {
		c.Spec.AWSRegion = awsRegion
	}
}

func WithAWSIamBackendMode(backendMode string) AWSIamConfigOpt {
	return func(c *v1alpha1.AWSIamConfig) {
		c.Spec.BackendMode = append(c.Spec.BackendMode, backendMode)
	}
}

func WithAWSIamClusterID(clusterId string) AWSIamConfigOpt {
	return func(c *v1alpha1.AWSIamConfig) {
		c.Spec.ClusterID = clusterId
	}
}

func WithAWSIamRole(arn, username string, groups []string) *v1alpha1.MapRoles {
	return &v1alpha1.MapRoles{
		RoleARN:  arn,
		Username: username,
		Groups:   groups,
	}
}

func AddAWSIamMapRoles(mapRoles *v1alpha1.MapRoles) AWSIamConfigOpt {
	return func(c *v1alpha1.AWSIamConfig) {
		c.Spec.MapRoles = append(c.Spec.MapRoles, *mapRoles)
	}
}

func WithAWSIamUser(arn, username string, groups []string) *v1alpha1.MapUsers {
	return &v1alpha1.MapUsers{
		UserARN:  arn,
		Username: username,
		Groups:   groups,
	}
}

func AddAWSIamMapUsers(mapUsers *v1alpha1.MapUsers) AWSIamConfigOpt {
	return func(c *v1alpha1.AWSIamConfig) {
		c.Spec.MapUsers = append(c.Spec.MapUsers, *mapUsers)
	}
}

func WithAWSIamPartition(partition string) AWSIamConfigOpt {
	return func(c *v1alpha1.AWSIamConfig) {
		c.Spec.Partition = partition
	}
}
