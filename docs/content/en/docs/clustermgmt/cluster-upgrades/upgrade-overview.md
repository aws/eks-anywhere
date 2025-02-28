---
title: "Upgrade Overview"
linkTitle: "Overview"
weight: 10
date: 2024-02-23
description: >
  Overview of EKS Anywhere and Kubernetes version upgrades
---

Version upgrades in EKS Anywhere and Kubernetes are events that should be carefully planned, tested, and implemented. New EKS Anywhere and Kubernetes versions can introduce significant changes, and we recommend that you test the behavior of your applications against new EKS Anywhere and Kubernetes versions before you update your production clusters. Cluster [backups]({{< relref "../cluster-backup-restore/backup-cluster" >}}) should always be performed before initiating an upgrade. When initiating cluster version upgrades, new virtual or bare metal machines are provisioned and the machines on older versions are deprovisioned in a rolling fashion by default. 

Unlike Amazon EKS, there are no automatic upgrades in EKS Anywhere and you have full control over when you upgrade. On the end of support date, you can still create new EKS Anywhere clusters with the unsupported Kubernetes version if the EKS Anywhere version you are using includes it. Any existing EKS Anywhere clusters with the unsupported Kubernetes version continue to function. As new Kubernetes versions become available in EKS Anywhere, we recommend that you proactively update your clusters to use the latest available Kubernetes version to remain on versions that receive CVE patches and bug fixes. 

Reference the [EKS Anywhere Changelog]({{< relref "../../whatsnew/changelog/" >}}) for information on fixes, features, and changes included in each EKS Anywhere release. For details on the EKS Anywhere version support policy, reference the [Versioning page.]({{< relref "../../concepts/support-versions" >}})

### Upgrade Version Skew

{{% content "version-skew.md" %}}

### User Interfaces

EKS Anywhere versions for management and standalone clusters must be upgraded with the `eksctl anywhere` CLI. Kubernetes versions for management, standalone, and workload clusters, and EKS Anywhere versions for workload clusters can be upgraded with the `eksctl anywhere` CLI or with Kubernetes API-compatible clients such as the `kubectl` CLI, GitOps, or Terraform. For an overview of the differences between management, standalone, workload clusters, reference the [Architecture page.]({{< relref "../../concepts/architecture" >}})

If you are using the `eksctl anywhere` CLI, there are `eksctl anywhere upgrade plan cluster` and `eksctl anywhere upgrade cluster` commands. The former shows the components and versions that will be upgraded. The latter runs the upgrade, first validating a set of preflight checks and then upgrading your cluster to match the updated spec.

If you are using an Kubernetes API-compatible client, you modify your workload cluster spec yaml and apply the modified yaml to your management cluster. The EKS Anywhere lifecycle controller, which runs on the management cluster, reconciles the desired changes on the workload cluster.

As of EKS Anywhere version `v0.19.0`, management components can be upgraded separately from cluster components. This is enables you to get the latest updates to the management components such as Cluster API controller, EKS Anywhere controller, and provider-specific controllers without impact to your workload clusters. Management components can only be upgraded with the `eksctl anywhere` CLI, which has new `eksctl anywhere upgrade plan management-components` and `eksctl anywhere upgrade management-component` commands. For more information, reference the [Upgrade Management Components page.]({{< relref "./management-components-upgrade" >}})

### Upgrading EKS Anywhere Versions

Each EKS Anywhere version includes all components required to create and manage EKS Anywhere clusters. For example, this includes:

- Administrative / CLI components (`eksctl anywhere` CLI, image-builder, diagnostics-collector)
- Management components (Cluster API controller, EKS Anywhere controller, provider-specific controllers)
- Cluster components (Kubernetes, Cilium)

You can find details about each EKS Anywhere releases in the EKS Anywhere release manifest. The release manifest contains references to the corresponding bundle manifest for each EKS Anywhere version. Within the bundle manifest, you will find the components included in a specific EKS Anywhere version. The images running in your deployment use the same URI values specified in the bundle manifest for that component. For example, see the [bundle manifest](https://anywhere-assets.eks.amazonaws.com/releases/bundles/92/manifest.yaml) for EKS Anywhere version `v0.22.0`.

To upgrade the EKS Anywhere version of a management or standalone cluster, you install a new version of the `eksctl anywhere` CLI, change the `eksaVersion` field in your management or standalone cluster's spec yaml, and then run the `eksctl anywhere upgrade management-components -f cluster.yaml` (as of EKS Anywhere version v0.19) or `eksctl anywhere upgrade cluster -f cluster.yaml` command. The `eksctl anywhere upgrade cluster` command upgrades both management and cluster components.

To upgrade the EKS Anywhere version of a workload cluster, you change the `eksaVersion` field in your workload cluster's spec yaml, and apply the new workload cluster's spec yaml to your management cluster using the `eksctl anywhere` CLI or with Kubernetes API-compatible clients.

### Upgrading Kubernetes Versions

Each EKS Anywhere version supports at least 4 minor versions of Kubernetes. Kubernetes patch version increments are included in EKS Anywhere minor and patch releases. There are two places in the cluster spec where you can configure the Kubernetes version, `Cluster.Spec.KubernetesVersion` and `Cluster.Spec.WorkerNodeGroupConfiguration[].KubernetesVersion`. If only `Cluster.Spec.KubernetesVersion` is set, then that version will apply to both control plane and worker nodes. You can use `Cluster.Spec.WorkerNodeGroupConfiguration[].KubernetesVersion` to upgrade your worker nodes separately from control plane nodes. 

The `Cluster.Spec.WorkerNodeGroupConfiguration[].KubernetesVersion` cannot be greater than `Cluster.Spec.KubernetesVersion`. In Kubernetes versions lower than `v1.28.0`, the `Cluster.Spec.WorkerNodeGroupConfiguration[].KubernetesVersion` can be at most 2 versions lower than the `Cluster.Spec.KubernetesVersion`. In Kubernetes versions `v1.28.0` or greater, the `Cluster.Spec.WorkerNodeGroupConfiguration[].KubernetesVersion` can be at most 3 versions lower than the `Cluster.Spec.KubernetesVersion`.

### Upgrade Controls

By default, when you upgrade EKS Anywhere or Kubernetes versions, nodes are upgraded one at a time in a rolling fashion. All control plane nodes are upgraded before worker nodes. To control the speed and behavior of rolling upgrades, you can use the `upgradeRolloutStrategy.rollingUpdate.maxSurge` and `upgradeRolloutStrategy.rollingUpdate.maxUnavailable` fields in the cluster spec (available on all providers as of EKS Anywhere version v0.19). The `maxSurge` setting controls how many new machines can be queued for provisioning simultaneously, and the `maxUnavailable` setting controls how many machines must remain available during upgrades. For more information on these controls, reference [Advanced configuration]({{< relref "./vsphere-and-cloudstack-upgrades#advanced-configuration-for-rolling-upgrade" >}}) for vSphere, CloudStack, Nutanix, and Snow upgrades and [Advanced configuration]({{< relref "./baremetal-upgrades#advanced-configuration-for-upgrade-rollout-strategy" >}}) for bare metal upgrades.

As of EKS Anywhere version `v0.19.0`, if you are running EKS Anywhere on bare metal, you can use the in-place rollout strategy to upgrade EKS Anywhere and Kubernetes versions, which upgrades the components on the same physical machines without requiring additional server capacity. In-place upgrades are not available for other providers.
