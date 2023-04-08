---
title: "DHCP options for EKS Anywhere"
linkTitle: "Use an existing DHCP service"
weight: 35
description: >
  Using your existing DHCP service with EKS Anywhere Bare Metal
---

In order to facilitate network booting machines, EKS Anywhere bare metal runs its own DHCP server, Boots (a standalone service in the Tinkerbell stack). In environments where there is an existing DHCP service, this DHCP service can be configured to interoperate with EKS Anywhere. This document will cover how to make your existing DHCP service interoperate with EKS Anywhere bare metal. In this scenario EKS Anywhere will have no layer 2 DHCP responsibilities. It is important to note that currently, Boots is responsible for more than just DHCP. So Boots can't be entirely avoided in the provisioning process.

#### Additional Services in Boots

- serving iPXE binaries via HTTP and TFTP
- serving an iPXE script via HTTP
- functions as a SYSLOG server (receiver)

## Process

The first step is to configure your existing DHCP to serve [host/address/static reservations](https://kb.isc.org/docs/what-are-host-reservations-how-to-use-them) for all machines that EKS Anywhere bare metal will be provisioning. This means that the IPAM details you enter into the [`hardware.csv`]({{< relref "./bare-preparation#prepare-hardware-inventory" >}}) must be used to create host/address/static reservations in your existing DHCP service. The second step is to configure your existing DHCP to serve the location of the iPXE binary and script used by EKS Anywhere.

There is a 2 step interaction between a netboot client and a DHCP service in order to kick off the provisioning process.

- __Step 1__: The machine broadcasts a request to network boot. The DHCP service then provides the machine with all IPAM info as well as the location of the Tinkerbell iPXE binary. The machine configures its network interface with the IPAM info then downloads the Tinkerbell iPXE binary from the location provided by the DHCP service and runs it.

- __Step 2__: Now in the Tinkerbell iPXE binary, iPXE broadcasts a request to network boot. The DHCP service again provides all IPAM info as well as provides the location of the Tinkerbell iPXE script `auto.ipxe` (see note below). iPXE configures its network interface using the IPAM info and then downloads the Tinkerbell iPXE script from the location provided by the DHCP service and runs it.

>Note The `auto.ipxe` is an [iPXE script](https://ipxe.org/scripting) that tells iPXE from where to download the [HookOS]({{< relref "./bare-custom-hookos" >}}) kernel and initrd so that they can be loaded into memory.

The following diagram illustrates the process described above. Note that the diagram only describes the network booting parts of the DHCP interaction, not the exchange of IPAM info.

![process](/images/eksa-baremetal-bring-your-own-dhcp.png)

## Configuration

Below you will find code snippets showing how to add the 2 step process from above to an existing DHCP service. Each config checks if DHCP option 77 ([user class option](https://www.rfc-editor.org/rfc/rfc3004.html)) equals "`Tinkerbell`". If it does match, then the Tinkerbell iPXE script (`auto.ipxe`) will be served. If option 77 does not match, then the iPXE binary will be served.

### DHCP option: `next server`

Most DHCP services define a `next server` option. This option generally corresponds to either DHCP option 66 or the DHCP header `sname`, [reference](https://www.rfc-editor.org/rfc/rfc2132.html#section-9.4).

Special consideration is required when using EKS Anywhere to create your initial [management cluster]({{< relref "../../concepts/cluster-topologies" >}}). This is because during this initial create phase a temporary [bootstrap cluster]({{< relref "../../concepts/cluster-topologies#whats-the-difference-between-a-management-cluster-and-a-bootstrap-cluster-for-eks-anywhere" >}}) is created and used to provision the management cluster.

As a temporary and one time step, the IP address used by the existing DHCP service for `next server` will need to be the IP address of the temporary bootstrap cluster. This will be the IP of the [Admin node]({{< relref "../../getting-started/install#administrative-machine-prerequisites" >}}) or the if you use the cli flag [`--tinkerbell-bootstrap-ip`]({{< relref "../eksctl/anywhere_create_cluster#options" >}}) then that IP should be used for the `next server` in your existing DHCP service.

Once the management cluster is created, the IP address used by the existing DHCP service for `next server` will need to be updated to the `tinkerbellIP`. This IP is defined in your cluster spec at [`tinkerbellDatacenterConfig.spec.tinkerbellIP`]({{< relref "../clusterspec/baremetal#example-tinkerbelldatacenterconfigspec" >}}). The `next server` IP will not need to be updated again.

>Note: The upgrade phase of a management cluster or the creation of any [workload clusters]({{< relref "../../concepts/cluster-topologies" >}}) will not require you to change the `next server` IP in the existing DHCP service config.

### Code snippets

[dnsmasq](https://linux.die.net/man/8/dnsmasq)

dnsmasq.conf

```text
dhcp-match=tinkerbell, option:user-class, Tinkerbell
dhcp-boot=tag:!tinkerbell,ipxe.efi,none,192.168.2.112
dhcp-boot=tag:tinkerbell,http://192.168.2.112/auto.ipxe
```

[Kea DHCP](https://www.isc.org/kea/)

kea.json

```json
{
    "Dhcp4": {
        "client-classes": [
            {
                "name": "tinkerbell",
                "test": "substring(option[77].hex,0,10) == 'Tinkerbell'",
                "boot-file-name": "http://192.168.2.112/auto.ipxe"
            },
            {
                "name": "default",
                "test": "not(substring(option[77].hex,0,10) == 'Tinkerbell')",
                "boot-file-name": "ipxe.efi"
            }
        ],
        "subnet4": [
            {
                "next-server": "192.168.2.112"
            }
        ]
    }
}
```

[ISC DHCP](https://ipxe.org/howto/dhcpd)

dhcpd.conf

```text
 if exists user-class and option user-class = "Tinkerbell" {
     filename "http://192.168.2.112/auto.ipxe";
 } else {
     filename "ipxe.efi";
 }
 next-server "192.168.1.112";
```

[Microsoft DHCP server](https://learn.microsoft.com/en-us/windows-server/networking/technologies/dhcp/dhcp-top)

Please follow the ipxe.org [guide](https://ipxe.org/howto/msdhcp) on how to configure Microsoft DHCP server.
