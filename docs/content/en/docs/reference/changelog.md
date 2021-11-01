
# Changelog

## [Unreleased]

### Added

- aws-iam-authenticator support
- Allow the retention of node taints during an upgrade

## v0.6.0 - 2021-10-29

### Added

- Support to create and manage a management cluster
- Support for upgrading eks-anywhere components
- k8s CIS compliance
- Support bundle improvements
- Ability to upgrade control plane nodes before worker nodes
- Ability to "bring your own registry" [issue 98](https://github.com/aws/eks-anywhere/issues/98)
- Make namespace configurable for anywhere resources [#177](hsuttps://github.com/aws/eks-anywhere/pull/177/files)
- Add proxy configuration to etcd nodes for bottlerocket [#195](https://github.com/aws/eks-anywhere/issues/195)

### Fixed
- Fix ova auto-import issue for multi-datacenter environments [#437](https://github.com/aws/eks-anywhere/issues/437)
- OVA import via EKS-A CLI sometimes fails [#254](https://github.com/aws/eks-anywhere/issues/254)

### Removed
- overrideClusterSpecFile field in cluster config

## v0.5.0

### Added

- Initial release of EKS-A