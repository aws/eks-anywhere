package hardware_test

import (
	"bytes"
	stdcsv "encoding/csv"
	"errors"
	"testing"
	"testing/iotest"

	csv "github.com/gocarina/gocsv"
	"github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestCsvReaderReads(t *testing.T) {
	g := gomega.NewWithT(t)

	buf := NewBufferedCSV()

	expect := NewValidMachine()

	err := csv.MarshalCSV([]hardware.Machine{expect}, buf)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	reader, err := hardware.NewCsvReader(buf.Buffer)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	machine, err := reader.Read()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(machine).To(gomega.BeEquivalentTo(expect))
}

func TestCsvReaderReadsWithNoIdSpecified(t *testing.T) {
	g := gomega.NewWithT(t)

	buf := NewBufferedCSV()

	expect := NewValidMachine()
	expect.Id = ""

	err := csv.MarshalCSV([]hardware.Machine{expect}, buf)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	const uuid = "unique-id"
	reader, err := hardware.NewCsvReaderWithUUIDGenerator(buf.Buffer, func() string { return uuid })
	g.Expect(err).ToNot(gomega.HaveOccurred())

	machine, err := reader.Read()
	g.Expect(err).ToNot(gomega.HaveOccurred())

	expect.Id = uuid // patch the expected machine with the expected uuid
	g.Expect(machine).To(gomega.BeEquivalentTo(expect))
}

func TestNewCsvReaderWithIOReaderError(t *testing.T) {
	g := gomega.NewWithT(t)

	expect := errors.New("read err")

	_, err := hardware.NewCsvReader(iotest.ErrReader(expect))
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring(expect.Error()))
}

// BufferedCSV is an in-memory CSV that satisfies io.Reader and io.Writer.
type BufferedCSV struct {
	*bytes.Buffer
	*stdcsv.Writer
	*stdcsv.Reader
}

func NewBufferedCSV() *BufferedCSV {
	buf := &BufferedCSV{Buffer: &bytes.Buffer{}}
	buf.Writer = stdcsv.NewWriter(buf.Buffer)
	buf.Reader = stdcsv.NewReader(buf.Buffer)
	return buf
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
