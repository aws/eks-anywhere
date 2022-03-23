package hardware

import (
	"fmt"
	"io"

	pbnjv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/pbnj/api/v1alpha1"
	tinkv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/tink/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterctlv1alpha3 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/constants"
)

// DefaultHardwareManifestYamlFilename is the default file for writing yinkerbell yaml manifests
const DefaultHardwareManifestYamlFilename = "hardware.yaml"

// Kubernetes related constants for describing kinds and api versions.
const (
	tinkerbellApiVersion = "tinkerbell.org/v1alpha1"
	hardwareKind         = "Hardware"
	bmcKind              = "BMC"

	secretApiVersion = "v1"
	secretKind       = "Secret"
)

// TinkerbellManifestYaml is a MachineWriter that writes Tinkerbell manifests to a destination.
type TinkerbellManifestYaml struct {
	writer io.Writer
}

// NewTinkerbellManifestYaml creates a TinkerbellManifestYaml instance that writes its manifests to w.
func NewTinkerbellManifestYaml(w io.Writer) *TinkerbellManifestYaml {
	return &TinkerbellManifestYaml{writer: w}
}

// Write m as a set of Kubernetes manifests for use with Cluster API Tinkerbell Provider. This includes writing a
// Hardware, BMC and Secret (for the BMC).
func (yw *TinkerbellManifestYaml) Write(m Machine) error {
	hardware, err := marshalTinkerbellHardwareYaml(m)
	if err != nil {
		return fmt.Errorf("marshalling tinkerbell hardware yaml (id=%v): %v", m.Id, err)
	}
	if err := yw.writeWithPrependedSeparator(hardware); err != nil {
		return fmt.Errorf("writing tinkerbell hardware yaml (id=%v): %v", m.Id, err)
	}

	bmc, err := marshalTinkerbellBmcYaml(m)
	if err != nil {
		return fmt.Errorf("marshalling tinkerbell bmc yaml (id=%v): %v", m.Id, err)
	}
	if err := yw.writeWithPrependedSeparator(bmc); err != nil {
		return fmt.Errorf("writing tinkerbell bmc yaml (id=%v): %v", m.Id, err)
	}

	secret, err := marshalSecretYaml(m)
	if err != nil {
		return fmt.Errorf("marshalling bmc secret yaml (id=%v): %v", m.Id, err)
	}
	if err := yw.writeWithPrependedSeparator(secret); err != nil {
		return fmt.Errorf("writing bmc secret yaml (id=%v): %v", m.Id, err)
	}

	return nil
}

var yamlSeparatorWithNewline = []byte("---\n")

func (yw *TinkerbellManifestYaml) writeWithPrependedSeparator(data []byte) error {
	if err := yw.write(append(data, yamlSeparatorWithNewline...)); err != nil {
		return err
	}

	return nil
}

func (yw *TinkerbellManifestYaml) write(data []byte) error {
	if _, err := yw.writer.Write(data); err != nil {
		return err
	}
	return nil
}

func marshalTinkerbellHardwareYaml(m Machine) ([]byte, error) {
	return yaml.Marshal(
		tinkv1alpha1.Hardware{
			TypeMeta: metav1.TypeMeta{
				Kind:       hardwareKind,
				APIVersion: tinkerbellApiVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      m.Hostname,
				Namespace: constants.EksaSystemNamespace,
				Labels: map[string]string{
					clusterctlv1alpha3.ClusterctlMoveLabelName: "true",
				},
			},
			Spec: tinkv1alpha1.HardwareSpec{
				ID:     m.Id,
				BmcRef: formatBmcRef(m),
			},
		},
	)
}

func marshalTinkerbellBmcYaml(m Machine) ([]byte, error) {
	return yaml.Marshal(
		pbnjv1alpha1.BMC{
			TypeMeta: metav1.TypeMeta{
				Kind:       bmcKind,
				APIVersion: tinkerbellApiVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      formatBmcRef(m),
				Namespace: constants.EksaSystemNamespace,
				Labels: map[string]string{
					clusterctlv1alpha3.ClusterctlMoveLabelName: "true",
				},
			},
			Spec: pbnjv1alpha1.BMCSpec{
				Host:   m.BmcIpAddress,
				Vendor: m.BmcVendor,
				AuthSecretRef: corev1.SecretReference{
					Name:      formatBmcSecretRef(m),
					Namespace: constants.EksaSystemNamespace,
				},
			},
		},
	)
}

func marshalSecretYaml(m Machine) ([]byte, error) {
	return yaml.Marshal(
		corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       secretKind,
				APIVersion: secretApiVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      formatBmcSecretRef(m),
				Namespace: constants.EksaSystemNamespace,
				Labels: map[string]string{
					clusterctlv1alpha3.ClusterctlMoveLabelName: "true",
				},
			},
			Type: "kubernetes.io/basic-auth",
			Data: map[string][]byte{
				"username": []byte(m.BmcUsername),
				"password": []byte(m.BmcPassword),
			},
		},
	)
}

func formatBmcRef(m Machine) string {
	return fmt.Sprintf("bmc-%s", m.Hostname)
}

func formatBmcSecretRef(m Machine) string {
	return fmt.Sprintf("%s-auth", formatBmcRef(m))
}
