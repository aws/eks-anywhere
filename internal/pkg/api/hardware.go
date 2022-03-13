package api

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gocarina/gocsv"
)

const (
	HardwareVendorDell       = "dell"
	HardwareVendorHP         = "hp"
	HardwareVendorSuperMicro = "supermicro"
	HardwareVendorAgnostic   = "agnostic"
)

type Hardware struct {
	Id           string `csv:"guid"`
	IpAddress    string `csv:"ip_address"`
	Gateway      string `csv:"gateway"`
	Nameservers  string `csv:"nameservers"`
	Netmask      string `csv:"netmask"`
	MacAddress   string `csv:"mac"`
	Hostname     string `csv:"hostname"`
	Vendor       string `csv:"vendor"`
	BmcIpAddress string `csv:"bmc_ip"`
	BmcUsername  string `csv:"bmc_username"`
	BmcPassword  string `csv:"bmc_password"`
}

func NewHardwareSlice(csvFile string) ([]*Hardware, error) {
	hardwareFile, err := os.OpenFile(csvFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create hardware slice from hardware csv file: %v", err)
	}

	defer hardwareFile.Close()

	hardware := []*Hardware{}

	if err := gocsv.UnmarshalFile(hardwareFile, &hardware); err != nil {
		return nil, fmt.Errorf("failed to create hardware slice from hardware csv file: %v", err)
	}

	return hardware, nil
}

func NewHardwareMap(csvFile string) (map[string]*Hardware, error) {
	slice, err := NewHardwareSlice(csvFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create hardware map from hardware csv file: %v", err)
	}
	return HardwareSliceToMap(slice), nil
}

func HardwareSliceToMap(slice []*Hardware) map[string]*Hardware {
	hardwareMap := make(map[string]*Hardware)

	for _, h := range slice {
		if _, exists := hardwareMap[h.Id]; !exists {
			hardwareMap[h.Id] = h
		}
	}

	return hardwareMap
}

func HardwareSliceToCSV(hardware []*Hardware, csvFile string) error {
	csvdir := filepath.Dir(csvFile)
	err := os.MkdirAll(csvdir, 0755)
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

func HardwareMapToCSV(hardware map[string]*Hardware, csvFile string) error {
	slice := HardwareMapToSlice(hardware)
	return HardwareSliceToCSV(slice, csvFile)
}
