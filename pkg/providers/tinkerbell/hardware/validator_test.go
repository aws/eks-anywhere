package hardware_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
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
		"Ids": {
			Assertion: hardware.UniqueIds(),
			Machines: []hardware.Machine{
				{Id: "foo"},
				{Id: "bar"},
			},
		},
		"IpAddresses": {
			Assertion: hardware.UniqueIpAddress(),
			Machines: []hardware.Machine{
				{IpAddress: "foo"},
				{IpAddress: "bar"},
			},
		},
		"MacAddresses": {
			Assertion: hardware.UniqueMacAddress(),
			Machines: []hardware.Machine{
				{MacAddress: "foo"},
				{MacAddress: "bar"},
			},
		},
		"Hostnames": {
			Assertion: hardware.UniqueHostnames(),
			Machines: []hardware.Machine{
				{Hostname: "foo"},
				{Hostname: "bar"},
			},
		},
		"BmcIpAddresses": {
			Assertion: hardware.UniqueBmcIpAddress(),
			Machines: []hardware.Machine{
				{BmcIpAddress: "foo"},
				{BmcIpAddress: "bar"},
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
		"Ids": {
			Assertion: hardware.UniqueIds(),
			Machines: []hardware.Machine{
				{Id: "foo"},
				{Id: "foo"},
			},
		},
		"IpAddresses": {
			Assertion: hardware.UniqueIpAddress(),
			Machines: []hardware.Machine{
				{IpAddress: "foo"},
				{IpAddress: "foo"},
			},
		},
		"MacAddresses": {
			Assertion: hardware.UniqueMacAddress(),
			Machines: []hardware.Machine{
				{MacAddress: "foo"},
				{MacAddress: "foo"},
			},
		},
		"Hostnames": {
			Assertion: hardware.UniqueHostnames(),
			Machines: []hardware.Machine{
				{Hostname: "foo"},
				{Hostname: "foo"},
			},
		},
		"BmcIpAddresses": {
			Assertion: hardware.UniqueBmcIpAddress(),
			Machines: []hardware.Machine{
				{BmcIpAddress: "foo"},
				{BmcIpAddress: "foo"},
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

func NewValidMachine() hardware.Machine {
	return hardware.Machine{
		Id:           uuid.NewString(),
		IpAddress:    "10.10.10.10",
		Gateway:      "10.10.10.1",
		Nameservers:  []string{"ns1"},
		MacAddress:   "00:00:00:00:00:00",
		Netmask:      "255.255.255.255",
		Hostname:     "localhost",
		BmcIpAddress: "10.10.10.11",
		BmcUsername:  "username",
		BmcPassword:  "password",
		BmcVendor:    "dell",
	}
}
