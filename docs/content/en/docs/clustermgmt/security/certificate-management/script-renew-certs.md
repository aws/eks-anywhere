---
title: "Manually Renew Cluster Certificates"
linkTitle: "Script To Renew Certificates"
description: >
  Step-by-step guide to renew Kubernetes certificates on EKS Anywhere clusters using a script.
weight: 20
---

## Overview

Use this script **if your cluster certificates are nearing expiration**.

If any certificate is **already expired**, follow the [manual renewal steps](https://anywhere.eks.amazonaws.com/docs/clustermgmt/security/manually-renew-certs/) instead.

This script automates:

- Certificate renewal for etcd and control plane nodes.
- Safe backup of existing certificates and the `kubeadm-config` ConfigMap.
- Cleanup of temporary files if cetificates are renewed and cluster is healthy.

## Prerequisites

- Admin machine with:
  - `kubectl`, `jq`, `scp`, `ssh`, and `sudo` installed
- SSH access to all control plane and etcd nodes
- Cluster should be reachable and working.

## Steps

### 1. Setup environment variable:

```bash
export KUBECONFIG=<path-to-management-cluster-kubeconfig>
```

### 2. Download the Script

{{< tabpane >}}
{{< tab header="Ubuntu or RHEL" lang="bash" >}}
```bash
curl -O https://raw.githubusercontent.com/aws/eks-anywhere/refs/heads/main/scripts/renew_certificates.sh
chmod +x renew_certificates.sh
```
{{< /tab >}}
{{< tab header="Bottlerocket" lang="bash" >}}
```bash
curl -O https://raw.githubusercontent.com/aws/eks-anywhere/refs/heads/main/scripts/renew_certificates_bottlerocket.sh
chmod +x renew_certificates_bottlerocket.sh
```
{{< /tab >}}
{{< /tabpane >}}

### 3. Run the Script

```bash
./renew_certificates.sh <cluster-name> <ssh-user> <path-to-ssh-private-key>
```


### What the Script Does

- Detects control plane and external etcd nodes.
- Backs up:
    - All etcd certificates (in case of external etcd).
    - Control plane certificates.
    - kubeadm-config ConfigMap.
- Renews external etcd certificates.
- Updates the Kubernetes secret apiserver-etcd-client.
- Renews all kubeadm certificates.
- Restart static control plane pods.
- Cleans up temporary certs and backup folders (only if certificates are renewed successfully and cluster is healthy).