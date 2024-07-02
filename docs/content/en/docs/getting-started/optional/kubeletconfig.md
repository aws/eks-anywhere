---
title: "KubeletConfiguration"
linkTitle: "KubeletConfiguration"
weight: 40
aliases:
    /docs/reference/clusterspec/optional/kubeletconfig/
description: >
  EKS Anywhere cluster yaml specification for Kubelet Configuration
---

## Kubelet Configuration Support

#### Provider support details
|                     | vSphere | Bare Metal | Nutanix | CloudStack | Snow |
|:-------------------:|:-------:|:----------:|:-------:|:----------:|:----:|
|    Ubuntu 20.04     |    ✔    |     ✔      |    ✔    |     —      |  —   |
|    Ubuntu 22.04     |    ✔    |     ✔      |    ✔    |     —      |  —   |
| Bottlerocket        |    ✔    |     —      |    —    |     —      |  —   |
|      RHEL 8.x       |    ✔    |     ✔      |    ✔    |     ✔      |  —   |
|      RHEL 9.x       |    —    |     —      |    ✔    |     ✔      |  —   |


You can configure EKS Anywhere to specify Kubelet settings and configure those for control plane and/or worker nodes starting from `v0.20.0`. This can be done using `kubeletConfiguration`.
The following cluster spec shows an example of how to configure `kubeletConfiguration`:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
   name: my-cluster-name
spec:
   ...
   ...
 controlPlaneConfiguration:         # Kubelet configuration for Control plane nodes
    kubeletConfiguration:
      kind: KubeletConfiguration
      maxPods: 80
   ...
  workerNodeGroupConfigurations:    # Kubelet configuration for Worker nodes
  - count: 1
    kubeletConfiguration:
      kind: KubeletConfiguration
      maxPods: 85
   ...
```

`kubeletConfiguration` should contain the configuration to be used by the kubelet while creating or updating a node. It must contain the `kind` key with the value `KubeletConfiguration` for EKS Anywhere to process the settings. This configuration must only be used with valid settings as it may cause unexpected behavior from the Kubelet if misconfigured. EKS Anywhere performs a limited set of data type validations for the Kubelet Configuration, however it is ultimately the user's responsibility to make sure that valid configuration is set for Kubelet Configuration.

More details on the Kubelet Configuration object and its supported fields can be found [here](https://kubernetes.io/docs/reference/config-api/kubelet-config.v1beta1/). EKS Anywhere only supports the latest Kubernetes version's `KubeletConfiguration`.

## Bottlerocket Support

The only provider that supports `kubeletConfiguration` with Bottlerocket is vSphere. The list of settings that can be configured for Bottlerocket can be found [here](https://bottlerocket.dev/en/os/1.19.x/api/settings/kubernetes/#alphaorder). This page also describes other various settings like Kubelet Options. The settings supported by Bottlerocket will have information specific to the `Kubelet Configuration` keyword in there. Refer to the documentation to learn about the supported fields as well as their data types as they may vary from the upstream object's data types.

Note that this is the preferred and supported way to specify any Kubelet settings from the release `v0.20.0` onwards. Previously the [`hostOSConfiguration.bottlerocketConfiguration.kubernetes`](https://anywhere.eks.amazonaws.com/docs/getting-started/optional/hostosconfig/#kubernetes) field was used to specify Bottlerocket Kubernetes settings. That has been deprecated from `v0.20.0`

Here's a list of supported fields by Bottlerocket for Kubelet Configuration -
<details>

- 	`allowedUnsafeSysctls`
-	`clusterDNSIPs`
-	`clusterDomain`
-	`containerLogMaxFiles`
-	`containerLogMaxSize`
-	`cpuCFSQuota`
-	`cpuManagerPolicy`
-	`cpuManagerPolicyOptions`
-	`cpuManagerReconcilePeriod`
-	`eventBurst`
-	`eventRecordQPS`
-	`evictionHard`
-	`evictionMaxPodGracePeriod`
-	`evictionSoft`
-	`evictionSoftGracePeriod`
-	`imageGCHighThresholdPercent`
-	`imageGCLowThresholdPercent`
-	`kubeAPIBurst`
-	`kubeAPIQPS`
-	`kubeReserved`
-	`maxPods`
-	`memoryManagerPolicy`
-	`podPidsLimit`
-	`providerID`
-	`registryBurst`
-	`registryPullQPS`
-	`shutdownGracePeriod`
-	`shutdownGracePeriodCriticalPods`
-	`systemReserved` 
-	`topologyManagerPolicy`
-   `topologyManagerScope`

</details>

## Special fields

### Duplicate fields

The `clusterNetwork.dns.resolvConf` is the file path to a file containing a custom DNS resolver configuration. This can now be provided in the Kubelet Configuration using the `resolvConf` field. Note that if both these fields are set, the Kubelet Configuration's field will take precendence and override the value from the `clusterNetwork.dns.resolvConf`.

### Blocked fields

Fields like `providerID` or `cloudProvider` are set by EKS Anywhere and can't be set by users. This is to maintain seamless support for all providers.

## Node Rollouts

Adding, updating, or deleting the Kubelet Configuration will cause node rollouts to the respective nodes that the configuration affects. This is especially important to consider in providers like Baremetal since the node rollouts that are caused by the Kubelet config changes could require extra hardware provisioned depending on your rollout strategy.