---
title: "Changelog"
linkTitle: "Changelog"
weight: 10
aliases:
    /docs/reference/changelog/
description: >
  Changelog for EKS Anywhere releases
---

{{% alert title="Warnings" color="warning" %}}
* EKS Anywhere releases `v0.15.0` - `v0.15.2` have an issue with Tinkerbell provider where BMC/IPMI calls time out for certain hardware.<br>
Please upgrade to `v0.15.3` if you are using Tinkerbell (Bare Metal) provider.
* Installing CSI as part of VSphere cluster creation was deprecated in version `v0.16.0` and has been removed in `v0.17.0`. Please refer to the [deprecation section]({{< relref "../getting-started/vsphere/vsphere-getstarted/#vsphere-csi-driver-deprecation" >}}) in the vSphere provider documentation for more information.
{{% /alert %}}

## Unreleased

## [v0.16.0](https://github.com/aws/eks-anywhere/releases/tag/v0.16.0)

### Added
- Workload clusters full lifecycle API support for CloudStack provider ([#2754](https://github.com/aws/eks-anywhere/issues/2754))
- Enable proxy configuration for Bare Metal provider ([#5925](https://github.com/aws/eks-anywhere/issues/5925))
- Kubernetes 1.27 support ([#5929](https://github.com/aws/eks-anywhere/pull/5929))
- Support for upgrades for clusters with pod disruption budgets ([#5697](https://github.com/aws/eks-anywhere/pull/5697))
- BottleRocket network config uses mac addresses instead of interface names for configuring interfaces for the Bare Metal provider ([#3411](https://github.com/aws/eks-anywhere/issues/3411))
- Allow users to configure additional BottleRocket settings 
  - kernel sysctl settings ([#5304](https://github.com/aws/eks-anywhere/pull/5304)) 
  - boot kernel parameters ([#5359](https://github.com/aws/eks-anywhere/pull/5359)) 
  - custom trusted cert bundles ([#5625](https://github.com/aws/eks-anywhere/pull/5625))
- Add support for IRSA on Nutanix ([#5698](https://github.com/aws/eks-anywhere/pull/5698))
- Add support for aws-iam-authenticator on Nutanix ([#5698](https://github.com/aws/eks-anywhere/pull/5698))
- Enable proxy configuration for Nutanix ([#5779](https://github.com/aws/eks-anywhere/pull/5779))

### Upgraded
- Management cluster upgrades will only move management cluster’s components to bootstrap cluster and back. ([#5914](https://github.com/aws/eks-anywhere/pull/5914))

### Fixed
- CloudStack control plane host port is only defaulted in CAPI objects if not provided. ([#5792](https://github.com/aws/eks-anywhere/pull/5792)) ([#5736](https://github.com/aws/eks-anywhere/pull/5736))

### Deprecated
- Add warning to deprecate disableCSI through CLI ([#5918](https://github.com/aws/eks-anywhere/pull/5918)). Refer to the [deprecation section]({{< relref "../getting-started/vsphere/vsphere-getstarted/#vsphere-csi-driver-deprecation" >}}) in the vSphere provider documentation for more information.

### Removed
- Kubernetes 1.22 support

## [v0.15.4](https://github.com/aws/eks-anywhere/releases/tag/v0.15.4)

### Fixed

- Add validation for tinkerbell ip for workload cluster to match management cluster (#[5798](https://github.com/aws/eks-anywhere/pull/5798))
- Update datastore usage validation to account for space that will free up during upgrade (#[5524](https://github.com/aws/eks-anywhere/issues/5524))
- Expand GITHUB_TOKEN regex to support fine-grained access tokens (#[5764](https://github.com/aws/eks-anywhere/issues/5764))
- Display the timeout flags in CLI help (#[5637](https://github.com/aws/eks-anywhere/issues/5637))

## [v0.15.3](https://github.com/aws/eks-anywhere/releases/tag/v0.15.3)

### Added

- Added bundles-override to package cli commands ([#5695](https://github.com/aws/eks-anywhere/pull/5695))

### Fixed

- Remove last-applied annotation for kubectl replace ([#5684](https://github.com/aws/eks-anywhere/pull/5684))
- Fixed bmclib timeout issues when using Tinkerbell provider with older hardware ([aws/eks-anywhere-build-tooling#2117](https://github.com/aws/eks-anywhere-build-tooling/pull/2117))

## [v0.15.2](https://github.com/aws/eks-anywhere/releases/tag/v0.15.2)

### Supported OS version details
|              | vSphere | Baremetal |  Nutanix | Cloudstack | Snow  |
|    :---:     |  :---:  |   :---:   |   :---:  |    :---:   | :---: |
| Ubuntu       | 20.04	 | 20.04     |	20.04             | Not supported	     | 20.04 |  
| Bottlerocket | 1.13.1  | 1.13.1    |  Not supported     | Not supported	     | Not supported   |
| RHEL         | 8.7	   | 8.7	     |  Not supported     | 8.7	               | Not supported   |

### Added
- Support for no-timeouts to more EKS Anywhere operations (#[5565](https://github.com/aws/eks-anywhere/issues/5565))

### Changed
- Use kubectl for kube-proxy upgrader calls (#[5609](https://github.com/aws/eks-anywhere/pull/5609))

### Fixed
- Fixed the failure to delete a Tinkerbell workload cluster due to an incorrect SSH key update during reconciliation (#[5554](https://github.com/aws/eks-anywhere/issues/5554))
- Fixed `machineGroupRef` updates for CloudStack and Vsphere  (#[5313](https://github.com/aws/eks-anywhere/issues/5313))

## [v0.15.1](https://github.com/aws/eks-anywhere/releases/tag/v0.15.1)

### Supported OS version details
|              | vSphere | Baremetal |  Nutanix | Cloudstack | Snow  |
|    :---:     |  :---:  |   :---:   |   :---:  |    :---:   | :---: |
| Ubuntu       | 20.04	 | 20.04     |	20.04             | Not supported	     | 20.04 |  
| Bottlerocket | 1.13.1  | 1.13.1    |  Not supported     | Not supported	     | Not supported   |
| RHEL         | 8.7	   | 8.7	     |  Not supported     | 8.7	               | Not supported   |

### Added
- Kubernetes 1.26 support

### Upgraded
- Cilium updated from version `v1.11.10` to version `v1.11.15`

### Fixed
- Fix http client in file reader to honor the provided HTTP_PROXY, HTTPS_PROXY and NO_PROXY env variables ([#5488](https://github.com/aws/eks-anywhere/pull/5488))

## [v0.15.0](https://github.com/aws/eks-anywhere/releases/tag/v0.15.0)

### Supported OS version details
|              | vSphere | Baremetal |  Nutanix | Cloudstack | Snow  |
|    :---:     |  :---:  |   :---:   |   :---:  |    :---:   | :---: |
| Ubuntu       | 20.04	 | 20.04     |	20.04             | Not supported	     | 20.04 |  
| Bottlerocket | 1.13.1  | 1.13.1    |  Not supported     | Not supported	     | Not supported   |
| RHEL         | 8.7	   | 8.7	     |  Not supported     | 8.7	               | Not supported   |

### Added
- Workload clusters full lifecycle API support for Bare Metal provider ([#5237](https://github.com/aws/eks-anywhere/issues/5237))
- IRSA support for Bare Metal ([#4361](https://github.com/aws/eks-anywhere/issues/4361))
- Support for mixed disks within the same node grouping for BareMetal clusters ([#3234](https://github.com/aws/eks-anywhere/issues/3234))
- Workload clusters full lifecycle API support for Nutanix provider ([#5190](https://github.com/aws/eks-anywhere/pull/5190))
- OIDC support for Nutanix ([#4711](https://github.com/aws/eks-anywhere/pull/4711))
- Registry mirror support for Nutanix ([#5236](https://github.com/aws/eks-anywhere/pull/5236))
- Support for linking EKS Anywhere node VMs to Nutanix projects ([#5266](https://github.com/aws/eks-anywhere/pull/5266))
- Add `CredentialsRef` to `NutanixDatacenterConfig` to specify Nutanix credentials for workload clusters ([#5114](https://github.com/aws/eks-anywhere/pull/5114))
- Support for taints and labels for Nutanix provider ([#5172](https://github.com/aws/eks-anywhere/issues/5172))
- Support for InsecureSkipVerify for RegistryMirrorConfiguration across all providers. Currently only supported for Ubuntu and RHEL OS. ([#1647](https://github.com/aws/eks-anywhere/issues/1647))
- Support for configuring of Bottlerocket settings. ([#707](https://github.com/aws/eks-anywhere/issues/707))
- Support for using a custom CNI ([#5217](https://github.com/aws/eks-anywhere/issues/5217))
- Ability to configure NTP servers on EKS Anywhere nodes for vSphere and Tinkerbell providers ([#4760](https://github.com/aws/eks-anywhere/issues/4760))
- Support for nonRootVolumes option in SnowMachineConfig ([#5199](https://github.com/aws/eks-anywhere/issues/5199))
- Validate template disk size with vSphere provider using Bottlerocket ([#1571](https://github.com/aws/eks-anywhere/issues/1571))
- Allow users to specify `cloneMode` for different `VSphereMachineConfig` ([#4634](https://github.com/aws/eks-anywhere/pull/4634))
- Validate management cluster bundles version is the same or newer than bundle version used to upgrade a workload cluster([#5105](https://github.com/aws/eks-anywhere/issues/5105))
- Set hostname for Bottlerocket nodes ([#3629](https://github.com/aws/eks-anywhere/issues/3629))
- Curated Package controller as a package ([#831](https://github.com/aws/eks-anywhere-packages/pull/831))
- Curated Package Credentials Package ([#829](https://github.com/aws/eks-anywhere-packages/pull/829))
- Enable Full Cluster Lifecycle for curated packages ([#807](https://github.com/aws/eks-anywhere-packages/issues/807))
- Curated Package Controller Configuration in Cluster Spec ([#5031](https://github.com/aws/eks-anywhere/pull/5031))

### Upgraded

- Bottlerocket upgraded from `v1.13.0` to `v1.13.1`
- Upgrade EKS Anywhere admin AMI to Kernel 5.15
- Tinkerbell stack upgraded ([#3233](https://github.com/aws/eks-anywhere/issues/3233)):
  - Cluster API Provider Tinkerbell `v0.4.0`
  - Hegel `v0.10.1`
  - Rufio `v0.2.1`
  - Tink `v0.8.0`
- Curated Package Harbor upgraded from `2.5.1` to `2.7.1`
- Curated Package Prometheus upgraded from `2.39.1` to `2.41.0`
- Curated Package Metallb upgraded from `0.13.5` to `0.13.7`
- Curated Package Emissary upgraded from `3.3.0` to `3.5.1`

### Fixed
- Applied a patch that fixes vCenter sessions leak ([#1767](https://github.com/aws/eks-anywhere-build-tooling/issues/2057))

### Breaking changes
- Removed support for Kubernetes 1.21

## [v0.14.6](https://github.com/aws/eks-anywhere/releases/tag/v0.14.6)

### Fixed

- Fix clustermanager no-timeouts option ([#5445](https://github.com/aws/eks-anywhere/pull/5445))

## [v0.14.5](https://github.com/aws/eks-anywhere/releases/tag/v0.14.5)

### Fixed

- Fix kubectl get call to point to full API name ([#5342](https://github.com/aws/eks-anywhere/pull/5342))
- Expand all kubectl calls to fully qualified names ([#5347](https://github.com/aws/eks-anywhere/pull/5347))

## [v0.14.4](https://github.com/aws/eks-anywhere/releases/tag/v0.14.4)

### Added

- `--no-timeouts` flag in create and upgrade commands to disable timeout for all wait operations
- Management resources backup procedure with clusterctl

## [v0.14.3](https://github.com/aws/eks-anywhere/releases/tag/v0.14.3)

### Added

- `--aws-region` flag to `copy packages` command.

### Upgraded

- CAPAS from `v0.1.22` to `v0.1.24`.

## [v0.14.2](https://github.com/aws/eks-anywhere/releases/tag/v0.14.2)

### Added
- Enabled support for Kubernetes version 1.25

## [v0.14.1](https://github.com/aws/eks-anywhere/releases/tag/v0.14.1)

### Added
- support for authenticated pulls from registry mirror ([#4796](https://github.com/aws/eks-anywhere/pull/4796))
- option to override default nodeStartupTimeout in machine health check ([#4800](https://github.com/aws/eks-anywhere/pull/4800))
- Validate control plane endpoint with pods and services CIDR blocks([#4816](https://github.com/aws/eks-anywhere/pull/4816))


### Fixed
- Fixed a issue where registry mirror settings weren’t being applied properly on Bottlerocket nodes for Tinkerbell provider

## [v0.14.0](https://github.com/aws/eks-anywhere/releases/tag/v0.14.0)

### Added
- Add support for EKS Anywhere on AWS Snow ([#1042](https://github.com/aws/eks-anywhere/issues/1042))
- Static IP support for BottleRocket ([#4359](https://github.com/aws/eks-anywhere/pull/4359))
- Add registry mirror support for curated packages
- Add copy packages command ([#4420](https://github.com/aws/eks-anywhere/pull/4420))

### Fixed
- Improve management cluster name validation for workload clusters

## [v0.13.1](https://github.com/aws/eks-anywhere/releases/tag/v0.13.1)

### Added
- Multi-region support for all supported curated packages

### Fixed
- Fixed nil pointer in `eksctl anywhere upgrade plan` command

## [v0.13.0](https://github.com/aws/eks-anywhere/releases/tag/v0.13.0)
### Added
- Workload clusters full lifecycle API support for vSphere and Docker ([#1090](https://github.com/aws/eks-anywhere/issues/1090))
- Single node cluster support for Bare Metal provider
- Cilium updated to version `v1.11.10`
- CLI high verbosity log output is automatically included in the support bundle after a CLI `cluster` command error ([#1703](https://github.com/aws/eks-anywhere/issues/1703) implemented by [#4289](https://github.com/aws/eks-anywhere/issues/1703))
- Allow to configure machine health checks timeout through a new flag `--unhealthy-machine-timeout` ([#3918](https://github.com/aws/eks-anywhere/issues/3918) implemented by [#4123](https://github.com/aws/eks-anywhere/pull/4123))
- Ability to configure rolling upgrade for Bare Metal and Cloudstack via `maxSurge` and `maxUnavailable` parameters
- New Nutanix Provider
- Workload clusters support for Bare Metal
- VM Tagging support for vSphere VM's created in the cluster ([#4228](https://github.com/aws/eks-anywhere/pull/4228))
- Support for new curated packages:
  - Prometheus `v2.39.1`
- Updated curated packages' versions:
  - ADOT `v0.23.0` upgraded from `v0.21.1`
  - Emissary `v3.3.0` upgraded from `v3.0.0`
  - Metallb `v0.13.7` upgraded from `v0.13.5`
- Support for packages controller to create target namespaces [#601](https://github.com/aws/eks-anywhere-packages/issues/601)
- (For more EKS Anywhere packages info: [v0.13.0](https://github.com/aws/eks-anywhere-packages/releases/tag/v0.2.22))

### Fixed
- Kubernetes version upgrades from 1.23 to 1.24 for Docker clusters ([#4266](https://github.com/aws/eks-anywhere/pull/4266))
- Added missing docker login when doing authenticated registry pulls

### Breaking changes
- Removed support for Kubernetes 1.20

## [v0.12.2](https://github.com/aws/eks-anywhere/releases/tag/v0.12.2)

### Added
- Add support for Kubernetes 1.24 (CloudStack support to come in future releases)[#3491](https://github.com/aws/eks-anywhere/issues/3491)

### Fixed
- Fix authenticated registry mirror validations
- Fix capc bug causing orphaned VM's in slow environments
- Bundle activation problem for package controller

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
- EKS Anywhere no longer distributes Ubuntu OVAs for use with EKS Anywhere clusters. Building your own Ubuntu-based nodes as described in [Building node images]({{< relref "../osmgmt/artifacts.md/#building-node-images" >}}) is the only supported way to get that functionality.

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
- Support for upgrading eks-anywhere components [#93](https://github.com/aws/eks-anywhere/issues/93), [Cluster upgrades]({{< relref "../clustermgmt/cluster-upgrades" >}})
  - IMPORTANT: Currently upgrading existing flux manged clusters requires performing a few [additional steps]({{< relref "../clustermgmt/cluster-upgrades" >}}). The fix for upgrading the existing clusters will be published in `0.6.1` release 
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
