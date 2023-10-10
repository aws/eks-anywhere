---
title: "Versioning"
linkTitle: "Versioning"
weight: 30
aliases:
    /docs/reference/support/support-versions/
description: >
  EKS Anywhere and Kubernetes version support policy
---

To see supported versions of Kubernetes for each release of EKS Anywhere, and information about that support, refer to the content below.

## Kubernetes support

Each EKS Anywhere version generally includes support for multiple Kubernetes versions, with the exception of the initial few releases.
Starting from EKS Anywhere version 0.11, the latest version supports at least four recent versions of Kubernetes.
The end of support date of a Kubernetes version aligns with Amazon EKS in AWS as documented on the [Amazon EKS Kubernetes release calendar](https://docs.aws.amazon.com/eks/latest/userguide/kubernetes-versions.html#kubernetes-release-calendar). 

Common vulnerabilities and exposures (CVE) patches and bug fixes, including those for the supported Kubernetes versions, are back-ported to the latest EKS Anywhere version (version n).
The following table shows EKS Anywhere version support for different Kubernetes versions:

| Kubernetes version | Supported EKS Anywhere version                        | First supported    | End of support   |
|--------------------|-------------------------------------------------------|--------------------|------------------|
| 1.27               | 0.17, 0.16                                                  | June 1, 2023       | September 2024   |
| 1.26               | 0.17, 0.16, 0.15                                            | April 5, 2023	     | July 2024        | 
| 1.25               | 0.17, 0.16, 0.15, 0.14                                      | February 16, 2023	 | April 2024       | 
| 1.24               | 0.17, 0.16, 0.15, 0.14, 0.13, 0.12                          | November 17, 2022	 | January 2024     | 
| 1.23               | 0.17, 0.16, 0.15, 0.14, 0.13, 0.12, 0.11                    | August 18, 2022	   | October 2023     | 
| 1.22               | 0.15, 0.14, 0.13, 0.12, 0.11, 0.10, 0.9, 0.8          | March 31, 2022     | May 2023         | 
| 1.21               | 0.14, 0.13, 0.12, 0.11, 0.10, 0.9, 0.8, 0.7, 0.6, 0.5 | September 8, 2021  | March 30, 2023   | 
| 1.20               | 0.12, 0.11, 0.10, 0.9, 0.8, 0.7, 0.6, 0.5             | September 8, 2021  | November 1, 2022 | 
| 1.19               | Not supported                                         | --                 | --               | 
| 1.18               | Not supported                                         | --                 | --               | 

The following table notes which EKS Anywhere and related Kubernetes versions are currently supported for CVE patches and bug fixes:

| EKS Anywhere version      | Kubernetes versions included | EKS Anywhere Release Date | CVE patches and bug fixes back-ported? |
|------------|------------------------------|---------------------------------|-------------------------|
| 0.17 | 1.27, 1.26, 1.25, 1.24, 1.23 | August 16, 2023 | Yes |
| 0.16 | 1.27, 1.26, 1.25, 1.24, 1.23 | June 1, 2023 | No |
| 0.15 | 1.26, 1.25, 1.24, 1.23, 1.22 | March 30, 2023 | No |
| 0.14 | 1.25, 1.24, 1.23, 1.22, 1.21 | January 19, 2023 | No |
| 0.13 | 1.24, 1.23, 1.22, 1.21       | December 15, 2022 | No |
| 0.12 | 1.24, 1.23, 1.22, 1.21, 1.20 | October 20, 2022 | No |
| 0.11 | 1.23, 1.22, 1.21, 1.20       | August 18, 2022 | No | 
| 0.10 | 1.22, 1.21, 1.20             | June 30, 2022 | No | 
| 0.9 | 1.21, 1.20                   | May 12, 2022 | No | 
| 0.8 | 1.22, 1.21, 1.20             | March 31, 2022 | No | 
| 0.7 | 1.21, 1.20                   | January 27, 2022 | No | 
| 0.6 | 1.21, 1.20                   | October 29, 2021 | No | 
| 0.5 | 1.21, 1.20                   | September 8, 2021 | No | 

* An [EKS Anywhere Enterprise Subscription](https://aws.amazon.com/eks/eks-anywhere/pricing/) is required to receive support for EKS Anywhere from AWS.

## EKS Anywhere versions and bundles

Each EKS Anywhere version installs a suite of componenets when provisioning your clusters. The versions of these components are specified in an EKS Anywhere versions bundle.

You may find a map of EKS Anywhere releases in the release manifest located in the [manifest.yaml](https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml) file. It contains references to the corresponding bundle manifest for each EKS Anywhere version.

Within the bundle manifest files, you will find the components that will be used to provision your cluster along with their versions. The images running in your cluster would use the same uri values specified in this bundles for that component.

For example, see the corresponding [bundle manifest](https://anywhere-assets.eks.amazonaws.com/releases/bundles/45/manifest.yaml) for EKS Anywhere v0.17.0.

## EKS Anywhere supported OS version

These are supported operating system versions that EKS Anywhere may use to provision control plane and worker nodes. For specifics on support OS for each provider, see the [configuration guide]({{< relref "../getting-started/" >}}) for your chosen provider.


| OS         | OS Versions                  | Supported EKS Anywhere version  |
|------------|------------------------------|---------------------------------|
| Ubuntu        | 22.04     | 0.17
|               | 20.04     | 0.17, 0.16, 0.15, 0.14, 0.13, 0.12, 0.11, 0.10, 0.9, 0.8, 0.7, 0.6, 0.5
| Bottlerocket  | 1.13.1    | 0.17, 0.16, 0.15
|               | 1.12.0    | 0.14
|               | 1.10.1    | 0.12
|               | 1.9.2     | 0.11
|               | 1.8.0     | 0.10
|               | 1.7.2     | 0.9
|               | 1.6.2     | 0.8
|               | 1.6.0     | 0.7
|               | 1.3.0     | 0.6
| RHEL          | 8.7	      | 0.17, 0.16, 0.15, 0.13, 0.14, 0.13, 0.12

For details on supported operating systems for Admin machines, [see the install guide.]({{< relref "../getting-started/install/" >}})

## EKS Anywhere version support FAQs

### What is the difference between an EKS Anywhere minor version versus a patch version?

An EKS Anywhere minor version includes new EKS Anywhere capabilities, bug fixes, security patches, and a new Kubernetes minor version if there is one. An EKS Anywhere patch version generally includes only bug fixes, security patches, and Kubernetes patch version. EKS Anywhere patch versions are released more frequently than the EKS Anywhere minor versions so you can receive the latest security and bug fixes sooner. 

### Where can I find the content of the EKS Anywhere versions?

You can find the content of the previous EKS Anywhere minor versions and patch versions on the [What’s New]({{< relref "../whatsnew/" >}}) page.

### Will I get notified when there is a new EKS Anywhere version release?

You will get notified if you have subscribed as documented on the Release Alerts page.

### Can I use different EKS Anywhere minor versions for my management cluster and workload clusters?

Yes, the management cluster can be upgraded to newer EKS Anywhere versions than its workload clusters.
However, we only support a maximum skew of one EKS Anywhere minor version for management and workload clusters.
This means that we support the management cluster being one EKS Anywhere minor version newer than the workload clusters (such as v0.15 for workload clusters if the management cluster is at v0.16).
In the event that you want to upgrade your management cluster to a version that does not satisfy this condition, we recommend upgrading the workload cluster's EKS Anywhere version first, followed by upgrading to your desired EKS Anywhere version for the management cluster.

> **NOTE**: Workload clusters can only be created on EKS Anywhere versions that have been used to create or upgrade the management cluster.
For example, if you create your management cluster with v0.15.0, you can only create workload clusters with v0.15.0.
However, if you create your management cluster with version v0.15.0 and then upgrade to v0.16.0, you can create workload clusters in either v0.15.0 or v0.16.0.

### Can I skip EKS Anywhere minor versions during cluster upgrade (such as going from v0.9 directly to v0.11)?

No. We perform regular upgrade reliability testing for sequential version upgrade (e.g. going from version 0.9 to 0.10, then from version 0.10 to 0.11), but we do not perform testing on non-sequential upgrade path (e.g. going from version 0.9 directly to 0.11). You should _not_ skip minor versions during cluster upgrade. However, you can choose to skip patch versions.

### What kind of fixes are back-ported to the previous versions?

Back-ported fixes include CVE patches and bug fixes for EKS Anywhere components and the Kubernetes versions that are supported by the corresponding EKS Anywhere version.

### What happens on the end of support date for a Kubernetes version?

On the end of support date, you can still create a new cluster with the unsupported Kubernetes version using an old version of the EKS Anywhere toolkit that was released with this Kubernetes version. Any existing EKS Anywhere clusters with the unsupported Kubernetes version will continue to function. However, you will not be able to receive CVE patches or bug fixes for the unsupported Kubernetes version.

### What EKS Anywhere versions are supported if you have the EKS Anywhere Enterprise Subscription?

If you have an [EKS Anywhere Enterprise Subscription](https://aws.amazon.com/eks/eks-anywhere/pricing/), AWS will provide troubleshooting support and configuration guidance for your licensed cluster, irrespective of the specific EKS Anywhere version it's running on. However, as the CVE patches and bug fixes are only back-ported to the latest EKS Anywhere version, it is highly recommended to keep your clusters updated with the latest EKS Anywhere release. With an EKS Anywhere Enterprise Subscription, AWS will assist you in upgrading your licensed cluster to the most recent EKS Anywhere version.

### Does EKS extended support for Kubernetes versions apply to EKS Anywhere clusters?

No. EKS extended support for Kubernetes versions does not apply to EKS Anywhere at this time. To request this capability, please comment or upvote on this [EKS Anywhere GitHub issue](https://github.com/aws/eks-anywhere/issues/6793).

### Will I get notified when support is ending for a Kubernetes version on EKS Anywhere?

Not automatically. You should check this page regularly and take note of the end of support date for the Kubernetes version you’re using.
