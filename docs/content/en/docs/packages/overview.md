---
title: Overview of EKS Anywhere curated packages
weight: 20
---

### Overview
EKS Anywhere Curated Packages is a management system for installation, configuration and maintenance of additional components for your Kubernetes cluster. Examples of these components may include Container Registry, Ingress, and LoadBalancer, etc.

The major components of EKS Anywhere Curated Packages are the [package controller]({{< relref "package-controller" >}}), the [package build artifacts]({{< relref "artifacts" >}}) and the [command line interface]({{< relref "cli" >}}). The package controller will run in a pod in an EKS Anywhere cluster. The package controller will manage the lifecycle of packages tested and maintained by EKS Anywhere.

### Packages
Please check out [curated package list]({{< relref "../reference/packagespec" >}}) for the complete list of EKS Anywhere curated packages.