# Manifests: Release and Bundles

## Release manifest
The Release manifest contains all the available EKS-A releases. Each entry contains information about the version (sem ver for the release), URL to the proper Bundles manifest for that release and url to download the CLI binary for the supported architectures and platforms.

## Bundles manifest
The Bundles manifest contains all the EKS-A components, including management (Cluster-API, EKS-A controller, CRDs, etc.) and cluster components (EKS-D, CNI, CCM, etc.). Each Bundles manifest contains multiple VersionsBundles, once per supported Kubernetes version for that specific release. A VersionsBundles groups together all the components that should be used in a cluster for a particular Kubernetes version.

Each EKS-A release has a unique Bundles manifest. In order to bump a EKS-A component to a new version, a new Bundles needs to be created and shipped through a new EKS-A patch release.

## CLI startup flow
Each CLI is built with a particular EKS-A semver in its metadata. This pins each binary built to a particular EKS-A release. However, the Release and Bundles manifests are not included or shipped with the binary. These are served through cloudfront and the CLI dynamically fetches them during each command startup. The logic is as follows:
1. Fetch the Release manifest. As part of its metadata, the CLI is built with a static URL (is the same for all EKS-A releases) pointing to this manifest. Dev CLIs will use a different URL than prod CLIs.
1. Iterate over the available EKS-A releases in the manifest and look for an exact version match with semver included in the build metadata of the binary.
1.
	1. Select the exact match release and continue.
	1. Or return an error if the release it not found. This should never happen for prod releases. Until now, we keep all releases in the manifest from the beginning of times. At some point we might stop dropping very old releases. That would make those old binaries fail.
1. Grab the Bundles URL for the selected release and download it. This components in this Bundles will be used to create/upgrade the cluster during this command execution by selecting a VersionsBundle using the desired cluster's `spec.kubernetesVersion`.

### Dev releases
Dev releases are a bit special: we generate new them all the time, very fast. For this reason, we don't use a simple major.minor.patch semver, but we include build metadata. In particular we use `v{major}.{minor}.{patch}-dev+build.{number}` with `number` being a monotonically increasing integer that is bumped every time a new dev release is built.

The version we use for the first part depends on the HEAD: `main` vs release branches:
- For `main`, we use the next minor version to the latest tag available. For example, if the latest prod release is `v0.18.5`, the version used for dev releases will be `v0.19.0-dev+build.{number}`. This aligns with the fact that the code in `main` belongs to the next future prod release `v0.19.0`.
- For `release-*` branches, we use the next patch version to the latest available tag for that minor version. For example, for `release-0.17`, if the latest latest prod release is for v0.17 is `v0.17.7`, dev releases will follow `v0.17.8-dev+build.{number}`.

In order to avoid the dev Release manifest growing forever, we trim the included releases to a max size, dropping always the oldest one. Take this in mind if using a particular version locally. If you do it for too long, it might become unavailable. If it does, just rebuild your CLI.

### E2E tests
When a CLI is built for dev E2E tests, it's given the latest available EKS-A dev version. This pins that particular binary to one EKS-A version and hence one Bundles manifest. This is important in order to guarantee that the same Bundles is used for the whole execution of the test pipelines and avoids a race condition where a new EKS-A dev release and Bundles are published during an E2E test run.

### Locally building the CLI
When writing and testing code for the CLI/Controller, most of the time we don't care about particular releases and we just want to use the latest available Bundles that contains the latest available set of components. this verifies that our changes are compatible with the current state of EKS-A dependencies.

To avoid having to rebuild the CLI every time we want to refresh the pulled Bundles or even having to care about fetching the latest version, we introduced a special build metadata identifier `+latest`. This instructs the CLI to not look for an exact match with an EKS-A version, but select the newest one that matches our pre-release. For example: if the release manifest has two releases [`v0.19.0-dev+build.1234`, `v0.19.0-dev+build.1233`], then if the CLI has version `v0.19.0-dev+latest`, then the release `v0.19.0-dev+build.1234` will be selected.

This is the default behavior when building a CLI locally: the Makefile will calculate the appropriate major.minor.patch based on the current HEAD and its closest branch ancestor (either `main` or a `release-*` branch). If you wish to pin your local CLI to a particular version, pass the `DEV_GIT_VERSION` to the make target.