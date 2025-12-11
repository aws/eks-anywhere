---
title: "Scale Bare Metal cluster"
linkTitle: "Scale Bare Metal cluster"
weight: 20
aliases:
    /docs/tasks/cluster/cluster-scale/baremetal-scale/
date: 2017-01-05
description: >
  How to scale your Bare Metal cluster
---

### Scaling nodes on Bare Metal clusters
When you are horizontally scaling your Bare Metal EKS Anywhere cluster, consider the number of nodes you need for your control plane and for your data plane.

See the [Kubernetes Components](https://kubernetes.io/docs/concepts/overview/components/) documentation to learn the differences between the control plane and the data plane (worker nodes).

Horizontally scaling the cluster is done by increasing the number for the control plane or worker node groups under the Cluster specification.

>**_NOTE:_** If etcd is running on your control plane (the default configuration) you should scale your control plane in odd numbers (3, 5, 7...).

```
apiVersion: anywhere.eks.amazonaws.com/v1
kind: Cluster
metadata:
  name: test-cluster
spec:
  controlPlaneConfiguration:
    count: 1     # increase this number to horizontally scale your control plane
...    
  workerNodeGroupConfigurations:
  - count: 1     # increase this number to horizontally scale your data plane
```

Next, you must ensure you have enough available hardware for the scale-up operation to function. Available hardware could have been fed to the cluster as extra hardware during a prior create command, or could be fed to the cluster during the scale-up process by providing the hardware CSV file to the upgrade cluster command (explained in detail below).
For scale-down operation, you can skip directly to the [upgrade cluster command]({{< relref "baremetal-scale/#upgrade-cluster-command-for-scale-updown" >}}).

To check if you have enough available hardware for scale up, you can use the `kubectl` command below to check if there are hardware with the selector labels corresponding to the controlplane/worker node group and without the `ownerName` label. 

```bash
kubectl get hardware -n eksa-system --show-labels
```

For example, if you want to scale a worker node group with selector label `type=worker-group-1`, then you must have an additional hardware object in your cluster with the label `type=worker-group-1` that doesn't have the `ownerName` label. 

In the command shown below, `eksa-worker2` matches the selector label and it doesn't have the `ownerName` label. Thus, it can be used to scale up `worker-group-1` by 1.

```bash
kubectl get hardware -n eksa-system --show-labels 
NAME                STATE       LABELS
eksa-controlplane               type=controlplane,v1alpha1.tinkerbell.org/ownerName=abhnvp-control-plane-template-1656427179688-9rm5f,v1alpha1.tinkerbell.org/ownerNamespace=eksa-system
eksa-worker1                    type=worker-group-1,v1alpha1.tinkerbell.org/ownerName=abhnvp-md-0-1656427179689-9fqnx,v1alpha1.tinkerbell.org/ownerNamespace=eksa-system
eksa-worker2                    type=worker-group-1
```

If you don't have any available hardware that match this requirement in the cluster, you can [setup a new hardware CSV]({{< relref "../../getting-started/baremetal/bare-preparation/#prepare-hardware-inventory" >}}). You can feed this hardware inventory file during the [upgrade cluster command]({{< relref "baremetal-scale/#upgrade-cluster-command-for-scale-updown" >}}).

#### Upgrade Cluster Command for Scale Up/Down

1. **eksctl CLI**: To upgrade a workload cluster with eksctl, run:
    ```bash
    eksctl anywhere upgrade cluster 
    -f cluster.yaml 
    # --hardware-csv <hardware.csv> \ # uncomment to add more hardware
   --kubeconfig mgmt/mgmt-eks-a-cluster.kubeconfig
    ```
    As noted earlier, adding the `--kubeconfig` option tells `eksctl` to use the management cluster identified by that kubeconfig file to create a different workload cluster.

2. **kubectl CLI**: The cluster lifecycle feature lets you use kubectl to talk to the Kubernetes API to scale a workload cluster. kubectl can also be used for management cluster scaling, except when upgrading the EKS Anywhere CLI version which requires `eksctl anywhere upgrade`. To use kubectl, run:
     ```bash
     kubectl apply -f eksa-w01-cluster.yaml --kubeconfig mgmt/mgmt-eks-a-cluster.kubeconfig
     ```

    To check the state of a cluster managed with the cluster lifecyle feature, use `kubectl` to show the cluster object with its status.
    
    The `status` field on the cluster object field holds information about the current state of the cluster.

    ```
    kubectl get clusters w01 -o yaml
    ```

    The cluster has been fully upgraded once the status of the `Ready` condition is marked `True`.
    See the [cluster status]({{< relref "../cluster-status" >}}) guide for more information.

3. **GitOps**: See [Manage separate workload clusters with GitOps]({{< relref "../cluster-flux.md#manage-separate-workload-clusters-using-gitops" >}})

4. **Terraform**: See [Manage separate workload clusters with Terraform]({{< relref "../cluster-terraform.md#manage-separate-workload-clusters-using-terraform" >}})

   >**NOTE**:For kubectl, GitOps and Terraform:
   > * The baremetal controller does not support scaling upgrades and Kubernetes version upgrades in the same request.
   > * While scaling workload cluster if you need to add additional machines, run:
   >   ```
   >   eksctl anywhere generate hardware -z updated-hardware.csv > updated-hardware.yaml
   >   kubectl apply -f updated-hardware.yaml
   >   ```
   > *  For scaling multiple workload clusters, it is essential that the hardware that will be used for scaling up clusters has labels and selectors that are unique to the target workload cluster. For instance, for an EKSA cluster named `eksa-workload1`, the hardware that is assigned for this cluster should have labels that are only going to be used for this cluster like `type=eksa-workload1-cp` and `type=eksa-workload1-worker`. Another workload cluster named `eksa-workload2` can have labels like `type=eksa-workload2-cp` and `type=eksa-workload2-worker`. Please note that even though labels can be arbitrary, they need to be unique for each workload cluster. Not specifying unique cluster labels can cause cluster upgrades to behave in unexpected ways which may lead to unsuccessful upgrades and unstable clusters.

### Selective Server Removal During Scale Down

When scaling down your Bare Metal EKS Anywhere cluster, you may want to remove a specific server rather than letting the system automatically choose which node to remove. By default, Cluster API will remove nodes in a non-deterministic order during scale down operations.

>**_IMPORTANT:_** Do not attempt to remove specific servers by deleting entries from the hardware CSV file and running the upgrade command. This approach will not have deterministic behavior.

To target a specific server for removal during scale down, use the Cluster API [delete annotation](https://main.cluster-api.sigs.k8s.io/reference/api/labels-and-annotations#:~:text=cluster.x%2Dk8s.io/delete%2Dmachine,User) on the corresponding machine resource. This ensures that the specific machine is prioritized for removal during the scale-down operation.

#### Steps to Remove a Specific Server

1. **Identify the target node**: First, determine which node you want to remove by listing all nodes in your cluster:
   ```bash
   kubectl get nodes
   ```

2. **Find the corresponding CAPI machine**: List all CAPI machines to find the machine resource that corresponds to your target node:
   ```bash
   kubectl get machines.cluster.x-k8s.io -n eksa-system
   ```
   
   You can also find the specific machine for a node using:
   ```bash
   NODE_NAME="your-target-node-name"
   kubectl get machines.cluster.x-k8s.io -n eksa-system --no-headers | awk -v node="$NODE_NAME" '$3==node {print $1}'
   ```

3. **Add the delete annotation**: Mark the machine for deletion by adding the Cluster API delete annotation:
   ```bash
   MACHINE_NAME="your-machine-name"
   kubectl annotate machines.cluster.x-k8s.io $MACHINE_NAME cluster.x-k8s.io/delete-machine=yes -n eksa-system
   ```

4. **Update your cluster specification**: Reduce the node count in your EKS Anywhere cluster specification file:
   ```yaml
   apiVersion: anywhere.eks.amazonaws.com/v1
   kind: Cluster
   metadata:
     name: test-cluster
   spec:
     workerNodeGroupConfigurations:
     - count: 2     # decrease this number for scale down
   ```

5. **Apply the scale down operation**: Run the upgrade command to perform the scale down:
   ```bash
   eksctl anywhere upgrade cluster -f cluster.yaml --kubeconfig mgmt/mgmt-eks-a-cluster.kubeconfig
   ```

The annotated machine will be prioritized for removal during the scale-down operation, ensuring that your specific target server is removed from the cluster.

>**_NOTE:_** The hardware CSV is designed to manage the hardware catalog by creating [Hardware](https://doc.crds.dev/github.com/tinkerbell/tink/tinkerbell.org/Hardware/v1alpha1@v0.12.2), [Machine](https://doc.crds.dev/github.com/tinkerbell/rufio/bmc.tinkerbell.org/Machine/v1alpha1@v0.6.3), and BMC credential resources. For operational tasks like scaling, always use the cluster specification count and CAPI machine annotations rather than modifying the hardware CSV.

### Autoscaling


EKS Anywhere supports autoscaling of worker node groups using the [Kubernetes Cluster Autoscaler](https://github.com/kubernetes/autoscaler/) and as a [curated package]({{< relref "../../packages/cluster-autoscaler/" >}}).

See [here](https://github.com/kubernetes/autoscaler/) and as a [curated package]({{< relref "../../getting-started/optional/autoscaling/" >}}) for details on how to configure your cluster spec to autoscale worker node groups for autoscaling.
