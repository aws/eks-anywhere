package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidateAWSIamConfig(t *testing.T) {
	tests := []struct {
		testName     string
		awsIamConfig *AWSIamConfig
		wantErr      bool
	}{
		{
			testName: "invalid AWSIamConfig no aws region",
			awsIamConfig: &AWSIamConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AWSIamConfig",
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: AWSIamConfigSpec{
					BackendMode: []string{"mode1", "mode2"},
					MapRoles: []MapRoles{
						{
							RoleARN:  "test-role-arn",
							Username: "test",
							Groups:   []string{"group1", "group2"},
						},
					},
					MapUsers: []MapUsers{
						{
							UserARN:  "test-user-arn",
							Username: "test",
							Groups:   []string{"group1", "group2"},
						},
					},
					Partition: "aws",
				},
			},
			wantErr: true,
		},
		{
			testName: "invalid AWSIamConfig no backend mode",
			awsIamConfig: &AWSIamConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       AWSIamConfigKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: AWSIamConfigSpec{
					AWSRegion: "test-region",
					MapRoles: []MapRoles{
						{
							RoleARN:  "test-role-arn",
							Username: "test",
							Groups:   []string{"group1", "group2"},
						},
					},
					MapUsers: []MapUsers{
						{
							UserARN:  "test-user-arn",
							Username: "test",
							Groups:   []string{"group1", "group2"},
						},
					},
					Partition: "aws",
				},
			},
			wantErr: true,
		},
		{
			testName: "invalid AWSIamConfig unsupported MountedFile",
			awsIamConfig: &AWSIamConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       AWSIamConfigKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: AWSIamConfigSpec{
					AWSRegion:   "test-region",
					BackendMode: []string{mountedFile},
					MapRoles: []MapRoles{
						{
							RoleARN:  "test-role-arn",
							Username: "test",
							Groups:   []string{"group1", "group2"},
						},
					},
					MapUsers: []MapUsers{
						{
							UserARN:  "test-user-arn",
							Username: "test",
							Groups:   []string{"group1", "group2"},
						},
					},
					Partition: "aws",
				},
			},
			wantErr: true,
		},
		{
			testName: "invalid AWSIamConfig no arn",
			awsIamConfig: &AWSIamConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       AWSIamConfigKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: AWSIamConfigSpec{
					AWSRegion:   "test-region",
					BackendMode: []string{"mode1", "mode2"},
					MapRoles: []MapRoles{
						{
							Username: "test",
							Groups:   []string{"group1", "group2"},
						},
					},
					MapUsers: []MapUsers{
						{
							Username: "test",
							Groups:   []string{"group1", "group2"},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			testName: "valid AWSIamConfig no mapping eksconfigmap backend",
			awsIamConfig: &AWSIamConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       AWSIamConfigKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: AWSIamConfigSpec{
					AWSRegion:   "test-region",
					BackendMode: []string{eksConfigMap},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid AWSIamConfig",
			awsIamConfig: &AWSIamConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       AWSIamConfigKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: AWSIamConfigSpec{
					AWSRegion:   "test-region",
					BackendMode: []string{"mode1", "mode2"},
					MapRoles: []MapRoles{
						{
							RoleARN:  "test-role-arn",
							Username: "test",
							Groups:   []string{"group1", "group2"},
						},
					},
					MapUsers: []MapUsers{
						{
							UserARN:  "test-user-arn",
							Username: "test",
							Groups:   []string{"group1", "group2"},
						},
					},
					Partition: "aws",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			err := tt.awsIamConfig.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("AWSIamConfig.Validate() error = %v\nwantErr %v", err, tt.wantErr)
			}
		})
	}
}
