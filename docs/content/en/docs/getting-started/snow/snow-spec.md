---
title: "Configure for Snow"
linkTitle: "Configuration"
weight: 40
aliases:
    /docs/reference/clusterspec/snow/
description: >
  Full EKS Anywhere configuration reference for a AWS Snow cluster.
---

This is a generic template with detailed descriptions below for reference.
The following additional optional configuration can also be included:

* [CNI]({{< relref "../optional/cni.md" >}})
* [IAM Authenticator]({{< relref "../optional/iamauth.md" >}})
* [OIDC]({{< relref "../optional/oidc.md" >}})
* [GitOps]({{< relref "../optional/gitops.md" >}})
* [Proxy]({{< relref "../optional/proxy.md" >}})
* [Registry Mirror]({{< relref "../optional/registrymirror.md" >}})
* [Machine Health Check Timeouts]({{< relref "../optional/healthchecks.md" >}})

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
      - 10.1.0.0/16
    services:
      cidrBlocks:
      - 10.96.0.0/12
  controlPlaneConfiguration:
    count: 3
    endpoint:
      host: ""
    machineGroupRef:
      kind: SnowMachineConfig
      name: my-cluster-machines
  datacenterRef:
    kind: SnowDatacenterConfig
    name: my-cluster-datacenter
  externalEtcdConfiguration:
    count: 3
    machineGroupRef:
      kind: SnowMachineConfig
      name: my-cluster-machines
  kubernetesVersion: "1.28"
  workerNodeGroupConfigurations:
  - count: 1
    machineGroupRef:
      kind: SnowMachineConfig
      name: my-cluster-machines
    name: md-0
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: SnowDatacenterConfig
metadata:
  name: my-cluster-datacenter
spec: {}

---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: SnowMachineConfig
metadata:
  name: my-cluster-machines
spec:
  amiID: ""
  instanceType: sbe-c.large
  sshKeyName: ""
  osFamily: ubuntu
  devices:
  - ""
  containersVolume:
    size: 25
  network:
    directNetworkInterfaces:
    - index: 1
      primary: true
      ipPoolRef:
        kind: SnowIPPool
        name: ip-pool-1
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: SnowIPPool
metadata:
  name: ip-pool-1
spec:
  pools:
  - ipStart: 192.168.1.2
    ipEnd: 192.168.1.14
    subnet: 192.168.1.0/24
    gateway: 192.168.1.1
  - ipStart: 192.168.1.55
    ipEnd: 192.168.1.250
    subnet: 192.168.1.0/24
    gateway: 192.168.1.1
```

## Cluster Fields

### name (required)
Name of your cluster `my-cluster-name` in this example

{{% include "../_configuration/cluster_clusterNetwork.html" %}}

### controlPlaneConfiguration (required)
Specific control plane configuration for your Kubernetes cluster.

### controlPlaneConfiguration.count (required)
Number of control plane nodes

### controlPlaneConfiguration.machineGroupRef (required)
Refers to the Kubernetes object with Snow specific configuration for your nodes. See `SnowMachineConfig Fields` below.

### controlPlaneConfiguration.endpoint.host (required)
A unique IP you want to use for the control plane VM in your EKS Anywhere cluster. Choose an IP in your network
range that does not conflict with other devices.

>**_NOTE:_** This IP should be outside the network DHCP range as it is a floating IP that gets assigned to one of
the control plane nodes for kube-apiserver loadbalancing.

### controlPlaneConfiguration.taints
A list of taints to apply to the control plane nodes of the cluster.

Replaces the default control plane taint. For k8s versions prior to 1.24, it replaces `node-role.kubernetes.io/master`. For k8s versions 1.24+, it replaces `node-role.kubernetes.io/control-plane`. The default control plane components will tolerate the provided taints.

Modifying the taints associated with the control plane configuration will cause new nodes to be rolled-out, replacing the existing nodes.

>**_NOTE:_** The taints provided will be used instead of the default control plane taint.
Any pods that you run on the control plane nodes must tolerate the taints you provide in the control plane configuration.
>

### controlPlaneConfiguration.labels
A list of labels to apply to the control plane nodes of the cluster. This is in addition to the labels that
EKS Anywhere will add by default.

Modifying the labels associated with the control plane configuration will cause new nodes to be rolled out, replacing
the existing nodes.

### workerNodeGroupConfigurations (required)
This takes in a list of node groups that you can define for your workers.
You may define one or more worker node groups.

### workerNodeGroupConfigurations.count
Number of worker nodes. Optional if autoscalingConfiguration is used, in which case count will default to `autoscalingConfiguration.minCount`.

Refers to [troubleshooting machine health check remediation not allowed]({{< relref "../../troubleshooting/troubleshooting/#machine-health-check-shows-remediation-is-not-allowed" >}}) and choose a sufficient number to allow machine health check remediation.

### workerNodeGroupConfigurations.machineGroupRef (required)
Refers to the Kubernetes object with Snow specific configuration for your nodes. See `SnowMachineConfig Fields` below.

### workerNodeGroupConfigurations.name (required)
Name of the worker node group (default: md-0)

### workerNodeGroupConfigurations.autoscalingConfiguration.minCount
Minimum number of nodes for this node group's autoscaling configuration.

### workerNodeGroupConfigurations.autoscalingConfiguration.maxCount
Maximum number of nodes for this node group's autoscaling configuration.

### workerNodeGroupConfigurations.taints
A list of taints to apply to the nodes in the worker node group.

Modifying the taints associated with a worker node group configuration will cause new nodes to be rolled-out, replacing the existing nodes associated with the configuration.

At least one node group must not have `NoSchedule` or `NoExecute` taints applied to it.

### workerNodeGroupConfigurations.labels
A list of labels to apply to the nodes in the worker node group. This is in addition to the labels that
EKS Anywhere will add by default.

Modifying the labels associated with a worker node group configuration will cause new nodes to be rolled out, replacing
the existing nodes associated with the configuration.

### workerNodeGroupConfigurations.kubernetesVersion
The Kubernetes version you want to use for this worker node group. Supported values: 1.28, 1.27, 1.26, 1.25, 1.24

### externalEtcdConfiguration.count
Number of etcd members.

### externalEtcdConfiguration.machineGroupRef
Refers to the Kubernetes object with Snow specific configuration for your etcd members. See `SnowMachineConfig Fields` below.

### datacenterRef
Refers to the Kubernetes object with Snow environment specific configuration. See `SnowDatacenterConfig Fields` below.

### kubernetesVersion (required)
The Kubernetes version you want to use for your cluster. Supported values: `1.28`, `1.27`, `1.26`, `1.25`, `1.24`

## SnowDatacenterConfig Fields

### identityRef
Refers to the Kubernetes secret object with Snow devices credentials used to reconcile the cluster.

## SnowMachineConfig Fields

### amiID (optional)
AMI ID from which to create the machine instance. Snow provider offers an AMI lookup logic which will look for a suitable AMI ID based on the Kubernetes version and osFamily if the field is empty.

### instanceType (optional)
Type of the Snow EC2 machine instance. See [Quotas for Compute Instances on a Snowball Edge Device](https://docs.aws.amazon.com/snowball/latest/developer-guide/ec2-edge-limits.html) for supported instance types on Snow (Default: `sbe-c.large`).

### osFamily
Operating System on instance machines. Permitted value: ubuntu.

### physicalNetworkConnector (optional)
Type of snow physical network connector to use for creating direct network interfaces. Permitted values: `SFP_PLUS`, `QSFP`, `RJ45` (Default: `SFP_PLUS`).

### sshKeyName (optional)
Name of the AWS Snow SSH key pair you want to configure to access your machine instances.

The default is `eksa-default-{cluster-name}-{uuid}`.

### devices
A device IP list from which to bootstrap and provision machine instances.

### network
Custom network setting for the machine instances. DHCP and static IP configurations are supported.

### network.directNetworkInterfaces[0].index (optional)
Index number of a direct network interface (DNI) used to clarify the position in the list. Must be no smaller than 1 and no greater than 8.

### network.directNetworkInterfaces[0].primary (optional)
Whether the DNI is primary or not. One and only one primary DNI is required in the directNetworkInterfaces list.

### network.directNetworkInterfaces[0].vlanID (optional)
VLAN ID to use for the DNI.

### network.directNetworkInterfaces[0].dhcp (optional)
Whether DHCP is to be used to assign IP for the DNI.

### network.directNetworkInterfaces[0].ipPoolRef (optional)
Refers to a `SnowIPPool` object which provides a range of ip addresses. When specified, an IP address selected from the pool will be allocated to the DNI.

### containersVolume (optional)
Configuration option for customizing containers data storage volume.

### containersVolume.size
Size of the storage for containerd runtime in Gi.

The field is optional for Ubuntu and if specified, the size must be no smaller than 8 Gi.

### containersVolume.deviceName (optional)
Containers volume device name.

### containersVolume.type (optional)
Type of the containers volume. Permitted values: `sbp1`, `sbg1`. (Default: `sbp1`)

`sbp1` stands for capacity-optimized HDD. `sbg1` is performance-optimized SSD.

### nonRootVolumes (optional)
Configuration options for the non root storage volumes.

### nonRootVolumes[0].deviceName
Non root volume device name. Must be specified and cannot have prefix "/dev/sda" as it is reserved for root volume and containers volume.

### nonRootVolumes[0].size
Size of the storage device for the non root volume. Must be no smaller than 8 Gi.

### nonRootVolumes[0].type (optional)
Type of the non root volume. Permitted values: `sbp1`, `sbg1`. (Default: `sbp1`)

`sbp1` stands for capacity-optimized HDD. `sbg1` is performance-optimized SSD.

## SnowIPPool Fields

### pools[0].ipStart
Start address of an IP range.

### pools[0].ipEnd
End address of an IP range.

### pools[0].subnet
An IP subnet for determining whether an IP is within the subnet.

### pools[0].gateway
Gateway of the subnet for routing purpose.
