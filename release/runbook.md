# EKS Anywhere Release Runbook

The following steps are to be followed when performing a full EKS Anywhere release (bundle + EKS-A CLI).

## Bundle Release

### Create Development Versioned Bundle Release PR

* release/triggers/bundle-release/development/BUNDLE_NUMBER
* release/triggers/bundle-release/development/CLI_MIN_VERSION
* release/triggers/bundle-release/development/CLI_MAX_VERSION

Merging this PR will trigger
1. Builds of artifacts for all the upstream dependencies of EKS-A.
2. Upload of artifacts and bundle manifest to the staging S3 bucket and public ECR repositories.

### Create Production Versioned Bundle Release PR

* release/triggers/bundle-release/production/BUNDLE_NUMBER
* release/triggers/bundle-release/production/CLI_MIN_VERSION
* release/triggers/bundle-release/production/CLI_MAX_VERSION

Merging this PR will trigger
1. Download of artifacts from staging S3 bucket and public ECR repositories.
2. Upload of artifacts and bundle manifest to the production S3 buckets and public ECR repositories.

**Note:** Since we carry out the bundle release first, it is possible that the CLI version mentioned in the CLI max version file has not been released yet. After a successful bundle release, we shall create a PR for the EKS-A CLI staging release, so we use the same version in the CLI max version file as the version we will eventually be releasing (the RELEASE_VERSION file below).

## EKS Anywhere CLI release

### Create Development EKS-A Release PR

* release/triggers/eks-a-release/development/RELEASE_NUMBER
* release/triggers/eks-a-release/development/RELEASE_VERSION

Merging this PR will trigger
1. Build of EKS Anywhere CLI from the branch specified.
2. Integration test of the CLI with the previously released bundle manifest.
3. Upload of CLI tarball and EKS-A releases manifest to the staging S3 bucket.

### Create Production EKS-A Release PR

* release/triggers/eks-a-release/production/RELEASE_NUMBER
* release/triggers/eks-a-release/production/RELEASE_VERSION

Merging this PR will trigger
1. Download of CLI tarball from the staging S3 bucket.
2. Upload of CLI tarball and EKS-A releases manifest to the production S3 bucket.

## Release Steps on EKS Anywhere Repository

Once we finish the production EKS-A release, we can mark the release accordingly on the repository.This includes adding a tag as well as a separate branch and updating our website to point to the new branch.

* Update the [Changelog](docs/content/en/docs/reference/changelog.md) with changes for the release and rename `[Unreleased]` to the release version.
  * If this is not just a patch release, update the value under `versions` in the docs [config](docs/config.toml).
  * Follow format [here](https://keepachangelog.com/).
* Tag the commit with the release version (ex. `v0.6.0`).
* Create a branch from the tag
* Have one of the maintainers, who has access to the infrastructure cdk package, change the website configuration to point to the newly created branch
* Add a file to the branch called `PUBLISHED_VERSION` to trigger a change to the infrastructure to deploy the new changes

Now the [homepage](https://anywhere.eks.amazonaws.com/) for EKS-A should be updated with the changes from this new branch
