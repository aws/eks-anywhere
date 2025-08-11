---
title: "Renew certificates using eksctl anywhere"
linkTitle: "Using eksctl anywhere"
weight: 15
description: >
  How to renew EKS Anywhere cluster certificates using the eksctl anywhere CLI
---

## Overview

EKS Anywhere provides a simple and recommended way to renew cluster certificates using the `eksctl anywhere renew certificates` command. This is the recommended approach for certificate renewal. 
Get more information on EKS Anywhere cluster certificates from [Monitoring Certificate Expiration]({{< relref "monitoring-certificates.md" >}})

{{% alert title="Note" color="primary" %}}
This feature is available starting with EKS Anywhere version `v0.23.1`.
{{% /alert %}}


## Prerequisites

- Admin machine with:
  - `eksctl anywhere` CLI installed
  - SSH access to all control plane and etcd nodes

## Configuration File

Create a YAML configuration file that specifies the cluster details and SSH access information. Example configuration file:

```yaml
clusterName: my-cluster
os: ubuntu  # Options: ubuntu, rhel, bottlerocket
controlPlane: 
  nodes: 
  - 192.168.1.10
  - 192.168.1.11
  - 192.168.1.12
  ssh:
    sshKey: /path/to/private/key
    sshUser: ssh-user
etcd: 
  nodes: 
  - 192.168.1.20
  - 192.168.1.21
  - 192.168.1.22
  ssh:
    sshKey: /path/to/private/key
    sshUser: ssh-user
```

### Configuration Fields

#### clusterName (required)
Name of the EKS Anywhere cluster.

#### os (required)
Operating system of the nodes. Permitted values: ubuntu, rhel, bottlerocket.

#### controlPlane.nodes (optional)
List of control plane node IPs. Node IPs can be omitted if the cluster is accessible.

#### controlPlane.ssh.sshKey (required)
Path to SSH private key for control plane nodes.

#### controlPlane.ssh.sshUser (required)
SSH user for control plane nodes.

#### etcd (optional)
Required only if the cluster uses external etcd nodes.

#### etcd.nodes (optional)
List of external etcd node IPs. Node IPs can be omitted if the cluster is accessible.

#### etcd.ssh.sshKey (required if using external etcd)
Path to SSH private key for etcd nodes. Required only if the cluster uses external etcd nodes.

#### etcd.ssh.sshUser (required if using external etcd)
SSH user for etcd nodes. Required only if the cluster uses external etcd nodes.

### Using Password-Protected SSH Keys

If your SSH keys are password protected, you can use environment variables to provide the passphrases instead of including them in the configuration file:

- `EKSA_SSH_KEY_PASSPHRASE_CP`: Passphrase for the control plane SSH key
- `EKSA_SSH_KEY_PASSPHRASE_ETCD`: Passphrase for the etcd SSH key

When using these environment variables, you can leave the `sshKey` field empty in your configuration file.

## Steps to Renew Certificates

1. Create the configuration file as described above (e.g., `cert-renewal-config.yaml`).

2. Run the certificate renewal command:

```bash
eksctl anywhere renew certificates -f cert-renewal-config.yaml
```

You can set the log level verbosity using the `-v` or `--verbosity` flag:

```bash
eksctl anywhere renew certificates -f cert-renewal-config.yaml -v 9
```

You can also specify which component's certificates to renew using the `--component` flag:

```bash
# Renew only control plane certificates
eksctl anywhere renew certificates -f cert-renewal-config.yaml --component control-plane

# Renew only etcd certificates
eksctl anywhere renew certificates -f cert-renewal-config.yaml --component etcd
```

This is useful when you want to renew certificates for only specific components rather than all components at once.

### Renew certificates for a cluster with accessible nodes

For clusters that are accessible via kubectl, follow these steps:

1. Set the KUBECONFIG environment variable:
   ```bash
   export KUBECONFIG=mgmt/mgmt-eks-a-cluster.kubeconfig
   ```

2. Create a simplified configuration file without node IPs:
   ```yaml
   clusterName: my-cluster
   os: ubuntu
   controlPlane:
     ssh:
       sshKey: /path/to/private/key
       sshUser: ssh-user
   etcd:
     ssh:
       sshKey: /path/to/private/key
       sshUser: ssh-user
   ```

3. Run the certificate renewal command:
   ```bash
   eksctl anywhere renew certificates -f cert-renewal-config.yaml
   ```

## What the command does

The `eksctl anywhere renew certificates` command automates the certificate renewal process by:

1. Connecting to each node via SSH
2. Backing up existing certificates
3. For external etcd nodes:
   - Regenerating etcd certificates
   - Verifying etcd health
4. For control plane nodes:
   - Renewing all kubeadm certificates
   - Restarting static pods
   - Updating external etcd key cert (if present)
   - Verifying API server health
5. Verifying overall cluster health
