package hardware

import (
	"fmt"
	"io"

	pbnjv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/pbnj/api/v1alpha1"
	tinkv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/tink/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// DefaultHardwareManifestYAMLFilename is the default file for writing yinkerbell yaml manifests
const DefaultHardwareManifestYAMLFilename = "hardware.yaml"

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
		return fmt.Errorf("marshalling tinkerbell hardware yaml (id=%v): %v", m.ID, err)
	}
	if err := yw.writeWithPrependedSeparator(hardware); err != nil {
		return fmt.Errorf("writing tinkerbell hardware yaml (id=%v): %v", m.ID, err)
	}

	bmc, err := marshalTinkerbellBMCYAML(m)
	if err != nil {
		return fmt.Errorf("marshalling tinkerbell bmc yaml (id=%v): %v", m.ID, err)
	}
	if err := yw.writeWithPrependedSeparator(bmc); err != nil {
		return fmt.Errorf("writing tinkerbell bmc yaml (id=%v): %v", m.ID, err)
	}

	secret, err := marshalSecretYAML(m)
	if err != nil {
		return fmt.Errorf("marshalling bmc secret yaml (id=%v): %v", m.ID, err)
	}
	if err := yw.writeWithPrependedSeparator(secret); err != nil {
		return fmt.Errorf("writing bmc secret yaml (id=%v): %v", m.ID, err)
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
	return yaml.Marshal(tinkv1alpha1.Hardware{})
}

func marshalTinkerbellBMCYAML(m Machine) ([]byte, error) {
	return yaml.Marshal(
		pbnjv1alpha1.BMC{},
	)
}

func marshalSecretYAML(m Machine) ([]byte, error) {
	return yaml.Marshal(
		corev1.Secret{},
	)
}
