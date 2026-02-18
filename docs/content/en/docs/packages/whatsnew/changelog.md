---
title: "Changelog"
linkTitle: "Changelog"
weight: 7
description: >
  Changelog for Curated packages release
---
## Package Bundle Release (12-29-2025)
#### Changed
- ADOT `0.43.1` to `0.45.1`
- Emissary `3.9.1` to `3.10.0`
- Metrics-Server `3.12.2` to `3.13.0`
- Prometheus `2.55.1` to `3.8.0`

## Package Bundle Release (11-08-2025)
#### Changed
- Fix Critical CVEs in all curated packages
- Credential-Provider-Package `0.4.8` to `0.4.9`
  - Cleanup pkg controller private registry field [#1226](https://github.com/aws/eks-anywhere-packages/pull/1226)
- Cert-Manager `1.16.4` to `1.18.2`
- Metallb `0.14.9` to `0.15.2`

## Package Bundle Release (08-19-2025)
#### Changed
- Upgrade helm version `3.14.3` to `3.18.4`
- Credential-Provider-Package `0.4.6` to `0.4.8`
  - Enable proxy configuration [#1216](https://github.com/aws/eks-anywhere-packages/pull/1216)
- Cluster-Autoscaler `9.46.6` to `9.47.0` for Kubernetes version 1.33
- Metallb
  - Enable configuration for flag `ignoreExcludeLB`
  

## Package Bundle Release (05-22-2025)

#### Changed
- Fix Critical CVEs in all curated packages
- Harbor `2.12.1` to `2.12.2`
- ADOT `0.42.0` to `0.43.1`
  - Update [schema.json](https://github.com/aws/eks-anywhere-build-tooling/blob/main/projects/aws-observability/aws-otel-collector/helm/schema.json) to sync with upstream and support `extraManifests` field
- Cert-Manager `1.16.1` to `1.16.4`
- Metrics-Server `0.7.2` to `0.7.3`
- Credential-Provider-Package `0.4.5` to `0.4.6`
  - Use session token for ecr aws authentication [#1190](https://github.com/aws/eks-anywhere-packages/pull/1190)
- Cluster-Autoscaler `9.43.2` to `9.46.6`
  - Update [schema.json](https://github.com/aws/eks-anywhere-build-tooling/blob/main/projects/kubernetes/autoscaler/1-32/helm/schema.json) to allow additional properties in `extraArgs` field [example args](https://github.com/kubernetes/autoscaler/blob/cluster-autoscaler-chart-9.46.6/cluster-autoscaler/FAQ.md#what-are-the-parameters-to-ca)


## Package Bundle Release (02-28-2025)

#### Changed

- Harbor `2.11.1` to `2.12.1`
  - Fixes [schema.json](https://github.com/aws/eks-anywhere-build-tooling/blob/main/projects/goharbor/harbor/helm/schema.json) to sync with upstream version [PR](https://github.com/aws/eks-anywhere-build-tooling/pull/4373)
- ADOT `0.41.1` to `0.42.0`
  - The `logging` exporter is now [deprecated](https://github.com/open-telemetry/opentelemetry-collector/pull/11037), users should update the config to the `debug` exporter instead. Example configuration can be found [here](https://github.com/open-telemetry/opentelemetry-collector/blob/main/exporter/debugexporter/README.md)


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
