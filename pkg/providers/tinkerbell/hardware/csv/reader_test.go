package csv_test

import (
	"bytes"
	stdcsv "encoding/csv"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware/csv"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware/csv/mocks"
)

func TestReaderCreateNewInstance(t *testing.T) {
	g := gomega.NewWithT(t)

	buf := NewBufferedCSV()
	err := buf.Write(csv.Headers())
	g.Expect(err).ToNot(gomega.HaveOccurred())

	_, err = csv.NewReader(buf)

	g.Expect(err).ToNot(gomega.HaveOccurred())
}

func TestReaderCreateNewInstanceWithMissingHeaders(t *testing.T) {
	g := gomega.NewWithT(t)

	// Build cases where 1 header is removed at a time rusulting in len(Headers()) cases.
	cases := make(map[string][]string)
	for i := 0; i < len(csv.Headers()); i++ {
		headers := csv.Headers()
		name := headers[i]
		cases[name] = append(headers[:i], headers[i+1:]...)
	}

	for name, record := range cases {
		t.Run(name, func(t *testing.T) {
			buf := NewBufferedCSV()
			err := buf.Write(record)
			g.Expect(err).ToNot(gomega.HaveOccurred())
			_, err = csv.NewReader(buf)
			g.Expect(err).To(gomega.HaveOccurred())
		})
	}
}

func TestReaderReadsMachines(t *testing.T) {
	g := gomega.NewWithT(t)

	buf := NewBufferedCSV()
	err := buf.Write(csv.Headers())
	g.Expect(err).ToNot(gomega.HaveOccurred())

	const id = "foo-bar"
	expect := hardware.Machine{
		Id:       id,
		Hostname: "domain.com",
		Network: hardware.Network{
			Ip:          "1.1.1.2",
			Netmask:     "255.255.0.0",
			Gateway:     "1.1.1.1",
			Mac:         "00:00:00:00:00:00",
			NameServers: []string{"foo.com", "bar.com"},
		},
		Bmc: &hardware.Bmc{
			Ip:       "1.1.1.3",
			Username: "foo",
			Password: "bar",
			Vendor:   "linux",
		},
	}
	err = WriteMachine(expect, buf)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	reader, err := csv.NewReader(buf)
	csv.WithIDGenerator(reader, func() string {
		return id
	})

	g.Expect(err).ToNot(gomega.HaveOccurred())

	machine, err := reader.Read()

	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(machine).To(gomega.BeEquivalentTo(expect))
}

func TestReaderRecordReaderErrorsOnConstruction(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	recordReader := mocks.NewMockRecordReader(ctrl)

	expect := errors.New("foo something bar")
	recordReader.EXPECT().Read().Return(([]string)(nil), expect)

	_, err := csv.NewReader(recordReader)

	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring(expect.Error()))
}

func TestReaderRecordReaderErrorsOnRead(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewWithT(t)

	recordReader := mocks.NewMockRecordReader(ctrl)

	expect := errors.New("foo something bar")

	// track the count so we can return something differnet on the second call.
	var count int
	recordReader.EXPECT().Read().Times(2).DoAndReturn(func() ([]string, error) {
		if count == 1 {
			return nil, expect
		}
		count += 1
		return csv.Headers(), nil
	})

	reader, err := csv.NewReader(recordReader)

	g.Expect(err).ToNot(gomega.HaveOccurred())

	_, err = reader.Read()

	g.Expect(err.Error()).To(gomega.ContainSubstring(expect.Error()))
}

func TestOpen(t *testing.T) {
	g := gomega.NewWithT(t)

	reader, err := csv.Open("./testdata/hardware.csv")

	g.Expect(err).ToNot(gomega.HaveOccurred())

	_, err = reader.Read()

	g.Expect(err).ToNot(gomega.HaveOccurred())
}

// WriteMachine applies machine in the same order as csv.Headers(). writer defines an interface with 1 method
// that accepts a record and meets the standard library csv.Writer API.
func WriteMachine(machine hardware.Machine, writer interface{ Write([]string) error }) error {
	return writer.Write([]string{
		machine.Hostname,
		machine.Network.Ip,
		machine.Network.Gateway,
		machine.Network.Netmask,
		machine.Network.Mac,
		csv.JoinNameServers(machine.Network.NameServers),
		machine.Bmc.Vendor,
		machine.Bmc.Ip,
		machine.Bmc.Username,
		machine.Bmc.Password,
	})
}

// BufferedCSV is an in-memory CSV that satisfies io.Reader and io.Writer.
type BufferedCSV struct {
	*stdcsv.Writer
	*stdcsv.Reader
}

func NewBufferedCSV() *BufferedCSV {
	var buf bytes.Buffer
	writer := stdcsv.NewWriter(&buf)
	reader := stdcsv.NewReader(&buf)

	return &BufferedCSV{
		Writer: writer,
		Reader: reader,
	}
}

// Write writes record to b using the underlying csv.Writer but immediately flushes. This
// ensures the in-memory buffer is always up-to-date.
func (b *BufferedCSV) Write(record []string) error {
	if err := b.Writer.Write(record); err != nil {
		return err
	}
	b.Flush()
	return nil
}
