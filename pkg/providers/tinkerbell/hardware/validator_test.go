package hardware_test

import (
	"errors"
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
		Nameservers:  []string{"ns1"},
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
		Nameservers:  []string{"ns1"},
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
