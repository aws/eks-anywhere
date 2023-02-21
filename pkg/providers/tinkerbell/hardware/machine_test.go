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

func TestNewMachineFromHardware(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name        string
		wantMachine hardware.Machine
	}{
		{
			name: "new machine with bmc",
			wantMachine: hardware.Machine{
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
			wantMachine: hardware.Machine{
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
			hw := tinkHardware(&tt.wantMachine)
			rm := rufioMachine(&tt.wantMachine)
			secret := bmcAuthSecret(&tt.wantMachine)

			newMachine := hardware.NewMachineFromHardware(*hw, rm, secret)
			g.Expect(newMachine).To(Equal(tt.wantMachine))
		})
	}
}

func tinkHardware(machine *hardware.Machine) *tinkv1alpha1.Hardware {
	allow := true

	return &tinkv1alpha1.Hardware{
		ObjectMeta: metav1.ObjectMeta{
			Name:      machine.Hostname,
			Namespace: constants.EksaSystemNamespace,
			Labels: map[string]string{
				"type":                                   "cp",
				"v1alpha1.tinkerbell.org/ownerName":      "mgmt-control-plane-template",
				"v1alpha1.tinkerbell.org/ownerNamespace": "eksa-system",
			},
		},
		Spec: tinkv1alpha1.HardwareSpec{
			Disks: []tinkv1alpha1.Disk{{Device: machine.Disk}},
			Interfaces: []tinkv1alpha1.Interface{
				{
					Netboot: &tinkv1alpha1.Netboot{
						AllowPXE:      &allow,
						AllowWorkflow: &allow,
					},
					DHCP: &tinkv1alpha1.DHCP{
						MAC: machine.MACAddress,
						IP: &tinkv1alpha1.IP{
							Address: machine.IPAddress,
							Netmask: machine.Netmask,
							Gateway: machine.Gateway,
							Family:  4,
						},
						Hostname:    machine.Hostname,
						NameServers: machine.Nameservers,
						VLANID:      machine.VLANID,
					},
				},
			},
		},
	}
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
