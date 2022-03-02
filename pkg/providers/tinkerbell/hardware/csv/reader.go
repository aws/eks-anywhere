package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

// CSV header fields for configuring metadata field positions within a record.
const (
	HostnameHeader    = "hostname"
	IpAddressHeader   = "ip_address"
	GatewayHeader     = "gateway"
	NetmaskHeader     = "netmask"
	MacHeader         = "mac"
	NameServersHeader = "nameservers"
	BmcVendorHeader   = "bmc_vendor"
	BmcIpHeader       = "bmc_ip"
	BmcUsernameHeader = "bmc_username"
	BmcPasswordHeader = "bmc_password"
	// new headers should be added to Headers pkg var.
)

// Headers retrieves a slice of expected headers in the CSV.
func Headers() []string {
	return []string{
		HostnameHeader,
		IpAddressHeader,
		GatewayHeader,
		NetmaskHeader,
		MacHeader,
		NameServersHeader,
		BmcVendorHeader,
		BmcIpHeader,
		BmcUsernameHeader,
		BmcPasswordHeader,
	}
}

// RecordReader reads CSV records returning each field as a value in the slice.
type RecordReader interface {
	Read() (record []string, err error)
}

// Reader implements hardware.MachineReader. It converts CSV metadata into Machine instances. Reader expects the
// first record to represent headers. It uses the headers to determine what fields represent which metadata.
type Reader struct {
	reader RecordReader
	idx    headersIndex

	// a uuid generator func that can be monkey patched for testing.
	generateUuid func() string
}

// NewCSV creates a CSV instance. NewCSV assumes the first record in reader represents headers indicating the position
// of the expected fields. The expected headers are defined as constants in this package and begin with Header.
func NewReader(reader RecordReader) (*Reader, error) {
	csv := Reader{
		reader:       reader,
		generateUuid: func() string { return uuid.NewString() },
	}

	if err := csv.recordHeaderIndexes(); err != nil {
		return nil, err
	}

	return &csv, nil
}

// Open creates a new Reader instance that consumes filename as input. filename should be a path to a valid
// CSV file with the correct format including the header record.
func Open(filename string) (*Reader, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	recordReader := csv.NewReader(fh)

	return NewReader(recordReader)
}

// headersIndex defines the position of metadata in a CSV record. Positions are defined by the first
// record in the CSV that should contain the *Header strings.
type headersIndex struct {
	Hostname    int
	IpAddress   int
	Gateway     int
	Netmask     int
	Mac         int
	NameServer  int
	BmcVendor   int
	BmcIp       int
	BmcUsername int
	BmcPassword int
}

// Read a record from the CSV and construct a Machine from it. The Id of the Machine is randomly generated. If
// the BMC IP is specified Read assumes all BMC metadata is defined and configures the Machine's Bmc. If a BMC IP is
// not specified the Machine's Bmc field will be nil.
func (c *Reader) Read() (hardware.Machine, error) {
	line, err := c.reader.Read()
	if err != nil {
		return hardware.Machine{}, fmt.Errorf("csv: reading record: %v", err)
	}

	machine := hardware.Machine{
		Id:       c.generateUuid(),
		Hostname: line[c.idx.Hostname],
		Network: hardware.Network{
			Ip:          line[c.idx.IpAddress],
			Gateway:     line[c.idx.Gateway],
			NameServers: SplitNameServers(line[c.idx.NameServer]),
			Netmask:     line[c.idx.Netmask],
			Mac:         line[c.idx.Mac],
		},
	}

	// If there's an Ip address assume they've supplied a full Bmc configuration.
	if ip := line[c.idx.BmcIp]; ip != "" {
		machine.Bmc = &hardware.Bmc{
			Ip:       ip,
			Username: line[c.idx.BmcUsername],
			Password: line[c.idx.BmcPassword],
			Vendor:   line[c.idx.BmcVendor],
		}
	}

	return machine, nil
}

func (r *Reader) recordHeaderIndexes() error {
	headers, err := r.reader.Read()
	if err != nil {
		return fmt.Errorf("csv: parsing headers: %v", err)
	}

	var ok bool
	m := make(map[string]int, len(headers))
	for i, header := range headers {
		m[header] = i
	}

	if r.idx.Hostname, ok = m[HostnameHeader]; !ok {
		return newMissingHeaderError(HostnameHeader)
	}

	if r.idx.IpAddress, ok = m[IpAddressHeader]; !ok {
		return newMissingHeaderError(IpAddressHeader)
	}

	if r.idx.Gateway, ok = m[GatewayHeader]; !ok {
		return newMissingHeaderError(GatewayHeader)
	}

	if r.idx.Netmask, ok = m[NetmaskHeader]; !ok {
		return newMissingHeaderError(NetmaskHeader)
	}

	if r.idx.Mac, ok = m[MacHeader]; !ok {
		return newMissingHeaderError(MacHeader)
	}

	if r.idx.NameServer, ok = m[NameServersHeader]; !ok {
		return newMissingHeaderError(NameServersHeader)
	}

	if r.idx.BmcVendor, ok = m[BmcVendorHeader]; !ok {
		return newMissingHeaderError(BmcVendorHeader)
	}

	if r.idx.BmcIp, ok = m[BmcIpHeader]; !ok {
		return newMissingHeaderError(BmcIpHeader)
	}

	if r.idx.BmcUsername, ok = m[BmcUsernameHeader]; !ok {
		return newMissingHeaderError(BmcUsernameHeader)
	}

	if r.idx.BmcPassword, ok = m[BmcPasswordHeader]; !ok {
		return newMissingHeaderError(BmcPasswordHeader)
	}

	return nil
}

// NameServerJoinChar is used to join name servers so they are writable to a single CSV filed.
const NameServerJoinChar = "|"

// SplitNameServers splits a name servers string on NameServerJoinChar.
func SplitNameServers(servers string) []string {
	return strings.Split(servers, NameServerJoinChar)
}

// JoinNameServers joins several server strings using NameServerJoinChar.
func JoinNameServers(servers []string) string {
	return strings.Join(servers, NameServerJoinChar)
}

// WithIDGenerator sets reader's ID generator func to generator allowing callers to control
// the ID generation randomness.
func WithIDGenerator(reader *Reader, generator func() string) {
	reader.generateUuid = generator
}

func newMissingHeaderError(header string) error {
	return fmt.Errorf("csv: missing header %v", header)
}
