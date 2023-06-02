---
title: "Delete cluster"
linkTitle: "Delete cluster"
weight: 90
aliases:
    /docs/tasks/cluster/cluster-delete/
date: 2017-01-05
description: >
  How to delete an EKS Anywhere cluster
---

>**_NOTE_**: EKS Anywhere Bare Metal clusters do not yet support separate workload and management clusters. Use the instructions for Deleting a management cluster to delete a Bare Metal cluster.
>

### Deleting a workload cluster

Follow these steps to delete your EKS Anywhere cluster that is managed by a separate management cluster.

To delete a workload cluster, you will need:
- name of your workload cluster
- kubeconfig of your workload cluster
- kubeconfig of your management cluster

Run the following commands to delete the cluster:

1. Set up `CLUSTER_NAME` and `KUBECONFIG` environment variables:
    ```bash
    export CLUSTER_NAME=eksa-w01-cluster
    export KUBECONFIG=${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
    export MANAGEMENT_KUBECONFIG=<path-to-management-cluster-kubeconfig>
    ```

2. Run the delete command:
- If you are running the delete command from the directory which has the cluster folder with `${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.yaml`:

  ```bash
  eksctl anywhere delete cluster ${CLUSTER_NAME} --kubeconfig ${MANAGEMENT_KUBECONFIG}
  ```

### Deleting a management cluster

Follow these steps to delete your management cluster.

To delete a cluster you will need:
- cluster name or cluster configuration 
- kubeconfig of your cluster

Run the following commands to delete the cluster:

1. Set up `CLUSTER_NAME` and `KUBECONFIG` environment variables:
    ```bash
    export CLUSTER_NAME=mgmt
    export KUBECONFIG=${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
    ```

2. Run the delete command:
- If you are running the delete command from the directory which has the cluster folder with `${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.yaml`:

  ```bash
  eksctl anywhere delete cluster ${CLUSTER_NAME}
  ```

- Otherwise, use this command to manually specify the clusterconfig file path:
  ```bash
  export CONFIG_FILE=<path-to-config-file>
  eksctl anywhere delete cluster -f ${CONFIG_FILE}
  ```

Example output:
```
Performing provider setup and validations
Creating management cluster
Installing cluster-api providers on management cluster
Moving cluster management from workload cluster
Deleting workload cluster
Clean up Git Repo
GitOps field not specified, clean up git repo skipped
🎉 Cluster deleted!
```

For vSphere, CloudStack, and Nutanix, this will delete all of the VMs that were created in your provider.
For Bare Metal, the servers will be powered off if BMC information has been provided.
If your workloads created external resources such as external DNS entries or load balancer endpoints you may need to delete those resources manually.
