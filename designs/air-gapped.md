# Air Gapped user experience consolidation

## Introduction

**Problem:** we offer a set features that allow users to run EKS-A cluster in disconnected environments, but they lack cohesion.

Overtime, we have developed a particular set of features and tools that, even though they were aimed to solve independent problems, they all converge to offer to support for running EKS-A in environments without Internet connection.
As a result of their independent development, their interfaces and behaviors are not in sync, which makes the user experience not ideal.

This document aims to solve that, learning from the experience gained building such tools and offering a more cohesive and simple to use solution, without drastically changing the backbone of those features and how they interact with each other.

### Tenets

* ****Simple:**** minimal interface, minimum number of manual steps.

### Goals and Objectives

As an EKS-A cluster administrator I want to:

* Manage clusters in an environment without Internet connection.
* Populate my environment with the necessary dependencies without having to worry about what those dependencies are and with minimal effort.
* Interact with EKS-A clusters in air-gapped environments in the same way I interact with any other EKS-A cluster.

### Statement of Scope

**In scope**
* Download dependencies and package them.
* Populate disconnected environments with dependencies without Internet connection.

**Not in scope**
* Setup additional infrastructure to support disconnected environments (eg. private registries).

**Future scope**
* Package dependencies selectively.
* Populate disconnected environments with dependencies with Internet connection.
 
## Current state
In order to support disconnected environments, users currently have to:
1. Provision an OCI registry.
1. Run `eksctl anywhere download artifacts` to download manifests to disk.
1. Run `eksctl anywhere download images` to download images and charts to disk.
1. Run `eksctl anywhere import images` to push container images and charts to a registry.
1. Configure their cluster spec file to point to the registry 
1. Run all cluster commands (`create`, `upgrade`, etc.) with the `--bundles-override` flag pointing to the `Bundles` manifest downloaded in the first step.

## Overview of Solution
1. Users will provide an OCI registry.
1. We will store all dependencies (container images, helm charts and yaml manifests) in that registry.
1. If a cluster config with a [`registryMirrorConfiguration`](https://github.com/aws/eks-anywhere/blob/main/pkg/api/v1alpha1/cluster_types.go#L51) is provided, the CLI and the rest of components will always pull them from it.
No extra flags/configuration will be needed and the user will interact with this cluster as with any other one.

## Solution Details

### Populating the registry

In order to populate the registry, we will offer two commands:
```sh
eksctl anywhere export artifacts --output artifacts.tar
eksctl anywhere import artifacts --input artifacts.tar --registry myregistry.com
```

The first one downloads all dependencies (yaml manifests, images and helm charts). The only command argument here is the destination file. This command will create a tarball containing all 3 types of artifacts in that location.
The second one unpackages the tarball created in the first command, reads the packaged `Bundles` manifest and imports the referenced dependencies to a registry. The input here will be the artifacts tarball, the registry endpoint and the registry credentials (provided through env vars).

This two commands can be expanded in the future to add more capabilities like selectively exporting dependencies (eg. only for one kubernetes version or for only one provider).

### Pre packaged dependencies

A specific version of the CLI tied to one `Bundles` manifest will always produce the exact same artifacts tarball.
This means that we can prepackage the dependencies, store them in a public bucket and reference the tarball in the `Release` manifest.
This simplifies the experience for users who are interested in the default dependencies bundle.

## Implementation

The proposed design can be implemented in 4 incremental phases:
1. Add the two new commands but only push images and charts to the registry. This will require to keep using the `--bundles-override` pointing to a `Bundles` manifest in disk. This is an incremental improvement over the current state of `download images` so work should be minimal.
2. Push manifests to OCI registry and add the capability to the CLI to download them.
3. Start storing our default manifests in public ECR as opposed to Cloudfront + S3. This unifies even more the behavior of connected and disconnected environments.
4. Start packaging dependencies and serving them from the `Release` manifest.

## User Experience

### Deprecated commands
This proposal deprecates several commands:
* `download artifacts`
* `download images`
* `import-images`
* `import images`

We should inform users in the next release notes and keep the old commands in the codebase for at least one more release, printing a warning when executed.

## Testing

We should add at least one E2E test for the whole flow:
1. Download artifacts to disk
1. Import artifacts to a private registry
1. Create a cluster with `registryMirrorConfiguration` (ideally in an environment without external Internet connection).