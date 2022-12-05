---
title: "What's New?"
linkTitle: "What's New?"
weight: 35
---
## Unreleased



## [v0.12.2](https://github.com/aws/eks-anywhere/releases/tag/v0.12.2)

### Added
- Add support for Kubernetes 1.24 (CloudStack support to come in future releases)[#3491](https://github.com/aws/eks-anywhere/issues/3491)

### Fixed
- Fix authenticated registry mirror validations
- Fix capc bug causing orphaned VM's in slow environments
- Bundle activation problem for pacakge controller

## [v0.12.1](https://github.com/aws/eks-anywhere/releases/tag/v0.12.1)

### Changed
- Setting minimum wait time for nodes and machinedeployments (#3868, fixes #3822)

### Fixed
- Fixed worker node count pointer dereference issue (#3852)
- Fixed eks-anywhere-packages reference in go.mod (#3902)
- Surface dropped error in Cloudstack validations (#3832)

## [v0.12.0](https://github.com/aws/eks-anywhere/releases/tag/v0.12.0)


### ⚠️ Breaking changes
- Certificates signed with SHA-1 are not supported anymore for Registry Mirror. Users with a registry mirror and providing a custom CA cert will need to rotate the certificate served by the registry mirror endpoint before using the new EKS-A version. This is true for both new clusters (`create cluster` command) and existing clusters (`upgrade cluster` command).
- The `--source` option was removed from several package commands. Use either `--kube-version` for registry or `--cluster` for cluster.

### Added
- Add support for EKS Anywhere with provider CloudStack
- Add support to upgrade Bare Metal cluster
- Add support for using Registry Mirror for Bare Metal
- Redhat-based node image support for vSphere, CloudStack and Bare Metal EKS Anywhere clusters
- Allow authenticated image pull using Registry Mirror for Ubuntu on vSphere cluster
- Add option to disable vSphere CSI driver [#3148](https://github.com/aws/eks-anywhere/issues/3148)
- Add support for skipping load balancer deployment for Bare Metal so users can use their own load balancers [#3608](https://github.com/aws/eks-anywhere/pull/3608)
- Add support to configure aws-iam-authenticator on workload clusters independent of management cluster [#2814](https://github.com/aws/eks-anywhere/issues/2814)
- Add EKS Anywhere Packages support for remote management on workload clusters. (For more EKS Anywhere packages info: [v0.12.0](https://github.com/aws/eks-anywhere-packages/releases/tag/v0.2.14))
- Add new EKS Anywhere Packages
  - AWS Distro for OpenTelemetry (ADOT)
  - Cert Manager
  - Cluster Autoscaler
  - Metrics Server

### Fixed
- Remove special cilium network policy with `policyEnforcementMode` set to `always` due to lack of pod network connectivity for vSphere CSI
- Fixed [#3391](https://github.com/aws/eks-anywhere/issues/3391) [#3560](https://github.com/aws/eks-anywhere/issues/3560) for AWSIamConfig upgrades on EKS Anywhere workload clusters

## [v0.11.4](https://github.com/aws/eks-anywhere/releases/tag/v0.11.4)

### Added
- Add validate session permission for vsphere

### Fixed
- Fix datacenter naming bug for vSphere [#3381](https://github.com/aws/eks-anywhere/issues/3381)
- Fix os family validation for vSphere
- Fix controller overwriting secret for vSphere [#3404](https://github.com/aws/eks-anywhere/issues/3404)
- Fix unintended rolling upgrades when upgrading from an older EKS-A version for CloudStack

## [v0.11.3](https://github.com/aws/eks-anywhere/releases/tag/v0.11.3)

### Added
- Add some bundleRef validation
- Enable kube-rbac-proxy on CloudStack cluster controller's metrics port

### Fixed
- Fix issue with fetching EKS-D CRDs/manifests with retries
- Update BundlesRef when building a Spec from file
- Fix worker node upgrade inconsistency in Cloudstack

## [v0.11.2](https://github.com/aws/eks-anywhere/releases/tag/v0.11.2)

### Added
- Add a preflight check to validate vSphere user's permissions [#2744](https://github.com/aws/eks-anywhere/issues/2744)

### Changed
- Make `DiskOffering` in `CloudStackMachineConfig` optional

### Fixed
- Fix upgrade failure when flux is enabled [#3091](https://github.com/aws/eks-anywhere/issues/3091)[#3093](https://github.com/aws/eks-anywhere/issues/3093)
- Add token-refresher to default images to fix import/download images commands
- Improve retry logic for transient issues with kubectl applies and helm pulls [#3167](https://github.com/aws/eks-anywhere/issues/3167)
- Fix issue fetching curated packages images

## [v0.11.1](https://github.com/aws/eks-anywhere/releases/tag/v0.11.1)

### Added
- Add `--insecure` flag to import/download images commands [#2878](https://github.com/aws/eks-anywhere/issues/2878)

## [v0.11.0](https://github.com/aws/eks-anywhere/releases/tag/v0.11.0)

### Breaking Changes
- EKS Anywhere no longer distributes Ubuntu OVAs for use with EKS Anywhere clusters. Building your own Ubuntu-based nodes as described in [Building node images]({{< relref "./artifacts.md/#building-node-images" >}}) is the only supported way to get that functionality.

### Added
- Add support for Kubernetes 1.23 [#2159](https://github.com/aws/eks-anywhere/issues/2159)
- Add support for Support Bundle for validating control plane IP with vSphere provider
- Add support for aws-iam-authenticator on Bare Metal
- Curated Packages General Availability
- Added Emissary Ingress Curated Package

### Changed
- Install and enable GitOps in the existing cluster with upgrade command

## [v0.10.1](https://github.com/aws/eks-anywhere/releases/tag/v0.10.1)

### Changed
- Updated EKS Distro versions to latest release

### Fixed
- Fixed control plane nodes not upgraded for same kube version [#2636](https://github.com/aws/eks-anywhere/issues/2636)

## [v0.10.0](https://github.com/aws/eks-anywhere/releases/tag/v0.10.0)

### Added
- Added support for EKS Anywhere on bare metal with provider [tinkerbell](https://tinkerbell.org/). EKS Anywhere on bare metal supports complete provisioning cycle, including power on/off and PXE boot for standing up a cluster with the given hardware data.
- Support for node CIDR mask config exposed via the cluster spec. [#488](https://github.com/aws/eks-anywhere/issues/488)

### Changed
- Upgraded cilium from 1.9 to 1.10. [#1124](https://github.com/aws/eks-anywhere/issues/1124)
- Changes for EKS Anywhere packages [v0.10.0](https://github.com/aws/eks-anywhere-packages/releases/tag/v0.10.0)

### Fixed
- Fix issue using self-signed certificates for registry mirror [#1857](https://github.com/aws/eks-anywhere/issues/1857)

## [v0.9.2](https://github.com/aws/eks-anywhere/releases/tag/v0.9.0)

### Fixed
- Fix issue by avoiding processing Snow images when URI is empty

## [v0.9.1](https://github.com/aws/eks-anywhere/releases/tag/v0.9.0)

## [v0.9.0](https://github.com/aws/eks-anywhere/releases/tag/v0.9.0)

### Added
- Adding support to EKS Anywhere for a generic git provider as the source of truth for GitOps configuration management. [#9](https://github.com/aws/eks-anywhere/projects/9)
- Allow users to configure Cloud Provider and CSI Driver with different credentials. [#1730](https://github.com/aws/eks-anywhere/pull/1730)
- Support to install, configure and maintain operational components that are secure and tested by Amazon on EKS Anywhere clusters.[#2083](https://github.com/aws/eks-anywhere/issues/2083)
- A new Workshop section has been added to EKS Anywhere documentation.
- Added support for curated packages behind a feature flag [#1893](https://github.com/aws/eks-anywhere/pull/1893)

### Fixed
- Fix issue specifying proxy configuration for helm template command [#2009](https://github.com/aws/eks-anywhere/issues/2009)

## [v0.8.2](https://github.com/aws/eks-anywhere/releases/tag/v0.8.2)

### Fixed
- Fix issue with upgrading cluster from a previous minor version [#1819](https://github.com/aws/eks-anywhere/issues/1819)

## [v0.8.1](https://github.com/aws/eks-anywhere/releases/tag/v0.8.1)

### Fixed
- Fix issue with downloading artifacts [#1753](https://github.com/aws/eks-anywhere/issues/1753)

## [v0.8.0](https://github.com/aws/eks-anywhere/releases/tag/v0.8.0)
### Added
- SSH keys and Users are now mutable [#1208](https://github.com/aws/eks-anywhere/issues/1208)
- OIDC configuration is now mutable [#676](https://github.com/aws/eks-anywhere/issues/676)
- Add support for Cilium's policy enforcement mode [#726](https://github.com/aws/eks-anywhere/issues/726)

### Changed
- Install Cilium networking through Helm instead of static manifest
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
