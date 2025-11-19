---
title: "Upgrade Bare Metal cluster"
linkTitle: "Upgrade Bare Metal cluster"
weight: 20
aliases:
    /docs/tasks/cluster/cluster-upgrades/baremetal-upgrades/
date: 2017-01-05
description: >
  How to perform Bare Metal cluster upgrades
---

{{% alert title="Note" color="warning" %}}

**Upgrade overview information was moved to a dedicated [Upgrade Overview page.]({{< relref "./upgrade-overview" >}})**

{{% /alert %}}

### Considerations

- Only EKS Anywhere and Kubernetes version upgrades are supported for Bare Metal clusters. You cannot update other cluster configuration.
- **Upgrades should never be run from ephemeral nodes (short-lived systems that spin up and down on a regular basis). If the EKS Anywhere version is lower than `v0.18.0` and upgrade fails, _you must not delete the KinD bootstrap cluster Docker container_. During an upgrade, the bootstrap cluster contains critical EKS Anywhere components. If it is deleted after a failed upgrade, they cannot be recovered.**
- It is highly recommended to run the `eksctl anywhere upgrade cluster` command with the `--no-timeouts` option when the command is executed through automation. This prevents the CLI from timing out and enables cluster operators to fix issues preventing the upgrade from completing while the process is running. 
- In EKS Anywhere version `v0.15.0`, we introduced the EKS Anywhere cluster lifecycle controller that runs on management clusters and manages workload clusters. The EKS Anywhere lifecycle controller enables you to use Kubernetes API-compatible clients such as `kubectl`, GitOps, or Terraform for managing workload clusters. In this EKS Anywhere version, the EKS Anywhere cluster lifecycle controller rolls out new nodes in workload clusters when management clusters are upgraded. In EKS Anywhere version `v0.16.0`, this behavior was changed such that management clusters can be upgraded separately from workload clusters.
- When running workload cluster upgrades after upgrading a management cluster, a machine rollout may be triggered on workload clusters during the workload cluster upgrade, even if the changes to the workload cluster spec didn't require one (for example scaling down a worker node group).
- Starting with EKS Anywhere `v0.18.0`, the `osImageURL` must include the Kubernetes minor version (`Cluster.Spec.KubernetesVersion` or `Cluster.Spec.WorkerNodeGroupConfiguration[].KubernetesVersion` in the cluster spec). For example, if the Kubernetes version is 1.33, the `osImageURL` must include 1.33, 1_33, 1-33 or 133. If you are upgrading Kubernetes versions, you must have a new OS image with your target Kubernetes version components.
- If you are running EKS Anywhere in an airgapped environment, you must download the new artifacts and images prior to initiating the upgrade. Reference the [Airgapped Upgrades page]({{< relref "./airgapped-upgrades" >}}) page for more information.

### Upgrade Version Skew

{{% content "version-skew.md" %}}

### Prerequisites

EKS Anywhere upgrades on Bare Metal require at least one spare hardware server for control plane upgrade and one for each worker node group upgrade. During upgrade, the spare hardware server is provisioned with the new version and then an old server is deprovisioned. The deprovisioned server is then reprovisioned with
the new version while another old server is deprovisioned. This happens one at a time until all the control plane components have been upgraded, followed by
worker node upgrades.

### Check upgrade components
Before you perform an upgrade, check the current and new versions of components that are ready to upgrade by typing:

```bash
eksctl anywhere upgrade plan cluster -f cluster.yaml
```

The output should appear similar to the following:

```
Checking new release availability...
NAME                 CURRENT VERSION                NEXT VERSION
EKS-A Management     v0.19.0-dev+build.20+a0037f0   v0.19.0-dev+build.26+3bc5008
cert-manager         v1.13.2+129095a                v1.13.2+bb56494
cluster-api          v1.6.1+5efe087                 v1.6.1+9cf3436
kubeadm              v1.6.1+8ceb315                 v1.6.1+82f1c0a
tinkerbell           v0.4.0+cdde180                 v0.4.0+e848206
kubeadm              v1.6.1+6420e1c                 v1.6.1+2f0b35f
etcdadm-bootstrap    v1.0.10+7094b99                v1.0.10+a3f0355
etcdadm-controller   v1.0.17+0259550                v1.0.17+ba86997
```
To format the output in json, add `-o json` to the end of the command line.

### Check hardware availability

Next, you must ensure you have enough available hardware for the rolling upgrade operation to function. This type of upgrade requires you to have one spare hardware server for control plane upgrade and one for each worker node group upgrade. Check [prerequisites]({{< relref "baremetal-upgrades/#prerequisites" >}}) for more information.
Available hardware could have been fed to the cluster as extra hardware during a prior create command, or could be fed to the cluster during the upgrade process by providing the hardware CSV file to the [upgrade cluster command]({{< relref "baremetal-upgrades/#upgrade-cluster-command" >}}).

To check if you have enough available hardware for rolling upgrade, you can use the `kubectl` command below to check if there are hardware objects with the selector labels corresponding to the controlplane/worker node group and without the `ownerName` label. 

```bash
kubectl get hardware -n eksa-system --show-labels
```

For example, if you want to perform upgrade on a cluster with one worker node group with selector label `type=worker-group-1`, then you must have an additional hardware object in your cluster with the label `type=controlplane` (for control plane upgrade) and one with `type=worker-group-1` (for worker node group upgrade) that doesn't have the `ownerName` label. 

In the command shown below, `eksa-worker2` matches the selector label and it doesn't have the `ownerName` label. Thus, it can be used to perform rolling upgrade of `worker-group-1`. Similarly, `eksa-controlplane-spare` will be used for rolling upgrade of control plane.

```bash
kubectl get hardware -n eksa-system --show-labels 
NAME                STATE       LABELS
eksa-controlplane               type=controlplane,v1alpha1.tinkerbell.org/ownerName=abhnvp-control-plane-template-1656427179688-9rm5f,v1alpha1.tinkerbell.org/ownerNamespace=eksa-system
eksa-controlplane-spare         type=controlplane
eksa-worker1                    type=worker-group-1,v1alpha1.tinkerbell.org/ownerName=abhnvp-md-0-1656427179689-9fqnx,v1alpha1.tinkerbell.org/ownerNamespace=eksa-system
eksa-worker2                    type=worker-group-1
```

If you don't have any available hardware that match this requirement in the cluster, you can [setup a new hardware CSV]({{< relref "../../getting-started/baremetal/bare-preparation/#prepare-hardware-inventory" >}}). You can feed this hardware inventory file during the [upgrade cluster command]({{< relref "baremetal-upgrades/#upgrade-cluster-command" >}}).

#### Skip BMC connectivity checks for faulty machines

EKS Anywhere validates that all BMC machines are contactable before performing cluster upgrades. If you have faulty BMC machines with connectivity issues or hardware faults, you can skip validation for those specific machines so they don't block your cluster upgrade.

To skip BMC validation for a machine:

```
kubectl label machine.bmc.tinkerbell.org <machine-name> \
  -n eksa-system \
  bmc.tinkerbell.org/skip-contact-check=true
```

After you have repaired or replaced the faulty hardware and verified that BMC connectivity is restored, remove the skip label to re-enable normal BMC validation for future upgrades:

```
kubectl label machine.bmc.tinkerbell.org <machine-name> \
  -n eksa-system \
  bmc.tinkerbell.org/skip-contact-check-
```

>**_NOTE:_** The skip label should only be used temporarily during upgrades when a machine has known issues. Once the machine is healthy, remove the label to ensure proper BMC validation in subsequent operations.

### Performing a cluster upgrade

To perform a cluster upgrade you can modify your cluster specification `kubernetesVersion` field to the desired version.

As an example, to upgrade a cluster with version 1.32 to 1.33 you would change your spec as follows:

```
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: dev
spec:
  controlPlaneConfiguration:
    count: 1
    endpoint:
      host: "198.18.99.49"
    machineGroupRef:
      kind: TinkerbellMachineConfig
      name: dev
      ...
  kubernetesVersion: "1.33"
      ...
```

>**_NOTE:_** If you have a custom machine image for your nodes in your cluster config yaml or to upgrade a node or group of nodes to a new operating system version (ie. RHEL 8.7 to RHEL 8.8), you may also need to update your [`TinkerbellDatacenterConfig`]({{< relref "../../getting-started/baremetal/bare-spec/#tinkerbelldatacenterconfig-fields" >}}) or [`TinkerbellMachineConfig`]({{< relref "../../getting-started/baremetal/bare-spec/#tinkerbellmachineconfig-fields" >}}) with the new operating system image URL [`osImageURL`]({{< relref "../../getting-started/baremetal/bare-spec/#osimageurl-required" >}}). 

and then you will run the [upgrade cluster command]({{< relref "baremetal-upgrades/#upgrade-cluster-command" >}}).


#### Upgrade cluster command
* **kubectl CLI**: The cluster lifecycle feature lets you use kubectl to talk to the Kubernetes API to upgrade a workload cluster. To use kubectl, run:
   ```bash
   kubectl apply -f eksa-w01-cluster.yaml 
  --kubeconfig mgmt/mgmt-eks-a-cluster.kubeconfig
   ```
 
    To check the state of a cluster managed with the cluster lifecyle feature, use `kubectl` to show the cluster object with its status.
    
    The `status` field on the cluster object field holds information about the current state of the cluster.

    ```
    kubectl get clusters w01 -o yaml
    ```

    The cluster has been fully upgraded once the status of the `Ready` condition is marked `True`.
    See the [cluster status]({{< relref "../cluster-status" >}}) guide for more information.
  
* **GitOps**: See [Manage separate workload clusters with GitOps]({{< relref "../cluster-flux.md#manage-separate-workload-clusters-using-gitops" >}})

* **Terraform**: See [Manage separate workload clusters with Terraform]({{< relref "../cluster-terraform.md#manage-separate-workload-clusters-using-terraform" >}})

  >**NOTE**:For kubectl, GitOps and Terraform:
  > * The baremetal controller does not support scaling upgrades and Kubernetes version upgrades in the same request.
  > * While scaling a workload cluster if you need to add additional machines, run:
  >   ```
  >   eksctl anywhere generate hardware -z updated-hardware.csv > updated-hardware.yaml
  >   kubectl apply -f updated-hardware.yaml
  >   ```
  >
  > *  If you want to upgrade multiple workload clusters, make sure that the spare hardware that is available for new nodes to rollout has labels unique to the workload cluster you are trying to upgrade. For instance, for an EKSA cluster named `eksa-workload1`, the hardware that is assigned for this cluster should have labels that are only going to be used for this cluster like `type=eksa-workload1-cp` and `type=eksa-workload1-worker`. Another workload cluster named `eksa-workload2` can have labels like `type=eksa-workload2-cp` and `type=eksa-workload2-worker`. Please note that even though labels can be arbitrary, they need to be unique for each workload cluster. Not specifying unique cluster labels can cause cluster upgrades to behave in unexpected ways which may lead to unsuccessful upgrades and unstable clusters.

* **eksctl CLI**: To upgrade a workload cluster with eksctl, run:

  ```bash
  eksctl anywhere upgrade cluster -f cluster.yaml 
  # --hardware-csv <hardware.csv> \ # uncomment to add more hardware
  --kubeconfig mgmt/mgmt-eks-a-cluster.kubeconfig
  ```
  As noted earlier, adding the `--kubeconfig` option tells `eksctl` to use the management cluster identified by that kubeconfig file to create a different workload cluster.

  This will upgrade the cluster specification (if specified), upgrade the core components to the latest available versions and apply the changes using the provisioner controllers.

  #### Output
  
  Example output:

  ```
  ✅ control plane ready
  ✅ worker nodes ready
  ✅ nodes ready
  ✅ cluster CRDs ready
  ✅ cluster object present on workload cluster
  ✅ upgrade cluster kubernetes version increment
  ✅ validate immutable fields
  🎉 all cluster upgrade preflight validations passed
  Performing provider setup and validations
  Ensuring etcd CAPI providers exist on management cluster before upgrade
  Pausing GitOps cluster resources reconcile
  Upgrading core components
  Backing up management cluster's resources before upgrading
  Upgrading management cluster
  Updating Git Repo with new EKS-A cluster spec
  Forcing reconcile Git repo with latest commit
  Resuming GitOps cluster resources kustomization
  Writing cluster config file
  🎉 Cluster upgraded!
  Cleaning up backup resources
  ```
  
  Starting in EKS Anywhere v0.18.0, when upgrading management cluster the CLI depends on the EKS Anywhere Controller to perform the upgrade. In the event an issue occurs and the CLI times out, it may be possible to fix the issue and have the upgrade complete as the EKS Anywhere Controller will continually attempt to complete the upgrade.

  During the workload cluster upgrade process, EKS Anywhere pauses the cluster controller reconciliation by adding the paused annotation `anywhere.eks.amazonaws.com/paused: true` to the EKS Anywhere cluster, provider datacenterconfig and machineconfig resources, before the components upgrade. After upgrade completes, the annotations are removed so that the cluster controller resumes reconciling the cluster. If the CLI execution is interrupted or times out, the controller won't reconcile changes to the EKS-A objects until these annotations are removed. You can re-run the CLI to restart the upgrade process or remove the annotations manually with `kubectl`.
  
  Though not recommended, you can manually pause the EKS Anywhere cluster controller reconciliation to perform extended maintenance work or interact with Cluster API objects directly. To do it, you can add the paused annotation to the cluster resource:
  
  ```bash
  kubectl annotate clusters.anywhere.eks.amazonaws.com ${CLUSTER_NAME} -n ${CLUSTER_NAMESPACE} anywhere.eks.amazonaws.com/paused=true
  ```
  
  After finishing the task, make sure you resume the cluster reconciliation by removing the paused annotation, so that EKS Anywhere cluster controller can continue working as expected.
  
  ```bash
  kubectl annotate clusters.anywhere.eks.amazonaws.com ${CLUSTER_NAME} -n ${CLUSTER_NAMESPACE} anywhere.eks.amazonaws.com/paused-
  ```

### Upgradeable cluster attributes

`Cluster`:
- `kubernetesVersion`
- `controlPlaneConfiguration.count`
- `controlPlaneConfiguration.upgradeRolloutStrategy.rollingUpdate.maxSurge`
- `workerNodeGroupConfigurations.count`
- `workerNodeGroupConfigurations.kubernetesVersion` (in case of modular upgrade)
- `workerNodeGroupConfigurations.upgradeRolloutStrategy.rollingUpdate.maxSurge`
- `workerNodeGroupConfigurations.upgradeRolloutStrategy.rollingUpdate.maxUnavailable`

`TinkerbellDatacenterConfig`:
- `osImageURL`

### Advanced configuration for upgrade rollout strategy

EKS Anywhere allows an optional configuration to customize the behavior of upgrades. 

`upgradeRolloutStrategy` can be configured separately for control plane and for each worker node group.
This template contains an example for control plane under the `controlPlaneConfiguration` section and for worker node group under `workerNodeGroupConfigurations`:

```bash
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
    count: 1
    endpoint:
      host: "10.61.248.209"
    machineGroupRef:
      kind: TinkerbellMachineConfig
      name: my-cluster-name-cp
    upgradeRolloutStrategy:
      type: RollingUpdate
      rollingUpdate:
        maxSurge: 1
  datacenterRef:
    kind: TinkerbellDatacenterConfig
    name: my-cluster-name
  kubernetesVersion: "1.33"
  managementCluster:
    name: my-cluster-name 
  workerNodeGroupConfigurations:
  - count: 2
    machineGroupRef:
      kind: TinkerbellMachineConfig
      name: my-cluster-name 
    name: md-0
    upgradeRolloutStrategy:
      type: RollingUpdate
      rollingUpdate:
        maxSurge: 1
        maxUnavailable: 0

---
...
```

#### upgradeRolloutStrategy
Configuration parameters for upgrade strategy.

#### upgradeRolloutStrategy.type
Default: `RollingUpdate`

Type of rollout strategy. Supported values: `RollingUpdate`,`InPlace`.

>**_NOTE:_** The upgrade rollout strategy type must be the same for all control plane and worker nodes.

#### upgradeRolloutStrategy.rollingUpdate
Configuration parameters for customizing rolling upgrade behavior.

>**_NOTE:_** The rolling update parameters can only be configured if `upgradeRolloutStrategy.type` is `RollingUpdate`.

#### upgradeRolloutStrategy.rollingUpdate.maxSurge
Default: 1

This can not be 0 if maxUnavailable is 0.

The maximum number of machines that can be scheduled above the desired number of machines.

Example: When this is set to n, the new worker node group can be scaled up immediately by n when the rolling upgrade starts. Total number of machines in the cluster (old + new) never exceeds (desired number of machines + n). Once scale down happens and old machines are brought down, the new worker node group can be scaled up further ensuring that the total number of machines running at any time does not exceed the desired number of machines + n.

#### upgradeRolloutStrategy.rollingUpdate.maxUnavailable
Default: 0

This can not be 0 if MaxSurge is 0.

The maximum number of machines that can be unavailable during the upgrade.

This can only be configured for worker nodes.

Example: When this is set to n, the old worker node group can be scaled down by n machines immediately when the rolling upgrade starts. Once new machines are ready, old worker node group can be scaled down further, followed by scaling up the new worker node group, ensuring that the total number of machines unavailable at all times during the upgrade never falls below n.

### Rolling Upgrades

The `RollingUpdate` rollout strategy type allows the specification of two parameters that control the desired behavior of rolling upgrades: 
* `maxSurge` - The maximum number of machines that can be scheduled above the desired number of machines. When not specified, the current CAPI default of 1 is used.
* `maxUnavailable` - The maximum number of machines that can be unavailable during the upgrade. When not specified, the current CAPI default of 0 is used.

Example configuration:

```bash
upgradeRolloutStrategy:
  type: RollingUpdate
  rollingUpdate:
    maxSurge: 1
    maxUnavailable: 0    # only configurable for worker nodes
```

### Rolling upgrades with no additional hardware

When maxSurge is set to 0 and maxUnavailable is set to 1, it allows for a rolling upgrade without need for additional hardware. Use this configuration if your workloads can tolerate node unavailability.

>**_NOTE:_** This could ONLY be used if unavailability of a maximum of 1 node is acceptable. For single node clusters, an additional temporary machine is a must. Alternatively, you may recreate the single node cluster for upgrading and handle data recovery manually.

With this kind of configuration, the rolling upgrade will proceed node by node, deprovision and delete a node fully before re-provisioning it with upgraded version, and re-join it to the cluster. This means that any point during the course of the rolling upgrade, there could be one unavailable node.


### In-Place Upgrades

As of EKS Anywhere version `v0.19.0`, the `InPlace` rollout strategy type can be used to upgrade the EKS Anywhere and Kubernetes versions by upgrading the components on the same physical machines without requiring additional server capacity.
EKS Anywhere schedules a privileged pod that executes the upgrade logic as a sequence of init containers on each node to be upgraded.
This upgrade logic includes updating the containerd, cri-tools, kubeadm, kubectl and kubelet binaries along with core Kubernetes components and restarting those services.

Due to the nature of this upgrade, temporary downtime of workloads can be expected.
It is best practice to configure your clusters in a way that they are resilient to having one node down.

During in place upgrades, EKS Anywhere pauses machine health checks to ensure that new nodes are not rolled out while the node is temporarily down during the upgrade process.
Moreover, autoscaler configuration is not supported when using `InPlace` upgrade rollout strategy to further ensure that no new nodes are rolled out unexpectedly.

Example configuration:

```bash
upgradeRolloutStrategy:
  type: InPlace
```

### Troubleshooting

Attempting to upgrade a cluster with more than 1 minor release will result in receiving the following error.

```
✅ validate immutable fields
❌ validation failed    {"validation": "Upgrade preflight validations", "error": "validation failed with 1 errors: WARNING: version difference between upgrade version (1.21) and server version (1.19) do not meet the supported version increment of +1", "remediation": ""}
Error: failed to upgrade cluster: validations failed
```

For more errors you can see the [troubleshooting section]({{< relref "../../troubleshooting" >}}).
