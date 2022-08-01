package hardware

import (
	"fmt"
	"sort"
	"strings"
)

// Machine is a machine configuration with optional BMC interface configuration.
type Machine struct {
	Hostname    string      `csv:"hostname"`
	IPAddress   string      `csv:"ip_address"`
	Netmask     string      `csv:"netmask"`
	Gateway     string      `csv:"gateway"`
	Nameservers Nameservers `csv:"nameservers"`
	MACAddress  string      `csv:"mac"`

	// Disk used to populate the default workflow actions.
	// Currently needs to be the same for all hardware residing in the same group where a group
	// is either: control plane hardware, external etcd hard, or the definable worker node groups.
	Disk string `csv:"disk"`

	// Labels to be applied to the Hardware resource.
	Labels Labels `csv:"labels"`

	BMCIPAddress string `csv:"bmc_ip, omitempty"`
	BMCUsername  string `csv:"bmc_username, omitempty"`
	BMCPassword  string `csv:"bmc_password, omitempty"`
}

// HasBMC determines if m has a BMC configuration. A BMC configuration is present if any of the BMC fields
// contain non-empty strings.
func (m *Machine) HasBMC() bool {
	return m.BMCIPAddress != "" || m.BMCUsername != "" || m.BMCPassword != ""
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

// LabelSSeparator is used to separate key value label pairs.
const LabelsSeparator = "|"

// Labels defines a lebsl set. It satisfies https://pkg.go.dev/k8s.io/apimachinery/pkg/labels#Labels.
type Labels map[string]string

// Get returns the value for the provided label.
func (l Labels) Has(k string) bool {
	_, ok := l[k]
	return ok
}

// See https://pkg.go.dev/k8s.io/apimachinery/pkg/labels#Labels
func (l Labels) Get(k string) string {
	return l[k]
}

func (l *Labels) MarshalCSV() (string, error) {
	return l.String(), nil
}

func (l *Labels) UnmarshalCSV(s string) error {
	// Ensure we make the map so consumers of l don't segfault.
	*l = make(Labels)

	// Cater for no labels being specified.
	split := strings.Split(s, LabelsSeparator)
	if len(split) == 1 && split[0] == "" {
		return nil
	}

	for _, pair := range split {
		keyValue := strings.Split(strings.TrimSpace(pair), "=")
		if len(keyValue) != 2 {
			return fmt.Errorf("badly formatted key-value pair: %v", pair)
		}

		(*l)[strings.TrimSpace(keyValue[0])] = strings.TrimSpace(keyValue[1])
	}
	return nil
}

func (l Labels) String() string {
	labels := make([]string, 0, len(l))
	for key, value := range l {
		labels = append(labels, fmt.Sprintf("%v=%v", key, value))
	}
	// Sort for determinism.
	sort.StringSlice(labels).Sort()
	return strings.Join(labels, LabelsSeparator)
}

func newEmptyFieldError(s string) error {
	return newMachineError(fmt.Sprintf("%v is empty", s))
}

func newMachineError(s string) error {
	return fmt.Errorf("machine: %v", s)
}
