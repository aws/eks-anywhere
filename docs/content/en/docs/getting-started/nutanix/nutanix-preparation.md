---
title: "Preparing Nutanix Cloud Infrastructure for EKS Anywhere"
linkTitle: "2. Prepare Nutanix"
weight: 20
aliases:
    /docs/reference/nutanix/nutanix-preparation/
description: >
  Set up a Nutanix cluster to prepare it for EKS Anywhere
---

Certain resources must be in place with appropriate user permissions to create an EKS Anywhere cluster using the Nutanix provider.

## Configuring Nutanix User
You need a Prism Admin user to create EKS Anywhere clusters on top of your Nutanix cluster.

## Build Nutanix AHV node images
Follow the steps outlined in [artifacts]({{< relref "../../osmgmt/artifacts/" >}}) to create a Ubuntu-based image for Nutanix AHV and import it into the AOS Image Service.

