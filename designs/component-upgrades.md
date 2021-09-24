# EKS-A components upgrade


## Introduction

**Problem:** customers are only able to upgrade the cluster components that are part of the kubernetes distribution. Any other core component of EKS-A remains untouched from the cluster creation.

This stops customers from enjoying the latest features and bug fixes from eks-a and any of its core components as well as from being able to get security patches.

### Tenets

* **Simple:** simple to use, simple to maintain, simple to build

### Goals and objectives

As an eks-a customer:

* I want to able to enjoy the latest eks-a features and bug fixes
* I want to keep my cluster up to date with security patches
* I don’t want to need to stay informed about latest versions, security patches or compatibility between components

### Statement of scope

**In scope**

* Upgrade core eks-a components to latest versions available for an specific cli version
    * Core CAPI
    * CAPI providers
    * Cert-manager
    * Etcdadm CAPI provider
    * CNIs (cilium)
    * EKS-A controller and CRDs
    * Flux

**Not in scope**

* Install/upgrade add-ons
* Upgrade cli
* Upgrades to customer defined component versions

**Future scope**

*  Version changes report
* “Notification” for new available versions
* Automatic/scheduled upgrades

## Overview of solution

**Upgrade all eks-a components to the versions declared in the latest bundles manifest when running the `upgrade cluster` command.**

## Solution details

Since we already have:

* An interface that customers use to upgrade clusters (`upgrade cluster` command)
* An declarative list of all eks-a components at their latest available versions that have been tested together (bundles manifest)

It seems like the simplest solution is to take advantage of both.

This solution fits very well with both the goal of a customer experience as simple as possible and the concept of the versions bundle (a group of versioned components that work well together):

* We don’t introduce more concepts/commands. Customers just have to use the same command they are already used to. The syntax stays the same and there is no need to define what needs to be upgraded and to what version.
The cli takes care of this transparently.
* Components bundled together are always upgraded together. This reduces our testing matrix considerably, since there is no room for version miss-alignments.

The cli needs the ability to detect both component version changes as well as new builds of the same versioned component (such as a new container build due to a patch in a dependency like AL2). The process to upgrade some components could be different for these two scenarios.

### Order of execution

Components should be **upgraded** **before the new Kubernetes version is rolled out** (if needed). It’s more common for an old component to have problems with a new Kubernetes version than vice-versa. There might be case by case exceptions. Right now there are no exceptions, at least with the current set of core components.

This also facilitates CAPI API version upgrades, since updating its components (CRDs and controllers) will also update the current objects in the cluster (through conversion webhooks). This allows us to only maintain only one API version of provider templates, since these are only reapplied when rolling out the new Kubernetes version.

### How to detect new builds for components?

All the CAPI providers objects in the bundles manifest already contain a `version` field. Unfortunately it only reflects the original upstream version that component was built from. We should add the build metadata information to that semver, like `v0.4.2+build-2`. This will allow us to detect new builds as well as preserve the original upstream version, which is important in order to verify API version changes. We might need to trim the build metadata info when interacting with `clusterctl`, I haven’t tested if it’s supported yet.

We will need to add such field to all the other components (Flux, eks-a controllers and crds, cilium...) that don’t have it yet.

### CAPI

All the CAPI providers, components and cert-manager can be updated with `clusterctl`. This works both with version bumps and new builds of the same version. We just need to create the proper overrides layer, configure the images in the clusterctl config file and pass specific versions for every provider to the `upgrade apply` command.

Even if there is CAPI API version change, the process is the same. The providers will take care of converting the objects in the cluster to the new API version. The only requirement is that all the installed providers need to be upgraded to a version that supports the new API version.

### EKS-A controller and CRDs

Both the controller and CRDs can be upgraded by just applying the new components manifest with kubectl. This should work fine for both new versions and new builds of the same version.

We have to be careful and make sure that any changes to the CRDs are backwards compatible, since at this point there are already EKS Anywhere objects in the cluster. Examples of changes that are not backwards compatible include: adding a required field, changing the type of field, making an existing field required, and so on. It’s important to note that applying the manifest with the new CRD won’t fail, even if the changes are not backwards compatible. But the objects already in the cluster will become invalid. So any kind of update to those (such as flux reconciliation) will be be rejected by the Kubernetes API server.

Updates to the controller code should also take this into consideration. They always need to consider how to handle existing objects in the cluster. For example, if changing the pattern to name certain CAPI objects in the provider templates, the controller has to be able to handle both the old and the new pattern.

If there is a new EKS Anywhere API version, we will need to write conversion webhooks.

### Flux

The process to upgrade the Flux components is very similar to the installation process. It uses the `flux` cli:

1. Write the “kustomization” layer to disk to set our own Flux images
2. Run the `bootstrap` command. This will generate the new manifest (`flux-system/gotk-components.yaml`) and commit all files to the repository
3. Run the `reconcile` command for `flux-system` which tells Flux to pull the manifests from git and upgrade itself

This process should work for both new versions and new builds of the same version.

Full documentation: https://fluxcd.io/docs/installation/#bootstrap-upgrade

### Cilium

The process to upgrade Cilium is a bit more involved. It uses `helm` to generate intermediate manifests and `kubectl` to apply them.

1. Generate pre-flight checks with `helm` and apply them with `kubectl`. This “pre-pulls" the new images on the nodes to avoid `ErrImagePull` errors later during the upgrade
2. Once it’s finished, delete the pre-flight components with `kubectl`
3. Generate the new Cilium manifest using the right “upgrade compatibility” options with `helm` (the generated manifest is different depending on the minor version jump)  and apply it with `kubectl`.

Depending on the implementation, we might be able to avoid creating some/all of these manifests by including them in the bundle manifest.

Full documentation: https://docs.cilium.io/en/v1.10/operations/upgrade/

### CAPI upgrade to `v1alpha4`

As mentioned in the [CAPI](#capi) section, we don’t need any extra process to perform CAPI component upgrades that involve an API version change. However, there will some extra effort needed to support `v1alpha4` in the cli. Most of this will probably apply to future API version changes.

* Add support for `v1alpha4`  in the etcdadm provider. This includes writing conversion webhooks.
* Port all the patches made to upstream cluster-api to `v1alpha4` CRDs and new controller’s code (build on top of latest patch of `v0.4.x`)
* Update all provider templates in the cli to generate `v1alpha4` CAPI specs
* Build compatible versions of dependencies like `clusterctl` and cert-manager

## Customer experience

### Behavior change

There are not backwards breaking changes. However, the expected behavior of the `upgrade cluster` command will be changing, so it should be clearly communicated in the changelog.

### Documentation

We should include a section about how to “get the latest version of components” or something similar.

In the `upgrade cluster` command documentation we should document that we will always upgrade internal eks-a components to the latest available version, even when the Kubernetes version remains unchanged.

## Testing

Since the plan is to only support two minor versions of the cli at the same time, we should add E2E tests that:

* Create the cluster using the previous minor release and upgrade them using the latest version from the current branch
    * For `main`, this makes sure that the next (future) minor release is compatible with clusters created by the current latest release
    * When running this in a release branch, this makes sure that the next (future) patch release is compatible with clusters created by the previous minor release

When building a new bundle release for a certain version we should, at least, run E2E tests that:

* Create a cluster using the current latest release of the cli and the current latest bundle, followed by upgrading the cluster using the same version of the cli and the soon-to-be-released bundles manifest
