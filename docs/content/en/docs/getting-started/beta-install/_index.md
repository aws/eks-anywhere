---
title: Install EKS Anywhere for Beta
weight: 5
---

To create an EKS Anywhere (EKS-A) cluster you will need [`eksctl`](https://eksctl.io) and the `eks-a` binaries.
These binaries are provided as part of the beta package. 
This will let you create a cluster in multiple providers for local development or production workloads.

### Local machine prerequisites

- Docker 20.x.x
- Operating System:
    - Mac OS (10.15) or
    - Ubuntu (20.04.2 LTS)
- 4 CPU cores
- 16GB memory
- 30GB free disk space

> **_NOTE:_** If you are using Ubuntu, use the installation instructions below to install Docker and not the Snap installation.

### Set up your Ubuntu administrative machine

All of these commands should be run from your administrative machine

```bash
sudo apt update
sudo apt install -y docker.io
sudo usermod -a -G docker $USER
wget https://distro.eks.amazonaws.com/kubernetes-1-21/releases/4/artifacts/kubernetes/v1.21.2/bin/linux/amd64/kubectl
mkdir -p $HOME/bin
chmod +x kubectl
mv kubectl $HOME/bin/
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc
```

Now you need to log out and log back into the system to get the correct group permissions and path.

### Install EKS Anywhere

To install the `eksctl` and `eks-a` commands you can extract the pre-built binaries from the provided archive.

Give the binaries execute permission and move them into your `$HOME/bin` folder.

```bash
chmod +x ./eksctl
chmod +x ./eks-a
mv ./eksctl $HOME/bin/
mv ./eks-a $HOME/bin/
```

### Verify installation

You can verify your installed version with

```bash
eksctl anywhere version
```

## Deploy a cluster

Once you have the tools installed you can deploy a local cluster or production cluster in the next steps.

* [Create local cluster](../local-environment/)
* [Create production cluster](../production-environment/)
