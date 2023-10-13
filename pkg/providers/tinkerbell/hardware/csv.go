package hardware

import (
	"bufio"
	"bytes"
	stdcsv "encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	csv "github.com/gocarina/gocsv"

	unstructuredutil "github.com/aws/eks-anywhere/pkg/utils/unstructured"
)

// CSVReader reads a CSV file and provides Machine instances. It satisfies the MachineReader interface. The ID field of
// the Machine is optional in the CSV. If unspecified, CSVReader will generate a UUID and apply it to the machine.
type CSVReader struct {
	reader *csv.Unmarshaller
	// BMCOptions used in a Machine that do not have a corresponding column in the CSV.
	BMCOptions *BMCOptions
}

// NewCSVReader returns a new CSVReader instance that consumes csv data from r. r should return io.EOF when no more
// records are available.
func NewCSVReader(r io.Reader, opts *BMCOptions) (CSVReader, error) {
	stdreader := stdcsv.NewReader(r)

	reader, err := csv.NewUnmarshaller(stdreader, Machine{})
	if err != nil {
		return CSVReader{}, err
	}

	if err := ensureRequiredColumnsInCSV(reader.MismatchedStructFields); err != nil {
		return CSVReader{}, err
	}

	return CSVReader{reader: reader, BMCOptions: opts}, nil
}

// Read reads a single entry from the CSV data source and returns a new Machine representation.
func (cr CSVReader) Read() (Machine, error) {
	machine, err := cr.reader.Read()
	if err != nil {
		return Machine{}, err
	}
	m := machine.(Machine)
	if cr.BMCOptions != nil {
		m.BMCOptions = cr.BMCOptions
	}

	return m, nil
}

// NewNormalizedCSVReaderFromFile creates a MachineReader instance backed by a CSVReader reading from path
// that applies default normalizations to machines.
func NewNormalizedCSVReaderFromFile(path string, opts *BMCOptions) (MachineReader, error) {
	fh, err := os.Open(path)
	if err != nil {
		return CSVReader{}, err
	}

	reader, err := NewCSVReader(bufio.NewReader(fh), opts)
	if err != nil {
		return nil, err
	}

	return NewNormalizer(reader), nil
}

// requiredColumns matches the csv tags on the Machine struct. These must remain in sync with
// the struct. We may consider an alternative that uses reflection to interpret whether a field
// is required in the future.
var requiredColumns = map[string]struct{}{
	"hostname":    {},
	"ip_address":  {},
	"netmask":     {},
	"gateway":     {},
	"nameservers": {},
	"mac":         {},
	"disk":        {},
	"labels":      {},
}

func ensureRequiredColumnsInCSV(unmatched []string) error {
	var intersection []string
	for _, column := range unmatched {
		if _, ok := requiredColumns[column]; ok {
			intersection = append(intersection, column)
		}
	}

	if len(intersection) > 0 {
		return fmt.Errorf("missing required columns in csv: %v", strings.Join(intersection, ", "))
	}

	return nil
}

// BuildHardwareYAML builds a hardware yaml from the csv at the provided path.
func BuildHardwareYAML(path string, opts *BMCOptions) ([]byte, error) {
	reader, err := NewNormalizedCSVReaderFromFile(path, opts)
	if err != nil {
		return nil, fmt.Errorf("reading csv: %v", err)
	}

	var b bytes.Buffer
	writer := NewTinkerbellManifestYAML(&b)

	validator := NewDefaultMachineValidator()

	err = TranslateAll(reader, writer, validator)
	if err != nil {
		return nil, fmt.Errorf("generating hardware yaml: %v", err)
	}

	return unstructuredutil.StripNull(b.Bytes())
}
