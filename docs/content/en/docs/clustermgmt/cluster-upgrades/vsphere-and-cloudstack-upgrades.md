---
title: "Upgrade vSphere, CloudStack, Nutanix, or Snow cluster"
linkTitle: "Upgrade vSphere, CloudStack, Nutanix, or Snow cluster"
weight: 20
aliases:
    /docs/tasks/cluster/cluster-upgrades/vsphere-and-cloudstack-upgrades/
date: 2017-01-05
description: >
  How to perform a cluster upgrade for vSphere, CloudStack, Nutanix, or Snow cluster
---
EKS Anywhere provides the command `upgrade`, which allows you to `upgrade` various aspects of your EKS Anywhere cluster.
When you run `eksctl anywhere upgrade cluster -f ./cluster.yaml`, EKS Anywhere runs a set of preflight checks to ensure your cluster is ready to be upgraded.
EKS Anywhere then performs the upgrade, modifying your cluster to match the updated specification. 
The upgrade command also upgrades core components of EKS Anywhere and lets the user enjoy the latest features, bug fixes and security patches.

**Upgrades should never be run from ephemeral nodes (short-lived systems that spin up and down on a regular basis). It is highly recommended to run the `upgrade` command with the `--no-timeouts` option when the command is executed through automation. This prevents the CLI from timing out and enables cluster operators to fix issues preventing the upgrade from completing while the process is running. If an `eksctl anywhere` version is older than `v0.18.0` and upgrade fails, _you must not delete the KinD bootstrap cluster Docker container_. During an upgrade, the bootstrap cluster contains critical EKS Anywhere components. If it is deleted after a failed upgrade, they cannot be recovered.**

{{% alert title="Important" color="warning" %}}

In `eksctl anywhere` version `v0.13.0`, we introduced the full lifecycle controller to fully manage new workload clusters.
In this version, the controller rolls out new nodes in the workload cluster whenever the user upgrades the management cluster and its management components.

In `eksctl anywhere` version `v0.16.0`, we changed this behavior to allow users to be explicit when deciding which clusters to upgrade.
Therefore, workload clusters are no longer affected by management cluster upgrades.
Due to this change, each cluster must be individually upgraded to enjoy the latest features on all clusters.
If you have a management cluster running EKS Anywhere version v0.15, you can successfully upgrade to EKS Anywhere version v0.16 and observe no changes to any of its workload clusters.

When triggering a workload cluster upgrade after upgrading the management cluster, please keep in mind that it will not only apply your changes in the workload cluster spec, but also any new improvements included in the new EKS Anywhere controller that was upgraded on the management cluster.
The changes in the EKS Anywhere controller can trigger a machine rollout on the workload cluster during upgrade, even if the changes to the workload cluster spec didn't require one (for example, scaling down a worker node group).

Starting with EKS Anywhere v0.18, the `image`/`template` should include the Kubernetes version (either `Cluster.Spec.KubernetesVersion` or `Cluster.Spec.WorkerNodeGroupConfiguration[].KubernetesVersion` in the case of modular upgrades). For example, if the Kubernetes version is 1.24, the `image`/`template` must include 1.24, 1_24, 1-24 or 124.
{{% /alert %}}

### Prepare DHCP IP addresses pool

Please make sure to have sufficient available IP addresses in your DHCP pool to cover the new machines. The number of necessary IPs can be calculated from the machine counts and [maxSurge config]({{< relref "./baremetal-upgrades.md/#upgraderolloutstrategyrollingupdatemaxsurge" >}}). For create operation, each machine needs 1 IP. For upgrade operation, control plane and workers need just 1 extra IP (total, not per node) due to rolling upgrade strategy. Each external etcd machine needs 1 extra IP address (ex: 3 etcd nodes would require 3 more IP addresses) because EKS Anywhere needs to create all the new etcd machines before removing any old ones. You will also need additional IPs to be equal to the number used for maxSurge. After calculating the required IPs, please make sure your environment has enough available IPs before performing the upgrade operation.

* Example 1, to create a cluster with 3 control plane node, 2 worker nodes and 3 stacked etcd, you will need at least 5 (3+2+0, as stacked etcd is deployed as part of the control plane nodes) available IPs. To upgrade the same cluster with default maxSurge (0), you will need 1 (1+0+0) additional available IPs.
* Example 2, to create a cluster with 1 control plane node, 2 worker nodes and 3 unstacked (external) etcd nodes, you will need at least 6 (1+2+3) available IPs. To upgrade the same cluster with default maxSurge (0), you will need at least 4 (1+3+0) additional available IPs.
* Example 3, to upgrade a cluster with 1 control plane node, 2 worker nodes and 3 unstacked (external) etcd nodes, with maxSurge set to 2, you will need at least 6 (1+3+2) additional available IPs.

### EKS Anywhere Version Upgrades

We encourage that you stay up to date with the latest EKS Anywhere version, as new features, bug fixes, or security patches might be added in each release. You can find the content, such as supported OS versions and changelog, of each EKS Anywhere version on the [What’s New]({{< relref "../../whatsnew/changelog/" >}}) page.

Download the [latest or target EKS Anywhere release](https://github.com/aws/eks-anywhere/releases/) and run `eksctl anywhere upgrade cluster` command to upgrade a cluster to a specific EKS Anywhere version.

Workload clusters can also be upgraded via the API (`kubectl apply` or GitOps) by changing `bundlesRef` for EKS Anywhere version 0.16.5 or below. Starting from v0.17.0, `bundlesRef` can be set to null and `eksaVersion` can be used instead.

`bundlesRef` is a reference to a bundles resource (collection of dependencies needed by an EKS Anywhere cluster) on the cluster whereas `eksaVersion` must be a valid SemVer value that maps to an EKSARelease resource on the cluster via the EKSARelease name. Both of these fields are automatically updated by EKS Anywhere and only need to be manually changed when upgrading via the API. The supported values for `eksaVersion` can be obtained by running `kubectl get eksareleases -n eksa-system`. For an EKSARelease with the name eksa-vX-X-X-prereleaseMetadata-plus-buildMetadata, `eksaVersion` can be set to vX-X-X-preleaseMetadata+buildMetadata.
The workload's version may not exceed the management cluster. Any upgrades to `eksaVersion` must also be sequential relative to minor version. However, you can choose to skip patch versions.

**Skipping Amazon EKS Anywhere minor versions during cluster upgrade (such as going from v0.14 directly to v0.16) is NOT allowed.** EKS Anywhere team performs regular upgrade reliability testing for sequential version upgrade (e.g. going from version 0.14 to 0.15, then from version 0.15 to 0.16), but we do not perform testing on non-sequential upgrade path (e.g. going from version 0.14 directly to 0.16). You should not skip minor versions during cluster upgrade. However, you can choose to skip patch versions.

To upgrade EKS Anywhere version for an airgapped cluster, you need to [download new artifacts and images]({{< relref "./airgapped-upgrades" >}}).

{{% alert title="Important" color="warning" %}}

We provide a maximum skew version support of one EKS Anywhere minor version for management and workload clusters.
This means that we support the management cluster being one EKS Anywhere minor version newer than the workload clusters (such as v0.15 for workload clusters if the management cluster is at v0.16).
In the event that you want to upgrade your management cluster to a version that does not satisfy this condition, we recommend upgrading the workload cluster's EKS Anywhere version first, followed by upgrading to your desired EKS Anywhere version for the management cluster.

{{% /alert %}}

### Kubernetes Minor Version Upgrades

Kubernetes has minor releases [three times per year](https://kubernetes.io/releases/release/) and EKS Distro follows a similar cadence.
EKS Anywhere will add support for new EKS Distro releases as they are released, and you are advised to upgrade your cluster when possible.

Cluster upgrades are not handled automatically and require administrator action to modify the cluster specification and perform an upgrade. This may also require [building a node image or OVA]({{< relref "../../osmgmt/artifacts/" >}}) for the new Kubernetes version being upgraded to.
You are advised to upgrade your clusters in development environments first and verify your workloads and controllers are compatible with the new version.

Cluster upgrades are performed in place using a rolling process (similar to Kubernetes Deployments).
Upgrades can only happen one minor version at a time (e.g. `1.24` -> `1.25`).
Control plane components will be upgraded before worker nodes.

A new VM is created with the new version and then an old VM is removed.
This happens one at a time until all the control plane components have been upgraded.

Worker node groups can optionally be upgraded separately from the control plane by setting `workerNodeGroupConfiguration.kubernetesVersion`. There can only be a skew of two minor versions between the control plane and each worker node. Removing `workerNodeGroupConfiguration.kubernetesVersion` will trigger an upgrade to that node group to upgrade to the root level `kubernetesVersion`.

### Core component upgrades

EKS Anywhere `upgrade` also supports upgrading the following core components:

* Core CAPI
* CAPI providers
* Cilium CNI plugin
* Cert-manager
* Etcdadm CAPI provider
* EKS Anywhere controllers and CRDs
* GitOps controllers (Flux) - this is an optional component, will be upgraded only if specified

The latest versions of these core EKS Anywhere components are embedded into a bundles manifest that the CLI uses to fetch the latest versions 
and image builds needed for each component upgrade. 
The command detects both component version changes and new builds of the same versioned component.
If there is a new Kubernetes version that is going to get rolled out, the core components get upgraded before the Kubernetes
version. 
Irrespective of a Kubernetes version change, the upgrade command will always upgrade the internal EKS
Anywhere components mentioned above to their latest available versions. All upgrade changes are backwards compatible.

Specifically for Snow provider, a new Admin instance is needed when upgrading to the new versions of EKS Anywhere. See [Upgrade EKS Anywhere AMIs in Snowball Edge devices](https://docs.aws.amazon.com/snowball/latest/developer-guide/CrUD-clusters.html) to upgrade and use a new Admin instance in Snow devices. After that, ugrades of other components can be done as described in this document.

### Check upgrade components
Before you perform an upgrade, check the current and new versions of components that are ready to upgrade by typing:

**Management Cluster**

```bash
eksctl anywhere upgrade plan cluster -f mgmt-cluster.yaml
```

**Workload Cluster**

```bash
eksctl anywhere upgrade plan cluster -f workload-cluster.yaml --kubeconfig mgmt/mgmt-eks-a-cluster.kubeconfig
```

The output should appear similar to the following:

```
Worker node group name not specified. Defaulting name to md-0.
Warning: The recommended number of control plane nodes is 3 or 5
Worker node group name not specified. Defaulting name to md-0.
Checking new release availability...
NAME                     CURRENT VERSION                 NEXT VERSION
EKS-A                    v0.0.0-dev+build.1000+9886ba8   v0.0.0-dev+build.1105+46598cb
cluster-api              v1.0.2+e8c48f5                  v1.0.2+1274316
kubeadm                  v1.0.2+92c6d7e                  v1.0.2+aa1a03a
vsphere                  v1.0.1+efb002c                  v1.0.1+ef26ac1
kubadm                   v1.0.2+f002eae                  v1.0.2+f443dcf
etcdadm-bootstrap        v1.0.2-rc3+54dcc82              v1.0.0-rc3+df07114
etcdadm-controller       v1.0.2-rc3+a817792              v1.0.0-rc3+a310516
```
To the format output in json, add `-o json` to the end of the command line.

### Performing a cluster upgrade

To perform a cluster upgrade you can modify your cluster specification `kubernetesVersion` field to the desired version.

As an example, to upgrade a cluster with version 1.26 to 1.27 you would change your spec

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
      kind: VSphereMachineConfig
      name: dev
      ...
  kubernetesVersion: "1.27"
      ...
```

>**_NOTE:_** If you have a custom machine image for your nodes you may also need to update your `vsphereMachineConfig` with a new `template`. Refer to [vSphere Artifacts]({{< relref "../../osmgmt/artifacts/#vsphere-artifacts" >}}) to build a new OVA template.

and then you will run the [upgrade cluster command]({{< relref "vsphere-and-cloudstack-upgrades/#upgrade-cluster-command" >}}).

#### Upgrade cluster command

* **kubectl CLI**: The cluster lifecycle feature lets you use `kubectl` to talk to the Kubernetes API to upgrade an EKS Anywhere cluster. For example, to use `kubectl` to upgrade a management or workload cluster, you can run:
   ```bash
   # Upgrade a management cluster with cluster name "mgmt"
   kubectl apply -f mgmt-cluster.yaml --kubeconfig mgmt/mgmt-eks-a-cluster.kubeconfig

  # Upgrade a workload cluster with cluster name "eksa-w01"
   kubectl apply -f eksa-w01-cluster.yaml --kubeconfig mgmt/mgmt-eks-a-cluster.kubeconfig
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

{{% alert title="Important" color="warning" %}}

**For kubectl, GitOps and Terraform**

If you want to update the [registry mirror]({{< relref "../../getting-started/optional/registrymirror" >}}) credential with `kubectl`, GitOps or Terraform, you need to update the `registry-credentials` secret in the `eksa-system` namespace of your management cluster. For example with `kubectl`, you can run:

```bash
kubectl edit secret -n eksa-system registry-credentials --kubeconfig mgmt/mgmt-eks-a-cluster.kubeconfig
```

Replace username and password fields with the base64-encoded values of your new username and password. You can encode the values using the `echo` command, for example:

```bash
echo -n 'newusername' | base64
echo -n 'newpassword' | base64
```

{{% /alert %}}

* **eksctl CLI**: To upgrade an EKS Anywhere cluster with `eksctl`, run:

  ```bash
  # Upgrade a management cluster with cluster name "mgmt"
   eksctl anywhere upgrade cluster -f mgmt-cluster.yaml

  # Upgrade a workload cluster with cluster name "eksa-w01"
   eksctl anywhere upgrade cluster -f eksa-w01-cluster.yaml --kubeconfig mgmt/mgmt-eks-a-cluster.kubeconfig
  ```
  As noted earlier, adding the `--kubeconfig` option tells `eksctl` to use the management cluster identified by that kubeconfig file to upgrade a different workload cluster.

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

>**_NOTE (vSphere only):_** If you are upgrading a vSphere cluster created using EKS Anywhere version prior to `v0.16.0` that has the vSphere CSI Driver installed in it, please refer to the additional steps listed [here]({{< relref "../storage/vsphere-storage#csi-driver-cleanup-for-upgrades" >}}) before attempting an upgrade.

### Upgradeable Cluster Attributes
EKS Anywhere `upgrade` supports upgrading more than just the `kubernetesVersion`, 
allowing you to upgrade a number of fields simultaneously with the same procedure.

#### Upgradeable Attributes

`Cluster`:
- `kubernetesVersion`
- `controlPlaneConfig.count`
- `controlPlaneConfigurations.machineGroupRef.name`
- `controlPlaneConfigurations.upgradeRolloutStrategy.rollingUpdate.maxSurge`
- `workerNodeGroupConfigurations.count`
- `workerNodeGroupConfigurations.machineGroupRef.name`
- `workerNodeGroupConfigurations.kubernetesVersion`
- `workerNodeGroupConfigurations.upgradeRolloutStrategy.rollingUpdate.maxSurge`
- `workerNodeGroupConfigurations.upgradeRolloutStrategy.rollingUpdate.maxUnavailable`
- `externalEtcdConfiguration.machineGroupRef.name`
- `identityProviderRefs` (Only for `kind:OIDCConfig`, `kind:AWSIamConfig` is immutable)
- `gitOpsRef` (Once set, you can't change or delete the field's content later)
- `registryMirrorConfiguration` (for non-authenticated registry mirror)
  - `endpoint`
  - `port` 
  - `caCertContent`
  - `insecureSkipVerify` 

`VSphereMachineConfig`:
- `datastore`
- `diskGiB`
- `folder`
- `memoryMiB`
- `numCPUs`
- `resourcePool`
- `template`
- `users`

`NutanixMachineConfig`:
- `vcpusPerSocket`
- `vcpuSockets`
- `memorySize`
- `image`
- `cluster`
- `subnet`
- `systemDiskSize`

`SnowMachineConfig`:
- `amiID`
- `instanceType`
- `physicalNetworkConnector`
- `sshKeyName`
- `devices`
- `containersVolume`
- `osFamily`
- `network`

`CloudStackDatacenterConfig`:
- `availabilityZones` (Can add and remove availability zones provided at least 1 previously configured zone is still present)

`CloudStackMachineConfig`:
- `template`
- `computeOffering`
- `diskOffering`
- `userCustomDetails`
- `symlinks`
- `users`

`OIDCConfig`:
- `clientID`
- `groupsClaim`
- `groupsPrefix`
- `issuerUrl`
- `requiredClaims.claim`
- `requiredClaims.value`
- `usernameClaim`
- `usernamePrefix`

`AWSIamConfig`:
- `mapRoles`
- `mapUsers`

EKS Anywhere `upgrade` also supports adding more worker node groups post-creation.
To add more worker node groups, modify your cluster config file to define the additional group(s).
Example:
```
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: dev
spec:
  controlPlaneConfiguration:
     ...
  workerNodeGroupConfigurations:
  - count: 2
    machineGroupRef:
      kind: VSphereMachineConfig
      name: my-cluster-machines
    name: md-0
  - count: 2
    machineGroupRef:
      kind: VSphereMachineConfig
      name: my-cluster-machines
    name: md-1
      ...
```

Worker node groups can use the same machineGroupRef as previous groups, or you can define a new machine configuration for your new group.

#### Advanced configuration for rolling upgrade

EKS Anywhere allows an optional configuration to customize the behavior of upgrades. 

It allows the specification of 
Two parameters that control the desired behavior of rolling upgrades: 
* maxSurge - The maximum number of machines that can be scheduled above the desired number of machines. When not specified, the current CAPI default of 1 is used.
* maxUnavailable - The maximum number of machines that can be unavailable during the upgrade. When not specified, the current CAPI default of 0 is used.

Example configuration:

```bash
upgradeRolloutStrategy:
  type: RollingUpdate
  rollingUpdate:
    maxSurge: 1
    maxUnavailable: 0    # only configurable for worker nodes
```

'upgradeRolloutStrategy' configuration can be specified separately for control plane and for each worker node group. This template contains an example for control plane under the 'controlPlaneConfiguration' section and for worker node group under 'workerNodeGroupConfigurations':

```bash
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: my-cluster-name
spec:
  controlPlaneConfiguration:
    count: 1
    endpoint:
      host: "xx.xx.xx.xx"
    machineGroupRef:
      kind: VSphereMachineConfig
      name: my-cluster-name-cp
    upgradeRolloutStrategy:
      type: RollingUpdate
      rollingUpdate:
        maxSurge: 1 
  workerNodeGroupConfigurations:
  - count: 2
    machineGroupRef:
      kind: VSphereMachineConfig
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
Type of rollout strategy. Currently only `RollingUpdate` is supported.

#### upgradeRolloutStrategy.rollingUpdate
Configuration parameters for customizing rolling upgrade behavior.

#### upgradeRolloutStrategy.rollingUpdate.maxSurge
Default: 1

This can not be 0 if maxUnavailable is 0.

The maximum number of machines that can be scheduled above the desired number of machines. 

Example: When this is set to n, the new worker node group can be scaled up immediately by n when the rolling upgrade starts. Total number of machines in the cluster (old + new) never exceeds (desired number of machines + n). Once scale down happens and old machines are brought down, the new worker node group can be scaled up further ensuring that the total number of machines running at any time does not exceed the desired number of machines + n.

#### upgradeRolloutStrategy.rollingUpdate.maxUnavailable
Default: 0

This can not be 0 if MaxSurge is 0.

The maximum number of machines that can be unavailable during the upgrade.

Example: When this is set to n, the old worker node group can be scaled down by n machines immediately when the rolling upgrade starts. Once new machines are ready, old worker node group can be scaled down further, followed by scaling up the new worker node group, ensuring that the total number of machines unavailable at all times during the upgrade never falls below n.

### Resume upgrade after failure

EKS Anywhere supports re-running the `upgrade` command post-failure as an experimental feature.
If the `upgrade` command fails, the user can manually fix the issue (when applicable) and simply rerun the same command.  At this point, the CLI will skip the completed tasks, restore the state of the operation, and resume the upgrade process.
The completed tasks are stored in the `generated` folder as a file named `<clusterName>-checkpoint.yaml`.

This feature is experimental. To enable this feature, export the following environment variable:<br/>
`export CHECKPOINT_ENABLED=true`

### Troubleshooting

Attempting to upgrade a cluster with more than 1 minor release will result in receiving the following error.

```
✅ validate immutable fields
❌ validation failed    {"validation": "Upgrade preflight validations", "error": "validation failed with 1 errors: WARNING: version difference between upgrade version (1.21) and server version (1.19) do not meet the supported version increment of +1", "remediation": ""}
Error: failed to upgrade cluster: validations failed
```

For more errors you can see the [troubleshooting section]({{< relref "../../troubleshooting" >}}).
