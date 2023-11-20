---
title: "Requirements for EKS Anywhere on VMware vSphere "
linkTitle: "1. Requirements"
weight: 15
aliases:
    /docs/reference/vsphere/vsphere-prereq/
description: >
  VMware vSphere provider requirements for EKS Anywhere
---

To run EKS Anywhere, you will need:

## Prepare Administrative machine
Set up an Administrative machine as described in [Install EKS Anywhere ]({{< relref "../../getting-started/install/" >}}).

## Prepare a VMware vSphere environment
To prepare a VMware vSphere environment to run EKS Anywhere, you need the following:
* A vSphere 7+ environment running vCenter.
* Capacity to deploy 6-10 VMs.
* DHCP service running in vSphere environment in the primary VM network for your workload cluster.
  * [Prepare DHCP IP addresses pool]({{< relref "../../clustermgmt/cluster-upgrades/vsphere-and-cloudstack-upgrades.md/#prepare-dhcp-ip-addresses-pool" >}})
* One network in vSphere to use for the cluster. EKS Anywhere clusters need access to vCenter through the network to enable self-managing and storage capabilities.
* An [OVA]({{< relref "customize/vsphere-ovas/" >}}) imported into vSphere and converted into a template for the workload VMs
* It's critical that you set up your [vSphere user credentials properly.]({{< relref "./vsphere-preparation#configuring-vsphere-user-group-and-roles" >}})
* One IP address routable from cluster but excluded from DHCP offering.
  This IP address is to be used as the [Control Plane Endpoint IP.]({{< relref "./vsphere-spec/#controlplaneconfigurationendpointhost-required" >}})

  Below are some suggestions to ensure that this IP address is never handed out by your DHCP server.

  You may need to contact your network engineer.

   *  Pick an IP address reachable from cluster subnet which is excluded from DHCP range OR
   *  Alter DHCP ranges to leave out an IP address(s) at the top and/or the bottom of the range OR
   *  Create an IP reservation for this IP on your DHCP server. This is usually accomplished by adding
a dummy mapping of this IP address to a non-existent mac address.


Each VM will require:

* 2 vCPUs
* 8GB RAM
* 25GB Disk

The administrative machine and the target workload environment will need network access (TCP/443) to:

{{% content "./domains.md" %}}


## vSphere information needed before creating the cluster
You need to get the following information before creating the cluster:

* **Static IP Addresses**:
You will need one IP address for the management cluster control plane endpoint, and a separate IP address for the control plane of each workload cluster you add.

  Let’s say you are going to have the management cluster and two workload clusters.
For those, you would need three IP addresses, one for each cluster.
All of those addresses will be configured the same way in the configuration file you will generate for each cluster.

  A static IP address will be used for each control plane VM in your EKS Anywhere cluster.
Choose IP addresses in your network range that do not conflict with other VMs and make sure they are excluded from your DHCP offering.

  An IP address will be the value of the property `controlPlaneConfiguration.endpoint.host` in the config file of the management cluster.
A separate IP address must be assigned for each workload cluster.

  ![Import ova wizard](/images/ip.png)

* **vSphere Datacenter Name**: The vSphere datacenter to deploy the EKS Anywhere cluster on.

  ![Import ova wizard](/images/datacenter.png)

* **VM Network Name**: The VM network to deploy your EKS Anywhere cluster on.

  ![Import ova wizard](/images/networkname.png)

* **vCenter Server Domain Name**: The vCenter server fully qualified domain name or IP address. If the server IP is used, the thumbprint must be set or insecure must be set to true.

  ![Import ova wizard](/images/domainname.png)

* **thumbprint** (required if insecure=false): The SHA1 thumbprint of the vCenter server certificate which is only required if you have a self-signed certificate for your vSphere endpoint.

  There are several ways to obtain your vCenter thumbprint.
If you have [govc installed,](https://github.com/vmware/govmomi/blob/master/govc/README.md) you can run the following command in the Administrative machine terminal, and take a note of the output:

  ```bash
  govc about.cert -thumbprint -k
  ```

* **template**: The VM template to use for your EKS Anywhere cluster.
This template was created when you imported the [OVA file]({{< relref "./vsphere-preparation#deploy-an-ova-template" >}}) into vSphere.

  ![Import ova wizard](/images/ovatemplate.png)

* **datastore**: The vSphere [datastore](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.storage.doc/GUID-3CC7078E-9C30-402C-B2E1-2542BEE67E8F.html) to deploy your EKS Anywhere cluster on.

  ![Import ova wizard](/images/storage.png)


* **folder**:
The [folder]({{< relref "./vsphere-preparation#configuring-folder-resources" >}}) parameter in VSphereMachineConfig allows you to organize the VMs of an EKS Anywhere cluster.
With this, each cluster can be organized as a folder in vSphere.
You will have a separate folder for the management cluster and each cluster you are adding.

  ![Import ova wizard](/images/folder.png)


* **resourcePool**:
The vSphere resource pools for your VMs in the EKS Anywhere cluster. If there is a resource pool: `/<datacenter>/host/<resource-pool-name>/Resources`

  ![Import ova wizard](/images/resourcepool.png)
