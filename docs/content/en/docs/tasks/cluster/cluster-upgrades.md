---
title: "Upgrade cluster"
linkTitle: "Upgrade cluster"
weight: 20
date: 2017-01-05
description: >
  How to perform a cluster version upgrade
---
EKS Anywhere provides the command `upgrade`, which allows you to `upgrade` various aspects of your EKS Anywhere cluster.
When you run `eksctl anywhere upgrade cluster -f ./cluster.yaml`, EKS Anywhere runs a set of preflight checks to ensure your cluster is ready to be upgraded.
EKS Anywhere then performs the upgrade, modifying your cluster to match the updated specification. 
The upgrade command also upgrades core components of EKS Anywhere and lets the user enjoy the latest features, bug fixes and security patches.


### Minor Version Upgrades

Kubernetes has minor releases [three times per year](https://kubernetes.io/releases/release/) and EKS Distro follows a similar cadence.
EKS Anywhere will add support for new EKS Distro releases as they are released, and you are advised to upgrade your cluster when possible.

Cluster upgrades are not handled automatically and require administrator action to modify the cluster specification and perform an upgrade.
You are advised to upgrade your clusters in development environments first and verify your workloads and controllers are compatible with the new version.

Cluster upgrades are performed in place using a rolling process (similar to Kubernetes Deployments).
Upgrades can only happen one minor version at a time (e.g. `1.20` -> `1.21`).
Control plane components will be upgraded before worker nodes.

A new VM is created with the new version and then an old VM is removed.
This happens one at a time until all the control plane components have been upgraded.

### Core component upgrades

EKS Anywhere `upgrade` also supports upgrading the following core components:

* Core CAPI
* CAPI providers
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
Anywhere components mentioned above to their latest available versions. All upgrade changes are backwards compatible except GitOps.

#### GitOps controller upgrade

__Upgrading an existing GitOps enabled cluster created by an older release of the EKS-A CLI (v0.5.0 and below) requires a manual update to the git repository structure.__ 
	
Before running the `upgrade` command on an existing GitOps enabled cluster, you must:
- move the eksa cluster config into a subdirectory under the management cluster's root level, following the [repo directory convention]({{< relref "./cluster-flux/#create-gitops-configuration-repo" >}})
- update the `clusterConfigPath` field of the `GitOpsConfig` of the `eksa-cluster.yaml` to reflect the new path of the `eksa-system` directory. For example:

```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: GitOpsConfig
metadata:
  name: management-cluster
spec:
  flux:
    github:
      clusterConfigPath: clusters/management-cluster/management-cluster # old: clusters/management-cluster
      ...
```

After pushing the changes to the git repo, user can safely run the `upgrade` command, and the Flux components will be upgraded.


### Performing a cluster upgrade

To perform a cluster upgrade you can modify your cluster specification `kubernetesVersion` field to the desired version.

As an example, to upgrade a cluster with version 1.20 to 1.21 you would change your spec

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
  kubernetesVersion: "1.21"
      ...
```

>**_NOTE:_** If you have a custom machine image for your nodes you may also need to update your `vsphereMachineConfig` with a new `template`.

and then you will run the command

```
eksctl anywhere upgrade cluster -f cluster.yaml
```

This will upgrade the cluster specification (if specified), upgrade the core components to the latest available versions and apply the changes using the provisioner controllers.

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
Pausing EKS-A cluster controller reconcile
Pausing Flux kustomization
GitOps field not specified, pause flux kustomization skipped
Creating bootstrap cluster
Installing cluster-api providers on bootstrap cluster
Moving cluster management from workload to bootstrap cluster
Upgrading workload cluster
Moving cluster management from bootstrap to workload cluster
Applying new EKS-A cluster resource; resuming reconcile
Resuming EKS-A controller reconciliation
Updating Git Repo with new EKS-A cluster spec
GitOps field not specified, update git repo skipped
Forcing reconcile Git repo with latest commit
GitOps not configured, force reconcile flux git repo skipped
Resuming Flux kustomization
GitOps field not specified, resume flux kustomization skipped
```

### Upgradeable Cluster Attributes
EKS Anywhere `upgrade` supports upgrading more than just the `kubernetesVersion`, 
allowing you to upgrade a number of fields simultaneously with the same procedure.

#### Upgradeable Attributes

`Cluster`:
- `kubernetesVersion`
- `controlPlaneConfig.count`
- `controlPlaneConfigurations.machineGroupRef.name`
- `workerNodeGroupConfigurations.count`
- `workerNodeGroupConfigurations.machineGroupRef.name`
- `etcdConfiguration.externalConfiguration.machineGroupRef.name`

`VSphereMachineConfig`:
- `VSphereMachineConfig.datastore`
- `VSphereMachineConfig.diskGiB`
- `VSphereMachineConfig.folder`
- `VSphereMachineConfig.memoryMiB`
- `VSphereMachineConfig.numCPUs`
- `VSphereMachineConfig.resourcePool`
- `VSphereMachineConfig.template`

### Troubleshooting

Attempting to upgrade a cluster with more than 1 minor release will result in receiving the following error.

```
✅ validate immutable fields
❌ validation failed    {"validation": "Upgrade preflight validations", "error": "validation failed with 1 errors: WARNING: version difference between upgrade version (1.21) and server version (1.19) do not meet the supported version increment of +1", "remediation": ""}
Error: failed to upgrade cluster: validations failed
```

For more errors you can see the [troubleshooting section]({{< relref "../troubleshoot" >}}).
