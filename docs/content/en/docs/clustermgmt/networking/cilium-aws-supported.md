---
title: "Cilium supported by AWS"
linkTitle: "Cilium supported by AWS"
weight: 40
description: >
  How to install and manage versions of the Cilium CNI build maintained by AWS for EKS Anywhere clusters
---

## Overview

As of EKS Anywhere v0.24, EKS Anywhere installs an AWS build of Cilium with all the features of open source Cilium as the default Kubernetes CNI plugin. Previous versions of EKS Anywhere included a Cilium build with a limited set of features compared to open source Cilium. With this update, for clusters with an EKS Anywhere Enterprise Subscription license, AWS will continue to provide technical support for Networking Routing (CNI), Identity-Based Network Policy (Labels, CIDR), and Load-Balancing (L3/L4) capabilities of Cilium, but you no longer need to replace the default EKS Anywhere CNI to use other capabilities available with open source Cilium. If you plan to use functionality outside the scope of AWS support, AWS recommends obtaining alternative commercial support for Cilium or have the in-house expertise to troubleshoot and contribute fixes to the Cilium project.

EKS Anywhere supports two approaches for managing the versions of the Cilium build maintained by AWS:

- **AWS build of Cilium with automatic version upgrades:**: EKS Anywhere automatically manages Cilium installation and upgrades (default behavior)
- **AWS build of Cilium with self-managed version upgrades**: You manage Cilium deployment using Helm while using AWS-maintained images from the public ECR registry

The AWS build of Cilium version provides images that include security patches and bug fixes validated by Amazon, available at `public.ecr.aws/eks/cilium/cilium`.

## Installing AWS build of Cilium with automatic version upgrades

If you want to use the full open source Cilium version with advanced features, the transition is seamless and requires no configuration changes. AWS provides support for the subset of Cilium features mentioned in the Overview section above. You can customize Cilium settings using the `helmValues` parameter (see [Helm Values Configuration for Cilium Plugin]({{< relref "../../getting-started/optional/cni#helm-values-configuration-for-cilium-plugin" >}})).

## Installing AWS build of Cilium with self-managed version upgrades

If you want to continue managing the Cilium CNI version yourself while using the AWS build of Cilium, you can install the AWS build of Cilium using Helm.

### Prerequisites

- Helm 3.x installed
- Access to your EKS Anywhere cluster with `kubectl` configured
- Appropriate permissions to manage resources in the `kube-system` namespace

### Installation

Install the AWS build of Cilium version using Helm (replace `1.17.6-0` with your desired version from https://gallery.ecr.aws/eks/cilium/cilium):

```bash
helm install cilium oci://public.ecr.aws/eks/cilium/cilium \
  --version 1.17.6-0 \
  --namespace kube-system \
  --set envoy.enabled=false \
  --set ingressController.enabled=false \
  --set loadBalancer.l7.backend=none
```

### Pros and Cons

**Pros:**
- Access to full Cilium open source features with AWS support for a subset of features
- Security patches and bug fixes handled by AWS
- Control over upgrade timing and version selection
- Greater flexibility to customize Cilium configuration

**Cons:**
- You are responsible for upgrading and validating Cilium versions
- Additional operational overhead for validating that your Cilium version is compatible with your Kubernetes version
- No automatic upgrades during EKS Anywhere cluster upgrades

## Migrating from upstream open source Cilium to AWS build of Cilium

If you're currently running the upstream open source Cilium version and want to switch to the AWS build of Cilium, you can perform an in-place upgrade using Helm. This migration allows you to benefit from images maintained by AWS while preserving your existing Cilium configuration and network policies.

The AWS build of Cilium uses images from the AWS public ECR registry (`public.ecr.aws/eks/cilium/cilium`) and includes security patches and bug fixes validated by AWS. This upgrade is performed as a rolling update, minimizing disruption to your workloads.

To perform the migration, run :

```bash
helm upgrade cilium oci://public.ecr.aws/eks/cilium/cilium \
  --version 1.17.6-0 \
  --namespace kube-system \
  --set envoy.enabled=false \
  --set ingressController.enabled=false \
  --set loadBalancer.l7.backend=none
```

Cilium 1.17.6 is the only Cilium version that AWS currently builds.


After the upgrade, verify the installation status:

```bash
cilium status --wait
```

Expected output showing AWS ECR images:

```
    /¯¯\
 /¯¯\__/¯¯\    Cilium:             OK
 \__/¯¯\__/    Operator:           OK
 /¯¯\__/¯¯\    Envoy DaemonSet:    disabled (using embedded mode)
 \__/¯¯\__/    Hubble Relay:       disabled
    \__/       ClusterMesh:        disabled

DaemonSet              cilium                   Desired: 4, Ready: 4/4, Available: 4/4
Deployment             cilium-operator          Desired: 2, Ready: 2/2, Available: 2/2
Containers:            cilium                   Running: 4
                       cilium-operator          Running: 2
                       clustermesh-apiserver
                       hubble-relay
Cluster Pods:          17/17 managed by Cilium
Helm chart version:    1.17.6-0
Image versions         cilium             public.ecr.aws/eks/cilium/cilium:v1.17.6-0: 4
                       cilium-operator    public.ecr.aws/eks/cilium/operator-generic:v1.17.6-0: 2
```

After aligning your Cilium installation with the AWS-supported version, you can optionally transition management of Cilium upgrades to EKS Anywhere.This allows EKS Anywhere to automatically upgrade your Cilium deployment to the latest AWS build of Cilium for the Kubernetes you are upgrading to, reducing operational overhead.

To complete the transition:

1. **Upgrade Your Cluster**: Upgrade your cluster to the latest EKS Anywhere version.

2. **Toggle the skipUpgrade Configuration**: Modify the `skipUpgrade` setting in your cluster configuration:

   ```bash
   kubectl edit cluster <cluster-name>
   ```

   Choose one of the following options:

   **Option 1**: Remove the `skipUpgrade` line entirely and use default settings:

   ```yaml
   cilium: {}
   ```

   **Option 2**: Explicitly set `skipUpgrade` to false:

   ```yaml
   cilium:
     skipUpgrade: false
   ```

3. **Save and Exit**: Save the configuration file and exit the editor.

Once configured, EKS Anywhere will automatically manage Cilium upgrades during cluster upgrades, ensuring compatibility and reducing manual maintenance.
