---
title: "Overview"
linkTitle: "Overview"
weight: 5
description: >
  Overview of operating system management for nodes in EKS Anywhere clusters.
---

Bottlerocket, Ubuntu, and Red Hat Enterprise Linux (RHEL) can be used as operating systems for nodes in EKS Anywhere clusters. You can only use a single operating system per cluster. Bottlerocket is the only operating system distributed and fully supported by AWS. If you are using the other operating systems, you must build the operating system images and configure EKS Anywhere to use the images you built when installing or updating clusters. AWS will assist with troubleshooting and configuration guidance for Ubuntu and RHEL as part of EKS Anywhere Enterprise Subscriptions. For official support for Ubuntu and RHEL operating systems, you must purchase support through their respective vendors.

Reference the table below for the operating systems supported per deployment option for the latest version of EKS Anywhere. See [Admin machine]({{<  relref "/docs/getting-started/install" >}}) for supported operating systems.

|| vSphere | Bare metal | Snow | CloudStack | Nutanix |
| --- | :---: | :---: | :---: | :---: | :---: |
| Bottlerocket | &#10004; | &#10004; | &mdash; | &mdash; | &mdash; |
| Ubuntu | &#10004; | &#10004; | &#10004; | &mdash; | &#10004; |
| RHEL | &#10004; | &#10004; | &mdash; | &#10004; | &mdash; |

| OS | Supported Versions |
| :---: | :---: |
| Bottlerocket | 1.19.x |
| Ubuntu | 20.04.x, 22.04.x |
| RHEL | 8.x, 9.x<sup>*</sup> |

<sup>*</sup>Nutanix only

With the vSphere, bare metal, Snow, CloudStack and Nutanix deployment options, EKS Anywhere provisions the operating system when new machines are deployed during cluster creation, upgrade, and scaling operations. You can configure the operating system to use through the EKS Anywhere cluster spec, which varies by deployment option. See the deployment option sections below for an overview of how the operating system configuration works per deployment option.

## vSphere
To configure the operating system to use for EKS Anywhere clusters on vSphere, use the [`VSphereMachingConfig` `spec.template` field]({{< ref "/docs/getting-started/vsphere/vsphere-spec#template-optional" >}}). The template name corresponds to the template you imported into your vSphere environment. See the [Customize OVAs]({{< ref "/docs/getting-started/vsphere/customize/customize-ovas" >}}) and [Import OVAs]({{< ref "/docs/getting-started/vsphere/customize/vsphere-ovas" >}}) documentation pages for more information. Changing the template after cluster creation will result in the deployment of new machines.

## Bare metal
To configure the operating system to use for EKS Anywhere clusters on bare metal, use the [`TinkerbellDatacenterConfig` `spec.osImageURL` field]({{< ref "/docs/getting-started/baremetal/bare-spec#osimageurl" >}}). This field can be used to stream the operating system from a custom location and is required to use Ubuntu or RHEL. You cannot change the `osImageURL` after creating your cluster. To upgrade the operating system, you must replace the image at the existing `osImageURL` location with a new image. Operating system changes are only deployed when an action that triggers a deployment of new machines is triggered, which includes Kubernetes version upgrades only at this time.

## Snow
To configure the operating to use for EKS Anywhere clusters on Snow, use the [`SnowMachineConfig` `spec.osFamily` field]({{< ref "/docs/getting-started/snow/snow-spec#osfamily" >}}). At this time, only Ubuntu is supported for use with EKS Anywhere clusters on Snow. You can customize the instance image with the [`SnowMachineConfig` `spec.amiID` field]({{< ref "/docs/getting-started/snow/snow-spec#amiid-optional" >}}) and the instance type with the [`SnowMachineConfig` `spec.instanceType` field]({{< ref "/docs/getting-started/snow/snow-spec#instancetype-optional" >}}). Changes to these fields after cluster creation will result in the deployment of new machines.

## CloudStack
To configure the operating system to use for EKS Anywhere clusters on CloudStack, use the [`CloudStackMachineConfig` `spec.template.name` field]({{< ref "/docs/getting-started/cloudstack/cloud-spec#templateidname-required" >}}). At this time, only RHEL is supported for use with EKS Anywhere clusters on CloudStack. Changing the template name field after cluster creation will result in the deployment of new machines.

## Nutanix
To configure the operating system to use for EKS Anywhere clusters on Nutanix, use the [`NutanixMachineConfig` `spec.image.name` field]({{< ref "/docs/getting-started/nutanix/nutanix-spec#imagename-name-or-uuid-required" >}}) or the image uuid field. At this time, only Ubuntu is supported for use with EKS Anywhere clusters on Nutanix. Changing the image name or uuid field after cluster creation will result in the deployment of new machines.
