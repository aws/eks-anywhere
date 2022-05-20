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

// ErrTinkebellHardwareJSONRepeatWrites occurs when a TinkerbellHardwareJSON receives multiple calls to Write().
var ErrTinkebellHardwareJSONRepeatWrites = errors.New("TinkerbellHardwareJSON can only be written to once")

// TinkerbellHardwareJSONFactory creates new TinkerbellHardwareJSON instances.
type TinkerbellHardwareJSONFactory interface {
	Create(name string) (*TinkerbellHardwareJSON, error)
}

// TinkerbellHardwareJSONWriter writes discrete instances of TinkerbellHardwareJSON's. Paths for files that were
// successfully written can be retrieved from Journal().
type TinkerbellHardwareJSONWriter struct {
	json TinkerbellHardwareJSONFactory
}

// NewTinkerbellHardwareJSONWriter creates a newTinkerbellHardwareJSONWriter instance that uses factory to create
// TinkerbellHardwareJSON instances and write to them.
func NewTinkerbellHardwareJSONWriter(factory TinkerbellHardwareJSONFactory) *TinkerbellHardwareJSONWriter {
	return &TinkerbellHardwareJSONWriter{json: factory}
}

// Write creates a new TinkerbellHardwareJSON instance and writes m to it.
func (tw *TinkerbellHardwareJSONWriter) Write(m Machine) error {
	file, err := tw.json.Create(fmt.Sprintf("%v.json", m.Hostname))
	if err != nil {
		return err
	}

	if err := file.Write(m); err != nil {
		return err
	}

	return nil
}

// TinkerbellHardwareJSON represents a discrete Tinkerbell hardware json file. It can only be written to once.
type TinkerbellHardwareJSON struct {
	writer io.WriteCloser
	closed bool
}

// NewTinkerbellHardwareJSON creates a new TinkerbellHardwareJSON instance that uses w as its destination for Write calls.
func NewTinkerbellHardwareJSON(w io.WriteCloser) *TinkerbellHardwareJSON {
	return &TinkerbellHardwareJSON{writer: w}
}

// Write marshals m as a Tinkerbell hardware json object and writes it to tj's writer. Upon successfully completing the
// write, tj will close its writer. Subsequent calls to Write will return ErrTinkebellHardwareJSONRepeatWrites.
func (tj *TinkerbellHardwareJSON) Write(m Machine) error {
	if tj.closed {
		return ErrTinkebellHardwareJSONRepeatWrites
	}
	defer func() { tj.closed = true; tj.writer.Close() }()

	marshalled, err := marshalTinkerbellHardwareJSON(m)
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
	ID       string                     `json:"id"`
	Metadata *packet.Metadata           `json:"metadata"`
	Network  *hardware.Hardware_Network `json:"network"`
}

func marshalTinkerbellHardwareJSON(m Machine) ([]byte, error) {
	return json.Marshal(
		Hardware{
			ID: m.ID,
			Metadata: &packet.Metadata{
				Facility: &packet.Metadata_Facility{
					FacilityCode: "onprem",
					PlanSlug:     "c2.medium.x86",
				},
				Instance: &packet.Metadata_Instance{
					Id:       m.ID,
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
								Address: m.IPAddress,
								Gateway: m.Gateway,
								Netmask: m.Netmask,
							},
							Mac:         m.MACAddress,
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

// tinkerbellHardwareJSONFactoryFunc is a convinience function type that satisfies the TinkerbellHardwareJSONFactory
// interface.
type tinkerbellHardwareJSONFactoryFunc func(name string) (*TinkerbellHardwareJSON, error)

// Create calls fn returning its return values to the caller.
func (fn tinkerbellHardwareJSONFactoryFunc) Create(name string) (*TinkerbellHardwareJSON, error) {
	return fn(name)
}

// RecordingTinkerbellHardwareJSONFactory creates a new TinkerbellHardwareJSON where all writes to the
// TinkerbellHardwareJSON are recorded in journal.
func RecordingTinkerbellHardwareJSONFactory(basepath string, journal *Journal) (TinkerbellHardwareJSONFactory, error) {
	info, err := os.Stat(basepath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("basepath does not exist: %v", basepath)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("basepath is not a directory: %v", basepath)
	}

	return tinkerbellHardwareJSONFactoryFunc(func(name string) (*TinkerbellHardwareJSON, error) {
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

		return NewTinkerbellHardwareJSON(writer), nil
	}), nil
}

// TinkerbellHardwarePusher registers hardware with a Tinkerbell stack.
type TinkerbellHardwarePusher interface {
	PushHardware(ctx context.Context, hardware []byte) error
}

// RegisterTinkerbellHardware uses client to push all serializedJSONs representing TinkerbellHardwareJSON to a
// Tinkerbell server.
func RegisterTinkerbellHardware(ctx context.Context, client TinkerbellHardwarePusher, serializedJSONs [][]byte) error {
	for _, json := range serializedJSONs {
		if err := client.PushHardware(ctx, json); err != nil {
			return err
		}
	}
	return nil
}
