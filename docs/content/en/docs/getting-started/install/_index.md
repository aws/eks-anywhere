---
title: Install EKS Anywhere (under construction)
weight: 10
toc_hide: true
hide_summary: true
---

To create an EKS Anywhere (EKS-A) cluster you will need [`eksctl`](https://eksctl.io) and the `eksctl-anywhere` plugin.
This will let you create a cluster in multiple providers for local development or production workloads.

### Local machine prerequisites

- Docker 20.x.x
- Mac OS (10.15) / Ubuntu (20.04.2 LTS)
- 4 CPU cores
- 16GB memory
- 30GB free disk space

> **_NOTE:_** If you are using Ubuntu use the [Docker CE](https://docs.docker.com/engine/install/ubuntu/) installation instructions to install Docker and not the Snap installation.

### Install EKS Anywhere

#### Via Homebrew (macOS and Linux)

You can install `eksctl` and `eksctl-anywhere` with [homebrew](http://homebrew.sh/).

```bash
brew install aws/tap/eks-anywhere
```

#### Manually (macOS and Linux)

Install the latest release of `eksctl`.

```bash
curl "https://github.com/weaveworks/eksctl/releases/latest/download/eksctl_$(uname -s)_amd64.tar.gz" \
    --silent --location \
    | tar xz -C /tmp
sudo mv /tmp/eksctl /usr/local/bin/
```

Install the `eksctl-anywhere` plugin.

```bash
curl "https://github.com/aws/eks-anywhere/releases/latest/download/eksctl-anywhere_$(uname -s)_amd64.tar.gz" \
    --silent --location \
    | tar xz -C /tmp
sudo mv /tmp/eksctl-anywhere /usr/local/bin/
```

### (Optional) Install additional tools

There are some additional tools you may want for your EKS-A clusters.
The EKS Distro project publishes some additional binaries and EKS-A bundles them for usage.

This brew formula includes

* eksctl
* eksctl-anywhere
* kubectl
* aws-iam-authenticator

```bash
brew install aws/tap/eks-anywhere-bundle
```

### Upgrade eksctl-anywhere

If you installed `eksctl-anywhere` via homebrew you can upgrade the binary with

```bash
brew update
brew upgrade eksctl-anywhere
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