package hardware_test

import (
	"testing"

	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell/rufio"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestCatalogue_BMC_Insert(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()

	err := catalogue.InsertBMC(&v1alpha1.Machine{})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(catalogue.TotalBMCs()).To(gomega.Equal(1))
}

func TestCatalogue_BMC_UnknownIndexErrors(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()

	_, err := catalogue.LookupBMC(hardware.BMCNameIndex, "Name")
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestCatalogue_BMC_Indexed(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue(hardware.WithBMCNameIndex())

	const name = "hello"
	expect := &v1alpha1.Machine{ObjectMeta: metav1.ObjectMeta{Name: name}}
	err := catalogue.InsertBMC(expect)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	received, err := catalogue.LookupBMC(hardware.BMCNameIndex, name)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(received).To(gomega.HaveLen(1))
	g.Expect(received[0]).To(gomega.Equal(expect))
}

func TestCatalogue_BMC_AllBMCsReceivesCopy(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue(hardware.WithHardwareIDIndex())

	const totalHardware = 1
	err := catalogue.InsertBMC(&v1alpha1.Machine{ObjectMeta: metav1.ObjectMeta{Name: "foo"}})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	changedHardware := catalogue.AllBMCs()
	g.Expect(changedHardware).To(gomega.HaveLen(totalHardware))

	changedHardware[0] = &v1alpha1.Machine{ObjectMeta: metav1.ObjectMeta{Name: "qux"}}

	unchangedHardware := catalogue.AllBMCs()
	g.Expect(unchangedHardware).ToNot(gomega.Equal(changedHardware))
}

func TestBMCCatalogueWriter_Write(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	writer := hardware.NewBMCCatalogueWriter(catalogue)
	machine := NewValidMachine()

	err := writer.Write(machine)
	g.Expect(err).To(gomega.Succeed())

	bmcs := catalogue.AllBMCs()
	g.Expect(bmcs).To(gomega.HaveLen(1))
	g.Expect(bmcs[0].Name).To(gomega.ContainSubstring(machine.Hostname))
	g.Expect(bmcs[0].Spec.Connection.Host).To(gomega.Equal(machine.BMCIPAddress))
	g.Expect(bmcs[0].Spec.Connection.AuthSecretRef.Name).To(gomega.ContainSubstring(machine.Hostname))
}

func TestBMCMachineWithOptions(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	writer := hardware.NewBMCCatalogueWriter(catalogue)
	machine := NewMachineWithOptions()
	want := &v1alpha1.Machine{Spec: v1alpha1.MachineSpec{
		Connection: v1alpha1.Connection{
			Host: "10.10.10.11",
			Port: 0,
			AuthSecretRef: v1.SecretReference{
				Name:      "bmc-localhost-auth",
				Namespace: constants.EksaSystemNamespace,
			},
			InsecureTLS: true,
			ProviderOptions: &v1alpha1.ProviderOptions{
				RPC: &v1alpha1.RPCOptions{
					ConsumerURL: "https://example.net",
					Request: &v1alpha1.RequestOpts{
						HTTPContentType: "application/vnd.api+json",
						HTTPMethod:      "POST",
						StaticHeaders:   map[string][]string{"myheader": {"myvalue"}},
						TimestampFormat: "2006-01-02T15:04:05Z07:00", // time.RFC3339
						TimestampHeader: "X-Example-Timestamp",
					},
					Signature: &v1alpha1.SignatureOpts{
						HeaderName:                 "X-Example-Signature",
						AppendAlgoToHeaderDisabled: true,
						IncludedPayloadHeaders:     []string{"X-Example-Timestamp"},
					},
					HMAC: &v1alpha1.HMACOpts{
						PrefixSigDisabled: true,
						Secrets: map[v1alpha1.HMACAlgorithm][]v1.SecretReference{
							v1alpha1.HMACAlgorithm("sha256"): {
								{Name: "bmc-localhost-auth-0", Namespace: constants.EksaSystemNamespace},
								{Name: "bmc-localhost-auth-1", Namespace: constants.EksaSystemNamespace},
							},
							v1alpha1.HMACAlgorithm("sha512"): {
								{Name: "bmc-localhost-auth-0", Namespace: constants.EksaSystemNamespace},
								{Name: "bmc-localhost-auth-1", Namespace: constants.EksaSystemNamespace},
							},
						},
					},
					Experimental: &v1alpha1.ExperimentalOpts{
						CustomRequestPayload: `{"data":{"type":"articles","id":"1","attributes":{"title": "Rails is Omakase"},"relationships":{"author":{"links":{"self":"/articles/1/relationships/author","related":"/articles/1/author"},"data":null}}}}`,
						DotPath:              "data.relationships.author.data",
					},
				},
			},
		},
	}}

	err := writer.Write(machine)
	g.Expect(err).To(gomega.Succeed())

	got := catalogue.AllBMCs()[0]
	g.Expect(got.Spec).To(gomega.Equal(want.Spec))
}
