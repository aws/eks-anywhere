---
title: "Ubuntu and RHEL"
linkTitle: "On Ubuntu and RHEL"
weight: 11
aliases:
    /docs/tasks/etcd-backup-restore/ubuntu-rhel-etcd-backup/
date: 2021-11-04
description: >
  How to backup and restore External ETCD on Ubuntu/RHEL OS EKS Anywhere cluster
---

{{% alert title="Note" color="warning" %}}
External etcd topology is supported for vSphere, CloudStack and Snow clusters, but not yet for Bare Metal or Nutanix clusters.
{{% /alert %}}

This page contains steps for backing up a cluster by taking an etcd snapshot, and restoring the cluster from a snapshot. These steps are for an EKS Anywhere cluster provisioned using the external etcd topology (selected by default) with Ubuntu OS.

### Use case

EKS-Anywhere clusters use etcd as the backing store. Taking a snapshot of etcd backs up the entire cluster data. This can later be used to restore a cluster back to an earlier state if required. Etcd backups can be taken prior to cluster upgrade, so if the upgrade doesn't go as planned you can restore from the backup.


### Backup

Etcd offers a built-in snapshot mechanism. You can take a snapshot using the `etcdctl snapshot save` or `etcdutl snapshot save` command by following the steps given below. 

1. Login to any one of the etcd VMs
```bash
ssh -i $PRIV_KEY ec2-user@$ETCD_VM_IP
```
2. Run the etcdctl or etcdutl command to take a snapshot with the following steps
    {{< tabpane >}}

    {{< tab header="Using etcd < v3.5.x" lang="bash" >}}
    sudo su
    source /etc/etcd/etcdctl.env
    etcdctl snapshot save snapshot.db
    chown ec2-user snapshot.db
    {{< /tab >}}

    {{< tab header="Using etcd >= v3.5.x" lang="bash" >}}
    sudo su
    etcdutl snapshot save snapshot.db
    chown ec2-user snapshot.db
    {{< /tab >}}
    
    {{< /tabpane >}}
```bash

```
3. Exit the VM. Copy the snapshot from the VM to your local/admin setup where you can save snapshots in a secure place. Before running scp, make sure you don't already have a snapshot file saved by the same name locally. 
```bash
scp -i $PRIV_KEY ec2-user@$ETCD_VM_IP:/home/ec2-user/snapshot.db . 
```

NOTE: This snapshot file contains all information stored in the cluster, so make sure you save it securely (encrypt it).

### Restore

Restoring etcd is a 2-part process. The first part is restoring etcd using the snapshot, creating a new data-dir for etcd. The second part is replacing the current etcd data-dir with the one generated after restore. During etcd data-dir replacement, we cannot have any kube-apiserver instances running in the cluster. So we will first stop all instances of kube-apiserver and other controlplane components using the following steps for every controlplane VM:

#### Pausing Etcdadm cluster and control plane machine health check reconcile

During restore, it is required to pause the Etcdadm controller reconcile and the control plane machine healths checks for the target cluster (whether it is management or workload cluster). To do that, you need to add a `cluster.x-k8s.io/paused` annotation to the target cluster's `etcdadmclusters` and `machinehealthchecks` resources. For example,

```bash
kubectl annotate clusters.anywhere.eks.amazonaws.com $CLUSTER_NAME anywhere.eks.amazonaws.com/paused=true --kubeconfig mgmt-cluster.kubeconfig

kubectl patch clusters.cluster.x-k8s.io $CLUSTER_NAME --type merge -p '{"spec":{"paused": true}}' -n eksa-system --kubeconfig mgmt-cluster.kubeconfig
```

#### Stopping the controlplane components
1. Login to a controlplane VM
```bash
ssh -i $PRIV_KEY ec2-user@$CONTROLPLANE_VM_IP
```
2. Stop controlplane components by moving the static pod manifests under a temp directory:
```bash
sudo su
mkdir temp-manifests
mv /etc/kubernetes/manifests/*.yaml temp-manifests
```
3. Repeat these steps for all other controlplane VMs

After this you can restore etcd from a saved snapshot using the `snapshot save` command following the steps given below.

#### Restoring from the snapshot

1. The snapshot file should be made available in every etcd VM of the cluster. You can copy it to each etcd VM using this command:
```bash
scp -i $PRIV_KEY snapshot.db ec2-user@$ETCD_VM_IP:/home/ec2-user
```
2. To run the etcdctl or etcdutl snapshot restore command, you need to provide the following configuration parameters:
* name: This is the name of the etcd member. The value of this parameter should match the value used while starting the member. This can be obtained by running:
```bash
export ETCD_NAME=$(cat /etc/etcd/etcd.env | grep ETCD_NAME | awk -F'=' '{print $2}')
```  
* initial-advertise-peer-urls: This is the advertise peer URL with which this etcd member was configured. It should be the exact value with which this etcd member was started. This can be obtained by running:
```bash
export ETCD_INITIAL_ADVERTISE_PEER_URLS=$(cat /etc/etcd/etcd.env | grep ETCD_INITIAL_ADVERTISE_PEER_URLS | awk -F'=' '{print $2}')
```
* initial-cluster: This should be a comma-separated mapping of etcd member name and its peer URL. For this, get the `ETCD_NAME` and `ETCD_INITIAL_ADVERTISE_PEER_URLS` values for each member and join them. And then use this exact value for all etcd VMs. For example, for a 3 member etcd cluster this is what the value would look like (The command below cannot be run directly without substituting the required variables and is meant to be an example)
```bash
export ETCD_INITIAL_CLUSTER=${ETCD_NAME_1}=${ETCD_INITIAL_ADVERTISE_PEER_URLS_1},${ETCD_NAME_2}=${ETCD_INITIAL_ADVERTISE_PEER_URLS_2},${ETCD_NAME_3}=${ETCD_INITIAL_ADVERTISE_PEER_URLS_3}
```  
* initial-cluster-token: Set this to a unique value and use the same value for all etcd members of the cluster. It can be any value such as `etcd-cluster-1` as long as it hasn't been used before.  
3. Gather the required env vars for the restore command
```bash
cat <<EOF >> restore.env
export ETCD_NAME=$(cat /etc/etcd/etcd.env | grep ETCD_NAME | awk -F'=' '{print $2}')
export ETCD_INITIAL_ADVERTISE_PEER_URLS=$(cat /etc/etcd/etcd.env | grep ETCD_INITIAL_ADVERTISE_PEER_URLS | awk -F'=' '{print $2}')
EOF

cat /etc/etcd/etcdctl.env >> restore.env
```
4. Make sure you form the correct `ETCD_INITIAL_CLUSTER` value using all etcd members, and set it as an env var in the restore.env file created in the above step.
5. Once you have obtained all the right values, run the following commands to restore etcd replacing the required values:
    {{< tabpane >}}

    {{< tab header="Using etcd < v3.5.x" lang="bash" >}}
    sudo su
    source restore.env
    etcdctl snapshot restore snapshot.db \
        --name=${ETCD_NAME} \
        --initial-cluster=${ETCD_INITIAL_CLUSTER} \
        --initial-cluster-token=etcd-cluster-1 \
        --initial-advertise-peer-urls=${ETCD_INITIAL_ADVERTISE_PEER_URLS}
    {{< /tab >}}

    {{< tab header="Using etcd >= v3.5.x" lang="bash" >}}
    sudo su
    source restore.env
    etcdutl snapshot restore snapshot.db \
        --name=${ETCD_NAME} \
        --initial-cluster=${ETCD_INITIAL_CLUSTER} \
        --initial-cluster-token=etcd-cluster-1 \
        --initial-advertise-peer-urls=${ETCD_INITIAL_ADVERTISE_PEER_URLS}
    {{< /tab >}}
    
    {{< /tabpane >}}

5. This is going to create a new data-dir for the restored contents under a new directory `{ETCD_NAME}.etcd`. To start using this, restart etcd with the new data-dir with the following steps:
```bash
systemctl stop etcd.service
mv /var/lib/etcd/member /var/lib/etcd/member.bak
mv ${ETCD_NAME}.etcd/member /var/lib/etcd/
```
6. Perform this directory swap on all etcd VMs, and then start etcd again on those VMs
```bash
systemctl start etcd.service
```
NOTE: Until the etcd process is started on all VMs, it might appear stuck on the VMs where it was started first, but this should be temporary.

#### Starting the controlplane components
1. Login to a controlplane VM
```bash
ssh -i $PRIV_KEY ec2-user@$CONTROLPLANE_VM_IP
```
2. Start the controlplane components by moving back the static pod manifests from under the temp directory to the /etc/kubernetes/manifests directory:
```bash
mv temp-manifests/*.yaml /etc/kubernetes/manifests
```
3. Repeat these steps for all other controlplane VMs
4. It may take a few minutes for the kube-apiserver and the other components to get restarted. After this you should be able to access all objects present in the cluster at the time the backup was taken.

#### Resuming Etcdadm cluster and control plane machine health checks reconcile

Resume Etcdadm cluster reconcile and control plane machine health checks for the target cluster by removing the `cluster.x-k8s.io/paused` annotation in the target cluster's  resource. For example,

```bash
kubectl annotate clusters.anywhere.eks.amazonaws.com $CLUSTER_NAME anywhere.eks.amazonaws.com/paused- --kubeconfig mgmt-cluster.kubeconfig

kubectl patch clusters.cluster.x-k8s.io $CLUSTER_NAME --type merge -p '{"spec":{"paused": false}}' -n eksa-system --kubeconfig mgmt-cluster.kubeconfig
```
