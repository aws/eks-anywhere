package hardware

import (
	"bufio"
	stdcsv "encoding/csv"
	"io"
	"os"

	csv "github.com/gocarina/gocsv"
	"github.com/google/uuid"
)

// CSVReader reads a CSV file and provides Machine instances. It satisfies the MachineReader interface. The ID field of
// the Machine is optional in the CSV. If unspecified, CSVReader will generate a UUID and apply it to the machine.
type CSVReader struct {
	reader        *csv.Unmarshaller
	uuidGenerator func() string
}

// NewCSVReader returns a new CSVReader instance that consumes csv data from r. r should return io.EOF when no more
// records are available.
func NewCSVReader(r io.Reader) (CSVReader, error) {
	stdreader := stdcsv.NewReader(r)

	reader, err := csv.NewUnmarshaller(stdreader, Machine{})
	if err != nil {
		return CSVReader{}, err
	}

	return CSVReader{reader: reader, uuidGenerator: uuid.NewString}, nil
}

// NewCSVReaderFromFile creates a CSVReader instance that reads from path.
func NewCSVReaderFromFile(path string) (CSVReader, error) {
	fh, err := os.Open(path)
	if err != nil {
		return CSVReader{}, err
	}

	return NewCSVReader(bufio.NewReader(fh))
}

// NewCSVReaderWithUUIDGenerator returns a new CSVReader instance as defined in NewCSVReader with its internal
// UUID generator configured as generator.
func NewCSVReaderWithUUIDGenerator(r io.Reader, generator func() string) (CSVReader, error) {
	reader, err := NewCSVReader(r)
	if err != nil {
		return CSVReader{}, err
	}

	reader.uuidGenerator = generator
	return reader, nil
}

// Read reads a single entry from the CSV data source and returns a new Machine representation.
func (cr CSVReader) Read() (Machine, error) {
	machine, err := cr.reader.Read()
	if err != nil {
		return Machine{}, err
	}
	m := machine.(Machine)
	if m.ID == "" {
		m.ID = cr.uuidGenerator()
	}
	return m, nil
}
