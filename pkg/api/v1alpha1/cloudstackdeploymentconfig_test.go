package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetCloudStackDeploymentConfig(t *testing.T) {
	tests := []struct {
		testName                 string
		fileName                 string
		wantCloudStackDeployment *CloudStackDeploymentConfig
		wantErr                  bool
	}{
		{
			testName:                 "file doesn't exist",
			fileName:                 "testdata/fake_file.yaml",
			wantCloudStackDeployment: nil,
			wantErr:                  true,
		},
		{
			testName:                 "not parseable file",
			fileName:                 "testdata/not_parseable_cluster_cloudstack.yaml",
			wantCloudStackDeployment: nil,
			wantErr:                  true,
		},
		{
			testName: "valid 1.19",
			fileName: "testdata/cluster_1_19_cloudstack.yaml",
			wantCloudStackDeployment: &CloudStackDeploymentConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       CloudStackDeploymentKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: CloudStackDeploymentConfigSpec{
					Domain:                "domain1",
					Zone:                  "zone1",
					Account:               "admin",
					Network:               "net1",
					ManagementApiEndpoint: "https://127.0.0.1:8080/client/api",
					Insecure:              false,
				},
			},
			wantErr: false,
		},
		{
			testName: "valid with extra delimiters",
			fileName: "testdata/cluster_extra_delimiters_cloudstack.yaml",
			wantCloudStackDeployment: &CloudStackDeploymentConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       CloudStackDeploymentKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: CloudStackDeploymentConfigSpec{
					Domain:                "domain1",
					Zone:                  "zone1",
					Account:               "admin",
					Network:               "net1",
					ManagementApiEndpoint: "https://127.0.0.1:8080/client/api",
					Insecure:              false,
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.20",
			fileName: "testdata/cluster_1_20_cloudstack.yaml",
			wantCloudStackDeployment: &CloudStackDeploymentConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       CloudStackDeploymentKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: CloudStackDeploymentConfigSpec{
					Domain:                "domain1",
					Zone:                  "zone1",
					Account:               "admin",
					Network:               "net1",
					ManagementApiEndpoint: "https://127.0.0.1:8080/client/api",
					Insecure:              false,
				},
			},
			wantErr: false,
		},
		{
			testName:                 "invalid kind",
			fileName:                 "testdata/cluster_invalid_kinds.yaml",
			wantCloudStackDeployment: nil,
			wantErr:                  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := GetCloudStackDeploymentConfig(tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetCloudStackDeploymentConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.wantCloudStackDeployment) {
				t.Fatalf("GetCloudStackDeploymentConfig() = %#v, want %#v", got, tt.wantCloudStackDeployment)
			}
		})
	}
}
