---
title: "Version support"
linkTitle: "Version support"
weight: 50
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

| Kubernetes version | Supported EKS Anywhere version | First supported    | End of support   |
|--------------------|---------------|--------------------|------------------|
| 1.27               | 0.16 | June 1, 2023       | September 2024   |
| 1.26               | 0.15 | April 5, 2023	     | July 2024        | 
| 1.25               | 0.15, 0.14 | February 16, 2023	 | April 2024       | 
| 1.24               | 0.15, 0.14, 0.13, 0.12 | November 17, 2022	 | January 2024     | 
| 1.23               | 0.15, 0.14, 0.13, 0.12, 0.11 | August 18, 2022	   | October 2023     | 
| 1.22               | 0.15, 0.14, 0.13, 0.12, 0.11, 0.10, 0.9, 0.8 | March 31, 2022     | May 2023         | 
| 1.21               | 0.14, 0.13, 0.12, 0.11, 0.10, 0.9, 0.8, 0.7, 0.6, 0.5 | September 8, 2021  | March 30, 2023   | 
| 1.20               | 0.12, 0.11, 0.10, 0.9, 0.8, 0.7, 0.6, 0.5 | September 8, 2021  | November 1, 2022 | 
| 1.19               | Not supported | --                 | --               | 
| 1.18               | Not supported | --                 | --               | 

The following table notes which EKS Anywhere and related Kubernetes versions are currently supported for CVE patches and bug fixes:

| EKS Anywhere version      | Kubernetes versions included | EKS Anywhere Release Date | CVE patches and bug fixes back-ported? |
|------------|------------------------------|---------------------------------|-------------------------|
| 0.15 | 1.26, 1.25, 1.24, 1.23, 1.22 | March 30, 2023 | Yes |
| 0.14 | 1.25, 1.24, 1.23, 1.22, 1.21 | January 19, 2023 | Yes |
| 0.13 | 1.24, 1.23, 1.22, 1.21       | December 15, 2022 | Yes |
| 0.12 | 1.24, 1.23, 1.22, 1.21, 1.20 | October 20, 2022 | No |
| 0.11 | 1.23, 1.22, 1.21, 1.20       | August 18, 2022 | No | 
| 0.10 | 1.22, 1.21, 1.20             | June 30, 2022 | No | 
| 0.9 | 1.21, 1.20                   | May 12, 2022 | No | 
| 0.8 | 1.22, 1.21, 1.20             | March 31, 2022 | No | 
| 0.7 | 1.21, 1.20                   | January 27, 2022 | No | 
| 0.6 | 1.21, 1.20                   | October 29, 2021 | No | 
| 0.5 | 1.21, 1.20                   | September 8, 2021 | No | 

* [Amazon EKS Anywhere Enterprise subscription](https://aws.amazon.com/eks/eks-anywhere/pricing/) is required to receive AWS support on any Amazon EKS Anywhere clusters.

## EKS Anywhere version support FAQs

### What is the difference between an Amazon EKS Anywhere minor version versus a patch version?

An Amazon EKS Anywhere minor version includes new Amazon EKS Anywhere capabilities, bug fixes, security patches, and a new Kubernetes minor version if there is one. An Amazon EKS Anywhere patch version generally includes only bug fixes, security patches, and Kubernetes patch version. Amazon EKS Anywhere patch versions are released more frequently than the Amazon EKS Anywhere minor versions so you can receive the latest security and bug fixes sooner. 

### Where can I find the content of the Amazon EKS Anywhere versions?

You can find the content of the previous Amazon EKS Anywhere minor versions and patch versions on the [What’s New]({{< relref "../changelog/" >}}) page.

### Will I get notified when there is a new Amazon EKS Anywhere version release?
You will get notified if you have subscribed as documented on the [Release Alerts]({{< relref "../snsupdates/" >}}) page.

### Can I skip Amazon EKS Anywhere minor versions during cluster upgrade (such as going from v0.9 directly to v0.11)?

No. We perform regular upgrade reliability testing for sequential version upgrade (e.g. going from version 0.9 to 0.10, then from version 0.10 to 0.11), but we do not perform testing on non-sequential upgrade path (e.g. going from version 0.9 directly to 0.11). You should _not_ skip minor versions during cluster upgrade. However, you can choose to skip patch versions.

### What kind of fixes are back-ported to the previous versions?*
Back-ported fixes include CVE patches and bug fixes for the Amazon EKS Anywhere components and the Kubernetes versions that are supported by the corresponding Amazon EKS Anywhere versions. 

### What happens on the end of support date for a Kubernetes version?
On the end of support date, you can still create a new cluster with the unsupported Kubernetes version using an old version of the Amazon EKS Anywhere toolkit that was released with this Kubernetes version. Any existing Amazon EKS Anywhere clusters with the unsupported Kubernetes version will continue to function. If you have the [Amazon EKS Anywhere Enterprise subscription](https://aws.amazon.com/eks/eks-anywhere/pricing/), AWS Support will continue to provide troubleshooting support and configuration guidance to those clusters as long as their Amazon EKS Anywhere versions are still being supported. However, you will not be able to receive CVE patches or bug fixes for the unsupported Kubernetes version.

### Will I get notified when support is ending for a Kubernetes version on Amazon EKS Anywhere?
Not automatically. You should check this page regularly and take note of the end of support date for the Kubernetes version you’re using.

