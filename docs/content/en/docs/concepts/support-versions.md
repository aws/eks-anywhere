---
title: "Versioning"
linkTitle: "Versioning"
weight: 30
aliases:
    /docs/reference/support/support-versions/
description: >
  EKS Anywhere and Kubernetes version support policy and release cycle
---

This page contains information on the EKS Anywhere release cycle and support for Kubernetes versions.

When creating new clusters, we recommend that you use the latest available Kubernetes version supported by EKS Anywhere. If your application requires a specific version of Kubernetes, you can select older versions. You can create new EKS Anywhere clusters on any Kubernetes version that the EKS Anywhere version supports.

You must have an [EKS Anywhere Enterprise Subscription]({{< relref "./support-scope" >}}) to receive support for EKS Anywhere from AWS.

## Kubernetes versions

Each EKS Anywhere version includes support for multiple Kubernetes minor versions.

The release and support schedule for Kubernetes versions in EKS Anywhere aligns with the Amazon EKS standard support schedule as documented on the [Amazon EKS Kubernetes release calendar.](https://docs.aws.amazon.com/eks/latest/userguide/kubernetes-versions.html#kubernetes-release-calendar) A minor Kubernetes version is under standard support in EKS Anywhere for 14 months after it's released in EKS Anywhere. EKS Anywhere currently does not offer extended version support for Kubernetes versions. If you are interested in extended version support for Kubernetes versions in EKS Anywhere, please upvote or comment on [EKS Anywhere GitHub Issue #6793.](https://github.com/aws/eks-anywhere/issues/6793) Patch releases for Kubernetes versions are included in EKS Anywhere as they become available in EKS Distro.

Unlike Amazon EKS, there are no automatic upgrades in EKS Anywhere and you have full control over when you upgrade. On the end of support date, you can still create new EKS Anywhere clusters with the unsupported Kubernetes version if the EKS Anywhere version you are using includes it. Any existing EKS Anywhere clusters with the unsupported Kubernetes version continue to function. As new Kubernetes versions become available in EKS Anywhere, we recommend that you proactively update your clusters to use the latest available Kubernetes version to remain on versions that receive CVE patches and bug fixes.

Reference the table below for release and support dates for each Kubernetes version in EKS Anywhere. The Release Date column denotes the EKS Anywhere release date when the Kubernetes version was first supported in EKS Anywhere. Note, dates with only a month and a year are approximate and are updated with an exact date when it's known.

<!-- See /docs/data/version_support.yml -->
{{< kube_support >}}

* Older Kubernetes versions are omitted from this table for brevity, reference the [EKS Anywhere GitHub](https://github.com/aws/eks-anywhere/tree/main/docs/data/version_support.yml) for older versions.

## EKS Anywhere versions

Each EKS Anywhere version includes all components required to create and manage EKS Anywhere clusters. For example, this includes:

- Administrative / CLI components (eksctl CLI, image-builder, diagnostics-collector)
- Management components (Cluster API controller, EKS Anywhere controller, provider-specific controllers)
- Workload components (Kubernetes, Cilium)

You can find details about each EKS Anywhere releases in the [EKS Anywhere release manifest.](https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml) The release manifest contains references to the corresponding bundle manifest for each EKS Anywhere version. Within the bundle manifest, you will find the components included in a specific EKS Anywhere version. The images running in your deployment use the same uri values specified in the bundle manifest for that component. For example, see the [bundle manifest](https://anywhere-assets.eks.amazonaws.com/releases/bundles/57/manifest.yaml) for EKS Anywhere v0.18.5.

Starting in 2024, EKS Anywhere follows a 4-month release cadence for minor versions. EKS Anywhere has a 2-week cadence for patch versions. Common vulnerabilities and exposures (CVE) patches and bug fixes, including those for the supported Kubernetes versions, are included in the latest EKS Anywhere minor version (version N). If you are interested in patch support for EKS Anywhere N-1 versions, please upvote or comment on [EKS Anywhere GitHub Issue #7397.](https://github.com/aws/eks-anywhere/issues/7397)


Reference the table below for release dates and patch support for each EKS Anywhere version. This table shows the Kubernetes versions that are supported in each EKS Anywhere version.
<!-- See /docs/data/version_support.yml -->
{{< eksa_support >}}

* Older EKS Anywhere versions are omitted from this table for brevity, reference the [EKS Anywhere GitHub](https://github.com/aws/eks-anywhere/tree/main/docs/data/version_support.yml) for older versions.

## Operating System versions

Bottlerocket, Ubuntu, and Red Hat Enterprise Linux (RHEL) can be used as operating systems for nodes in EKS Anywhere clusters. Reference the table below for operating system version support in EKS Anywhere. For information on operating system management in EKS Anywhere, reference the [Operating System Management Overview page]({{< relref "../osmgmt/overview" >}})


| OS         | OS Versions                  | Supported EKS Anywhere version  |
|------------|------------------------------|---------------------------------|
| Ubuntu        | 22.04     | 0.17 and above
|               | 20.04     | 0.5 and above
| Bottlerocket  | 1.15.1    | 0.18
|               | 1.13.1    | 0.15-0.17
|               | 1.12.0    | 0.14
|               | 1.10.1    | 0.12
| RHEL          | 9.x<sup>*</sup>	      | 0.18
| RHEL          | 8.x	      | 0.12 and above

<sup>*</sup>Nutanix only

* For details on supported operating systems for Admin machines, [see the Admin Machine page.]({{< relref "../getting-started/install/" >}})
* Older Bottlerocket versions are omitted from this table for brevity

## Frequently Asked Questions (FAQs)

### Where can I find details of what changed in an EKS Anywhere version?

For changes included in an EKS Anywhere version, reference the [EKS Anywhere Changelog.]({{< relref "../whatsnew/changelog" >}})

### Will I get notified when there is a new EKS Anywhere version release?

You will get notified if you have subscribed as documented on the [Release Alerts page.]({{< relref "../whatsnew/snsUpdates" >}})

### Does Amazon EKS extended support for Kubernetes versions apply to EKS Anywhere clusters?

No. Amazon EKS extended support for Kubernetes versions does not apply to EKS Anywhere at this time. To request this capability, please comment or upvote on this [EKS Anywhere GitHub issue](https://github.com/aws/eks-anywhere/issues/6793).

### What happens on the end of support date for a Kubernetes version?

Unlike Amazon EKS, there are no forced upgrades in EKS Anywhere. On the end of support date, you can still create new EKS Anywhere clusters with the unsupported Kubernetes version if the EKS Anywhere version you are using includes it. Any existing EKS Anywhere clusters with the unsupported Kubernetes version will continue to function. However, you will not be able to receive CVE patches or bug fixes for the unsupported Kubernetes version. Troubleshooting support, configuration guidance, and upgrade assistance is available for all Kubernetes and EKS Anywhere versions for customers with EKS Anywhere Enterprise Subscriptions.

### What EKS Anywhere versions are supported if you have the EKS Anywhere Enterprise Subscription?

If you have purchased an EKS Anywhere Enterprise Subscription, AWS will provide troubleshooting support, configuration guidance, and upgrade assistance for your licensed clusters, irrespective of the EKS Anywhere version it's running on. However, as the CVE patches and bug fixes are only included in the latest EKS Anywhere version, it is recommended to keep your deployments updated with the latest EKS Anywhere release. With an EKS Anywhere Enterprise Subscription, AWS will assist you in upgrading your licensed clusters to the latest EKS Anywhere version.

### Can I use different EKS Anywhere minor versions for my management cluster and workload clusters?

Yes, the management cluster can be upgraded to newer EKS Anywhere versions than the workload clusters that it manages. However, we only support a maximum skew of one EKS Anywhere minor version for management and workload clusters. This means the management cluster can be at most one EKS Anywhere minor version newer than the workload clusters (ie. management cluster with v0.18.x and workload clusters with v0.17.x). In the event that you want to upgrade your management cluster to a version that does not satisfy this condition, we recommend upgrading the workload cluster's EKS Anywhere version first to match the current management cluster's EKS Anywhere version, followed by an upgrade to your desired EKS Anywhere version for the management cluster.

> **NOTE**: Workload clusters can only be created with or upgraded to the same EKS Anywhere version that the management cluster was created with.
For example, if you create your management cluster with v0.18.0, you can only create workload clusters with v0.18.0. However, if you create your management cluster with version v0.17.0 and then upgrade to v0.18.0, you can create workload clusters with either v0.17.0 or v0.18.0.

### Can I skip EKS Anywhere minor versions during cluster upgrade (such as going from v0.16 directly to v0.18)?

No. We perform regular upgrade reliability testing for sequential version upgrade (ie. going from version 0.16 to 0.17, then from version 0.17 to 0.18), but we do not perform testing on non-sequential upgrade path (ie. going from version 0.16 directly to 0.18). You should _not_ skip minor versions during cluster upgrade. However, you can choose to skip patch versions.

### What is the difference between an EKS Anywhere minor version versus a patch version?

An EKS Anywhere minor version includes new EKS Anywhere capabilities, bug fixes, security patches, and new Kubernetes minor versions if they are available. An EKS Anywhere patch version generally includes only bug fixes, security patches, and Kubernetes patch version increments. EKS Anywhere patch versions are released more frequently than EKS Anywhere minor versions so you can receive the latest security and bug fixes sooner.

### What kind of fixes are patched in the latest EKS Anywhere minor version?

Patches include CVE patches and bug fixes for EKS Anywhere components and the Kubernetes versions that are supported by the corresponding EKS Anywhere version.

### Will I get notified when support is ending for a Kubernetes version on EKS Anywhere?

Not automatically. You should check this page regularly and take note of the end of support date for the Kubernetes version you’re using.
