package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetAndValidateAWSIamConfig(t *testing.T) {
	c := &Cluster{}
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
					Partition:   "aws",
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
