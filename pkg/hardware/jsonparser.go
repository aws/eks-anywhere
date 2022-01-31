package hardware

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/eks-anywhere/pkg/filewriter"
)

const (
	hostname      = "hostname"
	ipAddress     = "ip_address"
	gateway       = "gateway"
	netmask       = "netmask"
	mac           = "mac"
	nameservers   = "nameservers"
	vendor        = "vendor"
	bmcIp         = "bmc_ip"
	bmcUsername   = "bmc_username"
	bmcPassword   = "bmc_password"
	eksaNamespace = "eksa-system"
	jsonPath      = "hardware-manifests/json"
)

type JsonParser struct {
	jsonWriter filewriter.FileWriter
}

func NewJsonParser() (*JsonParser, error) {
	filewriter, err := filewriter.NewWriter(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("error initializing JsonParser: %v", err)
	}
	return &JsonParser{
		jsonWriter: filewriter,
	}, nil
}

func (j *JsonParser) Write(filename string, data []byte) error {
	_, err := j.jsonWriter.Write(filename, data, filewriter.PersistentFile)
	return err
}

func (j *JsonParser) CleanUp() {
	j.jsonWriter.CleanUpTemp()
}

func (j *JsonParser) GetHardwareJson(id, hostname, ipAddress, gateway, netmask, mac, nameserver string) ([]byte, error) {
	nameservers := strings.Split(nameserver, "|")
	hardware := &Hardware{
		ID: id,
		Metadata: Metadata{
			Facility: Facility{
				FacilityCode:    "onprem",
				PlanSlug:        "c2.medium.x86",
				PlanVersionSlug: "",
			},
			Instance: Instance{
				ID:       id,
				Hostname: hostname,
				Storage: Storage{
					Disks: []Disk{
						{
							Device: "/dev/sda",
						},
					},
				},
			},
			State: "provisioning",
		},
		Network: Network{
			Interfaces: []Interface{
				{
					DHCP: DHCP{
						Arch: "x86_64",
						IP: IP{
							Address: ipAddress,
							Gateway: gateway,
							Netmask: netmask,
						},
						Mac:         mac,
						Nameservers: nameservers,
						UEFI:        true,
					},
					Netboot: Netboot{
						AllowPXE:      true,
						AllowWorkflow: true,
					},
				},
			},
		},
	}

	return json.MarshalIndent(hardware, "", "  ")
}
