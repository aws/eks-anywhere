---
title: 2. Airgapped (optional)
weight: 15
description: >
  Configure EKS Anywhere for airgapped environments
---

EKS Anywhere can be used in airgapped environments, where clusters are not connected to the internet or external networks.
The following diagrams illustrate how to set up for cluster creation in an airgapped environment:

![Download EKS Anywhere artifacts to Admin machine](/images/airgap-arch01.png)

If you are planning to run EKS Anywhere in an airgapped environments, before you create a cluster, you must temporarily connect your Admin machine to the internet to install the `eksctl` CLI and pull the required EKS Anywhere dependencies.

![Disconnect Admin machine from Internet to create cluster](/images/airgap-arch02.png)

Once these dependencies are downloaded and imported in a local registry, you no longer need internet access. In the EKS Anywhere cluster specification, you can configure EKS Anywhere to use your local registry mirror. When the registry mirror configuration is set in the EKS Anywhere cluster specification, EKS Anywhere configures containerd to pull from that registry instead of Amazon ECR during cluster creation and lifecycle operations. For more information, reference the [Registry Mirror Configuration documentation.]({{< relref "../optional/registrymirror" >}})

If you are using Ubuntu or RHEL as the operating system for nodes in your EKS Anywhere cluster, you must connect to the internet while building the images with the EKS Anywhere image-builder tool. After building the operating system images, you can configure EKS Anywhere to pull the operating system images from a location of your chosing in the EKS Anywhere cluster specification. For more information on the image building process and operating system cluster specification, reference the [Operating System Management documentation.]({{< relref "../../osmgmt/overview" >}})

### Overview

The process for preparing your airgapped environment for EKS Anywhere is summarized by the following steps:
1. Use the `eksctl anywhere` CLI to download EKS Anywhere artifacts. These artifacts are `yaml` files that contain the list and locations of the EKS Anywhere dependencies.
1. Use the `eksctl anywhere` CLI to download EKS Anywhere images. These images include EKS Anywhere dependencies including EKS Distro components, Cluster API provider components, and EKS Anywhere components such as the EKS Anywhere controllers, Cilium CNI, kube-vip, and cert-manager.
1. Set up your local registry following the steps in the [Registry Mirror Configuration documentation.]({{< relref "../optional/registrymirror" >}})
1. Use the `eksctl anywhere` CLI to import the EKS Anywhere images to your local registry.
1. Optionally use the `eksctl anywhere` CLI to copy EKS Anywhere Curated Packages images to your local registry.

### Prerequisites
- An existing [Admin machine]({{< relref "../install" >}})
- Docker running on the Admin machine
- At least 80GB in storage space on the Admin machine to temporarily store the EKS Anywhere images locally before importing them to your local registry. Currently, when downloading images, EKS Anywhere pulls all dependencies for all infrastructure providers and supported Kubernetes versions.
- The download and import images commands must be run on an amd64 machine to import amd64 images to the registry mirror.

### Procedure

{{% content "./airgap-steps.md" %}}

If the previous steps succeeded, all of the required EKS Anywhere dependencies are now present in your local registry. Before you create your EKS Anywhere cluster, configure `registryMirrorConfiguration` in your EKS Anywhere cluster specification with the information for your local registry. For details see the [Registry Mirror Configuration documentation.]({{< relref "../../getting-started/optional/registrymirror/#registry-mirror-cluster-spec" >}})

>**_NOTE:_** If you are running EKS Anywhere on bare metal, you must configure `osImageURL` and `hookImagesURLPath` in your EKS Anywhere cluster specification with the location of your node operating system image and the hook OS image. For details, reference the [bare metal configuration documentation.]({{< relref "../baremetal/bare-spec/#osimageurl" >}})

### Next Steps
- Review EKS Anywhere [cluster networking requirements]({{< relref "../ports" >}})
- Review EKS Anywhere [infrastructure providers and their prerequisites]({{< relref "../chooseprovider" >}})
- Review the [upgrade procedure]({{< relref "../../clustermgmt/cluster-upgrades/airgapped-upgrades.md" >}}) for EKS Anywhere in airgapped environments
