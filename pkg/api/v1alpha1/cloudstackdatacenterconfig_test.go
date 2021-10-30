package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetCloudstackDatacenterConfig(t *testing.T) {
	tests := []struct {
		testName              string
		fileName              string
		wantCloudstackDatacenter *CloudstackDatacenterConfig
		wantErr               bool
	}{
		{
			testName:              "file doesn't exist",
			fileName:              "testdata/fake_file.yaml",
			wantCloudstackDatacenter: nil,
			wantErr:               true,
		},
		{
			testName:              "not parseable file",
			fileName:              "testdata/not_parseable_cluster_cloudstack.yaml",
			wantCloudstackDatacenter: nil,
			wantErr:               true,
		},
		{
			testName: "valid 1.18",
			fileName: "testdata/cluster_1_18_cloudstack.yaml",
			wantCloudstackDatacenter: &CloudstackDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       CloudstackDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: CloudstackDatacenterConfigSpec{
					Domain: "domain1",
					Zone: "zone1",
					Project: "",
					Account: "admin",
					Network: "net1",
					ControlPlaneEndpoint: "https://127.0.0.1:8080/client/api",
					Insecure:   false,
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.19",
			fileName: "testdata/cluster_1_19_cloudstack.yaml",
			wantCloudstackDatacenter: &CloudstackDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       CloudstackDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: CloudstackDatacenterConfigSpec{
					Domain: "domain1",
					Zone: "zone1",
					Project: "",
					Account: "admin",
					Network: "net1",
					ControlPlaneEndpoint: "https://127.0.0.1:8080/client/api",
					Insecure:   false,
				},
			},
			wantErr: false,
		},
		{
			testName: "valid with extra delimiters",
			fileName: "testdata/cluster_extra_delimiters_cloudstack.yaml",
			wantCloudstackDatacenter: &CloudstackDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       CloudstackDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: CloudstackDatacenterConfigSpec{
					Domain: "domain1",
					Zone: "zone1",
					Project: "",
					Account: "admin",
					Network: "net1",
					ControlPlaneEndpoint: "https://127.0.0.1:8080/client/api",
					Insecure:   false,
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.20",
			fileName: "testdata/cluster_1_20_cloudstack.yaml",
			wantCloudstackDatacenter: &CloudstackDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       CloudstackDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: CloudstackDatacenterConfigSpec{
					Domain: "domain1",
					Zone: "zone1",
					Project: "",
					Account: "admin",
					Network: "net1",
					ControlPlaneEndpoint: "https://127.0.0.1:8080/client/api",
					Insecure:   false,
				},
			},
			wantErr: false,
		},
		{
			testName:              "invalid kind",
			fileName:              "testdata/cluster_invalid_kinds.yaml",
			wantCloudstackDatacenter: nil,
			wantErr:               true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := GetCloudstackDatacenterConfig(tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetCloudstackDatacenterConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.wantCloudstackDatacenter) {
				t.Fatalf("GetCloudstackDatacenterConfig() = %#v, want %#v", got, tt.wantCloudstackDatacenter)
			}
		})
	}
}
