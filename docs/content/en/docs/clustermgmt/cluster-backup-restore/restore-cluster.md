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

### Cluster accessible and the infrastructure components not changed after etcd backup was taken

Similar to the failed management cluster without infrastructure components change situation, follow the [External etcd backup and restore]({{< relref "../etcd-backup-restore/etcdbackup" >}}) to restore the workload cluster itself from the backup.

### Cluster not accessible or infrastructure components changed after etcd backup was taken

If the workload cluster is still accessible, but the infrastructure machines are changed after the etcd backup was taken, you can still try restoring the cluster itself from the etcd backup. Although doing so is risky: it can potentially cause the node names, IPs and other infrastructure configurations to revert to a state that is no longer valid. Restoring etcd effectively takes a cluster back in time and all clients will experience a conflicting, parallel history. This can impact the behavior of watching components like Kubernetes controller managers, EKS Anywhere cluster controller manager, and Cluster API controller managers.

If the original workload cluster becomes inaccessible or cannot be restored to a healthy state from an outdated etcd, a new workload cluster needs to be created. This new cluster should be managed by the same management cluster that oversaw the original. You must then restore your workload applications to this new cluster from the etcd backup of the original. This ensures the management cluster retains control, with all data from the old cluster intact. Below is an example of applying the etcd `backup etcd-snapshot-w01.db` from a failed workload cluster `w01` to a new cluster `w02`:

1. Create a new workload cluster to which you will be migrating your workloads and applications from the original failed workload cluster.

    You can define a cluster config similar to your old workload cluster, with a different cluster name (if the old workload cluster still exists), and run cluster creation of the new workload cluster with the **exact same EKS Anywhere version** used to create the old workload cluster.

    ```bash
    export MGMT_CLUSTER="mgmt"
    export MGMT_CLUSTER_KUBECONFIG=${MGMT_CLUSTER}/${MGMT_CLUSTER}-eks-a-cluster.kubeconfig

    eksctl anywhere create cluster -f w02.yaml --kubeconfig $MGMT_CLUSTER_KUBECONFIG
    ```

1. Save the config map objects of the new workload cluster to a file.

    Save a copy of the new workload cluster's `cluster-info`, `kube-proxy` and `kubeadm-config` config map objects before the restore. This is necessary as the etcd restore will override the config maps above with the metadata information (certificates, endpoint, etc.) from the old cluster.

    ```bash
    export WORKLOAD_CLUSTER_NAME="w02"
    export WORKLOAD_CLUSTER_KUBECONFIG=${WORKLOAD_CLUSTER_NAME}/${WORKLOAD_CLUSTER_NAME}-eks-a-cluster.kubeconfig

    cat <<EOF >> w02-cm.yaml
    $(kubectl get -n kube-public cm cluster-info -oyaml --kubeconfig $WORKLOAD_CLUSTER_KUBECONFIG)
    ---
    $(kubectl get -n kube-system cm kube-proxy -oyaml --kubeconfig $WORKLOAD_CLUSTER_KUBECONFIG)
    ---
    $(kubectl get -n kube-system cm kubeadm-config -oyaml --kubeconfig $WORKLOAD_CLUSTER_KUBECONFIG)
    EOF
    ```

    Manually remove the `creationTimestamp`, `resourceVersion`, `uid` from the config map objects, so that later you can run `kubectl apply` against this file without errors.


1. Follow the [External etcd backup and restore]({{< relref "../etcd-backup-restore/etcdbackup" >}}) to restore the old workload cluster's etcd backup `etcd-snapshot-w01.db` onto the new workload cluster `w02`. Use different restore process based on OS family:
    * [BottleRocket]({{< relref "../etcd-backup-restore/bottlerocket-etcd-backup/#restore-etcd-from-backup" >}})
    * [Ubuntu]({{< relref "../etcd-backup-restore/ubuntu-rhel-etcd-backup/#restore" >}})

    {{% alert title="Warning" color="warning" %}}

    Do not unpause the cluster reconcilers as the workload cluster is not completely setup with required configurations yet. Unpausing the cluster reconcilers might cause unexpected machine rollout and leave the cluster in unhealthy state.

    {{% /alert %}}

    You might notice that after restoring the original etcd backup to the new workload cluster `w02`, all the nodes go to `NotReady` state with node names changed to have prefix `w01-*`. This is because restoring etcd effectively applies the node data from the original cluster which causes a conflicting history and can impact the behavior of watching components like Kubelets, Kubernetes controller managers.

    ```bash
    kubectl get nodes --kubeconfig $WORKLOAD_CLUSTER_KUBECONFIG

    NAME                              STATUS     ROLES           AGE     VERSION
    w01-bbtdd                         NotReady   control-plane   3d23h   v1.27.3-eks-6f07bbc
    w01-md-0-66dbcfb56cxng8lc-8ppv5   NotReady   <none>          3d23h   v1.27.3-eks-6f07bbc
    ```

1. Restart Kubelet of the control plane and worker nodes of the new workload cluster after the restore.

    For some cases, Kubelet on the node will automatically restart and nodes becomes ready. For other cases, you need to manually restart the Kubelet on all the control plane and worker nodes in order to bring back the nodes to ready state. Kubelet registers the node itself with the apisever which then updates etcd with the correct node data of the new workload cluster `w02`.

    {{< tabpane >}}
    {{< tab header="Ubuntu or RHEL" lang="bash" >}}
# SSH into the control plane and worker nodes. You must do this for each node.
ssh -i ${SSH_KEY} ${SSH_USERNAME}@<node IP>
sudo su
systemctl restart kubelet
    {{< /tab >}}
    {{< tab header="Bottlerocket" lang="bash" >}}
# SSH into the control plane and worker nodes. You must do this for each node.
ssh -i ${SSH_KEY} ${SSH_USERNAME}@<node IP>
apiclient exec admin bash
sheltie
systemctl restart kubelet
    {{< /tab >}}
    {{< /tabpane >}}

1. Add back `node-role.kubernetes.io/control-plane` label to **all** the control plane nodes.

    ```bash
    kubectl label nodes <control-plane-node-name> node-role.kubernetes.io/control-plane= --kubeconfig $WORKLOAD_CLUSTER_KUBECONFIG
    ```

1. Remove the lagacy nodes (if any).

    ```bash
    kubectl get nodes --kubeconfig $WORKLOAD_CLUSTER_KUBECONFIG

    NAME                              STATUS     ROLES           AGE     VERSION
    w01-bbtdd                         NotReady   control-plane   3d23h   v1.27.3-eks-6f07bbc
    w01-md-0-66dbcfb56cxng8lc-8ppv5   NotReady   <none>          3d23h   v1.27.3-eks-6f07bbc
    w02-fcbm2j                        Ready      control-plane   91m     v1.27.3-eks-6f07bbc
    w02-md-0-b7cc67cd4xd86jf-4c9ktp   Ready      <none>          73m     v1.27.3-eks-6f07bbc
    ```
    
    ```bash
    kubectl delete node w01-bbtdd --kubeconfig $WORKLOAD_CLUSTER_KUBECONFIG
    kubectl delete node w01-md-0-66dbcfb56cxng8lc-8ppv5 --kubeconfig $WORKLOAD_CLUSTER_KUBECONFIG
    ```

1. Re-apply the original config map objects of the workload cluster.

    Re-apply the `cluster-info`, `kube-proxy` and `kubeadm-config` config map objects we saved in previous step to the workload cluster `w02`.

    ```bash
    kubectl apply -f w02-cm.yaml --kubeconfig $WORKLOAD_CLUSTER_KUBECONFIG
    ```

1. Validate the nodes are in ready state.

    ```bash
    kubectl get nodes --kubeconfig $WORKLOAD_CLUSTER_KUBECONFIG

    NAME                              STATUS   ROLES           AGE     VERSION
    w02-djshz                         Ready    control-plane   9m7s    v1.27.3-eks-6f07bbc
    w02-md-0-6bbc8dd6d4xbgcjh-wfmb6   Ready    <none>          3m55s   v1.27.3-eks-6f07bbc
    ```

1. Restart the system pods to ensure that they use the config maps you re-applied in previous step.

    ```bash
    kubectl rollout restart ds kube-proxy -n kube-system --kubeconfig $WORKLOAD_CLUSTER_KUBECONFIG
    ```

1. Unpause the cluster reconcilers

    ```bash
    kubectl annotate clusters.anywhere.eks.amazonaws.com $WORKLOAD_CLUSTER_NAME anywhere.eks.amazonaws.com/paused- --kubeconfig=$MGMT_CLUSTER_KUBECONFIG

    kubectl patch clusters.cluster.x-k8s.io $WORKLOAD_CLUSTER_NAME --type merge -p '{"spec":{"paused": false}}' -n eksa-system --kubeconfig=$MGMT_CLUSTER_KUBECONFIG
    ```

1. Rollout and restart all the machine objects so that the workload cluster has a clean state.

    ```bash
    # Substitute the EKS Anywhere release version with whatever CLI version you are using
    EKSA_RELEASE_VERSION=v0.18.3
    BUNDLE_MANIFEST_URL=$(curl -s https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
    CLI_TOOLS_IMAGE=$(curl -s $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[0].eksa.cliTools.uri")


    # Rollout restart all the control plane machines
    docker run -i --network host -w $(pwd) -v $(pwd):/$(pwd) --entrypoint clusterctl ${CLI_TOOLS_IMAGE} alpha rollout restart kubeadmcontrolplane/${WORKLOAD_CLUSTER_NAME} -n eksa-system --kubeconfig ${MGMT_CLUSTER_KUBECONFIG}

    # Rollout restart all the worker machines
    # You need to repeat below command for each worker node group
    docker run -i --network host -w $(pwd) -v $(pwd):/$(pwd) --entrypoint clusterctl ${CLI_TOOLS_IMAGE} alpha rollout restart machinedeployment/${WORKLOAD_CLUSTER_NAME}-md-0 -n eksa-system --kubeconfig ${MGMT_CLUSTER_KUBECONFIG}
    ```

1. Validate the new workload cluster is in the desired state.

    ```bash
    kubectl get clusters -n default -o custom-columns="NAME:.metadata.name,READY:.status.conditions[?(@.type=='Ready')].status" --kubeconfig $MGMT_CLUSTER_KUBECONFIG

    NAME   READY
    mgmt   True
    w02    True

    kubectl get clusters.cluster.x-k8s.io -n eksa-system --kubeconfig $MGMT_CLUSTER_KUBECONFIG

    NAME   PHASE         AGE
    mgmt   Provisioned   11h   
    w02    Provisioned   11h 

    kubectl get kcp -n eksa-system  --kubeconfig $MGMT_CLUSTER_KUBECONFIG

    NAME   CLUSTER    INITIALIZED   API SERVER AVAILABLE   REPLICAS   READY   UPDATED   UNAVAILABLE   AGE   VERSION
    mgmt   mgmt       true          true                   2          2       2                       11h   v1.27.1-eks-1-27-4
    w02    w02        true          true                   2          2       2                       11h   v1.27.1-eks-1-27-4
    ```