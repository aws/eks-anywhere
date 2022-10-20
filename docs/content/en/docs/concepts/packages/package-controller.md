---
title: EKS Anywhere curated package controller
linkTitle: Curated package controller
weight: 1
---

### Overview
The package controller will install, upgrade, configure and remove packages from the cluster. The package controller will watch the packages and packagebundle custom resources for the packages to run and their configuration values. The package controller only runs on the management cluster and manages packages on the management cluster and on the workload clusters.

Package release information is stored in a package bundle manifest. The package controller will continually monitor and download new package bundles. When a new package bundle is downloaded, it will show up as update available and users can use the CLI to activate the bundle to upgrade the installed packages.

Any changes to a package custom resource will trigger and install, upgrade, configuration or removal of that package. The package controller will use ECR or private registry to get all resources including bundle, helm charts, and container images.

### Installation
Please check out [create local cluster]({{< relref "../../getting-started/local-environment" >}}) and [create production cluster]({{< relref "../../getting-started/production-environment" >}}) for how to install package controller at the cluster creation time. 

Please check out [package management]({{< relref "../../tasks/packages" >}}) for how to install package controller after cluster creation and manage curated packages.
