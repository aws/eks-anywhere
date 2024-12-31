---
title: "Changelog"
linkTitle: "Changelog"
weight: 7
description: >
  Changelog for Curated packages release
---

## Package Bundle Release (12-26-2024)

#### Changed

- cert-manager `1.15.3` to `1.16.1`
  - **Updated helm patch to include properties for eksa-packages in values.schema.json** [#4171](https://github.com/aws/eks-anywhere-build-tooling/pull/4171)
- cluster-autoscaler `9.43.0` to `9.43.2`
- credential-provider-package `0.4.4` to `0.4.5`
  - **Added support to update both legacy and default path for kubelet-extra-args for ubuntu** [#1177](https://github.com/aws/eks-anywhere-packages/pull/1177)
- metallb `0.14.8` to `0.14.9`
- prometheus `2.54.1` to `2.55.1`

## Package Bundle Release (10-18-2024)

#### Changed
- adot `0.40.1` to `0.41.1`
- cert-manager `1.14.7` to `1.15.3`
  - **Startupapicheck image change** [#3790](https://github.com/aws/eks-anywhere-build-tooling/pull/3790)

    As of this release the `cert-manager ctl` is no longer part of the main repo, it has been broken out into its own project. As such the `startupapicheck` job uses a new OCI image called `startupapicheck`. If you run in an environment in which images cannot be pulled, be sure to include the new image.
- cluster-autoscaler `9.37.0` to `9.43.0`
- harbor `2.11.0` to `2.11.1`
- metrics-server `0.7.1` to `0.7.2`
- prometheus `2.54.0` to `2.54.1`
