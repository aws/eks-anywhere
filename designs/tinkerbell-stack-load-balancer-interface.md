# Tinkerbell stack load balancer interface customization

## Problem Statement:

Customer wants to specify the Tinkerbell stack load-balancer interface in order to override the default interface used by the current load-balancer(kube-vip) daemonset. This can be done via specifying the [vip_interface](https://github.com/kube-vip/kube-vip/blob/04ce471366c21d4586fb2d683cd166f0dc4e18ce/pkg/kubevip/config_envvar.go#L34) env variable in the kube-vip daemonset after the cluster is created. But the main issue with that is this change will not persist whenever the management cluster is upgraded. In order to solve this problem, we would allow users to specify the interface through the cluster spec and configure it in our kube-vip daemonset. This doc proposes a solution for where the interface can be specified in the cluster spec and discusses various trade-offs with alternate options.

## Proposed Solution:

Specify the interface in the TinkerbellDatacenterConfig object spec at the root level:

**API Schema:**

```
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: mgmt-cluster
spec:
  ...
  kubernetesVersion: 1.30
  ...

---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: TinkerbellDatacenterConfig
metadata:
  name: sjparekh-mgmt
spec:
  ...
  tinkerbellIP: "x.x.x.x"
  osImageURL: "https://s3-bucket-url/ubuntu.gz"
  loadBalancerInterface: "eth0"
  skipLoadBalancerDeployment: "false"
  ...
```

**Tradeoffs:**

This approach allows specifying the interface to a tinkerbell-specific custom resource which is where we already have some other tinkerbell config as well so it seems like a more appropriate place to have it but the drawback is that we are adding a field to the tinkerbell datacenter config custom resource when the field itself is not directly related to the datacenter. 

Another drawback is that adding more fields in the future will require changing the API again but at the same time, it allows us to fail the validations quickly and more easily if the field is misconfigured.

**Implementation Details:**

The load balancer interface specified in the cluster spec will be passed to the tinkerbell stack helm chart through the [createValuesOverride](https://github.com/aws/eks-anywhere/blob/e24df70ec55e1be403e19685aded8850d3c45dad/pkg/providers/tinkerbell/stack/stack.go#L511) method of the [Installer](https://github.com/aws/eks-anywhere/blob/e24df70ec55e1be403e19685aded8850d3c45dad/pkg/providers/tinkerbell/stack/stack.go#L78C6-L78C15) struct when installing the stack during cluster create/upgrade operations.

The upstream helm chart template already handles setting the [vip_interface](https://github.com/tinkerbell/charts/blob/95df5bc5f89c76dd0f6cc2955bb590f023d94f28/tinkerbell/stack/templates/kubevip.yaml#L34C9-L37C19) env variable in the kube-vip daemonset with the interface value from the [values.yaml](https://github.com/tinkerbell/charts/blob/95df5bc5f89c76dd0f6cc2955bb590f023d94f28/tinkerbell/stack/values.yaml#L38) file.

**Testing:**

* E2E tests would be added to verify that the load balancer is indeed deployed with the expected interface
* Unit tests would be added for any functional changes implemented


**Documentation:**

* We would have to add tinkerbellStackLoadBalancerInterface as an optional configuration for tinkerbell datacenter config [fields](https://anywhere.eks.amazonaws.com/docs/getting-started/baremetal/bare-spec/#tinkerbelldatacenterconfig-fields) in the EKS Anywhere docs
* We also need to document that in the case of single-node clusters, same interface will be used for load-balancing both the tinkerbell stack as well as control plane components. If a user wants separate interface for them in a single-node cluster, they would have to skip deploying kube-vip and instead deploy their own load balancers with one for cp components and one for tinkerbell stack configured separately with their custom interfaces.
* Specify that new nodes will not be rolled out when the cluster is created/upgrade with the custom interface

## Alternate Solutions Considered:

Specify the interface by exposing a tinkerbell config map at the root level of the cluster spec

**API Schema:** 

```
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: mgmt-cluster
spec:
  ...
  kubernetesVersion: 1.30
  tinkerbellConfig:
    loadBalancerInterface: "eth0"
  ...
```

**Tradeoffs:**

This approach will allow us to add more tinkerbell configuration fields in the future without having to change the API but the drawback is that we are adding a new provider-specific configuration to the root level of cluster spec which does not seem appropriate and we don’t have it for any other providers either.
