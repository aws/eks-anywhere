package hardware

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	apimachineryvalidation "k8s.io/apimachinery/pkg/util/validation"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/networkutils"
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
	RegisterDefaultAssertions(validator)
	return validator
}

// Validate validates machine by executing its Validate() method and passing it to all registered MachineAssertions.
func (mv *DefaultMachineValidator) Validate(machine Machine) error {
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

var (
	linuxPathRegex      = `^(/dev/[\w-]+)+$`
	linuxPathValidation = regexp.MustCompile(linuxPathRegex)
)

// StaticMachineAssertions defines all static data assertions performed on a Machine.
func StaticMachineAssertions() MachineAssertion {
	return func(m Machine) error {
		if m.IPAddress == "" {
			return newEmptyFieldError("IPAddress")
		}

		if err := networkutils.ValidateIP(m.IPAddress); err != nil {
			return fmt.Errorf("IPAddress: %v", err)
		}

		if m.Gateway == "" {
			return newEmptyFieldError("Gateway")
		}

		if err := networkutils.ValidateIP(m.Gateway); err != nil {
			return fmt.Errorf("Gateway: %v", err)
		}

		if len(m.Nameservers) == 0 {
			return newEmptyFieldError("Nameservers")
		}

		for _, nameserver := range m.Nameservers {
			if nameserver == "" {
				return newMachineError("Nameservers contains an empty entry")
			}
		}

		if m.Netmask == "" {
			return newEmptyFieldError("Netmask")
		}

		if m.MACAddress == "" {
			return newEmptyFieldError("MACAddress")
		}

		if _, err := net.ParseMAC(m.MACAddress); err != nil {
			return fmt.Errorf("MACAddress: %v", err)
		}

		if m.Hostname == "" {
			return newEmptyFieldError("Hostname")
		}

		if errs := apimachineryvalidation.IsDNS1123Subdomain(m.Hostname); len(errs) > 0 {
			return fmt.Errorf("invalid hostname: %v: %v", m.Hostname, errs)
		}

		if !linuxPathValidation.MatchString(m.Disk) {
			return fmt.Errorf(
				"disk must be a valid linux path (\"%v\")",
				linuxPathRegex,
			)
		}

		for key, value := range m.Labels {
			if err := validateLabelKey(key); err != nil {
				return err
			}

			if err := validateLabelValue(value); err != nil {
				return err
			}
		}

		if m.HasBMC() {
			if m.BMCIPAddress == "" {
				return newEmptyFieldError("BMCIPAddress")
			}

			if err := networkutils.ValidateIP(m.BMCIPAddress); err != nil {
				return fmt.Errorf("BMCIPAddress: %v", err)
			}

			if m.BMCOptions == nil || m.BMCOptions.RPC == nil {
				if m.BMCUsername == "" {
					return newEmptyFieldError("BMCUsername")
				}

				if m.BMCPassword == "" {
					return newEmptyFieldError("BMCPassword")
				}
			}
		}

		if m.VLANID != "" {
			i, err := strconv.Atoi(m.VLANID)
			if err != nil {
				return errors.New("VLANID: must be a string integer")
			}

			// valid VLAN IDs are between 1 and 4094 - https://en.m.wikipedia.org/wiki/VLAN#IEEE_802.1Q
			const (
				maxVLANID = 4094
				minVLANID = 1
			)
			if i < minVLANID || i > maxVLANID {
				return errors.New("VLANID: must be between 1 and 4094")
			}
		}

		return nil
	}
}

// UniqueIPAddress asserts a given Machine instance has a unique IPAddress field relative to previously seen Machine
// instances. It is not thread safe. It has a 1 time use.
func UniqueIPAddress() MachineAssertion {
	ips := make(map[string]struct{})
	return func(m Machine) error {
		if _, seen := ips[m.IPAddress]; seen {
			return fmt.Errorf("duplicate IPAddress: %v", m.IPAddress)
		}

		ips[m.IPAddress] = struct{}{}

		return nil
	}
}

// UniqueMACAddress asserts a given Machine instance has a unique MACAddress field relative to previously seen Machine
// instances. It is not thread safe. It has a 1 time use.
func UniqueMACAddress() MachineAssertion {
	macs := make(map[string]struct{})
	return func(m Machine) error {
		if _, seen := macs[m.MACAddress]; seen {
			return fmt.Errorf("duplicate MACAddress: %v", m.MACAddress)
		}

		macs[m.MACAddress] = struct{}{}

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

// UniqueBMCIPAddress asserts a given Machine instance has a unique BMCIPAddress field relative to previously seen
// Machine instances. If there is no BMC configuration as defined by machine.HasBMC() the check is a noop. It is
// not thread safe. It has a 1 time use.
func UniqueBMCIPAddress() MachineAssertion {
	ips := make(map[string]struct{})
	return func(m Machine) error {
		if !m.HasBMC() {
			return nil
		}

		if m.BMCIPAddress == "" {
			return fmt.Errorf("missing BMCIPAddress (mac=\"%v\")", m.MACAddress)
		}

		if _, seen := ips[m.BMCIPAddress]; seen {
			return fmt.Errorf("duplicate IPAddress: %v", m.BMCIPAddress)
		}

		ips[m.BMCIPAddress] = struct{}{}

		return nil
	}
}

// RegisterDefaultAssertions applies a set of default assertions to validator. The default assertions
// include UniqueHostnames and UniqueIDs.
func RegisterDefaultAssertions(validator *DefaultMachineValidator) {
	validator.Register([]MachineAssertion{
		StaticMachineAssertions(),
		UniqueIPAddress(),
		UniqueMACAddress(),
		UniqueHostnames(),
		UniqueBMCIPAddress(),
	}...)
}

func validateLabelKey(k string) error {
	if errs := apimachineryvalidation.IsQualifiedName(k); len(errs) != 0 {
		return fmt.Errorf("%v", strings.Join(errs, "; "))
	}
	return nil
}

func validateLabelValue(v string) error {
	if errs := apimachineryvalidation.IsValidLabelValue(v); len(errs) != 0 {
		return fmt.Errorf("%v", strings.Join(errs, "; "))
	}
	return nil
}

// LabelsMatchSelector ensures all selector key-value pairs can be found in labels.
// If selector is empty true is always returned.
func LabelsMatchSelector(selector v1alpha1.HardwareSelector, labels Labels) bool {
	for expectKey, expectValue := range selector {
		labelValue, hasLabel := labels[expectKey]
		if !hasLabel || labelValue != expectValue {
			return false
		}
	}
	return true
}
