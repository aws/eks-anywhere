package hardware_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware/mocks"
)

/*

MachineValidationDecorator
	- Register
	- Read

UniqueHostnames
UniqueIds
*/

func TestMachineValidationDecoratorValidationsAppliedToRead(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	expect := hardware.Machine{
		Id:       "qux-rules-all",
		Hostname: "youve-been-food",
	}

	reader := mocks.NewMockMachineReader(ctrl)
	reader.EXPECT().Read().Return(expect, (error)(nil))

	// check is set by assertion when its called and allows us to validate
	// registered assertions are infact called by the validation decorator.
	var check bool
	assertion := func(m hardware.Machine) error {
		check = true
		return nil
	}

	validator := hardware.NewMachineValidationDecorator(reader)
	validator.Register(assertion)

	machine, err := validator.Read()

	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(check).To(gomega.BeTrue())
	g.Expect(machine).To(gomega.BeEquivalentTo(expect))
}

func TestMachineValidationDecoratorErrorsWhenAssertionErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	reader := mocks.NewMockMachineReader(ctrl)
	reader.EXPECT().Read().Return(hardware.Machine{}, (error)(nil))

	// check is set by assertion when its called and allows us to validate
	// registered assertions are infact called by the validation decorator.
	expect := errors.New("something went wrong")
	assertion := func(hardware.Machine) error {
		return expect
	}

	validator := hardware.NewMachineValidationDecorator(reader)
	validator.Register(assertion)

	_, err := validator.Read()

	g.Expect(err).To(gomega.BeEquivalentTo(expect))
}

func TestMachineValidationDecoratorErrorsWhenReaderErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	reader := mocks.NewMockMachineReader(ctrl)

	expect := errors.New("something went wrong")
	reader.EXPECT().Read().Return(hardware.Machine{}, expect)

	validator := hardware.NewMachineValidationDecorator(reader)

	_, err := validator.Read()

	g.Expect(err).To(gomega.BeEquivalentTo(expect))
}

func TestUniqueIds(t *testing.T) {
	g := gomega.NewWithT(t)

	assertion := hardware.UniqueIds()

	err := assertion(hardware.Machine{Id: "foo"})

	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = assertion(hardware.Machine{Id: "bar"})

	g.Expect(err).ToNot(gomega.HaveOccurred())
}

func TestUniqueIdsWithDupes(t *testing.T) {
	g := gomega.NewWithT(t)

	assertion := hardware.UniqueIds()

	err := assertion(hardware.Machine{Id: "foo"})

	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = assertion(hardware.Machine{Id: "foo"})

	g.Expect(err).To(gomega.HaveOccurred())
}

func TestUniqueHostnames(t *testing.T) {
	g := gomega.NewWithT(t)

	assertion := hardware.UniqueHostnames()

	err := assertion(hardware.Machine{Hostname: "foo"})

	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = assertion(hardware.Machine{Hostname: "bar"})

	g.Expect(err).ToNot(gomega.HaveOccurred())
}

func TestUniqueHostnamesWithDupes(t *testing.T) {
	g := gomega.NewWithT(t)

	assertion := hardware.UniqueHostnames()

	err := assertion(hardware.Machine{Hostname: "foo"})

	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = assertion(hardware.Machine{Hostname: "foo"})

	g.Expect(err).To(gomega.HaveOccurred())
}
