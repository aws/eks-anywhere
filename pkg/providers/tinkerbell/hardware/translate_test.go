package hardware_test

import (
	"errors"
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware/mocks"
)

func TestTranslateReadsAndWrites(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	reader := mocks.NewMockMachineReader(ctrl)
	writer := mocks.NewMockMachineWriter(ctrl)
	validator := mocks.NewMockMachineValidator(ctrl)

	machine := hardware.Machine{
		Hostname: "foot-bar",
	}

	var receivedMachine hardware.Machine
	reader.EXPECT().Read().Return(machine, (error)(nil))
	validator.EXPECT().Validate(machine).Return((error)(nil))
	writer.EXPECT().
		Write(machine).
		Do(func(machine hardware.Machine) {
			receivedMachine = machine
		}).
		Return((error)(nil))

	err := hardware.Translate(reader, writer, validator)

	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(receivedMachine).To(gomega.BeEquivalentTo(machine))
}

func TestTranslateWithReadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	reader := mocks.NewMockMachineReader(ctrl)
	writer := mocks.NewMockMachineWriter(ctrl)
	validator := mocks.NewMockMachineValidator(ctrl)

	expect := errors.New("luck-number-10")

	reader.EXPECT().Read().Return(hardware.Machine{}, expect)

	err := hardware.Translate(reader, writer, validator)

	g.Expect(err.Error()).To(gomega.ContainSubstring(expect.Error()))
}

func TestTranslateWithWriteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	reader := mocks.NewMockMachineReader(ctrl)
	writer := mocks.NewMockMachineWriter(ctrl)
	validator := mocks.NewMockMachineValidator(ctrl)

	machine := hardware.Machine{Hostname: "lucky-number-10"}
	expect := errors.New("luck-number-10")

	reader.EXPECT().Read().Return(machine, (error)(nil))
	validator.EXPECT().Validate(machine).Return((error)(nil))
	writer.EXPECT().Write(machine).Return(expect)

	err := hardware.Translate(reader, writer, validator)

	g.Expect(err.Error()).To(gomega.ContainSubstring(expect.Error()))
}

func TestTranslateReturnsEOFWhenReaderEOFs(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	reader := mocks.NewMockMachineReader(ctrl)
	writer := mocks.NewMockMachineWriter(ctrl)
	validator := mocks.NewMockMachineValidator(ctrl)

	reader.EXPECT().Read().Return(hardware.Machine{}, io.EOF)

	err := hardware.Translate(reader, writer, validator)

	g.Expect(err).To(gomega.BeEquivalentTo(io.EOF))
}

func TestTranslateWithValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	reader := mocks.NewMockMachineReader(ctrl)
	writer := mocks.NewMockMachineWriter(ctrl)
	validator := mocks.NewMockMachineValidator(ctrl)

	expect := errors.New("validation error")
	reader.EXPECT().Read().Return(hardware.Machine{}, (error)(nil))
	validator.EXPECT().Validate(hardware.Machine{}).Return(expect)

	err := hardware.Translate(reader, writer, validator)

	g.Expect(err.Error()).To(gomega.ContainSubstring(expect.Error()))
}

func TestTranslateAllReadsAndWritesMaskingEOF(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	reader := mocks.NewMockMachineReader(ctrl)
	writer := mocks.NewMockMachineWriter(ctrl)
	validator := mocks.NewMockMachineValidator(ctrl)

	machine := hardware.Machine{Hostname: "lucky-number-10"}

	// use readCount to track how many times the Read() call has been made. On
	// the second call we return io.EOF.
	var readCount int
	reader.EXPECT().
		Read().
		Times(2).
		DoAndReturn(func() (hardware.Machine, error) {
			if readCount == 1 {
				return hardware.Machine{}, io.EOF
			}

			readCount++
			return machine, nil
		})

	validator.EXPECT().Validate(machine).Return((error)(nil))

	// we only expect Write() to bec alled once because the io.EOF shouldn't result in
	// a write.
	writer.EXPECT().Write(machine).Times(1).Return((error)(nil))

	err := hardware.TranslateAll(reader, writer, validator)

	g.Expect(err).ToNot(gomega.HaveOccurred())
}

func TestTranslateAllWithReadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	reader := mocks.NewMockMachineReader(ctrl)
	writer := mocks.NewMockMachineWriter(ctrl)
	validator := mocks.NewMockMachineValidator(ctrl)

	expect := errors.New("luck-number-10")

	reader.EXPECT().Read().Return(hardware.Machine{}, expect)

	err := hardware.TranslateAll(reader, writer, validator)

	g.Expect(err.Error()).To(gomega.ContainSubstring(expect.Error()))
}

func TestTranslateAllWithWriteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	reader := mocks.NewMockMachineReader(ctrl)
	writer := mocks.NewMockMachineWriter(ctrl)
	validator := mocks.NewMockMachineValidator(ctrl)

	machine := hardware.Machine{Hostname: "lucky-number-10"}
	expect := errors.New("luck-number-10")

	reader.EXPECT().Read().Return(machine, (error)(nil))
	validator.EXPECT().Validate(machine).Return((error)(nil))
	writer.EXPECT().Write(machine).Return(expect)

	err := hardware.TranslateAll(reader, writer, validator)

	g.Expect(err.Error()).To(gomega.ContainSubstring(expect.Error()))
}

func TestTranslateAllWithValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	reader := mocks.NewMockMachineReader(ctrl)
	writer := mocks.NewMockMachineWriter(ctrl)
	validator := mocks.NewMockMachineValidator(ctrl)

	expect := errors.New("validation error")
	reader.EXPECT().Read().Return(hardware.Machine{}, (error)(nil))
	validator.EXPECT().Validate(hardware.Machine{}).Return(expect)

	err := hardware.TranslateAll(reader, writer, validator)

	g.Expect(err.Error()).To(gomega.ContainSubstring(expect.Error()))
}

func TestTranslateReadsAndWritesWithModifier(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	reader := mocks.NewMockMachineReader(ctrl)
	writer := mocks.NewMockMachineWriter(ctrl)
	validator := mocks.NewMockMachineValidator(ctrl)

	machine := hardware.Machine{
		Hostname: "foot-bar",
	}

	validateMachine := hardware.Machine{
		Hostname:    "foot-bar",
		BMCUsername: "rufio",
		BMCPassword: "secret",
		ConsumerURL: "consumerURL",
	}

	mods := []func(hardware.Machine) hardware.Machine{}
	mods = append(mods, func(m hardware.Machine) hardware.Machine {
		m.BMCPassword = "secret"
		if m.BMCUsername == "" {
			// We update the BMCUsername so that validations pass.
			// When the secret != "" the webhook functionality will be used.
			// With this the BMCUsername is not relevant for any BMC machines,
			// hence the statically define name.
			m.BMCUsername = "rufio"
		}
		return m
	})
	mods = append(mods, func(m hardware.Machine) hardware.Machine {
		m.ConsumerURL = "consumerURL"
		return m
	})

	var receivedMachine hardware.Machine
	reader.EXPECT().Read().Return(machine, (error)(nil))
	validator.EXPECT().Validate(validateMachine).Return((error)(nil))
	writer.EXPECT().
		Write(validateMachine).
		Do(func(machine hardware.Machine) {
			receivedMachine = validateMachine
		}).
		Return((error)(nil))

	err := hardware.Translate(reader, writer, validator, mods...)

	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(receivedMachine).To(gomega.BeEquivalentTo(validateMachine))
}
