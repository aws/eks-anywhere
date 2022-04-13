package hardware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tinkerbell/tink/protos/hardware"
	"github.com/tinkerbell/tink/protos/packet"
)

// ErrTinkebellHardwareJsonRepeatWrites occurs when a TinkerbellHardwareJson receives multiple calls to Write().
var ErrTinkebellHardwareJsonRepeatWrites = errors.New("TinkerbellHardwareJson can only be written to once")

// TinkerbellHardwareJsonFactory creates new TinkerbellHardwareJson instances.
type TinkerbellHardwareJsonFactory interface {
	Create(name string) (*TinkerbellHardwareJson, error)
}

// TinkerbellHardwareJsonWriter writes discrete instances of TinkerbellHardwareJson's. Paths for files that were
// successfully written can be retrieved from Journal().
type TinkerbellHardwareJsonWriter struct {
	json TinkerbellHardwareJsonFactory
}

// NewTinkerbellHardwareJsonWriter creates a newTinkerbellHardwareJsonWriter instance that uses factory to create
// TinkerbellHardwareJson instances and write to them.
func NewTinkerbellHardwareJsonWriter(factory TinkerbellHardwareJsonFactory) *TinkerbellHardwareJsonWriter {
	return &TinkerbellHardwareJsonWriter{json: factory}
}

// Write creates a new TinkerbellHardwareJson instance and writes m to it.
func (tw *TinkerbellHardwareJsonWriter) Write(m Machine) error {
	file, err := tw.json.Create(fmt.Sprintf("%v.json", m.Hostname))
	if err != nil {
		return err
	}

	if err := file.Write(m); err != nil {
		return err
	}

	return nil
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

// Hardware describes the hardware json structure required by the Tinkerbell API when registering hardware.
type Hardware struct {
	Id       string                     `json:"id"`
	Metadata *packet.Metadata           `json:"metadata"`
	Network  *hardware.Hardware_Network `json:"network"`
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

// Journal is an io.Writer that records the byte data passed to Write() as distinct chunks.
type Journal [][]byte

// Write records b as a distinct chunk in journal before returning len(b). It never returns an error.
func (journal *Journal) Write(b []byte) (int, error) {
	*journal = append(*journal, b)
	return len(b), nil
}

// tinkerbellHardwareJsonFactoryFunc is a convinience function type that satisfies the TinkerbellHardwareJsonFactory
// interface.
type tinkerbellHardwareJsonFactoryFunc func(name string) (*TinkerbellHardwareJson, error)

// Create calls fn returning its return values to the caller.
func (fn tinkerbellHardwareJsonFactoryFunc) Create(name string) (*TinkerbellHardwareJson, error) {
	return fn(name)
}

// RecordingTinkerbellHardwareJsonFactory creates a new TinkerbellHardwareJson where all writes to the
// TinkerbellHardwareJson are recorded in journal.
func RecordingTinkerbellHardwareJsonFactory(basepath string, journal *Journal) (TinkerbellHardwareJsonFactory, error) {
	info, err := os.Stat(basepath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("basepath does not exist: %v", basepath)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("basepath is not a directory: %v", basepath)
	}

	return tinkerbellHardwareJsonFactoryFunc(func(name string) (*TinkerbellHardwareJson, error) {
		path := filepath.Join(basepath, name)

		fh, err := os.Create(path)
		if err != nil {
			return nil, err
		}

		// Create a simple construct that satisfies io.WriteCloser where the io.Writer writes to both the journal
		// and the file and the io.Closer closes the file.
		writer := struct {
			io.Writer
			io.Closer
		}{
			Writer: io.MultiWriter(fh, journal),
			Closer: fh,
		}

		return NewTinkerbellHardwareJson(writer), nil
	}), nil
}

// TinkerbellHardwarePusher registers hardware with a Tinkerbell stack.
type TinkerbellHardwarePusher interface {
	PushHardware(ctx context.Context, hardware []byte) error
}

// RegisterTinkerbellHardware uses client to push all serializedJsons representing TinkerbellHardwareJson to a
// Tinkerbell server.
func RegisterTinkerbellHardware(ctx context.Context, client TinkerbellHardwarePusher, serializedJsons [][]byte) error {
	for _, json := range serializedJsons {
		if err := client.PushHardware(ctx, json); err != nil {
			return err
		}
	}
	return nil
}
