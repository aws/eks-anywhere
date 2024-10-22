---
title: "Manage vSphere VMs with vMotion "
linkTitle: "Manage vSphere VMs"
weight: 20
description: >
  Using vMotion to manage vSphere VMs used in clusters
aliases:
---

## vMotion with EKS Anywhere


VMware vMotion is a feature within vSphere that allows live migration of virtual machines (VMs) between ESXi hypervisor hosts. This document outlines the guidelines for using vMotion to migrate EKS Anywhere nodes between vSphere ESXi hosts using vMotion while ensuring cluster stability.

### Considerations for node migration using vMotion

When migrating EKS Anywhere nodes with vMotion, several considerations must be kept in mind, particularly around configuration values defined in the [vSphere cluster spec file]({{< relref "/docs/getting-started/vsphere/vsphere-spec" >}}) . These configurations must remain unchanged during the migration and the infrastructure these configurations represent should also not change.


* **No cross-vCenter vMotion**

EKS Anywhere nodes cannot be migrated between different vCenter environments using vMotion. The nodes must remain within the same vCenter instance for proper EKS Anywhere operation. The vCenter Server managing the EKS Anywhere cluster is specified in the `VSphereDatacenterConfig` section of the EKS Anywhere [vSphere cluster spec file]({{< relref "/docs/getting-started/vsphere/vsphere-spec" >}}), under the `spec.server` field, and cannot be changed.


* **vSphere infrastructure settings in** `VSphereDatacenterConfig`

  In addition to the vCenter element, two additional elements defined in the `VSphereDatacenterConfig` section of the EKS Anywhere cluster spec file are immutable must remain unchanged during the vMotion process:


  * datacenter `(spec.datacenter)` - The datacenter specified in the EKS Anywhere cluster spec file must not change during the vMotion migration. This value refers to the vSphere datacenter that hosts the EKS Anywhere nodes.


  * network `(spec.network)` - The network defined in the EKS Anywhere cluster spec file must not change during the vMotion migration. This value refers the vSphere network in which the EKS Anywhere nodes are operating. Any changes to this network configuration would disrupt node connectivity and lead to outages in the EKS Anywhere cluster.


* **VMware Storage vMotion is not supported for EKS Anywhere nodes**

  datastore `(spec.datastore)` - Defined in the `VSphereMachineConfig` section of the EKS Anywhere cluster spec file is immutable.  This value refers to the vSphere datastore that holds EKS Anywhere node vm backing store. Modifying the datastore during vMotion (storage vMotion) would require a change to this value, which is not supported.


* **Node network configuration stability**

  The IP address, subnet mask, and default gateway of each EKS Anywhere node must remain unchanged during the vMotion process. Any modifications to the IP address configuration can cause communication failures between the EKS Anywhere nodes, pods, and the control plane, leading to disruptions in EKS Anywhere  cluster operations.


* **EKS Anywhere configuration stabiltiy**

  The EKS Anywhere environment itself should remain unchanged during vMotion.  Do not perform or trigger any EKS Anywhere changes or life cycle events while performing vmotion.


### Best practices for vMotion with EKS Anywhere clusters

* **Follow VMware vMotion best practices**

  * General best practices: Review VMware's general guidelines for optimal vMotion performance, such as ensuring sufficient CPU, memory, and network resources, and minimizing load on the ESXi hosts during the migration. Refer to the [VMware vMotion documentation](https://docs.vmware.com/) for details.

  * VMware vMotion Networking Best Practices: Whenever possible, follow the [Networking Best Practices for VMware vMotion](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.vcenterhost.doc/GUID-7DAD15D4-7F41-4913-9F16-567289E22977.html) to optimize performance and reduce the risk of issues during the migration process.

  * Use High-Speed Networks: A 10GbE or higher speed network is recommended to ensure smooth vMotion operations for EKS-A nodes, particularly those with large memory footprints.


* **Shared Storage**

  Shared storage is a requirement for vmotion of EKS-A clusters.  Storage such as vSAN, Fiber Channel SAN, or NFS should be shared between the supporting vSphere ESXi hosts for maintaining access to the VM's backing data without relying on storage vMotion, which is not supported in EKS-A environments.


* **Monitoring before and after migration**

  To verify cluster health and node stability,  monitor the EKS-A nodes and pods before and after the vMotion migration:

  * Before migration, run the following commands to check the current health and status of the EKS-A nodes and pods.
  * After vMotion activity is completed, run the commands again to verify that the nodes and pods are still operational and healthy.

```
  kubectl get nodes
  kubectl get pods -a
```



* **Infrastructure maintenance during vMotion**

  It is recommended that no other infrastructure maintenance activities be performed during the vMotion operation. The underlying datacenter infrastructure supporting the network, storage, and server resources utilized by VMware vSphere must remain stable during the vMotion process. Any interruptions in these services could lead to partial or complete failures in the vMotion process, potentially causing the EKS Anywhere nodes to lose connectivity or experience disruptions in normal operations.
