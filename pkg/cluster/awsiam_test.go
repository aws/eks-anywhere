package cluster_test

import (
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

func TestConfigManagerValidateAWSIamConfig(t *testing.T) {
	tests := []struct {
		testName string
		config   *cluster.Config
		wantErr  bool
	}{
		{
			testName: "valid awsiamconfig",
			config: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					TypeMeta: metav1.TypeMeta{
						Kind:       anywherev1.ClusterKind,
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "eksa-unit-test",
						Namespace: "default",
					},
					Spec: anywherev1.ClusterSpec{
						IdentityProviderRefs: []anywherev1.Ref{
							{
								Kind: "AWSIamConfig",
								Name: "test",
							},
						},
					},
				},
				AWSIAMConfigs: map[string]*anywherev1.AWSIamConfig{
					"test": {
						TypeMeta: metav1.TypeMeta{
							Kind:       "AWSIamConfig",
							APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "eksa-unit-test",
							Namespace: "default",
						},
						Spec: anywherev1.AWSIamConfigSpec{
							AWSRegion:   "test-region",
							BackendMode: []string{"mode1", "mode2"},
							MapRoles: []anywherev1.MapRoles{
								{
									RoleARN:  "test-role-arn",
									Username: "test",
									Groups:   []string{"group1", "group2"},
								},
							},
							MapUsers: []anywherev1.MapUsers{
								{
									UserARN:  "test-user-arn",
									Username: "test",
									Groups:   []string{"group1", "group2"},
								},
							},
							Partition: "aws",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "different namespace",
			config: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					TypeMeta: metav1.TypeMeta{
						Kind:       anywherev1.ClusterKind,
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "eksa-unit-test",
						Namespace: "default",
					},
					Spec: anywherev1.ClusterSpec{
						IdentityProviderRefs: []anywherev1.Ref{
							{
								Name: "test1",
								Kind: "AWSIamConfig",
							},
						},
					},
				},
				AWSIAMConfigs: map[string]*anywherev1.AWSIamConfig{
					"test": {
						TypeMeta: metav1.TypeMeta{
							Kind:       "AWSIamConfig",
							APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "eksa-unit-test",
							Namespace: "not-default",
						},
						Spec: anywherev1.AWSIamConfigSpec{
							AWSRegion:   "test-region",
							BackendMode: []string{"mode1", "mode2"},
							MapRoles: []anywherev1.MapRoles{
								{
									RoleARN:  "test-role-arn",
									Username: "test",
									Groups:   []string{"group1", "group2"},
								},
							},
							MapUsers: []anywherev1.MapUsers{
								{
									UserARN:  "test-user-arn",
									Username: "test",
									Groups:   []string{"group1", "group2"},
								},
							},
							Partition: "aws",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			testName: "no awsiam config",
			config: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					TypeMeta: metav1.TypeMeta{
						Kind:       anywherev1.ClusterKind,
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "eksa-unit-test",
						Namespace: "default",
					},
					Spec: anywherev1.ClusterSpec{
						IdentityProviderRefs: []anywherev1.Ref{
							{
								Name: "test1", Kind: "AWSIamConfig",
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			c := cluster.NewConfigManager()
			c.RegisterValidations(
				func(c *cluster.Config) error {
					for _, a := range c.AWSIAMConfigs {
						if a.Namespace != c.Cluster.Namespace {
							return fmt.Errorf("%s and Cluster objects must have the same namespace specified", anywherev1.AWSIamConfigKind)
						}
					}
					return nil
				},
				func(c *cluster.Config) error {
					for _, idr := range c.Cluster.Spec.IdentityProviderRefs {
						if idr.Kind == anywherev1.AWSIamConfigKind && c.AWSIAMConfigs == nil {
							return fmt.Errorf("%s/%s referenced in Cluster but not present in the cluster config", anywherev1.AWSIamConfigKind, idr.Name)
						}
					}
					return nil
				},
			)

			err := c.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
