---
title: Install EKS Anywhere
weight: 10
---

EKS Anywhere will create and manage Kubernetes clusters on multiple providers.
Currently we support creating development clusters locally using Docker and production clusters using Bare Metal or VMware vSphere.

Creating an EKS Anywhere cluster begins with setting up an Administrative machine where you will run Docker and add some binaries.
From there, you create the cluster for your chosen provider.
See [Create cluster workflow]({{< relref "../../concepts/clusterworkflow" >}}) for an overview of the cluster creation process.

To create an EKS Anywhere cluster you will need [`eksctl`](https://eksctl.io) and the `eksctl-anywhere` plugin.
This will let you create a cluster in multiple providers for local development or production workloads.

### Administrative machine prerequisites

- Docker 20.x.x
- Mac OS (10.15) / Ubuntu (20.04.2 LTS)
- 4 CPU cores
- 16GB memory
- 30GB free disk space
- Administrative machine must be on the same Layer 2 network as the cluster machines (Bare Metal provider only).

   {{% alert title="Note" color="primary" %}}
   * If you are using Ubuntu use the [Docker CE](https://docs.docker.com/engine/install/ubuntu/) installation instructions to install Docker and not the Snap installation.
   * If you are using Docker Desktop, you need to know that:
       * For EKS Anywhere Bare Metal, Docker Desktop is not supported
       * For EKS Anywhere vSphere, if you are using Mac OS Docker Desktop 4.4.2 or newer `"deprecatedCgroupv1": true` must be set in `~/Library/Group\ Containers/group.com.docker/settings.json`.
   * Currently newer versions of Ubuntu (21.10) and other linux distributions with cgroup v2 enabled are not supported.
   {{% /alert %}}


### Install EKS Anywhere CLI tools

#### Via Homebrew (macOS and Linux)

{{% alert title="Warning" color="warning" %}}
EKS Anywhere only works on computers with x86 and amd64 process architecture.
It currently will not work on computers with Apple Silicon or Arm based processors.
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
curl "https://github.com/weaveworks/eksctl/releases/latest/download/eksctl_$(uname -s)_amd64.tar.gz" \
    --silent --location \
    | tar xz -C /tmp
sudo mv /tmp/eksctl /usr/local/bin/
```

Install the `eksctl-anywhere` plugin.

```bash
export EKSA_RELEASE="0.10.1" OS="$(uname -s | tr A-Z a-z)" RELEASE_NUMBER=15
curl "https://anywhere-assets.eks.amazonaws.com/releases/eks-a/${RELEASE_NUMBER}/artifacts/eks-a/v${EKSA_RELEASE}/${OS}/amd64/eksctl-anywhere-v${EKSA_RELEASE}-${OS}-amd64.tar.gz" \
    --silent --location \
    | tar xz ./eksctl-anywhere
sudo mv ./eksctl-anywhere /usr/local/bin/
```

### Upgrade eksctl-anywhere

If you installed `eksctl-anywhere` via homebrew you can upgrade the binary with

```bash
brew update
brew upgrade eks-anywhere
```

If you installed `eksctl-anywhere` manually you should follow the installation steps to download the latest release.

You can verify your installed version with

```bash
eksctl anywhere version
```

## Deploy a cluster

Once you have the tools installed you can deploy a local cluster or production cluster in the next steps.

* [Create local cluster](../local-environment/)
* [Create production cluster](../production-environment/)
