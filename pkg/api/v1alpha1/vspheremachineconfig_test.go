package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetVSphereMachineConfigs(t *testing.T) {
	tests := []struct {
		testName                  string
		fileName                  string
		wantVSphereMachineConfigs map[string]*VSphereMachineConfig
		wantErr                   bool
	}{
		{
			testName:                  "file doesn't exist",
			fileName:                  "testdata/fake_file.yaml",
			wantVSphereMachineConfigs: nil,
			wantErr:                   true,
		},
		{
			testName:                  "not parseable file",
			fileName:                  "testdata/not_parseable_cluster.yaml",
			wantVSphereMachineConfigs: nil,
			wantErr:                   true,
		},
		{
			testName: "valid 1.18",
			fileName: "testdata/cluster_1_18.yaml",
			wantVSphereMachineConfigs: map[string]*VSphereMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       VSphereMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: VSphereMachineConfigSpec{
						DiskGiB:   25,
						MemoryMiB: 8192,
						NumCPUs:   2,
						OSFamily:  Ubuntu,
						Template:  "myTemplate",
						Users: []UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
						Datastore:         "myDatastore",
						Folder:            "myFolder",
						ResourcePool:      "myResourcePool",
						StoragePolicyName: "myStoragePolicyName",
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.19",
			fileName: "testdata/cluster_1_19.yaml",
			wantVSphereMachineConfigs: map[string]*VSphereMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       VSphereMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: VSphereMachineConfigSpec{
						DiskGiB:   25,
						MemoryMiB: 8192,
						NumCPUs:   2,
						OSFamily:  Ubuntu,
						Template:  "myTemplate",
						Users: []UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
						Datastore:         "myDatastore",
						Folder:            "myFolder",
						ResourcePool:      "myResourcePool",
						StoragePolicyName: "myStoragePolicyName",
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid with extra delimiters",
			fileName: "testdata/cluster_extra_delimiters.yaml",
			wantVSphereMachineConfigs: map[string]*VSphereMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       VSphereMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: VSphereMachineConfigSpec{
						DiskGiB:   25,
						MemoryMiB: 8192,
						NumCPUs:   2,
						OSFamily:  Ubuntu,
						Template:  "myTemplate",
						Users: []UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
						Datastore:         "myDatastore",
						Folder:            "myFolder",
						ResourcePool:      "myResourcePool",
						StoragePolicyName: "myStoragePolicyName",
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.20",
			fileName: "testdata/cluster_1_20.yaml",
			wantVSphereMachineConfigs: map[string]*VSphereMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       VSphereMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: VSphereMachineConfigSpec{
						DiskGiB:   25,
						MemoryMiB: 8192,
						NumCPUs:   2,
						OSFamily:  Ubuntu,
						Template:  "myTemplate",
						Users: []UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
						Datastore:         "myDatastore",
						Folder:            "myFolder",
						ResourcePool:      "myResourcePool",
						StoragePolicyName: "myStoragePolicyName",
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid different machine configs",
			fileName: "testdata/cluster_different_machine_configs.yaml",
			wantVSphereMachineConfigs: map[string]*VSphereMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       VSphereMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: VSphereMachineConfigSpec{
						DiskGiB:   25,
						MemoryMiB: 8192,
						NumCPUs:   2,
						OSFamily:  Ubuntu,
						Template:  "myTemplate",
						Users: []UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
						Datastore:         "myDatastore",
						Folder:            "myFolder",
						ResourcePool:      "myResourcePool",
						StoragePolicyName: "myStoragePolicyName",
					},
				},
				"eksa-unit-test-2": {
					TypeMeta: metav1.TypeMeta{
						Kind:       VSphereMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test-2",
					},
					Spec: VSphereMachineConfigSpec{
						DiskGiB:   20,
						MemoryMiB: 2048,
						NumCPUs:   4,
						OSFamily:  Bottlerocket,
						Template:  "myTemplate2",
						Users: []UserConfiguration{{
							Name:              "mySshUsername2",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey2"},
						}},
						Datastore:         "myDatastore2",
						Folder:            "myFolder2",
						ResourcePool:      "myResourcePool2",
						StoragePolicyName: "myStoragePolicyName2",
					},
				},
			},
			wantErr: false,
		},
		{
			testName:                  "invalid kind",
			fileName:                  "testdata/cluster_invalid_kinds.yaml",
			wantVSphereMachineConfigs: nil,
			wantErr:                   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := GetVSphereMachineConfigs(tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetVSphereMachineConfigs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.wantVSphereMachineConfigs) {
				t.Fatalf("GetVSphereMachineConfigs() = %#v, want %#v", got, tt.wantVSphereMachineConfigs)
			}
		})
	}
}
