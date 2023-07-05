# Kubernetes Version Modular Upgrades

## Introduction

Currently, when you upgrade the Kubernetes version of an EKS-A Cluster, the entire cluster is upgraded including the control plane, worker, etcd nodes, etc. Upgrading k8s can roll out new worker nodes which would disrupt pods running on those nodes and cause unwanted downtime. 
Customers should be able to schedule upgrades to worker nodes separately from the control plane, so that they have increased control over when downtime occurs without sacrificing upgrades. 

## Goals

As a user, I would like to upgrade the control plane on my cluster without experiencing any downtime on my workloads. I would like to upgrade worker nodes separately from control plane nodes. 

## Proposed Solution

Cluster configuration files should separate the kubernetesVersion field into two parts for control plane and worker node groups. We would then have to change the logic for the generation of machine deployments to use its respective k8s version. 

The field for worker node groups can be optional and should default to the top level kubernetesVersion value if not specified. The field should be filled in for the user if they don’t provide a value. 

## Implementation Details

The current kubernetesVersion field should be kept to upgrade only the control plane of a cluster. A secondary kubernetesVersion field should be added under each worker node group configuration for the purpose of upgrading only that worker nodes group in the cluster. The second k8s version field should have similar validations as the original field.
We also need validations for the top level field that ensure its semver is either equal to or not greater than 1 minor version from the older worker node groups. 

Since [kubernetesVersion](https://github.com/tatlat/eks-anywhere/blob/main/pkg/cluster/fetch.go#L153) is used to create an [EKSD Release and VersionsBundle](https://github.com/aws/eks-anywhere/blob/ab8bea7667b598ce7500d49b0a3d4726f0775c2a/pkg/cluster/spec.go#L40), we would need to create additional instances of each of these types within cluster.Spec for the secondary kubernetesVersion and store them in a map to retrieve the correct VersionsBundle for each worker node group.

The [apibuilder](https://github.com/aws/eks-anywhere/blob/ab8bea7667b598ce7500d49b0a3d4726f0775c2a/pkg/clusterapi/apibuilder.go#L244) would need to be changed to use the new instance of the VersionsBundle when creating the machine deployment. 

Additionally, each provider’s [NeedNewWorkloadTemplate](https://github.com/aws/eks-anywhere/blob/ab8bea7667b598ce7500d49b0a3d4726f0775c2a/pkg/providers/docker/docker.go#L346) needs to use the workerKubernetesVersion field and [buildTemplateMapMD](https://github.com/aws/eks-anywhere/blob/ab8bea7667b598ce7500d49b0a3d4726f0775c2a/pkg/providers/docker/docker.go#L307) would also need to use the new VersionsBundle.

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
    VersionsBundle    *VersionsBundle
    eksdRelease       *eksdv1alpha1.Release
    OIDCConfig        *eksav1alpha1.OIDCConfig
    AWSIamConfig      *eksav1alpha1.AWSIamConfig
    ManagementCluster *types.Cluster 
    workerVersions    map[string]*WorkerVersions // <-- ADD THIS
}

type WorkerVersions struct {
   VersionsBundle *VersionsBundle
   eksdRelease *eksdv1alpha1.Release
}
```
