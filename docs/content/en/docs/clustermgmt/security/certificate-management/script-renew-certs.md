---
title: "Manually renew cluster certificates"
linkTitle: "Script To Renew Certificates"
description: >
  Step-by-step guide to renew Kubernetes certificates on EKS Anywhere clusters using a script.
weight: 20
---

Certificates for external etcd and control plane nodes expire after 1 year in EKS Anywhere. EKS Anywhere automatically rotates these certificates when new machines are rolled out in the cluster. New machines are rolled out during cluster lifecycle operations such as `upgrade`. If you upgrade your cluster at least once a year, as recommended, manual rotation of cluster certificates will not be necessary.

This page shows the process for manually rotating certificates if you have not upgraded your cluster in 1 year.

The following table lists the cluster certificate files:

| etcd node             | control plane node       |
|-----------------------|--------------------------|
| apiserver-etcd-client | apiserver-etcd-client    |
| ca                    | ca                       |
| etcdctl-etcd-client   | front-proxy-ca           |
| peer                  | sa                       |
| server                | etcd/ca.crt              |
|                       | apiserver-kubelet-client |
|                       | apiserver                |
|                       | front-proxy-client       |

Commands below can be used for quickly checking your certificates' expiration dates:

```bash
# The expiry time of api-server certificate on you cp node
echo | openssl s_client -connect ${CONTROL_PLANE_IP}:6443 2>/dev/null | openssl x509 -noout -dates

# The expiry time of certificate used by your external etcd server, if you configured one
echo | openssl s_client -connect ${EXTERNAL_ETCD_IP}:2379 2>/dev/null | openssl x509 -noout -dates
```

{{% alert title="Note" color="primary" %}}
You can rotate certificates by following the steps given below. You cannot rotate the `ca` certificate because it is the root certificate. Note that the commands used for Bottlerocket nodes are different than those for Ubuntu and RHEL nodes.
{{% /alert %}}

## Overview
This script automates:

- Certificate renewal for etcd and control plane nodes
- Cleanup of temporary files if certificates are renewed and cluster is healthy

## Prerequisites

- Admin machine with:
  - `kubectl`, `yq`, `jq`, `scp`, `ssh`, and `sudo` installed
- SSH access to all control plane and etcd nodes

## Steps

### 1. Setup environment variable:

```bash
export KUBECONFIG=<path-to-management-cluster-kubeconfig>
```

### 2. Prepare below config yaml file by adding node and private key information of your control plane and/or external etcd to a file, such as `keys-config.yaml`:

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

### 3. Download the Script

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

### 4. Run the Script

```bash
./renew_certificates.sh -f keys-config.yaml
```


### What the Script Does

- Backs up:
    - All etcd certificates (in case of external etcd)
    - Control plane certificates
- Renews external etcd certificates
- Updates the Kubernetes secret apiserver-etcd-client
- Renews all kubeadm certificates
- Restart static control plane pods
- Cleans up temporary certs and backup folders (only if certificates are renewed successfully and cluster is healthy)