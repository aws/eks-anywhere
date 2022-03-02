---
title: Install EKS Anywhere
weight: 10
---

EKS Anywhere will create and manage Kubernetes clusters on multiple providers.
Currently we support creating development clusters locally with Docker and production clusters using VMware vSphere.
Other deployment targets will be added in the future, including bare metal support in 2022.

Creating an EKS Anywhere cluster begins with setting up an Administrative machine where you will run Docker and add some binaries.
From there, you create the cluster for your chosen provider.
See [Create cluster workflow]({{< relref "../../concepts/clusterworkflow" >}}) for an overview of the cluster creation process.

To create an EKS Anywhere cluster you will need [`eksctl`](https://eksctl.io) and the `eksctl-anywhere` plugin.
This will let you create a cluster in multiple providers for local development or production workloads.

To create an EKS Anywhere cluster you will need [`eksctl`](https://eksctl.io), the `eksctl-anywhere` plugin, `kubectl`, `Docker` and `Helm`. This will let you create a cluster in multiple providers for local development or production workloads, and deploying workloads to your clusters. 

> **_NOTE:_** You can install `eksctl-anywhere` on Linux, macOS. Windows is not supported at the time of writing this document. 

In addition, you would need to install `Go`, `kind`, `govc`, `jq` and `flux`. `Go` will be needed when it comes enabling IAM Role for Service Accounts on your EKS Anywhere cluster, to generate the keys.json file. `kind` and `govc` will be used for some troubleshooting tasks. `jq` to query JSON output of some command and pipe the result to another command. `flux` would be needed if you need to enable GitOps on your cluster, but with a self hosted Github (which is not supported at the time being with the addon GitOps which is on the roadmap to be supported soon)


### Administrative machine prerequisites

- Docker 20.x.x
- Mac OS (10.15) / Ubuntu (20.04.2 LTS)
- 4 CPU cores
- 16GB memory
- 30GB free disk space

> **_NOTE:_** If you are using Ubuntu use the [Docker CE](https://docs.docker.com/engine/install/ubuntu/) installation instructions to install Docker and not the Snap installation.

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

And you can also install `helm`, `govc`, `kind`, `jq`, `go` and `flux` with [homebrew](http://brew.sh/).

```bash
brew install helm
brew install govc
brew install kind
brew install jq
brew install go
brew install flux
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
export EKSA_RELEASE="0.7.0" OS="$(uname -s | tr A-Z a-z)" RELEASE_NUMBER=5
curl "https://anywhere-assets.eks.amazonaws.com/releases/eks-a/${RELEASE_NUMBER}/artifacts/eks-a/v${EKSA_RELEASE}/${OS}/amd64/eksctl-anywhere-v${EKSA_RELEASE}-${OS}-amd64.tar.gz" \
    --silent --location \
    | tar xz ./eksctl-anywhere
sudo mv ./eksctl-anywhere /usr/local/bin/
```

Install `kubectl`.

```bash
sudo curl --silent --location -o /usr/local/bin/kubectl \
   https://amazon-eks.s3.us-west-2.amazonaws.com/1.21.2/2021-07-05/bin/linux/amd64/kubectl
sudo chmod +x /usr/local/bin/kubectl
```

Install `helm`.

```bash
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
sudo chmod 700 get_helm.sh
./get_helm.sh
```

Install `go`.

```bash
wget https://go.dev/dl/go1.17.7.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.17.7.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go version
```

Install `jq`.

```bash
wget https://go.dev/dl/go1.17.7.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.17.7.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go version
```

Install Flux (if a self hosted Github is the only option available)

```bash
curl -s https://fluxcd.io/install.sh | sudo bash
```

Troubleshooting tools are needed on the admin machine as well to enable you troubleshooting the cluster later on, if needed:

The first tool called `govc`, which is a vSphere CLI, that you can use to make vSphere API calls. You can download and install it like this:

```bash
curl -L -o \
- "https://github.com/vmware/govmomi/releases/latest/download/govc_$(uname -s)_$(uname -m).tar.gz" \
| tar -C /usr/local/bin -xvzf - govc
```

The second tool called `kind`, which is  a tool for running local Kubernetes clusters using Docker container “nodes”, that runs during the bootstrap process. You can download and install it like this:

```bash
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-linux-amd64
chmod +x ./kind
mv ./kind /some-dir-in-your-PATH/kind
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
