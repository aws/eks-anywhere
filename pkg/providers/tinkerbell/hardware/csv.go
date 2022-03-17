package hardware

import (
	stdcsv "encoding/csv"
	"io"

	csv "github.com/gocarina/gocsv"
	"github.com/google/uuid"
)

// CsvReader reads a CSV file and provides Machine instances. It satisfies the MachineReader interface. The Id field of
// the Machine is optional in the CSV. If unspecified, CsvReader will generate a UUID and apply it to the machine.
type CsvReader struct {
	reader        *csv.Unmarshaller
	uuidGenerator func() string
}

// NewCsvReader returns a new CsvReader instance that consumes csv data from r. r should return io.EOF when no more
// records are available.
func NewCsvReader(r io.Reader) (CsvReader, error) {
	stdreader := stdcsv.NewReader(r)

	reader, err := csv.NewUnmarshaller(stdreader, Machine{})
	if err != nil {
		return CsvReader{}, err
	}

	return CsvReader{reader: reader, uuidGenerator: uuid.NewString}, nil
}

// NewCsvReaderWithUUIDGenerator returns a new CsvReader instance as defined in NewCsvReader with its internal
// UUID generator configured as generator.
func NewCsvReaderWithUUIDGenerator(r io.Reader, generator func() string) (CsvReader, error) {
	reader, err := NewCsvReader(r)
	if err != nil {
		return CsvReader{}, err
	}

	reader.uuidGenerator = generator
	return reader, nil
}

// Read reads a single entry from the CSV data source and returns a new Machine representation.
func (cr CsvReader) Read() (Machine, error) {
	machine, err := cr.reader.Read()
	if err != nil {
		return Machine{}, err
	}
	m := machine.(Machine)
	if m.Id == "" {
		m.Id = cr.uuidGenerator()
	}
	return m, nil
}
