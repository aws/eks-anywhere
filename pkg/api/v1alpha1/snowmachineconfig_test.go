package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

func TestSnowMachineConfigSetDefaults(t *testing.T) {
	tests := []struct {
		name   string
		before *SnowMachineConfig
		after  *SnowMachineConfig
	}{
		{
			name: "optional fields all empty",
			before: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{},
			},
			after: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					InstanceType:             DefaultSnowInstanceType,
					PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
				},
			},
		},
		{
			name: "instance type exists",
			before: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					InstanceType: "instance-type-1",
				},
			},
			after: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					InstanceType:             "instance-type-1",
					PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
				},
			},
		},
		{
			name: "ssh key name exists",
			before: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					SshKeyName: "ssh-name",
				},
			},
			after: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					SshKeyName:               "ssh-name",
					InstanceType:             DefaultSnowInstanceType,
					PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
				},
			},
		},
		{
			name: "physical network exists",
			before: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					PhysicalNetworkConnector: "network-1",
				},
			},
			after: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					PhysicalNetworkConnector: "network-1",
					InstanceType:             DefaultSnowInstanceType,
				},
			},
		},
		{
			name: "os family exists",
			before: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					OSFamily: "ubuntu",
				},
			},
			after: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					InstanceType:             DefaultSnowInstanceType,
					PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
					OSFamily:                 Ubuntu,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			tt.before.SetDefaults()
			g.Expect(tt.before).To(Equal(tt.after))
		})
	}
}

func TestSnowMachineConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		obj     *SnowMachineConfig
		wantErr string
	}{
		{
			name: "valid config with amiID, instance type, physical network interface, devices, network, container volume, osFamily",
			obj: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					AMIID:                    "ami-1",
					InstanceType:             DefaultSnowInstanceType,
					PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
					Devices:                  []string{"1.2.3.4"},
					OSFamily:                 Bottlerocket,
					Network: SnowNetwork{
						DirectNetworkInterfaces: []SnowDirectNetworkInterface{
							{
								Index:   1,
								DHCP:    true,
								Primary: true,
							},
						},
					},
					ContainersVolume: &snowv1.Volume{
						Size: 25,
					},
				},
			},
			wantErr: "",
		},
		{
			name: "valid without ami",
			obj: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					InstanceType:             "sbe-c.4xlarge",
					PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
					Devices:                  []string{"1.2.3.4"},
					OSFamily:                 Ubuntu,
					Network: SnowNetwork{
						DirectNetworkInterfaces: []SnowDirectNetworkInterface{
							{
								Index:   1,
								DHCP:    true,
								Primary: true,
							},
						},
					},
					ContainersVolume: &snowv1.Volume{
						Size: 25,
					},
				},
			},
			wantErr: "",
		},
		{
			name: "invalid instance type sbe-g.largex",
			obj: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					AMIID:                    "ami-1",
					InstanceType:             "sbe-g.largex",
					PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
					Devices:                  []string{"1.2.3.4"},
					OSFamily:                 Bottlerocket,
					Network: SnowNetwork{
						DirectNetworkInterfaces: []SnowDirectNetworkInterface{
							{
								Index:   1,
								DHCP:    true,
								Primary: true,
							},
						},
					},
					ContainersVolume: &snowv1.Volume{
						Size: 25,
					},
				},
			},
			wantErr: "SnowMachineConfig InstanceType sbe-g.largex is not supported",
		},
		{
			name: "invalid instance type sbe-c-xlarge",
			obj: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					AMIID:                    "ami-1",
					InstanceType:             "sbe-c-large",
					PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
					Devices:                  []string{"1.2.3.4"},
					OSFamily:                 Bottlerocket,
					Network: SnowNetwork{
						DirectNetworkInterfaces: []SnowDirectNetworkInterface{
							{
								Index:   1,
								DHCP:    true,
								Primary: true,
							},
						},
					},
					ContainersVolume: &snowv1.Volume{
						Size: 25,
					},
				},
			},
			wantErr: "SnowMachineConfig InstanceType sbe-c-large is not supported",
		},
		{
			name: "invalid instance type sbe-c.elarge",
			obj: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					AMIID:                    "ami-1",
					InstanceType:             "sbe-c.elarge",
					PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
					Devices:                  []string{"1.2.3.4"},
					OSFamily:                 Bottlerocket,
					Network: SnowNetwork{
						DirectNetworkInterfaces: []SnowDirectNetworkInterface{
							{
								Index:   1,
								DHCP:    true,
								Primary: true,
							},
						},
					},
					ContainersVolume: &snowv1.Volume{
						Size: 25,
					},
				},
			},
			wantErr: "SnowMachineConfig InstanceType sbe-c.elarge is not supported",
		},
		{
			name: "invalid physical network connector",
			obj: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					AMIID:                    "ami-1",
					InstanceType:             "sbe-c.64xlarge",
					PhysicalNetworkConnector: "invalid-physical-network",
					Devices:                  []string{"1.2.3.4"},
					OSFamily:                 Bottlerocket,
					Network: SnowNetwork{
						DirectNetworkInterfaces: []SnowDirectNetworkInterface{
							{
								Index:   1,
								DHCP:    true,
								Primary: true,
							},
						},
					},
					ContainersVolume: &snowv1.Volume{
						Size: 25,
					},
				},
			},
			wantErr: "PhysicalNetworkConnector invalid-physical-network is not supported",
		},
		{
			name: "empty devices",
			obj: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					AMIID:                    "ami-1",
					InstanceType:             "sbe-g.large",
					PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
					OSFamily:                 Bottlerocket,
					Network: SnowNetwork{
						DirectNetworkInterfaces: []SnowDirectNetworkInterface{
							{
								Index:   1,
								DHCP:    true,
								Primary: true,
							},
						},
					},
					ContainersVolume: &snowv1.Volume{
						Size: 25,
					},
				},
			},
			wantErr: "Devices must contain at least one device IP",
		},
		{
			name: "invalid container volume size for ubuntu",
			obj: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					AMIID:                    "ami-1",
					InstanceType:             DefaultSnowInstanceType,
					PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
					Devices:                  []string{"1.2.3.4"},
					OSFamily:                 Ubuntu,
					Network: SnowNetwork{
						DirectNetworkInterfaces: []SnowDirectNetworkInterface{
							{
								Index:   1,
								DHCP:    true,
								Primary: true,
							},
						},
					},
					ContainersVolume: &snowv1.Volume{
						Size: 7,
					},
				},
			},
			wantErr: "SnowMachineConfig ContainersVolume.Size must be no smaller than 8 Gi",
		},
		{
			name: "invalid container volume size for bottlerocket",
			obj: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					AMIID:                    "ami-1",
					InstanceType:             DefaultSnowInstanceType,
					PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
					Devices:                  []string{"1.2.3.4"},
					OSFamily:                 Bottlerocket,
					Network: SnowNetwork{
						DirectNetworkInterfaces: []SnowDirectNetworkInterface{
							{
								Index:   1,
								DHCP:    true,
								Primary: true,
							},
						},
					},
					ContainersVolume: &snowv1.Volume{
						Size: 24,
					},
				},
			},
			wantErr: "SnowMachineConfig ContainersVolume.Size must be no smaller than 25 Gi",
		},
		{
			name: "container volume not specified for bottlerocket",
			obj: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					AMIID:                    "ami-1",
					InstanceType:             DefaultSnowInstanceType,
					PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
					Devices:                  []string{"1.2.3.4"},
					OSFamily:                 Bottlerocket,
					Network: SnowNetwork{
						DirectNetworkInterfaces: []SnowDirectNetworkInterface{
							{
								Index:   1,
								DHCP:    true,
								Primary: true,
							},
						},
					},
				},
			},
			wantErr: "SnowMachineConfig ContainersVolume must be specified for Bottlerocket OS",
		},
		{
			name: "invalid os family",
			obj: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					AMIID:                    "ami-1",
					InstanceType:             DefaultSnowInstanceType,
					PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
					Devices:                  []string{"1.2.3.4"},
					OSFamily:                 "invalidOS",
					Network: SnowNetwork{
						DirectNetworkInterfaces: []SnowDirectNetworkInterface{
							{
								Index:   1,
								DHCP:    true,
								Primary: true,
							},
						},
					},
					ContainersVolume: &snowv1.Volume{
						Size: 25,
					},
				},
			},
			wantErr: "SnowMachineConfig OSFamily invalidOS is not supported",
		},
		{
			name: "empty os family",
			obj: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					AMIID:                    "ami-1",
					InstanceType:             DefaultSnowInstanceType,
					PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
					Devices:                  []string{"1.2.3.4"},
					OSFamily:                 "",
					Network: SnowNetwork{
						DirectNetworkInterfaces: []SnowDirectNetworkInterface{
							{
								Index:   1,
								DHCP:    true,
								Primary: true,
							},
						},
					},
					ContainersVolume: &snowv1.Volume{
						Size: 25,
					},
				},
			},
			wantErr: "SnowMachineConfig OSFamily must be specified",
		},
		{
			name: "empty network",
			obj: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					AMIID:                    "ami-1",
					InstanceType:             DefaultSnowInstanceType,
					PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
					Devices:                  []string{"1.2.3.4"},
					OSFamily:                 Bottlerocket,
					ContainersVolume: &snowv1.Volume{
						Size: 25,
					},
				},
			},
			wantErr: "SnowMachineConfig Network.DirectNetworkInterfaces length must be no smaller than 1",
		},
		{
			name: "invalid network",
			obj: &SnowMachineConfig{
				Spec: SnowMachineConfigSpec{
					AMIID:                    "ami-1",
					InstanceType:             DefaultSnowInstanceType,
					PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
					Devices:                  []string{"1.2.3.4"},
					OSFamily:                 Bottlerocket,
					ContainersVolume: &snowv1.Volume{
						Size: 25,
					},
					Network: SnowNetwork{
						DirectNetworkInterfaces: []SnowDirectNetworkInterface{
							{
								Index:   1,
								DHCP:    true,
								Primary: false,
							},
						},
					},
				},
			},
			wantErr: "SnowMachineConfig Network.DirectNetworkInterfaces list must contain one and only one primary DNI",
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

func TestSnowMachineConfigSetControlPlaneAnnotation(t *testing.T) {
	g := NewWithT(t)
	m := &SnowMachineConfig{}
	m.SetControlPlaneAnnotation()
	g.Expect(m.Annotations).To(Equal(map[string]string{"anywhere.eks.amazonaws.com/control-plane": "true"}))
}

func TestSnowMachineConfigSetEtcdAnnotation(t *testing.T) {
	g := NewWithT(t)
	m := &SnowMachineConfig{}
	m.SetEtcdAnnotation()
	g.Expect(m.Annotations).To(Equal(map[string]string{"anywhere.eks.amazonaws.com/etcd": "true"}))
}

func TestNewSnowMachineConfigGenerate(t *testing.T) {
	g := NewWithT(t)
	want := &SnowMachineConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       SnowMachineConfigKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: "snow-cluster",
		},
		Spec: SnowMachineConfigSpec{
			AMIID:                    "",
			Devices:                  []string{""},
			InstanceType:             DefaultSnowInstanceType,
			SshKeyName:               DefaultSnowSSHKeyName,
			PhysicalNetworkConnector: DefaultSnowPhysicalNetworkConnectorType,
			OSFamily:                 DefaultOSFamily,
			Network: SnowNetwork{
				DirectNetworkInterfaces: []SnowDirectNetworkInterface{
					{
						Index:   1,
						DHCP:    true,
						Primary: true,
					},
				},
			},
		},
	}
	g.Expect(NewSnowMachineConfigGenerate("snow-cluster")).To(Equal(want))
}
