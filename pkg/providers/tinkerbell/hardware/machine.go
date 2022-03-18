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
	Id          string      `csv:"id"`
	IpAddress   string      `csv:"ip_address"`
	Gateway     string      `csv:"gateway"`
	Nameservers Nameservers `csv:"nameservers"`
	Netmask     string      `csv:"netmask"`
	MacAddress  string      `csv:"mac"`
	Hostname    string      `csv:"hostname"`

	BmcIpAddress string `csv:"bmc_ip"`
	BmcUsername  string `csv:"bmc_username"`
	BmcPassword  string `csv:"bmc_password"`
	BmcVendor    string `csv:"vendor"`
}

// HasBmc determines if m has a Bmc configuration. A Bmc configuration is present if any of the Bmc fields
// contain non-empty strings.
func (m *Machine) HasBmc() bool {
	return m.BmcIpAddress != "" || m.BmcUsername != "" || m.BmcPassword != "" || m.BmcVendor != ""
}

// Validate ensures all fields on m are valid. Bmc configurationis only validated if m.HasBmc() returns true.
func (m *Machine) Validate() error {
	if m.Id == "" {
		return newEmptyFieldError("Id")
	}

	if m.IpAddress == "" {
		return newEmptyFieldError("IpAddress")
	}

	if err := networkutils.ValidateIP(m.IpAddress); err != nil {
		return fmt.Errorf("IpAddress: %v", err)
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

	if m.MacAddress == "" {
		return newEmptyFieldError("MacAddress")
	}

	if _, err := net.ParseMAC(m.MacAddress); err != nil {
		return fmt.Errorf("MacAddress: %v", err)
	}

	if m.Hostname == "" {
		return newEmptyFieldError("Hostname")
	}

	if errs := apimachineryvalidation.IsDNS1123Subdomain(m.Hostname); len(errs) > 0 {
		return fmt.Errorf("invalid hostname: %v: %v", m.Hostname, errs)
	}

	if m.HasBmc() {
		if m.BmcIpAddress == "" {
			return newEmptyFieldError("BmcIpAddress")
		}

		if err := networkutils.ValidateIP(m.BmcIpAddress); err != nil {
			return fmt.Errorf("BmcIpAddress: %v", err)
		}

		if m.BmcUsername == "" {
			return newEmptyFieldError("BmcUsername")
		}

		if m.BmcPassword == "" {
			return newEmptyFieldError("BmcPassword")
		}

		if m.BmcVendor == "" {
			return newEmptyFieldError("BmcVendor")
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
