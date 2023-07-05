# Cluster Status Proposal

## Introduction

**Problem:** The Full Cluster Lifecycle API allows users to manage (creating, updating and deleting) EKS Anywhere clusters with a native Kubernetes experience we want users to have. However, it is unclear and somewhat arduous to assess the state of cluster managed in this way, unlike when using the CLI to manage workload clusters where the state of any operation is made known through the run CLI run logs and the user is notified through success and failure logs.

**Context:** Currently, users must engage with Kubernetes resources that are not native to EKS Anywhere API to assess or debug a the state of a cluster. Additionally, there are multiple ways to do this, which can be inefficient, confusing and overall a bad user experience, depending on how it is done.

For example, when the user creates/updates EKS Anywhere workload cluster with the API, the user may do some of the following:

1. Check the status of their cluster by looking at the Conditions of the CAPI cluster and the related machine deployments.
2. Or get the CAPI Machine objects with `k get machines -A` to check that the machines specified in the `Cluster` spec are created and that the phases transition from `Provisioning` to `Running`, are using the correct k8s versions, etc and check the controller logs if something does not look right.
3. Or focus on watching logs in the controllers such as the `eksa-controller-manager` to monitor the reconciliation process. If there is an error in the logs here, this may give the user a hint to what the issue is, or where to dive deeper into the logs of other pods. For example, if CAPI control plane is not ready indefinitely, the user may then check the   `capi-kubeadm-control-plane-controller-manager` or `capi-kubeadm-bootstrap-controller-manager` for details. This could also cause some confusion because controllers WILL have logs of transient errors.

## Tenets
1. Independent: independent interface that abstracts the internal system details from the user
2. Simple: simple to use, simple to understand, simple to maintain

## Goals
To improve the user experience, EKS Anywhere needs a direct way of verifying the current state of the cluster through the API. For this design, we are concerned with the following users and their experience:

* EKS-Anywhere Cluster Administrator
* EKS Anywhere CLI


As an EKS-Anywhere cluster administrator, Larry needs to be able to:

* identify when the control plane is available, so that the kubeconfig  can be retrieved and used to contact the workload cluster with kubectl commands.
* easily verify when a cluster is ready for use, so that they may start deploying workloads to the cluster.
* see the name and version of the managed CNI installed when describing EKS-Anywhere cluster, so that he is able to quickly verify the installation while debugging possibly related networking issues.
* view information on control plane and worker node scaling when managing a cluster through the API, so that he can debug related issues where the operation fails and hangs.
* understand when and why an operation issued via the API succeeded or failed.

As the EKS-Anywhere CLI, it needs to be able to:

* report that the cluster has been successfully created or updated, so that the user can start scheduling workloads.
* report that the cluster has been successfully deleted, so that the user is aware.
* report that the input cluster spec is valid or not, so the user knows immediately if they must fix their `Cluster` spec.
* when the CAPI control plane to be available, so that it may retrieve the kubeconfig from the cluster and write it to disk.
* report through the terminal that the default networking solution is being installed, up to date, or not vs using a custom CNI.
* report progress on waiting for control plane and worker machines to be ready, so the user is aware does not think they’re stuck if it takes a long time.
* surface failures while trying to provisioning a new or updating an existing cluster after some time, so the user can respond as soon as possible.


## Scope 

### In

* Define the information necessary to perform lifecycle operations through the API.
* Design a solution to make all the required information needed to perform lifecycle operations through the API available to the user.

### Out

* Refactor the  EKS Anywhere CLI to use the Full Lifecycle API to manage workload cluster.
* Add cluster information to the table output when describing a cluster with kubectl.
* Surface cluster health issues to provide users enhanced visibility into the health of their cluster infrastructure. 
* Emitting metrics for this cluster status information.
* Add analyzers to surface this new cluster status information to support bundles.

## Current State
The current `ClusterStatus` is used to store a descriptive `ClusterStatus.FailureMessage` about a fatal problem while reconciling a cluster that the user can check to help when troubleshooting.

```
// ClusterStatus defines the observed state of Cluster.
type ClusterStatus struct {
    // Descriptive message about a fatal problem while reconciling a cluster
    // +optional
    FailureMessage string json:"failureMessage,omitempty"
    // EksdReleaseRef defines the properties of the EKS-D object on the cluster
    EksdReleaseRef EksdReleaseRef json:"eksdReleaseRef,omitempty"
    // +optional
    Conditions []clusterv1.Condition json:"conditions,omitempty"
    // ReconciledGeneration represents the .metadata.generation the last time the
    // cluster was successfully reconciled. It is the latest generation observed
    // by the controller.
    // NOTE: This field was added for internal use and we do not provide guarantees
    // to its behavior if changed externally. Its meaning and implementation are
    // subject to change in the future.
    ReconciledGeneration int64 `json:"reconciledGeneration,omitempty"`
    // ChildrenReconciledGeneration represents the sum of the .metadata.generation
    // for all the linked objects for the cluster, observed the last time the
    // cluster was successfully reconciled.
    // NOTE: This field was added for internal use and we do not provide guarantees
    // to its behavior if changed externally. Its meaning and implementation are
    // subject to change in the future.
    ChildrenReconciledGeneration int64 `json:"childrenReconciledGeneration,omitempty"`
}
```

E.g. of a workload clusters.anywhere.eks.amazonaws.com object test-docker-w01 with a status displaying a failureMessage

```
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: test-docker-w01
  ...
spec:
 ...
status:
  failureMessage: 'Dependent cluster objects don''t exist: DockerDatacenterConfig.anywhere.eks.amazonaws.com
    "test-docker-w01" not found'
```    


The `ClusterStatus.FailureMessage` field assigned a descriptive message when there is a failure reconciling the cluster. The other fields below go unused:

* `ClusterStatus.EksdReleaseRef`
* `ClusterStatus.Conditions`



## Solution
Use the `Cluster` status to represent the current state of the cluster by adding new fields that show a more accurate state of the  cluster after each reconciliation loop . This would allow the user more visibility so that they may responding accordingly to that information while performing management operations. 

**How do we ensure the viewed status represents the `Cluster` spec?**

The cluster spec represents the desired state of the cluster and the status  should represent the actual state of the cluster according to the Kubernetes [api conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status). These are separate objects in Kubernetes and updated independently, so there can be situations where the user views the `Cluster` status that represents an old spec. To clarify this ambiguity, we can use the Kubernetes concepts of generation and observedGenerations.

* `metadata.generation`: a monotonically increasing number that gets bumped every time you update your resource
* `status.observedGeneration`: owned an set by the controller when the status is updated on every reconciliation attempt. If this is not the same as the metadata.generation it means that the status being view represents an old generation of the spec.

There is currently two fields in the `Cluster` status that observe the generation of the EKS Anywhere cluster resource and its linked objects. They are 

* `reconciledGeneration` - represents the .metadata.generation the last time the cluster was successfully reconciled. It is the latest generation observed by the controller.
* `childrenReconciledGeneration` -  represents the sum of the .metadata.generation for all the linked objects for the cluster, observed the last time the cluster was successfully reconciled. 

These fields are marked as internal use and subject to change based on other internal designs and decisions. Note that the `reconciledGeneration` field offers something similar to what we need, however, you should note that it is only set when the cluster has been successfully reconciled. That means, the status would be updated to reflect the failure but the `reconciledGeneration` field would not be when there is a failed during reconciliation. This is an unconventional way of implementing the idea of a latest observed generation of a resource in a controller and it is not what we want to solve the problem of clarifying whether the current status represents the desired cluster spec update. We propose adding a new `observedGeneration` field because we want one field that is set every reconciliation loop to accurately show that the status is for the desired cluster spec even in cases where reconciliation fails. 

**What information do we need to perform lifecycle operations?**

Now, we need to define information is needed to perform lifecycle management operations. We propose the need for the following:

* The `Cluster` spec is valid - When the user submits a cluster spec, there may be provider specific validations at runtime for that check if the `Cluster` spec submitted is valid or not. We currently bubble up this up to the `status.failureMessage` as a terminal reconciliation failure, so nothing more needs to be done for this.
* The Control Plane has been initialized - The control plane houses components that make decisions about the cluster server, for example, the kubeapi server which exposes the Kubernetes API. The cluster is not contactable until the kubeapi server is available. The user needs to know when the first control plane is available, and therefore possible to contact the cluster using the kubeconfig that can be retrieved from the management cluster to contact the newly created cluster.
* The Control Plane is ready - the control plane can only be initialized once, but the overall condition of the control plane can change. So, we need another indicator of the current readiness of the cluster’s control plane.
* Default managed CNI networking solution is installed or not - By default, EKS Anywhere installs and manages a networking solution for the user, who may want to see this information while viewing their cluster to make informed decisions  and debug issues related to networking.
* Control Plane and Worker machines are ready - A number of control plane and worker nodes are scaled up and down, depending on the desired state given in the `Cluster` spec during an operation. Any number of these nodes may be in transitionary unusable states during that period of time, so it is important for the user to know the current vs total expected number of nodes that are ready is important to enable them to make informed follow up operations like scheduling workloads and even for debugging issues.
* Cluster is ready - The cluster is fully ready and  matches the desired state of the `Cluster` spec.


We would also deprecate any existing unnecessary or unused fields from the `Cluster` status.

## Solution Details

### API changes

**Adding new fields to the `Cluster` status**

We propose adding the following fields to the Cluster status:

```
// ClusterStatus defines the observed state of Cluster.
type ClusterStatus struct {
    ...
    // Machine readable value about a terminal problem while reconciling the cluster
    // set at the same time as failureMessage
    // +optional
    FailureReason string `json:"failureReason,omitempty"`
   
    // information on the EKS Anywhere CNI configured
    // +optional
    DefaultCNI ClusterCNI`json:"components,omitempty"`
        
    // ObservedGeneration is the latest generation observed by the controller
    // set everytime the status is updated
    ObservedGeneration int64`json:"observedGeneration,omitempty"`
  }
```

```
// ClusterCNI represents a CNI cluster component.
type ClusterCNI struct {
    // E.g. cilium
    Name string`json:"name,omitempty"`
    
    // Empty if not applied or no version can be inferred.
    // +optional
    Version string`json:"version,omitempty"`
    
    // Applied, Not Applied
    Status string`json:"status,omitempty"`
}
```

**Default CNI field**

EKS Anywhere deploys EKS Anywhere Cilium as the default managed CNI option. However, this design is informed by the fact that we do support other CNI solutions like `kindnetd`. Although, this is only for docker and thus, non-production clusters, EKS Anywhere has the capability for configuring managed CNI solutions beyond cilium and can possibly support more in the future.

The user may need information on managed CNI component in the `Cluster` status for a few reasons. First, the name and version of CNI component installed may serve as initial point for debugging related networking issues. Users may also opt to use a custom CNI instead of the default we use in EKS Anywhere (Cilium), where at first Cilium is installed on cluster creations because it’s necessary for provisioning. However, after the user would have to manually uninstall  EKS Anywhere Cilium and replace it with a different CNI. In this case, it would be useful for the user to be able to verify the managed CNI is installed or not. 

**Adding new conditions**

Conditions provide a mechanism for high-level status reporting from the controller and a collection of conditions can be used represent a more specific assessment of cluster readiness. We should use multiple conditions to assessment  the user a the status of the individual checkpoints defined above and then use the conditions to determine overall cluster readiness. This gives the user insight into the checkpoint states of cluster readiness that we’ve defined, and allows other components to avoid duplicating the logic of each condition in the controller. Keeping the information that we need to collect that we’ve defined above in mind, we propose adding the following condition types:


* `ControlPlaneInitialized` - reports the first control plane has been initialized and the cluster’s API server is contactable via the available kubeconfig generated in the management cluster as a secret. Once this condition is marked true, it’s value never changes.
* `ControlPlaneReady` -  reports that the condition of the current state of the specified control plane machines vs the desired state in the Cluster spec. e.g. replicasReady vs replicas,
* `DefaultCNIConfigured` - reports that the CNI has been successfully configured. If an EKS Anywhere managed CNI solution is not configured and the `Cluster` spec is configured to skip EKS Anywhere's default CNI upgrades, this condition will be marked as “False” with the reason `SkipUpgradesForDefaultCNIConfigured`.
* `WorkersReady` - reports that the condition of the current state of the specified worker machines vs the desired state in the Cluster spec spec e.g. replicasReady vs replicas
* `Ready` - reports a summary of other conditions, indicating an overall operational state of the EKS Anywhere cluster. It will be marked “True” once the current state of the cluster has fully reached the desired state specified in the `Cluster` spec.


**Cleanup unnecessary fields**

Lastly, we can safely deprecate the `EksdReleaseRef` field from the `Cluster` status object and remove any reference to it in the docs.  The user is able to check the `BundlesRef` in the `Cluster` spec, and subsequently the bundle to see which release of Eksd is being used.

## Cluster Status Examples

In these examples, the `Cluster` spec submitted is configured with 1 Control Plane and 1 Worker Node replica unless otherwise specified.

**Cluster spec invalid**

```
  status:
    failureMessage: "Invalid cluster CloudStackDatacenterConfig: bad datacenter config"
    failureReason: "BadDatacenterConfig"
    childrenReconciledGeneration: 5
    reconciledGeneration: 2
    observedGeneration: 3
    conditions:
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      message: "First control plane not ready yet"
      reason: "WaitingForControlPlaneInitialized"
      status: "False"
      Type: "ControlPlaneInitialized"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      message: "Managed CNI not configured yet"
      reason: "WaitingForDefaultCNIConfigured"
      status: "False"
      Type: "DefaultCNIConfigured"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      message: "Scaling up control plane to 1 replicas (actual 0)"
      reason: "ScalingUp"
      status: "False"
      Type: "ControlPlaneReady"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      message: "Workers expected not ready yet, 1 replicas (actual 0)"
      reason: "ScalingUp"
      status: "False"
      Type: "WorkersReady"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      message: "Scaling up control plane to 1 replicas (actual 0)"
      reason: "ScalingUp"
      status: "False"
      Type: "Ready"
```

**Control plane not initialized**

```
  status:
    childrenReconciledGeneration: 5
    reconciledGeneration: 2
    observedGeneration: 3
    conditions:
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      message: "First control plane not ready yet"
      reason: "WaitingForControlPlaneInitialized"
      status: "False"
      Type: "ControlPlaneInitialized"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      message: "Managed CNI not configured yet"
      reason: "WaitingForDefaultCNIConfigured"
      status: "False"
      Type: "DefaultCNIConfigured"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      message: "Scaling up control plane to 1 replicas (actual 0)"
      reason: "ScalingUp"
      status: "False"
      Type: "ControlPlaneReady"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      message: "Workers expected not ready yet, 1 replicas (actual 0)"
      reason: "ScalingUp"
      status: "False"
      Type: "WorkersReady"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      message: "Scaling up control plane to 1 replicas (actual 0)"
      reason: "ScalingUp"
      status: "False"
      Type: "Ready"
```

**Control plane initialized and kubeconfig available**

```
  status:
    childrenReconciledGeneration: 5
    reconciledGeneration: 2
    observedGeneration: 3
    defaultCNI:
      name: "cilium"
      version: "v1.11.15-eksa.1"
      status: "applied"
    conditions:
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      status: "True"
      Type: "ControlPlaneInitialized"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      status: "True"
      Type: "DefaultCNIConfigured"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      message: "Scaling up control plane, 1 replicas (actual 0)"
      reason: "ScalingUp"
      status: "False"
      Type: "ControlPlaneReady"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      message: "Workers expected not ready yet, 1 replicas (actual 0)"
      reason: "ScalingUp"
      status: "False"
      Type: "WorkersReady"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      message: "Scaling up control plane and workers to 2 replicas (actual 1)"
      reason: "ScalingUp"
      status: "False"
      Type: "Ready"
```

**Cluster is ready**

```
  status:
    childrenReconciledGeneration: 5
    reconciledGeneration: 2
    observedGeneration: 3
    defaultCNI:
      name: "cilium"
      version: "v1.11.15-eksa.1"
      status: "applied"
    conditions:
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      status: "True"
      Type: "ControlPlaneInitialized"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      status: "True"
      Type: "DefaultCNIConfigured"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      status: "True"
      Type: "ControlPlaneReady"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      status: "True"
      Type: "WorkersReady"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      status: "True"
      Type: "Ready"
 ```     


**Rolling upgrade in progress (upgraded to 3 CP and 2 Worker Nodes)**

```
  status:
    childrenReconciledGeneration: 6
    reconciledGeneration: 3
    observedGeneration: 4
    defaultCNI:
      name: "cilium"
      version: "v1.11.15-eksa.1"
      status: "applied"
    conditions:
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      status: "True"
      Type: "ControlPlaneInitialized"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      status: "True"
      Type: "DefaultCNIConfigured"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      message: "Scaling up control plane to 3 replicas (actual 2)"
      reason: "ScalingUp"
      status: "False"
      Type: "ControlPlaneReady"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      message: "Workers expected not ready yet, 2 replicas (actual 1)"
      reason: "ScalingUp"
      status: "False"
      Type: "WorkersReady"
    - lastTransitionTime: "2023-06-05T21:58:57Z"
      message: "Scaling up control plane to 3 replicas (actual 2)"
      reason: "ScalingUp"
      status: "False"
      Type: "Ready"
```



## Alternative Solution

EKS Anywhere could also provide a CLI commands to poll cluster directly and return a summary of the cluster state as output.

Pros:
* No new fields need to be added to the Cluster status resource

Cons:
* Inefficient as the logic to assess the state of the cluster would execute on every run of the command
* Kubernetes [api conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status) favors adding to the Cluster status to represent the current state  of the cluster


The preferred way would be to represent the needed information in the `Cluster` status, which would avoid the need for users to use more complex logic as there would be a direct way to fetch the needed information on the state of the cluster.

## Testing

Testing if there is a failure related to upstream CAPI components that trickles down to `Cluster` status is not advised. We do not have any way of covering that directly.

We recommend testing using robust unit tests on the controller. We could also have unit tests on lower level entities depending on the implementation.

Additionally, we should update the API end-to-end tests infrastructure to use the `Cluster` status to determine the progress of the lifecycle operation. If the status is not populated correctly, then the tests won’t be able to detect when the cluster is ready and will time out.
