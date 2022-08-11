---
title: EKS Anywhere curated package build artifacts
linkTitle: "Curated package build artifacts"
weight: 2
---
There are three types of build artifacts for packages: the container images, the helm charts and the package bundle manifests. The container images, helm charts and bundle manifests for all of the packages will be built and stored in EKS Anywhere ECR repository. Each package may have multiple versions specified in the packages bundle. The bundle will reference the helm chart tag in the ECR repository. The helm chart will reference the container images for the package.