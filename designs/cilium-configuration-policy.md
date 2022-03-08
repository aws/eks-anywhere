# Configure Cilium CNI

## Introduction

**Problem:** The Cilium CNI plugin (built-in default for EKS Anywhere) supports defining and enforcement of network policies. Network policies determine how pods within a cluster are able to communicate with one another. Cilium accepts policy enforcement modes from the users to determine the allowed traffic between endpoints. Currently there is no way to configure Cilium launched by EKS Anywhere to use any mode apart from the default mode, which allows all traffic.

### Goals and Objectives

As an EKS Anywhere user:

* I want to have the ability to configure Cilium's [supported policy enforcement modes](https://docs.cilium.io/en/v1.9/policy/intro/).
* I want that, depending on the policy enforcement mode I select, EKS Anywhere should configure the network policies required for its core components to ensure that the cluster gets provisioned successfully.

### Non-Goals/Future Work

* Accepting network policies for user workloads through EKS Anywhere spec.


## Overview of Solution

With this feature, a user can launch Cilium with one of its supported policy enforcement modes: default, always, never. The default mode will be Cilium's default, which is also named `default`. The user can specify the mode through the EKS Anywhere cluster spec, and EKS Anywhere will use that while launching the Cilium Helm chart in the workload cluster.

### Implementation details

#### API Changes

A new type will be added to accept the cni plugin name and additional options from the user:

```go
type CNIConfig struct {
  Cilium *CiliumConfig `json:"cilium,omitempty"`
}

type CiliumConfig struct {
  PolicyEnforcementMode string `json:"policyEnforcementMode,omitempty"`
}
```

The user can then configure cilium, starting with the policy enforcement mode as follows:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: eks-anywhere
spec:
  clusterNetwork:
    cniConfig: 
      cilium: 
        policyEnforcementMode: "always"
```

We can further expand the new CNIConfig type to support more CNI plugins. We can also expand the CiliumConfig type to accept more values that the Cilium Helm chart accepts.

##### Unsupported CNI plugins

Currently we consider Kindnetd and Cilium Enterprise to be valid CNI plugins for EKS Anywhere, however we don't support them for the workload cluster as of now. So the CNIConfig type won't accept Kindnetd or Cilium Enterprise plugins to begin with, we can add configuration options for plugins as we start supporting them.

##### Validation

As we support more plugins, we will add validations to check that the user has specified only one of the CNIConfigs. These validations will be performed first during the create preflight validations and also as a part of the cluster validating webhook.


#### Changes required for the "always" policy enforcement mode

Cilium's `always` policy enforcement mode denies all traffic to and from all pods unless rules/network policies supporting the communication are specified. This means that the EKS Anywhere components, which includes the core Kubernetes deployments, EKS Anywhere deployments and Cluster API deployments will not be able to run without specifying the required network policies. So to allow traffic for these deployments, EKS Anywhere will configure [Kubernetes NetworkPolicy](https://kubernetes.io/docs/concepts/services-networking/network-policies/) resource that selects these deployments and allows traffic to and from them.  

Kubernetes NetworkPolicy is a namespaced resource that allows users to configure ingress and egress rules for pods specified using podSelectors. If a NetworkPolicy's pod selector field is left empty, then its rules apply to all pods within the NetworkPolicy's namespace. 
When the `always` policy enforcement mode is selected, EKS Anywhere will create NetworkPolicy resources in all core EKS Anywhere namespaces, allowing all ingress and egress traffic to all pods within those namespaces. This will ensure the cluster provisioning and management will go on uninterrupted.  
This is a sample NetworkPolicy to allow all ingress and egress traffic for workloads in a namespace called `test`:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-all-ingress-egress
  namespace: test
spec:
  podSelector: {}
  ingress:
  - {}
  egress:
  - {}
  policyTypes:
  - Ingress
  - Egress
```

The namespaces within which we need to create such policies will depend on:
* Whether the cluster is a self-managed/management cluster or a workload cluster managed by another cluster.
* The infrastructure provider being used.


For self-managed/management cluster, EKS Anywhere will create NetworkPolicy resources in the following namespaces allowing all ingress/egress traffic by default:
- kube-system
- eksa-system
- All core Cluster API namespaces:
  + capi-system
  + capi-kubeadm-bootstrap-system
  + capi-kubeadm-control-plane-system
  + etcdadm-bootstrap-provider-system
  + etcdadm-controller-system
  + cert-manager
- Infrastruture provider's namespace (for instance, capd-system OR capv-system)

For a workload cluster managed by another cluster, EKS Anywhere will create NetworkPolicy resource only in the following namespace by default:
- kube-system

For the workload clusters, we will limit the ingress/egress of pods in the kube-system namespace
to other pods only in the kube-system namespace by using the following NetworkPolicy:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-all-ingress-egress
  namespace: test
spec:
  podSelector: {}
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: kube-system
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: kube-system
  policyTypes:
  - Ingress
  - Egress
```

#### Changes in the output of generated cluster config

The generate clusterconfig command will continue to create a cluster config with cilium as the selected CNI. But instead of Cilium being specified as the name of the CNI plugin, it will be specified with default the (none) configuration options.
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: eks-anywhere
spec:
  clusterNetwork:
    cniConfig: 
      cilium: {}
```
