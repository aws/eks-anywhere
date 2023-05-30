---
title: "Bottlerocket"
linkTitle: "On Bottlerocket"
weight: 11
aliases:
    /docs/tasks/etcd-backup-restore/bottlerocket-etcd-backup/
date: 2021-11-04
description: >
  How to backup and restore External ETCD on Bottlerocket OS
---
{{% alert title="Note" color="warning" %}}
External etcd topology is supported for vSphere, CloudStack and Snow clusters, but not yet for Bare Metal or Nutanix clusters.
{{% /alert %}}

This guide requires some common shell tools such as: 
* `grep`
* `xargs`
* `ssh`
* `scp`
* `cut`

Make sure you have these installed on your admin machine before continuing.

### Admin machine environment variables setup
On your admin machine, set the following environment variables that will later come in handy

```bash
export MANAGEMENT_CLUSTER_NAME="eksa-management"     # Set this to the management cluster name
export CLUSTER_NAME="eksa-workload"                  # Set this to name of the cluster you want to backup (management or workload)
export SSH_KEY="path-to-private-ssh-key"             # Set this to the cluster's private SSH key path
export SSH_USERNAME="ec2-user"                       # Set this to the SSH username
export SNAPSHOT_PATH="/tmp/snapshot.db"              # Set this to the path where you want the ETCD snapshot to be saved

export MANAGEMENT_KUBECONFIG=${MANAGEMENT_CLUSTER_NAME}/${MANAGEMENT_CLUSTER_NAME}-eks-a-cluster.kubeconfig
export CLUSTER_KUBECONFIG=${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
export ETCD_ENDPOINTS=$(kubectl --kubeconfig=${MANAGEMENT_KUBECONFIG} -n eksa-system get machines --selector cluster.x-k8s.io/cluster-name=${CLUSTER_NAME},cluster.x-k8s.io/etcd-cluster=${CLUSTER_NAME}-etcd -ojsonpath='{.items[*].status.addresses[0].address}')
export CONTROL_PLANE_ENDPOINTS=($(kubectl --kubeconfig=${MANAGEMENT_KUBECONFIG} -n eksa-system get machines --selector cluster.x-k8s.io/control-plane-name=${CLUSTER_NAME} -ojsonpath='{.items[*].status.addresses[0].address}'))
```

### Prepare ETCD nodes for backup and restore
Install SCP on the ETCD nodes:
```bash
echo -n ${ETCD_ENDPOINTS} | xargs -I {} -d" " ssh -o StrictHostKeyChecking=no -i ${SSH_KEY} ${SSH_USERNAME}@{} sudo yum -y install openssh-clients
```

### Create ETCD Backup
Make sure to setup the [admin environment variables]({{< relref "#admin-machine-environment-variables-setup" >}}) and [prepare your ETCD nodes for backup]({{< relref "#prepare-etcd-nodes-for-backup-and-restore" >}}) before moving forward.

1. SSH into one of the ETCD nodes
    ```bash
    export ETCD_NODE=$(echo -n ${ETCD_ENDPOINTS} | cut -d " " -f1)
    ssh -i ${SSH_KEY} ${SSH_USERNAME}@${ETCD_NODE}
    ```

1. Drop into Bottlerocket's root shell
    ```bash
    sudo sheltie
    ```

1. Set these environment variables
    ```bash
    # get the container ID corresponding to ETCD pod
    export ETCD_CONTAINER_ID=$(ctr -n k8s.io c ls | grep -w "etcd-io" | cut -d " " -f1)

    # get the ETCD endpoint
    export ETCD_ENDPOINT=$(cat /etc/kubernetes/manifests/etcd | grep -wA1 ETCD_ADVERTISE_CLIENT_URLS | tail -1 | grep -oE '[^ ]+$')
    ```

1. Create the ETCD snapshot
    ```bash
    ctr -n k8s.io t exec -t --exec-id etcd ${ETCD_CONTAINER_ID} etcdctl \
        --endpoints=${ETCD_ENDPOINT} \
        --cacert=/var/lib/etcd/pki/ca.crt \
        --cert=/var/lib/etcd/pki/server.crt \
        --key=/var/lib/etcd/pki/server.key \
        snapshot save /var/lib/etcd/data/etcd-backup.db
    ```

1. Move the snapshot to another directory and set proper permissions
    ```bash
    mv /var/lib/etcd/data/etcd-backup.db /run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/home/ec2-user/snapshot.db
    chown 1000 /run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/home/ec2-user/snapshot.db
    ```

1. Exit out of ETCD node. You will have to type `exit` twice to get back to the admin machine
    ```bash
    exit
    exit
    ```

1. Copy over the snapshot from the ETCD node
    ```bash
    scp -i ${SSH_KEY} ${SSH_USERNAME}@${ETCD_NODE}:/home/ec2-user/snapshot.db ${SNAPSHOT_PATH}
    ```

You should now have the ETCD snapshot in your current working directory.

### Restore ETCD from Backup
Make sure to setup the [admin environment variables]({{< relref "#admin-machine-environment-variables-setup" >}}) and [prepare your ETCD nodes for restore]({{< relref "#prepare-etcd-nodes-for-backup-and-restore" >}}) before moving forward.

1. Pause cluster reconciliation

    Before starting the process of restoring ETCD, you have to pause some cluster reconciliation objects so EKS Anywhere doesn't try to perform any operations on the cluster while you restore the ETCD snapshot.
    ```bash
    # Pause control plane reconcilation
    kubectl --kubeconfig=${MANAGEMENT_KUBECONFIG} -n eksa-system annotate machinehealthchecks ${CLUSTER_NAME}-kcp-unhealthy cluster.x-k8s.io/paused=true

    # Pause ETCD reconcilation
    kubectl --kubeconfig=${MANAGEMENT_KUBECONFIG} -n eksa-system annotate etcdadmclusters ${CLUSTER_NAME}-etcd cluster.x-k8s.io/paused=true
    ```

2. Stop control plane core components
    
    You also need to stop the control plane core components so the Kubernetes API server doesn't try to communicate with ETCD while you perform ETCD operations.

    - You can use this command to get the control plane node IPs which you can use to SSH
    ```bash
    echo -n ${CONTROL_PLANE_ENDPOINTS[@]} | xargs -I {} -d " " echo "{}"
    ```

    - SSH into the node and stop the core components. You must do this for each control plane node.
    ```bash
    # SSH into the control plane node using the IPs printed in previous command
    ssh -i ${SSH_KEY} ${SSH_USERNAME}@<Control Plane IP from previous command>

    # drop into bottlerocket root shell
    sudo sheltie

    # create a temporary directory and move the static manifests to it
    mkdir -p /tmp/temp-manifests
    mv /etc/kubernetes/manifests/* /tmp/temp-manifests
    ```

    - Exit out of the Bottlerocket node
    ```bash
    # exit from bottlerocket's root shell
    exit

    # exit from bottlerocket node
    exit
    ```
    Repeat these steps for each control plane node.

1. Copy the backed-up ETCD snapshot to all the ETCD nodes
    ```bash
    echo -n ${ETCD_ENDPOINTS} | xargs -I {} -d" " scp -o StrictHostKeyChecking=no -i ${SSH_KEY} ${SNAPSHOT_PATH} ${SSH_USERNAME}@{}:/home/ec2-user
    ```

1. Perform the ETCD restore
    
    For this step, you have to SSH into each ETCD node and run the restore command.
    - Get ETCD nodes IPs for SSH'ing into the nodes
    ```bash
    # This should print out all the control plane IPs
    echo -n ${ETCD_ENDPOINTS} | xargs -I {} -d " " echo "{}"
    ```

    ```bash
    # SSH into the control plane node using the IPs printed in previous command
    ssh -i ${SSH_KEY} ${SSH_USERNAME}@<ETCD IP from previous command>

    # drop into bottlerocket's root shell
    sudo sheltie

    # copy over the ETCD snapshot to the appropriate location
    cp /run/host-containerd/io.containerd.runtime.v2.task/default/admin/rootfs/home/ec2-user/snapshot.db /var/lib/etcd/data/etcd-snapshot.db

    # setup the ETCD environment
    export ETCD_NAME=$(cat /etc/kubernetes/manifests/etcd | grep -wA1 ETCD_NAME | tail -1 | grep -oE '[^ ]+$')
    export ETCD_INITIAL_ADVERTISE_PEER_URLS=$(cat /etc/kubernetes/manifests/etcd | grep -wA1 ETCD_INITIAL_ADVERTISE_PEER_URLS | tail -1 | grep -oE '[^ ]+$')
    export ETCD_INITIAL_CLUSTER=$(cat /etc/kubernetes/manifests/etcd | grep -wA1 ETCD_INITIAL_CLUSTER | tail -1 | grep -oE '[^ ]+$')
    export INITIAL_CLUSTER_TOKEN="etcd-cluster-1"

    # get the container ID corresponding to ETCD pod
    export ETCD_CONTAINER_ID=$(ctr -n k8s.io c ls | grep -w "etcd-io" | cut -d " " -f1)

    # run the restore command
    ctr -n k8s.io t exec -t --exec-id etcd ${ETCD_CONTAINER_ID} etcdctl \
        snapshot restore /var/lib/etcd/data/etcd-snapshot.db \
        --name=${ETCD_NAME} \
        --initial-cluster=${ETCD_INITIAL_CLUSTER} \
        --initial-cluster-token=${INITIAL_CLUSTER_TOKEN} \
        --initial-advertise-peer-urls=${ETCD_INITIAL_ADVERTISE_PEER_URLS} \
        --cacert=/var/lib/etcd/pki/ca.crt \
        --cert=/var/lib/etcd/pki/server.crt \
        --key=/var/lib/etcd/pki/server.key 

    # move the ETCD data files out of the container to a temporary location
    mkdir -p /tmp/etcd-files
    $(ctr -n k8s.io snapshot mounts /tmp/etcd-files/ ${ETCD_CONTAINER_ID})
    mv /tmp/etcd-files/${ETCD_NAME}.etcd /tmp/

    # stop the ETCD pod
    mkdir -p /tmp/temp-manifests
    mv /etc/kubernetes/manifests/* /tmp/temp-manifests

    # backup the previous ETCD data files
    mv /var/lib/etcd/data/member /var/lib/etcd/data/member.backup

    # copy over the new ETCD data files to the data directory
    mv /tmp/${ETCD_NAME}.etcd/member /var/lib/etcd/data/

    # re-start the ETCD pod
    mv /tmp/temp-manifests/* /etc/kubernetes/manifests/
    ```

    - Cleanup temporary files and folders
    ```bash
    # clean up all the temporary files
    umount /tmp/etcd-files
    rm -rf /tmp/temp-manifests /tmp/${ETCD_NAME}.etcd /tmp/etcd-files/ /var/lib/etcd/data/etcd-snapshot.db
    ```

    - Exit out of the Bottlerocket node
    ```bash
    # exit from bottlerocket's root shell
    exit

    # exit from bottlerocket node
    exit
    ```

    Repeat this step for each ETCD node.

1. Restart control plane core components

    - You can use this command to get the control plane node IPs which you can use to SSH
    ```bash
    echo -n ${CONTROL_PLANE_ENDPOINTS[@]} | xargs -I {} -d " " echo "{}"
    ```

    - SSH into the node and restart the core components. You must do this for each control plane node.
    ```bash
    # SSH into the control plane node using the IPs printed in previous command
    ssh -i ${SSH_KEY} ${SSH_USERNAME}@<Control Plane IP from previous command>

    # drop into bottlerocket root shell
    sudo sheltie

    # move the static manifests back to the right directory
    mv /tmp/temp-manifests/* /etc/kubernetes/manifests/
    ```

    - Exit out of the Bottlerocket node
    ```bash
    # exit from bottlerocket's root shell
    exit

    # exit from bottlerocket node
    exit
    ```
    Repeat these steps for each control plane node.

1. Unpause the cluster reconcilers

    Once the ETCD restore is complete, you can resume the cluster reconcilers.
    ```bash
    # unpause control plane reconcilation
    kubectl --kubeconfig=${MANAGEMENT_KUBECONFIG} -n eksa-system annotate machinehealthchecks ${CLUSTER_NAME}-kcp-unhealthy cluster.x-k8s.io/paused-

    # unpause ETCD reconcilation
    kubectl --kubeconfig=${MANAGEMENT_KUBECONFIG} -n eksa-system annotate etcdadmclusters ${CLUSTER_NAME}-etcd cluster.x-k8s.io/paused-
    ```

At this point you should have the ETCD cluster restored to snapshot.
To verify, you can run the following commands:
```bash
kubectl --kubeconfig=${CLUSTER_KUBECONFIG} get nodes

kubectl --kubeconfig=${CLUSTER_KUBECONFIG} get pods -A
```

You may also need to restart some deployments/daemonsets manually if they are stuck in an unhealthy state.
