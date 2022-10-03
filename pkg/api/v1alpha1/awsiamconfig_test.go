package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetAndValidateAWSIamConfig(t *testing.T) {
	c := &Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "eksa-unit-test-cluster",
		},
	}
	tests := []struct {
		testName         string
		fileName         string
		refName          string
		wantAWSIamConfig *AWSIamConfig
		wantErr          bool
	}{
		{
			testName:         "file doesn't exist",
			fileName:         "testdata/fake_file.yaml",
			wantAWSIamConfig: nil,
			wantErr:          true,
		},
		{
			testName:         "invalid ref name",
			fileName:         "testdata/cluster_1_21_awsiam_invalid_refname.yaml",
			refName:          "eksa-unit-test",
			wantAWSIamConfig: nil,
			wantErr:          true,
		},
		{
			testName:         "invalid namespace",
			fileName:         "testdata/cluster_1_21_awsiam_invalid_namespace.yaml",
			refName:          "eksa-unit-test",
			wantAWSIamConfig: nil,
			wantErr:          true,
		},
		{
			testName:         "invalid AWSIamConfig no aws region",
			fileName:         "testdata/cluster_1_21_awsiam_no_awsregion.yaml",
			refName:          "eksa-unit-test",
			wantAWSIamConfig: nil,
			wantErr:          true,
		},
		{
			testName:         "invalid AWSIamConfig no aws region",
			fileName:         "testdata/cluster_1_21_awsiam_no_backendmode.yaml",
			refName:          "eksa-unit-test",
			wantAWSIamConfig: nil,
			wantErr:          true,
		},
		{
			testName:         "invalid AWSIamConfig unsupported MountedFile",
			fileName:         "testdata/cluster_1_21_awsiam_unsupported_mountedfile.yaml",
			refName:          "eksa-unit-test",
			wantAWSIamConfig: nil,
			wantErr:          true,
		},
		{
			testName:         "invalid AWSIamConfig no arn",
			fileName:         "testdata/cluster_1_21_awsiam_no_arn.yaml",
			refName:          "eksa-unit-test",
			wantAWSIamConfig: nil,
			wantErr:          true,
		},
		{
			testName: "valid AWSIamConfig no mapping eksconfigmap backend",
			fileName: "testdata/cluster_1_21_awsiam_no_mapping_eksconfigmap.yaml",
			refName:  "eksa-unit-test",
			wantAWSIamConfig: &AWSIamConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AWSIamConfig",
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: AWSIamConfigSpec{
					AWSRegion:   "test-region",
					BackendMode: []string{"EKSConfigMap"},
					Partition:   "aws",
				},
			},
			wantErr: false,
		},
		{
			testName: "valid AWSIamConfig",
			fileName: "testdata/cluster_1_21_awsiam.yaml",
			refName:  "eksa-unit-test",
			wantAWSIamConfig: &AWSIamConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AWSIamConfig",
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
			got, err := GetAndValidateAWSIamConfig(tt.fileName, tt.refName, c)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetAndValidateAWSIamConfig() error = %v\nwantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.wantAWSIamConfig) {
				t.Fatalf("GetAndValidateAWSIamConfig() = %v\nwant %v", got, tt.wantAWSIamConfig)
			}
		})
	}
}
