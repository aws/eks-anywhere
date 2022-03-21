package api

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gocarina/gocsv"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/pbnj"
)

const (
	HardwareVendorDell        = "dell"
	HardwareVendorHP          = "hp"
	HardwareVendorSuperMicro  = "supermicro"
	HardwareVendorUnspecified = "unspecified"
)

// Alias for backwards compatibility.
type Hardware = hardware.Machine

func NewBmcSecretConfig(h *Hardware) pbnj.BmcSecretConfig {
	return pbnj.BmcSecretConfig{
		Host:     h.BmcIpAddress,
		Username: h.BmcUsername,
		Password: h.BmcPassword,
		Vendor:   h.BmcVendor,
	}
}

func NewHardwareSlice(r io.Reader) ([]*Hardware, error) {
	hardware := []*Hardware{}

	if err := gocsv.Unmarshal(r, &hardware); err != nil {
		return nil, fmt.Errorf("failed to create hardware slice from reader: %v", err)
	}

	return hardware, nil
}

func NewHardwareSliceFromFile(file string) ([]*Hardware, error) {
	hardwareFile, err := os.OpenFile(file, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create hardware slice from hardware file: %v", err)
	}
	return NewHardwareSlice(hardwareFile)
}

func NewHardwareMapFromFile(file string) (map[string]*Hardware, error) {
	slice, err := NewHardwareSliceFromFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create hardware map from hardware file: %v", err)
	}
	return HardwareSliceToMap(slice), nil
}

// converts a hardware slice to a map. The first instance of the slice is used in case slice contains duplicates
func HardwareSliceToMap(slice []*Hardware) map[string]*Hardware {
	hardwareMap := make(map[string]*Hardware)

	for _, h := range slice {
		if _, exists := hardwareMap[h.Id]; !exists {
			hardwareMap[h.Id] = h
		}
	}

	return hardwareMap
}

func WriteHardwareSliceToCSV(hardware []*Hardware, csvFile string) error {
	csvdir := filepath.Dir(csvFile)
	err := os.MkdirAll(csvdir, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create hardware csv file from slice: %v", err)
	}

	hardwareFile, err := os.OpenFile(csvFile, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create hardware csv file from slice: %v", err)
	}

	defer hardwareFile.Close()

	if err := gocsv.MarshalFile(&hardware, hardwareFile); err != nil {
		return fmt.Errorf("failed to create hardware csv file from slice: %v", err)
	}

	return nil
}

func HardwareMapToSlice(hardware map[string]*Hardware) []*Hardware {
	harwareSlice := []*Hardware{}
	for _, value := range hardware {
		harwareSlice = append(harwareSlice, value)
	}
	return harwareSlice
}

func WriteHardwareMapToCSV(hardware map[string]*Hardware, csvFile string) error {
	slice := HardwareMapToSlice(hardware)
	return WriteHardwareSliceToCSV(slice, csvFile)
}
