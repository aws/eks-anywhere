package hardware_test

import (
	"testing"

	. "github.com/onsi/gomega"
	rufioalphav1 "github.com/tinkerbell/rufio/api/v1alpha1"
	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestMachineFromHardwareSuccess(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name        string
		hw          *tinkv1alpha1.Hardware
		wantMachine *hardware.Machine
	}{
		{
			name: "new machine with bmc",
			hw:   tinkHardware(),
			wantMachine: &hardware.Machine{
				Hostname:    "hw1",
				IPAddress:   "10.10.10.10",
				Netmask:     "10.10.10.1",
				Gateway:     "10.10.10.1",
				Nameservers: []string{"1.1.1.1"},
				MACAddress:  "00:00:00:00:00:01",
				Disk:        "/dev/sda",
				Labels: map[string]string{
					"type": "cp",
				},
				BMCIPAddress: "192.168.0.10",
				BMCUsername:  "Admin",
				BMCPassword:  "admin",
				VLANID:       "",
			},
		},
		{
			name: "new machine without bmc",
			hw:   tinkHardware(),
			wantMachine: &hardware.Machine{
				Hostname:    "hw1",
				IPAddress:   "10.10.10.10",
				Netmask:     "10.10.10.1",
				Gateway:     "10.10.10.1",
				Nameservers: []string{"1.1.1.1"},
				MACAddress:  "00:00:00:00:00:01",
				Disk:        "/dev/sda",
				Labels: map[string]string{
					"type": "cp",
				},
				BMCIPAddress: "",
				BMCUsername:  "",
				BMCPassword:  "",
				VLANID:       "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := rufioMachine(tt.wantMachine)
			secret := bmcAuthSecret(tt.wantMachine)

			newMachine, err := hardware.MachineFromHardware(*tt.hw, rm, secret)

			g.Expect(err).To(BeNil())
			g.Expect(newMachine).To(Equal(tt.wantMachine))

		})
	}
}

func TestMachineFromHardwareFailure(t *testing.T) {
	g := NewWithT(t)

	machine := &hardware.Machine{
		Hostname:    "hw1",
		IPAddress:   "10.10.10.10",
		Netmask:     "10.10.10.1",
		Gateway:     "10.10.10.1",
		Nameservers: []string{"1.1.1.1"},
		MACAddress:  "00:00:00:00:00:01",
		Disk:        "/dev/sda",
		Labels: map[string]string{
			"type": "cp",
		},
		BMCIPAddress: "192.168.0.10",
		BMCUsername:  "Admin",
		BMCPassword:  "admin",
		VLANID:       "",
	}

	tests := []struct {
		name        string
		wantError   string
		hw          *tinkv1alpha1.Hardware
		wantMachine *hardware.Machine
	}{
		{
			name: "hardware no interface",
			hw: tinkHardware(func(hw *tinkv1alpha1.Hardware) {
				hw.Spec.Interfaces = []tinkv1alpha1.Interface{}
			}),
			wantError:   "interfaces is empty",
			wantMachine: machine,
		},
		{
			name: "hardware no DHCP",
			hw: tinkHardware(func(hw *tinkv1alpha1.Hardware) {
				hw.Spec.Interfaces[0].DHCP = nil
			}),
			wantError:   "no DHCP on interface",
			wantMachine: machine,
		},
		{
			name: "hardware no DHCP IP",
			hw: tinkHardware(func(hw *tinkv1alpha1.Hardware) {
				hw.Spec.Interfaces[0].DHCP.IP = nil
			}),
			wantError:   "no IP on DHCP",
			wantMachine: machine,
		},
		{
			name: "hardware no disks",
			hw: tinkHardware(func(hw *tinkv1alpha1.Hardware) {
				hw.Spec.Disks = []tinkv1alpha1.Disk{}
			}),
			wantError:   "disks is empty",
			wantMachine: machine,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := rufioMachine(tt.wantMachine)
			secret := bmcAuthSecret(tt.wantMachine)

			newMachine, err := hardware.MachineFromHardware(*tt.hw, rm, secret)

			if tt.wantError != "" {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantError)))
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(newMachine).To(Equal(tt.wantMachine))
			}

		})
	}
}

type tinkHardwareOpt func(hw *tinkv1alpha1.Hardware)

func tinkHardware(opts ...tinkHardwareOpt) *tinkv1alpha1.Hardware {
	allow := true
	name := "hw1"
	hw := &tinkv1alpha1.Hardware{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: constants.EksaSystemNamespace,
			Labels: map[string]string{
				"type":                                   "cp",
				"v1alpha1.tinkerbell.org/ownerName":      "mgmt-control-plane-template",
				"v1alpha1.tinkerbell.org/ownerNamespace": "eksa-system",
			},
		},
		Spec: tinkv1alpha1.HardwareSpec{
			Disks: []tinkv1alpha1.Disk{{Device: "/dev/sda"}},
			Interfaces: []tinkv1alpha1.Interface{
				{
					Netboot: &tinkv1alpha1.Netboot{
						AllowPXE:      &allow,
						AllowWorkflow: &allow,
					},
					DHCP: &tinkv1alpha1.DHCP{
						MAC: "00:00:00:00:00:01",
						IP: &tinkv1alpha1.IP{
							Address: "10.10.10.10",
							Netmask: "10.10.10.1",
							Gateway: "10.10.10.1",
							Family:  4,
						},
						Hostname:    name,
						NameServers: []string{"1.1.1.1"},
						VLANID:      "",
					},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(hw)
	}
	return hw
}

func rufioMachine(machine *hardware.Machine) *rufioalphav1.Machine {
	return &rufioalphav1.Machine{
		Spec: rufioalphav1.MachineSpec{
			Connection: rufioalphav1.Connection{
				Host: machine.BMCIPAddress,
			},
		},
	}
}

func bmcAuthSecret(machine *hardware.Machine) *corev1.Secret {
	return &corev1.Secret{
		Data: map[string][]byte{
			"username": []byte(machine.BMCUsername),
			"password": []byte(machine.BMCPassword),
		},
	}
}
