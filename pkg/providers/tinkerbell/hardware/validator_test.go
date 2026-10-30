package hardware_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestDefaultMachineValidatorValidationsRun(t *testing.T) {
	g := gomega.NewWithT(t)

	// check is set by assertion when its called and allows us to validate
	// registered assertions are infact called by the validation decorator.
	var check bool
	assertion := func(m hardware.Machine) error {
		check = true
		return nil
	}

	validator := &hardware.DefaultMachineValidator{}
	validator.Register(assertion)

	err := validator.Validate(NewValidMachine())

	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(check).To(gomega.BeTrue())
}

func TestDefaultMachineValidatorErrorsWhenAssertionErrors(t *testing.T) {
	g := gomega.NewWithT(t)

	// check is set by assertion when its called and allows us to validate
	// registered assertions are infact called by the validation decorator.
	expect := errors.New("something went wrong")
	assertion := func(hardware.Machine) error {
		return expect
	}

	validator := &hardware.DefaultMachineValidator{}
	validator.Register(assertion)

	err := validator.Validate(NewValidMachine())

	g.Expect(err).To(gomega.BeEquivalentTo(expect))
}

func TestUniquenessAssertions(t *testing.T) {
	cases := map[string]struct {
		Assertion hardware.MachineAssertion
		Machines  []hardware.Machine
	}{
		"IPAddresses": {
			Assertion: hardware.UniqueIPAddress(),
			Machines: []hardware.Machine{
				{IPAddress: "foo"},
				{IPAddress: "bar"},
			},
		},
		"MACAddresses": {
			Assertion: hardware.UniqueMACAddress(),
			Machines: []hardware.Machine{
				{MACAddress: "foo"},
				{MACAddress: "bar"},
			},
		},
		"Hostnames": {
			Assertion: hardware.UniqueHostnames(),
			Machines: []hardware.Machine{
				{Hostname: "foo"},
				{Hostname: "bar"},
			},
		},
		"BMCIPAddresses": {
			Assertion: hardware.UniqueBMCIPAddress(),
			Machines: []hardware.Machine{
				{BMCIPAddress: "foo"},
				{BMCIPAddress: "bar"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			g.Expect(tc.Assertion(tc.Machines[0])).ToNot(gomega.HaveOccurred())
			g.Expect(tc.Assertion(tc.Machines[1])).ToNot(gomega.HaveOccurred())
		})
	}
}

func TestUniquenessAssertionsWithDupes(t *testing.T) {
	cases := map[string]struct {
		Assertion hardware.MachineAssertion
		Machines  []hardware.Machine
	}{
		"IPAddresses": {
			Assertion: hardware.UniqueIPAddress(),
			Machines: []hardware.Machine{
				{IPAddress: "foo"},
				{IPAddress: "foo"},
			},
		},
		"MACAddresses": {
			Assertion: hardware.UniqueMACAddress(),
			Machines: []hardware.Machine{
				{MACAddress: "foo"},
				{MACAddress: "foo"},
			},
		},
		"Hostnames": {
			Assertion: hardware.UniqueHostnames(),
			Machines: []hardware.Machine{
				{Hostname: "foo"},
				{Hostname: "foo"},
			},
		},
		"BMCIPAddresses": {
			Assertion: hardware.UniqueBMCIPAddress(),
			Machines: []hardware.Machine{
				{BMCIPAddress: "foo"},
				{BMCIPAddress: "foo"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			g.Expect(tc.Assertion(tc.Machines[0])).ToNot(gomega.HaveOccurred())
			g.Expect(tc.Assertion(tc.Machines[1])).To(gomega.HaveOccurred())
		})
	}
}

func TestStaticMachineAssertions_ValidMachine(t *testing.T) {
	g := gomega.NewWithT(t)

	machine := NewValidMachine()

	validate := hardware.StaticMachineAssertions()
	g.Expect(validate(machine)).ToNot(gomega.HaveOccurred())
}

func TestMachineValidatorAllowsAddresslessHardware(t *testing.T) {
	g := gomega.NewWithT(t)
	machine := newSLAACMachine()

	g.Expect(hardware.NewMachineValidator(true).Validate(machine)).To(gomega.Succeed())
}

func TestMachineValidatorRejectsSLAACStaticAddresses(t *testing.T) {
	for _, tt := range []struct {
		name       string
		ip, gw, nm string
		field      string
	}{
		{name: "IP only", ip: "2001:db8::100", field: "ip_address"},
		{name: "gateway only", gw: "2001:db8::1", field: "gateway"},
		{name: "netmask only", nm: "64", field: "netmask"},
		{name: "IP and gateway", ip: "2001:db8::100", gw: "2001:db8::1", field: "ip_address"},
		{name: "IP and netmask", ip: "2001:db8::100", nm: "64", field: "ip_address"},
		{name: "gateway and netmask", gw: "2001:db8::1", nm: "64", field: "gateway"},
		{name: "full IPv6 triplet", ip: "2001:db8::100", gw: "2001:db8::1", nm: "64", field: "ip_address"},
		{name: "full IPv4 triplet", ip: "192.0.2.100", gw: "192.0.2.1", nm: "255.255.255.0", field: "ip_address"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			machine := newSLAACMachine()
			machine.IPAddress, machine.Gateway, machine.Netmask = tt.ip, tt.gw, tt.nm

			err := hardware.NewMachineValidator(true).Validate(machine)
			g.Expect(err).To(gomega.MatchError(fmt.Sprintf("machine: %s must be empty for IPv6 SLAAC (hostname=%q)", tt.field, machine.Hostname)))
		})
	}
}

func TestMachineValidatorSLAACRetainsOtherAssertions(t *testing.T) {
	for name, mutate := range map[string]func(*hardware.Machine){
		"missing DNS":      func(m *hardware.Machine) { m.Nameservers = nil },
		"empty DNS":        func(m *hardware.Machine) { m.Nameservers = []string{""} },
		"invalid DNS":      func(m *hardware.Machine) { m.Nameservers = []string{"invalid"} },
		"missing MAC":      func(m *hardware.Machine) { m.MACAddress = "" },
		"invalid MAC":      func(m *hardware.Machine) { m.MACAddress = "invalid" },
		"missing hostname": func(m *hardware.Machine) { m.Hostname = "" },
		"invalid hostname": func(m *hardware.Machine) { m.Hostname = "invalid!" },
		"invalid disk":     func(m *hardware.Machine) { m.Disk = "/invalid" },
		"invalid label":    func(m *hardware.Machine) { m.Labels = hardware.Labels{"invalid!": "cp"} },
		"invalid value":    func(m *hardware.Machine) { m.Labels = hardware.Labels{"type": "invalid!"} },
		"missing BMC IP":   func(m *hardware.Machine) { m.BMCIPAddress = "" },
		"invalid BMC IP":   func(m *hardware.Machine) { m.BMCIPAddress = "invalid" },
		"missing BMC user": func(m *hardware.Machine) { m.BMCUsername = "" },
		"missing BMC pass": func(m *hardware.Machine) { m.BMCPassword = "" },
		"invalid VLAN":     func(m *hardware.Machine) { m.VLANID = "4095" },
	} {
		t.Run(name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			machine := newSLAACMachine()
			mutate(&machine)
			g.Expect(hardware.NewMachineValidator(true).Validate(machine)).NotTo(gomega.Succeed())
		})
	}
}

func TestMachineValidatorSLAACUniqueness(t *testing.T) {
	for _, tt := range []struct {
		name    string
		mutate  func(*hardware.Machine)
		wantErr string
	}{
		{name: "distinct machines"},
		{
			name:    "duplicate MAC",
			mutate:  func(m *hardware.Machine) { m.MACAddress = "02:00:00:00:00:01" },
			wantErr: "duplicate MACAddress",
		},
		{
			name:    "duplicate hostname",
			mutate:  func(m *hardware.Machine) { m.Hostname = "node-example" },
			wantErr: "duplicate Hostname",
		},
		{
			name:    "duplicate BMC",
			mutate:  func(m *hardware.Machine) { m.BMCIPAddress = "192.0.2.20" },
			wantErr: "192.0.2.20",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			first, second := newSLAACMachine(), newSLAACMachine()
			second.Hostname, second.MACAddress, second.BMCIPAddress = "node-second", "02:00:00:00:00:02", "192.0.2.21"
			if tt.mutate != nil {
				tt.mutate(&second)
			}
			validator := hardware.NewMachineValidator(true)
			g.Expect(validator.Validate(first)).To(gomega.Succeed())
			err := validator.Validate(second)
			if tt.wantErr == "" {
				g.Expect(err).To(gomega.Succeed())
			} else {
				g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(tt.wantErr)))
			}
		})
	}
}

func TestMachineValidatorStaticMode(t *testing.T) {
	g := gomega.NewWithT(t)
	machine := newSLAACMachine()
	g.Expect(hardware.NewMachineValidator(false).Validate(machine)).To(gomega.MatchError("machine: IPAddress is empty"))
	machine.IPAddress, machine.Gateway, machine.Netmask = "192.0.2.100", "192.0.2.1", "255.255.255.0"
	validator := hardware.NewMachineValidator(false)
	g.Expect(validator.Validate(machine)).To(gomega.Succeed())
	machine.Hostname, machine.MACAddress, machine.BMCIPAddress = "node-second", "02:00:00:00:00:02", "192.0.2.21"
	g.Expect(validator.Validate(machine)).To(gomega.MatchError("duplicate IPAddress: 192.0.2.100"))
}

func newSLAACMachine() hardware.Machine {
	return hardware.Machine{
		Hostname:     "node-example",
		MACAddress:   "02:00:00:00:00:01",
		Nameservers:  []string{"2001:db8::53"},
		Disk:         "/dev/sda",
		Labels:       hardware.Labels{"type": "cp"},
		BMCIPAddress: "192.0.2.20",
		BMCUsername:  "dummy-user",
		BMCPassword:  "dummy-password",
		VLANID:       "123",
	}
}

func TestStaticMachineAssertions_InvalidMachines(t *testing.T) {
	g := gomega.NewWithT(t)

	cases := map[string]func(*hardware.Machine){
		"EmptyIPAddress": func(h *hardware.Machine) {
			h.IPAddress = ""
		},
		"InvalidIPAddress": func(h *hardware.Machine) {
			h.IPAddress = "invalid"
		},
		"EmptyGateway": func(h *hardware.Machine) {
			h.Gateway = ""
		},
		"InvalidGateway": func(h *hardware.Machine) {
			h.Gateway = "invalid"
		},
		"NoNameservers": func(h *hardware.Machine) {
			h.Nameservers = []string{}
		},
		"EmptyNameserver": func(h *hardware.Machine) {
			h.Nameservers = []string{""}
		},
		"EmptyNetmask": func(h *hardware.Machine) {
			h.Netmask = ""
		},
		"EmptyMACAddress": func(h *hardware.Machine) {
			h.MACAddress = ""
		},
		"InvalidMACAddress": func(h *hardware.Machine) {
			h.MACAddress = "invalid mac"
		},
		"EmptyHostname": func(h *hardware.Machine) {
			h.Hostname = ""
		},
		"InvalidHostname": func(h *hardware.Machine) {
			h.Hostname = "!@#$%"
		},
		"EmptyBMCIPAddress": func(h *hardware.Machine) {
			h.BMCIPAddress = ""
		},
		"InvalidBMCIPAddress": func(h *hardware.Machine) {
			h.BMCIPAddress = "invalid"
		},
		"EmptyBMCUsername": func(h *hardware.Machine) {
			h.BMCUsername = ""
		},
		"EmptyBMCPassword": func(h *hardware.Machine) {
			h.BMCPassword = ""
		},
		"InvalidLabelKey": func(h *hardware.Machine) {
			h.Labels["?$?$?"] = "foo"
		},
		"InvalidLabelValue": func(h *hardware.Machine) {
			h.Labels["foo"] = "\\/dsa"
		},
		"InvalidDisk": func(h *hardware.Machine) {
			h.Disk = "*&!@#!%"
		},
		"InvalidWithJustDev": func(h *hardware.Machine) {
			h.Disk = "/dev/"
		},
		"InvalidVLANUnder": func(h *hardware.Machine) {
			h.VLANID = "0"
		},
		"InvalidVLANOver": func(h *hardware.Machine) {
			h.VLANID = "4095"
		},
		"NonIntVLAN": func(h *hardware.Machine) {
			h.VLANID = "im not an int"
		},
	}

	validate := hardware.StaticMachineAssertions()
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			machine := NewValidMachine()
			mutate(&machine)
			g.Expect(validate(machine)).To(gomega.HaveOccurred())
		})
	}
}

func NewValidMachine() hardware.Machine {
	return hardware.Machine{
		IPAddress:    "10.10.10.10",
		Gateway:      "10.10.10.1",
		Nameservers:  []string{"192.0.2.53"},
		MACAddress:   "00:00:00:00:00:00",
		Netmask:      "255.255.255.255",
		Hostname:     "localhost",
		Labels:       hardware.Labels{"type": "cp"},
		Disk:         "/dev/sda",
		BMCIPAddress: "10.10.10.11",
		BMCUsername:  "username",
		BMCPassword:  "password",
		VLANID:       "200",
	}
}

func NewMachineWithOptions() hardware.Machine {
	return hardware.Machine{
		IPAddress:    "10.10.10.10",
		Gateway:      "10.10.10.1",
		Nameservers:  []string{"192.0.2.53"},
		MACAddress:   "00:00:00:00:00:00",
		Netmask:      "255.255.255.255",
		Hostname:     "localhost",
		Labels:       hardware.Labels{"type": "cp"},
		Disk:         "/dev/sda",
		BMCIPAddress: "10.10.10.11",
		BMCUsername:  "username",
		BMCPassword:  "password",
		VLANID:       "200",
		BMCOptions: &hardware.BMCOptions{
			RPC: &hardware.RPCOpts{
				ConsumerURL: "https://example.net",
				Request: hardware.RequestOpts{
					HTTPContentType: "application/vnd.api+json",
					HTTPMethod:      http.MethodPost,
					StaticHeaders:   map[string][]string{"myheader": {"myvalue"}},
					TimestampFormat: time.RFC3339,
					TimestampHeader: "X-Example-Timestamp",
				},
				Signature: hardware.SignatureOpts{
					HeaderName:                 "X-Example-Signature",
					AppendAlgoToHeaderDisabled: true,
					IncludedPayloadHeaders:     []string{"X-Example-Timestamp"},
				},
				HMAC: hardware.HMACOpts{
					PrefixSigDisabled: true,
					Secrets:           []string{"superSecret1", "superSecret2"},
				},
				Experimental: hardware.ExperimentalOpts{
					CustomRequestPayload: `{"data":{"type":"articles","id":"1","attributes":{"title": "Rails is Omakase"},"relationships":{"author":{"links":{"self":"/articles/1/relationships/author","related":"/articles/1/author"},"data":null}}}}`,
					DotPath:              "data.relationships.author.data",
				},
			},
		},
	}
}
