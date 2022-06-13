---
title: "Requirements for EKS Anywhere on Bare Metal"
linkTitle: "Requirements"
weight: 15
description: >
  Bare Metal provider requirements for EKS Anywhere
---

To run EKS Anywhere on Bare Metal, you need the following:


## Administrative machine

Set up an Administrative machine as described in [Install EKS Anywhere ]({{< relref "../../getting-started/install/" >}}).

## Computer requirements 

The minimum number of physical computers needed to run EKS Anywhere in a non-production mode is:

* Control plane physical machines: 1
* Worker physical machines: 1

The recommended number of physical machines for production is at least:

* Control plane physical machines: 3
* Worker physical machines: 2

You will need an additional, temporary machine later when you go to upgrade a node.

The computer hardware you need for your Bare Metal cluster must meet the following capacity requirements:

* CPU: 2
* Memory: 8GB RAM
* Storage: 25GB

## Network requirements

Each machine must include the following features:

* Network Interface Cards: At least one NIC is required. It must be capable of netbooting from iPXE. It must have at least 4GB of RAM for the OSIE/Hook boot and operation.
* IPMI integration: An IPMI implementation (such a Dell Drac or HP iLO) on the computer's motherboard or on a separate expansion card. This feature is used to allow remote management of the machine, such as turning the machine on and off.

Here are other network requirements:

* All EKS Anywhere machine, including the Admin, control plane and worker machines, must be on the same level 2 connection to the other machines in the cluster.

* Two network switches are required: one for connectivity between the nodes and one for IPMI connnectivity.

* You must be able to run DHCP on control plane/worker machine network.

* The administrative machine and the target workload environment will need network access to:

  * public.ecr.aws
  * anywhere-assets.eks.amazonaws.com: To download the EKS Anywhere binaries, manifests and OVAs
  * distro.eks.amazonaws.com: To download EKS Distro binaries and manifests
  * d2glxqk2uabbnd.cloudfront.net: For EKS Anywhere and EKS Distro ECR container images

* One IP address routable from cluster but excluded from DHCP offering. This IP address is to be used as the Control Plane Endpoint IP or kube-vip VIP address. Below are some suggestions to ensure that this IP address is never handed out by your DHCP server. You may need to contact your network engineer.

  * Pick an IP address reachable from cluster subnet which is excluded from DHCP range OR
  * Alter DHCP ranges to leave out an IP address(s) at the top and/or the bottom of the range OR
  * Create an IP reservation for this IP on your DHCP server. This is usually accomplished by adding a dummy mapping of this IP address to a non-existent mac address.

## Hardware suggestions

While many different hardware options that meet the criteria listed above should work with EKS Anywhere, the following hardware has been tested and shown to work:

### Computer hardware (NEED MORE SPECIFIC INFO)

* Dell
  * PowerEdge R340
  * Dell PowerEdge R750
  * Dell PowerEdge R240
* HP
* Supermicro


### Network switches (NEED MORE SPECIFIC INFO)

* Cisco Nexus 9300 (48-port, 1/10G/25G, and 6p 40G/100G) switches
