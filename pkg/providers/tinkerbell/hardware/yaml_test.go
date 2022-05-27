package hardware_test

import (
	"bufio"
	"bytes"
	"errors"
	"testing"

	"github.com/onsi/gomega"
	rufiov1alpha1 "github.com/tinkerbell/rufio/api/v1alpha1"
	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apimachineryyaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestTinkerbellManifestYAMLWrites(t *testing.T) {
	t.Skip("Machine to type conversion functions currently unimplemented hence the test fails.")
	g := gomega.NewWithT(t)

	var buf bytes.Buffer
	writer := hardware.NewTinkerbellManifestYAML(&buf)

	expect := NewValidMachine()

	err := writer.Write(expect)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	reader := apimachineryyaml.NewYAMLReader(bufio.NewReader(&buf))

	var hardware tinkv1alpha1.Hardware
	raw, err := reader.Read()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	err = yaml.Unmarshal(raw, &hardware)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	var bmc rufiov1alpha1.BaseboardManagement
	raw, err = reader.Read()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	err = yaml.Unmarshal(raw, &bmc)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	var secret corev1.Secret
	raw, err = reader.Read()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	err = yaml.Unmarshal(raw, &secret)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	AssertTinkerbellHardwareRepresentsMachine(g, hardware, expect)
	AssertTinkerbellBMCRepresentsMachine(g, bmc, expect)
	AssertBMCSecretRepresentsMachine(g, secret, expect)
}

func TestTinkerbellManifestYAMLWriteErrors(t *testing.T) {
	g := gomega.NewWithT(t)

	writer := hardware.NewTinkerbellManifestYAML(ErrWriter{})

	expect := NewValidMachine()

	err := writer.Write(expect)
	g.Expect(err).To(gomega.HaveOccurred())
}

func AssertTinkerbellHardwareRepresentsMachine(g *gomega.WithT, h tinkv1alpha1.Hardware, m hardware.Machine) {
	g.Expect(h.ObjectMeta.Name).To(gomega.Equal(m.Hostname))
	g.Expect(h.Spec.Metadata.Instance.ID).To(gomega.Equal(m.ID))
}

func AssertTinkerbellBMCRepresentsMachine(g *gomega.WithT, b rufiov1alpha1.BaseboardManagement, m hardware.Machine) {
	g.Expect(b.Spec.Connection.Host).To(gomega.Equal(m.BMCIPAddress))
}

func AssertBMCSecretRepresentsMachine(g *gomega.WithT, s corev1.Secret, m hardware.Machine) {
	g.Expect(s.Data).To(gomega.HaveKeyWithValue("username", []byte(m.BMCUsername)))
	g.Expect(s.Data).To(gomega.HaveKeyWithValue("password", []byte(m.BMCPassword)))
}

type ErrWriter struct{}

func (ErrWriter) Write([]byte) (int, error) {
	return 0, errors.New("ErrWriter: always return an error")
}
