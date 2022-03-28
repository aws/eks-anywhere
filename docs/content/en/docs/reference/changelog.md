---
title: "What's New?"
linkTitle: "What's New?"
weight: 55
menu:
  main:
    weight: 10
---

## [v0.7.2](https://github.com/aws/eks-anywhere/releases/tag/v0.7.2) - 2022-02-28

### Fixed
- Fix issue with downloading artifacts [#1327](https://github.com/aws/eks-anywhere/pull/1327)

## [v0.7.1](https://github.com/aws/eks-anywhere/releases/tag/v0.7.1) - 2022-02-25

### Added
- Support for taints in worker node group configurations [#189](https://github.com/aws/eks-anywhere/issues/189)
- Support for taints in control plane configurations [#189](https://github.com/aws/eks-anywhere/issues/189)
- Support for labels in worker node group configuration [#486](https://github.com/aws/eks-anywhere/issues/486)
- Allow removal of worker node groups using the `eksctl anywhere upgrade` command [#1054](https://github.com/aws/eks-anywhere/issues/1054)


## [v0.7.0](https://github.com/aws/eks-anywhere/releases/tag/v0.7.0) - 2022-01-27

### Added

- Support for [`aws-iam-authenticator`](https://github.com/kubernetes-sigs/aws-iam-authenticator) as an authentication option in EKS-A clusters [#90](https://github.com/aws/eks-anywhere/issues/90)
- Support for multiple worker node groups in EKS-A clusters [#840](https://github.com/aws/eks-anywhere/issues/840)
- Support for IAM Role for Service Account (IRSA) [#601](https://github.com/aws/eks-anywhere/issues/601)
- New command `upgrade plan cluster` lists core component changes affected by `upgrade cluster` [#499](https://github.com/aws/eks-anywhere/issues/499)
- Support for workload cluster's control plane and etcd upgrade through GitOps [#1007](https://github.com/aws/eks-anywhere/issues/1007)
- Upgrading a Flux managed cluster previously required manual steps. These steps have now been automated.
  [#759](https://github.com/aws/eks-anywhere/pull/759), [#1019](https://github.com/aws/eks-anywhere/pull/1019)
- Cilium CNI will now be upgraded by the `upgrade cluster` command [#326](https://github.com/aws/eks-anywhere/issues/326)

### Changed
- EKS-A now uses Cluster API (CAPI) v1.0.1 and v1beta1 manifests, upgrading from v0.3.23 and v1alpha3 manifests.
- Kubernetes components and etcd now use TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 as the
  configured TLS cipher suite [#657](https://github.com/aws/eks-anywhere/pull/657), 
  [#759](https://github.com/aws/eks-anywhere/pull/759)
- Automated git repository structure changes during Flux component `upgrade` workflow [#577](https://github.com/aws/eks-anywhere/issues/577)


## v0.6.0 - 2021-10-29

### Added

- Support to create and manage workload clusters [#94](https://github.com/aws/eks-anywhere/issues/94)
- Support for upgrading eks-anywhere components [#93](https://github.com/aws/eks-anywhere/issues/93), [Cluster upgrades]({{< relref "/docs/tasks/cluster/cluster-upgrades" >}})
  - IMPORTANT: Currently upgrading existing flux manged clusters requires performing a few [additional steps]({{< relref "/docs/tasks/cluster/cluster-upgrades" >}}). The fix for upgrading the existing clusters will be published in `0.6.1` release 
    to improve the upgrade experience.
- k8s CIS compliance [#193](https://github.com/aws/eks-anywhere/pull/192/)
- Support bundle improvements [#92](https://github.com/aws/eks-anywhere/issues/92)
- Ability to upgrade control plane nodes before worker nodes [#100](https://github.com/aws/eks-anywhere/issues/100)
- Ability to use your own container registry [#98](https://github.com/aws/eks-anywhere/issues/98)
- Make namespace configurable for anywhere resources [#177](https://github.com/aws/eks-anywhere/pull/177/files)


### Fixed
- Fix ova auto-import issue for multi-datacenter environments [#437](https://github.com/aws/eks-anywhere/issues/437)
- OVA import via EKS-A CLI sometimes fails [#254](https://github.com/aws/eks-anywhere/issues/254)
- Add proxy configuration to etcd nodes for bottlerocket [#195](https://github.com/aws/eks-anywhere/issues/195)


### Removed
- overrideClusterSpecFile field in cluster config

## v0.5.0

### Added

- Initial release of EKS-A
