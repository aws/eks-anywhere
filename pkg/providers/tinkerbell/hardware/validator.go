package hardware

import (
	"fmt"
)

// MachineAssertion defines a condition that Machine must meet.
type MachineAssertion func(Machine) error

// DefaultMachineValidator validated Machine instances.
type DefaultMachineValidator struct {
	assertions []MachineAssertion
}

var _ MachineValidator = &DefaultMachineValidator{}

// NewDefaultMachineValidator creates a machineValidator instance with default assertions registered.
func NewDefaultMachineValidator() *DefaultMachineValidator {
	validator := &DefaultMachineValidator{}
	WithDefaultAssertions(validator)
	return validator
}

// Validate validates machine by executing its Validate() method and passing it to all registered MachineAssertions.
func (mv *DefaultMachineValidator) Validate(machine Machine) error {
	if err := machine.Validate(); err != nil {
		return err
	}

	for _, fn := range mv.assertions {
		if err := fn(machine); err != nil {
			return err
		}
	}

	return nil
}

// Register registers v MachineAssertions with m.
func (mv *DefaultMachineValidator) Register(v ...MachineAssertion) {
	mv.assertions = append(mv.assertions, v...)
}

// UniqueIds asserts a given Machine instance has a unique Id field relative to previously seen Machine instances.
// It is not thread safe. It has a 1 time use.
func UniqueIds() MachineAssertion {
	ids := make(map[string]struct{})
	return func(m Machine) error {
		if _, seen := ids[m.Id]; seen {
			return fmt.Errorf("duplicate Id: %v", m.Id)
		}

		ids[m.Id] = struct{}{}

		return nil
	}
}

// UniqueIpAddress asserts a given Machine instance has a unique IpAddress field relative to previously seen Machine
// instances. It is not thread safe. It has a 1 time use.
func UniqueIpAddress() MachineAssertion {
	ips := make(map[string]struct{})
	return func(m Machine) error {
		if _, seen := ips[m.IpAddress]; seen {
			return fmt.Errorf("duplicate IpAddress: %v", m.IpAddress)
		}

		ips[m.IpAddress] = struct{}{}

		return nil
	}
}

// UniqueMacAddress asserts a given Machine instance has a unique MacAddress field relative to previously seen Machine
// instances. It is not thread safe. It has a 1 time use.
func UniqueMacAddress() MachineAssertion {
	macs := make(map[string]struct{})
	return func(m Machine) error {
		if _, seen := macs[m.MacAddress]; seen {
			return fmt.Errorf("duplicate MacAddress: %v", m.MacAddress)
		}

		macs[m.MacAddress] = struct{}{}

		return nil
	}
}

// UniqueHostnames asserts a given Machine instance has a unique Hostname field relative to previously seen Machine
// instances. It is not thread safe. It has a 1 time use.
func UniqueHostnames() MachineAssertion {
	hostnames := make(map[string]struct{})
	return func(m Machine) error {
		if _, seen := hostnames[m.Hostname]; seen {
			return fmt.Errorf("duplicate Hostname: %v", m.Hostname)
		}

		hostnames[m.Hostname] = struct{}{}

		return nil
	}
}

// UniqueBmcIpAddress asserts a given Machine instance has a unique BmcIpAddress field relative to previously seen
// Machine instances. If there is no Bmc configuration as defined by machine.HasBmc() the check is a noop. It is
// not thread safe. It has a 1 time use.
func UniqueBmcIpAddress() MachineAssertion {
	ips := make(map[string]struct{})
	return func(m Machine) error {
		if !m.HasBmc() {
			return nil
		}

		if m.BmcIpAddress == "" {
			return fmt.Errorf("missing BmcIpAddress (id=\"%v\")", m.Id)
		}

		if _, seen := ips[m.BmcIpAddress]; seen {
			return fmt.Errorf("duplicate IpAddress: %v", m.BmcIpAddress)
		}

		ips[m.BmcIpAddress] = struct{}{}

		return nil
	}
}

var defaultAssertions = []MachineAssertion{
	UniqueIds(),
	UniqueIpAddress(),
	UniqueMacAddress(),
	UniqueHostnames(),
	UniqueBmcIpAddress(),
}

// WithDefaultAssertions applies a set of default assertions to validator. The default assertions include
// UniqueHostnames and UniqueIds.
func WithDefaultAssertions(validator *DefaultMachineValidator) {
	validator.Register(defaultAssertions...)
}
