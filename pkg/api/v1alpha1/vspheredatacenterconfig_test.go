package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetVSphereDatacenterConfig(t *testing.T) {
	tests := []struct {
		testName              string
		fileName              string
		wantVSphereDatacenter *VSphereDatacenterConfig
		wantErr               bool
	}{
		{
			testName:              "file doesn't exist",
			fileName:              "testdata/fake_file.yaml",
			wantVSphereDatacenter: nil,
			wantErr:               true,
		},
		{
			testName:              "not parseable file",
			fileName:              "testdata/not_parseable_cluster.yaml",
			wantVSphereDatacenter: nil,
			wantErr:               true,
		},
		{
			testName: "valid 1.18",
			fileName: "testdata/cluster_1_18.yaml",
			wantVSphereDatacenter: &VSphereDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       VSphereDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: VSphereDatacenterConfigSpec{
					Datacenter: "myDatacenter",
					Network:    "myNetwork",
					Server:     "myServer",
					Thumbprint: "myTlsThumbprint",
					Insecure:   false,
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.19",
			fileName: "testdata/cluster_1_19.yaml",
			wantVSphereDatacenter: &VSphereDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       VSphereDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: VSphereDatacenterConfigSpec{
					Datacenter: "myDatacenter",
					Network:    "myNetwork",
					Server:     "myServer",
					Thumbprint: "myTlsThumbprint",
					Insecure:   false,
				},
			},
			wantErr: false,
		},
		{
			testName: "valid with extra delimiters",
			fileName: "testdata/cluster_extra_delimiters.yaml",
			wantVSphereDatacenter: &VSphereDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       VSphereDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: VSphereDatacenterConfigSpec{
					Datacenter: "myDatacenter",
					Network:    "myNetwork",
					Server:     "myServer",
					Thumbprint: "myTlsThumbprint",
					Insecure:   false,
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.20",
			fileName: "testdata/cluster_1_20.yaml",
			wantVSphereDatacenter: &VSphereDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       VSphereDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: VSphereDatacenterConfigSpec{
					Datacenter: "myDatacenter",
					Network:    "myNetwork",
					Server:     "myServer",
					Thumbprint: "myTlsThumbprint",
					Insecure:   false,
				},
			},
			wantErr: false,
		},
		{
			testName:              "invalid kind",
			fileName:              "testdata/cluster_invalid_kinds.yaml",
			wantVSphereDatacenter: nil,
			wantErr:               true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := GetVSphereDatacenterConfig(tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetVSphereDatacenterConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.wantVSphereDatacenter) {
				t.Fatalf("GetVSphereDatacenterConfig() = %#v, want %#v", got, tt.wantVSphereDatacenter)
			}
		})
	}
}
