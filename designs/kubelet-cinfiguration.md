# Kubelet Configuration

## Introduction

### Problem

Currently, EKS-A doesn’t provide users the ability to customize any Kubelet configuration parameters besides a small subset of configs like node labels, custom DNS resolve conf path, etc. If users want to customize anything beyond these parameters, they have manually go in the node and configure them. If the nodes get rolled out, any custom configs they did on the cluster gets wiped and they have to re-configure it on the new nodes all over again.

### Goals and objectives

As an EKS-A user, I would like to:

* configure kubelet parameters through EKS-A cluster spec.
* modify the kubelet configuration as needed by updating the EKS-A cluster spec.

### Statement of Scope

#### In Scope

* Allowing users to configure kubelet configuration parameters through EKS-A.
* Using the kubelet config file in favor of kubelet flags to set the parameters, since the flags are deprecated now.
* Exposing all config parameters that kubelet supports.
* Support for all kubelet configs that bottlerocket supports.

#### Out of Scope

* Ensuring that passing certain parameters won’t cause cluster to fail.
    * Kubelet configs should be used by advanced users who know what they are doing. Some configs can cause the clusters to potentially be non-operable. These are hard to validate so EKS-A will only validate that the config is of proper format.

## Overview of solution

### Design Details

Since we want to expose all the kubelet configuration parameters, the simplest solution is to just make the kubelet config a string type and let users pass it in as a multiline string in the cluster spec. EKS-A can do some basic validations like yaml parsing and type validation and then simply propagate this file down to CAPI. However, this won’t surface any errors for the users regarding type validations. We have another approach that we can use which is pulling the upstream object for Kubelet Configuration. This let's us surface type errors as well as errors related to invalid keys for the Kubelet Configuration.

#### Proposed and Finalized Solution 

Pass yaml as a config object - upstream kubernetes object.
We will be using the most recent version's upstream Kubelet Configuration object in our EKSA Spec. This is the most recent kubernetes version EKSA supports.

```
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: eksa-cluster
spec:
  controlPlaneConfiguration:
    kubeletConfiguration:
      apiVersion: kubelet.config.k8s.io/v1beta1
      kind: KubeletConfiguration
      address: "192.168.0.8"
      port: 20250
      serializeImagePulls: false
      evictionHard:
          memory.available:  "100Mi"
          nodefs.available:  "10%"
          nodefs.inodesFree: "5%"
          imagefs.available: "15%"
```
Pros:

* Easier for the user to do YAML parsing while giving input - type safety at compile time
* Can handle complex and larger yamls
* Surface type errors for users

Cons

* EKSA will need to create code for managing the object

On CAPI/kubeadm side, we can leverage the kubeadm configuration patches to apply KubeletConfiguration  that users pass in like this :

      kubeadmConfigSpec:
        files:
        - path: /etc/kubernetes/patches/kubeletconfiguration.yaml
          owner: "root:root"
          permissions: "0644"
          content: |
            apiVersion: kubelet.config.k8s.io/v1beta1
            kind: KubeletConfiguration
            address: "192.168.0.8"
            port: 20250
            serializeImagePulls: false
            evictionHard:
                memory.available:  "100Mi"
                nodefs.available:  "10%"
                nodefs.inodesFree: "5%"
                imagefs.available: "15%"
        initConfiguration:
          patches:
            directory: /etc/kubernetes/patches
        joinConfiguration:
          patches:
            directory: /etc/kubernetes/patches

Kubeadm config patches takes the default kubelet config that kubeadm would generate and overrides it with whatever fields are set in the patch files. This allows us to let users pass any subset of kubelet parameters they want to configure while also keeping the kubeadm defaults for other fields. This approach is explained in more detail in the CAPI doc here https://cluster-api.sigs.k8s.io/tasks/bootstrap/kubeadm-bootstrap/kubelet-config#use-kubeadms-kubeletconfiguration-patch-target


#### Alternate Solution 

Pass as string in the cluster spec

```
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: eksa-cluster
spec:
  controlPlaneConfiguration:
    kubeletConfiguration: |
      apiVersion: kubelet.config.k8s.io/v1beta1
      kind: KubeletConfiguration
      address: "192.168.0.8"
      port: 20250
      serializeImagePulls: false
      evictionHard:
          memory.available:  "100Mi"
          nodefs.available:  "10%"
          nodefs.inodesFree: "5%"
          imagefs.available: "15%"
```
Pros

* Simplicity of the input
* Can be provided to the CAPI template as the direct string as that is the type of input CAPI would expect (no type conversions)

Cons

* EKSA will need additional yaml parsing validations to verify that its a valid yaml
* Difficult to maintain for larger strings


### Handling Kubelet params that EKS-A currently configures

EKS-A spec currently already sets a few configs that are passed to kubelet through kubelet extra args. These include user configurable fields like node-labels and resolv-conf path, and also some EKS-A default ones like tls-cipher-suites, cloud-provider, read-only-port,  anonymous-auth and eviction-hard (some are specific to certain providers).
There’s two questions that need we need to answer here:

=> What happens when a user tries to configure a kubelet parameter directly and also through another cluster spec field?

These are the current kubelet arguments we expose via our EKSA cluster spec today. 
```
   Resolv Conf (DNS) - clusterNetwork.dns.resolvConf
   Node labels (CP and MD) - controlPlaneConfiguration.labels; workerNodeGroupConfigurations.labels
   ```
#### PROPOSED SOLUTION 
The proposed approach is to use the explicit EKS-A field to override the fields in KubeletConfiguration. So for example, if a user tries to configure node labels through `KubeletConfiguration.nodeLabels` and also through `Cluster.spec.controlPlaneConfiguration.labels`, the latter would take precedence. We can also throw an error in EKS-A if both are set. In terms of implementation, this would mean that we would still use the Kubelet flags to set the values of these fields if specified. The pro in using this approach is also that we would prevent node rollouts.

#### ALTERNATE SOLUTION 
 We deprecate the fields in favor of having the Kubelet configuration be the source of truth. So for example, if a user tries to configure node labels through `KubeletConfiguration.nodeLabels` and also through `Cluster.spec.controlPlaneConfiguration.labels`, the former would take precedence. We can throw an error and also suggest the user to use the KubeletConfiguration to set the labels. The cons for this would be unexpected node rollouts even if the same value does not change for the node. This would especially cause problems for Tinkerbell since customers might not have extra nodes around when these rollouts happen. To manage these rollouts on the CAPI side, we would need to do some sort of compare for these configs and check that there isn’t really any difference in these specific values we are setting.

=> What happens when a user tries to configure a kubelet parameter that EKS-A sets as well?

We have similar options for this case. 

#### PROPOSED SOLUTION 
 EKSA sets a few field defaults for Kubelet args. If users try to set any of those fields, the proposed approach here is to block any field that EKS-A configures. We would throw an error if EKSA default fields are set in the Kubelet Configuration. In this case, the KubeletConfiguration on the EKS-A spec won’t be the source of truth anymore if it’ll always get overridden by EKS-A. This will also ensure users don’t create a cluster with certain kubelet configurations without realizing that some of those fields could be altered by EKS-A.
Fields EKSA defaults today that are provider specific:
- cloud-provider (vsp, nutanix)
- read-only-port   (all but docker)
- anonymous-auth    (all but docker)
- provider-id (tinkerbell, cs)
- eviction-hard (nutanix, docker)
- tls-cipher-suites
 
We will take a look at these fields on a case-by-case basis and check if we can let users configure these settings.

#### ALTERNATE SOLUTION 
Allow the field in KubeletConfiguration but always override it with the EKS-A specific values.

### BOTTLEROCKET SUPPORT

Bottlerocket doesn’t support providing a custom kubelet configuration file. Instead, you have to set the relevant kubelet fields using the bottlerocket API, which itself supports only a subset of kubelet parameters. 

#### PROPOSED SOLUTION 
To support this feature for bottlerocket, we’ll have to maintain a mapping of kubelet configs to bottlerocket settings on CAPI side to translate the kubelet configs to the corresponding bottlerocket settings.
For example, the following KubeletConfiguration:
```
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
maxPods: 100
allowedUnsafeSysctls:
- "net.core.somaxconn"
- "net.ipv4.ip_local_port_range"
```
will have to be converted to the following bottlerocket settings:

```
[settings.kubernetes]
max-pods = 100
allowed-unsafe-sysctls = ["net.core.somaxconn", "net.ipv4.ip_local_port_range"]
```
On EKS-A side, we can add validations to only allow users to set kubelet fields that bottlerocket supports and error out if an unsupported field is set.

Bottlerocket only supports configuration of 32 fields for Kubelet as opposed to a total of 113 fields that other OSes can configure. We would need to map all those like the way we map the 3 existing settings we expose in the OS Host configuration section of machine configs.
Ref https://anywhere.eks.amazonaws.com/docs/getting-started/optional/hostosconfig/#bottlerocketconfiguration

We finalized on providing support for Bottlerocket and deprecating the settings from the OS Host configuration spec in the Machine Configs in the favor of using the new Kubelet Configuration struct. With this we will introduce the new bottlerocket settings on the CAPI side and maintain those with every new version we support. Additionally we will throw errors if user tries to set the settings that aren’t supported by the most recent config struct upstream.


#### ALTERNATE SOLUTION 
The other option would be to not support KubeletConfiguration for bottlerocket directly but instead consolidate all the other bottlerocket kubelet settings in the bottlerocket configuration section of EKS-A spec. This field exposes 3 bottlerocket kubelet settings (max-pods, allowed-unsafe-sysctls and cluster-dns-ips) already so it would make sense to expose all the other kubelet configs here as well. Bottlerocket doesn’t support providing a custom kubelet configuration file. Instead, you have to set the relevant kubelet fields using the bottlerocket API, which itself supports only a subset of kubelet parameters. 

In regards to the EKSA cluster spec, the existing settings are exposed in the provider machine configs. Exposing the rest of the supported fields in the same host configuration would make it easier to also validate the OS specific kubelet settings.
However, this would mean maintaining two separate configurations for Kubelet Configuration based on the Operating systems used.