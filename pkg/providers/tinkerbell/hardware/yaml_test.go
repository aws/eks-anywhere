package hardware_test

import (
	"bufio"
	"bytes"
	"errors"
	"testing"

	"github.com/onsi/gomega"
	pbnjv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/pbnj/api/v1alpha1"
	tinkv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/tink/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apimachineryyaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestTinkerbellManifestYAMLWrites(t *testing.T) {
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

	var bmc pbnjv1alpha1.BMC
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
	AsserBMCSecretRepresentsMachine(g, secret, expect)
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
	g.Expect(h.Spec.ID).To(gomega.Equal(m.ID))
}

func AssertTinkerbellBMCRepresentsMachine(g *gomega.WithT, b pbnjv1alpha1.BMC, m hardware.Machine) {
	g.Expect(b.Spec.Host).To(gomega.Equal(m.BMCIPAddress))
	g.Expect(b.Spec.Vendor).To(gomega.Equal(m.BMCVendor))
}

func AsserBMCSecretRepresentsMachine(g *gomega.WithT, s corev1.Secret, m hardware.Machine) {
	g.Expect(s.Data).To(gomega.HaveKeyWithValue("username", []byte(m.BMCUsername)))
	g.Expect(s.Data).To(gomega.HaveKeyWithValue("password", []byte(m.BMCPassword)))
}

type ErrWriter struct{}

func (ErrWriter) Write([]byte) (int, error) {
	return 0, errors.New("ErrWriter: always return an error")
}
