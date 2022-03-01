package hardware

import "fmt"

// MachineAssertion defines a condition that Machine must meet.
type MachineAssertion func(Machine) error

// MachineValidationDecorator decorates a MachineReader and runs a set of MachineAssertions on
// each Machine instance as they are read.
type MachineValidationDecorator struct {
	MachineReader
	assertions []MachineAssertion
}

// NewMachineValidationDedcorator creates a MachineValidationDecorator instance with no assertions
// registered.
func NewMachineValidationDecorator(reader MachineReader) *MachineValidationDecorator {
	return &MachineValidationDecorator{MachineReader: reader}
}

// NewDefaultMachineValidationDecorator creates a MachineValidationDecorator with a set of default
// assertions registered. See WithDefaultAssertions() for a list of defaults.
func NewDefaultMachineValidationDecorator(reader MachineReader) *MachineValidationDecorator {
	validator := NewMachineValidationDecorator(reader)
	WithDefaultAssertions(validator)
	return validator
}

// Read reads a Machine instance, runs the set of MachineAssertions registered with m
// against the Machine instance returning the first error.
func (m *MachineValidationDecorator) Read() (Machine, error) {
	machine, err := m.MachineReader.Read()
	if err != nil {
		return Machine{}, err
	}

	for _, fn := range m.assertions {
		if err := fn(machine); err != nil {
			return Machine{}, err
		}
	}

	return machine, nil
}

// Register registers v MachineAssertions with m.
func (m *MachineValidationDecorator) Register(v ...MachineAssertion) {
	m.assertions = append(m.assertions, v...)
}

var defaultAssertions = []MachineAssertion{
	UniqueHostnames(),
	UniqueIds(),
}

// WithDefaultAssertions applies a set of default assertions to validator. The default assertions include
// UniqueHostnames and UniqueIds.
func WithDefaultAssertions(validator *MachineValidationDecorator) {
	validator.Register(defaultAssertions...)
}

// UniqueHostnames asserts a given Machine instance does not contain a hostname previously observed
// in another Machine instance. It is not thread safe. It has a 1 time use.
func UniqueHostnames() MachineAssertion {
	hostnames := make(map[string]struct{})
	return func(m Machine) error {
		if _, ok := hostnames[m.Hostname]; ok {
			return fmt.Errorf("hostname re-use: %v", m.Hostname)
		}

		hostnames[m.Hostname] = struct{}{}

		return nil
	}
}

// UniqueIds asserts a given Machine instance does not contain a unique ID previously observed
// in another Machine instance. It is not thread safe. It has a 1 time use.
func UniqueIds() MachineAssertion {
	ids := make(map[string]struct{})
	return func(m Machine) error {
		if _, ok := ids[m.Id]; ok {
			return fmt.Errorf("cannot re-use an id for machines: %v", m.Id)
		}

		ids[m.Id] = struct{}{}

		return nil
	}
}
