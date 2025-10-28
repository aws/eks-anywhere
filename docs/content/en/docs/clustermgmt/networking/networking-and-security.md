---
title: "Secure connectivity with CNI and Network Policy"
linkTitle: "Secure connectivity with CNI and Network Policy"
weight: 30
aliases:
    /docs/tasks/workload/networking-and-security/
description: >
  How to validate the setup of Cilium CNI and deploy network policies to secure workload connectivity.
---

{{% alert title="Announcements" color="warning" %}}
* EKS Anywhere release `v0.24.0` introduces the First-Party Supported Cilium as the default Cilium CNI in an EKS Anywhere cluster. The image for First-Party Supported Cilium is available in AWS public ECR gallery https://gallery.ecr.aws/eks/cilium/cilium. It is recommended to use this first-party supported Cilium which has been tested by AWS as the CNI in your cluster. 
{{% /alert %}}

<!-- overview -->

EKS Anywhere uses [Cilium](https://cilium.io) for pod networking and security.

<!-- body -->

Cilium is installed by default as a Kubernetes CNI plugin and so is already running in your EKS Anywhere cluster.

This section provides information about:

* Understanding Cilium components and requirements

* Validating your Cilium networking setup.

* Using Cilium to securing workload connectivity using Kubernetes Network Policy.


## Cilium Features
The following table lists Cilium features and notes which of those features are built into EKS Anywhere's default Cilium version , upstream Open Source, and Cilium Enterprise.

<details><summary>Expand to see Cilium Features</summary>

| Headline/Feature   | &nbsp;&nbsp;EKS Anywhere Default Cilium | &nbsp;&nbsp;Cilium OSS |  &nbsp;&nbsp;Isovalent Cilium Enterprise |
| ---------------    | ---------- | ---------- |  --------------------        |
| Networking Routing with tunneling mode |  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |  &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |  &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |
| Networking Routing with native routing mode |  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |  &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |  &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |
| Kubernetes Host Scope IPAM  | &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |  &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** | &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |
| Kubernetes Network Policy  | &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |  &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** | &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |
| Egress Masquerade  | &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |  &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** | &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |
| CNI Exclusive Configuration  | &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |  &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** | &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |
| Policy Enforcement Modes  | &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |  &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** | &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |
| Load-Balancing (L3/L4) | &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** | &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |  &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |
| Advanced Network Policy & Encryption (DNS, L7, TLS/SNI, ...) | &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;**&mdash;** | &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |  &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |
| Ingress, Gateway API, & Service Mesh | &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;**&mdash;** | &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |  &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |
| Multi-Cluster, Egress Gateway, BGP | &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;**&mdash;** | &nbsp;&nbsp;&nbsp;&nbsp;**&mdash;** | &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |
| Hubble Network Observability (Metrics, Logs, Prometheus, Grafana, OpenTelemetry) | &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;**&mdash;** | &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** | &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |
| SIEM Integration & Timescape Observability Storage | &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;**&mdash;** | &nbsp;&nbsp;&nbsp;&nbsp;**&mdash;** | &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |
| Tetragon Runtime Security | &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;**&mdash;** | &nbsp;&nbsp;&nbsp;&nbsp;**&mdash;** | &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |
| Enterprise-hardened Cilium Distribution, Training, 24x7 Enterprise Grade Support | &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;**&mdash;** | &nbsp;&nbsp;&nbsp;&nbsp;**&mdash;** | &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |

</details>

## Cilium Components

The primary Cilium Agent runs as a DaemonSet on each Kubernetes node.  Each cluster also includes a Cilium Operator Deployment to handle certain cluster-wide operations.  For EKS Anywhere, Cilium is configured to use the Kubernetes API server as the identity store, so no etcd cluster connectivity is required.

In a properly working environment, each Kubernetes node should have a Cilium Agent pod (`cilium-WXYZ`) in “Running” and ready (1/1) state.
By default there will be two
Cilium Operator pods (`cilium-operator-123456-WXYZ`) in “Running” and ready (1/1) state on different Kubernetes nodes for high-availability.

Run the following command to ensure all cilium related pods are in a healthy state.

```bash
kubectl get pods -n kube-system | grep cilium
```

Example output for this command in a 3 node environment is:

```
kube-system   cilium-fsjmd                                1/1     Running           0          4m
kube-system   cilium-nqpkv                                1/1     Running           0          4m
kube-system   cilium-operator-58ff67b8cd-jd7rf            1/1     Running           0          4m
kube-system   cilium-operator-58ff67b8cd-kn6ss            1/1     Running           0          4m
kube-system   cilium-zz4mt                                1/1     Running           0          4m
```

## Network Connectivity Requirements

To provide pod connectivity within an on-premises environment, the Cilium agent implements an overlay network using the GENEVE tunneling protocol.   As a result,
**UDP port 6081 connectivity MUST be allowed by any firewall running between Kubernetes nodes** running the Cilium agent.

Allowing ICMP Ping (type = 8, code = 0) as well as TCP port 4240 is also recommended in order for Cilium Agents to validate node-to-node connectivity as
part of internal status reporting.

## Validating Connectivity

Install the latest version of [Cilium CLI](https://docs.cilium.io/en/stable/gettingstarted/k8s-install-default/#install-the-cilium-cli).
The Cilium CLI has connectivity test functionality to validate proper installation and connectivity within a Kubernetes cluster.

By default, Cilium CLI will run tests in the `cilium-test-1` namespace which can be changed by using `--test-namespace` flag.   For example:

```bash
cilium connectivity test
```

Successful test output will show all tests in a "successful" (some tests might be in "skipped") state.   For example:

```
✅ [cilium-test-1] All 12 tests (139 actions) successful, 72 tests skipped, 0 scenarios skipped.
```

Afterward, simply delete the namespace to clean-up the connectivity test:

```bash
kubectl delete ns cilium-test-1
```

## Kubernetes Network Policy

By default, all Kubernetes workloads within a cluster can talk to any other workloads in the cluster, as well as any workloads outside the cluster.  To enable a stronger security posture, Cilium implements the Kubernetes Network Policy specification to provide identity-aware firewalling / segmentation of Kubernetes workloads.

Network policies are defined as Kubernetes YAML specifications that are applied to a particular namespaces to describe that connections should be allowed to or from a given set of pods.  These network policies are “identity-aware” in that they describe workloads within the cluster using Kubernetes metadata like namespace and labels, rather than by IP Address.

Basic network policies are validated as part of the above Cilium connectivity check test.

For next steps on leveraging Network Policy, we encourage you to explore:

* A hands-on [Network Policy Intro Tutorial](https://github.com/networkpolicy/tutorial) .

* The visual [Network Policy Editor](https://editor.cilium.io) .

* The #networkpolicy channel on [Cilium Slack](https://cilium.io/slack) .

* Other resources on [networkpolicy.io](https://networkpolicy.io) .


## Additional Cilium Features

Some advanced features of Cilium are not enabled as part of EKS Anywhere, including:

* [Hubble observability](https://docs.isovalent.com/operations-guide/features/network-visibility/index.html)
* [DNS-aware and HTTP-Aware Network Policy](https://docs.isovalent.com/quick-start/policy_lifecycle.html)
* [Multi-cluster Routing](https://docs.isovalent.com/operations-guide/features/cluster-mesh/index.html)
* [Transparent Encryption](https://docs.cilium.io/en/v1.13/security/network/encryption/)
* [Advanced Load-balancing](https://docs.isovalent.com/operations-guide/features/cilium-standalone-gateway.html)

Please contact the EKS Anywhere team if you are interested in leveraging these advanced features along with EKS Anywhere.
