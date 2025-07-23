---
title: "Preparing Bare Metal for EKS Anywhere"
linkTitle: "2. Prepare hardware"
weight: 20
aliases:
    /docs/reference/baremetal/bare-preparation/
description: >
  Set up a Bare Metal cluster to prepare it for EKS Anywhere
---
After gathering hardware described in Bare Metal [Requirements]({{< relref "./bare-prereq" >}}), you need to prepare the hardware and create a CSV file describing that hardware.

## Prepare hardware
To prepare your computer hardware for EKS Anywhere, you need to connect your computer hardware and do some configuration.
Once the hardware is in place, you need to:

* Obtain IP and MAC addresses for your machines' NICs.
* Obtain IP addresses for your machines' BMC interfaces.
* Obtain the gateway address for your network to reach the Internet.
* Obtain the IP address for your DNS servers.
* Make sure the following settings are in place:
  * UEFI is enabled on all target cluster machines, unless you are provisioning RHEL systems. Enable legacy BIOS on any RHEL machines.
  * Netboot (PXE or HTTP) boot is enabled for the NIC on each machine for which you provided the MAC address. This is the interface on which the operating system will be provisioned.
  * IPMI over LAN and/or Redfish is enabled on all BMC interfaces.
* Go to the BMC settings for each machine and set the IP address (bmc_ip), username (bmc_username), and password (bmc_password) to use later in the CSV file.

## Prepare hardware inventory
Create a CSV file to provide information about all physical machines that you are ready to add to your target Bare Metal cluster.
This file will be used:

* When you generate the hardware file to be included in the cluster creation process described in the Create Bare Metal production cluster Getting Started guide.
* To provide information that is passed to each machine from the Tinkerbell DHCP server when the machine is initially network booted.

**NOTE**:While using kubectl, GitOps and Terraform for workload cluster creation, please make sure to refer [this]({{< relref "./baremetal-getstarted/#create-separate-workload-clusters" >}}) section.

The following is an example of an EKS Anywhere Bare Metal hardware CSV file:

```
hostname,bmc_ip,bmc_username,bmc_password,mac,ip_address,netmask,gateway,nameservers,labels,disk
eksa-cp01,10.10.44.1,root,PrZ8W93i,CC:48:3A:00:00:01,10.10.50.2,255.255.254.0,10.10.50.1,8.8.8.8|8.8.4.4,type=cp,/dev/sda
eksa-cp02,10.10.44.2,root,Me9xQf93,CC:48:3A:00:00:02,10.10.50.3,255.255.254.0,10.10.50.1,8.8.8.8|8.8.4.4,type=cp,/dev/sda
eksa-cp03,10.10.44.3,root,Z8x2M6hl,CC:48:3A:00:00:03,10.10.50.4,255.255.254.0,10.10.50.1,8.8.8.8|8.8.4.4,type=cp,/dev/sda
eksa-wk01,10.10.44.4,root,B398xRTp,CC:48:3A:00:00:04,10.10.50.5,255.255.254.0,10.10.50.1,8.8.8.8|8.8.4.4,type=worker,/dev/sda
eksa-wk02,10.10.44.5,root,w7EenR94,CC:48:3A:00:00:05,10.10.50.6,255.255.254.0,10.10.50.1,8.8.8.8|8.8.4.4,type=worker,/dev/sda

```

The CSV file is a comma-separated list of values in a plain text file, holding information about the physical machines in the datacenter that are intended to be a part of the cluster creation process.
Each line represents a physical machine (not a virtual machine).

The following sections describe each value.

### hostname
The hostname assigned to the machine.
### bmc_ip (optional)
The IP address assigned to the BMC interface on the machine.
### bmc_username (optional)
The username assigned to the BMC interface on the machine.
### bmc_password (optional)
The password associated with the `bmc_username` assigned to the BMC interface on the machine.
### mac
The MAC address of the network interface card (NIC) that provides access to the host computer.
### ip_address
The IP address providing access to the host computer.
### netmask
The netmask associated with the `ip_address` value.
In the example above, a /23 subnet mask is used, allowing you to use up to 510 IP addresses in that range. 
### gateway
IP address of the interface that provides access (the gateway) to the Internet.
### nameservers
The IP address of the server that you want to provide DNS service to the cluster.
### labels
The optional labels field can consist of a key/value pair to use in conjunction with the `hardwareSelector` field when you set up your Bare Metal configuration.
The key/value pair is connected with an equal (`=`) sign.

For example, a `TinkerbellMachineConfig` with a `hardwareSelector` containing `type: cp` will match entries in the CSV containing `type=cp` in its label definition.

### disk
The device name of the disk on which the operating system will be installed.
For example, it could be `/dev/sda` for the first SCSI disk or `/dev/nvme0n1` for the first NVME storage device.

## Hardware Management 

### Hardware Objects and Spare Nodes
EKS Anywhere creates Kubernetes hardware objects for all entries in the `hardware.csv` file, regardless of the node count specified in your cluster configuration. Any machines in the CSV file that exceed the count specified in your cluster configuration are automatically treated as spare nodes.

For example, if your cluster configuration specifies:
```yaml
controlPlaneConfiguration:
  count: 1
workerNodeGroupConfigurations:
- count: 1
```

But your `hardware.csv` contains 3 control plane entries and 3 worker entries:
```
cplane-0,XX:XX:XX:XX:XX:01,10.162.10.131,...
cplane-1,XX:XX:XX:XX:XX:02,10.162.10.132,...
cplane-2,XX:XX:XX:XX:XX:03,10.162.10.133,...
worker-0,XX:XX:XX:XX:XX:04,10.162.10.134,...
worker-1,XX:XX:XX:XX:XX:05,10.162.10.135,...
worker-2,XX:XX:XX:XX:XX:06,10.162.10.136,...
```

EKS Anywhere will create Hardware objects for all 6 machines, but initially only use `cplane-0` and `worker-0` for the cluster. The remaining machines (`cplane-1`, `cplane-2`, `worker-1`, and `worker-2`) will be treated as spare nodes.

### Important Considerations for Hardware Lifecycle Management
During operations like rolling upgrades, EKS Anywhere's selection of spare nodes is not deterministic. Any of the available spare nodes could be selected for use during maintenance operations.
If you repurpose machines that were previously part of your EKS Anywhere cluster but still have hardware objects in the system:

1. During subsequent cluster operations (like upgrades), these repurposed machines could be selected as spare nodes
2. If BMC credentials are still valid, the machine may be automatically re-imaged, potentially destroying any workloads running on the repurposed machine

#### Best Practices for Hardware Management

To avoid unintended consequences when managing your hardware:

1. **Align CSV with Intended Use**: Ensure your hardware.csv only contains entries for machines intended for immediate cluster use or as designated spares
2. **Clean Up Hardware Objects**: Regularly audit and remove Hardware objects for machines that have been repurposed for other uses
3. **Manage BMC Credentials**: Remove or rotate BMC credentials for hardware that has been repurposed to prevent unintended re-imaging
4. **Hardware Inventory**: Maintain an up-to-date inventory of which physical machines are currently part of the cluster, designated as spares, or repurposed for other workloads
