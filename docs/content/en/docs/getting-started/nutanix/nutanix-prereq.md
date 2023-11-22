---
title: "Requirements for EKS Anywhere on Nutanix Cloud Infrastructure"
linkTitle: "1. Requirements"
weight: 10
aliases:
    /docs/reference/nutanix/nutanix-prereq/
description: >
  Preparing a Nutanix Cloud Infrastructure provider for EKS Anywhere
---

To run EKS Anywhere, you will need:

## Prepare Administrative machine
Set up an Administrative machine as described in [Install EKS Anywhere ]({{< relref "../../getting-started/install/" >}}).

## Prepare a Nutanix environment
To prepare a Nutanix environment to run EKS Anywhere, you need the following:
* A Nutanix environment running AOS 5.20.4+ with AHV and Prism Central 2022.1+
* Capacity to deploy 6-10 VMs
* DHCP service or Nutanix IPAM running in your environment in the primary VM network for your workload cluster
* [Prepare DHCP IP addresses pool]({{< relref "../../clustermgmt/cluster-upgrades/vsphere-and-cloudstack-upgrades.md/#prepare-dhcp-ip-addresses-pool" >}})
* A VM image imported into the Prism Image Service for the workload VMs
* User credentials to create VMs and attach networks, etc
* One IP address routable from cluster but excluded from DHCP/IPAM offering.
  This IP address is to be used as the [Control Plane Endpoint IP]({{< relref "./nutanix-spec/#controlplaneconfigurationendpointhost-required" >}})

  Below are some suggestions to ensure that this IP address is never handed out by your DHCP server.

  You may need to contact your network engineer.

   *  Pick an IP address reachable from cluster subnet which is excluded from DHCP range OR
   *  Alter DHCP ranges to leave out an IP address(s) at the top and/or the bottom of the range OR
   *  Create an IP reservation for this IP on your DHCP server. This is usually accomplished by adding
a dummy mapping of this IP address to a non-existent mac address.
   *  Block an IP address from the Nutanix IPAM managed network using aCLI


Each VM will require:

* 2 vCPUs
* 4GB RAM
* 40GB Disk

The administrative machine and the target workload environment will need network access (TCP/443) to:

{{% content "./domains.md" %}}


## Nutanix information needed before creating the cluster
You need to get the following information before creating the cluster:

* **Static IP Addresses**:
You will need one IP address for the management cluster control plane endpoint, and a separate one for the controlplane of each workload cluster you add.

  Let’s say you are going to have the management cluster and two workload clusters.
For those, you would need three IP addresses, one for each.
All of those addresses will be configured the same way in the configuration file you will generate for each cluster.

  A static IP address will be used for control plane API server HA in each of your EKS Anywhere clusters.
Choose IP addresses in your network range that do not conflict with other VMs and make sure they are excluded from your DHCP offering.

  An IP address will be the value of the property `controlPlaneConfiguration.endpoint.host` in the config file of the management cluster.
A separate IP address must be assigned for each workload cluster.

  ![Import ova wizard](/images/ip.png)

* **Prism Central FQDN or IP Address**: The Prism Central fully qualified domain name or IP address.

* **Prism Element Cluster Name**: The AOS cluster to deploy the EKS Anywhere cluster on.

* **VM Subnet Name**: The VM network to deploy your EKS Anywhere cluster on.

* **Machine Template Image Name**: The VM image to use for your EKS Anywhere cluster.

* **additionalTrustBundle** (required if using a self-signed PC SSL certificate): The PEM encoded CA trust bundle of the root CA that issued the certificate for Prism Central.




