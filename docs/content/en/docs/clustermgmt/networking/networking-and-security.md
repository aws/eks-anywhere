---
title: "Secure connectivity with CNI and Network Policy"
linkTitle: "Secure connectivity with CNI and Network Policy"
weight: 30
aliases:
    /docs/tasks/workload/networking-and-security/
description: >
  How to validate the setup of Cilium CNI and deploy network policies to secure workload connectivity.
---

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
| ---------------    | ----------       | ---------- |  --------------------        |
| Networking Routing (CNI) |  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |  &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |  &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |
| Identity-Based Network Policy (Labels, CIDR)  | &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |  &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** | &nbsp;&nbsp;&nbsp;&nbsp;**&#10004;** |
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

Cilium includes a connectivity check YAML that can be deployed into a test namespace in order to validate proper installation and connectivity within a Kubernetes cluster.   If the connectivity check passes, all pods created by the YAML manifest will reach “Running” and ready (1/1) state.    We recommend running this test only once you have multiple worker nodes in your environment to ensure you are validating cross-node connectivity.

It is important that this test is run in a dedicated namespace, with no existing network policy.   For example:

```bash
kubectl create ns cilium-test
```

```bash
kubectl apply -n cilium-test -f https://docs.isovalent.com/v1.10/public/connectivity-check-eksa.yaml
```

Once all pods have started, simply checking the status of pods in this namespace will indicate whether the tests have passed:

```bash
kubectl get pods -n cilium-test
```

Successful test output will show all pods in a "Running" and ready (1/1) state:

```
NAME                                                     READY   STATUS    RESTARTS   AGE
echo-a-d576c5f8b-zlfsk                                   1/1     Running   0          59s
echo-b-787dc99778-sxlcc                                  1/1     Running   0          59s
echo-b-host-675cd8cfff-qvvv8                             1/1     Running   0          59s
host-to-b-multi-node-clusterip-6fd884bcf7-pvj5d          1/1     Running   0          58s
host-to-b-multi-node-headless-79f7df47b9-8mzbp           1/1     Running   0          58s
pod-to-a-57695cc7ff-6tqpv                                1/1     Running   0          59s
pod-to-a-allowed-cnp-7b6d5ff99f-4rhrs                    1/1     Running   0          59s
pod-to-a-denied-cnp-6887b57579-zbs2t                     1/1     Running   0          59s
pod-to-b-intra-node-hostport-7d656d7bb9-6zjrl            1/1     Running   0          57s
pod-to-b-intra-node-nodeport-569d7c647-76gn5             1/1     Running   0          58s
pod-to-b-multi-node-clusterip-fdf45bbbc-8l4zz            1/1     Running   0          59s
pod-to-b-multi-node-headless-64b6cbdd49-9hcqg            1/1     Running   0          59s
pod-to-b-multi-node-hostport-57fc8854f5-9d8m8            1/1     Running   0          58s
pod-to-b-multi-node-nodeport-54446bdbb9-5xhfd            1/1     Running   0          58s
pod-to-external-1111-56548587dc-rmj9f                    1/1     Running   0          59s
pod-to-external-fqdn-allow-google-cnp-5ff4986c89-z4h9j   1/1     Running   0          59s
```

Afterward, simply delete the namespace to clean-up the connectivity test:

```bash
kubectl delete ns cilium-test
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
