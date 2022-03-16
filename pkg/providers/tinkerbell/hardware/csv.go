package hardware

import (
	stdcsv "encoding/csv"
	"io"

	csv "github.com/gocarina/gocsv"
)

// CsvReader reads a CSV file and provides Machine instances. It satisfies the MachineReader interface.
type CsvReader struct {
	reader *csv.Unmarshaller
}

// NewCsvReader returns a new CsvReader instance that consumes csv data from r. r should return io.EOF when no more
// records are available.
func NewCsvReader(r io.Reader) (CsvReader, error) {
	stdreader := stdcsv.NewReader(r)

	reader, err := csv.NewUnmarshaller(stdreader, Machine{})
	if err != nil {
		return CsvReader{}, err
	}

	return CsvReader{reader: reader}, nil
}

// Read reads a single entry from the CSV data source and returns a new Machine representation.
func (cr CsvReader) Read() (Machine, error) {
	machine, err := cr.reader.Read()
	if err != nil {
		return Machine{}, err
	}
	return machine.(Machine), nil
}
