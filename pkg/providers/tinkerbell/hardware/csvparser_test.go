package hardware_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

const (
	valid_testdata                  = "testdata/valid_hardware.csv"
	invalid_testdata_empty_file     = "testdata/invalid_hardware_empty_file.csv"
	invalid_testdata_missing_file   = "testdata/invalid_hardware_missing_file.go"
	invalid_testdata_missing_header = "testdata/invalid_hardware_missing_header.csv"
)

func TestNewCsvParserSuccess(t *testing.T) {
	hi := hardware.HeadersIndex{
		HostnameIndex:    5,
		IpAddressIndex:   0,
		GatewayIndex:     1,
		NetmaskIndex:     3,
		MacIndex:         4,
		NameServerIndex:  2,
		VendorIndex:      6,
		BmcIpIndex:       7,
		BmcUsernameIndex: 8,
		BmcPasswordIndex: 9,
	}

	csv, err := hardware.NewCsvParser(valid_testdata)
	if err != nil {
		t.Fatalf("hardware.NewCsvParser() error = %v, want nil", err)
	}

	defer csv.Close()

	if !reflect.DeepEqual(csv.HeadersIndex, hi) {
		t.Fatalf("CsvParser.HeadersIndex = %#v, want %#v", csv.HeadersIndex, hi)
	}
}

func TestNewCsvParserFailureMissingHeader(t *testing.T) {
	_, err := hardware.NewCsvParser(invalid_testdata_missing_header)
	expectedErr := "error initializing CsvParser: error finding header bmc_password"
	if err == nil || err.Error() != expectedErr {
		t.Fatalf("hardware.NewCsvParser() error = %v, expected = %s", err, expectedErr)
	}
}

func TestNewCsvParserFailureMissingFile(t *testing.T) {
	_, err := hardware.NewCsvParser(invalid_testdata_missing_file)
	expectedErr := fmt.Sprintf("error initializing CsvParser: error opening file %s: open %s: no such file or directory", invalid_testdata_missing_file, invalid_testdata_missing_file)
	if err == nil || err.Error() != expectedErr {
		t.Fatalf("hardware.NewCsvParser() error = %v, expected = %s", err, expectedErr)
	}
}

func TestNewCsvParserFailureEmptyFile(t *testing.T) {
	_, err := hardware.NewCsvParser(invalid_testdata_empty_file)
	expectedErr := "error initializing CsvParser: error parsing CSV headers: EOF"
	if err == nil || err.Error() != expectedErr {
		t.Fatalf("hardware.NewCsvParser() error = %v, expected = %s", err, expectedErr)
	}
}
