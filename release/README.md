# EKS Anywhere Release Tooling

This folder contains the release tooling code for EKS Anywhere. It defines the APIs and CRDs corresponding to the two kinds of release supported by EKS Anywhere - the versioned bundle release and EKS Anywhere CLI release. It also contains logic to perform the release process, that includes downloading artifacts from the source locations, uploading them to release locations, and generating the versioned bundles manifest and EKS Anywhere releases manifest.

## Versioned Bundles Release

The versioned bundles are comprised of version-tagged bundles of EKS-A/EKS-D artifacts. This includes the list of container images and manifests from EKS-A as well as EKS-D dependent artifacts such as OVAs and kind images. These image and S3 URIs will be embedded into a bundles manifest that the CLI will use to fetch the images and manifests that it needs to pull when creating a cluster. The bundles release manifest URI will be referenced in the EKS-A CLI release manifest, so that customers will know what version bundle each release version of the CLI will be supporting.

## EKS-A CLI release

The EKS-A CLI release is the build and release of the tagged version of the EKS-A CLI, along with a reference to the corresponding bundle release manifest. The staging EKS-A release will be kicked off after the staging bundle release, so that this release manifest can reference a bundle manifest that exists in the artifacts S3. After the staging release has been uploaded to S3, we can do integration tests in staging, by obtaining the CLI from the release manifest and running it against the staging versioned bundles. Once the tests pass, all the manifests and artifacts can be moved from staging to prod.

## Testing release tooling changes locally

Changes made to release tooling, such as modifying the release API or adding a new component to the versioned bundle, can change the manifest specs produced during release. To visualize these changes, you can simply run `make dev-release` which will simulate the release process in dry-run mode and generate the resultant versioned bundle and release manifests.
