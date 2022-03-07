package hardware

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tinkerbell/tink/protos/hardware"
	"github.com/tinkerbell/tink/protos/packet"

	"github.com/aws/eks-anywhere/pkg/filewriter"
)

const (
	guid          = "guid"
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

type Hardware struct {
	Id       string                     `json:"id"`
	Metadata *packet.Metadata           `json:"metadata"`
	Network  *hardware.Hardware_Network `json:"network"`
}

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
		Id: id,
		Metadata: &packet.Metadata{
			Facility: &packet.Metadata_Facility{
				FacilityCode: "onprem",
				PlanSlug:     "c2.medium.x86",
			},
			Instance: &packet.Metadata_Instance{
				Id:       id,
				Hostname: hostname,
				Storage: &packet.Metadata_Instance_Storage{
					Disks: []*packet.Metadata_Instance_Storage_Disk{
						{
							Device: "/dev/sda",
						},
					},
				},
			},
			State: "provisioning",
		},
		Network: &hardware.Hardware_Network{
			Interfaces: []*hardware.Hardware_Network_Interface{
				{
					Dhcp: &hardware.Hardware_DHCP{
						Arch:     "x86_64",
						Hostname: hostname,
						Ip: &hardware.Hardware_DHCP_IP{
							Address: ipAddress,
							Gateway: gateway,
							Netmask: netmask,
						},
						Mac:         mac,
						NameServers: nameservers,
						Uefi:        true,
					},
					Netboot: &hardware.Hardware_Netboot{
						AllowPxe:      true,
						AllowWorkflow: true,
					},
				},
			},
		},
	}

	return json.MarshalIndent(hardware, "", "  ")
}
