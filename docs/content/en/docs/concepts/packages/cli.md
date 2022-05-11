---
title: EKS Anywhere curated package CLI
linkTitle: Curated package CLI
weight: 3
---

### Overview
The Curated Packages CLI provides the user experience required to manage curated packages.
Through the CLI, a user is able to discover, create, delete, and upgrade curated packages to a cluster.
These functionalities can be achieved during and after an eks anywhere cluster is created.

The CLI provides both imperative and declarative mechanisms to manage curated packages. These 
packages will be included as part of a packagebundle that will be provided by the eks-anywhere team.
Whenever a user requests a package creation through the CLI, a custom resource is created on the cluster
indicating the existence of a new package that needs to be installed. When a user executes a delete operation,
the custom resource will be removed from the cluster indicating the need for uninstalling a package. 
An upgrade through the CLI upgrades all packages to the latest release.

### Installation
Please check out [eks anywhere cli]({{< relref "../../getting-started/install" >}}) to install the CLI on your machine.

Also check out [create local cluster]({{< relref "../../getting-started/local-environment" >}}) and [create production cluster]({{< relref "../../getting-started/production-environment" >}}) for how to use the CLI on/after cluster creation

Check out [package management]({{< relref "../../tasks/packages" >}}) for how to use the CLI after a cluster is created and manage curated packages.