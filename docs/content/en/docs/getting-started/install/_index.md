---
title: 1. Admin Machine
weight: 10
description: >
  Steps for setting up the Admin Machine
---

{{% alert title="Warning" color="warning" %}}
The Administrative machine (Admin machine) is required to run cluster lifecycle operations, but EKS Anywhere clusters do not require a continuously running Admin machine to function. During cluster creation, critical cluster artifacts including the kubeconfig file, SSH keys, and the full cluster specification yaml are saved to the Admin machine. These files are required when running any subsequent cluster lifecycle operations. For this reason, it is recommended to save a backup of these files and to use the same Admin machine for all subsequent cluster lifecycle operations.
{{% /alert %}}

EKS Anywhere will create and manage Kubernetes clusters on multiple providers.
Currently we support creating development clusters locally using Docker and production clusters from providers listed on the [providers]({{< relref "/docs/getting-started/chooseprovider/" >}}) page.

Creating an EKS Anywhere cluster begins with setting up an Administrative machine where you run all EKS Anywhere lifecycle operations as well as Docker, `kubectl` and prerequisite utilites.
From here you will need to install [`eksctl`](https://eksctl.io), a CLI tool for creating and managing clusters on EKS, and the [`eksctl-anywhere`](/docs/reference/eksctl/anywhere/) plugin, an extension to create and manage EKS Anywhere clusters on-premises, on your Administrative machine.
You can then proceed to the [cluster networking]({{< relref "../ports" >}}) and [provider specific steps]({{< relref "../chooseprovider" >}}). 
See [Create cluster workflow]({{< relref "../overview" >}}) for an overview of the cluster creation process.

>**_NOTE:_** For Snow provider, if you ordered a Snowball Edge device with EKS Anywhere enabled, it is preconfigured with an Admin AMI which contains the necessary binaries, dependencies, and artifacts to create an EKS Anywhere cluster. Skip to the steps on [Create Snow production cluster]({{< relref "../snow/snow-getstarted" >}})to get started with EKS Anywhere on Snow.

### Administrative machine prerequisites

#### System and network requirements
- Mac OS 10.15+ / Ubuntu 20.04.2 LTS or 22.04 LTS / RHEL or Rocky Linux 8.8+
- 4 CPU cores
- 16GB memory
- 30GB free disk space
- If you are running in an airgapped environment, the Admin machine must be amd64.
- If you are running EKS Anywhere on bare metal, the Admin machine must be on the same Layer 2 network as the cluster machines.

Here are a few other things to keep in mind:

* If you are using Ubuntu, use the Docker CE installation instructions to install Docker and not the Snap installation, as described [here.](https://docs.docker.com/engine/install/ubuntu/)

* If you are using EKS Anywhere v0.15 or earlier and Ubuntu 21.10 or 22.04, you will need to switch from _cgroups v2_ to _cgroups v1_. For details, see [Troubleshooting Guide.]({{< relref "../../troubleshooting/troubleshooting.md#for-eks-anywhere-v015-and-earlier-cgroups-v2-is-not-supported-in-ubuntu-2110-and-2204" >}})

* If you are using Docker Desktop, you need to know that:

  * For EKS Anywhere Bare Metal, Docker Desktop is not supported
  * For EKS Anywhere vSphere, if you are using EKS Anywhere v0.15 or earlier and Mac OS Docker Desktop 4.4.2 or newer `"deprecatedCgroupv1": true` must be set in `~/Library/Group\ Containers/group.com.docker/settings.json`.

#### Tools
- [Docker 20.x.x](https://docs.docker.com/engine/install/)
- [`curl`](https://everything.curl.dev/get)
- [`yq`](https://github.com/mikefarah/yq/#install)

### Install EKS Anywhere CLI tools

#### Via Homebrew (macOS and Linux)

{{% alert title="Warning" color="warning" %}}
EKS Anywhere works on computers with x86 and amd64 process architecture.
From v0.18.1 it also works with Apple Silicon or Arm based processors.
{{% /alert %}}

You can install `eksctl` and `eksctl-anywhere` with [homebrew](http://brew.sh/).
This package will also install `kubectl` and the `aws-iam-authenticator` which will be helpful to test EKS Anywhere clusters.

```bash
brew install aws/tap/eks-anywhere
```

#### Manually (macOS and Linux)

Install the latest release of `eksctl`.
The EKS Anywhere plugin requires `eksctl` version 0.66.0 or newer.

```bash
curl "https://github.com/eksctl-io/eksctl/releases/latest/download/eksctl_$(uname -s)_amd64.tar.gz" \
    --silent --location \
    | tar xz -C /tmp
sudo install -m 0755 /tmp/eksctl /usr/local/bin/eksctl
```

Install the `eksctl-anywhere` plugin.

```bash
RELEASE_VERSION=$(curl https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml --silent --location | yq ".spec.latestVersion")
EKS_ANYWHERE_TARBALL_URL=$(curl https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml --silent --location | yq ".spec.releases[] | select(.version==\"$RELEASE_VERSION\").eksABinary.$(uname -s | tr A-Z a-z).uri")
curl $EKS_ANYWHERE_TARBALL_URL \
    --silent --location \
    | tar xz ./eksctl-anywhere
sudo install -m 0755 ./eksctl-anywhere /usr/local/bin/eksctl-anywhere
```

Install the `kubectl` Kubernetes command line tool.
This can be done by following the instructions [here](https://kubernetes.io/docs/tasks/tools/#kubectl).

Or you can install the latest kubectl directly with the following.

```bash
export OS="$(uname -s | tr A-Z a-z)" ARCH=$(test "$(uname -m)" = 'x86_64' && echo 'amd64' || echo 'arm64')
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/${OS}/${ARCH}/kubectl"
sudo install -m 0755 ./kubectl /usr/local/bin/kubectl
```

### Upgrade eksctl-anywhere

If you installed `eksctl-anywhere` via homebrew you can upgrade the binary with

```bash
brew update
brew upgrade aws/tap/eks-anywhere
```

If you installed `eksctl-anywhere` manually you should follow the installation steps to download the latest release.

You can verify your installed version with

```bash
eksctl anywhere version
```

## Prepare for airgapped deployments (optional)

For more information on how to prepare the Administrative machine for airgapped environments, go to the [Airgapped](/docs/getting-started/airgapped/) page. 

## Deploy a cluster

Once you have the tools installed, go to the [EKS Anywhere providers]({{< relref "/docs/getting-started/chooseprovider" >}}) page for instructions on creating a cluster on your chosen provider.
