# Support Tagging VM's in a vSphere cluster

## Introduction

**Problem:** User's have no way to Tag VM's provisioned by EKS-A in vSphere. Tags are needed by customers for various reasons.
A specific use-case is for NSX. NSX polices & rules can be configured based on tags on a VM. Customers may control network traffic based on these tags. Therefore, we need to support adding tags at deployment time.

### Goals and Objectives

As an EKS Anywhere user:

* I want to have the ability to add tags to VM's in my vSphere cluster.

## Overview of Solution

CAPV already supports adding tags on nodes as you can see below. The plan is for EKS-A to expose the same values in the EKS-A cluster.yaml.

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: VSphereMachineTemplate
metadata:
  name: vsphere-quickstart-worker
  namespace: default
spec:
  template:
    spec:
      tagIDs:
      - urn:vmomi:InventoryServiceTag:8e0ce079-0675-47d6-8665-16ada4e6dabd:GLOBAL
...
```

### Solution Details

```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: VSphereMachineConfig
metadata:
   name: my-cluster-machines
spec:
  tags:
  - urn:vmomi:InventoryServiceTag:8e0ce079-0675-47d6-8665-16ada4e6dabd:GLOBAL
  diskGiB: 25
  datastore: "Datastore"
  folder: "Folder"
...
```

### Side Effects

With these changes, whichever tags are added to the `VSphereMachineConfig` will be added to the VM during deployment.
