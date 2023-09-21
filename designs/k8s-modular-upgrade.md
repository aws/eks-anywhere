# Kubernetes Version Modular Upgrades

## Introduction

Currently, when you upgrade the Kubernetes version of an EKS-A Cluster, the entire cluster is upgraded including the control plane, worker, etcd nodes, etc. Upgrading k8s can roll out new worker nodes which would disrupt pods running on those nodes. Customers should be able to upgrade worker nodes separately from the control plane, so that they can schedule upgrades to worker nodes at more convenient times.

## Goals

As a user, I would like to upgrade the control plane on my cluster without also upgrading all of my worker nodes. I would like to upgrade each worker node group separately from each other and control plane nodes. 

## Proposed Solution

Cluster configuration files should separate the kubernetesVersion field into two parts for control plane and worker node groups. We would then have to change the logic for the generation of machine deployments to use its respective k8s version. 

The field for worker node groups should be opt in and default to the top level kubernetesVersion value if not specified. When not specified, the controller and provider specific code should attempt to use the top level version and the secondary version should remain empty. Users should also be able to remove the secondary field if they wish to allow that worker node group to be upgraded with the control plane. 

## Implementation Details

The current kubernetesVersion field should be kept to upgrade the control plane of a cluster and any worker node groups that don’t have their own kubernetesVersion field. A secondary kubernetesVersion field should be added under worker node group configurations for the purpose of upgrading only the specified worker nodes group in the cluster. 

The second k8s version field should have similar validations as the top level field such as an upgrade skew of one compared to old values. We also need validations for the top level field that ensure its semver is either equal to or not greater than 2 minor versions from the older worker node groups. An additional validation may be needed to prevent users from removing the secondary field if it would cause an upgrade skew greater than two.

Since kubernetesVersion is used to create an EKSD Release and VersionsBundle, we would need to create additional instances of each of these types within cluster.Spec for the secondary kubernetesVersion and store them in a map to retrieve the correct VersionsBundle for each node's respective Kubernetes version.

The apibuilder would need to be changed to use the new instance of the VersionsBundle when creating the machine deployment. 

Additionally, each provider’s NeedNewWorkloadTemplate needs to use the workerKubernetesVersion field and buildTemplateMapMD would also need to use the new VersionsBundle.


```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: mgmt
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
    count: 1
  datacenterRef:
    kind: DockerDatacenterConfig
    name: mgmt
  externalEtcdConfiguration:
    count: 1
  kubernetesVersion: "1.26"
  managementCluster:
    name: mgmt
  workerNodeGroupConfigurations:
  - count: 1
    name: md-0
  - count: 1
    kubernetesVersion: "1.25"
    name: md-1 
```

```golang
type WorkerNodeGroupConfiguration struct {
    // Name refers to the name of the worker node group
    Name string `json:"name,omitempty"`
    // Count defines the number of desired worker nodes. Defaults to 1.
    Count *int `json:"count,omitempty"`
    // AutoScalingConfiguration defines the auto scaling configuration
    AutoScalingConfiguration *AutoScalingConfiguration `json:"autoscalingConfiguration,omitempty"`
    // MachineGroupRef defines the machine group configuration for the worker nodes.
    MachineGroupRef *Ref `json:"machineGroupRef,omitempty"`
    // Taints define the set of taints to be applied on worker nodes
    Taints []corev1.Taint `json:"taints,omitempty"`
    // Labels define the labels to assign to the node
    Labels map[string]string `json:"labels,omitempty"`
    // UpgradeRolloutStrategy determines the rollout strategy to use for rolling upgrades
    // and related parameters/knobs
    UpgradeRolloutStrategy *WorkerNodesUpgradeRolloutStrategy `json:"upgradeRolloutStrategy,omitempty"`
    
    // ADD THIS
    KubernetesVersion *KubernetesVersion `json:"kubernetesVersion,omitempty"`
}
```

```golang
type Spec struct {
    *Config
    Bundles           *v1alpha1.Bundles
    // VersionsBundle    *VersionsBundle <-- remove this
    eksdRelease       *eksdv1alpha1.Release
    OIDCConfig        *eksav1alpha1.OIDCConfig
    AWSIamConfig      *eksav1alpha1.AWSIamConfig
    ManagementCluster *types.Cluster 
    VersionsBundles   map[eksav1alpha1.KubernetesVersion]*VersionsBundle // <-- Add this
}
```
