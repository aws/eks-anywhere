---
title: "Preparing Bare Metal for EKS Anywhere"
linkTitle: "Preparing Bare Metal"
weight: 20
description: >
  Set up a Bare Metal cluster to prepare it for EKS Anywhere
---
After gathering hardware described in Bare Metal [Requirements]({{< relref "./bare-prereq" >}}), you need to prepare the hardware and create a CSV file describing that hardware.

## Prepare hardware
To prepare your computer hardware for EKS Anywhere, you need to connect your computer hardware and do some configuration.
The following figure provides an example of an EKS Anywhere Bare Metal cluster consisting of three control plane and two worker machines:

**_NEED TO ADD A HARDWARE DIAGRAM AND SHORT DESCRIPTION OF EACH COMPONENT._**

Once the hardware is in place, you need to:

* Obtain IP addresses to assign to the IPMI and regular NIC interface on each machine in the cluster.
* Obtain the gateway address for your network to reach the Internet.
* Obtain the IP address for your DNS servers.
* Make sure the following settings are in place:
  * UEFI is enabled on all target cluster machines
  * PXE boot is enabled on each NIC
  * PXE is set as the first device in each machine's boot order
  * IPMI over LAN is enabled on the IPMI interfaces
* Go to the IPMI settings for each machine and set the IP address (bmc_ip), username (bmc_username), and password (bmc_password) to use later in the CSV file.

## Prepare hardware inventory
Create a CSV file to provide information about all physical machines that you are ready to add to your target Bare Metal cluster.
This file will be used:

* When you generate the hardware file to be included in the cluster creation process described in the [Create Bare Metal production cluster]({{< relref "../../getting-started/production-environment" >}}) Getting Started guide.
* To provide information that is passed to each machine from the Tinkerbell DHCP server when the machine is initially PXE booted.

The following is an example of an EKS Anywhere Bare Metal hardware CSV file:

```
hostname,bmc_ip,bmc_username,bmc_password,mac,ip_address,netmask,gateway,nameservers,labels,disk
eksa-cp01,10.10.44.1,root,PrZ8W93i,CC:48:3A:00:00:01,10.10.50.2,255.255.254.0,10.10.50.1,X.X.X.X,,/dev/sda
eksa-cp02,10.10.44.2,root,Me9xQf93,CC:48:3A:00:00:02,10.10.50.3,255.255.254.0,10.10.50.1,X.X.X.X,,/dev/sda
eksa-wk01,10.10.44.3,root,Z8x2M6hl,CC:48:3A:00:00:03,10.10.50.4,255.255.254.0,10.10.50.1,X.X.X.X,,/dev/sda
eksa-wk02,10.10.44.4,root,B398xRTp,CC:48:3A:00:00:04,10.10.50.5,255.255.254.0,10.10.50.1,X.X.X.X,,/dev/sda
eksa-wk03,10.10.44.5,root,w7EenR94,CC:48:3A:00:00:05,10.10.50.6,255.255.254.0,10.10.50.1,X.X.X.X,,/dev/sda

```

The CSV file is a comma-separated list of values in a plain text file, holding information about the physical machines in the datacenter that are intended to be a part of the cluster creation process.
Each line represents a physical machine (not a virtual machine).
Multiple values in an item can be separated by pipe symbols (|).

The following sections describe each value.

### hostname
The hostname assigned to the machine.
### bmc_ip
The IP address assigned to the IPMI interface on the machine.
### bmc_username
The username assigned to the IPMI interface on the machine.
### bmc_password
The password associated with the `bmc_username` assigned to the IPMI interface on the machine.
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
Optional labels you can set on each node that are added to any labels set on that node by the cluster.
These labels consist of key/value pairs that can be used by Kubernetes to match workloads that request nodes with those labels.
See Kubernetes [Labels and Selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) for details.

Labels can also be used in conjunction with the `hardwareSelector` field when you set up your [Bare Metal configuration]({{< relref "../clusterspec/baremetal/" >}}).
For example, if you had the `controlPlaneConfiguration.machineGroupRef.TinkerbellMachineConfig` name set to `my-cluster-name-cp` and a `hardwareSelector` under the `TinkerbellMachineConfig` definition for that name set to `node: "cp-machine"`, that machine would be added to the control plane.

In this CSV field, each key/value pair is connected with an equal (`=`) sign.
For example, you could set `environment=test` on one node and `environment=production` on another to allow workloads to request the appropriate node to run on based on their stage of development.

Use a pipe (|) symbol to add multiple labels to one field. For example, to set a machine as a worker node to be used for production workloads, the field might appear as follows: `node=worker-node|environment=production`.
In this way, the cluster specification could be set to recognize this machine as a worker node and workloads could later request to run on a production system by matching that label in its pod spec.

### disk
The device name of the disk on which the operating system will be installed.
For example, it could be `/dev/sda` for the first SCSI disk or `/dev/nvme0n1` for the first NVME storage device.
