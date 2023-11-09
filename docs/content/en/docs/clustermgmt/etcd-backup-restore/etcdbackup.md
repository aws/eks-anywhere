---
title: "External etcd backup and restore"
linkTitle: "External etcd backup/restore"
weight: 10
description: >
  How to Backup and Restore External ETCD
---
{{% alert title="Note" color="warning" %}}
External ETCD topology is supported for vSphere, CloudStack and Snow clusters, but not yet for Bare Metal or Nutanix clusters.
{{% /alert %}}

This page contains steps for backing up a cluster by taking an ETCD snapshot, and restoring the cluster from a snapshot.

## Use case

EKS-Anywhere clusters use ETCD as the backing store. Taking a snapshot of ETCD backs up the entire cluster data. This can later be used to restore a cluster back to an earlier state if required. 

ETCD backups can be taken prior to cluster upgrade, so if the upgrade doesn't go as planned, you can restore from the backup.

{{% alert title="Important" color="warning" %}}

Restoring to a previous cluster state is a destructive and destablizing action to take on a running cluster. It should be considered only when all other options have been exhausted.

If you are able to retrieve data using the Kubernetes API server, then etcd is available and you should not restore using an etcd backup.

{{% /alert %}}
