---
title: "EKS Anywhere curated packages"
linkTitle: "EKS Anywhere curated packages"
weight: 4
date: 2022-04-12
description: >
  Explanation of the process of managing curated packages in an EKS Anywhere cluster
---

EKS Anywhere curated packages is a framework to manage installation, configuration and maintenance of components that provide general operational capabilities for Kubernetes applications. Examples of these components include CSI, Ingress, LoadBalancer, Container Registry, Service Mesh, Secret Management, etc.

## Overview
The major components of EKS Anywhere curated packages are the package controller, the package build artifacts and the command line interface. The package controller will run in a pod in an EKS Anywhere cluster. The package controller will install packages from EKS Anywhere public ECR repository of tested package images.

## Package controller
The package controller will install, configure and remove packages from the cluster. The package controller will watch the packages and packagebundle custom resources for the packages to run and their configuration values. The user will need to perform install, uninstall and updates of package controller using helm.

The package controller will continual monitor for the latest bundle for the current Kubernetes release. The bundle controller will download and make active the latest bundle and all new installations will be from that active bundle. Any existing packages will show up as having an update available if the new bundle provides a newer release. It will be up to the customer to actually install the update. Any changes to an package will trigger and install, update or removal of that package. The package controller will use ECR to get all resources including bundle, helm charts, and container images.

## Package build artifacts
There are three types of build artifacts for packages: the container images, the helm charts and the package bundle manifests. The container images, helm charts and bundle manifests for all of the packages will be built and stored in EKS Anywhere public ECR repository. Each package may have multiple versions specified in the packages bundle. The bundle will reference package the helm chart tag in the ECR repository. The helm chart will reference the container images for the package.
