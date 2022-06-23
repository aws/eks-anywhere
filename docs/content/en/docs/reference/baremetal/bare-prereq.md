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

## Computer requirements 

The minimum number of physical computers needed to run EKS Anywhere in a non-production mode is:

* Control plane physical machines: 1
* Worker physical machines: 1

The recommended number of physical machines for production is at least:

* Control plane physical machines: 3
* Worker physical machines: 2

You will need an additional, temporary machine for each control plane node grouping and worker node grouping later when you go to upgrade a node.
That machine must have the same specs as all the machines in that group.
This comes from the need to use the same template to populate data on the disks for all nodes in a group.

The computer hardware you need for your Bare Metal cluster must meet the following capacity requirements:

* CPU: 2
* Memory: 8GB RAM
* Storage: 25GB

## Network requirements

Each machine should include the following features:

* Network Interface Cards: At least one NIC is required. It must be capable of netbooting from PXE. 
* IPMI integration (recommended): An IPMI implementation (such a Dell iDRAC or HP iLO) on the computer's motherboard or on a separate expansion card. This feature is used to allow remote management of the machine, such as turning the machine on and off.

>**_NOTE:_** IPMI is not required for an EKS Anywhere cluster. However, without IPMI, upgrades are not supported and you will have to physically turn machines off and on when appropriate.

Here are other network requirements:

* All EKS Anywhere machine, including the Admin, control plane and worker machines, must be on the same layer 2 connection to the other machines in the cluster and have network connectivity to the BMC (IPMI, Redfish, and so on). The hardware does not need to be on the same layer 2 as the BMC, but the Admin machine and management cluster does need routes configured so it can communicate with the BMC API.

* You must be able to run DHCP on the control plane/worker machine network.

* The administrative machine and the target workload environment will need network access to:

  * public.ecr.aws
  * anywhere-assets.eks.amazonaws.com: To download the EKS Anywhere binaries, manifests and OVAs
  * distro.eks.amazonaws.com: To download EKS Distro binaries and manifests
  * d2glxqk2uabbnd.cloudfront.net: For EKS Anywhere and EKS Distro ECR container images

* Two IP addresses routable from the cluster, but excluded from DHCP offering. One IP address is to be used as the Control Plane Endpoint IP or kube-vip VIP address. The other is for the Tinkerbell IP address on the target cluster. Below are some suggestions to ensure that these IP addresses are never handed out by your DHCP server. You may need to contact your network engineer to manage these addresses.

  * Pick IP addresses reachable from the cluster subnet that are excluded from the DHCP range or
  * Create an IP reservation for these addresses on your DHCP server. This is usually accomplished by adding a dummy mapping of this IP address to a non-existent mac address.

>**_NOTE:_** When your set up your cluster configuration YAML file, the endpoint and Tinkerbell addresses are set in the `ControlPlaneConfiguration.endpoint.host` and `tinkerbellIP` fields, respectively.

* Ports must be open to the Admin machine and cluster machines as described in [Ports and protocols]({{< relref "../ports/" >}}).

## Hardware suggestions

While many different hardware options that meet the criteria listed above should work with EKS Anywhere, the following hardware has been tested and shown to work:

* Dell
  * Dell PowerEdge R340 (iDRAC 9)
  * Dell PowerEdge R750 (iDRAC 9)
  * Dell PowerEdge R740xd (iDRAC 9)
  * Dell PowerEdge R240 (iDRAC 9)
* HP
* Supermicro
  * SYS-510-M (IPMI 2.0)
