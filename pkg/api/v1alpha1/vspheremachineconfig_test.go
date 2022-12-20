package v1alpha1

import (
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
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

func TestVSphereMachineConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		obj     *VSphereMachineConfig
		wantErr string
	}{
		{
			name: "valid config",
			obj: &VSphereMachineConfig{
				Spec: VSphereMachineConfigSpec{
					MemoryMiB:    64,
					DiskGiB:      100,
					NumCPUs:      3,
					Template:     "templateA",
					ResourcePool: "poolA",
					Datastore:    "ds-aaa",
					Folder:       "folder/A",
					OSFamily:     "ubuntu",
					Users: []UserConfiguration{
						{
							Name: "test",
							SshAuthorizedKeys: []string{
								"ssh_rsa",
							},
						},
					},
				},
			},
			wantErr: "",
		},
		{
			name: "valid without folder",
			obj: &VSphereMachineConfig{
				Spec: VSphereMachineConfigSpec{
					MemoryMiB:    64,
					DiskGiB:      100,
					NumCPUs:      3,
					Template:     "templateA",
					ResourcePool: "poolA",
					Datastore:    "ds-aaa",
					OSFamily:     "ubuntu",
					Users: []UserConfiguration{
						{
							Name: "test",
							SshAuthorizedKeys: []string{
								"ssh_rsa",
							},
						},
					},
				},
			},
			wantErr: "",
		},
		{
			name: "invalid - datastore not set",
			obj: &VSphereMachineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec: VSphereMachineConfigSpec{
					MemoryMiB:    64,
					DiskGiB:      100,
					NumCPUs:      3,
					Template:     "templateA",
					ResourcePool: "poolA",
					Folder:       "folder/A",
					OSFamily:     "ubuntu",
					Users: []UserConfiguration{
						{
							Name: "test",
							SshAuthorizedKeys: []string{
								"ssh_rsa",
							},
						},
					},
				},
			},
			wantErr: "VSphereMachineConfig test datastore is not set or is empty",
		},
		{
			name: "invalid - resource pool not set",
			obj: &VSphereMachineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec: VSphereMachineConfigSpec{
					MemoryMiB: 64,
					DiskGiB:   100,
					NumCPUs:   3,
					Template:  "templateA",
					Datastore: "ds-aaa",
					Folder:    "folder/A",
					OSFamily:  "ubuntu",
					Users: []UserConfiguration{
						{
							Name: "test",
							SshAuthorizedKeys: []string{
								"ssh_rsa",
							},
						},
					},
				},
			},
			wantErr: "VSphereMachineConfig test VM resourcePool is not set or is empty",
		},
		{
			name: "unsupported os family",
			obj: &VSphereMachineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec: VSphereMachineConfigSpec{
					MemoryMiB:    64,
					DiskGiB:      100,
					NumCPUs:      3,
					Template:     "templateA",
					ResourcePool: "poolA",
					Datastore:    "ds-aaa",
					Folder:       "folder/A",
					OSFamily:     "suse",
					Users: []UserConfiguration{
						{
							Name: "test",
							SshAuthorizedKeys: []string{
								"ssh_rsa",
							},
						},
					},
				},
			},
			wantErr: "VSphereMachineConfig test osFamily: suse is not supported, please use one of the following: bottlerocket, ubuntu",
		},
		{
			name: "invalid ssh username",
			obj: &VSphereMachineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec: VSphereMachineConfigSpec{
					MemoryMiB:    64,
					DiskGiB:      100,
					NumCPUs:      3,
					Template:     "templateA",
					ResourcePool: "poolA",
					Datastore:    "ds-aaa",
					Folder:       "folder/A",
					OSFamily:     "bottlerocket",
					Users: []UserConfiguration{
						{
							Name: "test",
							SshAuthorizedKeys: []string{
								"ssh_rsa",
							},
						},
					},
				},
			},
			wantErr: "SSHUsername test is invalid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			err := tt.obj.Validate()
			if tt.wantErr == "" {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}
