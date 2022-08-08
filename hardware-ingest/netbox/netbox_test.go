package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-openapi/runtime"
	"github.com/google/go-cmp/cmp"
	"github.com/netbox-community/go-netbox/netbox/client"
	"github.com/netbox-community/go-netbox/netbox/client/dcim"
	"github.com/netbox-community/go-netbox/netbox/client/ipam"
	"github.com/netbox-community/go-netbox/netbox/models"
)

func TestCheckIP(t *testing.T) {
	type checkIPTest struct {
		toCheck, IPStart, IPEnd string
		want                    bool
	}

	checkIPTests := []checkIPTest{
		{"10.80.21.32", "10.80.21.31/21", "10.80.21.51/21", true},
		{"10.80.21.35", "10.80.21.31/21", "10.80.21.51/21", true},
		{"25.82.21.32", "10.80.21.31/21", "10.80.21.51/21", false},
		{"100.100.100.1000", "10.80.21.31/21", "10.80.21.51/21", false},
		{"25.82.21.32", "10.800.21.31/21", "10.80.21.51/21", false},
		{"25.82.21.32", "10.80.21.31/21", "10.800.21.51/21", false},
	}

	n := new(Netbox)
	n.logger = logr.Discard()
	for _, test := range checkIPTests {
		if output := n.checkIP(test.toCheck, test.IPStart, test.IPEnd); output != test.want {
			t.Errorf("output %v not equal to expected %v", test.toCheck, test.want)
		}
	}
}

func toPointer(v string) *string { return &v }

func TestReadDevicesFromNetbox(t *testing.T) {
	type outputs struct {
		bmcIP       string
		bmcUsername string
		bmcPassword string
		disk        string
		label       string
		name        string
		primIP      string
		ifError     error
	}

	type inputs struct {
		v    outputs
		err  error
		want []*Machine
	}

	tests := []inputs{
		// Checking happy flow with control-plane
		{
			v: outputs{
				bmcIP:       "192.168.2.5/22",
				bmcUsername: "root",
				bmcPassword: "root",
				disk:        "/dev/sda",
				label:       "control-plane",
				name:        "dev",
				primIP:      "192.18.2.5/22",
				ifError:     nil,
			},
			err: nil, want: []*Machine{
				{
					Hostname:  "dev",
					IPAddress: "192.18.2.5",
					Netmask:   "255.255.252.0",
					Disk:      "/dev/sda",
					Labels: map[string]string{
						"type": "control-plane",
					},
					BMCIPAddress: "192.168.2.5",
					BMCUsername:  "root",
					BMCPassword:  "root",
				},
			},
		},
		// Checking happy flow with worker-plane
		{
			v: outputs{
				bmcIP:       "192.168.2.5/22",
				bmcUsername: "root",
				bmcPassword: "root",
				disk:        "/dev/sda",
				name:        "dev",
				primIP:      "192.18.2.5/22",
				ifError:     nil,
			},
			err: nil, want: []*Machine{
				{
					Hostname:  "dev",
					IPAddress: "192.18.2.5",
					Netmask:   "255.255.252.0",
					Disk:      "/dev/sda",
					Labels: map[string]string{
						"type": "worker-plane",
					},
					BMCIPAddress: "192.168.2.5",
					BMCUsername:  "root",
					BMCPassword:  "root",
				},
			},
		},

		// Checking unhappy flow with bmcIPwithout Mask
		{
			v: outputs{
				bmcIP:       "192.168.2.5",
				bmcUsername: "root",
				bmcPassword: "root",
				disk:        "/dev/sda",
				name:        "dev",
				primIP:      "192.18.2.5/22",
				ifError:     &IPError{"192.168.2.5"},
			},
			err: nil, want: []*Machine{
				{},
			},
		},
		// Checking unhappy flow with IPV6 address for prim IP
		{
			v: outputs{
				bmcIP:       "192.168.2.5/22",
				bmcUsername: "root",
				bmcPassword: "root",
				disk:        "/dev/sda",
				label:       "control-plane",
				name:        "dev",
				primIP:      "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
				ifError:     &IPError{"2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
			},
			err: nil, want: []*Machine{
				{},
			},
		},
		// Checking unhappy flow with invalid IPv4 address with mask
		{
			v: outputs{
				bmcIP:       "192.460.634.516/22",
				bmcUsername: "root",
				bmcPassword: "root",
				disk:        "/dev/sda",
				label:       "",
				name:        "dev",
				primIP:      "192.18.2.5/22",
				ifError:     &IPError{"192.460.634.516/22"},
			},
			err: nil, want: []*Machine{
				{},
			},
		},
		{
			v: outputs{
				ifError: &NetboxError{"cannot get Devices list", "error code 500-Internal Server Error"},
			},
			err: errors.New("error code 500-Internal Server Error"), want: []*Machine{},
		},
	}

	for idx, tt := range tests {
		t.Run(fmt.Sprintf("%v", idx), func(t *testing.T) {
			n := new(Netbox)
			n.logger = logr.Discard()
			d := new(models.DeviceWithConfigContext)
			d.Tags = []*models.NestedTag{{Name: &tt.v.label}}
			d.Name = toPointer(tt.v.name)
			d.CustomFields = map[string]interface{}{
				"bmc_ip":       map[string]interface{}{"address": tt.v.bmcIP},
				"bmc_username": tt.v.bmcUsername,
				"bmc_password": tt.v.bmcPassword,
				"disk":         tt.v.disk,
			}
			d.PrimaryIp4 = &models.NestedIPAddress{Address: toPointer(tt.v.primIP)}
			dummyDevListOK := new(dcim.DcimDevicesListOK)
			dummyDevListOKBody := new(dcim.DcimDevicesListOKBody)

			// dummyDevListOK.Payload = new(models.Device)
			dummyDevListOKBody.Results = []*models.DeviceWithConfigContext{d}
			dummyDevListOK.Payload = dummyDevListOKBody
			v := &mock{v: dummyDevListOK, err: tt.err}
			c := &client.NetBoxAPI{Dcim: v}
			deviceReq := dcim.NewDcimDevicesListParams()
			err := n.readDevicesFromNetbox(context.TODO(), c, deviceReq)

			if err != nil {
				if !errors.Is(err, tt.v.ifError) {
					t.Fatal("Got: ", err.Error(), "want: ", tt.v.ifError)
				}
			} else {
				if diff := cmp.Diff(n.Records, tt.want); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestReadInterfacesFromNetbox(t *testing.T) {
	type outputs struct {
		MacAddress []string
		Name       []string
		device     string
		Tag        int
		ifError    error
	}

	type inputs struct {
		v    outputs
		err  error
		want []*Machine
	}

	tests := []inputs{
		// Checking happy flow with 1 interface mapped to device
		{
			v: outputs{
				MacAddress: []string{"CC:48:3A:11:F4:C1"},
				Name:       []string{"GigabitEthernet1"},
				device:     "eksa-dev01",
				ifError:    nil,
			},
			err: nil, want: []*Machine{
				{
					Hostname:   "eksa-dev01",
					MACAddress: "CC:48:3A:11:F4:C1",
				},
			},
		},
		// Checking happy flow with 3 interfaces mapped to device and primary interface being 1st interface (0-based indexing)
		{
			v: outputs{
				MacAddress: []string{"CC:48:3A:11:F4:C1", "CC:48:3A:11:EA:11", "CC:48:3A:11:EA:61"},
				Name:       []string{"GigabitEthernet1", "GigabitEthernet1-a", "GigabitEthernet1-b"},
				device:     "eksa-dev01",
				Tag:        1,
				ifError:    nil,
			},
			err: nil, want: []*Machine{
				{
					Hostname:   "eksa-dev01",
					MACAddress: "CC:48:3A:11:EA:11",
				},
			},
		},
		// Checking Unhappy flow by generating error from API
		{
			v: outputs{
				device:  "errorDev",
				ifError: &NetboxError{"cannot get Interfaces list", "error code 500-Internal Server Error"},
			},
			err: errors.New("error code 500-Internal Server Error"), want: []*Machine{},
		},
	}
	for idx, tt := range tests {
		t.Run(fmt.Sprintf("%v", idx), func(t *testing.T) {
			n := new(Netbox)
			dummyMachine := &Machine{
				Hostname: tt.v.device,
			}

			n.Records = append(n.Records, dummyMachine)
			n.logger = logr.Discard()

			dummyInterfaceList := make([]*models.Interface, len(tt.v.MacAddress))
			for idx := range tt.v.MacAddress {
				i := new(models.Interface)
				i.Name = &tt.v.Name[idx]

				i.MacAddress = &tt.v.MacAddress[idx]
				if idx == tt.v.Tag {
					i.Tags = []*models.NestedTag{{Name: toPointer("eks-a")}}
				}
				dummyInterfaceList[idx] = i
			}

			dummyIntListOK := new(dcim.DcimInterfacesListOK)
			dummyIntListOKBody := new(dcim.DcimInterfacesListOKBody)
			dummyIntListOKBody.Results = dummyInterfaceList
			dummyIntListOK.Payload = dummyIntListOKBody
			i := &mock{i: dummyIntListOK, err: tt.err}
			c := &client.NetBoxAPI{Dcim: i}

			err := n.readInterfacesFromNetbox(context.TODO(), c)

			if err != nil {
				if !errors.Is(err, tt.v.ifError) {
					t.Fatal("Got: ", err.Error(), "want: ", tt.v.ifError)
				}
			} else {
				if diff := cmp.Diff(n.Records, tt.want); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestTypeAssertions(t *testing.T) {
	type outputs struct {
		bmcIP       interface{}
		bmcUsername interface{}
		bmcPassword interface{}
		disk        interface{}
		name        string
		primIP      string
	}

	type inputs struct {
		v    outputs
		err  error
		want error
	}

	tests := []inputs{
		{
			v: outputs{
				bmcIP:       "192.168.2.5/22",
				bmcUsername: "root",
				bmcPassword: "root",
				disk:        "/dev/sda",
				name:        "dev",
				primIP:      "192.18.2.5/22",
			},
			err: nil, want: &TypeAssertError{"bmc_ip", "map[string]interface{}", "string"},
		},
		{
			v: outputs{
				bmcIP:       map[string]interface{}{"address": 192.431},
				bmcUsername: "root",
				bmcPassword: "root",
				disk:        "/dev/sda",
				name:        "dev",
				primIP:      "192.18.2.5/22",
			},
			err: nil, want: &TypeAssertError{"bmc_ip_address", "string", "float64"},
		},
		{
			v: outputs{
				bmcIP:       map[string]interface{}{"address": "192.168.2.5/22"},
				bmcUsername: []string{"root1", "root2"},
				bmcPassword: "root",
				disk:        "/dev/sda",
				name:        "dev",
				primIP:      "192.18.2.5/22",
			},
			err: nil, want: &TypeAssertError{"bmc_username", "string", "[]string"},
		},
		{
			v: outputs{
				bmcIP:       map[string]interface{}{"address": "192.168.2.5/22"},
				bmcUsername: "root1",
				bmcPassword: []string{"root1", "root2"},
				disk:        "/dev/sda",
				name:        "dev",
				primIP:      "192.18.2.5/22",
			},
			err: nil, want: &TypeAssertError{"bmc_password", "string", "[]string"},
		},
		{
			v: outputs{
				bmcIP:       map[string]interface{}{"address": "192.168.2.5/22"},
				bmcUsername: "root",
				bmcPassword: "root",
				disk:        123,
				name:        "dev",
				primIP:      "192.18.2.5/22",
			},
			err: nil, want: &TypeAssertError{"disk", "string", "int"},
		},
	}

	for idx, tt := range tests {
		t.Run(fmt.Sprintf("%v", idx), func(t *testing.T) {
			n := new(Netbox)
			n.logger = logr.Discard()
			d := new(models.DeviceWithConfigContext)
			d.Name = toPointer(tt.v.name)

			d.CustomFields = map[string]interface{}{
				"bmc_ip":       tt.v.bmcIP,
				"bmc_username": tt.v.bmcUsername,
				"bmc_password": tt.v.bmcPassword,
				"disk":         tt.v.disk,
			}
			d.PrimaryIp4 = &models.NestedIPAddress{Address: toPointer(tt.v.primIP)}
			dummyDevListOK := new(dcim.DcimDevicesListOK)
			dummyDevListOKBody := new(dcim.DcimDevicesListOKBody)

			dummyDevListOKBody.Results = []*models.DeviceWithConfigContext{d}
			dummyDevListOK.Payload = dummyDevListOKBody
			v := &mock{v: dummyDevListOK, err: tt.err}
			c := &client.NetBoxAPI{Dcim: v}
			deviceReq := dcim.NewDcimDevicesListParams()
			err := n.readDevicesFromNetbox(context.TODO(), c, deviceReq)

			if err != nil {
				if !errors.Is(err, tt.want) {
					t.Fatalf("Got: %v, want: %v", err.Error(), tt.want)
				}
			} else {
				if diff := cmp.Diff(n.Records, tt.want); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestReadIPRangeFromNetbox(t *testing.T) {
	type outputs struct {
		gatewayIP    interface{}
		nameserverIP []interface{}
		startIP      string
		endIP        string
		ifError      error
	}

	type inputs struct {
		v    outputs
		err  error
		want []*Machine
	}

	tests := []inputs{
		{
			v: outputs{
				gatewayIP:    map[string]interface{}{"address": "10.80.8.1/22"},
				nameserverIP: []interface{}{map[string]interface{}{"address": "208.91.112.53/22"}},
				startIP:      "10.80.12.20/22",
				endIP:        "10.80.12.30/22",
			},
			err: nil, want: []*Machine{
				{
					IPAddress:   "10.80.12.25",
					Gateway:     "10.80.8.1",
					Nameservers: Nameservers{"208.91.112.53"},
				},
			},
		},
		{
			v: outputs{
				gatewayIP:    map[string]interface{}{"address": "10.800.8.1/22"},
				nameserverIP: []interface{}{map[string]interface{}{"address": "208.91.112.53/22"}},
				startIP:      "10.80.12.20/22",
				endIP:        "10.80.12.30/22",
				ifError:      &IPError{"10.800.8.1/22"},
			},
			err: nil, want: []*Machine{},
		},
		{
			v: outputs{
				gatewayIP:    map[string]interface{}{"address": "10.80.8.1/22"},
				nameserverIP: []interface{}{map[string]interface{}{"address": "208.910.112.53/22"}},
				startIP:      "10.80.12.20/22",
				endIP:        "10.80.12.30/22",
				ifError:      &IPError{"208.910.112.53/22"},
			},
			err: nil, want: []*Machine{},
		},
		{
			v: outputs{
				gatewayIP:    map[string]string{"address": "10.80.8.1/22"},
				nameserverIP: []interface{}{map[string]interface{}{"address": "208.91.112.53/22"}},
				startIP:      "10.80.12.20/22",
				endIP:        "10.80.12.30/22",
				ifError:      &TypeAssertError{"gatewayIP", "map[string]interface{}", "map[string]string"},
			},
			err: nil, want: []*Machine{},
		},
		{
			v: outputs{
				gatewayIP:    map[string]interface{}{"address": 102.45},
				nameserverIP: []interface{}{map[string]interface{}{"address": "208.91.112.53/22"}},
				startIP:      "10.80.12.20/22",
				endIP:        "10.80.12.30/22",
				ifError:      &TypeAssertError{"gatewayAddr", "string", "float64"},
			},
			err: nil, want: []*Machine{},
		},
		{
			v: outputs{
				gatewayIP:    map[string]interface{}{"address": "10.80.8.1/22"},
				nameserverIP: []interface{}{"208.91.112.53/22", "208.91.112.53/22"},
				startIP:      "10.80.12.20/22",
				endIP:        "10.80.12.30/22",
				ifError:      &TypeAssertError{"nameserversIPMap", "map[string]interface{}", "string"},
			},
			err: nil, want: []*Machine{},
		},
		{
			v: outputs{
				gatewayIP:    map[string]interface{}{"address": "10.80.8.1/22"},
				nameserverIP: []interface{}{map[string]interface{}{"address": 208.91}},
				startIP:      "10.80.12.20/22",
				endIP:        "10.80.12.30/22",
				ifError:      &TypeAssertError{"nameserversIPMap", "string", "float64"},
			},
			err: nil, want: []*Machine{},
		},
	}

	for _, tt := range tests {
		n := new(Netbox)
		dummyMachine := &Machine{
			IPAddress: "10.80.12.25",
		}

		n.Records = append(n.Records, dummyMachine)
		n.logger = logr.Discard()

		d := new(models.IPRange)
		d.StartAddress = &tt.v.startIP
		d.EndAddress = &tt.v.endIP
		d.CustomFields = map[string]interface{}{
			"gateway":     tt.v.gatewayIP,
			"nameservers": tt.v.nameserverIP,
		}
		dummyIPrangeListOk := new(ipam.IpamIPRangesListOK)
		dummyIPrangeListOkBody := new(ipam.IpamIPRangesListOKBody)
		dummyIPrangeListOkBody.Results = []*models.IPRange{d}
		dummyIPrangeListOk.Payload = dummyIPrangeListOkBody
		i := &mock{IP: dummyIPrangeListOk, err: tt.err}
		c := &client.NetBoxAPI{Ipam: i}

		iPRangeReq := ipam.NewIpamIPRangesListParams()
		err := n.readIPRangeFromNetbox(context.TODO(), c, iPRangeReq)

		if err != nil {
			if !errors.Is(err, tt.v.ifError) {
				t.Fatal("Got: ", err.Error(), "want: ", tt.v.ifError)
			}
		} else {
			if diff := cmp.Diff(n.Records, tt.want); diff != "" {
				t.Fatal(diff)
			}
		}
	}
}

func TestSerializeMachines(t *testing.T) {
	test := []*Machine{
		{Hostname: "Dev1", IPAddress: "10.80.8.21", Netmask: "255.255.255.0", Gateway: "192.168.2.1", Nameservers: []string{"1.1.1.1"}, MACAddress: "CC:48:3A:11:F4:C1", Disk: "/dev/sda", Labels: map[string]string{"type": "worker-plane"}, BMCIPAddress: "10.80.12.20", BMCUsername: "root", BMCPassword: "pPyU6mAO"},
		{Hostname: "Dev2", IPAddress: "10.80.8.22", Netmask: "255.255.255.0", Gateway: "192.168.2.1", Nameservers: []string{"1.1.1.1"}, MACAddress: "CC:48:3A:11:EA:11", Disk: "/dev/sda", Labels: map[string]string{"type": "control-plane"}, BMCIPAddress: "10.80.12.21", BMCUsername: "root", BMCPassword: "pPyU6mAO"},
	}

	want := createMachineString(test)
	n := new(Netbox)
	n.logger = logr.Discard()

	got, err := n.serializeMachines(test)
	if err != nil {
		t.Fatal("Error: ", err)
	}

	if !bytes.EqualFold(got, []byte(want)) {
		t.Fatal(cmp.Diff(got, []byte(want)))
	}
}

type mock struct {
	v   *dcim.DcimDevicesListOK
	i   *dcim.DcimInterfacesListOK
	IP  *ipam.IpamIPRangesListOK
	err error
}

func (m *mock) DcimCablesBulkDelete(_ *dcim.DcimCablesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimCablesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimCablesBulkPartialUpdate(_ *dcim.DcimCablesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimCablesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimCablesBulkUpdate(_ *dcim.DcimCablesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimCablesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimCablesCreate(_ *dcim.DcimCablesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimCablesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimCablesDelete(_ *dcim.DcimCablesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimCablesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimCablesList(_ *dcim.DcimCablesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimCablesListOK, error) {
	return nil, nil
}

func (m *mock) DcimCablesPartialUpdate(_ *dcim.DcimCablesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimCablesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimCablesRead(_ *dcim.DcimCablesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimCablesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimCablesUpdate(_ *dcim.DcimCablesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimCablesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConnectedDeviceList(_ *dcim.DcimConnectedDeviceListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConnectedDeviceListOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortTemplatesBulkDelete(_ *dcim.DcimConsolePortTemplatesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortTemplatesBulkPartialUpdate(_ *dcim.DcimConsolePortTemplatesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortTemplatesBulkUpdate(_ *dcim.DcimConsolePortTemplatesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortTemplatesCreate(_ *dcim.DcimConsolePortTemplatesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortTemplatesDelete(_ *dcim.DcimConsolePortTemplatesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortTemplatesList(_ *dcim.DcimConsolePortTemplatesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortTemplatesPartialUpdate(_ *dcim.DcimConsolePortTemplatesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortTemplatesRead(_ *dcim.DcimConsolePortTemplatesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortTemplatesUpdate(_ *dcim.DcimConsolePortTemplatesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsolePortTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsBulkDelete(_ *dcim.DcimConsolePortsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsolePortsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsBulkPartialUpdate(_ *dcim.DcimConsolePortsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsolePortsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsBulkUpdate(_ *dcim.DcimConsolePortsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsolePortsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsCreate(_ *dcim.DcimConsolePortsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsolePortsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsDelete(_ *dcim.DcimConsolePortsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsolePortsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsList(_ *dcim.DcimConsolePortsListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsolePortsListOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsPartialUpdate(_ *dcim.DcimConsolePortsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsolePortsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsRead(_ *dcim.DcimConsolePortsReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsolePortsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsTrace(_ *dcim.DcimConsolePortsTraceParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsolePortsTraceOK, error) {
	return nil, nil
}

func (m *mock) DcimConsolePortsUpdate(_ *dcim.DcimConsolePortsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsolePortsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortTemplatesBulkDelete(_ *dcim.DcimConsoleServerPortTemplatesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortTemplatesBulkPartialUpdate(_ *dcim.DcimConsoleServerPortTemplatesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortTemplatesBulkUpdate(_ *dcim.DcimConsoleServerPortTemplatesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortTemplatesCreate(_ *dcim.DcimConsoleServerPortTemplatesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortTemplatesDelete(_ *dcim.DcimConsoleServerPortTemplatesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortTemplatesList(_ *dcim.DcimConsoleServerPortTemplatesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortTemplatesPartialUpdate(_ *dcim.DcimConsoleServerPortTemplatesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortTemplatesRead(_ *dcim.DcimConsoleServerPortTemplatesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortTemplatesUpdate(_ *dcim.DcimConsoleServerPortTemplatesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsoleServerPortTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsBulkDelete(_ *dcim.DcimConsoleServerPortsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsBulkPartialUpdate(_ *dcim.DcimConsoleServerPortsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsBulkUpdate(_ *dcim.DcimConsoleServerPortsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsCreate(_ *dcim.DcimConsoleServerPortsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsDelete(_ *dcim.DcimConsoleServerPortsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsList(_ *dcim.DcimConsoleServerPortsListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsListOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsPartialUpdate(_ *dcim.DcimConsoleServerPortsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsRead(_ *dcim.DcimConsoleServerPortsReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsTrace(_ *dcim.DcimConsoleServerPortsTraceParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsTraceOK, error) {
	return nil, nil
}

func (m *mock) DcimConsoleServerPortsUpdate(_ *dcim.DcimConsoleServerPortsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimConsoleServerPortsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBayTemplatesBulkDelete(_ *dcim.DcimDeviceBayTemplatesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBayTemplatesBulkPartialUpdate(_ *dcim.DcimDeviceBayTemplatesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBayTemplatesBulkUpdate(_ *dcim.DcimDeviceBayTemplatesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBayTemplatesCreate(_ *dcim.DcimDeviceBayTemplatesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBayTemplatesDelete(_ *dcim.DcimDeviceBayTemplatesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBayTemplatesList(_ *dcim.DcimDeviceBayTemplatesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBayTemplatesPartialUpdate(_ *dcim.DcimDeviceBayTemplatesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBayTemplatesRead(_ *dcim.DcimDeviceBayTemplatesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBayTemplatesUpdate(_ *dcim.DcimDeviceBayTemplatesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceBayTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBaysBulkDelete(_ *dcim.DcimDeviceBaysBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceBaysBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBaysBulkPartialUpdate(_ *dcim.DcimDeviceBaysBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceBaysBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBaysBulkUpdate(_ *dcim.DcimDeviceBaysBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceBaysBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBaysCreate(_ *dcim.DcimDeviceBaysCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceBaysCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBaysDelete(_ *dcim.DcimDeviceBaysDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceBaysDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBaysList(_ *dcim.DcimDeviceBaysListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceBaysListOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBaysPartialUpdate(_ *dcim.DcimDeviceBaysPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceBaysPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBaysRead(_ *dcim.DcimDeviceBaysReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceBaysReadOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceBaysUpdate(_ *dcim.DcimDeviceBaysUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceBaysUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceRolesBulkDelete(_ *dcim.DcimDeviceRolesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceRolesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDeviceRolesBulkPartialUpdate(_ *dcim.DcimDeviceRolesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceRolesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceRolesBulkUpdate(_ *dcim.DcimDeviceRolesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceRolesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceRolesCreate(_ *dcim.DcimDeviceRolesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceRolesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimDeviceRolesDelete(_ *dcim.DcimDeviceRolesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceRolesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDeviceRolesList(_ *dcim.DcimDeviceRolesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceRolesListOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceRolesPartialUpdate(_ *dcim.DcimDeviceRolesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceRolesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceRolesRead(_ *dcim.DcimDeviceRolesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceRolesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceRolesUpdate(_ *dcim.DcimDeviceRolesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceRolesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceTypesBulkDelete(_ *dcim.DcimDeviceTypesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceTypesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDeviceTypesBulkPartialUpdate(_ *dcim.DcimDeviceTypesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceTypesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceTypesBulkUpdate(_ *dcim.DcimDeviceTypesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceTypesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceTypesCreate(_ *dcim.DcimDeviceTypesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceTypesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimDeviceTypesDelete(_ *dcim.DcimDeviceTypesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceTypesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDeviceTypesList(_ *dcim.DcimDeviceTypesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceTypesListOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceTypesPartialUpdate(_ *dcim.DcimDeviceTypesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceTypesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceTypesRead(_ *dcim.DcimDeviceTypesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceTypesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimDeviceTypesUpdate(_ *dcim.DcimDeviceTypesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDeviceTypesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDevicesBulkDelete(_ *dcim.DcimDevicesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDevicesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDevicesBulkPartialUpdate(_ *dcim.DcimDevicesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDevicesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDevicesBulkUpdate(_ *dcim.DcimDevicesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDevicesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDevicesCreate(_ *dcim.DcimDevicesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDevicesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimDevicesDelete(_ *dcim.DcimDevicesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDevicesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimDevicesList(_ *dcim.DcimDevicesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDevicesListOK, error) {
	return m.v, m.err
}

func (m *mock) DcimDevicesNapalm(_ *dcim.DcimDevicesNapalmParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDevicesNapalmOK, error) {
	return nil, nil
}

func (m *mock) DcimDevicesPartialUpdate(_ *dcim.DcimDevicesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDevicesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimDevicesRead(_ *dcim.DcimDevicesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDevicesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimDevicesUpdate(_ *dcim.DcimDevicesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimDevicesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortTemplatesBulkDelete(_ *dcim.DcimFrontPortTemplatesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortTemplatesBulkPartialUpdate(_ *dcim.DcimFrontPortTemplatesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortTemplatesBulkUpdate(_ *dcim.DcimFrontPortTemplatesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortTemplatesCreate(_ *dcim.DcimFrontPortTemplatesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortTemplatesDelete(_ *dcim.DcimFrontPortTemplatesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortTemplatesList(_ *dcim.DcimFrontPortTemplatesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortTemplatesPartialUpdate(_ *dcim.DcimFrontPortTemplatesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortTemplatesRead(_ *dcim.DcimFrontPortTemplatesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortTemplatesUpdate(_ *dcim.DcimFrontPortTemplatesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimFrontPortTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsBulkDelete(_ *dcim.DcimFrontPortsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimFrontPortsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsBulkPartialUpdate(_ *dcim.DcimFrontPortsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimFrontPortsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsBulkUpdate(_ *dcim.DcimFrontPortsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimFrontPortsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsCreate(_ *dcim.DcimFrontPortsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimFrontPortsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsDelete(_ *dcim.DcimFrontPortsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimFrontPortsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsList(_ *dcim.DcimFrontPortsListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimFrontPortsListOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsPartialUpdate(_ *dcim.DcimFrontPortsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimFrontPortsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsPaths(_ *dcim.DcimFrontPortsPathsParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimFrontPortsPathsOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsRead(_ *dcim.DcimFrontPortsReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimFrontPortsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimFrontPortsUpdate(_ *dcim.DcimFrontPortsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimFrontPortsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfaceTemplatesBulkDelete(_ *dcim.DcimInterfaceTemplatesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInterfaceTemplatesBulkPartialUpdate(_ *dcim.DcimInterfaceTemplatesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfaceTemplatesBulkUpdate(_ *dcim.DcimInterfaceTemplatesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfaceTemplatesCreate(_ *dcim.DcimInterfaceTemplatesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimInterfaceTemplatesDelete(_ *dcim.DcimInterfaceTemplatesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInterfaceTemplatesList(_ *dcim.DcimInterfaceTemplatesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfaceTemplatesPartialUpdate(_ *dcim.DcimInterfaceTemplatesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfaceTemplatesRead(_ *dcim.DcimInterfaceTemplatesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfaceTemplatesUpdate(_ *dcim.DcimInterfaceTemplatesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInterfaceTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfacesBulkDelete(_ *dcim.DcimInterfacesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInterfacesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInterfacesBulkPartialUpdate(_ *dcim.DcimInterfacesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInterfacesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfacesBulkUpdate(_ *dcim.DcimInterfacesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInterfacesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfacesCreate(_ *dcim.DcimInterfacesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInterfacesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimInterfacesDelete(_ *dcim.DcimInterfacesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInterfacesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInterfacesList(_ *dcim.DcimInterfacesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInterfacesListOK, error) {
	return m.i, m.err
}

func (m *mock) DcimInterfacesPartialUpdate(_ *dcim.DcimInterfacesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInterfacesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfacesRead(_ *dcim.DcimInterfacesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInterfacesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfacesTrace(_ *dcim.DcimInterfacesTraceParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInterfacesTraceOK, error) {
	return nil, nil
}

func (m *mock) DcimInterfacesUpdate(_ *dcim.DcimInterfacesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInterfacesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemRolesBulkDelete(_ *dcim.DcimInventoryItemRolesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemRolesBulkPartialUpdate(_ *dcim.DcimInventoryItemRolesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemRolesBulkUpdate(_ *dcim.DcimInventoryItemRolesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemRolesCreate(_ *dcim.DcimInventoryItemRolesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemRolesDelete(_ *dcim.DcimInventoryItemRolesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemRolesList(_ *dcim.DcimInventoryItemRolesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesListOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemRolesPartialUpdate(_ *dcim.DcimInventoryItemRolesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemRolesRead(_ *dcim.DcimInventoryItemRolesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemRolesUpdate(_ *dcim.DcimInventoryItemRolesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemRolesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemTemplatesBulkDelete(_ *dcim.DcimInventoryItemTemplatesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemTemplatesBulkPartialUpdate(_ *dcim.DcimInventoryItemTemplatesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemTemplatesBulkUpdate(_ *dcim.DcimInventoryItemTemplatesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemTemplatesCreate(_ *dcim.DcimInventoryItemTemplatesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemTemplatesDelete(_ *dcim.DcimInventoryItemTemplatesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemTemplatesList(_ *dcim.DcimInventoryItemTemplatesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemTemplatesPartialUpdate(_ *dcim.DcimInventoryItemTemplatesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemTemplatesRead(_ *dcim.DcimInventoryItemTemplatesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemTemplatesUpdate(_ *dcim.DcimInventoryItemTemplatesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemsBulkDelete(_ *dcim.DcimInventoryItemsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemsBulkPartialUpdate(_ *dcim.DcimInventoryItemsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemsBulkUpdate(_ *dcim.DcimInventoryItemsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemsCreate(_ *dcim.DcimInventoryItemsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemsDelete(_ *dcim.DcimInventoryItemsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemsList(_ *dcim.DcimInventoryItemsListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemsListOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemsPartialUpdate(_ *dcim.DcimInventoryItemsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemsRead(_ *dcim.DcimInventoryItemsReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimInventoryItemsUpdate(_ *dcim.DcimInventoryItemsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimInventoryItemsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimLocationsBulkDelete(_ *dcim.DcimLocationsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimLocationsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimLocationsBulkPartialUpdate(_ *dcim.DcimLocationsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimLocationsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimLocationsBulkUpdate(_ *dcim.DcimLocationsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimLocationsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimLocationsCreate(_ *dcim.DcimLocationsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimLocationsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimLocationsDelete(_ *dcim.DcimLocationsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimLocationsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimLocationsList(_ *dcim.DcimLocationsListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimLocationsListOK, error) {
	return nil, nil
}

func (m *mock) DcimLocationsPartialUpdate(_ *dcim.DcimLocationsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimLocationsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimLocationsRead(_ *dcim.DcimLocationsReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimLocationsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimLocationsUpdate(_ *dcim.DcimLocationsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimLocationsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimManufacturersBulkDelete(_ *dcim.DcimManufacturersBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimManufacturersBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimManufacturersBulkPartialUpdate(_ *dcim.DcimManufacturersBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimManufacturersBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimManufacturersBulkUpdate(_ *dcim.DcimManufacturersBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimManufacturersBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimManufacturersCreate(_ *dcim.DcimManufacturersCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimManufacturersCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimManufacturersDelete(_ *dcim.DcimManufacturersDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimManufacturersDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimManufacturersList(_ *dcim.DcimManufacturersListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimManufacturersListOK, error) {
	return nil, nil
}

func (m *mock) DcimManufacturersPartialUpdate(_ *dcim.DcimManufacturersPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimManufacturersPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimManufacturersRead(_ *dcim.DcimManufacturersReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimManufacturersReadOK, error) {
	return nil, nil
}

func (m *mock) DcimManufacturersUpdate(_ *dcim.DcimManufacturersUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimManufacturersUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBayTemplatesBulkDelete(_ *dcim.DcimModuleBayTemplatesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimModuleBayTemplatesBulkPartialUpdate(_ *dcim.DcimModuleBayTemplatesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBayTemplatesBulkUpdate(_ *dcim.DcimModuleBayTemplatesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBayTemplatesCreate(_ *dcim.DcimModuleBayTemplatesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimModuleBayTemplatesDelete(_ *dcim.DcimModuleBayTemplatesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimModuleBayTemplatesList(_ *dcim.DcimModuleBayTemplatesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBayTemplatesPartialUpdate(_ *dcim.DcimModuleBayTemplatesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBayTemplatesRead(_ *dcim.DcimModuleBayTemplatesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBayTemplatesUpdate(_ *dcim.DcimModuleBayTemplatesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleBayTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBaysBulkDelete(_ *dcim.DcimModuleBaysBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleBaysBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimModuleBaysBulkPartialUpdate(_ *dcim.DcimModuleBaysBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleBaysBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBaysBulkUpdate(_ *dcim.DcimModuleBaysBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleBaysBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBaysCreate(_ *dcim.DcimModuleBaysCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleBaysCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimModuleBaysDelete(_ *dcim.DcimModuleBaysDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleBaysDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimModuleBaysList(_ *dcim.DcimModuleBaysListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleBaysListOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBaysPartialUpdate(_ *dcim.DcimModuleBaysPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleBaysPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBaysRead(_ *dcim.DcimModuleBaysReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleBaysReadOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleBaysUpdate(_ *dcim.DcimModuleBaysUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleBaysUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleTypesBulkDelete(_ *dcim.DcimModuleTypesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleTypesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimModuleTypesBulkPartialUpdate(_ *dcim.DcimModuleTypesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleTypesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleTypesBulkUpdate(_ *dcim.DcimModuleTypesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleTypesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleTypesCreate(_ *dcim.DcimModuleTypesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleTypesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimModuleTypesDelete(_ *dcim.DcimModuleTypesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleTypesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimModuleTypesList(_ *dcim.DcimModuleTypesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleTypesListOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleTypesPartialUpdate(_ *dcim.DcimModuleTypesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleTypesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleTypesRead(_ *dcim.DcimModuleTypesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleTypesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimModuleTypesUpdate(_ *dcim.DcimModuleTypesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModuleTypesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModulesBulkDelete(_ *dcim.DcimModulesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModulesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimModulesBulkPartialUpdate(_ *dcim.DcimModulesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModulesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModulesBulkUpdate(_ *dcim.DcimModulesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModulesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModulesCreate(_ *dcim.DcimModulesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModulesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimModulesDelete(_ *dcim.DcimModulesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModulesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimModulesList(_ *dcim.DcimModulesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModulesListOK, error) {
	return nil, nil
}

func (m *mock) DcimModulesPartialUpdate(_ *dcim.DcimModulesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModulesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimModulesRead(_ *dcim.DcimModulesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModulesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimModulesUpdate(_ *dcim.DcimModulesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimModulesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPlatformsBulkDelete(_ *dcim.DcimPlatformsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPlatformsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPlatformsBulkPartialUpdate(_ *dcim.DcimPlatformsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPlatformsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPlatformsBulkUpdate(_ *dcim.DcimPlatformsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPlatformsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPlatformsCreate(_ *dcim.DcimPlatformsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPlatformsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimPlatformsDelete(_ *dcim.DcimPlatformsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPlatformsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPlatformsList(_ *dcim.DcimPlatformsListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPlatformsListOK, error) {
	return nil, nil
}

func (m *mock) DcimPlatformsPartialUpdate(_ *dcim.DcimPlatformsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPlatformsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPlatformsRead(_ *dcim.DcimPlatformsReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPlatformsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimPlatformsUpdate(_ *dcim.DcimPlatformsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPlatformsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsBulkDelete(_ *dcim.DcimPowerFeedsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerFeedsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsBulkPartialUpdate(_ *dcim.DcimPowerFeedsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerFeedsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsBulkUpdate(_ *dcim.DcimPowerFeedsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerFeedsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsCreate(_ *dcim.DcimPowerFeedsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerFeedsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsDelete(_ *dcim.DcimPowerFeedsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerFeedsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsList(_ *dcim.DcimPowerFeedsListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerFeedsListOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsPartialUpdate(_ *dcim.DcimPowerFeedsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerFeedsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsRead(_ *dcim.DcimPowerFeedsReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerFeedsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsTrace(_ *dcim.DcimPowerFeedsTraceParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerFeedsTraceOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerFeedsUpdate(_ *dcim.DcimPowerFeedsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerFeedsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletTemplatesBulkDelete(_ *dcim.DcimPowerOutletTemplatesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletTemplatesBulkPartialUpdate(_ *dcim.DcimPowerOutletTemplatesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletTemplatesBulkUpdate(_ *dcim.DcimPowerOutletTemplatesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletTemplatesCreate(_ *dcim.DcimPowerOutletTemplatesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletTemplatesDelete(_ *dcim.DcimPowerOutletTemplatesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletTemplatesList(_ *dcim.DcimPowerOutletTemplatesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletTemplatesPartialUpdate(_ *dcim.DcimPowerOutletTemplatesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletTemplatesRead(_ *dcim.DcimPowerOutletTemplatesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletTemplatesUpdate(_ *dcim.DcimPowerOutletTemplatesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerOutletTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsBulkDelete(_ *dcim.DcimPowerOutletsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerOutletsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsBulkPartialUpdate(_ *dcim.DcimPowerOutletsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerOutletsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsBulkUpdate(_ *dcim.DcimPowerOutletsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerOutletsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsCreate(_ *dcim.DcimPowerOutletsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerOutletsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsDelete(_ *dcim.DcimPowerOutletsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerOutletsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsList(_ *dcim.DcimPowerOutletsListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerOutletsListOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsPartialUpdate(_ *dcim.DcimPowerOutletsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerOutletsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsRead(_ *dcim.DcimPowerOutletsReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerOutletsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsTrace(_ *dcim.DcimPowerOutletsTraceParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerOutletsTraceOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerOutletsUpdate(_ *dcim.DcimPowerOutletsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerOutletsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPanelsBulkDelete(_ *dcim.DcimPowerPanelsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPanelsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerPanelsBulkPartialUpdate(_ *dcim.DcimPowerPanelsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPanelsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPanelsBulkUpdate(_ *dcim.DcimPowerPanelsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPanelsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPanelsCreate(_ *dcim.DcimPowerPanelsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPanelsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimPowerPanelsDelete(_ *dcim.DcimPowerPanelsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPanelsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerPanelsList(_ *dcim.DcimPowerPanelsListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPanelsListOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPanelsPartialUpdate(_ *dcim.DcimPowerPanelsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPanelsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPanelsRead(_ *dcim.DcimPowerPanelsReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPanelsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPanelsUpdate(_ *dcim.DcimPowerPanelsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPanelsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortTemplatesBulkDelete(_ *dcim.DcimPowerPortTemplatesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortTemplatesBulkPartialUpdate(_ *dcim.DcimPowerPortTemplatesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortTemplatesBulkUpdate(_ *dcim.DcimPowerPortTemplatesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortTemplatesCreate(_ *dcim.DcimPowerPortTemplatesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortTemplatesDelete(_ *dcim.DcimPowerPortTemplatesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortTemplatesList(_ *dcim.DcimPowerPortTemplatesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortTemplatesPartialUpdate(_ *dcim.DcimPowerPortTemplatesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortTemplatesRead(_ *dcim.DcimPowerPortTemplatesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortTemplatesUpdate(_ *dcim.DcimPowerPortTemplatesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPortTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsBulkDelete(_ *dcim.DcimPowerPortsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPortsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsBulkPartialUpdate(_ *dcim.DcimPowerPortsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPortsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsBulkUpdate(_ *dcim.DcimPowerPortsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPortsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsCreate(_ *dcim.DcimPowerPortsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPortsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsDelete(_ *dcim.DcimPowerPortsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPortsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsList(_ *dcim.DcimPowerPortsListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPortsListOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsPartialUpdate(_ *dcim.DcimPowerPortsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPortsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsRead(_ *dcim.DcimPowerPortsReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPortsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsTrace(_ *dcim.DcimPowerPortsTraceParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPortsTraceOK, error) {
	return nil, nil
}

func (m *mock) DcimPowerPortsUpdate(_ *dcim.DcimPowerPortsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimPowerPortsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRackReservationsBulkDelete(_ *dcim.DcimRackReservationsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRackReservationsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRackReservationsBulkPartialUpdate(_ *dcim.DcimRackReservationsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRackReservationsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRackReservationsBulkUpdate(_ *dcim.DcimRackReservationsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRackReservationsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRackReservationsCreate(_ *dcim.DcimRackReservationsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRackReservationsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimRackReservationsDelete(_ *dcim.DcimRackReservationsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRackReservationsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRackReservationsList(_ *dcim.DcimRackReservationsListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRackReservationsListOK, error) {
	return nil, nil
}

func (m *mock) DcimRackReservationsPartialUpdate(_ *dcim.DcimRackReservationsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRackReservationsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRackReservationsRead(_ *dcim.DcimRackReservationsReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRackReservationsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimRackReservationsUpdate(_ *dcim.DcimRackReservationsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRackReservationsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRackRolesBulkDelete(_ *dcim.DcimRackRolesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRackRolesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRackRolesBulkPartialUpdate(_ *dcim.DcimRackRolesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRackRolesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRackRolesBulkUpdate(_ *dcim.DcimRackRolesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRackRolesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRackRolesCreate(_ *dcim.DcimRackRolesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRackRolesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimRackRolesDelete(_ *dcim.DcimRackRolesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRackRolesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRackRolesList(_ *dcim.DcimRackRolesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRackRolesListOK, error) {
	return nil, nil
}

func (m *mock) DcimRackRolesPartialUpdate(_ *dcim.DcimRackRolesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRackRolesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRackRolesRead(_ *dcim.DcimRackRolesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRackRolesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimRackRolesUpdate(_ *dcim.DcimRackRolesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRackRolesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRacksBulkDelete(_ *dcim.DcimRacksBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRacksBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRacksBulkPartialUpdate(_ *dcim.DcimRacksBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRacksBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRacksBulkUpdate(_ *dcim.DcimRacksBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRacksBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRacksCreate(_ *dcim.DcimRacksCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRacksCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimRacksDelete(_ *dcim.DcimRacksDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRacksDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRacksElevation(_ *dcim.DcimRacksElevationParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRacksElevationOK, error) {
	return nil, nil
}

func (m *mock) DcimRacksList(_ *dcim.DcimRacksListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRacksListOK, error) {
	return nil, nil
}

func (m *mock) DcimRacksPartialUpdate(_ *dcim.DcimRacksPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRacksPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRacksRead(_ *dcim.DcimRacksReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRacksReadOK, error) {
	return nil, nil
}

func (m *mock) DcimRacksUpdate(_ *dcim.DcimRacksUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRacksUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortTemplatesBulkDelete(_ *dcim.DcimRearPortTemplatesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRearPortTemplatesBulkPartialUpdate(_ *dcim.DcimRearPortTemplatesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortTemplatesBulkUpdate(_ *dcim.DcimRearPortTemplatesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortTemplatesCreate(_ *dcim.DcimRearPortTemplatesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimRearPortTemplatesDelete(_ *dcim.DcimRearPortTemplatesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRearPortTemplatesList(_ *dcim.DcimRearPortTemplatesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortTemplatesPartialUpdate(_ *dcim.DcimRearPortTemplatesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortTemplatesRead(_ *dcim.DcimRearPortTemplatesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortTemplatesUpdate(_ *dcim.DcimRearPortTemplatesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRearPortTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsBulkDelete(_ *dcim.DcimRearPortsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRearPortsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsBulkPartialUpdate(_ *dcim.DcimRearPortsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRearPortsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsBulkUpdate(_ *dcim.DcimRearPortsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRearPortsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsCreate(_ *dcim.DcimRearPortsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRearPortsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsDelete(_ *dcim.DcimRearPortsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRearPortsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsList(_ *dcim.DcimRearPortsListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRearPortsListOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsPartialUpdate(_ *dcim.DcimRearPortsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRearPortsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsPaths(_ *dcim.DcimRearPortsPathsParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRearPortsPathsOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsRead(_ *dcim.DcimRearPortsReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRearPortsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimRearPortsUpdate(_ *dcim.DcimRearPortsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRearPortsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRegionsBulkDelete(_ *dcim.DcimRegionsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRegionsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRegionsBulkPartialUpdate(_ *dcim.DcimRegionsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRegionsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRegionsBulkUpdate(_ *dcim.DcimRegionsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRegionsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRegionsCreate(_ *dcim.DcimRegionsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRegionsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimRegionsDelete(_ *dcim.DcimRegionsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRegionsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimRegionsList(_ *dcim.DcimRegionsListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRegionsListOK, error) {
	return nil, nil
}

func (m *mock) DcimRegionsPartialUpdate(_ *dcim.DcimRegionsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRegionsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimRegionsRead(_ *dcim.DcimRegionsReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRegionsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimRegionsUpdate(_ *dcim.DcimRegionsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimRegionsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimSiteGroupsBulkDelete(_ *dcim.DcimSiteGroupsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimSiteGroupsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimSiteGroupsBulkPartialUpdate(_ *dcim.DcimSiteGroupsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimSiteGroupsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimSiteGroupsBulkUpdate(_ *dcim.DcimSiteGroupsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimSiteGroupsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimSiteGroupsCreate(_ *dcim.DcimSiteGroupsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimSiteGroupsCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimSiteGroupsDelete(_ *dcim.DcimSiteGroupsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimSiteGroupsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimSiteGroupsList(_ *dcim.DcimSiteGroupsListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimSiteGroupsListOK, error) {
	return nil, nil
}

func (m *mock) DcimSiteGroupsPartialUpdate(_ *dcim.DcimSiteGroupsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimSiteGroupsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimSiteGroupsRead(_ *dcim.DcimSiteGroupsReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimSiteGroupsReadOK, error) {
	return nil, nil
}

func (m *mock) DcimSiteGroupsUpdate(_ *dcim.DcimSiteGroupsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimSiteGroupsUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimSitesBulkDelete(_ *dcim.DcimSitesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimSitesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimSitesBulkPartialUpdate(_ *dcim.DcimSitesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimSitesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimSitesBulkUpdate(_ *dcim.DcimSitesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimSitesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimSitesCreate(_ *dcim.DcimSitesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimSitesCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimSitesDelete(_ *dcim.DcimSitesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimSitesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimSitesList(_ *dcim.DcimSitesListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimSitesListOK, error) {
	return nil, nil
}

func (m *mock) DcimSitesPartialUpdate(_ *dcim.DcimSitesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimSitesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimSitesRead(_ *dcim.DcimSitesReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimSitesReadOK, error) {
	return nil, nil
}

func (m *mock) DcimSitesUpdate(_ *dcim.DcimSitesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimSitesUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimVirtualChassisBulkDelete(_ *dcim.DcimVirtualChassisBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimVirtualChassisBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimVirtualChassisBulkPartialUpdate(_ *dcim.DcimVirtualChassisBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimVirtualChassisBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimVirtualChassisBulkUpdate(_ *dcim.DcimVirtualChassisBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimVirtualChassisBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimVirtualChassisCreate(_ *dcim.DcimVirtualChassisCreateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimVirtualChassisCreateCreated, error) {
	return nil, nil
}

func (m *mock) DcimVirtualChassisDelete(_ *dcim.DcimVirtualChassisDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimVirtualChassisDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) DcimVirtualChassisList(_ *dcim.DcimVirtualChassisListParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimVirtualChassisListOK, error) {
	return nil, nil
}

func (m *mock) DcimVirtualChassisPartialUpdate(_ *dcim.DcimVirtualChassisPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimVirtualChassisPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) DcimVirtualChassisRead(_ *dcim.DcimVirtualChassisReadParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimVirtualChassisReadOK, error) {
	return nil, nil
}

func (m *mock) DcimVirtualChassisUpdate(_ *dcim.DcimVirtualChassisUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...dcim.ClientOption) (*dcim.DcimVirtualChassisUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamAggregatesBulkDelete(_ *ipam.IpamAggregatesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamAggregatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamAggregatesBulkPartialUpdate(_ *ipam.IpamAggregatesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamAggregatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamAggregatesBulkUpdate(_ *ipam.IpamAggregatesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamAggregatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamAggregatesCreate(_ *ipam.IpamAggregatesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamAggregatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamAggregatesDelete(_ *ipam.IpamAggregatesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamAggregatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamAggregatesList(_ *ipam.IpamAggregatesListParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamAggregatesListOK, error) {
	return nil, nil
}

func (m *mock) IpamAggregatesPartialUpdate(_ *ipam.IpamAggregatesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamAggregatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamAggregatesRead(_ *ipam.IpamAggregatesReadParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamAggregatesReadOK, error) {
	return nil, nil
}

func (m *mock) IpamAggregatesUpdate(_ *ipam.IpamAggregatesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamAggregatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamAsnsBulkDelete(_ *ipam.IpamAsnsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamAsnsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamAsnsBulkPartialUpdate(_ *ipam.IpamAsnsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamAsnsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamAsnsBulkUpdate(_ *ipam.IpamAsnsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamAsnsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamAsnsCreate(_ *ipam.IpamAsnsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamAsnsCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamAsnsDelete(_ *ipam.IpamAsnsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamAsnsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamAsnsList(_ *ipam.IpamAsnsListParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamAsnsListOK, error) {
	return nil, nil
}

func (m *mock) IpamAsnsPartialUpdate(_ *ipam.IpamAsnsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamAsnsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamAsnsRead(_ *ipam.IpamAsnsReadParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamAsnsReadOK, error) {
	return nil, nil
}

func (m *mock) IpamAsnsUpdate(_ *ipam.IpamAsnsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamAsnsUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupAssignmentsBulkDelete(_ *ipam.IpamFhrpGroupAssignmentsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamFhrpGroupAssignmentsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupAssignmentsBulkPartialUpdate(_ *ipam.IpamFhrpGroupAssignmentsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamFhrpGroupAssignmentsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupAssignmentsBulkUpdate(_ *ipam.IpamFhrpGroupAssignmentsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamFhrpGroupAssignmentsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupAssignmentsCreate(_ *ipam.IpamFhrpGroupAssignmentsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamFhrpGroupAssignmentsCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupAssignmentsDelete(_ *ipam.IpamFhrpGroupAssignmentsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamFhrpGroupAssignmentsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupAssignmentsList(_ *ipam.IpamFhrpGroupAssignmentsListParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamFhrpGroupAssignmentsListOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupAssignmentsPartialUpdate(_ *ipam.IpamFhrpGroupAssignmentsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamFhrpGroupAssignmentsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupAssignmentsRead(_ *ipam.IpamFhrpGroupAssignmentsReadParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamFhrpGroupAssignmentsReadOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupAssignmentsUpdate(_ *ipam.IpamFhrpGroupAssignmentsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamFhrpGroupAssignmentsUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupsBulkDelete(_ *ipam.IpamFhrpGroupsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamFhrpGroupsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupsBulkPartialUpdate(_ *ipam.IpamFhrpGroupsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamFhrpGroupsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupsBulkUpdate(_ *ipam.IpamFhrpGroupsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamFhrpGroupsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupsCreate(_ *ipam.IpamFhrpGroupsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamFhrpGroupsCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupsDelete(_ *ipam.IpamFhrpGroupsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamFhrpGroupsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupsList(_ *ipam.IpamFhrpGroupsListParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamFhrpGroupsListOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupsPartialUpdate(_ *ipam.IpamFhrpGroupsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamFhrpGroupsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupsRead(_ *ipam.IpamFhrpGroupsReadParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamFhrpGroupsReadOK, error) {
	return nil, nil
}

func (m *mock) IpamFhrpGroupsUpdate(_ *ipam.IpamFhrpGroupsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamFhrpGroupsUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamIPAddressesBulkDelete(_ *ipam.IpamIPAddressesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPAddressesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamIPAddressesBulkPartialUpdate(_ *ipam.IpamIPAddressesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPAddressesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamIPAddressesBulkUpdate(_ *ipam.IpamIPAddressesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPAddressesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamIPAddressesCreate(_ *ipam.IpamIPAddressesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPAddressesCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamIPAddressesDelete(_ *ipam.IpamIPAddressesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPAddressesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamIPAddressesList(_ *ipam.IpamIPAddressesListParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPAddressesListOK, error) {
	return nil, nil
}

func (m *mock) IpamIPAddressesPartialUpdate(_ *ipam.IpamIPAddressesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPAddressesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamIPAddressesRead(_ *ipam.IpamIPAddressesReadParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPAddressesReadOK, error) {
	return nil, nil
}

func (m *mock) IpamIPAddressesUpdate(_ *ipam.IpamIPAddressesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPAddressesUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesAvailableIpsCreate(_ *ipam.IpamIPRangesAvailableIpsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPRangesAvailableIpsCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesAvailableIpsList(_ *ipam.IpamIPRangesAvailableIpsListParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPRangesAvailableIpsListOK, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesBulkDelete(_ *ipam.IpamIPRangesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPRangesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesBulkPartialUpdate(_ *ipam.IpamIPRangesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPRangesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesBulkUpdate(_ *ipam.IpamIPRangesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPRangesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesCreate(_ *ipam.IpamIPRangesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPRangesCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesDelete(_ *ipam.IpamIPRangesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPRangesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesList(_ *ipam.IpamIPRangesListParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPRangesListOK, error) {
	return m.IP, nil
}

func (m *mock) IpamIPRangesPartialUpdate(_ *ipam.IpamIPRangesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPRangesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesRead(_ *ipam.IpamIPRangesReadParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPRangesReadOK, error) {
	return nil, nil
}

func (m *mock) IpamIPRangesUpdate(_ *ipam.IpamIPRangesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamIPRangesUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesAvailableIpsCreate(_ *ipam.IpamPrefixesAvailableIpsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamPrefixesAvailableIpsCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesAvailableIpsList(_ *ipam.IpamPrefixesAvailableIpsListParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamPrefixesAvailableIpsListOK, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesAvailablePrefixesCreate(_ *ipam.IpamPrefixesAvailablePrefixesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamPrefixesAvailablePrefixesCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesAvailablePrefixesList(_ *ipam.IpamPrefixesAvailablePrefixesListParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamPrefixesAvailablePrefixesListOK, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesBulkDelete(_ *ipam.IpamPrefixesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamPrefixesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesBulkPartialUpdate(_ *ipam.IpamPrefixesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamPrefixesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesBulkUpdate(_ *ipam.IpamPrefixesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamPrefixesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesCreate(_ *ipam.IpamPrefixesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamPrefixesCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesDelete(_ *ipam.IpamPrefixesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamPrefixesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesList(_ *ipam.IpamPrefixesListParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamPrefixesListOK, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesPartialUpdate(_ *ipam.IpamPrefixesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamPrefixesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesRead(_ *ipam.IpamPrefixesReadParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamPrefixesReadOK, error) {
	return nil, nil
}

func (m *mock) IpamPrefixesUpdate(_ *ipam.IpamPrefixesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamPrefixesUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRirsBulkDelete(_ *ipam.IpamRirsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRirsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamRirsBulkPartialUpdate(_ *ipam.IpamRirsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRirsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRirsBulkUpdate(_ *ipam.IpamRirsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRirsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRirsCreate(_ *ipam.IpamRirsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRirsCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamRirsDelete(_ *ipam.IpamRirsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRirsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamRirsList(_ *ipam.IpamRirsListParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRirsListOK, error) {
	return nil, nil
}

func (m *mock) IpamRirsPartialUpdate(_ *ipam.IpamRirsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRirsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRirsRead(_ *ipam.IpamRirsReadParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRirsReadOK, error) {
	return nil, nil
}

func (m *mock) IpamRirsUpdate(_ *ipam.IpamRirsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRirsUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRolesBulkDelete(_ *ipam.IpamRolesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRolesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamRolesBulkPartialUpdate(_ *ipam.IpamRolesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRolesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRolesBulkUpdate(_ *ipam.IpamRolesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRolesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRolesCreate(_ *ipam.IpamRolesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRolesCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamRolesDelete(_ *ipam.IpamRolesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRolesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamRolesList(_ *ipam.IpamRolesListParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRolesListOK, error) {
	return nil, nil
}

func (m *mock) IpamRolesPartialUpdate(_ *ipam.IpamRolesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRolesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRolesRead(_ *ipam.IpamRolesReadParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRolesReadOK, error) {
	return nil, nil
}

func (m *mock) IpamRolesUpdate(_ *ipam.IpamRolesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRolesUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRouteTargetsBulkDelete(_ *ipam.IpamRouteTargetsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRouteTargetsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamRouteTargetsBulkPartialUpdate(_ *ipam.IpamRouteTargetsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRouteTargetsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRouteTargetsBulkUpdate(_ *ipam.IpamRouteTargetsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRouteTargetsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRouteTargetsCreate(_ *ipam.IpamRouteTargetsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRouteTargetsCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamRouteTargetsDelete(_ *ipam.IpamRouteTargetsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRouteTargetsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamRouteTargetsList(_ *ipam.IpamRouteTargetsListParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRouteTargetsListOK, error) {
	return nil, nil
}

func (m *mock) IpamRouteTargetsPartialUpdate(_ *ipam.IpamRouteTargetsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRouteTargetsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamRouteTargetsRead(_ *ipam.IpamRouteTargetsReadParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRouteTargetsReadOK, error) {
	return nil, nil
}

func (m *mock) IpamRouteTargetsUpdate(_ *ipam.IpamRouteTargetsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamRouteTargetsUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamServiceTemplatesBulkDelete(_ *ipam.IpamServiceTemplatesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamServiceTemplatesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamServiceTemplatesBulkPartialUpdate(_ *ipam.IpamServiceTemplatesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamServiceTemplatesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamServiceTemplatesBulkUpdate(_ *ipam.IpamServiceTemplatesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamServiceTemplatesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamServiceTemplatesCreate(_ *ipam.IpamServiceTemplatesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamServiceTemplatesCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamServiceTemplatesDelete(_ *ipam.IpamServiceTemplatesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamServiceTemplatesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamServiceTemplatesList(_ *ipam.IpamServiceTemplatesListParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamServiceTemplatesListOK, error) {
	return nil, nil
}

func (m *mock) IpamServiceTemplatesPartialUpdate(_ *ipam.IpamServiceTemplatesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamServiceTemplatesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamServiceTemplatesRead(_ *ipam.IpamServiceTemplatesReadParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamServiceTemplatesReadOK, error) {
	return nil, nil
}

func (m *mock) IpamServiceTemplatesUpdate(_ *ipam.IpamServiceTemplatesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamServiceTemplatesUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamServicesBulkDelete(_ *ipam.IpamServicesBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamServicesBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamServicesBulkPartialUpdate(_ *ipam.IpamServicesBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamServicesBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamServicesBulkUpdate(_ *ipam.IpamServicesBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamServicesBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamServicesCreate(_ *ipam.IpamServicesCreateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamServicesCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamServicesDelete(_ *ipam.IpamServicesDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamServicesDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamServicesList(_ *ipam.IpamServicesListParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamServicesListOK, error) {
	return nil, nil
}

func (m *mock) IpamServicesPartialUpdate(_ *ipam.IpamServicesPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamServicesPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamServicesRead(_ *ipam.IpamServicesReadParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamServicesReadOK, error) {
	return nil, nil
}

func (m *mock) IpamServicesUpdate(_ *ipam.IpamServicesUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamServicesUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsAvailableVlansCreate(_ *ipam.IpamVlanGroupsAvailableVlansCreateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlanGroupsAvailableVlansCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsAvailableVlansList(_ *ipam.IpamVlanGroupsAvailableVlansListParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlanGroupsAvailableVlansListOK, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsBulkDelete(_ *ipam.IpamVlanGroupsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlanGroupsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsBulkPartialUpdate(_ *ipam.IpamVlanGroupsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlanGroupsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsBulkUpdate(_ *ipam.IpamVlanGroupsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlanGroupsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsCreate(_ *ipam.IpamVlanGroupsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlanGroupsCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsDelete(_ *ipam.IpamVlanGroupsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlanGroupsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsList(_ *ipam.IpamVlanGroupsListParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlanGroupsListOK, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsPartialUpdate(_ *ipam.IpamVlanGroupsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlanGroupsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsRead(_ *ipam.IpamVlanGroupsReadParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlanGroupsReadOK, error) {
	return nil, nil
}

func (m *mock) IpamVlanGroupsUpdate(_ *ipam.IpamVlanGroupsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlanGroupsUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVlansBulkDelete(_ *ipam.IpamVlansBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlansBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamVlansBulkPartialUpdate(_ *ipam.IpamVlansBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlansBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVlansBulkUpdate(_ *ipam.IpamVlansBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlansBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVlansCreate(_ *ipam.IpamVlansCreateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlansCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamVlansDelete(_ *ipam.IpamVlansDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlansDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamVlansList(_ *ipam.IpamVlansListParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlansListOK, error) {
	return nil, nil
}

func (m *mock) IpamVlansPartialUpdate(_ *ipam.IpamVlansPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlansPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVlansRead(_ *ipam.IpamVlansReadParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlansReadOK, error) {
	return nil, nil
}

func (m *mock) IpamVlansUpdate(_ *ipam.IpamVlansUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVlansUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVrfsBulkDelete(_ *ipam.IpamVrfsBulkDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVrfsBulkDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamVrfsBulkPartialUpdate(_ *ipam.IpamVrfsBulkPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVrfsBulkPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVrfsBulkUpdate(_ *ipam.IpamVrfsBulkUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVrfsBulkUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVrfsCreate(_ *ipam.IpamVrfsCreateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVrfsCreateCreated, error) {
	return nil, nil
}

func (m *mock) IpamVrfsDelete(_ *ipam.IpamVrfsDeleteParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVrfsDeleteNoContent, error) {
	return nil, nil
}

func (m *mock) IpamVrfsList(_ *ipam.IpamVrfsListParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVrfsListOK, error) {
	return nil, nil
}

func (m *mock) IpamVrfsPartialUpdate(_ *ipam.IpamVrfsPartialUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVrfsPartialUpdateOK, error) {
	return nil, nil
}

func (m *mock) IpamVrfsRead(_ *ipam.IpamVrfsReadParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVrfsReadOK, error) {
	return nil, nil
}

func (m *mock) IpamVrfsUpdate(_ *ipam.IpamVrfsUpdateParams, _ runtime.ClientAuthInfoWriter, _ ...ipam.ClientOption) (*ipam.IpamVrfsUpdateOK, error) {
	return nil, nil
}

func (m *mock) SetTransport(_ runtime.ClientTransport) {
}
