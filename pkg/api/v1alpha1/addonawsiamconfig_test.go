package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetAndValidateAddOnAWSIamConfig(t *testing.T) {
	c := &Cluster{
		ObjectMeta: metav1.ObjectMeta{
			ClusterName: "eksa-unit-test-cluster",
		},
	}
	tests := []struct {
		testName              string
		fileName              string
		refName               string
		wantAddOnAWSIamConfig *AddOnAWSIamConfig
		wantErr               bool
	}{
		{
			testName:              "file doesn't exist",
			fileName:              "testdata/fake_file.yaml",
			wantAddOnAWSIamConfig: nil,
			wantErr:               true,
		},
		{
			testName:              "invalid ref name",
			fileName:              "testdata/cluster_1_21_awsiam_invalid_refname.yaml",
			refName:               "eksa-unit-test",
			wantAddOnAWSIamConfig: nil,
			wantErr:               true,
		},
		{
			testName:              "invalid namespace",
			fileName:              "testdata/cluster_1_21_awsiam_invalid_namespace.yaml",
			refName:               "eksa-unit-test",
			wantAddOnAWSIamConfig: nil,
			wantErr:               true,
		},
		{
			testName:              "invalid AddOnAWSIamConfig no aws region",
			fileName:              "testdata/cluster_1_21_awsiam_no_awsregion.yaml",
			refName:               "eksa-unit-test",
			wantAddOnAWSIamConfig: nil,
			wantErr:               true,
		},
		{
			testName:              "invalid AddOnAWSIamConfig no aws region",
			fileName:              "testdata/cluster_1_21_awsiam_no_backendmode.yaml",
			refName:               "eksa-unit-test",
			wantAddOnAWSIamConfig: nil,
			wantErr:               true,
		},
		{
			testName:              "invalid AddOnAWSIamConfig unsupported MountedFile",
			fileName:              "testdata/cluster_1_21_awsiam_unsupported_mountedfile.yaml",
			refName:               "eksa-unit-test",
			wantAddOnAWSIamConfig: nil,
			wantErr:               true,
		},
		{
			testName:              "invalid AddOnAWSIamConfig no arn",
			fileName:              "testdata/cluster_1_21_awsiam_no_arn.yaml",
			refName:               "eksa-unit-test",
			wantAddOnAWSIamConfig: nil,
			wantErr:               true,
		},
		{
			testName: "valid AddOnAWSIamConfig default cluster id",
			fileName: "testdata/cluster_1_21_awsiam_no_clusterid.yaml",
			refName:  "eksa-unit-test",
			wantAddOnAWSIamConfig: &AddOnAWSIamConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AddOnAWSIamConfig",
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: AddOnAWSIamConfigSpec{
					AWSRegion:   "test-region",
					BackendMode: []string{"mode1", "mode2"},
					ClusterID:   "eksa-unit-test-cluster",
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
		{
			testName: "valid AddOnAWSIamConfig no mapping eksconfigmap backend",
			fileName: "testdata/cluster_1_21_awsiam_no_mapping_eksconfigmap.yaml",
			refName:  "eksa-unit-test",
			wantAddOnAWSIamConfig: &AddOnAWSIamConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AddOnAWSIamConfig",
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: AddOnAWSIamConfigSpec{
					AWSRegion:   "test-region",
					BackendMode: []string{"EKSConfigMap"},
					ClusterID:   "eksa-unit-test",
					Partition:   "aws",
				},
			},
			wantErr: false,
		},
		{
			testName: "valid AddOnAWSIamConfig",
			fileName: "testdata/cluster_1_21_awsiam.yaml",
			refName:  "eksa-unit-test",
			wantAddOnAWSIamConfig: &AddOnAWSIamConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AddOnAWSIamConfig",
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: AddOnAWSIamConfigSpec{
					AWSRegion:   "test-region",
					BackendMode: []string{"mode1", "mode2"},
					ClusterID:   "eksa-unit-test",
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
			got, err := GetAndValidateAddOnAWSIamConfig(tt.fileName, tt.refName, c)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetAndValidateAddOnAWSIamConfig() error = %v\nwantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.wantAddOnAWSIamConfig) {
				t.Fatalf("GetAndValidateAddOnAWSIamConfig() = %v\nwant %v", got, tt.wantAddOnAWSIamConfig)
			}
		})
	}
}
