---
title: "EKS Anywhere curated packages"
linkTitle: "Curated packages"
weight: 60
date: 2022-05-09
description: >
  All information you may need for EKS Anywhere curated packages
---

{{% alert title="Note" color="primary" %}}
*The EKS Anywhere package controller and the EKS Anywhere Curated Packages (referred to as “features”) are provided as “preview features” subject to the AWS Service Terms (including Section 2 “Betas and Previews”) of the same. During the EKS Anywhere Curated Packages Public Preview, the AWS Service Terms are extended to provide customers access to these features free of charge. These features will be subject to a service charge and fee structure at ”General Availability“ of the features.*
{{% /alert %}}

### Overview
Amazon EKS Anywhere Curated Packages are Amazon-curated software packages that extend the core functionalities of Kubernetes on your EKS Anywhere clusters. If you operate EKS Anywhere clusters on-premises, you probably install additional software to ensure the security and reliability of your clusters. However, you may be spending a lot of effort researching for the right software, tracking updates, and testing them for compatibility. Now with the EKS Anywhere Curated Packages, you can rely on Amazon to provide trusted, up-to-date, and compatible software that are supported by Amazon, reducing the need for multiple vendor support agreements. 

* *Amazon-built*: All container images of the packages are built from source code by Amazon, including the open source (OSS) packages. OSS package images are built from the open source upstream.
* *Amazon-scanned*: Amazon continuously scans the container images including the OSS package images for security vulnerabilities and provides remediation. 
* *Amazon-signed*: Amazon signs the package bundle manifest (a Kubernetes manifest) for the list of curated packages. The manifest is signed with AWS Key Management Service (AWS KMS) managed private keys. The curated packages are installed and managed by a package controller on the clusters. Amazon provides validation of signatures through an admission control webhook in the package controller and the public keys distributed in the bundle manifest file. 
* *Amazon-tested*: Amazon tests the compatibility of all curated packages including the OSS packages with each new version of EKS Anywhere.
* *Amazon-supported*: All curated packages including the curated OSS packages are supported under the EKS Anywhere Support Subscription. 

The main components of EKS Anywhere Curated Packages are the [package controller]({{< relref "./package-controller" >}}), the [package build artifacts]({{< relref "./artifacts" >}}) and the [command line interface]({{< relref "./cli" >}}). The package controller will run in a pod in an EKS Anywhere cluster. The package controller will manage the lifecycle of all curated packages.

### Curated packages
Please check out [curated package list]({{< relref "../../reference/packagespec" >}}) for the complete list of EKS Anywhere curated packages.

### Workshop
Please check out [workshop]({{< relref "../../workshops/packages" >}}) for curated packages.


### FAQ
1. *Can I install software not from the curated package list?*

    Yes. You can install any optional software of your choice. Be aware you cannot use EKS Anywhere tooling to install or update your self-managed software. Amazon does not provide testing, security patching, software updates, or customer support for your self-managed software.


2. *Can I install software that’s on the curated package list but not sourced from EKS Anywhere repository?*

    If, for example, you deploy a Harbor image that is not built and signed by Amazon, Amazon will not provide testing or customer support to your self-built images.