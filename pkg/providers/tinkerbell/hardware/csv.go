package hardware

import (
	"bufio"
	stdcsv "encoding/csv"
	"io"
	"os"

	csv "github.com/gocarina/gocsv"
)

// CSVReader reads a CSV file and provides Machine instances. It satisfies the MachineReader interface. The ID field of
// the Machine is optional in the CSV. If unspecified, CSVReader will generate a UUID and apply it to the machine.
type CSVReader struct {
	reader *csv.Unmarshaller
}

// NewCSVReader returns a new CSVReader instance that consumes csv data from r. r should return io.EOF when no more
// records are available.
func NewCSVReader(r io.Reader) (CSVReader, error) {
	stdreader := stdcsv.NewReader(r)

	reader, err := csv.NewUnmarshaller(stdreader, Machine{})
	if err != nil {
		return CSVReader{}, err
	}

	return CSVReader{reader: reader}, nil
}

// Read reads a single entry from the CSV data source and returns a new Machine representation.
func (cr CSVReader) Read() (Machine, error) {
	machine, err := cr.reader.Read()
	if err != nil {
		return Machine{}, err
	}
	return machine.(Machine), nil
}

// NewNormalizedCSVReaderFromFile creates a MachineReader instance backed by a CSVReader reading from path
// that applies default normalizations to machines.
func NewNormalizedCSVReaderFromFile(path string) (MachineReader, error) {
	fh, err := os.Open(path)
	if err != nil {
		return CSVReader{}, err
	}

	reader, err := NewCSVReader(bufio.NewReader(fh))
	if err != nil {
		return nil, err
	}

	return NewNormalizer(reader), nil
}
