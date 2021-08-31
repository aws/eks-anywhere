# EKS Anywhere Release Runbook

EKS Anywhere releases are of two variants - the versioned bundle release and the EKS-A CLI release.

### Versioned Bundles Release

The versioned bundles are comprised of version-tagged bundles of EKS-A/EKS-D components. This includes the list of container images/manifests from EKS-A as well as EKS-D dependent artifacts such as OVAs and kind images. These image and S3 URIs will be embedded into a bundles manifest that the CLI will use to fetch the images and manifests that it needs to pull when creating a cluster. The bundles release manifest URI will be referenced in the EKS-A CLI release manifest, so that customers will know what version bundle each release version of the CLI will be supporting.

### EKS-A CLI release

The EKS-A CLI release is the build and release of the tagged version of the EKS-A CLI, along with a reference to the corresponding bundle release manifest. The staging EKS-A release will be kicked off after the staging bundle release, so that this release manifest can reference a bundle manifest that exists in the artifacts S3. After the staging release has been uploaded to S3, we can do integration tests in staging, by obtaining the CLI from the release manifest and running it against the staging versioned bundles. Once the tests pass, all the manifests and artifacts can be moved from staging to prod.

## Staging Release

### Create Development Versioned Bundle Release PR

* release/triggers/bundle-release/development/BUNDLE_NUMBER
* release/triggers/bundle-release/development/CLI_MIN_VERSION
* release/triggers/bundle-release/development/CLI_MAX_VERSION

When the above PR gets merged, it will build all the upstream dependencies of EKS-A after which the release tool will pull them down and upload them to the Artifacts beta S3 and public ECR destinations. 

**Note:** Since we carry out the bundle release first, it is possible that the CLI version mentioned in the CLI max version file has not been released yet. After a successful bundle release, we shall create a PR for the EKS-A CLI staging release, so we use the same version in the CLI max version file as the version we will eventually be releasing (the RELEASE_VERSION file below).

### Create Development EKS-A Release PR

* release/triggers/eks-a-release/development/RELEASE_NUMBER
* release/triggers/eks-a-release/development/RELEASE_VERSION



