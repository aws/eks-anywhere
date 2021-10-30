package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetCloudstackMachineConfigs(t *testing.T) {
	tests := []struct {
		testName                  string
		fileName                  string
		wantCloudstackMachineConfigs map[string]*CloudstackMachineConfig
		wantErr                   bool
	}{
		{
			testName:                  "file doesn't exist",
			fileName:                  "testdata/fake_file.yaml",
			wantCloudstackMachineConfigs: nil,
			wantErr:                   true,
		},
		{
			testName:                  "not parseable file",
			fileName:                  "testdata/not_parseable_cluster.yaml",
			wantCloudstackMachineConfigs: nil,
			wantErr:                   true,
		},
		{
			testName: "valid 1.18",
			fileName: "testdata/cluster_1_18_cloudstack.yaml",
			wantCloudstackMachineConfigs: map[string]*CloudstackMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       CloudstackMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: CloudstackMachineConfigSpec{
						Template:  "centos7-k8s-118",
						ComputeOffering: "m4-large",
						DiskOffering: "ssd-100GB",
						OSFamily:  Ubuntu,
						KeyPair: "cloudstack-keypair",
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.19",
			fileName: "testdata/cluster_1_19_cloudstack.yaml",
			wantCloudstackMachineConfigs: map[string]*CloudstackMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       CloudstackMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: CloudstackMachineConfigSpec{
						Template: "centos7-k8s-119",
						ComputeOffering: "m4-large",
						DiskOffering: "ssd-100GB",
						OSFamily:  Ubuntu,
						KeyPair: "cloudstack-keypair",
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid with extra delimiters",
			fileName: "testdata/cluster_extra_delimiters_cloudstack.yaml",
			wantCloudstackMachineConfigs: map[string]*CloudstackMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       CloudstackMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: CloudstackMachineConfigSpec{
						Template: "centos7-k8s-118",
						ComputeOffering: "m4-large",
						DiskOffering: "ssd-100GB",
						OSFamily:  Ubuntu,
						KeyPair: "cloudstack-keypair",
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.20",
			fileName: "testdata/cluster_1_20_cloudstack.yaml",
			wantCloudstackMachineConfigs: map[string]*CloudstackMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       CloudstackMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: CloudstackMachineConfigSpec{
						Template: "centos7-k8s-120",
						ComputeOffering: "m4-large",
						DiskOffering: "ssd-100GB",
						OSFamily:  Ubuntu,
						KeyPair: "cloudstack-keypair",
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid different machine configs",
			fileName: "testdata/cluster_different_machine_configs_cloudstack.yaml",
			wantCloudstackMachineConfigs: map[string]*CloudstackMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       CloudstackMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: CloudstackMachineConfigSpec{
						Template: "centos7-k8s-118",
						ComputeOffering: "m4-large",
						DiskOffering: "ssd-100GB",
						OSFamily:  Ubuntu,
						KeyPair: "cloudstack-keypair",
					},
				},
				"eksa-unit-test-2": {
					TypeMeta: metav1.TypeMeta{
						Kind:       CloudstackMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test-2",
					},
					Spec: CloudstackMachineConfigSpec{
						Template: "centos7-k8s-118",
						ComputeOffering: "m5-xlarge",
						DiskOffering: "ssd-100GB",
						OSFamily:  Ubuntu,
						KeyPair: "cloudstack-keypair",
					},
				},
			},
			wantErr: false,
		},
		{
			testName:                  "invalid kind",
			fileName:                  "testdata/cluster_invalid_kinds.yaml",
			wantCloudstackMachineConfigs: nil,
			wantErr:                   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := GetCloudstackMachineConfigs(tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetCloudstackMachineConfigs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.wantCloudstackMachineConfigs) {
				t.Fatalf("GetCloudstackMachineConfigs() = %#v, want %#v", got, tt.wantCloudstackMachineConfigs)
			}
		})
	}
}
