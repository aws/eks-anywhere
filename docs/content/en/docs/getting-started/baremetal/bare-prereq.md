---
title: "Requirements for EKS Anywhere on Bare Metal"
linkTitle: "1. Requirements"
weight: 15
aliases:
    /docs/reference/baremetal/bare-prereq/
description: >
  Bare Metal provider requirements for EKS Anywhere
---

To run EKS Anywhere on Bare Metal, you need to meet the hardware and networking requirements described below.


## Administrative machine

Set up an Administrative machine as described in Install EKS Anywhere.

## Compute server requirements

The minimum number of physical machines needed to run EKS Anywhere on bare metal is 1. To configure EKS Anywhere to run on a single server, set `controlPlaneConfiguration.count` to 1, and omit `workerNodeGroupConfigurations` from your cluster configuration.

The recommended number of physical machines for production is at least:

* Control plane physical machines: 3
* Worker physical machines: 2

The compute hardware you need for your Bare Metal cluster must meet the following capacity requirements:

* vCPU: 2
* Memory: 8GB RAM
* Storage: 25GB

## Operating system requirements

If you intend on using a non-Bottlerocket OS you must build it using `image-builder`. See the [OS Management Artifacts]({{< relref "../../osmgmt/artifacts#bare-metal-artifacts" >}}) page for help building the OS.

## Upgrade requirements
If you are running a standalone cluster with only one control plane node, you will need at least one additional, temporary machine for each control plane node grouping. For cluster with multiple control plane nodes, you can perform a rolling upgrade with or without an extra temporary machine. For worker node upgrades, you can perform a rolling upgrade with or without an extra temporary machine.

When upgrading without an extra machine, keep in mind that your control plane and your workload must be able to tolerate node unavailability. When upgrading with extra machine(s), you will need additional temporary machine(s) for each control plane and worker node grouping. Refer to [Upgrade Bare Metal Cluster]({{< relref "../../clustermgmt/cluster-upgrades/baremetal-upgrades/" >}}) and [Advanced configuration for rolling upgrade]({{< relref "../../clustermgmt/cluster-upgrades/baremetal-upgrades/#advanced-configuration-for-rolling-upgrade" >}}).

> **_NOTE_**: For single-node clusters that require an additional temporary machine for upgrading, if you don't want to set up the extra hardware, you may recreate the cluster for upgrading and handle data recovery manually.

## Network requirements

Each machine should include the following features:

* Network Interface Cards: at least one NIC is required. It must be capable of network booting.

* BMC integration (recommended): an IPMI or Redfish implementation (such a Dell iDRAC, RedFish-compatible, legacy or HP iLO) on the computer's motherboard or on a separate expansion card. This feature is used to allow remote management of the machine, such as turning the machine on and off.

> **_NOTE:_** BMC integration is not required for an EKS Anywhere cluster. However, without BMC integration, upgrades are not supported and you will have to physically turn machines off and on when appropriate.

Here are other network requirements:

* All EKS Anywhere machines, including the Admin, control plane and worker machines, must be on the same layer 2 network and have network connectivity to the BMC (IPMI, Redfish, and so on).

* You must be able to run DHCP on the control plane/worker machine network.

> **_NOTE:_** If you have another DHCP service running on the network, you need to prevent it from interfering with the EKS Anywhere DHCP service. You can do that by configuring the other DHCP service to explicitly block all MAC addresses and exclude all IP addresses that you plan to use with your EKS Anywhere clusters.

* If you have not followed the [steps for airgapped environments]({{< relref "../airgapped" >}}), then the administrative machine and the target workload environment need network access (TCP/443) to:

  * `public.ecr.aws`

  * `anywhere-assets.eks.amazonaws.com`: to download the EKS Anywhere binaries, manifests and OVAs

  * `distro.eks.amazonaws.com`: to download EKS Distro binaries and manifests

  * `d2glxqk2uabbnd.cloudfront.net`: for EKS Anywhere and EKS Distro ECR container images

* Two IP addresses routable from the cluster, but excluded from DHCP offering. One IP address is to be used as the Control Plane Endpoint IP. The other is for the Tinkerbell IP address on the target cluster. Below are some suggestions to ensure that these IP addresses are never handed out by your DHCP server. You may need to contact your network engineer to manage these addresses.

  * Pick IP addresses reachable from the cluster subnet that are excluded from the DHCP range or

  * Create an IP reservation for these addresses on your DHCP server. This is usually accomplished by adding a dummy mapping of this IP address to a non-existent mac address.

> **_NOTE:_** When you set up your cluster configuration YAML file, the endpoint and Tinkerbell addresses are set in the `controlPlaneConfiguration.endpoint.host` and `tinkerbellIP` fields, respectively.

* Ports must be open to the Admin machine and cluster machines as described in the [Cluster Networking documentation]({{< relref "../ports" >}}).

## Validated hardware

Through extensive testing in a variety of on-premises environments, we expect Amazon EKS Anywhere on bare metal to run on most generic hardware that meets the above requirements.
In addition, we have collaborated with our hardware original equipment manufacturer (OEM) partners to provide you a list of validated hardware:

| Bare metal servers  | BMC   | NIC     |
|---------------------|-------|---------|
| Dell PowerEdge R740 | iDRAC9 |  Mellanox ConnectX-4 LX 25GbE  |
| Dell PowerEdge R7525 (NVIDIA Tesla™ T4 GPU's) | iDRAC9 |  Mellanox ConnectX-4 LX 25GbE & Intel Ethernet 10G 4P X710 OCP |
| Dell PowerFlex (R640) | iDRAC9 | Mellanox ConnectX-4 LX 25GbE |
| SuperServer SYS-510P-M | IPMI2.0/Redfish API | Intel® Ethernet Controller i350 2x 1GbE |
| Dell PowerEdge R240 | iDRAC9 | Broadcom 57414 Dual Port 10/25GbE |
| HPE ProLiant DL20 | iLO5 | HPE 361i 1G |
| HPE ProLiant DL160 Gen10 | iLO5 | HPE Eth 10/25Gb 2P 640SFP28 A |
| Dell PowerEdge R340 | iDRAC9 | Broadcom 57416 Dual Port 10GbE |
| HPE ProLiant DL360 | iLO5 | HPE Ethernet 1Gb 4-port 331i |
| Lenovo ThinkSystem SR650 V2 | XClarity Controller Enterprise v7.92 | Intel I350 1GbE RJ45 4-port OCP & Marvell QL41232 10/25GbE SFP28<br>2-Port PCIe Ethernet Adapter |
