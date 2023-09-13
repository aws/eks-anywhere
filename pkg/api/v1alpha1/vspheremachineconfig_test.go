package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
			wantErr: "users[0].name test is invalid. Please use 'ec2-user' for Bottlerocket",
		},
		{
			name: "invalid hostOSConfiguration",
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
					OSFamily:     "ubuntu",
					Users: []UserConfiguration{
						{
							Name: "test",
							SshAuthorizedKeys: []string{
								"ssh_rsa",
							},
						},
					},
					HostOSConfiguration: &HostOSConfiguration{
						NTPConfiguration: &NTPConfiguration{},
					},
				},
			},
			wantErr: "HostOSConfiguration is invalid for VSphereMachineConfig test: NTPConfiguration.Servers can not be empty",
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

func TestVSphereMachineConfigValidateUsers(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name          string
		machineConfig *VSphereMachineConfig
		wantErr       string
	}{
		{
			name: "machineconfig with bottlerocket user valid",
			machineConfig: &VSphereMachineConfig{
				Spec: VSphereMachineConfigSpec{
					OSFamily: "bottlerocket",
					Users: []UserConfiguration{{
						Name:              "ec2-user",
						SshAuthorizedKeys: []string{"ssh-rsa AAAA..."},
					}},
				},
			},
		},
		{
			name: "machineconfig with bottlerocket user valid",
			machineConfig: &VSphereMachineConfig{
				Spec: VSphereMachineConfigSpec{
					OSFamily: "ubuntu",
					Users: []UserConfiguration{{
						Name:              "capv",
						SshAuthorizedKeys: []string{"ssh-rsa AAAA..."},
					}},
				},
			},
		},
		{
			name: "machineconfig users not set",
			machineConfig: &VSphereMachineConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-cp",
				},
				Spec: VSphereMachineConfigSpec{},
			},
			wantErr: "users is not set for VSphereMachineConfig test-cp, please provide a user",
		},
		{
			name: "machineconfig with bottlerocket user name invalid",
			machineConfig: &VSphereMachineConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-cp",
				},
				Spec: VSphereMachineConfigSpec{
					OSFamily: "bottlerocket",
					Users: []UserConfiguration{
						{
							Name:              "capv",
							SshAuthorizedKeys: []string{"ssh-rsa AAAA..."},
						},
					},
				},
			},
			wantErr: "users[0].name capv is invalid. Please use 'ec2-user' for Bottlerocket",
		},
		{
			name: "machineconfig user name empty",
			machineConfig: &VSphereMachineConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-cp",
				},
				Spec: VSphereMachineConfigSpec{
					Users: []UserConfiguration{
						{
							Name:              "",
							SshAuthorizedKeys: []string{"ssh-rsa AAAA..."},
						},
					},
				},
			},
			wantErr: "users[0].name is not set or is empty for VSphereMachineConfig test-cp, please provide a username",
		},
		{
			name: "user ssh authorized key empty or not set",
			machineConfig: &VSphereMachineConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-cp",
				},
				Spec: VSphereMachineConfigSpec{
					Users: []UserConfiguration{{
						Name:              "Jeff",
						SshAuthorizedKeys: []string{""},
					}},
				},
			},
			wantErr: "users[0].SshAuthorizedKeys is not set or is empty for VSphereMachineConfig test-cp, please provide a valid ssh authorized key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.machineConfig.ValidateUsers()
			if tt.wantErr == "" {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}

func TestVSphereMachineConfigSetDefaultUsers(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name                  string
		machineConfig         *VSphereMachineConfig
		expectedMachineConfig *VSphereMachineConfig
	}{
		{
			name: "machine config with bottlerocket",
			machineConfig: &VSphereMachineConfig{
				Spec: VSphereMachineConfigSpec{
					OSFamily: "bottlerocket",
				},
			},
			expectedMachineConfig: &VSphereMachineConfig{
				Spec: VSphereMachineConfigSpec{
					OSFamily: "bottlerocket",
					Users: []UserConfiguration{
						{
							Name:              "ec2-user",
							SshAuthorizedKeys: []string{""},
						},
					},
				},
			},
		},
		{
			name: "machine config with ubuntu",
			machineConfig: &VSphereMachineConfig{
				Spec: VSphereMachineConfigSpec{
					OSFamily: "ubuntu",
				},
			},
			expectedMachineConfig: &VSphereMachineConfig{
				Spec: VSphereMachineConfigSpec{
					OSFamily: "ubuntu",
					Users: []UserConfiguration{
						{
							Name:              "capv",
							SshAuthorizedKeys: []string{""},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.machineConfig.SetUserDefaults()
			g.Expect(tt.machineConfig).To(Equal(tt.expectedMachineConfig))
		})
	}
}
