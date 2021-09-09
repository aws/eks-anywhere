
---
title: "Delete cluster"
linkTitle: "Delete cluster"
weight: 40
date: 2017-01-05
description: >
  How to delete an EKS Anywhere cluster
---

To delete a cluster you will need:
- cluster name or cluster configuration 
- kubeconfig for your cluster

Run the following commands to delete the cluster:

1. Set up `CLUSTER_NAME` and `KUBECONFIG` environment variables:
    ```bash
    export CLUSTER_NAME=dev
    export KUBECONFIG=${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
    ```

2. Run the delete command:
- If you are running the delete command from the directory that has the cluster folder with `${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.yaml`:

  ```bash
  eksctl anywhere delete cluster ${CLUSTER_NAME}
  ```

- Otherwise, use this command to specify the clusterconfig filepath manually:
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

This will delete all of the VMs that were created in your provider.
If your workloads created external resources such as external DNS entries or load balancer endpoints, you may need to delete those resources manually.
