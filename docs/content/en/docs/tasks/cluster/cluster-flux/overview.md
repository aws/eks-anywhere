---
toc_hide: true
hide_summary: true
---

EKS Anywhere supports a [GitOps](https://www.weave.works/technologies/gitops/) workflow for the management of your cluster.

When you create a cluster with GitOps enabled, EKS Anywhere will automatically commit your cluster configuration to the provided GitHub repository and install a GitOps toolkit on your cluster which watches that committed configuration file.
You can then manage the scale of the cluster by making changes to the version controlled cluster configuration file and committing the changes.
Once a change is detected by the GitOps controller running in your cluster, the scale of the cluster will be adjusted to match the committed configuration file.

If you'd like to learn more about GitOps and the associated best practices, [check out this introduction from Weaveworks](https://www.weave.works/technologies/gitops/).

>**_NOTE:_** Installing a GitOps controller needs to be done during cluster creation.
In the event that GitOps installation fails, EKS Anywhere cluster creation will continue.

### Supported Cluster Properties

Currently, you can manage a subset of cluster properties with GitOps:

`Cluster`:
- `Cluster.workerNodeGroupConfigurations[0].count`
- `Cluster.workerNodeGroupConfigurations[0].machineGroupRef.name`

`Worker Nodes`:
- `VSphereMachineConfig.datastore`
- `VSphereMachineConfig.diskGiB`
- `VSphereMachineConfig.folder`
- `VSphereMachineConfig.memoryMiB`
- `VSphereMachineConfig.numCPUs`
- `VSphereMachineConfig.resourcePool`

Any other changes to the cluster configuration in the git repository will be ignored.
If an immutable immutable field is changed in Git repsitory, there are two ways to find the error message:
1. If a notification webhook is set up, check the error message in notification channel.
2. Check the Flux Kustomization Controller log: `kubectl logs -f -n flux-system kustomize-controller-******` for error message containing text similar to `Invalid value: 1: field is immutable`
