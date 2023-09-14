---
title: 4. Choose provider
main_menu: true
weight: 17
description: >
  Choose an infrastructure provider for EKS Anywhere clusters
---

EKS Anywhere supports many different types of infrastructure including VMWare vSphere, bare metal, Snow, Nutanix, and Apache CloudStack. You can also run EKS Anywhere on Docker for dev/test use cases only. EKS Anywhere clusters can only run on a single infrastructure provider. For example, you cannot have some vSphere nodes, some bare metal nodes, and some Snow nodes in a single EKS Anywhere cluster. Management clusters also must run on the same infrastructure provider as workload clusters. 

Detailed information on each infrastructure provider can be found in the sections below. Review the infrastructure provider's prerequisites in-depth before creating your first cluster.

##### [**Install on vSphere**]({{< relref "../vsphere/" >}})
##### [**Install on Bare Metal**]({{< relref "../baremetal/" >}})
##### [**Install on Snow**]({{< relref "../snow/" >}})
##### [**Install on CloudStack**]({{< relref "../cloudstack/" >}})
##### [**Install on Nutanix**]({{< relref "../nutanix/" >}})
##### [**Install on Docker (dev only)**]({{< relref "../docker/" >}})
