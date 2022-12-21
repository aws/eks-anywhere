package hardware_test

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/tinkerbell/rufio/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestCatalogue_BMC_Insert(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()

	err := catalogue.InsertBMC(&v1alpha1.BaseboardManagement{})
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
	expect := &v1alpha1.BaseboardManagement{ObjectMeta: metav1.ObjectMeta{Name: name}}
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
	err := catalogue.InsertBMC(&v1alpha1.BaseboardManagement{ObjectMeta: metav1.ObjectMeta{Name: "foo"}})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	changedHardware := catalogue.AllBMCs()
	g.Expect(changedHardware).To(gomega.HaveLen(totalHardware))

	changedHardware[0] = &v1alpha1.BaseboardManagement{ObjectMeta: metav1.ObjectMeta{Name: "qux"}}

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
