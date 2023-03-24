---
title: Install EKS Anywhere
weight: 10
---

EKS Anywhere will create and manage Kubernetes clusters on multiple providers.
Currently we support creating development clusters locally using Docker and production clusters from providers listed on the [Create production cluster]({{< relref "/docs/getting-started/production-environment" >}}) page.

Creating an EKS Anywhere cluster begins with setting up an Administrative machine where you will run Docker and add some binaries.
From there, you create the cluster for your chosen provider.
See [Create cluster workflow]({{< relref "../../concepts/clusterworkflow" >}}) for an overview of the cluster creation process.

To create an EKS Anywhere cluster you will need [`eksctl`](https://eksctl.io) and the `eksctl-anywhere` plugin.
This will let you create a cluster in multiple providers for local development or production workloads.

>**_NOTE:_** For Snow provider, the Snow devices will come with a pre-configured Admin AMI which can be used to create an Admin instance with all the necessary binaries, dependencies and artifacts to create an EKS Anywhere cluster. Skip the below steps and see [Create Snow production cluster]({{< relref "../production-environment/snow-getstarted" >}}) to get started with EKS Anywhere on Snow.

### Administrative machine prerequisites

- Docker 20.x.x
- Mac OS 10.15 / Ubuntu 20.04.2 LTS (See Note on newer Ubuntu versions)
- 4 CPU cores
- 16GB memory
- 30GB free disk space
- Administrative machine must be on the same Layer 2 network as the cluster machines (Bare Metal provider only).

If you are using Ubuntu, use the Docker CE installation instructions to install Docker and not the Snap installation, as described [here.](https://docs.docker.com/engine/install/ubuntu/)

If you are using Ubuntu 21.10 or 22.04, you will need to switch from _cgroups v2_ to _cgroups v1_. For details, see [Troubleshooting Guide.]({{< relref "../../tasks/troubleshoot/troubleshooting.md#cgroups-v2-is-not-supported-in-ubuntu-2110-and-2204" >}})

If you are using Docker Desktop, you need to know that:

* For EKS Anywhere Bare Metal, Docker Desktop is not supported
* For EKS Anywhere vSphere, if you are using Mac OS Docker Desktop 4.4.2 or newer `"deprecatedCgroupv1": true` must be set in `~/Library/Group\ Containers/group.com.docker/settings.json`.


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
export EKSA_RELEASE="0.14.3" OS="$(uname -s | tr A-Z a-z)" RELEASE_NUMBER=30
curl "https://anywhere-assets.eks.amazonaws.com/releases/eks-a/${RELEASE_NUMBER}/artifacts/eks-a/v${EKSA_RELEASE}/${OS}/amd64/eksctl-anywhere-v${EKSA_RELEASE}-${OS}-amd64.tar.gz" \
    --silent --location \
    | tar xz ./eksctl-anywhere
sudo mv ./eksctl-anywhere /usr/local/bin/
```

Install the `kubectl` Kubernetes command line tool.
This can be done by following the instructions [here](https://kubernetes.io/docs/tasks/tools/).

Or you can install the latest kubectl directly with the following.

```bash
export OS="$(uname -s | tr A-Z a-z)" ARCH=$(test "$(uname -m)" = 'x86_64' && echo 'amd64' || echo 'arm64')
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/${OS}/${ARCH}/kubectl"
sudo mv ./kubectl /usr/local/bin
sudo chmod +x /usr/local/bin/kubectl
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

## Prepare for airgapped deployments (optional)

When creating an EKS Anywhere cluster, there may be times where you need to do so in an airgapped
environment.
In this type of environment, cluster nodes are connected to the Admin Machine, but not to the
internet.
In order to download images and artifacts, however, the Admin machine needs to be temporarily
connected to the internet.

An airgapped environment is especially important if you require the most secure networks.
EKS Anywhere supports airgapped installation for creating clusters using a registry mirror.
For airgapped installation to work, the Admin machine must have:

* Temporary access to the internet to download images and artifacts
* Ample space (80 GB or more) to store artifacts locally


To create a cluster in an airgapped environment, perform the following:

1. Download the artifacts and images that will be used by the cluster nodes to the Admin machine using the following command:
   ```bash
   eksctl anywhere download artifacts
   ```
   A compressed file `eks-anywhere-downloads.tar.gz` will be downloaded.

1. To decompress this file, use the following command:
   ```bash
   tar -xvf eks-anywhere-downloads.tar.gz
   ```
   This will create an eks-anywhere-downloads folder that we’ll be using later.

1. In order for the next command to run smoothly, ensure that Docker has been pre-installed and is running. Then run the following:
   ```bash
   eksctl anywhere download images -o images.tar
   ```

   **For the remaining steps, the Admin machine no longer needs to be connected to the internet or the bastion host.**

1. Next, you will need to set up a local registry mirror to host the downloaded EKS Anywhere images. In order to set one up, refer to [Registry Mirror configuration.]({{< relref "/docs/reference/clusterspec/optional/registrymirror.md" >}})

1. Now that you’ve configured your local registry mirror, you will need to import images to the local registry mirror using the following command (be sure to replace <registryUrl> with the url of the local registry mirror you created in step 4):
   ```bash
   eksctl anywhere import images -i images.tar -r <registryUrl> \
      -- bundles ./eks-anywhere-downloads/bundle-release.yaml
   ```
You are now ready to deploy a cluster by following instructions to [Create local cluster]({{< relref "/docs/getting-started/local-environment/" >}}) or  [Create production cluster.]({{< relref "/docs/getting-started/production-environment/" >}}) See text below for specific provider instructions.

### For Bare Metal (Tinkerbell)
You will need to have hookOS and its OS artifacts downloaded and served locally from an HTTP file server.
You will also need to modify the [hookImagesURLPath]({{< relref "../../reference/clusterspec/baremetal/#hookimagesurlpath" >}}) and the [osImageURL]({{< relref "../../reference/clusterspec/baremetal/#osimageurl" >}}) in the cluster configuration files.
Ensure that structure of the files is set up as described in [hookImagesURLPath.]({{< relref "../../reference/clusterspec/baremetal/#hookimagesurlpath" >}})

### For vSphere
If you are using the vSphere provider, be sure that the requirements in the
[Prerequisite checklist]({{< relref "../../reference/vsphere/vsphere-prereq" >}}) have been met.

## Deploy a cluster

Once you have the tools installed you can deploy a local cluster or production cluster in the next steps.

* [Create local cluster]({{< relref "/docs/getting-started/local-environment" >}})
* [Create production cluster]({{< relref "/docs/getting-started/production-environment" >}})
