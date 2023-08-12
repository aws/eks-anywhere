---
title: "3. Cluster Networking"
linkTitle: "3. Cluster Networking"
weight: 16
aliases:
    /docs/reference/ports/
description: >
  EKS Anywhere cluster networking
---

#### Cluster Networking
EKS Anywhere clusters use the `clusterNetwork` field in the cluster spec to allocate pod and service IPs. Once the cluster is created, the `pods.cidrBlocks`, `services.cidrBlocks` and `nodes.cidrMaskSize` fields are immutable. As a result, extra care should be taken to ensure that there are sufficient IPs and IP blocks available when provisioning large clusters.
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: my-cluster-name
spec:
  clusterNetwork:
    pods:
      cidrBlocks:
      - 192.168.0.0/16
    services:
      cidrBlocks:
      - 10.96.0.0/12
```

The cluster `pods.cidrBlocks` is subdivided between nodes with a default block of size `/24` per node, which can also be [configured via]({{< relref "../optional/cni/#node-ips-configuration-option" >}}) the  `nodes.cidrMaskSize` field. This node CIDR block is then used to assign pod IPs on the node.

{{% alert title="Warning" color="warning" %}}
The maximum number of nodes will be limited to the number of subnets of size `/24` (or `nodes.cidrMaskSize` if configured) that can fit in the cluster `pods.cidrBlocks`.

The maximum number of pods per node is also limited by the size of the node CIDR block. For example with the default `/24` node CIDR mask size, there are a maximum of 256 IPs available for pods. Kubernetes recommends [no more than 110 pods per node.](https://kubernetes.io/docs/setup/best-practices/cluster-large/)
{{% /alert %}}

#### Ports and Protocols

EKS Anywhere requires that various ports on control plane and worker nodes be open.
Some Kubernetes-specific ports need open access only from other Kubernetes nodes, while others are exposed externally.
Beyond Kubernetes ports, someone managing an EKS Anywhere cluster must also have external access to ports on the underlying EKS Anywhere provider (such as VMware) and to external tooling (such as Jenkins).

If you are responsible for network firewall rules between nodes on your EKS Anywhere clusters, the following tables describe both Kubernetes and EKS Anywhere-specific ports you should be aware of.

##### Kubernetes control plane
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


##### Kubernetes worker nodes
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

##### Bare Metal provider

On the Admin machine for a Bare Metal provider, the following ports need to be accessible to all the nodes in the cluster, from the same level 2 network, for initially network booting:

| Protocol | Direction | Port Range | Purpose                 | Used By                       |
|----------|-----------|------------|-------------------------|------------------------------ |
| UDP      | Inbound   | 67         | Boots DHCP              | All nodes, for network boot   |
| UDP      | Inbound   | 69         | Boots TFTP              | All nodes, for network boot   |
| UDP      | Inbound   | 514        | Boots Syslog            | All nodes, for provisioning logs |
| TCP      | Inbound   | 80         | Boots HTTP              | All nodes, for network boot   |
| TCP      | Inbound   | 42113      | Tink-server gRPC        | All nodes, talk to Tinkerbell |
| TCP      | Inbound   | 50061      | Hegel HTTP              | All nodes, talk to Tinkerbell |
| TCP      | Outbound  | 623        | Rufio IPMI              | All nodes, out-of-band power and next boot ([optional]({{< relref "../baremetal/bare-prereq/#network-requirements" >}})) |
| TCP      | Outbound  | 80,443     | Rufio Redfish           | All nodes, out-of-band power and next boot ([optional]({{< relref "../baremetal/bare-prereq/#network-requirements" >}})) |

##### VMware provider

The following table displays ports that need to be accessible from the VMware provider running EKS Anywhere:


| Protocol | Direction | Port Range  | Purpose                 | Used By                 |
|----------|-----------|-------------|-------------------------|-------------------------|
| TCP      | Inbound   | 443         | vCenter Server          | vCenter API endpoint    |
| TCP      | Inbound   | 6443        | Kubernetes API server   | Kubernetes API endpoint |
| TCP      | Inbound   | 2379        | Manager                 | Etcd API endpoint       |
| TCP      | Inbound   | 2380        | Manager                 | Etcd API endpoint       |

##### Nutanix provider

The following table displays ports that need to be accessible from the Nutanix provider running EKS Anywhere:

| Protocol | Direction | Port Range  | Purpose                 | Used By                    |
|----------|-----------|-------------|-------------------------|----------------------------|
| TCP      | Inbound   | 9440        | Prism Central Server    | Prism Central API endpoint |
| TCP      | Inbound   | 6443        | Kubernetes API server   | Kubernetes API endpoint    |
| TCP      | Inbound   | 2379        | Manager                 | Etcd API endpoint          |
| TCP      | Inbound   | 2380        | Manager                 | Etcd API endpoint          |

##### Snow provider

In addition to the [Ports Required to Use AWS Services on an AWS Snowball Edge Device](https://docs.aws.amazon.com/snowball/latest/developer-guide/port-requirements.html), the following table displays ports that need to be accessible from the Snow provider running EKS Anywhere:

| Protocol | Direction | Port Range  | Purpose                 | Used By                           |
|----------|-----------|-------------|-------------------------|-----------------------------------|
| TCP      | Inbound   | 9092        | Device Controller       | EKS Anywhere and CAPAS controller |
| TCP      | Inbound   | 8242        | EC2 HTTPS endpoint      | EKS Anywhere and CAPAS controller |
| TCP      | Inbound   | 6443        | Kubernetes API server   | Kubernetes API endpoint           |
| TCP      | Inbound   | 2379        | Manager                 | Etcd API endpoint                 |
| TCP      | Inbound   | 2380        | Manager                 | Etcd API endpoint                 |

##### Control plane management tools

A variety of control plane management tools are available to use with EKS Anywhere.
One example is Jenkins.


| Protocol | Direction | Port Range  | Purpose                 | Used By                 |
|----------|-----------|-------------|-------------------------|-------------------------|
| TCP      | Inbound   | 8080        | Jenkins Server          | HTTP Jenkins endpoint   |
| TCP      | Inbound   | 8443        | Jenkins Server          | HTTPS Jenkins endpoint  |
