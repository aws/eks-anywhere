# Add support for node labels

## Introduction

**Problem:** Within an EKS-A cluster, the user might want to run specific workloads only on certain nodes.
A great way to do this is by assigning labels to the desired nodes.
However, there is currently no way for users to add labels to their nodes.
This limits the users from being able to add this extra level of customization when deploying applications and other workloads on their cluster.

### Goals and Objectives

As an EKS Anywhere user, I want to:

* Have the ability to run specific workloads on specific nodes
* Have the ability to assign one or more labels to a node

## Overview of Solution

With this feature, a user can specify node labels in the node configuration.
This info will be fetched and added to the `kubeletExtraArgs` in the kubelet configuration.

### Solution Details

The user can specify the labels as key-value pairs under the node configurations.

Example in cluster config:
```
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: cluster-name
spec:
   ...
   workerNodeGroupConfigurations:
   - count: 1
     labels:
       label1: foo
       label2: bar
     machineGroupRef:
       kind: VSphereMachineConfig
       name: my-cluster-machines
   ...        
```
The `labels` field will be of type `map[string]string` to stay consistent with how labels are specified in the `metadata` for other k8s objects.
However, `kubeletExtraArgs` expects a string of key-value pairs for the `node-labels` flag, so the above example would need to be converted to a string that looks like: `"label1=foo,label2=bar"`.
After converting that map to a string, we can proceed to add this information to `kubeletExtraArgs`.

Example:
```
kubeletExtraArgs:
  cloud-provider: external
  node-labels: label1=foo,label2=bar
  tls-cipher-suites: TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
```

After the extra args are applied and the cluster is created, the user will now be able to specify which nodes to run workloads on based on those labels.
The user can also view the labels on their nodes using `kubectl get nodes -A --show-labels`.

Example output (labels applied to worker nodes only in this example):

```
NAME                      STATUS    ROLES                  AGE      LABELS
bob-kbf58                  Ready    control-plane,master   5m20s    beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=vsphere-vm.cpu-2.mem-8gb.os-ubuntu,beta.kubernetes.io/os=linux,kubernetes.io/arch=amd64
bob-md-0-56f765c98-pvwj4   Ready    <none>                 4m19s    beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=vsphere-vm.cpu-2.mem-8gb.os-ubuntu,beta.kubernetes.io/os=linux,label1=foo,label2=bar,kubernetes.io/arch=amd64
bob-md-0-56f765c98-vp87t   Ready    <none>                 4m19s    beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=vsphere-vm.cpu-2.mem-8gb.os-ubuntu,beta.kubernetes.io/os=linux,label1=foo,label2=bar,kubernetes.io/arch=amd64
bob-s9shf                  Ready    control-plane,master   4m9s     beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=vsphere-vm.cpu-2.mem-8gb.os-ubuntu,beta.kubernetes.io/os=linux,kubernetes.io/arch=amd64

```



---
Things to note:
- For vSphere Ubuntu, adding these labels to `kubeletExtraArgs` is sufficient.
  However, we have to configure additional settings to add this extra customization when using Bottlerocket.
  Example user data for setting up labels with Bottlerocket:
```
[settings.kubernetes.node-labels]
"label1" = "foo"
"label2" = "bar"
```