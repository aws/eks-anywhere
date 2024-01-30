---
title: "Backup cluster"
linkTitle: "Backup cluster"
weight: 20
aliases:
    /docs/tasks/cluster/cluster-backup-restore/backup-cluster/
description: >
  How to backup your EKS Anywhere cluster
---

We strongly advise performing regular cluster backups of all the EKS Anywhere clusters. This ensures that you always have an up-to-date cluster state available for restoration in case the cluster experiences issues or becomes unrecoverable. This document outlines the steps for creating the two essential types of backups required for the [EKS Anywhere cluster restore process]({{< relref "./restore-cluster" >}}).

## Etcd backup

For optimal cluster maintenance, it is crucial to perform regular etcd backups on all your EKS Anywhere management and workload clusters. **Always** take an etcd backup before performing an upgrade so it can be used to restore the cluster to a previous state in the event of a cluster upgrade failure. To create an etcd backup for your cluster, follow the guidelines provided in the [External etcd backup and restore]({{< relref "../etcd-backup-restore/etcdbackup" >}}) section.


## Cluster API backup

Since cluster failures primarily occur following unsuccessful cluster upgrades, EKS Anywhere takes the proactive step of automatically creating backups for the Cluster API objects. For the management cluster, it captures the states of both the management cluster and its workload clusters if all the clusters are in ready state. If one of the workload clusters is not ready, EKS Anywhere takes the best effort to backup the management cluster itself. For the workload cluster, it captures the state workload cluster's Cluster API objects. These backups are stored within the management cluster folder, where the upgrade command is initiated from the Admin machine, and are generated before each management and/or workload cluster upgrade process. 
For example, after executing a cluster upgrade command on `mgmt-cluster`, a backup folder is generated with the naming convention of `mgmt-cluster-backup-${timestamp}`:

```bash
mgmt-cluster/ 
├── mgmt-cluster-backup-2023-10-11T02_55_56 <------ Folder with a backup of the CAPI objects 
├── mgmt-cluster-eks-a-cluster.kubeconfig
├── mgmt-cluster-eks-a-cluster.yaml
└── generated
```

For workload cluster, a backup folder is generated with the naming convention of `wkld-cluster-backup-${timestamp}` under `mgmt-cluster` directory

```bash
mgmt-cluster/ 
├── wkld-cluster-backup-2023-10-11T02_55_56 <------ Folder with a backup of the CAPI objects 
├── mgmt-cluster-eks-a-cluster.kubeconfig
├── mgmt-cluster-eks-a-cluster.yaml
└── generated
```

Although the likelihood of a cluster failure occurring without any associated cluster upgrade operation is relatively low, it is still recommended to manually back up these Cluster API objects on a routine basis. For example, to create a Cluster API backup of a cluster:


```bash
MGMT_CLUSTER="mgmt"
MGMT_CLUSTER_KUBECONFIG=${MGMT_CLUSTER}/${MGMT_CLUSTER}-eks-a-cluster.kubeconfig
BACKUP_DIRECTORY=backup-mgmt

# Substitute the EKS Anywhere release version with whatever CLI version you are using
EKSA_RELEASE_VERSION=v0.17.3
BUNDLE_MANIFEST_URL=$(curl -s https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
CLI_TOOLS_IMAGE=$(curl -s $BUNDLE_MANIFEST_URL | yq ".spec.versionsBundles[0].eksa.cliTools.uri")


docker run -i --network host -w $(pwd) -v $(pwd):/$(pwd) --entrypoint clusterctl ${CLI_TOOLS_IMAGE} move \
        --namespace eksa-system \
        --kubeconfig $MGMT_CLUSTER_KUBECONFIG \
        --to-directory ${BACKUP_DIRECTORY}
```

This saves the Cluster API objects of the management cluster `mgmt` with all its workload clusters, to a local directory under the `backup-mgmt` folder.
