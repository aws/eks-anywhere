---
title: "Upgrade airgapped cluster"
linkTitle: "Upgrade airgapped cluster"
weight: 20
aliases:
   /docs/tasks/cluster/cluster-upgrades/airgapped-upgrades/
description: >
  How to perform eks-anywhere upgrade for an airgapped cluster
---
If you want to upgrade EKS Anywhere version, or your cluster upgrade requires EKS Anywhere version upgrade in airgapped environment, perform the following steps to prepare new artifacts in your registry mirror:

1. [Upgrade EKS Anywhere version]({{< relref "./vsphere-and-cloudstack-upgrades.md#eks-anywhere-version-upgrades" >}}).

1. Use the upgraded binary to download new artifacts that will be used by the cluster nodes to the Admin machine:
   ```bash
   eksctl anywhere download artifacts
   ```
   A compressed file `eks-anywhere-downloads.tar.gz` will be downloaded.

1. Decompress this file:
   ```bash
   tar -xvf eks-anywhere-downloads.tar.gz
   ```
   This will create an eks-anywhere-downloads folder that we’ll be using later.

1. Use the upgraded binary to download new images:
   ```bash
   eksctl anywhere download images -o images.tar
   ```

1. Use the upgraded binary to import new images to your local registry mirror.
   ```bash
   eksctl anywhere import images -i images.tar -r <registryUrl> \
      --bundles ./eks-anywhere-downloads/bundle-release.yaml
   ```
1. You are now ready to [upgrade your cluster based on the cluster provider]({{< relref "../cluster-upgrades/" >}})