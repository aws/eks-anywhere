package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/go-logr/logr"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/netbox-community/go-netbox/netbox/client"
	"github.com/netbox-community/go-netbox/netbox/client/dcim"
	"github.com/netbox-community/go-netbox/netbox/client/ipam"
)

type Netbox struct {
	Host    string
	User    string
	Pass    string
	Records []*Machine
	logger  logr.Logger
}

type IPError struct {
	act string
}

func (i *IPError) Error() string {
	return fmt.Sprintf("Error Parsing IP: expected: CIDR Address, got: %v", i.act)
}

func (i *IPError) Is(target error) bool {
	t, oK := target.(*IPError)
	if !oK {
		return false
	}
	return (i.act == t.act || t.act == "")
}

type TypeAssertError struct {
	field string
	exp   string
	act   string
}

func (t *TypeAssertError) Error() string {
	return fmt.Sprintf("Error in Type Assertion: field: %v, expected: %v, got: %v", t.field, t.exp, t.act)
}

func (t *TypeAssertError) Is(target error) bool {
	tar, oK := target.(*TypeAssertError)
	if !oK {
		return false
	}
	return (t.field == tar.field || t.field == "") && (t.exp == tar.exp || t.exp == "") && (t.act == tar.act || t.act == "")
}

type NetboxError struct {
	msg    string
	errMsg string
}

func (n *NetboxError) Error() string {
	return fmt.Sprintf(n.msg + " : " + n.errMsg)
}

func (n *NetboxError) Is(target error) bool {
	tar, oK := target.(*NetboxError)
	if !oK {
		return false
	}
	return (n.msg == tar.msg || n.msg == "") && (n.errMsg == tar.errMsg || n.errMsg == "")
}

// ReadFromNetbox Function calls 3 helper functions which makes API calls to Netbox and sets Records field with required Hardware value.
func (n *Netbox) ReadFromNetbox(ctx context.Context, host string, validationToKen string) error {
	transport := httptransport.New(host, client.DefaultBasePath, []string{"http"})
	transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", "ToKen "+validationToKen)

	c := client.New(transport, nil)

	// Get the devices list from netbox to populate the Machine values
	deviceReq := dcim.NewDcimDevicesListParams()
	err := n.readDevicesFromNetbox(ctx, c, deviceReq)
	if err != nil {
		return fmt.Errorf("cannot get Devices list: %v ", err)
	}

	err = n.readInterfacesFromNetbox(ctx, c)
	if err != nil {
		return fmt.Errorf("error reading Interfaces list: %v ", err)
	}

	// Get the Interfaces list from netbox to populate the Machine gateway and nameserver value
	ipamReq := ipam.NewIpamIPRangesListParams()
	err = n.readIPRangeFromNetbox(ctx, c, ipamReq)
	if err != nil {
		return fmt.Errorf("error reading IP ranges list: %v ", err)
	}
	n.logger.V(1).Info("ALL DEVICES")

	for _, machine := range n.Records {
		n.logger.V(1).Info("Device Read: ", "Host", machine.Hostname, "IP", machine.IPAddress, "MAC", machine.MACAddress, "BMC-IP", machine.BMCIPAddress)
	}

	return nil
}

// rd Function calls 3 helper functions with a filter tag which makes API calls to Netbox and sets Records field with required Hardware value.
func (n *Netbox) readFromNetboxFiltered(ctx context.Context, host string, validationToKen string, filterTag string) error {
	transport := httptransport.New(host, client.DefaultBasePath, []string{"http"})
	transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", "ToKen "+validationToKen)

	c := client.New(transport, nil)

	deviceReq := dcim.NewDcimDevicesListParams()
	deviceReq.Tag = &filterTag

	err := n.readDevicesFromNetbox(ctx, c, deviceReq)
	if err != nil {
		return fmt.Errorf("could not get Devices list: %v", err)
	}
	err = n.readInterfacesFromNetbox(ctx, c)

	if err != nil {
		return fmt.Errorf("error reading Interfaces list: %v ", err)
	}

	ipamReq := ipam.NewIpamIPRangesListParams()
	err = n.readIPRangeFromNetbox(ctx, c, ipamReq)
	if err != nil {
		return fmt.Errorf("error reading IP ranges list: %v ", err)
	}
	n.logger.V(1).Info("FILTERED DEVICES")
	for _, machine := range n.Records {
		n.logger.V(1).Info("Device Read: ", "Host", machine.Hostname, "IP", machine.IPAddress, "MAC", machine.MACAddress, "BMC-IP", machine.BMCIPAddress)
	}
	return nil
}

// checkIP Function to check if a given ip address falls in between a start and end IP address.
func (n *Netbox) checkIP(ip string, startIPRange string, endIPRange string) bool {
	startIP, _, err := net.ParseCIDR(startIPRange)
	if err != nil {
		n.logger.Error(err, "error parsing IP in start range")
		return false
	}

	endIP, _, err := net.ParseCIDR(endIPRange)
	if err != nil {
		n.logger.Error(err, "error parsing IP in end range")
		return false
	}

	trial := net.ParseIP(ip)
	if trial.To4() == nil {
		n.logger.Error(err, "error parsing IP to IP4 address")
		return false
	}

	if bytes.Compare(trial, startIP) >= 0 && bytes.Compare(trial, endIP) <= 0 {
		return true
	}

	return false
}

// readDevicesFromNetbox Function fetches the devices list from Netbox and sets HostName, BMC info, Ip addr, Disk and Labels.
func (n *Netbox) readDevicesFromNetbox(ctx context.Context, c *client.NetBoxAPI, deviceReq *dcim.DcimDevicesListParams) error {
	option := func(o *runtime.ClientOperation) {
		o.Context = ctx
	}

	deviceRes, err := c.Dcim.DcimDevicesList(deviceReq, nil, option)
	if err != nil {
		return &NetboxError{"cannot get Devices list", err.Error()}
	}

	devicePayload := deviceRes.GetPayload()

	for _, device := range devicePayload.Results {
		machine := new(Machine)
		machine.Hostname = *device.Name

		customFields, oK := device.CustomFields.(map[string]interface{})
		if !oK {
			return &TypeAssertError{"CustomFields", "map[string]interface{}", fmt.Sprintf("%T", device.CustomFields)}
		}

		bmcIPMap, oK := customFields["bmc_ip"].(map[string]interface{})
		if !oK {
			return &TypeAssertError{"bmc_ip", "map[string]interface{}", fmt.Sprintf("%T", customFields["bmc_ip"])}
		}

		bmcIPVal, oK := bmcIPMap["address"].(string)
		if !oK {
			return &TypeAssertError{"bmc_ip_address", "string", fmt.Sprintf("%T", bmcIPMap["address"])}
		}

		// Check if the string returned in for bmc_ip is a valid IP.
		bmcIPValAdd, bmcIPValMask, err := net.ParseCIDR(bmcIPVal)
		if err != nil {
			return &IPError{bmcIPVal}
		}

		machine.BMCIPAddress = bmcIPValAdd.String()

		machine.Netmask = net.IP(bmcIPValMask.Mask).String()
		bmcUserVal, oK := customFields["bmc_username"].(string)
		if !oK {
			return &TypeAssertError{"bmc_username", "string", fmt.Sprintf("%T", customFields["bmc_username"])}
		}
		machine.BMCUsername = bmcUserVal

		bmcPassVal, oK := customFields["bmc_password"].(string)
		if !oK {
			return &TypeAssertError{"bmc_password", "string", fmt.Sprintf("%T", customFields["bmc_password"])}
		}
		machine.BMCPassword = bmcPassVal

		diskVal, oK := customFields["disk"].(string)
		if !oK {
			return &TypeAssertError{"disk", "string", fmt.Sprintf("%T", customFields["disk"])}
		}
		machine.Disk = diskVal

		// Obtain the machine IP from primary IP which contains IP/mask value
		machineIPAdd, _, err := net.ParseCIDR(*device.PrimaryIp4.Address)
		if err != nil {
			return &IPError{*device.PrimaryIp4.Address}
			// return fmt.Errorf("cannot parse Machine IP Address, %v", err)
		}
		machine.IPAddress = machineIPAdd.String()

		labelMap := make(map[string]string)
		controlFlag := false
		for _, tag := range device.Tags {
			if *tag.Name == "control-plane" {
				labelMap["type"] = "control-plane"
				controlFlag = !controlFlag
				break
			}
		}
		if !controlFlag {
			labelMap["type"] = "worker-plane"
		}
		machine.Labels = labelMap
		n.Records = append(n.Records, machine)
	}

	n.logger.Info("step 1 - Reading devices successul", "num_machines", len(n.Records))
	return nil
}

// ReadInterfacesFromNetbox Function fetches the interfaces list from Netbox and sets the MAC address for each record.
func (n *Netbox) readInterfacesFromNetbox(ctx context.Context, c *client.NetBoxAPI) error {
	// Get the Interfaces list from netbox to populate the Machine mac value
	interfacesReq := dcim.NewDcimInterfacesListParams()

	option := func(o *runtime.ClientOperation) {
		o.Context = ctx
	}
	for _, record := range n.Records {
		interfacesReq.Device = &record.Hostname
		interfacesRes, err := c.Dcim.DcimInterfacesList(interfacesReq, nil, option)
		if err != nil {
			return &NetboxError{"cannot get Interfaces list", err.Error()}
		}
		interfacesResults := interfacesRes.GetPayload().Results
		if len(interfacesResults) == 1 {
			record.MACAddress = *interfacesResults[0].MacAddress
		} else {
			for _, interfaces := range interfacesResults {
				for _, tagName := range interfaces.Tags {
					if *tagName.Name == "eks-a" {
						record.MACAddress = *interfaces.MacAddress
					}
				}
			}
		}
	}

	n.logger.Info("step 2 - Reading intefaces successful, MAC addresses set")

	return nil
}

// ReadIpRangeFromNetbox Function fetches IP ranges from Netbox and sets the Gateway and nameserver address for each record.
func (n *Netbox) readIPRangeFromNetbox(ctx context.Context, c *client.NetBoxAPI, ipamReq *ipam.IpamIPRangesListParams) error {
	option := func(o *runtime.ClientOperation) {
		o.Context = ctx
	}
	ipamRes, err := c.Ipam.IpamIPRangesList(ipamReq, nil, option)
	if err != nil {
		return fmt.Errorf("cannot get IP ranges list: %v ", err)
	}
	ipamPayload := ipamRes.GetPayload()

	for _, record := range n.Records {
		for _, ipRange := range ipamPayload.Results {
			// nolint: nestif // Check for ip for optimization
			if n.checkIP(record.IPAddress, *ipRange.StartAddress, *ipRange.EndAddress) {
				customFields, oK := ipRange.CustomFields.(map[string]interface{})
				if !oK {
					return &TypeAssertError{"customFields", "map[string]interface{}", fmt.Sprintf("%T", ipRange.CustomFields)}
				}

				gatewayIPMap, oK := customFields["gateway"].(map[string]interface{})
				if !oK {
					return &TypeAssertError{"gatewayIP", "map[string]interface{}", fmt.Sprintf("%T", customFields["gateway"])}
				}

				gatewayIPVal, oK := gatewayIPMap["address"].(string)
				if !oK {
					return &TypeAssertError{"gatewayAddr", "string", fmt.Sprintf("%T", gatewayIPMap["address"])}
				}

				gatewayIPAdd, _, err := net.ParseCIDR(gatewayIPVal)
				if err != nil {
					return &IPError{gatewayIPVal}
				}

				nameserversIPs, oK := customFields["nameservers"].([]interface{})
				if !oK {
					return &TypeAssertError{"nameservers", "[]interface{}", fmt.Sprintf("%T", customFields["nameservers"])}
				}

				var nsIP Nameservers

				for _, nameserverIP := range nameserversIPs {
					nameserversIPsMap, oK := nameserverIP.(map[string]interface{})
					if !oK {
						return &TypeAssertError{"nameserversIPMap", "map[string]interface{}", fmt.Sprintf("%T", nameserverIP)}
					}

					nameserverIPVal, oK := nameserversIPsMap["address"].(string)
					if !oK {
						return &TypeAssertError{"nameserversIPMap", "string", fmt.Sprintf("%T", nameserversIPsMap["address"])}
					}

					// Parse CIDR reasoning and explanation about the type returned by netbox
					// Check if string returned by nameserverIPVal is a valid IP.
					nameserverIPAdd, _, err := net.ParseCIDR(nameserverIPVal)
					if err != nil {
						return &IPError{nameserverIPVal}
					}

					nsIP = append(nsIP, nameserverIPAdd.String())
				}
				record.Nameservers = nsIP
				record.Gateway = gatewayIPAdd.String()
			}
		}
	}

	n.logger.Info("step 3 - Reading IPAM data successful, all DCIM calls are complete")

	return nil
}

// serializeMachines Function takes in a arry of machine slices as input and converts them into byte array.
func (n *Netbox) serializeMachines(machines []*Machine) ([]byte, error) {
	ret, err := json.MarshalIndent(machines, "", " ")
	if err != nil {
		return nil, fmt.Errorf("error in encoding Machines to byte Array: %v", err)
	}
	return ret, nil
}
