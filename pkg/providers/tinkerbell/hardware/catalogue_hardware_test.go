package hardware_test

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/tinkerbell/cluster-api-provider-tinkerbell/tink/api/v1alpha1"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestCatalogue_Hardware_Insert(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()

	err := catalogue.InsertHardware(&v1alpha1.Hardware{})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(catalogue.TotalHardware()).To(gomega.Equal(1))
}

func TestCatalogue_Hardware_UnknownIndexErrors(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()

	_, err := catalogue.LookupHardware(hardware.HardwareIDIndex, "ID")
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestCatalogue_Hardware_IDIndex(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue(hardware.WithHardwareIDIndex())

	const id = "hello"
	expect := &v1alpha1.Hardware{Spec: v1alpha1.HardwareSpec{ID: id}}
	err := catalogue.InsertHardware(expect)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	received, err := catalogue.LookupHardware(hardware.HardwareIDIndex, id)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(received).To(gomega.HaveLen(1))
	g.Expect(received[0]).To(gomega.Equal(expect))
}

func TestCatalogue_Hardware_BmcRefIndex(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue(hardware.WithHardwareBMCRefIndex())

	const ref = "bmc-ref"
	expect := &v1alpha1.Hardware{Spec: v1alpha1.HardwareSpec{BmcRef: ref}}
	err := catalogue.InsertHardware(expect)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	received, err := catalogue.LookupHardware(hardware.HardwareBMCRefIndex, ref)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(received).To(gomega.HaveLen(1))
	g.Expect(received[0]).To(gomega.Equal(expect))
}

func TestCatalogue_Hardware_AllHardwareReceivesCopy(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue(hardware.WithHardwareIDIndex())

	const totalHardware = 1
	err := catalogue.InsertHardware(&v1alpha1.Hardware{Spec: v1alpha1.HardwareSpec{ID: "foo"}})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	changedHardware := catalogue.AllHardware()
	g.Expect(changedHardware).To(gomega.HaveLen(totalHardware))

	changedHardware[0] = &v1alpha1.Hardware{Spec: v1alpha1.HardwareSpec{ID: "qux"}}

	unchangedHardware := catalogue.AllHardware()
	g.Expect(unchangedHardware).ToNot(gomega.Equal(changedHardware))
}
