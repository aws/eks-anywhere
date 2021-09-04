---
title: "Requirements for EKS Anywhere on VMware vSphere "
linkTitle: "Requirements"
weight: 15
description: >
  Preparing a VMware vSphere provider for EKS Anywhere
---

To run EKS Anywhere, you will need:

* A vSphere 7+ environment running vCenter
* Capacity to deploy 8-10VMs
* DHCP service running in vSphere environment in the primary VM network for your workload cluster
* One network in vSphere to use for the cluster. This network must have inbound access into vCenter
* A OVA imported into vSphere and converted into template for the workload VMs
* User credentials to [create vms and attach networks, etc]({{< relref "user-permissions.md" >}})

A minimum workload cluster will require 6 to 10 VMs and each VM will require:

* 2 vCPU
* 8GB RAM
* 20GB Disk

The administrative machine and the target workload environment will need network access to:

{{% content "domains.md" %}}
