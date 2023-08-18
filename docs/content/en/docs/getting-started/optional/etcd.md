---
title: "etcd"
linkTitle: "etcd"
weight: 5
aliases:
    /docs/reference/clusterspec/optional/etcd/
description: >
  EKS Anywhere cluster yaml etcd specification reference
---

### Unstacked etcd topology (recommended)

#### Provider support details
|                | vSphere | Bare Metal | Nutanix | CloudStack | Snow |
|:--------------:|:-------:|:----------:|:-------:|:----------:|:----:|
| **Supported?** |   ✓	    |            |   	     |     ✓      |  ✓   |

There are two types of etcd topologies for configuring a Kubernetes cluster:

* Stacked: The etcd members and control plane components are colocated (run on the same node/machines)
* Unstacked/External: With the unstacked or external etcd topology, etcd members have dedicated machines and are not colocated with control plane components

The unstacked etcd topology is recommended for a HA cluster for the following reasons:  
  
* External etcd topology decouples the control plane components and etcd member.
  For example, if a control plane-only node fails, or if there is a memory leak in a component like kube-apiserver, it won't directly impact an etcd member.
* Etcd is resource intensive, so it is safer to have dedicated nodes for etcd, since it could use more disk space or higher bandwidth.
  Having a separate etcd cluster for these reasons could ensure a more resilient HA setup.

EKS Anywhere supports both topologies.
In order to configure a cluster with the unstacked/external etcd topology, you need to configure your cluster by updating the configuration file before creating the cluster.
This is a generic template with detailed descriptions below for reference:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
   name: my-cluster-name
spec:
   clusterNetwork:
      pods:
         cidrBlocks:
            - 192.168.0.0/16
      services:
         cidrBlocks:
            - 10.96.0.0/12
      cniConfig:
         cilium: {}
   controlPlaneConfiguration:
      count: 1
      endpoint:
         host: ""
      machineGroupRef:
         kind: VSphereMachineConfig
         name: my-cluster-name-cp
   datacenterRef:
      kind: VSphereDatacenterConfig
      name: my-cluster-name
   # etcd configuration
   externalEtcdConfiguration:
      count: 3
      machineGroupRef:
        kind: VSphereMachineConfig
        name: my-cluster-name-etcd
   kubernetesVersion: "1.27"
   workerNodeGroupConfigurations:
      - count: 1
        machineGroupRef:
           kind: VSphereMachineConfig
           name: my-cluster-name
        name: md-0
```
#### externalEtcdConfiguration (under Cluster)
External etcd configuration for your Kubernetes cluster.

#### count (required)
This determines the number of etcd members in the cluster.
The recommended number is 3.

#### machineGroupRef (required)
Refers to the Kubernetes object with provider specific configuration for your nodes.

