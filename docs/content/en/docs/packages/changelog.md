---
title: "Changelog"
linkTitle: "Changelog"
weight: 7
description: >
  Changelog for Curated packages release
---

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
