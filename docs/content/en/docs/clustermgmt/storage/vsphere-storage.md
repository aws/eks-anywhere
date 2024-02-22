---
title: "vSphere storage"
linkTitle: "vSphere storage"
weight: 80
date: 2017-01-05
description: >
  Managing storage on vSphere
---

EKS Anywhere clusters running on vSphere can leverage the vSphere Container Storage Plug-in (also called the [vSphere CSI Driver](https://github.com/kubernetes-sigs/vsphere-csi-driver)) for dynamic provisioning of persistent storage volumes on vSphere storage infrastructure. The CSI Driver integrates with the Cloud Native Storage (CNS) component in vCenter for the purpose of volume provisioning via vSAN, attaching and detaching volumes to/from VMs, mounting, formatting, and unmounting volumes on/from pods, snapshots, cloning, dynamic volume expansion, etc.

### Bundled vSphere CSI Driver Removal

EKS Anywhere versions prior to `v0.16.0` included built-in installation and management of the vSphere CSI Driver in EKS Anywhere clusters. The vSphere CSI driver components in EKS Anywhere included a Kubernetes CSI controller Deployment, a node-driver-registrar DaemonSet, a default Storage Class, and a number of related Secrets and RBAC entities.

In EKS Anywhere version `v0.16.0`, the built-in vSphere CSI driver feature was deprecated and was removed in EKS Anywhere version `v0.17.0`. You can still use the vSphere CSI driver with EKS Anywhere to make use of the storage options provided by vSphere in a Kubernetes-native way, but you must manage the installation and operation of the vSphere CSI driver on your EKS Anywhere clusters. 

Refer to the [vSphere CSI Driver documentation](https://docs.vmware.com/en/VMware-vSphere-Container-Storage-Plug-in) for the self-managed installation and management procedure. Refer to these [compatibiltiy matrices](https://docs.vmware.com/en/VMware-vSphere-Container-Storage-Plug-in/3.0/vmware-vsphere-csp-getting-started/GUID-D4AAD99E-9128-40CE-B89C-AD451DA8379D.html) to determine the correct version of the vSphere CSI Driver for the Kubernetes version and vSphere version you are running with EKS Anywhere.

### vSphere CSI Driver Cleanup for Upgrades

If you are using an EKS Anywhere version `v0.16.0` or below, you must remove the EKS Anywhere-managed version of the vSphere CSI driver prior to upgrading to EKS Anywhere version `v0.17.0` or later. You do not need to run these steps if you are not using the EKS Anywhere-managed version of the vSphere CSI driver. If you are self-managing your vSphere CSI driver installation, it will persist through EKS Anywhere upgrades.

See below for instructions on how to remove the EKS Anywhere vSphere CSI driver objects. You must delete the DaemonSet and Deployment first, as they rely on the other resources.

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