---
title: "Requirements for EKS Anywhere on VMware vSphere "
linkTitle: "Requirements"
weight: 15
description: >
  Preparing a VMware vSphere provider for EKS Anywhere
---

To run EKS Anywhere, you will need:

### Admin Machine:
An Admin Machine with eksctl and the eksctl-anywhere plugin and other binaries installed. Instructions are as outlined in the [administrative machine prerequisites and the Install EKS Anywhere CLi tools]({{< relref "../../ getting-started/install/#administrative-machine-prerequisites" >}}).

### Preparing a VMware vSphere environment for EKS Anywhere:
* A vSphere 7+ environment running vCenter
* Capacity to deploy 6-10 VMs
* DHCP service running in vSphere environment in the primary VM network for your workload cluster
* One network in vSphere to use for the cluster. This network must have inbound access into vCenter
* A OVA imported into vSphere and converted into template for the workload VMs
* User credentials to [create VMs and attach networks, etc]({{< relref "user-permissions.md" >}})
* One IP address routable from cluster but excluded from DHCP offering. 
  This IP address is to be used as the [Control Plane Endpoint IP or kube-vip VIP address]({{< relref "../clusterspec/vsphere/#controlplaneconfigurationendpointhost-required" >}})

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

The administrative machine and the target workload environment will need network access to:

{{% content "./domains.md" %}}


### vSphere Information Needed Before Creating the Cluster
We need to get the following information before creating the cluster:

* **Static IP Addresses**: 
You will need one IP address for the management cluster control plane endpoint, and a separate one for the controlplane of each workload cluster you add. 

So now let’s say you are going to have the management cluster and two workload clusters. Then you would need three IP addresses, one for each. And they all will be configured same way in the configuration file you are going to generate for each cluster.

A static IP addresses will be used for each control plane VM in your EKS Anywhere cluster. Choose IP addresses in your network range that do not conflict with other VMs, that must be excluded from your DHCP offering.

NOTE: This IP addresses should be outside the network DHCP range. Suggestions on how to ensure this IP does not cause issues during cluster creation process are here.

An IP address will be the value of the property `controlPlaneConfiguration.endpoint.host` in the config file of the management cluster, and a separate IP address for each workload cluster.
![Import ova wizard](/images/ip.png) 

* **vSphere Datacenter Name**:
The vSphere datacenter to deploy the EKS Anywhere cluster on.
![Import ova wizard](/images/datacenter.png) 

* **VM Network Name**:
The VM network to deploy your EKS Anywhere cluster on.
![Import ova wizard](/images/network.png) 

* **vCenter Server Domain Name**:
The vCenter server fully qualified domain name or IP address. If the server IP is used, the thumbprint must be set or insecure must be set to true.
![Import ova wizard](/images/domainname.png) 

* **thumbprint** (required if insecure=false):
The SHA1 thumbprint of the vCenter server certificate which is only required if you have a self-signed certificate for your vSphere endpoint.

There are several ways to obtain your vCenter thumbprint. The easiest way is if you have govc installed, you can run the following command in the admin machine terminal, and take a note of the output:

```bash
govc about.cert -thumbprint -k
```

* **template**:
The VM template to use for your EKS Anywhere cluster. This template was created when you imported the OVA file into vSphere . 
![Import ova wizard](/images/template.png) 

* **datastore**:
The vSphere [datastore](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.storage.doc/GUID-3CC7078E-9C30-402C-B2E1-2542BEE67E8F.html) to deploy your EKS Anywhere cluster on.
![Import ova wizard](/images/storage.png) 


* **folder**:
The folder parameter in VSphereMachineConfig allows you to organize your VMs of an EKS Anywhere cluster. With this each cluster can be organized as a folder in vSphere. It will be nice to have a separate folder for the management cluster and each cluster you are adding. 
![Import ova wizard](/images/folder.png) 


* **resourcePool**:
The vSphere Resource pools for your VMs in the EKS Anywhere cluster. If there is a resource pool: `/<datacenter>/host/<resource-pool-name>/Resources`
![Import ova wizard](/images/resourcepool.png) 