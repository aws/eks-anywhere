---
title: 2. Airgapped (optional)
weight: 15
description: >
  Configuring EKS Anywhere for airgapped environments
---

When creating an EKS Anywhere cluster, there may be times where you need to do so in an airgapped
environment.
In this type of environment, cluster nodes are connected to the Admin Machine, but not to the
internet.
In order to download images and artifacts, however, the Admin machine needs to be temporarily
connected to the internet.

An airgapped environment is especially important if you require the most secure networks.
EKS Anywhere supports airgapped installation for creating clusters using a registry mirror.
For airgapped installation to work, the Admin machine must have:

* Temporary access to the internet to download images and artifacts
* Ample space (80 GB or more) to store artifacts locally


To create a cluster in an airgapped environment, perform the following:

1. Download the artifacts and images that will be used by the cluster nodes to the Admin machine using the following command:
   ```bash
   eksctl anywhere download artifacts
   ```
   A compressed file `eks-anywhere-downloads.tar.gz` will be downloaded.

1. To decompress this file, use the following command:
   ```bash
   tar -xvf eks-anywhere-downloads.tar.gz
   ```
   This will create an eks-anywhere-downloads folder that we’ll be using later.

1. In order for the next command to run smoothly, ensure that Docker has been pre-installed and is running. Then run the following:
   ```bash
   eksctl anywhere download images -o images.tar
   ```

1. If you want to use curated packages, refer to [Curated Packages]({{< relref "../../packages/prereq#prepare-curated-packages-for-airapped-clusters" >}}) to copy curated packages to your registry mirror.

{{% alert title="Warning" color="warning" %}}
`eksctl anywhere download images` and `eksctl anywhere import images` command need to be run on an amd64 machine to import amd64 images to the registry mirror.
{{% /alert %}}

   **For the remaining steps, the Admin machine no longer needs to be connected to the internet or the bastion host.**

1. Next, you will need to set up a local registry mirror to host the downloaded EKS Anywhere images. In order to set one up, refer to [Registry Mirror configuration.]({{< relref "../../getting-started/optional/registrymirror.md" >}})

1. Now that you’ve configured your local registry mirror, you will need to import images to the local registry mirror using the following command (be sure to replace <registryUrl> with the url of the local registry mirror you created in step 4):
   ```bash
   eksctl anywhere import images -i images.tar -r <registryUrl> \
      --bundles ./eks-anywhere-downloads/bundle-release.yaml
   ```
You are now ready to deploy a cluster by selecting your provider from the [EKS Anywhere providers]({{< relref "/docs/getting-started/chooseprovider" >}}) page and following those instructions.

### For Bare Metal (Tinkerbell)
You will need to have hookOS and its OS artifacts downloaded and served locally from an HTTP file server.
You will also need to modify the [hookImagesURLPath]({{< relref "../baremetal/bare-spec/#hookimagesurlpath" >}}) and the [osImageURL]({{< relref "../baremetal/bare-spec/#osimageurl" >}}) in the cluster configuration files.
Ensure that structure of the files is set up as described in [hookImagesURLPath.]({{< relref "../baremetal/bare-spec/#hookimagesurlpath" >}})

### For vSphere
If you are using the vSphere provider, be sure that the requirements in the
[Prerequisite checklist]({{< relref "../vsphere/vsphere-prereq/" >}}) have been met.

## Deploy a cluster

Once you have the tools installed you can deploy a cluster by [choosing a provider]({{< relref "/docs/getting-started/chooseprovider/" >}})

## Upgrade an airgapped cluster

To upgrade an airgapped cluster, see [upgrade airgapped cluster]({{< relref "../../clustermgmt/cluster-upgrades/airgapped-upgrades.md" >}})
