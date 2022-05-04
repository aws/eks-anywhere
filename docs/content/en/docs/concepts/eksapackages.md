---
title: "EKS Anywhere curated packages"
linkTitle: "EKS Anywhere curated packages"
weight: 4
date: 2022-04-12
description: >
  Explanation of the process of managing curated packages in an EKS Anywhere cluster
---

EKS Anywhere Curated Packages is a management system for installation, configuration and maintenance of additional components for your Kubernetes applications. Examples of these components may include Container Registry, Ingress, and LoadBalancer, etc.

## Overview
The major components of EKS Anywhere Curated Packages are the package controller, the package build artifacts and the command line interface. The package controller will run in a pod in an EKS Anywhere cluster. The package controller will manage the lifecycle of packages tested and maintained by EKS Anywhere.

## Package controller
The package controller will install, upgrade, configure and remove packages from the cluster. The package controller will watch the packages and packagebundle custom resources for the packages to run and their configuration values.

Package release information is stored in a package bundle manifest. The package controller will continually monitor and download new package bundles. When a new package bundle is downloaded, it will show up as update available and users can use the CLI to activate the bundle to upgrade the installed packages.

Any changes to a package custom resource will trigger and install, upgrade, configuration or removal of that package. The package controller will use ECR or private registry to get all resources including bundle, helm charts, and container images.

## Package build artifacts
There are three types of build artifacts for packages: the container images, the helm charts and the package bundle manifests. The container images, helm charts and bundle manifests for all of the packages will be built and stored in EKS Anywhere public ECR repository. Each package may have multiple versions specified in the packages bundle. The bundle will reference the helm chart tag in the ECR repository. The helm chart will reference the container images for the package.
