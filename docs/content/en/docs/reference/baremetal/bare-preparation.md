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

NEED TO ADD A HARDWARE DIAGRAM AND SHORT DESCRIPTION OF EACH COMPONENT.

Once the hardware is in place, you need to:

* Obtain IP addresses to assign to the IPMI and regular NIC interface on each machine in the cluster.
* Obtain the gateway address for your network to reach the Internet.
* Go to the IPMI settings for each machine and set the IP address (bnc_ip), username (bnc_username), and password (bnc_password) to use later in the CSV file.

## Prepare hardware inventory
Create a CSV file to provide information about all physical machines that you are ready to add to your initial Bare Metal cluster.
This file will be used:

* When you generate the hardware file to be included in the cluster creation process described in the [Create production cluster]({{< relref "../../getting-started/production-environment" >}}) Getting Started guide.
* To provide information that is passed to each machine from the Tinkerbell DHCP server when the machine is initially PXE booted.

The following is an example of an EKS Anywhere Bare Metal hardware CSV file:

```
hostname,bmc_ip,bmc_username,bmc_password,mac,ip_address,netmask,gateway,nameservers,labels,disk
eksa-cp01,10.10.44.1,root,PrZ8W93i,Dell,CC:48:3A:00:00:01,10.10.50.2,255.255.254.0,10.10.50.1,X.X.X.X,,/dev/sda
eksa-cp02,10.10.44.2,root,Me9xQf93,Dell,CC:48:3A:00:00:02,10.10.50.3,255.255.254.0,10.10.50.1,X.X.X.X,,/dev/sda
eksa-wk01,10.10.44.3,root,Z8x2M6hl,Dell,CC:48:3A:00:00:03,10.10.50.4,255.255.254.0,10.10.50.1,X.X.X.X,,/dev/sda
eksa-wk02,10.10.44.4,root,B398xRTp,Dell,CC:48:3A:00:00:04,10.10.50.5,255.255.254.0,10.10.50.1,X.X.X.X,,/dev/sda
eksa-wk03,10.10.44.5,root,w7EenR94,Dell,CC:48:3A:00:00:05,10.10.50.6,255.255.254.0,10.10.50.1,X.X.X.X,,/dev/sda

```

The CSV file is a comma-separated list of values in a plain text file, representing physical control plane and worker machines (not virtual machines) in your cluster.
Multiple values in an item can be separated by pipe symbols (|).

The following sections describe each value.

### hostname
The hostname assigned to the machine.
### bnc_ip
The IP address assigned to IPMI interface on the machine.
### bmc_username
The username assigned to IPMI interface on the machine.
### bnc_password
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
This optional field lets you set lables on each node that are added to any labels set on that node by the cluster.
These labels consist of key/value pairs that can be used by Kubernetes to match workloads that request nodes with those lables.
See Kubernetes [Labels and Selectors]() for details.

In this CSV field, each key/value pair it connected with and equal (`=`) sign.
For example, you could set `environment=test` on one node and `environment=production` on another to allow workloads request the appropriate node to run on based on their stage of devleoment.
That field could appear as follows: `environment=test|environment=production`

### disk
The device name of the disk on which the operating system will be installed.
For example, it could be `/dev/sda` for the first SCSI disk or `/dev/nvme0n1` for the first NVME storage device.
