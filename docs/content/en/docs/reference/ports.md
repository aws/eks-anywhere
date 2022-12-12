---
title: "Ports and protocols"
weight: 60
description: >
  Ports used with an EKS Anywhere cluster
---

EKS Anywhere requires that various ports on control plane and worker nodes be open.
Some Kubernetes-specific ports need open access only from other Kubernetes nodes, while others are exposed externally.
Beyond Kubernetes ports, someone managing an EKS Anywhere cluster must also have external access to ports on the underlying EKS Anywhere provider (such as VMware) and to external tooling (such as Jenkins).

If you are responsible for network firewall rules between nodes on your EKS Anywhere clusters, the following tables describe both Kubernetes and EKS Anywhere-specific ports you should be aware of.

## Kubernetes control plane
The following table represents the ports published by the Kubernetes project that must be accessible on any Kubernetes control plane.


| Protocol | Direction | Port Range | Purpose                 | Used By                   |
|----------|-----------|------------|-------------------------|---------------------------|
| TCP      | Inbound   | 6443       | Kubernetes API server   | All                       |
| TCP      | Inbound   | 10250      | Kubelet API             | Self, Control plane       |
| TCP      | Inbound   | 10259      | kube-scheduler          | Self                      |
| TCP      | Inbound   | 10257      | kube-controller-manager | Self                      |

Although etcd ports are included in control plane section, you can also host your own
etcd cluster externally or on custom ports. 

| Protocol | Direction | Port Range | Purpose                 | Used By                   |
|----------|-----------|------------|-------------------------|---------------------------|
| TCP      | Inbound   | 2379-2380  | etcd server client API  | kube-apiserver, etcd      |

Use the following to access the SSH service on the control plane and etcd nodes:

| Protocol | Direction | Port Range | Purpose                 | Used By                   |
|----------|-----------|------------|-------------------------|---------------------------|
| TCP      | Inbound   | 22         | SSHD server             | SSH clients               |


## Kubernetes worker nodes
The following table represents the ports published by the Kubernetes project that must be accessible from worker nodes.


| Protocol | Direction | Port Range  | Purpose               | Used By                 |
|----------|-----------|-------------|-----------------------|-------------------------|
| TCP      | Inbound   | 10250       | Kubelet API           | Self, Control plane     |
| TCP      | Inbound   | 30000-32767 | [NodePort Services](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport)    | All                     |

The API server port that is sometimes switched to 443.
Alternatively, the default port is kept as is and API server is put behind a load balancer that listens on 443 and routes the requests to API server on the default port.
 
Use the following to access the SSH service on the worker nodes:

| Protocol | Direction | Port Range | Purpose                 | Used By                   |
|----------|-----------|------------|-------------------------|---------------------------|
| TCP      | Inbound   | 22         | SSHD server             | SSH clients               |

## Bare Metal provider

On the Admin machine for a Bare Metal provider, the following ports need to be accessible to all the nodes in the cluster, from the same level 2 network, for initially PXE booting:

| Protocol | Direction | Port Range | Purpose                 | Used By                       |
|----------|-----------|------------|-------------------------|-------------------------------|
| TCP      | Inbound   | 67         | boots DHCP              | All nodes, for network boot   |
| TCP      | Inbound   | 69         | boots TFTP              | All nodes, for network boot   |
| TCP      | Inbound   | 80         | boots HTTP              | All nodes, for network boot   |
| TCP      | Inbound   | 42113      | tink-server gRCP        | All nodes, talk to Tinkerbell |
| TCP      | Inbound   | 50061      | hegl HTTP               | All nodes, talk to Tinkerbell |

## VMware provider

The following table displays ports that need to be accessible from the VMware provider running EKS Anywhere:


| Protocol | Direction | Port Range  | Purpose                 | Used By                 |
|----------|-----------|-------------|-------------------------|-------------------------|
| TCP      | Inbound   | 443         | vCenter Server          | vCenter API endpoint    |
| TCP      | Inbound   | 6443        | Kubernetes API server   | Kubernetes API endpoint |
| TCP      | Inbound   | 2379        | Manager                 | Etcd API endpoint       |
| TCP      | Inbound   | 2380        | Manager                 | Etcd API endpoint       |

## Nutanix provider

The following table displays ports that need to be accessible from the Nutanix provider running EKS Anywhere:

| Protocol | Direction | Port Range  | Purpose                 | Used By                    |
|----------|-----------|-------------|-------------------------|----------------------------|
| TCP      | Inbound   | 9443        | Prism Central Server    | Prism Central API endpoint |
| TCP      | Inbound   | 6443        | Kubernetes API server   | Kubernetes API endpoint    |
| TCP      | Inbound   | 2379        | Manager                 | Etcd API endpoint          |
| TCP      | Inbound   | 2380        | Manager                 | Etcd API endpoint          |

## Control plane management tools

A variety of control plane management tools are available to use with EKS Anywhere.
One example is Jenkins.


| Protocol | Direction | Port Range  | Purpose                 | Used By                 |
|----------|-----------|-------------|-------------------------|-------------------------|
| TCP      | Inbound   | 8080        | Jenkins Server          | HTTP Jenkins endpoint   |
| TCP      | Inbound   | 8443        | Jenkins Server          | HTTPS Jenkins endpoint  |
