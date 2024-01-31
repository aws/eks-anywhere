
---
title: "Reboot nodes"
linkTitle: "Reboot nodes"
weight: 85
aliases:
    /docs/tasks/cluster-rebootnode/
date: 2017-01-05
description: >
  How to properly reboot a node in an EKS Anywhere cluster
---

If you need to reboot a node in your cluster for maintenance or any other reason, performing the following steps will help prevent possible disruption of services on those nodes:

{{% alert title="Warning" color="primary" %}}
Rebooting a cluster node as described here is good for all nodes, but is critically important when rebooting a Bottlerocket node running the `boots` service on a Bare Metal cluster.
If it does go down while running the `boots` service, the Bottlerocket node will not be able to boot again until the `boots` service is restored on another machine. This is because Bottlerocket must get its address from a DHCP service.
{{% /alert %}}

1. On your admin machine, set the following environment variables that will come in handy later
```bash
export CLUSTER_NAME=mgmt
export MGMT_KUBECONFIG=${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
```

1. [Backup cluster]({{< relref "/docs/clustermgmt/cluster-backup-restore/backup-cluster" >}}) 

    This ensures that there is an up-to-date cluster state available for restoration in the case that the cluster experiences issues or becomes unrecoverable during reboot.

1. Verify DHCP lease time will be longer than the maintenance time, and that IPs will be the same before and after maintenance. 
    
    This step is critical in ensuring the cluster will be healthy after reboot. If IPs are not preserved before and after reboot, the cluster may not be recoverable.
    
    {{% alert title="Warning" color="warning" %}}
If this cannot be verified, do not proceed any further
    {{% /alert %}}

1. Pause the reconciliation of the cluster being shut down. 

    This ensures that the EKS Anywhere cluster controller will not reconcile on the nodes that are down and try to remediate them.

    - add the paused annotation to the EKSA clusters and CAPI clusters: 
    ```bash
    kubectl annotate clusters.anywhere.eks.amazonaws.com $CLUSTER_NAME anywhere.eks.amazonaws.com/paused=true --kubeconfig=$MGMT_KUBECONFIG

    kubectl patch clusters.cluster.x-k8s.io $CLUSTER_NAME --type merge -p '{"spec":{"paused": true}}' -n eksa-system --kubeconfig=$MGMT_KUBECONFIG
    ```

1. For all of the nodes in the cluster, perform the following steps in this order: worker nodes, control plane nodes, and etcd nodes.

    1. Cordon the node so no further workloads are scheduled to run on it:

        ```bash
        kubectl cordon <node-name>
        ```

    1. Drain the node of all current workloads:

        ```bash
        kubectl drain <node-name>
        ```

    1. Using the appropriate method for your provider, shut down the node. 


1. Perform system maintenance or other tasks you need to do on each node. Then boot up the node in this order: etcd nodes, control plane nodes, and worker nodes.

1. Uncordon the nodes so that they can begin receiving workloads again.

    ```bash
    kubectl uncordon <node-name>
    ```

1. Remove the paused annotations from EKS Anywhere cluster.
    ```bash
    kubectl annotate clusters.anywhere.eks.amazonaws.com $CLUSTER_NAME anywhere.eks.amazonaws.com/paused- --kubeconfig=$MGMT_KUBECONFIG

    kubectl patch clusters.cluster.x-k8s.io $CLUSTER_NAME --type merge -p '{"spec":{"paused": false}}' -n eksa-system --kubeconfig=$MGMT_KUBECONFIG
    ```
