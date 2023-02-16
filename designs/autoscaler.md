# Autoscaling on EKS Anywhere Cluster

## Introduction

**Problem:** EKS Anywhere (EKS-A) is currently incompatible with most of the Kubernetes cluster autoscaling products on the market. Installing [Cluster Autoscaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler) in an EKS-A cluster will not work out of the box without API and design changes.

EKS Anywhere builds on top of Cluster API (CAPI) and maintains several CAPI cloud providers for cluster provisioning and management. [Cluster Autoscaler on Cluster API](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/cloudprovider/clusterapi) was chosen as the default EKS-A cluster autoscaling solution for its native CAPI support and easy integration, which avoids the need to drastically change the API definition and workflow. The following discussion will focus on Cluster Autoscaler as the autoscaling tool to integrate into EKS Anywhere.

At launch, EKS Anywhere user can specify a fixed number of nodes for each worker node group in the cluster. Once a cluster is created, EKS-A cluster controller watches this configuration for changes, generates CAPI resources, and re-applies them in the cluster. Since EKS Anywhere does not support user interaction with CAPI directly, most manual CAPI resource changes made by the user will be overwritten by the EKS-A cluster controller. This behavior prevents users from enabling Cluster Autoscaler themselves, since the required autoscaling annotations manually applied to CAPI resources will be removed by EKS-A cluster controller.

This document aims to solve this limitation and provides **autoscaling** **compatibility** in EKS Anywhere, so the user can either choose to install Cluster Autoscaler as an EKS-A curated package during or after cluster creation time, or install themselves from [upstream](https://github.com/kubernetes/autoscaler/releases) after a cluster is created.

### Goals and Objectives

As an EKS Anywhere user, I want to:

* create clusters in a streamlined fashion with autoscaling capabilities
* add autoscaling capabilities to existing EKS-A clusters
* have clusters automatically scale worker nodes up and down within the set boundaries

### Statement of Scope

#### In Scope

* Include Cluster Autoscaler as an [EKS-A curated package](https://anywhere.eks.amazonaws.com/docs/concepts/packages/).
* Integrate Cluster Autoscaler to all CAPI providers since the feature is provider agnostic.
* Integrate autoscaling configuration in cluster spec to make EKS-A cluster autoscaling-aware.
* Generate an EKS-A opinionated Cluster Autoscaler curated package template with default configurations.
* Install Cluster Autoscaler controller as a curated package during or after management cluster creation.
* Install Cluster Autoscaler controller as a curated package after workload cluster creation.
* Support upstream k8s Autoscaler installed/managed outside of EKS-A Curated Package.
* Update the autoscaling configuration to set up the min/max number of nodes in each worker node group.

#### Out of Scope

* Compare cluster autoscaler products (e.g. [Karpenter](https://karpenter.sh/))
* Autoscale at the pod level
* Control Plane nodes autoscaling

#### Future Scope

* Support other autoscaling solutions (e.g. [Karpenter](https://karpenter.sh/)).

## Overview of Solution

We provide a solution to add autoscaling configuration in the EKS Anywhere API, which marks the worker node groups to be Cluster Autoscaler compatible. EKS Anywhere user can then install Cluster Autoscaler in any EKS-A cluster with proper configuration, and expect node autoscaling to work for the associated resource group.

1. Introduce an optional type `AutoScalingConfiguration` in the cluster and populate it into each worker node group configuration.
2. If `AutoScalingConfiguration` is specified (not nil) for a worker node group, EKS-A converts its fields to CAPI autoscaler annotations and applies them to the underlying CAPI Machine Deployment resource. In this way, the specified worker node group will be discoverable by the Cluster Autoscaler controller once installed.
3. Install Cluster Autoscaler either (A) as an EKS-A curated package, or (B) from upstream; with configuration that uses Cluster API as cloud provider, and auto-discovers node groups in the target cluster.

### API Design

```go
type WorkerNodeGroupConfiguration struct {
    // Count defines the number of desired worker nodes.
    Count *int `json:"count,omitempty"`
    
    // AutoScalingConfiguration defines the auto scaling configuration
    // +optional
    AutoScalingConfiguration *AutoScalingConfiguration `json:"autoscalingConfiguration,omitempty"`
    ...
}

type AutoScalingConfiguration struct {
    // MinCount defines the minimum number of nodes for the associated resource group.
    // +optional
    MinCount int `json:"minCount,omitempty"`
    
    // MaxCount defines the maximum number of nodes for the associated resource group.
    // +optional
    MaxCount int `json:"maxCount,omitempty"`
}
```

### `AutoScalingConfiguration` Defaults and Validation

We consider the following semantics when setting initial Count that is passed down to MachineDeployment:

* If `Count` is specified, regardless the existence of `AutoScalingconfiguration`, `Count` is used as the initial count provided it passes validations outlined below.
* If `AutoScalingconfiguration` is specified, then `Count` is optional. When `Count` is not specified: `MinCount` is used as the initial count.

If `AutoScalingConfiguration` is not specified, we won’t add any annotation to the underlying MachineDeployment for the worker node group. Instead, we rely on the worker node group count to provision fixed number of nodes.

#### Validations

When `AutoScalingConfiguration` is specified, we validate

* minCount > 0, since Cluster Autoscaler with CAPI does not yet support scaling down to 0
* minCount <= count <= maxCount

### CAPI  Machine Deployment

Internally, the `minCount` and `maxCount` defined in the EKS-A cluster spec is converted to the corresponding worker node group MachineDeployment’s annotations `cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size` and `cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size`, as EKS-A builds the CAPI templates. The Cluster Autoscaler will monitor any of those MachineDeployment containing both of these annotations.

```yaml
apiVersion: cluster.x-k8s.io/v1beta1
kind: MachineDeployment
metadata:
  annotations:
    cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size: {{.maxCount}}
    cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size: {{.minCount}}
    cluster.x-k8s.io/cluster-name: {{.clusterName}}
  name: {{.workerNodeGroupName}}
  namespace: eksa-system
spec:
  clusterName: {{.clusterName}}
  replicas: {{.count}}
  ...
```

Note: the cluster will not have any autoscaling functionality until the Cluster Autoscaler controller is properly configured and installed in the cluster. Setting up the MachineDeployment annotations only guarantees that Cluster Autoscaler can recognize the worker node group and its autoscaling configuration. We will add a warning message in CLI logs when the autoscaling is configured in EKS-A cluster spec but autoscaler is not installed in the cluster. Though we won't be able to show that in an obvious way if user is creating or upgrading cluster through API (`kubectl` or GitOps).

### Cluster Autoscaler Controller

Cluster Autoscaler is a tool that automatically adjusts the size of the Kubernetes cluster. Specifically, it can:

* increase the size of the cluster when there are pending pods due to insufficient resources.
* decrease the size of the cluster when some nodes have low utilization for a significant amount of time.

Just like any other application, Cluster Autoscaler can be installed as a `Deployment` in any Kubernetes cluster. In order to connect Cluster Autoscaler to Cluster API managed EKS Anywhere cluster, a user needs to configure Cluster Autoscaler to use `clusterapi` as the cloud provider, and provide the path to the kubeconfig for the target cluster to autoscale.

Currently, it is impossible to have one autoscaler controller to watch multiple workload clusters with different kubeconfig at the same time, as there is no way currently to separate the pods between the workload clusters autoscaler is watching. The recommended method is to run multiple autoscalers in the management cluster with different options for how they discover the scalable resource sets.

Cluster Autoscaler provides the flexibility to enable autoscaling, whether it is within the cluster, or for a remote workload cluster. Cluster Autoscaler can be:

* deployed in the management cluster to autoscale the management cluster itself
* deployed in the management cluster to autoscale a remote workload cluster
* deployed in the workload cluster to autoscale the workload cluster itself

Since EKS Anywhere follows a standard management/workload clusters setup where all the CAPI machine objects are defined in the management cluster, we choose to **deploy autoscalers in the management cluster as well — to watch and manage management cluster, and all the remote workload clusters, as default option**. In this way, we have a centralized place — management cluster, to deploy and manage all the autoscaler controllers for workload clusters.

However, it is not uncommon that the management cluster size defined by the user is small, but the workload clusters are rather big, then it makes more sense to distribute the workload and run the autoscaler in each workload cluster to reduce resource usage in one management cluster. This configuration is also supported by us. Though this approach requires extra steps to create the management cluster’s kubeconfig in the workload cluster, comparing to the first approach where each workload cluster’s kubeconfig by default exists in the management cluster. Some limitations and concerns:

  1. limiting user permission to the raw CAPI components: having management cluster kubeconfig in workload cluster means workload cluster user can provision and manage clusters that are not built in a conformant, reviewed or secured manner.
  2. limiting user access to the underlying infrastructure: once non-admin workload cluster users have management cluster admin kubeconfig, which puts very high access credentials needed by CAPI. They then have high permission to the provider infrastructure (e.g vSphere: admin access to vCenter). We can certainly introduce RBAC roles to each management kubeconfig that is used in workload cluster to restrict user permission, but this is not currently supported by EKS-A.
  3. auditing and compliance of autoscaler: as the number of workload clusters arises, it is hard to version control each autoscaler deployment. Centralizing configuration makes sure compliance and configuration can be audited.
  4. lack of visibility and management: it is hard to list and visualize all the autoscaler deployed in each workload cluster at the same time without user building its own monitoring tools. User needs to find credentials for each cluster and run kubectl separately to get and manage each deployment.
  5. isolation of network connectivity: one benefit of having workload autoscaler deployed on management cluster is that we can completely cut any connection from workload to management.

EKS Anywhere user can either deploy Cluster Autoscaler controller themselves to target cluster by applying its [deployment manifest](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/clusterapi/examples/deployment.yaml) with custom arguments or using [Helm chart](https://github.com/kubernetes/autoscaler/tree/master/charts/cluster-autoscaler), or easy-install the Cluster Autoscaler as EKS-A curated package which is Amazon built, scanned, and tested. 

### EKS-A Curated Package

We will build the Cluster Autoscaler as an EKS-A curated package with an opinionated package bundle with default values. Once built and included in the package list, the autoscaler installation should fit into the standard curated package workflow. 

## User Experience

There are two requirements to enable cluster autoscaling:

1. annotate the CAPI MachineDeployment — configurable through EKS-A cluster spec
2. install Cluster Autoscaler controller in the cluster

Although it is not required, we recommend that users install the autoscaler as EKS-A curated package through the EKS-A CLI for a more streamlined installation experience and better support. There are a few possible scenarios when enabling autoscaling. 

**Enable Cluster Autoscaler as a Curated Package During Cluster Creation**

First, define a cluster spec file with autoscaling config enabled for selected worker node group(s):

```yaml
// cluster.yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: `${CLUSTER_NAME}`
spec:
  workerNodeGroupConfigurations:
  - autoscalingConfiguration:
      minCount: 1
      maxCount: 5
    machineGroupRef:
      kind: VSphereMachineConfig
      name: worker-machine-a
    name: md-0
  - count: 1
    autoscalingConfiguration:
      minCount: 1
      maxCount: 3
    machineGroupRef:
      kind: VSphereMachineConfig
      name: worker-machine-b
    name: md-1
...
```

Make sure cluster-autoscaler is available to install as an EKS-A package:

```sh
$ eksctl anywhere list packages --source registry --kube-version 1.21

// Example command output
Package                 Version(s)
-------                 ----------
cluster-autoscaler      1.0.0-164563266364sh4209000f3dgfea84eed36529sa
harbor                  2.5.0-4324383d8c5383bded5f7378efb98b4d50af827b
...
```

Generate a curated-package config for autoscaler and run create command with it plus the cluster spec:

```sh
$ eksctl anywhere generate package cluster-autoscaler --cluster test-cluster > cas-package.yaml
```

a) enables autoscaling in management cluster, where the autoscaler controller is deployed in management cluster and watches the management cluster itself.

```yaml
// cas-package.yaml
apiVersion: packages.eks.amazonaws.com/v1alpha1
kind: Package
metadata:
  name: cluster-autoscaler-management-cluster
  namespace: eksa-packages
spec:
  packageName: cluster-autoscaler
  config:
    cloudProvider: "clusterapi"
    autoDiscovery:
      clusterName: "management-cluster"
```

```sh
$ eksctl anywhere create cluster -f mgmt-cluster.yaml --install-packages cas-package.yaml
```

This command goes through the following steps:

1. create a new EKS-A cluster with proper autoscaling annotations on worker resources
2. install curated package controller on the management cluster
3. the package controller on the management cluster then installs the Cluster Autoscaler package on management cluster

b) enables autoscaling in workload cluster, where the autoscaler controller is deployed in management cluster and watches a remote workload cluster.

```yaml
// cas-package.yaml
apiVersion: packages.eks.amazonaws.com/v1alpha1
kind: Package
metadata:
  name: cluster-autoscaler-workload-cluster
  namespace: eksa-packages
spec:
  packageName: cluster-autoscaler
  config:
    cloudProvider: "clusterapi"
    autoDiscovery:
      clusterName: "workload-cluster"
    clusterAPIMode: "kubeconfig-incluster"
    clusterAPIKubeconfigSecret: "workload-cluster-kubeconfig"
    clusterAPIWorkloadKubeconfigPath: "/etc/kubernetes/value"
```

** *The command for this case — installing a package on the management cluster during a workload cluster creation — is TBD.*

**Enable Cluster Autoscaler as a Curated Package in Existing EKS-A Cluster**

If the original cluster spec does not have `AutoScalingConfiguration` defined, add it to the cluster spec and run upgrade command:

```sh
$ eksctl anywhere upgrade cluster -f cluster.yaml
```

If the original cluster does not have curated package controller installed:

```sh
$ eksctl anywhere install packagecontroller -f cluster.yaml
```

Install Cluster Autoscaler curated package:

```sh
$ eksctl anywhere create packages -f cas-package.yaml
```

*Notice if choosing to deploy autoscaler on workload cluster itself, user needs to manually store the management cluster’s kubeconfig as a secret in the workload cluster before running the `create packages` command above, with the package config:*

```yaml
// cas-package.yaml
apiVersion: packages.eks.amazonaws.com/v1alpha1
kind: Package
metadata:
  name: cluster-autoscaler-workload-cluster
  namespace: eksa-packages
spec:
  packageName: cluster-autoscaler
  config:
    cloudProvider: "clusterapi"
    autoDiscovery:
      clusterName: "workload-cluster"
    clusterAPIMode: "incluster-kubeconfig"
    clusterAPICloudConfigPath: "/etc/kubernetes/value"
    extraVolumeSecrets:
      cluster-autoscaler-cloud-config:
        mountPath: "/etc/kubernetes"
        name: "management-cluster-kubeconfig"
```

**Use Upstream Cluster Autoscaler in EKS-A Cluster**

Create a cluster with `AutoScalingConfiguration` specified for the worker node group(s) you want to enable autoscaling:

```sh
$ eksctl anywhere create cluster -f cluster.yaml
```

**Or** upgrade a cluster with `AutoScalingConfiguration` specified for the worker node group(s):

```sh
$ eksctl anywhere upgrade cluster -f cluster.yaml
```

Install Cluster Autoscaler from upstream following the [official instruction](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/clusterapi/README.md#connecting-cluster-autoscaler-to-cluster-api-management-and-workload-clusters). Deploy autoscaler as a [Deployment with proper RBAC](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/clusterapi/examples/deployment.yaml):

```yaml
// cas-deployment.yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cluster-1-autoscaler-management
  namespace: ${AUTOSCALER_NS}
  labels:
    app: cluster-autoscaler
spec:
  selector:
    matchLabels:
      app: cluster-autoscaler
  replicas: 1
  template:
    metadata:
      labels:
        app: cluster-autoscaler
    spec:
      containers:
      - image: ${AUTOSCALER_IMAGE} # that matches the K8s version. https://github.com/kubernetes/autoscaler/releases
        name: cluster-autoscaler
        command:
        - /cluster-autoscaler
        args:
        - --cloud-provider=clusterapi
        - --node-group-auto-discovery=clusterapi:clusterName=cluster-1
      serviceAccountName: cluster-autoscaler
      terminationGracePeriodSeconds: 10
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cluster-autoscaler-workload
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-autoscaler-workload
subjects:
- kind: ServiceAccount
  name: cluster-autoscaler
  namespace: ${AUTOSCALER_NS}
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cluster-autoscaler-management
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-autoscaler-management
subjects:
- kind: ServiceAccount
  name: cluster-autoscaler
  namespace: ${AUTOSCALER_NS}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cluster-autoscaler
  namespace: ${AUTOSCALER_NS}
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cluster-autoscaler-workload
rules:
  - apiGroups:
    - ""
    resources:
    - namespaces
    - persistentvolumeclaims
    - persistentvolumes
    - pods
    - replicationcontrollers
    - services
    verbs:
    - get
    - list
    - watch
  - apiGroups:
    - ""
    resources:
    - nodes
    verbs:
    - get
    - list
    - update
    - watch
  - apiGroups:
    - ""
    resources:
    - pods/eviction
    verbs:
    - create
  - apiGroups:
    - policy
    resources:
    - poddisruptionbudgets
    verbs:
    - list
    - watch
  - apiGroups:
    - storage.k8s.io
    resources:
    - csinodes
    - storageclasses
    - csidrivers
    - csistoragecapacities
    verbs:
    - get
    - list
    - watch
  - apiGroups:
    - batch
    resources:
    - jobs
    verbs:
    - list
    - watch
  - apiGroups:
    - apps
    resources:
    - daemonsets
    - replicasets
    - statefulsets
    verbs:
    - list
    - watch
  - apiGroups:
    - ""
    resources:
    - events
    verbs:
    - create
    - patch
  - apiGroups:
    - ""
    resources:
    - configmaps
    verbs:
    - create
    - delete
    - get
    - update
  - apiGroups:
    - coordination.k8s.io
    resources:
    - leases
    verbs:
    - create
    - get
    - update
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cluster-autoscaler-management
rules:
  - apiGroups:
    - cluster.x-k8s.io
    resources:
    - machinedeployments
    - machinedeployments/scale
    - machines
    - machinesets
    verbs:
    - get
    - list
    - update
    - watch
```

```sh
$ kubectl apply -f cas-deployment.yaml
```

## Testing

Add E2E tests for the below cases:

1. Create an autoscaling enabled management cluster
    1. generate autoscaling package config
    2. create a cluster with `AutoScalingConfiguration` in cluster spec and package config above
    3. scale up workload applications to create pending pods
    4. verify new nodes being created
    5. scale down workload applications
    6. wait and verify node size decreases
2. Enable autoscaling in an existing management cluster
    1. create a cluster without autoscaling feature
    2. generate autoscaling package config
    3. install curated package controller
    4. install autoscaling package
    5. scale workload applications up and down, and verify node size changes
3. Enable autoscaling in an existing workload cluster
    1. create a management cluster
    2. create a workload cluster
    3. generate autoscaling package config for workload cluster
    4. install curated package controller in management cluster
    5. apply autoscaling package config in management cluster
    6. scale workload applications in workload cluster up and down, and verify its node size changes

