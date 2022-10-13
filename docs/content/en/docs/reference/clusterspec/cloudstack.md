---
title: "CloudStack configuration"
linkTitle: "CloudStack configuration"
weight: 30
description: >
  Full EKS Anywhere configuration reference for a CloudStack cluster.
---
This is a generic template with detailed descriptions below for reference.
The following additional optional configuration can also be included:

* [CNI]({{< relref "optional/cni.md" >}})
* [IAM for pods]({{< relref "optional/irsa.md" >}})
* [IAM Authenticator]({{< relref "optional/iamauth.md" >}})
* [OIDC]({{< relref "optional/oidc.md" >}})
* [gitops]({{< relref "optional/gitops.md" >}})
* [proxy]({{< relref "optional/proxy.md" >}})
* [Registry Mirror]({{< relref "optional/registrymirror.md" >}})


```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: my-cluster-name
spec:
  clusterNetwork:
    cniConfig:
      cilium: {}
    pods:
      cidrBlocks:
      - 192.168.0.0/16
    services:
      cidrBlocks:
      - 10.96.0.0/12
  controlPlaneConfiguration:
    count: 2
    endpoint:
      host: ""
    machineGroupRef:
      kind: CloudStackMachineConfig
      name: my-cluster-name-cp
    taints:
    - key: ""
      value: ""
      effect: ""
    labels:
      "<key1>": ""
      "<key2>": ""
  datacenterRef:
    kind: CloudStackDatacenterConfig
    name: my-cluster-name
  externalEtcdConfiguration:
    count: 3
    machineGroupRef:
      kind: CloudStackMachineConfig
      name: my-cluster-name-etcd
  kubernetesVersion: "1.23"
  managementCluster:
    name: my-cluster-name
  workerNodeGroupConfigurations:
  - count: 2
    machineGroupRef:
      kind: CloudStackMachineConfig
      name: my-cluster-name
    taints:
    - key: ""
      value: ""
      effect: ""
    labels:
      "<key1>": ""
      "<key2>": ""
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: CloudStackDatacenterConfig
metadata:
  name: my-cluster-name-datacenter
spec:
  availabilityZones:
  - account: admin
    credentialsRef: global
    domain: domain1
    managementApiEndpoint: ""
    name: az-1
    zone:
      network: {}
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: CloudStackMachineConfig
metadata:
  name: my-cluster-name-cp
spec:
  computeOffering: {}
  template: {}
  users:
  - name: capc
    sshAuthorizedKeys:
    - ssh-rsa AAAA...
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: CloudStackMachineConfig
metadata:
  name: my-cluster-name
spec:
  computeOffering: {}
  template: {}
  users:
  - name: capc
    sshAuthorizedKeys:
    - ssh-rsa AAAA...
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: CloudStackMachineConfig
metadata:
  name: my-cluster-name-etcd
spec:
  computeOffering: {}
  template: {}
  users:
  - name: capc
    sshAuthorizedKeys:
    - ssh-rsa AAAA...
---
```
## Cluster Fields

### name (required)
Name of your cluster `my-cluster-name` in this example

### clusterNetwork (required)
Specific network configuration for your Kubernetes cluster.

### clusterNetwork.cniConfig (required)
CNI plugin configuration to be used in the cluster. The only supported configuration at the moment is `cilium`.

### clusterNetwork.cniConfig.cilium.policyEnforcementMode
Optionally, you may specify a policyEnforcementMode of default, always, never.

### clusterNetwork.pods.cidrBlocks[0] (required)
Subnet used by pods in CIDR notation. Please note that only 1 custom pods CIDR block specification is permitted.
This CIDR block should not conflict with the network subnet range selected for the VMs.

### clusterNetwork.services.cidrBlocks[0] (required)
Subnet used by services in CIDR notation. Please note that only 1 custom services CIDR block specification is permitted.
This CIDR block should not conflict with the network subnet range selected for the VMs.

### controlPlaneConfiguration (required)
Specific control plane configuration for your Kubernetes cluster.

### controlPlaneConfiguration.count (required)
Number of control plane nodes

### controlPlaneConfiguration.endpoint.host (required)
A unique IP you want to use for the control plane VM in your EKS Anywhere cluster. Choose an IP in your network
range that does not conflict with other VMs.

>**_NOTE:_** This IP should be outside the network DHCP range as it is a floating IP that gets assigned to one of
the control plane nodes for kube-apiserver loadbalancing. Suggestions on how to ensure this IP does not cause issues during cluster
creation process are [here]({{< relref "../cloudstack/cloudstack-prereq/." >}})

### controlPlaneConfiguration.machineGroupRef (required)
Refers to the Kubernetes object with CloudStack specific configuration for your nodes. See `CloudStackMachineConfig Fields` below.

### controlPlaneConfiguration.taints
A list of taints to apply to the control plane nodes of the cluster.

Replaces the default control plane taint, `node-role.kubernetes.io/master`. The default control plane components will tolerate the provided taints.

Modifying the taints associated with the control plane configuration will cause new nodes to be rolled-out, replacing the existing nodes.

>**_NOTE:_** The taints provided will be used instead of the default control plane taint `node-role.kubernetes.io/master`.
Any pods that you run on the control plane nodes must tolerate the taints you provide in the control plane configuration.
>

### controlPlaneConfiguration.labels
A list of labels to apply to the control plane nodes of the cluster. This is in addition to the labels that
EKS Anywhere will add by default.

Modifying the labels associated with the control plane configuration will cause new nodes to be rolled out, replacing
the existing nodes.

### datacenterRef
Refers to the Kubernetes object with CloudStack environment specific configuration. See `CloudStackDatacenterConfig Fields` below.

### externalEtcdConfiguration.count
Number of etcd members

### externalEtcdConfiguration.machineGroupRef
Refers to the Kubernetes object with CloudStack specific configuration for your etcd members. See `CloudStackMachineConfig Fields` below.

### kubernetesVersion (required)
The Kubernetes version you want to use for your cluster. Supported values: `1.23`, `1.22`, `1.21`, `1.20`

### managementCluster

### workerNodeGroupConfigurations (required)
This takes in a list of node groups that you can define for your workers.
You may define one or more worker node groups.

### workerNodeGroupConfigurations.count (required)
Number of worker nodes

### workerNodeGroupConfigurations.machineGroupRef (required)
Refers to the Kubernetes object with CloudStack specific configuration for your nodes. See `CloudStackMachineConfig Fields` below.

### workerNodeGroupConfigurations.name (required)
Name of the worker node group (default: md-0)

### workerNodeGroupConfigurations.taints
A list of taints to apply to the nodes in the worker node group.

Modifying the taints associated with a worker node group configuration will cause new nodes to be rolled-out, replacing the existing nodes associated with the configuration.

At least one node group must not have `NoSchedule` or `NoExecute` taints applied to it.

### workerNodeGroupConfigurations.labels
A list of labels to apply to the nodes in the worker node group. This is in addition to the labels that
EKS Anywhere will add by default.

Modifying the labels associated with a worker node group configuration will cause new nodes to be rolled out, replacing
the existing nodes associated with the configuration.

## CloudStackDatacenterConfig

### availabilityZones.account (required)
Account used to access CloudStack. The default is `admin`.

### availabilityZones.credentialsRef (required)

### availabilityZones.domain (required)
CloudStack domain to deploy the cluster. The default is `ROOT`.

### availabilityZones.managementApiEndpoint (required)
Location of the CloudStack API management endpoint. For example, `http://10.11.0.2:8080/client/api`.

### availabilityZones.name (required)
Name of the CloudStack zone on which to deploy the cluster.

### availabilityZones.zone.network (required)
CloudStack network name to use with the cluster.

## CloudStackMachineConfig
In the example above, there are separate `CloudStackMachineConfig` sections for the control plane (`my-cluster-name-cp`), worker (`my-cluster-name`) and etcd (`my-cluster-name-etcd`)  nodes.

### computeOfferings
Name of the CloudStack compute instance.

### users[0].name (optional)
The name of the user you want to configure to access your virtual machines through ssh.

The default is `capc`.

### users[0].sshAuthorizedKeys (optional)
The SSH public keys you want to configure to access your virtual machines through ssh (as described below). Only 1 is supported at this time.

### users[0].sshAuthorizedKeys[0] (optional)
This is the SSH public key that will be placed in `authorized_keys` on all EKS Anywhere cluster VMs so you can ssh into
them. The user will be what is defined under name above. For example:

```
ssh -i <private-key-file> <user>@<VM-IP>
```

The default is generating a key in your `$(pwd)/<cluster-name>` folder when not specifying a value.

### template (optional)
The VM template to use for your EKS Anywhere cluster. Currently, a VM based on RHEL 8.4 is required.
