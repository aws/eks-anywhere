package hardware

import (
	"fmt"
	"net"
	"strings"

	apimachineryvalidation "k8s.io/apimachinery/pkg/util/validation"

	"github.com/aws/eks-anywhere/pkg/networkutils"
)

// Machine is a machine configuration with optional BMC interface configuration.
type Machine struct {
	ID          string      `csv:"id"`
	IPAddress   string      `csv:"ip_address"`
	Gateway     string      `csv:"gateway"`
	Nameservers Nameservers `csv:"nameservers"`
	Netmask     string      `csv:"netmask"`
	MACAddress  string      `csv:"mac"`
	Hostname    string      `csv:"hostname"`

	BMCIPAddress string `csv:"bmc_ip"`
	BMCUsername  string `csv:"bmc_username"`
	BMCPassword  string `csv:"bmc_password"`
	BMCVendor    string `csv:"vendor"`
}

// HasBMC determines if m has a BMC configuration. A BMC configuration is present if any of the BMC fields
// contain non-empty strings.
func (m *Machine) HasBMC() bool {
	return m.BMCIPAddress != "" || m.BMCUsername != "" || m.BMCPassword != "" || m.BMCVendor != ""
}

// Validate ensures all fields on m are valid. BMC configurationis only validated if m.HasBMC() returns true.
func (m *Machine) Validate() error {
	if m.ID == "" {
		return newEmptyFieldError("ID")
	}

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

	if m.HasBMC() {
		if m.BMCIPAddress == "" {
			return newEmptyFieldError("BMCIPAddress")
		}

		if err := networkutils.ValidateIP(m.BMCIPAddress); err != nil {
			return fmt.Errorf("BMCIPAddress: %v", err)
		}

		if m.BMCUsername == "" {
			return newEmptyFieldError("BMCUsername")
		}

		if m.BMCPassword == "" {
			return newEmptyFieldError("BMCPassword")
		}

		if m.BMCVendor == "" {
			return newEmptyFieldError("BMCVendor")
		}
	}

	return nil
}

// NameserversSeparator is used to unmarshal Nameservers.
const NameserversSeparator = "|"

// Nameservers is a custom type that can unmarshal a CSV representation of nameservers.
type Nameservers []string

func (n *Nameservers) String() string {
	return strings.Join(*n, NameserversSeparator)
}

// UnmarshalCSV unmarshalls s where is is a list of nameservers separated by NameserversSeparator.
func (n *Nameservers) UnmarshalCSV(s string) error {
	servers := strings.Split(s, NameserversSeparator)
	*n = append(*n, servers...)
	return nil
}

// MarshalCSV marshalls Nameservers into a string list of nameservers separated by NameserversSeparator.
func (n *Nameservers) MarshalCSV() (string, error) {
	return n.String(), nil
}

func newEmptyFieldError(s string) error {
	return newMachineError(fmt.Sprintf("%v is empty", s))
}

func newMachineError(s string) error {
	return fmt.Errorf("machine: %v", s)
}
