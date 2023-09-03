---
title: "vSphere storage"
linkTitle: "vSphere storage"
weight: 80
date: 2017-01-05
description: >
  Managing storage on vSphere
---

EKS Anywhere clusters running on vSphere can leverage the vSphere Container Storage Plug-in (also called the vSphere CSI Driver) for dynamic provisioning of persistent storage volumes on vSphere storage infrastructure. The CSI Driver integrates with the Cloud Native Storage (CNS) component in vCenter for the purpose of volume provisioning via vSAN, attaching and detaching volumes to/from VMs, mounting, formatting, and unmounting volumes on/from pods, snapshots, cloning, dynamic volume expansion, etc.

### vSphere CSI Driver Deprecation

EKS Anywhere versions prior to `v0.16.0` supported the installation and management of the vSphere CSI Driver in EKS-A clusters. The CSI management components in EKS-A included a Kubernetes CSI controller Deployment, a node-driver-registrar DaemonSet, a default Storage Class, and a number of related Secrets and RBAC entities.

In EKS-A version `v0.16.0`, the CSI driver feature was deprecated as part of cluster creation and has been completely removed in version `v0.17.0`. However, you may self-manage this deployment to make use of the storage options provided by vSphere in a Kubernetes-native way. 

### CSI Driver Cleanup for Upgrades

If you are using EKS-A version `v0.16.0` and above to upgrade a cluster that has the vSphere CSI Driver installed in it, follow the below steps for proper cleanup of unmanaged vSphere CSI resources.

These are the resources you would need to delete from your cluster:

#### `default` namespace
* `vsphere-csi-controller-role` (kind: `ClusterRole`)
  ```bash
  kubectl delete clusterrole vsphere-csi-controller-role
  ```
* `vsphere-csi-controller-binding` (kind: `ClusterRoleBinding`)
  ```bash
  kubectl delete clusterrolebinding vsphere-csi-controller-binding
  ```
* `csi.vsphere.vmware.com` (kind: `CSIDriver`)
  ```bash
  kubectl delete csidriver csi.vsphere.vmware.com
  ```

#### `kube-system` namespace
* `vsphere-csi-node` (kind: `DaemonSet`)
  ```bash
  kubectl delete daemonset vsphere-csi-node -n kube-system
  ```
* `vsphere-csi-controller` (kind: `Deployment`)
  ```bash
  kubectl delete deployment vsphere-csi-controller -n kube-system
  ```
* `vsphere-csi-controller` (kind: `ServiceAccount`)
  ```bash
  kubectl delete serviceaccount vsphere-csi-controller -n kube-system
  ```
* `csi-vsphere-config` (kind: `Secret`)
  ```bash
  kubectl delete secret csi-vsphere-config -n kube-system
  ```

#### `eksa-system` namespace
* `<cluster-name>-csi` (kind: `ClusterResourceSet`)
  ```bash
  kubectl delete clusterresourceset <cluster-name>-csi -n eksa-system
  ```

>**_NOTE:_** Delete the `DaemonSet` and `Deployment` first, as they rely on the other resources.

Once these resources have been removed, you can refer to the [vSphere CSI Driver documentation](https://docs.vmware.com/en/VMware-vSphere-Container-Storage-Plug-in) for the installation and management procedure. Please refer to these [compatibiltiy matrices](https://docs.vmware.com/en/VMware-vSphere-Container-Storage-Plug-in/3.0/vmware-vsphere-csp-getting-started/GUID-D4AAD99E-9128-40CE-B89C-AD451DA8379D.html) to determine the correct version of the CSI Driver for the Kubernetes version and vSphere version you are running.