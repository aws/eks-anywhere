---
title: "CloudStack configuration"
linkTitle: "Configuration"
weight: 40
aliases:
    /docs/reference/clusterspec/cloudstack/
description: >
  Full EKS Anywhere configuration reference for a CloudStack cluster
---
This is a generic template with detailed descriptions below for reference.
The following additional optional configuration can also be included:

* [CNI]({{< relref "../optional/cni.md" >}})
* [IAM Roles for Service Accounts]({{< relref "../optional/irsa.md" >}})
* [IAM Authenticator]({{< relref "../optional/iamauth.md" >}})
* [OIDC]({{< relref "../optional/oidc.md" >}})
* [GitOps]({{< relref "../optional/gitops.md" >}})
* [Proxy]({{< relref "../optional/proxy.md" >}})
* [Registry Mirror]({{< relref "../optional/registrymirror.md" >}})
* [Machine Health Checks]({{< relref "../optional/healthchecks.md" >}})
* [API Server Extra Args]({{< relref "../optional/api-server-extra-args.md" >}})


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
    count: 3
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
  kubernetesVersion: "1.32"
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
      name: zone1
      network:
        name: "net1"
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: CloudStackMachineConfig
metadata:
  name: my-cluster-name-cp
spec:
  computeOffering:
    name: "m4-large"
  users:
  - name: capc
    sshAuthorizedKeys:
    - ssh-rsa AAAA...
  template:
    name: "rhel8-k8s-118"
  diskOffering:
    name: "Small"
    mountPath: "/data-small"
    device: "/dev/vdb"
    filesystem: "ext4"
    label: "data_disk"
  symlinks:
    /var/log/kubernetes: /data-small/var/log/kubernetes
  affinityGroupIds:
  - control-plane-anti-affinity
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: CloudStackMachineConfig
metadata:
  name: my-cluster-name
spec:
  computeOffering:
    name: "m4-large"
  users:
  - name: capc
    sshAuthorizedKeys:
    - ssh-rsa AAAA...
  template:
    name: "rhel8-k8s-118"
  diskOffering:
    name: "Small"
    mountPath: "/data-small"
    device: "/dev/vdb"
    filesystem: "ext4"
    label: "data_disk"
  symlinks:
    /var/log/pods: /data-small/var/log/pods
    /var/log/containers: /data-small/var/log/containers
  affinityGroupIds:
  - worker-affinity
  userCustomDetails:
    foo: bar
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: CloudStackMachineConfig
metadata:
  name: my-cluster-name-etcd
spec:
  computeOffering: {}
    name: "m4-large"
  users:
  - name: "capc"
    sshAuthorizedKeys:
    - "ssh-rsa AAAAB3N...
  template:
    name: "rhel8-k8s-118"
  diskOffering:
    name: "Small"
    mountPath: "/data-small"
    device: "/dev/vdb"
    filesystem: "ext4"
    label: "data_disk"
  symlinks:
    /var/lib: /data-small/var/lib
  affinityGroupIds:
  - etcd-affinity
---
```
## Cluster Fields

### name (required)
Name of your cluster `my-cluster-name` in this example

{{% include "../_configuration/cluster_clusterNetwork.html" %}}

### controlPlaneConfiguration (required)
Specific control plane configuration for your Kubernetes cluster.

### controlPlaneConfiguration.count (required)
Number of control plane nodes

### controlPlaneConfiguration.endpoint.host (required)
A unique IP you want to use for the control plane VM in your EKS Anywhere cluster. Choose an IP in your network
range that does not conflict with other VMs.

>**_NOTE:_** This IP should be outside the network DHCP range as it is a floating IP that gets assigned to one of
the control plane nodes for kube-apiserver loadbalancing. Suggestions on how to ensure this IP does not cause issues during cluster
creation process are [here]({{< relref "./cloudstack-prereq/." >}})

### controlPlaneConfiguration.machineGroupRef (required)
Refers to the Kubernetes object with CloudStack specific configuration for your nodes. See `CloudStackMachineConfig Fields` below.

### controlPlaneConfiguration.taints (optional)
A list of taints to apply to the control plane nodes of the cluster.

Replaces the default control plane taint, `node-role.kubernetes.io/master`. The default control plane components will tolerate the provided taints.

Modifying the taints associated with the control plane configuration will cause new nodes to be rolled-out, replacing the existing nodes.

>**_NOTE:_** The taints provided will be used instead of the default control plane taint `node-role.kubernetes.io/master`.
Any pods that you run on the control plane nodes must tolerate the taints you provide in the control plane configuration.
>

### controlPlaneConfiguration.labels (optional)
A list of labels to apply to the control plane nodes of the cluster. This is in addition to the labels that
EKS Anywhere will add by default.

A special label value is supported by the CAPC provider:
```yaml
    labels:
      cluster.x-k8s.io/failure-domain: ds.meta_data.failuredomain
```
The `ds.meta_data.failuredomain` value will be replaced with a failuredomain name where the node is deployed, such as `az-1`.

Modifying the labels associated with the control plane configuration will cause new nodes to be rolled out, replacing
the existing nodes.

### datacenterRef (required)
Refers to the Kubernetes object with CloudStack environment specific configuration. See `CloudStackDatacenterConfig Fields` below.

### externalEtcdConfiguration.count (optional)
Number of etcd members

### externalEtcdConfiguration.machineGroupRef (optional)
Refers to the Kubernetes object with CloudStack specific configuration for your etcd members. See `CloudStackMachineConfig Fields` below.

### kubernetesVersion (required)
The Kubernetes version you want to use for your cluster. The Kubernetes versions supported by your EKS Anywhere version are tabulated in [this]({{< relref "../../concepts/support-versions/#kubernetes-versions" >}}) section.

[Known issue related to Kubernetes versions whose minor version is a multiple of 10]({{< relref "../../troubleshooting/troubleshooting/#error-unable-to-get-cluster-config-from-file-kubernetes-version-13-is-not-supported-by-bundles-manifest-" >}})

### managementCluster (required)
Identifies the name of the management cluster.
If this is a standalone cluster or if it were serving as the management cluster for other workload clusters, this will be the same as the cluster name.

### workerNodeGroupConfigurations (required)
This takes in a list of node groups that you can define for your workers.
You may define one or more worker node groups.

### workerNodeGroupConfigurations[*].count (optional)
Number of worker nodes. (default: `1`) It will be ignored if the [cluster autoscaler curated package]({{< relref "../../packages/cluster-autoscaler/addclauto" >}}) is installed and `autoscalingConfiguration` is used to specify the desired range of replicas.

Refers to [troubleshooting machine health check remediation not allowed]({{< relref "../../troubleshooting/troubleshooting/#machine-health-check-shows-remediation-is-not-allowed" >}}) and choose a sufficient number to allow machine health check remediation.

### workerNodeGroupConfigurations[*].machineGroupRef (required)
Refers to the Kubernetes object with CloudStack specific configuration for your nodes. See `CloudStackMachineConfig Fields` below.

### workerNodeGroupConfigurations[*].name (required)
Name of the worker node group (default: md-0)

### workerNodeGroupConfigurations[*].autoscalingConfiguration.minCount (optional)
Minimum number of nodes for this node group's autoscaling configuration.

### workerNodeGroupConfigurations[*].autoscalingConfiguration.maxCount (optional)
Maximum number of nodes for this node group's autoscaling configuration.

### workerNodeGroupConfigurations[*].taints (optional)
A list of taints to apply to the nodes in the worker node group.

Modifying the taints associated with a worker node group configuration will cause new nodes to be rolled-out, replacing the existing nodes associated with the configuration.

At least one node group must not have `NoSchedule` or `NoExecute` taints applied to it.

### workerNodeGroupConfigurations[*].labels (optional)
A list of labels to apply to the nodes in the worker node group. This is in addition to the labels that
EKS Anywhere will add by default.
A special label value is supported by the CAPC provider:

```yaml
    labels:
      cluster.x-k8s.io/failure-domain: ds.meta_data.failuredomain
```
The `ds.meta_data.failuredomain` value will be replaced with a failuredomain name where the node is deployed, such as `az-1`.

Modifying the labels associated with a worker node group configuration will cause new nodes to be rolled out, replacing
the existing nodes associated with the configuration.

### workerNodeGroupConfigurations[*].kubernetesVersion (optional)
The Kubernetes version you want to use for this worker node group. The Kubernetes versions supported by your EKS Anywhere version are tabulated in [this]({{< relref "../../concepts/support-versions/#kubernetes-versions" >}}) section.

[Known issue related to Kubernetes versions whose minor version is a multiple of 10]({{< relref "../../troubleshooting/troubleshooting/#error-unable-to-get-cluster-config-from-file-kubernetes-version-13-is-not-supported-by-bundles-manifest-" >}})

Must be less than or equal to the cluster `kubernetesVersion` defined at the root level of the cluster spec. The worker node Kubernetes version must be no more than two minor Kubernetes versions lower than the cluster control plane's Kubernetes version. Removing `workerNodeGroupConfiguration.kubernetesVersion` will trigger an upgrade of the node group to the `kubernetesVersion` defined at the root level of the cluster spec.

## CloudStackDatacenterConfig

### availabilityZones.account (optional)
Account used to access CloudStack.
As long as you pass valid credentials, through `availabilityZones.credentialsRef`, this value is not required.

### availabilityZones.credentialsRef (required)
If you passed credentials through the environment variable `EKSA_CLOUDSTACK_B64ENCODED_SECRET` noted in [Create CloudStack production cluster]({{< relref "./cloudstack-getstarted/" >}}), you can identify those credentials here.
For that example, you would use the profile name `global`.
You can instead use a previously created secret on the Kubernetes cluster in the `eksa-system` namespace.

### availabilityZones.domain (optional)
CloudStack domain to deploy the cluster. The default is `ROOT`.

### availabilityZones.managementApiEndpoint (required)
Location of the CloudStack API management endpoint. For example, `http://10.11.0.2:8080/client/api`.

### availabilityZones.{id,name} (required)
Name or ID of the CloudStack zone on which to deploy the cluster.

### availabilityZones.zone.network.{id,name} (required)
CloudStack network name or ID to use with the cluster.

## CloudStackMachineConfig
In the example above, there are separate `CloudStackMachineConfig` sections for the control plane (`my-cluster-name-cp`), worker (`my-cluster-name`) and etcd (`my-cluster-name-etcd`) nodes.

### computeOffering.{id,name} (required)
Name or ID of the CloudStack compute instance.

### users[0].name (optional)
The name of the user you want to configure to access your virtual machines through ssh.
You can add as many users object as you want.

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

### template.{id,name} (required)
The VM template to use for your EKS Anywhere cluster. Currently, a VM based on RHEL 8.6 is required.
This can be a name or ID.
The `template.name` must contain the `Cluster.Spec.KubernetesVersion` or `Cluster.Spec.WorkerNodeGroupConfiguration[].KubernetesVersion` version (in case of modular upgrade). For example, if the Kubernetes version is 1.32, the `template.name` field name should include 1.32, 1_32, 1-32 or 132.
See the [Artifacts]({{< relref "../../osmgmt/artifacts" >}}) page for instructions for building RHEL-based images.

### diskOffering (optional)
Name representing a disk you want to mount into nodes for this CloudStackMachineConfig

### diskOffering.mountPath (optional)
Mount point on which to mount the disk.

### diskOffering.device (optional)
Device name of the disk partition to mount.

### diskOffering.filesystem (optional)
File system type used to format the filesystem on the disk.

### diskOffering.label (optional)
Label to apply to the disk partition.

### symlinks (optional)
Symbolic link of a directory or file you want to mount from the host filesystem to the mounted filesystem.

### userCustomDetails (optional)
Add key/value pairs to nodes in a `CloudStackMachineConfig`.
These can be used for things like identifying sets of nodes that you want to add to a security group that opens selected ports.

### affinityGroupIDs (optional)
Group ID to attach to the set of host systems to indicate how affinity is done for services on those systems.

### affinity (optional)
Allows you to set `pro` and `anti` affinity for the `CloudStackMachineConfig`.
This can be used in a mutually exclusive fashion with the affinityGroupIDs field.
