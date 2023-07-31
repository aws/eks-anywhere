---
title: "Overview of curated packages"
linkTitle: Overview
aliases:
    /docs/concepts/packages/artifacts/
    /docs/concepts/packages/cli/
    /docs/concepts/packages/package-controller/
    /docs/reference/packagespec/packages/
weight: 1
---

Components of EKS Anywhere curated packages consist of a controller, a CLI, and artifacts.

### Package controller
The package controller will install, upgrade, configure and remove packages from the cluster. The package controller will watch the packages and packagebundle custom resources for the packages to run and their configuration values. The package controller only runs on the management cluster and manages packages on the management cluster and on the workload clusters.

Package release information is stored in a package bundle manifest. The package controller will continually monitor and download new package bundles. When a new package bundle is downloaded, it will show up as update available and users can use the CLI to activate the bundle to upgrade the installed packages.

Any changes to a package custom resource will trigger and install, upgrade, configuration or removal of that package. The package controller will use ECR or private registry to get all resources including bundle, helm charts, and container images.

The Getting started page for each EKS Anywhere provider describes how to install the package controller at the cluster creation time. See the [EKS Anywhere providers]({{< relref "../getting-started/chooseprovider/" >}}) page the list of providers.

Please check out [package management]({{< relref "../packages/packages" >}}) for how to install package controller after cluster creation and manage curated packages.

### Packages CLI
The Curated Packages CLI provides the user experience required to manage curated packages.
Through the CLI, a user is able to discover, create, delete, and upgrade curated packages to a cluster.
These functionalities can be achieved during and after an EKS Anywhere cluster is created.

The CLI provides both imperative and declarative mechanisms to manage curated packages. These 
packages will be included as part of a `packagebundle` that will be provided by the EKS Anywhere team.
Whenever a user requests a package creation through the CLI (`eksctl anywhere create package`), a custom resource is created on the cluster
indicating the existence of a new package that needs to be installed. When a user executes a delete operation (`eksctl anywhere delete package`),
the custom resource will be removed from the cluster indicating the need for uninstalling a package. 
An upgrade through the CLI (`eksctl anywhere upgrade packages`) upgrades all packages to the latest release.

Please check out [Install EKS Anywhere]({{< relref "../getting-started/install" >}}) to install the `eksctl anywhere` CLI on your machine.

The Getting started page for each EKS Anywhere provider describes how to install the package controller at the cluster creation time. See the [EKS Anywhere providers]({{< relref "../getting-started/chooseprovider/" >}}) page the list of providers.

Please check out [package management]({{< relref "../packages/packages" >}}) for how to install package controller after cluster creation and manage curated packages.

### Curated packages artifacts
There are three types of build artifacts for packages: the container images, the helm charts and the package bundle manifests. The container images, helm charts and bundle manifests for all of the packages will be built and stored in EKS Anywhere ECR repository. Each package may have multiple versions specified in the packages bundle. The bundle will reference the helm chart tag in the ECR repository. The helm chart will reference the container images for the package.
