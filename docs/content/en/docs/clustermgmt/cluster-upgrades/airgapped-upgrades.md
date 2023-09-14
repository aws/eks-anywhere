---
title: "Upgrade airgapped cluster"
linkTitle: "Upgrade airgapped cluster"
weight: 20
aliases:
   /docs/tasks/cluster/cluster-upgrades/airgapped-upgrades/
description: >
  Upgrading EKS Anywhere clusters in airgapped environments
---
The procedure to upgrade EKS Anywhere clusters in airgapped environments is similar to the procedure for creating new clusters in airgapped environments. The only difference is that you must upgrade your `eksctl anywhere` CLI before running the steps to download and import the EKS Anywhere dependencies to your local registry mirror.

### Prerequisites
- An existing [Admin machine]({{< relref "../../getting-started/install" >}})
- **The upgraded version of the `eksctl anywhere` CLI installed on the Admin machine**
- Docker running on the Admin machine
- At least 80GB in storage space on the Admin machine to temporarily store the EKS Anywhere images locally before importing them to your local registry. Currently, when downloading images, EKS Anywhere pulls all dependencies for all infrastructure providers and supported Kubernetes versions.
- The download and import images commands must be run on an amd64 machine to import amd64 images to the registry mirror.

### Procedure

{{% content "../../getting-started/airgapped/airgap-steps.md" %}}

If the previous steps succeeded, all of the required EKS Anywhere dependencies are now present in your local registry. Before you upgrade your EKS Anywhere cluster, configure `registryMirrorConfiguration` in your EKS Anywhere cluster specification with the information for your local registry. For details see the [Registry Mirror Configuration documentation.]({{< relref "../../getting-started/optional/registrymirror/#registry-mirror-cluster-spec" >}})

>**_NOTE:_** If you are running EKS Anywhere on bare metal, you must configure `osImageURL` and `hookImagesURLPath` in your EKS Anywhere cluster specification with the location of the upgraded node operating system image and hook OS image. For details, reference the [bare metal configuration documentation.]({{< relref "../../getting-started/baremetal/bare-spec/#osimageurl" >}})

### Next Steps
- [Build upgraded node operating system images for your cluster]({{< relref "../../osmgmt/artifacts/#building-images-for-a-specific-eks-anywhere-version" >}})
- [Upgrade a cluster on vSphere, Snow, Cloudstack, or Nutanix]({{< relref "./vsphere-and-cloudstack-upgrades" >}})
- [Upgrade a cluster on bare metal]({{< relref "./baremetal-upgrades" >}})