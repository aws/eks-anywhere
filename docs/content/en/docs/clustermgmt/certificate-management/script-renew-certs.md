---
title: "Script to renew cluster certificates"
linkTitle: "Script to renew certificates"
description: >
  Step-by-step guide to renew Kubernetes certificates on EKS Anywhere clusters using a script
weight: 20
---

Get more information on EKS Anywhere cluster certificates from [here]({{< relref "monitoring-certificates.md" >}})

This script automates:

- Certificate renewal for etcd and control plane nodes
- Cleanup of temporary files if certificates are renewed and cluster is healthy

### Prerequisites

- Admin machine with:
  - `kubectl`, `yq`, `jq`, `scp`, `ssh`, and `sudo` installed
- SSH access to all control plane and etcd nodes

### Steps

1. Setup environment variable:

```bash
export KUBECONFIG=<path-to-management-cluster-kubeconfig>
```

2. Prepare a `keys-config.yaml` file

Add node and private key information of your control plane and/or external etcd to a file, such as `keys-config.yaml`:

```bash
clusterName: <cluster-name>
controlPlane:
  nodes:
  - <control-plane-1-ip>
  - <control-plane-2-ip>
  - <control-plane-3-ip>
  sshKey: <complete-path-to-private-ssh-key>
  sshUser: <ssh-user>
etcd:
  nodes:
  - <external-etcd-1-ip>
  - <external-etcd-2-ip>
  - <external-etcd-3-ip>
  sshKey: <complete-path-to-private-ssh-key>
  sshUser: <ssh-user>
```

3. Download the Script

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

4. Run the Script as a `sudo` user

```bash
sudo ./renew_certificates.sh -f keys-config.yaml
```


### What the Script Does

- Backs up:
    - All etcd certificates (in case of external etcd)
    - Control plane certificates
- Renews external etcd certificates
- Updates the Kubernetes secret `apiserver-etcd-client` if api server is reachable
- Renews all kubeadm certificates
- Restarts static control plane pods
- Cleans up temporary certs and backup folders (only if certificates are renewed successfully and cluster is healthy)