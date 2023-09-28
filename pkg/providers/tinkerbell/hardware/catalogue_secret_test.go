package hardware_test

import (
	"fmt"
	"testing"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestCatalogue_Secret_Insert(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()

	err := catalogue.InsertSecret(&corev1.Secret{})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(catalogue.TotalSecrets()).To(gomega.Equal(1))
}

func TestCatalogue_Secret_UnknownIndexErrors(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()

	_, err := catalogue.LookupSecret(hardware.SecretNameIndex, "Name")
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestCatalogue_Secret_Indexed(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue(hardware.WithSecretNameIndex())

	const name = "hello"
	expect := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name}}
	err := catalogue.InsertSecret(expect)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	received, err := catalogue.LookupSecret(hardware.SecretNameIndex, name)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(received).To(gomega.HaveLen(1))
	g.Expect(received[0]).To(gomega.Equal(expect))
}

func TestCatalogue_Secret_AllSecretsReceivesCopy(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue(hardware.WithHardwareIDIndex())

	const totalHardware = 1
	err := catalogue.InsertSecret(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "foo"}})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	changedHardware := catalogue.AllSecrets()
	g.Expect(changedHardware).To(gomega.HaveLen(totalHardware))

	changedHardware[0] = &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "qux"}}

	unchangedHardware := catalogue.AllSecrets()
	g.Expect(unchangedHardware).ToNot(gomega.Equal(changedHardware))
}

func TestSecretCatalogueWriter_Write(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	writer := hardware.NewSecretCatalogueWriter(catalogue)
	machine := NewValidMachine()

	err := writer.Write(machine)
	g.Expect(err).To(gomega.Succeed())

	secrets := catalogue.AllSecrets()
	g.Expect(secrets).To(gomega.HaveLen(1))
	g.Expect(secrets[0].Name).To(gomega.ContainSubstring(machine.Hostname))
	g.Expect(secrets[0].Data).To(gomega.HaveKeyWithValue("username", []byte(machine.BMCUsername)))
	g.Expect(secrets[0].Data).To(gomega.HaveKeyWithValue("password", []byte(machine.BMCPassword)))
}

// TestRPCSecrets add RPC secrets to the catalogue and verifies that they are added correctly.
func TestRPCSecrets(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	writer := hardware.NewSecretCatalogueWriter(catalogue)
	machine := NewValidMachine()
	machine.BMCOptions = &hardware.BMCOptions{
		RPC: &hardware.RPCOpts{
			ConsumerURL: "http://localhost:8080",
			HMAC: hardware.HMACOpts{
				Secrets: []string{"superSecret1", "superSecret2"},
			},
		},
	}

	err := writer.Write(machine)
	g.Expect(err).To(gomega.Succeed())

	secrets := catalogue.AllSecrets()
	g.Expect(secrets).To(gomega.HaveLen(2))
	for idx, secret := range secrets {
		g.Expect(secret.Name).To(gomega.ContainSubstring(fmt.Sprintf("bmc-%v-auth-%v", machine.Hostname, idx)))
		g.Expect(secret.Data).To(gomega.HaveKeyWithValue("secret", []byte(fmt.Sprintf("superSecret%v", idx+1))))
	}
}
