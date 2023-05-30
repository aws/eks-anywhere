---
title: "CIS Self-Assessment Guide"
linkTitle: "CIS Self-Assessment Guide"
weight: 60
aliases:
    /docs/reference/security/kube-bench/
description: >
  CIS Benchmark Self-Assessment Guide for EKS Anywhere clusters
---

The CIS Benchmark self-assessment guide serves to help EKS Anywhere users evaluate the level of security of the hardened cluster configuration against Kubernetes benchmark controls from the Center for Information Security (CIS). This guide will walk through the various controls and provide updated example commands to audit compliance in EKS Anywhere clusters.

You can verify the security posture of your EKS Anywhere cluster by using a tool called [`kube-bench`](https://github.com/aquasecurity/kube-bench). The ideal way to run the benchmark tests on your EKS Anywhere cluster is to apply the [Kube-bench Job YAMLs](https://github.com/aws/eks-anywhere/blob/main/test/kube-bench/jobs) to the cluster. This runs the `kube-bench` tests on a Pod on the cluster, and the logs of the Pod provide the test results.

Kube-bench currently does not support unstacked `etcd` topology (which is the default for EKS Anywhere), so the following checks are skipped in the default kube-bench Job YAML. If you created your EKS Anywhere cluster with stacked `etcd` configuration, you can apply the stacked `etcd` [Job YAML](https://github.com/aws/eks-anywhere/blob/main/test/kube-bench/jobs/controlplane/kube-bench-stacked-etcd.yaml) instead.

| Check number | Check description |
| :---: | :---: |
| 1.1.7 | Ensure that the etcd pod specification file permissions are set to 644 or more restrictive |
| 1.1.8 | Ensure that the etcd pod specification file ownership is set to root:root |
| 1.1.11 | Ensure that the etcd data directory permissions are set to 700 or more restrictive |
| 1.1.12 | Ensure that the etcd data directory ownership is set to etcd:etcd |

The following tests are also skipped, because they are not applicable or enforce settings that might make the cluster unstable.

| Check number | Check description | Reason for skipping |
| :---: | :---: | :---: |
| **Controlplane node configuration** |
| 1.2.6 | Ensure that the –kubelet-certificate-authority argument is set as appropriate | When generating serving certificates, functionality could break in conjunction with hostname overrides which are required for certain cloud providers |
| 1.2.16 | Ensure that the admission control plugin PodSecurityPolicy is set | Enabling Pod Security Policy can cause applications to unexpectedly fail |
| 1.2.32 | Ensure that the –encryption-provider-config argument is set as appropriate | Enabling encryption changes how data can be recovered as data is encrypted |
| 1.2.33 | Ensure that encryption providers are appropriately configured | Enabling encryption changes how data can be recovered as data is encrypted |
| **Worker node configuration** |
| 4.2.6 | Ensure that the –protect-kernel-defaults argument is set to true | System level configurations are required before provisioning the cluster in order for this argument to be set to true |
| 4.2.10 | Ensure that the –tls-cert-file and –tls-private-key-file arguments are set as appropriate | When generating serving certificates, functionality could break in conjunction with hostname overrides which are required for certain cloud providers |

{{< alert title="Note" color="primary" >}}
Running kube-bench on Bottlerocket controlplane nodes currently produces false negatives with respect to pod specification file (manifest) permissions, since the [default configuration](https://github.com/aquasecurity/kube-bench/blob/main/cfg/config.yaml) does not include the paths in which Bottlerocket places these manifests. This issue is being tracked [here](https://github.com/aquasecurity/kube-bench/issues/996).
{{< /alert >}}
