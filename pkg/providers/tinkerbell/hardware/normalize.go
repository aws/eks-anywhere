package hardware

import "strings"

// NormalizerFunc applies a normalization transformation to the Machine.
type NormalizerFunc func(Machine) Machine

// Normalizer is a decorator for a MachineReader that applies a set of normalization funcs
// to machines.
type Normalizer struct {
	reader      MachineReader
	normalizers []NormalizerFunc
}

// NewNormalizer creates a Normalizer instance that decorates r's Read(). A set of default
// normalization functions are pre-registered.
func NewNormalizer(r MachineReader) *Normalizer {
	normalizer := NewRawNormalizer(r)
	RegisterDefaultNormalizations(normalizer)
	return normalizer
}

// NewRawNormalizer returns a Normalizer with default normalizations registered by
// RegisterDefaultNormalizations.
func NewRawNormalizer(r MachineReader) *Normalizer {
	return &Normalizer{reader: r}
}

// Read reads an Machine from the decorated MachineReader, applies all normalization funcs and
// returns the machine. If the decorated MachineReader errors, it is returned.
func (n Normalizer) Read() (Machine, error) {
	machine, err := n.reader.Read()
	if err != nil {
		return Machine{}, err
	}

	for _, fn := range n.normalizers {
		machine = fn(machine)
	}

	return machine, nil
}

// Register fn to n such that fn is run over each machine read from the wrapped MachineReader.
func (n *Normalizer) Register(fn NormalizerFunc) {
	n.normalizers = append(n.normalizers, fn)
}

// LowercaseMACAddress ensures m's MACAddress field has lower chase characters.
func LowercaseMACAddress(m Machine) Machine {
	m.MACAddress = strings.ToLower(m.MACAddress)
	return m
}

// RegisterDefaultNormalizations registers a set of default normalizations on n.
func RegisterDefaultNormalizations(n *Normalizer) {
	for _, fn := range []NormalizerFunc{
		LowercaseMACAddress,
	} {
		n.Register(fn)
	}
}
