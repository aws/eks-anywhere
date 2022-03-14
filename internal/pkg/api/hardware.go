package api

import (
	"fmt"
	"os"

	"github.com/gocarina/gocsv"
)

const (
	Dell       = "dell"
	HP         = "hp"
	SuperMicro = "supermicro"
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
	hardwareFile, err := os.OpenFile(csvFile, os.O_RDWR|os.O_CREATE, os.ModePerm)
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
