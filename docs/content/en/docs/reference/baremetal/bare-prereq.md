---
title: "Requirements for EKS Anywhere on Bare Metal"
linkTitle: "Requirements"
weight: 15
description: >
  Bare Metal provider requirements for EKS Anywhere
---

To run EKS Anywhere on Bare Metal, you need to meet the hardware and networking requirements described below.


## Administrative machine

Set up an Administrative machine as described in [Install EKS Anywhere ]({{< relref "../../getting-started/install/" >}}).

## Compute server requirements

The minimum number of physical machines needed to run EKS Anywhere in a non-production mode is:

* Control plane physical machines: 1
* Worker physical machines: 1

The recommended number of physical machines for production is at least:

* Control plane physical machines: 3
* Worker physical machines: 2

You will need an additional, temporary machine for each control plane node grouping and worker node grouping later when you go to upgrade a node.
That machine must have the same specs as all the machines in that group.
This comes from the need to use the same template to populate data on the disks for all nodes in a group.

The compute hardware you need for your Bare Metal cluster must meet the following capacity requirements:

* CPU: 2
* Memory: 8GB RAM
* Storage: 25GB

## Network requirements

Each machine should include the following features:

* Network Interface Cards: At least one NIC is required. It must be capable of netbooting from PXE. 
* IPMI integration (recommended): An IPMI implementation (such a Dell iDRAC, RedFish-compatible, legacy or HP iLO) on the computer's motherboard or on a separate expansion card. This feature is used to allow remote management of the machine, such as turning the machine on and off.

>**_NOTE_**: IPMI is not required for an EKS Anywhere cluster. However, without IPMI, upgrades are not supported and you will have to physically turn machines off and on when appropriate.

Here are other network requirements:

* All EKS Anywhere machines, including the Admin, control plane and worker machines, must be on the same layer 2 network and have network connectivity to the BMC (IPMI, Redfish, and so on).

* You must be able to run DHCP on the control plane/worker machine network.

>**_NOTE:_**: If you have another DHCP service running on the network, you need to prevent it from interfering with the EKS Anywhere DHCP service. You can do that by configuring the other DHCP service to explicitly block all MAC addresses and exclude all IP addresses that you plan to use with your EKS Anywhere clusters.

* The administrative machine and the target workload environment will need network access to:

  * public.ecr.aws
  * anywhere-assets.eks.amazonaws.com: To download the EKS Anywhere binaries, manifests and OVAs
  * distro.eks.amazonaws.com: To download EKS Distro binaries and manifests
  * d2glxqk2uabbnd.cloudfront.net: For EKS Anywhere and EKS Distro ECR container images

* Two IP addresses routable from the cluster, but excluded from DHCP offering. One IP address is to be used as the Control Plane Endpoint IP. The other is for the Tinkerbell IP address on the target cluster. Below are some suggestions to ensure that these IP addresses are never handed out by your DHCP server. You may need to contact your network engineer to manage these addresses.

  * Pick IP addresses reachable from the cluster subnet that are excluded from the DHCP range or
  * Create an IP reservation for these addresses on your DHCP server. This is usually accomplished by adding a dummy mapping of this IP address to a non-existent mac address.

>**_NOTE:_** When you set up your cluster configuration YAML file, the endpoint and Tinkerbell addresses are set in the `ControlPlaneConfiguration.endpoint.host` and `tinkerbellIP` fields, respectively.

* Ports must be open to the Admin machine and cluster machines as described in [Ports and protocols]({{< relref "../ports/" >}}).

## Validated hardware

Through extensive testing in a variety of on premises customer environments during our beta phase, we expect Amazon EKS Anywhere on bare metal to run on most generic hardware that meets the above requirements.
In addition, we have collaborated with our hardware original equipment manufacturer (OEM) partners to provide you a list of validated hardware:

| Bare metal servers  | IPMI  | NIC     | OS      |
|---------------------|-------|---------|---------|
| Dell PowerEdge R740 | iDRAC9 |  Mellanox ConnectX-4 LX 25GbE  | Validated with Ubuntu v20.04.1 |
| Dell PowerEdge R7525 (NVIDIA Tesla™ T4 GPU's) | iDRAC9 |  Mellanox ConnectX-4 LX 25GbE & Intel Ethernet 10G 4P X710 OCP | Validated with Ubuntu v20.04.1 |
| Dell PowerFlex (R640) | iDRAC9 | Mellanox ConnectX-4 LX 25GbE | Validated with Ubuntu v20.04.1 |
| SuperServer SYS-510P-M | IPMI2.0/Redfish API | Intel® Ethernet Controller i350 2x 1GbE | Validated with Ubuntu v20.04.1 and Bottlerocket v1.8.0 |
| Dell PowerEdge R240 | iDRAC9 | Broadcom 57414 Dual Port 10/25GbE | Validated with Ubuntu v20.04 and Bottlerocket v1.8.0 |
| HPE ProLiant DL20 | iLO5 | HPE 361i 1G | Validated with Ubuntu v20.04 and Bottlerocket v1.8.0 |
| HPE ProLiant DL160 Gen10 | iLO5 | HPE Eth 10/25Gb 2P 640SFP28 A | Validated with Ubuntu v20.04.1 |
| Dell PowerEdge R340 | iDRAC9 | Broadcom 57416 Dual Port 10GbE | Validated with Ubuntu v20.04.1 and Bottlerocket v1.8.0 |
| HPE ProLiant DL360 | iLO5 | HPE Ethernet 1Gb 4-port 331i | Validated with Ubuntu v20.04.1 |
| Lenovo ThinkSystem SR650 V2 | XClarity Controller Enterprise v7.92 |<ul><li>Intel I350 1GbE RJ45 4-port OCP</li><li>Marvell QL41232 10/25GbE SFP28<br>2-Port PCIe Ethernet Adapter</li></ul>| Validated with Ubuntu v20.04.1 |
