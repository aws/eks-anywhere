package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetCloudStackMachineConfigs(t *testing.T) {
	tests := []struct {
		testName                     string
		fileName                     string
		wantCloudStackMachineConfigs map[string]*CloudStackMachineConfig
		wantErr                      bool
	}{
		{
			testName:                     "file doesn't exist",
			fileName:                     "testdata/fake_file.yaml",
			wantCloudStackMachineConfigs: nil,
			wantErr:                      true,
		},
		{
			testName:                     "not parseable file",
			fileName:                     "testdata/not_parseable_cluster.yaml",
			wantCloudStackMachineConfigs: nil,
			wantErr:                      true,
		},
		{
			testName: "valid 1.19",
			fileName: "testdata/cluster_1_19_cloudstack.yaml",
			wantCloudStackMachineConfigs: map[string]*CloudStackMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       CloudStackMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: CloudStackMachineConfigSpec{
						Template:        "centos7-k8s-119",
						ComputeOffering: "m4-large",
						DiskOffering:    "ssd-100GB",
						OSFamily:        Ubuntu,
						Users: []UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
						Details: map[string]string{
							"foo": "bar",
							"key": "value",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid with extra delimiters",
			fileName: "testdata/cluster_extra_delimiters_cloudstack.yaml",
			wantCloudStackMachineConfigs: map[string]*CloudStackMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       CloudStackMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: CloudStackMachineConfigSpec{
						Template:        "centos7-k8s-118",
						ComputeOffering: "m4-large",
						DiskOffering:    "ssd-100GB",
						OSFamily:        Ubuntu,
						Users: []UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
						Details: map[string]string{
							"foo": "bar",
							"key": "value",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.20",
			fileName: "testdata/cluster_1_20_cloudstack.yaml",
			wantCloudStackMachineConfigs: map[string]*CloudStackMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       CloudStackMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: CloudStackMachineConfigSpec{
						Template:        "centos7-k8s-120",
						ComputeOffering: "m4-large",
						DiskOffering:    "ssd-100GB",
						OSFamily:        Ubuntu,
						Users: []UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
						Details: map[string]string{
							"foo": "bar",
							"key": "value",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid different machine configs",
			fileName: "testdata/cluster_different_machine_configs_cloudstack.yaml",
			wantCloudStackMachineConfigs: map[string]*CloudStackMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       CloudStackMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: CloudStackMachineConfigSpec{
						Template:        "centos7-k8s-118",
						ComputeOffering: "m4-large",
						DiskOffering:    "ssd-100GB",
						OSFamily:        Ubuntu,
						Users: []UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
						Details: map[string]string{
							"foo": "bar",
							"key": "value",
						},
					},
				},
				"eksa-unit-test-2": {
					TypeMeta: metav1.TypeMeta{
						Kind:       CloudStackMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test-2",
					},
					Spec: CloudStackMachineConfigSpec{
						Template:        "centos7-k8s-118",
						ComputeOffering: "m5-xlarge",
						DiskOffering:    "ssd-100GB",
						OSFamily:        Ubuntu,
						Users: []UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
						Details: map[string]string{
							"foo": "bar",
							"key": "value",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName:                     "invalid kind",
			fileName:                     "testdata/cluster_invalid_kinds.yaml",
			wantCloudStackMachineConfigs: nil,
			wantErr:                      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := GetCloudStackMachineConfigs(tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetCloudStackMachineConfigs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.wantCloudStackMachineConfigs) {
				t.Fatalf("GetCloudStackMachineConfigs() = %#v, want %#v", got, tt.wantCloudStackMachineConfigs)
			}
		})
	}
}
