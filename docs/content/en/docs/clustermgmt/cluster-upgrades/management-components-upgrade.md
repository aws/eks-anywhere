---
title: "Upgrade management components"
linkTitle: "Upgrade management components"
weight: 21
aliases:
    /docs/tasks/cluster/cluster-upgrades/management-components-upgrade/
date: 2024-02-21
description: >
  How to upgrade EKS Anywhere management components
---

{{% alert title="Note" color="warning" %}}

The `eksctl anywhere upgrade management-components` subcommand was added in EKS Anywhere version `v0.19.0` for all providers. Management component upgrades can only be done through the `eksctl` CLI, not through the Kubernetes API.

{{% /alert %}}

### What are management components?

Management components run on management or standalone clusters and are responsible for managing the lifecycle of workload clusters. Management components include but are not limited to:

* Cluster API controller
* EKS Anywhere cluster lifecycle controller
* Curated Packages controller
* Provider-specific controllers (vSphere, Tinkerbell etc.)
* Tinkerbell services (Boots, Hegel, Rufio, etc.)
* Custom Resource Definitions (CRDs) (clusters, eksareleases, etc.)

### Why upgrade management components separately?

The existing `eksctl anywhere upgrade cluster` command, when run against management or standalone clusters, upgrades both the management and cluster components. When upgrading versions, this upgrade process performs a rolling replacement of nodes in the cluster, which brings operational complexity, and should be carefully planned and executed.

With the new `eksctl anywhere upgrade management-components` command, you can upgrade management components separately from cluster components. This enables you to get the latest updates to the management components such as Cluster API controller, EKS Anywhere controller, and provider-specific controllers without a rolling replacement of nodes in the cluster, which reduces the operational complexity of the operation.

### Check management components versions

You can check the current and new versions of management components with the `eksctl anywhere upgrade plan management-components` command:

```
eksctl anywhere upgrade plan management-components -f management-cluster.yaml
```

The output should appear similar to the following:

```
NAME                 CURRENT VERSION       NEXT VERSION
EKS-A Management     v0.18.3+cc70180       v0.19.0+a672f31
cert-manager         v1.13.0+68bec33       v1.13.2+a34c207
cluster-api          v1.5.2+b14378d        v1.6.0+04c07bc
kubeadm              v1.5.2+5762149        v1.6.0+5bf0931
vsphere              v1.7.4+6ecf386        v1.8.5+650acfa
etcdadm-bootstrap    v1.0.10+c9a5a8a       v1.0.10+1ceb898
etcdadm-controller   v1.0.16+0ed68e6       v1.0.17+5e33062
```

Alternatively, you can run the `eksctl anywhere upgrade plan cluster` command against your management cluster, which shows the version differences for both management and cluster components.

### Upgrade management components

To perform the management components upgrade, run the following command:

```
eksctl anywhere upgrade management-components -f management-cluster.yaml
```

The output should appear similar to the following:

```
Performing setup and validations
✅ Docker provider validation
✅ Control plane ready
✅ Cluster CRDs ready
Upgrading core components
Installing new eksa components
🎉 Management components upgraded!
```

At this point, a new `eksarelease` custom resource will be available in your management cluster, which means new cluster components that correspond to your current EKS Anywhere version are available for cluster upgrades. You can subsequently run a workload cluster upgrade with the `eksctl anywhere upgrade cluster command`, or by updating `eksaVersion` field in your workload cluster's spec and applying it to your management cluster with Kubernetes API-compatible tooling such as kubectl, GitOps, or Terraform.