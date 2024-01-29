---
title: "Restore cluster"
linkTitle: "Restore cluster"
weight: 20
aliases:
    /docs/tasks/cluster/cluster-backup-restore/restore-cluster/
description: >
  How to restore your EKS Anywhere cluster from backup
---

In certain unfortunate circumstances, an EKS Anywhere cluster may find itself in an unrecoverable state due to various factors such as a failed cluster upgrade, underlying infrastructure problems, or network issues, rendering the cluster inaccessible through conventional means. This document outlines detailed steps to guide you through the process of restoring a failed cluster from backups in these critical situations.

## Prerequisite

Always backup your EKS Anywhere cluster. Refer to the [Backup cluster]({{< relref "./backup-cluster" >}}) and make sure you have the updated etcd and Cluster API backup at hand.

## Restore a management cluster

As an EKS Anywhere management cluster contains the management components of itself, plus all the workload clusters it manages, the restoration process can be more complicated than just restoring all the objects from the etcd backup. To be more specific, all the core EKS Anywhere and Cluster API custom resources, that manage the lifecycle (provisioning, upgrading, operating, etc.) of the management and its workload clusters, are stored in the management cluster. This includes all the supporting infrastructure, like virtual machines, networks and load balancers. For example, after a failed cluster upgrade, the infrastructure components can change after the etcd backup was taken. Since the backup does not contain the new state of the half upgraded cluster, simply restoring it can create virtual machines UUID and IP mismatches, rendering EKS Anywhere incapable of healing the cluster.

Depending on whether the infrastructure components are changed or not after the etcd backup was taken (for example, if machines are rolled out and recreated and new IP addresses assigned to the machines), different strategy needs to be applied in order to restore the management cluster.

### Cluster accessible and the infrastructure components not changed after etcd backup was taken

If the management cluster is still accessible through the API server, and the underlying infrastructure layer (nodes, machines, VMs, etc.) are not changed after the etcd backup was taken, simply follow the [External etcd backup and restore]({{< relref "../etcd-backup-restore/etcdbackup" >}}) to restore the management cluster itself from the backup.

{{% alert title="Warning" color="warning" %}}

Do not apply the etcd restore unless you are very sure that the infrastructure layer is not changed after the etcd backup was taken. In other words, the nodes, machines, VMs, and their assigned IPs need to be exactly the same as when the backup was taken.

{{% /alert %}}

### Cluster not accessible or infrastructure components changed after etcd backup was taken

If the cluster is no longer accessible in any means, or the infrastructure machines are changed after the etcd backup was taken, restoring this management cluster itself from the outdated etcd backup will not work. Instead, you need to create a new management cluster, and migrate all the EKS Anywhere resources of the old workload clusters to the new one, so that the new management cluster can maintain the new ownership of managing the existing workload clusters. Below is an example of migrating a failed management cluster `mgmt-old` with its workload clusters `w01` and `w02` to a new management cluster `mgmt-new`:

1. Create a new management cluster to which you will be migrating your workload clusters later.

    You can define a cluster config similar to your old management cluster, and run cluster creation of the new management cluster with the **exact same EKS Anywhere version** used to create the old management cluster.

    If the original management cluster still exists with old infrastructure running, you need to create a new management cluster with a **different cluster name** to avoid conflict.

    ```sh
    eksctl anywhere create cluster -f mgmt-new.yaml
    ```

1. Move the custom resources of all the workload clusters to the new management cluster created above.

    Using the vSphere provider as an example, we are moving the Cluster API custom resources, such as `vpsherevms`, `vspheremachines` and `machines` of the **workload clusters**, from the old management cluster to the new management cluster created in above step. By using the `--filter-cluster` flag with the `clusterctl move` command, we are only targeting the custom resources from the workload clusters.


    ```bash
    # Use the same cluster name if the newly created management cluster has the same cluster name as the old one
    MGMT_CLUSTER_OLD="mgmt-old"
    MGMT_CLUSTER_NEW="mgmt-new"
    MGMT_CLUSTER_NEW_KUBECONFIG=${MGMT_CLUSTER_NEW}/${MGMT_CLUSTER_NEW}-eks-a-cluster.kubeconfig
    
    WORKLOAD_CLUSTER_1="w01"
    WORKLOAD_CLUSTER_2="w02"

    # Substitute the workspace path with the workspace you are using
    WORKSPACE_PATH="/home/ec2-user/eks-a"
    
    # Retrieve the Cluster API backup folder path that are automatically generated during the cluster upgrade
    # This folder contains all the resources that represent the cluster state of the old management cluster along with its workload clusters
    CLUSTER_STATE_BACKUP_LATEST=$(ls -Art ${WORKSPACE_PATH}/${MGMT_CLUSTER_OLD} | grep ${MGMT_CLUSTER_OLD}-backup | tail -1)
    CLUSTER_STATE_BACKUP_LATEST_PATH=${WORKSPACE_PATH}/${MGMT_CLUSTER_OLD}/${CLUSTER_STATE_BACKUP_LATEST}/

    # Substitute the EKS Anywhere release version with the EKS Anywhere version of the original management cluster
    EKSA_RELEASE_VERSION=v0.17.3
    BUNDLE_MANIFEST_URL=$(curl -s https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
    CLI_TOOLS_IMAGE=$(curl -s $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[0].eksa.cliTools.uri")

    # The clusterctl move command needs to be executed for each workload cluster.
    # It will only move the workload cluster resources from the EKS Anywhere backup to the new management cluster.
    # If you have multiple workload clusters, you have to run the command for each cluster as shown below.

    # Move workload cluster w01 resources to the new management cluster mgmt-new
    docker run -i --network host -w $(pwd) -v $(pwd):/$(pwd) --entrypoint clusterctl ${CLI_TOOLS_IMAGE} move \
        --namespace eksa-system \
        --filter-cluster ${WORKLOAD_CLUSTER_1} \
        --from-directory ${CLUSTER_STATE_BACKUP_LATEST_PATH} \
        --to-kubeconfig ${MGMT_CLUSTER_NEW_KUBECONFIG}
    
    # Move workload cluster w02 resources to the new management cluster mgmt-new
    docker run -i --network host -w $(pwd) -v $(pwd):/$(pwd) --entrypoint clusterctl ${CLI_TOOLS_IMAGE} move \
        --namespace eksa-system \
        --filter-cluster ${WORKLOAD_CLUSTER_2} \
        --from-directory ${CLUSTER_STATE_BACKUP_LATEST_PATH} \
        --to-kubeconfig ${MGMT_CLUSTER_NEW_KUBECONFIG}
    ```

1. (Optional) Update the cluster config file of the workload clusters if the new management cluster has a different cluster name than the original management cluster.

    You can **skip this step** if the new management cluster has the same cluster name as the old management cluster.

    ```yaml
    # workload cluster w01
    ---
    apiVersion: anywhere.eks.amazonaws.com/v1alpha1
    kind: Cluster
    metadata:
      name: w01
      namespace: default
    spec:
      managementCluster:
        name: mgmt-new # This needs to be updated with the new management cluster name.
      ...
    ```

    ```yaml
    # workload cluster w02
    ---
    apiVersion: anywhere.eks.amazonaws.com/v1alpha1
    kind: Cluster
    metadata:
      name: w02
      namespace: default
    spec:
      managementCluster:
        name: mgmt-new # This needs to be updated with the new management cluster name.
      ...
    ```

    Make sure that apart from the `managementCluster` field you updated above, all the other cluster configs of the workload clusters need to stay the same as the old workload clusters resources after the old management cluster fails.

1. Apply the updated cluster config of each workload cluster in the new management cluster.

    ```bash
    MGMT_CLUSTER_NEW="mgmt-new"
    MGMT_CLUSTER_NEW_KUBECONFIG=${MGMT_CLUSTER_NEW}/${MGMT_CLUSTER_NEW}-eks-a-cluster.kubeconfig

    kubectl apply -f w01/w01-eks-a-cluster.yaml --kubeconfig ${MGMT_CLUSTER_NEW_KUBECONFIG}
    kubectl apply -f w02/w02-eks-a-cluster.yaml --kubeconfig ${MGMT_CLUSTER_NEW_KUBECONFIG}
    ```

1. Validate all clusters are in the desired state.

    ```bash
    kubectl get clusters -n default -o custom-columns="NAME:.metadata.name,READY:.status.conditions[?(@.type=='Ready')].status" --kubeconfig ${MGMT_CLUSTER_NEW}/${MGMT_CLUSTER_NEW}-eks-a-cluster.kubeconfig

    NAME       READY
    mgmt-new   True
    w01        True
    w02        True

    kubectl get clusters.cluster.x-k8s.io -n eksa-system --kubeconfig ${MGMT_CLUSTER_NEW}/${MGMT_CLUSTER_NEW}-eks-a-cluster.kubeconfig

    NAME       PHASE         AGE
    mgmt-new   Provisioned   11h   
    w01        Provisioned   11h   
    w02        Provisioned   11h 

    kubectl get kcp -n eksa-system  --kubeconfig ${MGMT_CLUSTER_NEW}/${MGMT_CLUSTER_NEW}-eks-a-cluster.kubeconfig

    NAME       CLUSTER    INITIALIZED   API SERVER AVAILABLE   REPLICAS   READY   UPDATED   UNAVAILABLE   AGE   VERSION
    mgmt-new   mgmt-new   true          true                   2          2       2                       11h   v1.27.1-eks-1-27-4
    w01        w01        true          true                   2          2       2                       11h   v1.27.1-eks-1-27-4
    w02        w02        true          true                   2          2       2                       11h   v1.27.1-eks-1-27-4
    ```

## Restore a workload cluster

Restoring a workload cluster is a delicate process. If you have an [EKS Anywhere Enterprise Subscription](https://aws.amazon.com/eks/eks-anywhere/pricing/), please contact AWS support team if you wish to perform such operation.