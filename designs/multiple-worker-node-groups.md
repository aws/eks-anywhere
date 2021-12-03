# Multiple worker node groups

## Introduction

**Problem:** Currently users of eks-anywhere can only create one worker node group by specifying the configuration details in a cluster spec.

This limits the ability of a user to create worker nodes with different configurations and formulate application deployment strategy based on the configurations of the worker nodes.

Also, this creates a problem while tainting a node. Unless, eks-anywhere clusters support multiple worker node groups, worker nodes can not be tainted with `NoExecute` effect, since nodes in a node group have the same configuration, setting a taint with `NoExecute` effect would essentially make the worker nodes unusable for eks-anywhere specific deployments, as the general approach for eks-anywhere is to not add tolerations on deployments.


### Tenets

***Simple:*** Specifying different node group configurations in the cluster spec should be simple and readable.

### Goals and Objectives

As a Kubernetes administrator I want to:

* Add multiple worker node group configurations in the `workerNodeGroupConfigurations` array in cluster specs.
* Have the ability to point each worker node group configuration to a different machine configuration.
* Specify separate node counts and taints information for each node group.  

### Statement of Scope

**In scope**

* Providing users the ability to add multiple worker node groups configuration in the cluster spec and bootstrap kubernetes clusters with multiple worker node groups.

## Overview of Solution

With this feature, a user can create a cluster config file with multiple worker node group configurations. Upon running eks-anywhere cli, these info will be fetched and be added in the capi specification file. In the cape spec, configuration details will be appended for each worker node groups one after another. Examples of each of these two files are added in the next section.
 
## Solution Details

With this feature, cluster spec file will look like below.

```
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  creationTimestamp: null
  name: eksa-test
spec:
  clusterNetwork:
    cni: cilium
    pods:
    ...
  controlPlaneConfiguration:
    count: 3
    ...
    machineGroupRef:
      kind: VSphereMachineConfig
      name: eksa-test-cp
      ...
  datacenterRef:
    kind: VSphereDatacenterConfig
    name: eksa-test
  kubernetesVersion: "1.20"
  workerNodeGroupConfigurations:
  - count: 3
    machineGroupRef:
      kind: VSphereMachineConfig
      name: eksa-test-wl-1
      Taints:
      - key: Key2
        value: value2
        effect: PreferNoSchedule
  - count: 3
    machineGroupRef:
      kind: VSphereMachineConfig
      name: eksa-test-wl-2
      Taints:
      - key: Key3
        value: value3
        effect: PreferNoSchedule
status: {}

---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: VSphereDatacenterConfig
metadata:
  creationTimestamp: null
  name: eksa-test
spec:
....
status: {}

---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: VSphereMachineConfig
metadata:
  creationTimestamp: null
  name: eksa-test-cp
spec:
...
status: {}

---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: VSphereMachineConfig
metadata:
  creationTimestamp: null
  name: eksa-test-wl-1
spec:
...
status: {}

---

---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: VSphereMachineConfig
metadata:
  creationTimestamp: null
  name: eksa-test-wl-2
spec:
...
status: {}

---

---
```

Once it is processed through cli, the generated capi spec file for worker nodes should look like below.

```
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  labels:
    cluster.x-k8s.io/cluster-name: eksa-test
  name: eksa-test
  namespace: eksa-system
spec:
  ...
  controlPlaneRef:
    apiVersion: controlplane.cluster.x-k8s.io/v1alpha3
    kind: KubeadmControlPlane
    name: eksa-test
  infrastructureRef:
    apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
    kind: VSphereCluster
    name: eksa-test
---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: VSphereCluster
metadata:
  name: eksa-test
  namespace: eksa-system
spec:
  cloudProviderConfiguration:
  ...
---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: VSphereMachineTemplate
metadata:
  name: eksa-test-control-plane-template-1638469395664
  namespace: eksa-system
spec:
  template:
    spec:
    ...
---
apiVersion: controlplane.cluster.x-k8s.io/v1alpha3
kind: KubeadmControlPlane
metadata:
  name: eksa-test
  namespace: eksa-system
spec:
  infrastructureTemplate:
    apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
    kind: VSphereMachineTemplate
    name: eksa-test-control-plane-template-1638469395664
  kubeadmConfigSpec:
    clusterConfiguration:
    ...
    initConfiguration:
      nodeRegistration:
        criSocket: /var/run/containerd/containerd.sock
        kubeletExtraArgs:
          cloud-provider: external
          tls-cipher-suites: Something
        name: '{{ ds.meta_data.hostname }}'
        taints: 
          - key: key1
            value: val1
            effect: PreferNoSchedule
	  ....
    joinConfiguration:
      pause:
        imageRepository: public.ecr.aws/eks-distro/kubernetes/pause
        imageTag: v1.20.7-eks-1-20-8
      bottlerocketBootstrap:
        imageRepository: public.ecr.aws/l0g8r8j6/bottlerocket-bootstrap
        imageTag: v1-20-8-eks-a-v0.0.0-dev-build.579
      nodeRegistration:
        criSocket: /var/run/containerd/containerd.sock
        kubeletExtraArgs:
          cloud-provider: external
          tls-cipher-suites: Something
        name: '{{ ds.meta_data.hostname }}'
        taints: 
          - key: key1
            value: val1
            effect: PreferNoSchedule
	  ...
---
apiVersion: addons.cluster.x-k8s.io/v1alpha3
kind: ClusterResourceSet
metadata:
  labels:
    cluster.x-k8s.io/cluster-name: eksa-test
  name: eksa-test-crs-0
  namespace: eksa-system
spec:
...
---
apiVersion: v1
kind: Secret
metadata:
  name: vsphere-csi-controller
  namespace: eksa-system
stringData:
  data: |
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: vsphere-csi-controller
      namespace: kube-system
type: addons.cluster.x-k8s.io/resource-set
---
apiVersion: v1
data:
...
kind: ConfigMap
metadata:
  name: vsphere-csi-controller-role
  namespace: eksa-system
---
apiVersion: v1
data:
  data: |
  ...
kind: ConfigMap
metadata:
  name: vsphere-csi-controller-binding
  namespace: eksa-system
---
apiVersion: v1
data:
  data: |
  ...
kind: ConfigMap
metadata:
  name: csi.vsphere.vmware.com
  namespace: eksa-system
---
apiVersion: v1
data:
  data: |
    apiVersion: apps/v1
    kind: DaemonSet
    metadata:
      name: vsphere-csi-node
      namespace: kube-system
    spec:
      selector:
        matchLabels:
          app: vsphere-csi-node
      template:
        metadata:
          labels:
            app: vsphere-csi-node
            role: vsphere-csi
        spec:
	...
kind: ConfigMap
metadata:
  name: vsphere-csi-node
  namespace: eksa-system
---
apiVersion: v1
data:
  data: |
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: vsphere-csi-controller
      namespace: kube-system
    spec:
    ...
kind: ConfigMap
metadata:
  name: vsphere-csi-controller
  namespace: eksa-system
---
apiVersion: v1
data:
  data: |
    apiVersion: v1
    data:
      csi-migration: "false"
    kind: ConfigMap
    metadata:
      name: internal-feature-states.csi.vsphere.vmware.com
      namespace: kube-system
kind: ConfigMap
metadata:
  name: internal-feature-states.csi.vsphere.vmware.com
  namespace: eksa-system

---
apiVersion: bootstrap.cluster.x-k8s.io/v1alpha3
kind: KubeadmConfigTemplate
metadata:
  name: eksa-test-md-0
  namespace: eksa-system
spec:
  template:
    spec:
      joinConfiguration:
        pause:
          imageRepository: public.ecr.aws/eks-distro/kubernetes/pause
          imageTag: v1.20.7-eks-1-20-8
        bottlerocketBootstrap:
          imageRepository: public.ecr.aws/l0g8r8j6/bottlerocket-bootstrap
          imageTag: v1-20-8-eks-a-v0.0.0-dev-build.579
        nodeRegistration:
          criSocket: /var/run/containerd/containerd.sock
          kubeletExtraArgs:
            cloud-provider: external
            tls-cipher-suites: Something
          name: '{{ ds.meta_data.hostname }}'
	  ...
---
apiVersion: cluster.x-k8s.io/v1alpha3
kind: MachineDeployment
metadata:
  labels:
    cluster.x-k8s.io/cluster-name: eksa-test
  name: eksa-test-md-0
  namespace: eksa-system
spec:
  clusterName: eksa-test
  replicas: 3
  selector:
    matchLabels: {}
  template:
    metadata:
      labels:
        cluster.x-k8s.io/cluster-name: eksa-test
    spec:
      bootstrap:
        configRef:
          apiVersion: bootstrap.cluster.x-k8s.io/v1alpha3
          kind: KubeadmConfigTemplate
          name: eksa-test-md-0
      clusterName: eksa-test
      infrastructureRef:
        apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
        kind: VSphereMachineTemplate
        name: eksa-test-worker-node-template-1638469395669
      version: v1.20.7-eks-1-20-8
---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: VSphereMachineTemplate
metadata:
  name: eksa-test-worker-node-template-1638469395669
  namespace: eksa-system
spec:
  template:
    spec:
    ...
---
apiVersion: bootstrap.cluster.x-k8s.io/v1alpha3
kind: KubeadmConfigTemplate
metadata:
  name: eksa-test-md-1
  namespace: eksa-system
spec:
  template:
    spec:
      joinConfiguration:
        pause:
          imageRepository: public.ecr.aws/eks-distro/kubernetes/pause
          imageTag: v1.20.7-eks-1-20-8
        bottlerocketBootstrap:
          imageRepository: public.ecr.aws/l0g8r8j6/bottlerocket-bootstrap
          imageTag: v1-20-8-eks-a-v0.0.0-dev-build.579
        nodeRegistration:
          criSocket: /var/run/containerd/containerd.sock
          kubeletExtraArgs:
            cloud-provider: external
            tls-cipher-suites: Something
          name: '{{ ds.meta_data.hostname }}'
	  ...
---
apiVersion: cluster.x-k8s.io/v1alpha3
kind: MachineDeployment
metadata:
  labels:
    cluster.x-k8s.io/cluster-name: eksa-test
  name: eksa-test-md-1
  namespace: eksa-system
spec:
  clusterName: eksa-test
  replicas: 3
  selector:
    matchLabels: {}
  template:
    metadata:
      labels:
        cluster.x-k8s.io/cluster-name: eksa-test
    spec:
      bootstrap:
        configRef:
          apiVersion: bootstrap.cluster.x-k8s.io/v1alpha3
          kind: KubeadmConfigTemplate
          name: eksa-test-md-1
      clusterName: eksa-test
      infrastructureRef:
        apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
        kind: VSphereMachineTemplate
        name: test
      version: v1.20.7-eks-1-20-8
---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: VSphereMachineTemplate
metadata:
  name: test
  namespace: eksa-system
spec:
  template:
    spec:
    ...
---
```

For each worker node groups, capi spec file will continue to have the following 3 kind fields.

* KubeadmConfigTemplate
* Machime deployment
* VSphereMachineTemplate

For each group, we will append these three fields corresponding to that group in the capi spec.

Right now, the cli assumes that there will be only one group and it treats worker node group configuration array as a collection of only one element. As a result, the controller just refers to the first element of this array in different places of the code. So we need to do the same operations in loops, which includes capi spec creation, cluster spec validation etc. Once a capi spec is created with this approach, the workload cluster will be created with multiple worker nodes.

The examples in this design are for vsphere provider. But the same strategy applies for other providers as well.

## Testing

To make sure, the implementation of this feature is correct, we need to add unit tests for each providers to make sure, the capi specs are generated properly.
Also, we need to add e2e tests for each providers to test the following scenarios.

* Cluster creation with one worker node group
* Cluster creation with multiple worker node groups

## Conclusion

Current implementation of eks-anywhere cli implements an array of worker node group configurations and assumes that the array has only one element. With this design we can enhance the scope the current implementation to make sure it can handle multiple elements in that array. This will help us achieve our goal to support multiple worker node groups.
