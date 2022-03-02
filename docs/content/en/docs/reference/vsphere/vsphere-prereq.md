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
