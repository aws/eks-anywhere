package hardware_test

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestHardwareCatalogue_Insert(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewHardwareCatalogue()

	err := catalogue.InsertHardware(&v1alpha1.Hardware{})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(catalogue.TotalHardware()).To(gomega.Equal(1))
}

func TestHardwareCatalogue_UnknownIndexErrors(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewHardwareCatalogue()

	_, err := catalogue.LookupHardware(hardware.HardwareIDIndex, "ID")
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestHardwareCatalogue_AllHardwareReceivesCopy(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewHardwareCatalogue()

	const totalHardware = 1
	err := catalogue.InsertHardware(&v1alpha1.Hardware{ObjectMeta: metav1.ObjectMeta{Name: "foo"}})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	changedHardware := catalogue.AllHardware()
	g.Expect(changedHardware).To(gomega.HaveLen(totalHardware))

	changedHardware[0] = &v1alpha1.Hardware{ObjectMeta: metav1.ObjectMeta{Name: "bar"}}

	unchangedHardware := catalogue.AllHardware()
	g.Expect(unchangedHardware).ToNot(gomega.Equal(changedHardware))
}

func TestHardwareValidator_Defaults(t *testing.T) {
	g := gomega.NewWithT(t)
	h := &v1alpha1.Hardware{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}

	validator := hardware.NewHardwareValidator()

	g.Expect(validator.Validate(h)).To(gomega.Succeed())
}

func TestHardwareValidator_InvalidName(t *testing.T) {
	g := gomega.NewWithT(t)
	validator := hardware.NewHardwareValidator()

	// Name not set.
	h := &v1alpha1.Hardware{}
	g.Expect(validator.Validate(h)).To(gomega.HaveOccurred())

	// Name not RFC1123 compliant
	h.Name = "\\"
	g.Expect(validator.Validate(h)).To(gomega.HaveOccurred())
}

func TestValidateHardware_WithValidHardware(t *testing.T) {
	g := gomega.NewWithT(t)
	validator := hardware.NewHardwareValidator()
	catalogue := hardware.NewHardwareCatalogue()

	g.Expect(catalogue.InsertHardware(NewValidHardware())).To(gomega.Succeed())
	g.Expect(hardware.ValidateCataloguedHardware(validator, catalogue)).To(gomega.Succeed())
}

func TestValidateHardware_WithInvalidHardware(t *testing.T) {
	g := gomega.NewWithT(t)
	validator := hardware.NewHardwareValidator()
	catalogue := hardware.NewHardwareCatalogue()

	h := NewValidHardware()
	h.Name = ""
	g.Expect(catalogue.InsertHardware(h)).To(gomega.Succeed())
	g.Expect(hardware.ValidateCataloguedHardware(validator, catalogue)).ToNot(gomega.Succeed())
}

func NewValidHardware() *v1alpha1.Hardware {
	return &v1alpha1.Hardware{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hardware",
		},
		Spec: v1alpha1.HardwareSpec{
			Metadata: &v1alpha1.HardwareMetadata{
				State: hardware.HardwareProvisioningState,
			},
		},
	}
}
