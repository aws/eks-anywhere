package hardware

import (
	"encoding/csv"
	"fmt"
	"os"
)

type CsvParser struct {
	HeadersIndex
	csvFile *os.File
	reader  *csv.Reader
}

type HeadersIndex struct {
	HostnameIndex    int
	IpAddressIndex   int
	GatewayIndex     int
	NetmaskIndex     int
	MacIndex         int
	NameServerIndex  int
	VendorIndex      int
	BmcIpIndex       int
	BmcUsernameIndex int
	BmcPasswordIndex int
}

func NewCsvParser(filepath string) (*CsvParser, error) {
	csvFile, err := os.Open(filepath)
	if err != nil {
		csvFile.Close()
		return nil, fmt.Errorf("error initializing CsvParser: error opening file %s: %v", filepath, err)
	}

	csvParser := &CsvParser{
		HeadersIndex: HeadersIndex{},
		csvFile:      csvFile,
		reader:       csv.NewReader(csvFile),
	}

	if err := csvParser.parseHeaders(); err != nil {
		csvFile.Close()
		return nil, fmt.Errorf("error initializing CsvParser: %v", err)
	}

	return csvParser, nil
}

func (c *CsvParser) Close() {
	c.csvFile.Close()
}

func (c *CsvParser) Read() ([]string, error) {
	return c.reader.Read()
}

func (c *CsvParser) parseHeaders() error {
	headers, err := c.Read()
	if err != nil {
		return fmt.Errorf("error parsing CSV headers: %v", err)
	}

	var ok bool
	m := make(map[string]int)
	for i, header := range headers {
		m[header] = i
	}

	if c.HostnameIndex, ok = m[hostname]; !ok {
		return fmt.Errorf("error finding header %s", hostname)
	}

	if c.IpAddressIndex, ok = m[ipAddress]; !ok {
		return fmt.Errorf("error finding header %s", ipAddress)
	}

	if c.GatewayIndex, ok = m[gateway]; !ok {
		return fmt.Errorf("error finding header %s", gateway)
	}

	if c.NetmaskIndex, ok = m[netmask]; !ok {
		return fmt.Errorf("error finding header %s", netmask)
	}

	if c.MacIndex, ok = m[mac]; !ok {
		return fmt.Errorf("error finding header %s", mac)
	}

	if c.NameServerIndex, ok = m[nameservers]; !ok {
		return fmt.Errorf("error finding header %s", nameservers)
	}

	if c.VendorIndex, ok = m[vendor]; !ok {
		return fmt.Errorf("error finding header %s", vendor)
	}

	if c.BmcIpIndex, ok = m[bmcIp]; !ok {
		return fmt.Errorf("error finding header %s", bmcIp)
	}

	if c.BmcUsernameIndex, ok = m[bmcUsername]; !ok {
		return fmt.Errorf("error finding header %s", bmcUsername)
	}

	if c.BmcPasswordIndex, ok = m[bmcPassword]; !ok {
		return fmt.Errorf("error finding header %s", bmcPassword)
	}

	return nil
}
