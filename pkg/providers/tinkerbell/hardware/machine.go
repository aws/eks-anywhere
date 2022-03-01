package hardware

// Machine represents configuration for a single machine to be used in Tinkerbell provisioning.
type Machine struct {
	// A unique identifier for the machine.
	Id string

	// Hostname of the machine.
	Hostname string

	// Network describes the machines network configuration.
	Network Network

	// Optional BMC metadata. Nil means no Bmc configuration is specified.
	Bmc *Bmc
}

// Networks holds the metadata for a machines network configuration.
type Network struct {
	// The host IP address
	Ip string

	// The hosts default gateway address.
	Gateway string

	// The hosts netmask
	Netmask string

	// The hosts mac address.
	Mac string

	// The hosts name server configuration.
	NameServers []string
}

// Bmc holds metadata for connecting to a Baseboard Management Computer.
type Bmc struct {
	// IP address of the BMC.
	Ip string

	// The username to authenticate when issuing BMC commands.
	Username string

	// Password contains the BMC password for the user.
	Password string

	// A vendor identifier such as "Supermicro"
	Vendor string
}

// HasBmc determines if m has a Bmc configuration.
func (m Machine) HasBmc() bool {
	return m.Bmc != nil
}
