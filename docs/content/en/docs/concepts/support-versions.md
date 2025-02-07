---
title: "EKS Anywhere and Kubernetes version lifecycle"
linkTitle: "Version lifecycle"
weight: 30
aliases:
    /docs/reference/support/support-versions/
description: >
  EKS Anywhere and Kubernetes version support policy and release cycle
---

This page contains information on the EKS Anywhere release cycle and support for Kubernetes versions.

Each EKS Anywhere version contains at least four Kubernetes versions. When creating new clusters, we recommend that you use the latest available Kubernetes version supported by EKS Anywhere. However, you can create new EKS Anywhere clusters with any Kubernetes version that the EKS Anywhere version supports. To create new clusters or upgrade existing clusters with a Kubernetes version under extended support, you must have a valid and active license token for the cluster you are creating or upgrading.

## EKS Anywhere versions

Each EKS Anywhere version includes all components required to create and manage EKS Anywhere clusters. This includes but is not limited to:

- Administrative / CLI components (eksctl CLI, image-builder, diagnostics-collector)
- Management components (Cluster API controller, EKS Anywhere controller, provider-specific controllers)
- Cluster components (Kubernetes, Cilium)

You can find details about each EKS Anywhere release in the [EKS Anywhere release manifest](https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml). The release manifest contains references to the corresponding bundle manifest for each EKS Anywhere version. Within the bundle manifest, you will find the components included in a specific EKS Anywhere version. The images running in your deployment use the same URI values specified in the bundle manifest for that component. For example, see the [bundle manifest](https://anywhere-assets.eks.amazonaws.com/releases/bundles/89/manifest.yaml) for EKS Anywhere version `v0.21.5`.

EKS Anywhere follows a 4-month release cadence for minor versions and a 2-week cadence for patch versions. Common vulnerabilities and exposures (CVE) patches and bug fixes, including those for the supported Kubernetes versions, are included in the latest EKS Anywhere minor version (version N). High and critical CVE fixes and bug fixes are also backported to the penultimate EKS Anywhere minor version (version N-1), which follows a monthly patch release cadence.

Reference the table below for release dates and patch support for each EKS Anywhere version. This table shows the Kubernetes versions that are supported in each EKS Anywhere version.
<!-- See /docs/data/version_support.yml -->
{{< eksa_support >}}

* Older EKS Anywhere versions are omitted from this table for brevity, reference the [EKS Anywhere GitHub](https://github.com/aws/eks-anywhere/tree/main/docs/data/version_support.yml) for older versions.

## Kubernetes versions

The release and support schedule for Kubernetes versions in EKS Anywhere aligns with the Amazon EKS release and support schedule, including both standard and extended support Kubernetes versions, as documented on the [Amazon EKS Kubernetes release calendar.](https://docs.aws.amazon.com/eks/latest/userguide/kubernetes-versions.html#kubernetes-release-calendar) 

A Kubernetes version is under standard support in EKS Anywhere for 14 months after it's released in EKS Anywhere. Extended support for Kubernetes versions was released in EKS Anywhere version v0.22 and adds an additional 12 months of CVE and critical bug fix patches for the Kubernetes version. To use Kubernetes versions under extended support with EKS Anywhere clusters, you must have a valid and active license for each cluster you will create or upgrade with extended support Kubernetes versions. The patch releases for Kubernetes versions are included in EKS Anywhere according to the EKS Anywhere version release schedule defined in the previous section.

Unlike Amazon EKS in the cloud, there are no automatic upgrades in EKS Anywhere and you have full control over when you upgrade. On the end of the extended support date, you can still create new EKS Anywhere clusters with that Kubernetes version if the EKS Anywhere version includes it. Any existing EKS Anywhere clusters with the unsupported Kubernetes version continue to function. As new Kubernetes versions become available in EKS Anywhere, we recommend that you proactively update your clusters to use the latest available Kubernetes version to remain on versions that receive CVE patches and bug fixes. You can continue to get troubleshooting support for your EKS Anywhere clusters running on unsupported Kuberentes versions if you have an EKS Anywhere Enterprise Subscription. However, if there is an issue that requires code changes, the only resolution is to upgrade as patches are no longer made available for Kubernetes versions that have exited extended support.

Reference the table below for release and support dates for each Kubernetes version in EKS Anywhere. The Release Date column denotes the EKS Anywhere release date when the Kubernetes version was first supported in EKS Anywhere.

<!-- See /docs/data/version_support.yml -->
{{< kube_support >}}

* Older Kubernetes versions are omitted from this table for brevity, reference the [EKS Anywhere GitHub](https://github.com/aws/eks-anywhere/tree/main/docs/data/version_support.yml) for older versions.

## Operating System versions

Bottlerocket, Ubuntu, and Red Hat Enterprise Linux (RHEL) can be used as operating systems for nodes in EKS Anywhere clusters. Reference the table below for operating system version support in EKS Anywhere. For information on operating system management in EKS Anywhere, reference the [Operating System Management Overview page]({{< relref "../osmgmt/overview" >}})


| OS         | OS Versions                  | EKS Anywhere version  |
|------------|------------------------------|-----------------------|
| Ubuntu        | 22.04     | 0.17+
|               | 20.04     | 0.5+
| Bottlerocket  | 1.26.1    | 0.21
|               | 1.20.0    | 0.20
|               | 1.19.1    | 0.19
|               | 1.15.1    | 0.18
| RHEL          | 9.x<sup>*</sup>	      | 0.18+
|               | 8.x	      | 0.12+

<sup>*</sup>Bare Metal, CloudStack and Nutanix only

* For details on supported operating systems for Admin machines, [see the Admin Machine page.]({{< relref "../getting-started/install/" >}})
* Older Bottlerocket versions are omitted from this table for brevity

## Frequently Asked Questions (FAQs)

**1. Where can I find details of the changes included in an EKS Anywhere version?** 

For changes included in an EKS Anywhere version, reference the [EKS Anywhere Changelog.]({{< relref "../whatsnew/changelog" >}})

**2. How can I be notified of new EKS Anywhere version releases?**

To configure notifications for new EKS Anywhere versions, follow the instructions on the [Release Alerts page.]({{< relref "../whatsnew/snsUpdates" >}})

**3. Does EKS Anywhere have extended support for Kubernetes versions?**

Yes. Extended support for Kubernetes versions was released in EKS Anywhere version v0.22 and adds an additional 12 months of CVE and critical bug fix patches for the Kubernetes version. To use Kubernetes versions under extended support with EKS Anywhere clusters, you must an EKS Anywhere Enterprise Subscription, and a valid and active license for each cluster you will create or upgrade with extended support Kubernetes versions. You must use EKS Anywhere version v0.22 or above to use extended support for Kubernetes versions.

**4. What happens on the end of support date for a Kubernetes version?**

Unlike Amazon EKS in the cloud, there are not forced upgrades with EKS Anywhere. On the end of support date, you can still create new EKS Anywhere clusters with the unsupported Kubernetes version if the EKS Anywhere version includes it. Existing EKS Anywhere clusters with unsupported Kubernetes versions continue to function. However, you will not be able to receive CVE patches or bug fixes for unsupported Kubernetes versions. Troubleshooting support, configuration guidance, and upgrade assistance is available for all Kubernetes and EKS Anywhere versions for customers with EKS Anywhere Enterprise Subscriptions.

**5. What EKS Anywhere versions are supported if you have the EKS Anywhere Enterprise Subscription?**

If you have purchased an EKS Anywhere Enterprise Subscription, AWS will provide troubleshooting support, configuration guidance, and upgrade assistance for your licensed clusters, irrespective of the EKS Anywhere version it's running on. However, as the CVE patches and bug fixes are only included in the latest and penultimate EKS Anywhere versions, it is recommended to use either of these releases to manage your deployments and keep them up to date. With an EKS Anywhere Enterprise Subscription, AWS will assist you in upgrading your licensed clusters to the latest EKS Anywhere version.

**6. Can I use different EKS Anywhere versions for my management cluster and workload clusters?**

Yes, the management cluster can be upgraded to newer EKS Anywhere versions than the workload clusters that it manages. However, a maximum skew of one EKS Anywhere minor version for management and workload clusters. This means the management cluster can be at most one EKS Anywhere minor version newer than the workload clusters (ie. management cluster with `v0.22.x` and workload clusters with `v0.21.x`). In the event that you want to upgrade your management cluster to a version that does not satisfy this condition, we recommend upgrading the workload cluster's EKS Anywhere version first to match the current management cluster's EKS Anywhere version, followed by an upgrade to your desired EKS Anywhere version for the management cluster.

> **NOTE**: Workload clusters can only be created with or upgraded to the same EKS Anywhere version that the management cluster was created with.
For example, if you create your management cluster with `v0.22.0`, you can only create workload clusters with `v0.22.0`. However, if you create your management cluster with version `v0.21.0` and then upgrade to `v0.22.0`, you can create workload clusters with either EKS Anywhere version `v0.21.0` or `v0.22.0`.

**7. Can I skip EKS Anywhere minor versions during upgrades (such as going from v0.20.x directly to v0.22.x)?**

No, it is not recommended to skip EKS Anywhere versions during upgrade. Regular upgrade reliability testing is only performed for sequential version upgrades (ie. going from version `0.20.x` to `0.21.x`, then from version `0.21.x` to `0.22.x`). However, you can choose to skip EKS Anywhere patch versions (ie. upgrade from `v0.21.3` to `v0.21.5`).

**8. What is the difference between an EKS Anywhere minor version versus a patch version?**

An EKS Anywhere minor version includes new EKS Anywhere capabilities, bug fixes, security patches, and new Kubernetes minor versions if they are available. An EKS Anywhere patch version generally includes only bug fixes, security patches, and Kubernetes patch version increments. EKS Anywhere patch versions are released more frequently than EKS Anywhere minor versions so you can receive the latest security and bug fixes sooner. For example, patch releases for the latest EKS Anywhere version follow a biweekly release cadence and patches for the penultimate EKS Anywhere version follow a monthly cadence.

**9. What kind of fixes are patched in the latest EKS Anywhere minor version?**

The latest EKS Anywhere minor version will receive CVE patches and bug fixes for EKS Anywhere components and the Kubernetes versions that are supported by the corresponding EKS Anywhere version. New curated packages versions, if any, will be made available as upgrades for this minor version.

**10. What kind of fixes are patched in the penultimate EKS Anywhere minor version?**

The penultimate EKS Anywhere minor version receives high and critical CVE patches and updates only to those Kubernetes versions that are supported by the corresponding EKS Anywhere version. New curated packages versions, if any, will be made available as upgrades for this minor version.

**11. Can I get notified when support is ending for a Kubernetes version on EKS Anywhere?**

Not automatically. You should check this page regularly and take note of the end of support date for the Kubernetes version you’re using and plan your upgrades accordingly.
