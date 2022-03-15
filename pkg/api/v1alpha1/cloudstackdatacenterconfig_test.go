package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetCloudStackDatacenterConfig(t *testing.T) {
	tests := []struct {
		testName                 string
		fileName                 string
		wantCloudStackDatacenter *CloudStackDatacenterConfig
		wantErr                  bool
	}{
		{
			testName:                 "file doesn't exist",
			fileName:                 "testdata/fake_file.yaml",
			wantCloudStackDatacenter: nil,
			wantErr:                  true,
		},
		{
			testName:                 "not parseable file",
			fileName:                 "testdata/not_parseable_cluster_cloudstack.yaml",
			wantCloudStackDatacenter: nil,
			wantErr:                  true,
		},
		{
			testName: "valid 1.19",
			fileName: "testdata/cluster_1_19_cloudstack.yaml",
			wantCloudStackDatacenter: &CloudStackDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       CloudStackDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: CloudStackDatacenterConfigSpec{
					Domain:  "domain1",
					Account: "admin",
					Zones: []CloudStackZoneRef{
						{
							Zone: CloudStackResourceRef{
								Value: "zone1",
								Type:  Name,
							}, Network: CloudStackResourceRef{
								Value: "net1",
								Type:  Name,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid with extra delimiters",
			fileName: "testdata/cluster_extra_delimiters_cloudstack.yaml",
			wantCloudStackDatacenter: &CloudStackDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       CloudStackDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: CloudStackDatacenterConfigSpec{
					Domain:  "domain1",
					Account: "admin",
					Zones: []CloudStackZoneRef{
						{
							Zone: CloudStackResourceRef{
								Value: "zone1",
								Type:  Name,
							}, Network: CloudStackResourceRef{
								Value: "net1",
								Type:  Name,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.20",
			fileName: "testdata/cluster_1_20_cloudstack.yaml",
			wantCloudStackDatacenter: &CloudStackDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       CloudStackDatacenterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: CloudStackDatacenterConfigSpec{
					Domain:  "domain1",
					Account: "admin",
					Zones: []CloudStackZoneRef{
						{
							Zone: CloudStackResourceRef{
								Value: "zone1",
								Type:  Name,
							}, Network: CloudStackResourceRef{
								Value: "net1",
								Type:  Name,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName:                 "invalid kind",
			fileName:                 "testdata/cluster_invalid_kinds.yaml",
			wantCloudStackDatacenter: nil,
			wantErr:                  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := GetCloudStackDatacenterConfig(tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetCloudStackDatacenterConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.wantCloudStackDatacenter) {
				t.Fatalf("GetCloudStackDatacenterConfig() = %#v, want %#v", got, tt.wantCloudStackDatacenter)
			}
		})
	}
}
