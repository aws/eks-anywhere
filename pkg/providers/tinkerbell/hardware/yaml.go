package hardware

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"
)

// TinkerbellManifestYAML is a MachineWriter that writes Tinkerbell manifests to a destination.
type TinkerbellManifestYAML struct {
	writer io.Writer
}

// NewTinkerbellManifestYAML creates a TinkerbellManifestYAML instance that writes its manifests to w.
func NewTinkerbellManifestYAML(w io.Writer) *TinkerbellManifestYAML {
	return &TinkerbellManifestYAML{writer: w}
}

// Write m as a set of Kubernetes manifests for use with Cluster API Tinkerbell Provider. This includes writing a
// Hardware, BMC and Secret (for the BMC).
func (yw *TinkerbellManifestYAML) Write(m Machine) error {
	hardware, err := marshalTinkerbellHardwareYAML(m)
	if err != nil {
		return fmt.Errorf("marshalling tinkerbell hardware yaml (mac=%v): %v", m.MACAddress, err)
	}
	if err := yw.writeWithPrependedSeparator(hardware); err != nil {
		return fmt.Errorf("writing tinkerbell hardware yaml (mac=%v): %v", m.MACAddress, err)
	}

	bmc, err := marshalTinkerbellBMCYAML(m)
	if err != nil {
		return fmt.Errorf("marshalling tinkerbell bmc yaml (mac=%v): %v", m.MACAddress, err)
	}
	if err := yw.writeWithPrependedSeparator(bmc); err != nil {
		return fmt.Errorf("writing tinkerbell bmc yaml (mac=%v): %v", m.MACAddress, err)
	}

	secret, err := marshalSecretYAML(m)
	if err != nil {
		return fmt.Errorf("marshalling bmc secret yaml (mac=%v): %v", m.MACAddress, err)
	}
	if err := yw.writeWithPrependedSeparator(secret); err != nil {
		return fmt.Errorf("writing bmc secret yaml (mac=%v): %v", m.MACAddress, err)
	}

	return nil
}

var yamlSeparatorWithNewline = []byte("---\n")

func (yw *TinkerbellManifestYAML) writeWithPrependedSeparator(data []byte) error {
	if err := yw.write(append(data, yamlSeparatorWithNewline...)); err != nil {
		return err
	}

	return nil
}

func (yw *TinkerbellManifestYAML) write(data []byte) error {
	if _, err := yw.writer.Write(data); err != nil {
		return err
	}
	return nil
}

// TODO(chrisdoherty4) Patch these types so we can generate yamls again with the new Hardware
// and BaseboardManagement types.

func marshalTinkerbellHardwareYAML(m Machine) ([]byte, error) {
	return yaml.Marshal(hardwareFromMachine(m))
}

func marshalTinkerbellBMCYAML(m Machine) ([]byte, error) {
	return yaml.Marshal(toRufioMachine(m))
}

func marshalSecretYAML(m Machine) ([]byte, error) {
	var final []byte
	for _, s := range baseboardManagementSecretFromMachine(m) {
		data, err := yaml.Marshal(s)
		if err != nil {
			return nil, err
		}
		final = append(final, data...)
		final = append(final, yamlSeparatorWithNewline...)
	}

	return final, nil
}

// CreateOrStdout will create path and return an *os.File if path is not empty. If path is empty
// os.Stdout is returned.
func CreateOrStdout(path string) (*os.File, error) {
	if path != "" {
		dir := filepath.Dir(path)
		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			return nil, fmt.Errorf("failed to create hardware yaml file: %v", err)
		}
		return os.Create(path)
	}
	return os.Stdout, nil
}
