---
title: "Nutanix configuration"
linkTitle: "Nutanix configuration"
weight: 20
description: >
  Full EKS Anywhere configuration reference for a Nutanix cluster.
---

This is a generic template with detailed descriptions below for reference.

```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
 name: mgmt
 namespace: default
spec:
 bundlesRef:
   apiVersion: anywhere.eks.amazonaws.com/v1alpha1
   name: bundles-2
   namespace: eksa-system
 clusterNetwork:
   cniConfig:
     cilium: {}
   pods:
     cidrBlocks:
       - 192.168.0.0/16
   services:
     cidrBlocks:
       - 10.96.0.0/16
 controlPlaneConfiguration:
   count: 3
   endpoint:
     host: ""
   machineGroupRef:
     kind: NutanixMachineConfig
     name: mgmt-cp-machine
 datacenterRef:
   kind: NutanixDatacenterConfig
   name: nutanix-cluster
 kubernetesVersion: "1.23"
 workerNodeGroupConfigurations:
   - count: 1
     machineGroupRef:
       kind: NutanixMachineConfig
       name: mgmt-machine
     name: md-0
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: NutanixDatacenterConfig
metadata:
 name: nutanix-cluster
 namespace: default
spec:
 endpoint: pc01.cloud.internal
 port: 9440
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: NutanixMachineConfig
metadata:
 annotations:
   anywhere.eks.amazonaws.com/control-plane: "true"
 name: mgmt-cp-machine
 namespace: default
spec:
 cluster:
   name: nx-cluster-01
   type: name
   uuid: null
 image:
   name: eksa-ubuntu-2004-kube-v1.23
   type: name
   uuid: null
 memorySize: 4Gi
 osFamily: ubuntu
 subnet:
   name: vm-network
   type: name
   uuid: null
 systemDiskSize: 40Gi
 users:
   - name: nutanix-user
     sshAuthorizedKeys:
       - ssh-rsa AAAA…
 vcpuSockets: 2
 vcpusPerSocket: 1
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: NutanixMachineConfig
metadata:
 name: mgmt-machine
 namespace: default
spec:
 cluster:
   name: nx-cluster-01
   type: name
   uuid: null
 image:
   name: eksa-ubuntu-2004-kube-v1.23
   type: name
   uuid: null
 memorySize: 4Gi
 osFamily: ubuntu
 subnet:
   name: vm-network
   type: name
   uuid: null
 systemDiskSize: 40Gi
 users:
   - name: nutanix-user
     sshAuthorizedKeys:
       - ssh-rsa AAAA…
 vcpuSockets: 2
 vcpusPerSocket: 1
---
```

## Cluster Fields

### name (required)
Name of your cluster `mgmt` in this example.

### clusterNetwork (required)
Specific network configuration for your Kubernetes cluster.

### clusterNetwork.cniConfig (required)
CNI plugin configuration to be used in the cluster. The only supported configuration at the moment is `cilium`.

### clusterNetwork.cniConfig.cilium.policyEnforcementMode
Optionally, you may specify a `policyEnforcementMode` of `default`, `always`, `never`.

### clusterNetwork.pods.cidrBlocks[0] (required)
Subnet used by pods in CIDR notation. Please note that only 1 custom pods CIDR block specification is permitted. This CIDR block should not conflict with the network subnet range selected for the VMs.

### clusterNetwork.services.cidrBlocks[0] (required)
Subnet used by services in CIDR notation. Please note that only 1 custom services CIDR block specification is permitted. This CIDR block should not conflict with the network subnet range selected for the VMs.

### controlPlaneConfiguration (required)
Specific control plane configuration for your Kubernetes cluster.

### controlPlaneConfiguration.count (required)
Number of control plane nodes

### controlPlaneConfiguration.machineGroupRef (required)
Refers to the Kubernetes object with Nutanix specific configuration for your nodes. See `NutanixMachineConfig` fields below.

### controlPlaneConfiguration.endpoint.host (required)
A unique IP you want to use for the control plane VM in your EKS Anywhere cluster. Choose an IP in your network range that does not conflict with other VMs.

>**_NOTE:_** This IP should be outside the network DHCP range as it is a floating IP that gets assigned to one of
the control plane nodes for kube-apiserver loadbalancing. Suggestions on how to ensure this IP does not cause issues during cluster 
creation process are [here]({{< relref "../nutanix/nutanix-prereq/#prepare-a-nutanix-environment" >}}).

### controlPlaneConfiguration.taints
A list of taints to apply to the control plane nodes of the cluster.
Replaces the default control plane taint, `node-role.kubernetes.io/master`. The default control plane components will tolerate the provided taints.
 
Modifying the taints associated with the control plane configuration will cause new nodes to be rolled-out, replacing the existing nodes.
 
>**_NOTE:_** The taints provided will be used instead of the default control plane taint `node-role.kubernetes.io/master`.
Any pods that you run on the control plane nodes must tolerate the taints you provide in the control plane configuration.
> 

### controlPlaneConfiguration.labels
A list of labels to apply to the control plane nodes of the cluster. This is in addition to the labels that EKS Anywhere will add by default.
 
Modifying the labels associated with the control plane configuration will cause new nodes to be rolled out, replacing the existing nodes.

### workerNodeGroupConfigurations (required)
This takes in a list of node groups that you can define for your workers. You may define one or more worker node groups.

### workerNodeGroupConfigurations.count
Number of worker nodes. Optional if `autoscalingConfiguration` is used, in which case count will default to `autoscalingConfiguration.minCount`.

### workerNodeGroupConfigurations.machineGroupRef (required)
Refers to the Kubernetes object with Nutanix specific configuration for your nodes. See `NutanixMachineConfig` fields below.

### workerNodeGroupConfigurations.name (required)
Name of the worker node group (default: `md-0`)

### workerNodeGroupConfigurations.autoscalingConfiguration.minCount
Minimum number of nodes for this node group’s autoscaling configuration.

### workerNodeGroupConfigurations.autoscalingConfiguration.maxCount
Maximum number of nodes for this node group’s autoscaling configuration.

### workerNodeGroupConfigurations.taints
A list of taints to apply to the nodes in the worker node group.
 
Modifying the taints associated with a worker node group configuration will cause new nodes to be rolled-out, replacing the existing nodes associated with the configuration.
 
At least one node group must not have `NoSchedule` or `NoExecute` taints applied to it.

### workerNodeGroupConfigurations.labels
A list of labels to apply to the nodes in the worker node group. This is in addition to the labels that EKS Anywhere will add by default.
 
Modifying the labels associated with a worker node group configuration will cause new nodes to be rolled out, replacing the existing nodes associated with the configuration.

### datacenterRef
Refers to the Kubernetes object with Nutanix environment specific configuration. See `NutanixDatacenterConfig` fields below.

### kubernetesVersion (required)
The Kubernetes version you want to use for your cluster. Supported values: `1.23`, `1.22`, `1.21`, `1.20`

## NutanixDatacenterConfig Fields

### endpoint (required)
The Prism Central server fully qualified domain name or IP address. If the server IP is used, the PC SSL certificate must have an IP SAN configured.

### port (required) 
The Prism Central server port. (Default: `9443`) 

### insecure (optional)
Set insecure to `true` if the Prism Central server does not have a valid certificate. (Default: `false`)

### additionalTrustBundle (optional; required if using a self-signed PC SSL certificate)
The PEM encoded CA trust bundle.

The `additionalTrustBundle` needs to be populated with the PEM-encoded x509 certificate of the Root CA that issued the certificate for Prism Central. Suggestions on how to obtain this certificate are [here]({{< relref "../nutanix/nutanix-prereq/#prepare-a-nutanix-environment" >}}).

## NutanixMachineConfig Fields

### cluster
Reference to the Prism Element cluster.

### cluster.type	
Type to identify the Prism Element cluster. (Permitted values: `name` or `uuid`)
 
### cluster.name	
Name or UUID of the Prism Element cluster.
 
### image	
Reference to the OS image used for the system disk.
 
### image.type
Type to identify the OS image. (Permitted values: `name` or `uuid`)
 
### image.name (`name` or `UUID` required)
Name of the image
 
### image.uuid (`name` or `UUID` required)
UUID of the image
 
### memorySize
Size of RAM on virtual machines (Default: `4Gi`)

### osFamily (optional)
Operating System on virtual machines. (Permitted values: `ubuntu`)
 
### subnet
Reference to the subnet to be assigned to the VMs.
 
### subnet.name (`name` or `UUID` required)
Name of the subnet.
 
### subnet.type
Type to identify the subnet. (Permitted values: `name` or `uuid`)
 
### subnet.uuid (`name` or `UUID` required)
UUID of the subnet.
 
### systemDiskSize
Amount of storage assigned to the system disk. (Default: `40Gi`)
 
### vcpuSockets
Amount of vCPU sockets. (Default: `2`)
 
### vcpusPerSocket
Amount of vCPUs per socket. (Default: `1`)
 
### users (optional)
The users you want to configure to access your virtual machines. Only one is permitted at this time

### users[0].name (optional)
The name of the user you want to configure to access your virtual machines through ssh.

The default is `nutanix-user` if `osFamily=ubuntu`

### users[0].sshAuthorizedKeys (optional)
The SSH public keys you want to configure to access your virtual machines through ssh (as described below). Only 1 is supported at this time.

### users[0].sshAuthorizedKeys[0] (optional)
This is the SSH public key that will be placed in `authorized_keys` on all EKS Anywhere cluster VMs so you can ssh into
them. The user will be what is defined under name above. For example:

```
ssh -i <private-key-file> <user>@<VM-IP>
```

The default is generating a key in your `$(pwd)/<cluster-name>` folder when not specifying a value

## Upgradeable Cluster Attributes 
EKS Anywhere `upgrade` supports upgrading more than just the `kubernetesVersion`, allowing you to upgrade a number of fields simultaneously with the same procedure.

### Upgradeable Attributes
 
* `Cluster`
   * `kubernetesVersion`
   * `controlPlaneConfig.count`
   * `controlPlaneConfigurations.machineGroupRef.name`
   * `workerNodeGroupConfigurations.count`
   * `workerNodeGroupConfigurations.machineGroupRef.name`
* `NutanixMachineConfig`
   * `OSFamily`
   * `VCPUsPerSocket`
   * `VCPUSockets`
   * `MemorySize`
   * `Image`
   * `Cluster`
   * `Subnet`
   * `SystemDiskSize`
 
