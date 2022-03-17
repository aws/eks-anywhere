package hardware

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/tinkerbell/tink/protos/hardware"
	"github.com/tinkerbell/tink/protos/packet"
)

// ErrTinkebellHardwareJsonRepeatWrites occurs when a TinkerbellHardwareJson receives multiple calls to Write().
var ErrTinkebellHardwareJsonRepeatWrites = errors.New("TinkerbellHardwareJson can only be written to once")

// TinkerbellHardwareJsonFactory creates new TinkerbellHardwareJson instances. The path used by the
// TinkerbellHardwareJson is available as the second return parameter.
type TinkerbellHardwareJsonFactory interface {
	Create(name string) (*TinkerbellHardwareJson, string, error)
}

// TinkerbellHardwareJsonWriter writes discrete instances of TinkerbellHardwareJson's. Paths for files that were
// successfully written can be retrieved from Journal().
type TinkerbellHardwareJsonWriter struct {
	json    TinkerbellHardwareJsonFactory
	journal []string
}

// NewTinkerbellHardwareJsonWriter creates a newTinkerbellHardwareJsonWriter instance that uses factory to create
// TinkerbellHardwareJson instances and write to them.
func NewTinkerbellHardwareJsonWriter(factory TinkerbellHardwareJsonFactory) *TinkerbellHardwareJsonWriter {
	return &TinkerbellHardwareJsonWriter{json: factory}
}

// Write creates a new TinkerbellHardwareJson instance and writes m to it.
func (tw *TinkerbellHardwareJsonWriter) Write(m Machine) error {
	file, path, err := tw.json.Create(m.Hostname)
	if err != nil {
		return err
	}

	if err := file.Write(m); err != nil {
		return err
	}

	tw.journal = append(tw.journal, path)

	return nil
}

// Journal returns a list of json files that tw has created and successfully written a Machine instance to.
func (tw *TinkerbellHardwareJsonWriter) Journal() []string {
	return tw.journal
}

// TinkerbellHardwareJson represents a discrete Tinkerbell hardware json file. It can only be written to once.
type TinkerbellHardwareJson struct {
	writer io.WriteCloser
	closed bool
}

// NewTinkerbellHardwareJson creates a new TinkerbellHardwareJson instance that uses w as its destination for Write calls.
func NewTinkerbellHardwareJson(w io.WriteCloser) *TinkerbellHardwareJson {
	return &TinkerbellHardwareJson{writer: w}
}

// Write marshals m as a Tinkerbell hardware json object and writes it to tj's writer. Upon successfully completing the
// write, tj will close its writer. Subsequent calls to Write will return ErrTinkebellHardwareJsonRepeatWrites.
func (tj *TinkerbellHardwareJson) Write(m Machine) error {
	if tj.closed {
		return ErrTinkebellHardwareJsonRepeatWrites
	}
	defer func() { tj.closed = true; tj.writer.Close() }()

	marshalled, err := marshalTinkerbellHardwareJson(m)
	if err != nil {
		return err
	}

	if _, err := tj.writer.Write(marshalled); err != nil {
		return err
	}

	return nil
}

func marshalTinkerbellHardwareJson(m Machine) ([]byte, error) {
	return json.Marshal(
		Hardware{
			Id: m.Id,
			Metadata: &packet.Metadata{
				Facility: &packet.Metadata_Facility{
					FacilityCode: "onprem",
					PlanSlug:     "c2.medium.x86",
				},
				Instance: &packet.Metadata_Instance{
					Id:       m.Id,
					Hostname: m.Hostname,
					Storage: &packet.Metadata_Instance_Storage{
						Disks: []*packet.Metadata_Instance_Storage_Disk{
							{Device: "/dev/sda"},
						},
					},
				},
				State: "provisioning",
			},
			Network: &hardware.Hardware_Network{
				Interfaces: []*hardware.Hardware_Network_Interface{
					{
						Dhcp: &hardware.Hardware_DHCP{
							Arch:     "x86_64",
							Hostname: m.Hostname,
							Ip: &hardware.Hardware_DHCP_IP{
								Address: m.IpAddress,
								Gateway: m.Gateway,
								Netmask: m.Netmask,
							},
							Mac:         m.MacAddress,
							NameServers: m.Nameservers,
							Uefi:        true,
						},
						Netboot: &hardware.Hardware_Netboot{
							AllowPxe:      true,
							AllowWorkflow: true,
						},
					},
				},
			},
		},
	)
}

// tinkerbellHardwareJsonFactoryFunc is a convinience function type that satisfies the TinkerbellHardwareJsonFactory
// interface.
type tinkerbellHardwareJsonFactoryFunc func(name string) (*TinkerbellHardwareJson, string, error)

// Create calls fn returning its return values to the caller.
func (fn tinkerbellHardwareJsonFactoryFunc) Create(name string) (*TinkerbellHardwareJson, string, error) {
	return fn(name)
}

// NewTinkerbellHardwareJsonFactory creates a new TinkerbellHardwareJson with an *os.File. The file path will be rooted
// at basepath.
func NewTinkerbellHardwareJsonFactory(basepath string) TinkerbellHardwareJsonFactory {
	return tinkerbellHardwareJsonFactoryFunc(func(name string) (*TinkerbellHardwareJson, string, error) {
		path := filepath.Join(basepath, name)

		fh, err := os.Create(path)
		if err != nil {
			return nil, "", err
		}

		return NewTinkerbellHardwareJson(fh), path, nil
	})
}
