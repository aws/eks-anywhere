# Support advanced Virtual Machine Configuration (VMX) settings on vSphere node VMs 

## Introduction

**Problem:** 

vSphere allows users to configure [advanced VMX settings](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.vcenter.configuration.doc/GUID-62184858-32E6-4DAC-A700-2C80667C3558.html) for individual VMs. However, EKS Anywhere on vSphere currently doesn't provide a way for users to modify these settings.

One of the use-cases for exposing this is to allow users to [disable swap space](https://github.com/aws/eks-anywhere/issues/4237) on VMs by adding this VMX config `sched.swap.vmxSwapEnabled=FALSE`.

### Goals and Objectives

As an EKS Anywhere user, I want to:

* configure VMX settings through the EKS Anywhere clusterconfig.

### Statement of Scope

#### In Scope

* Allow users to enable/disable VMX settings on their cluster VMs through EKS Anywhere.

#### Future Scope

* Iteratively add more VMX configuration options, depending upon [the approach](#option-1-opinionated-approach-expose-specific-configs).

## Overview of Solution

CAPV already supports configuring custom VMX keys.
```bash
$ kubectl explain vspheremachinetemplate.spec.template.spec.customVMXKeys
KIND:     VSphereMachineTemplate
VERSION:  infrastructure.cluster.x-k8s.io/v1beta1

FIELD:    customVMXKeys <map[string]string>

DESCRIPTION:
     CustomVMXKeys is a dictionary of advanced VMX options that can be set on VM
     Defaults to empty map
```

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: VSphereMachineTemplate
metadata:
  name: vsphere-quickstart-worker
  namespace: default
spec:
  template:
    spec:
      customVMXKeys:
        "sched.swap.vmxSwapEnabled": "FALSE"
...
```

The plan is for EKS-A to expose the same values in the EKS-A cluster.yaml.

### Solution Details

We can implement this in one of two ways:

1. Expose some VMX options and add more incremently as users ask for it.
1. Expose the custom VMX options as a generic map and allow users to pass in specific config themselves.

#### Option 1: Opinionated approach: Expose specific configs

In this approach, EKS-A will abstract away the exact VMX options and provide a user-friendly way to configure these settings.

The keys and values for the options will be pre-defined types instead of generic strings.

This is what the VSphereMachineConfigSpec API will look like with this approach:
```go
type VSphereMachineConfigSpec struct {
  VMXOptions CustomVMXOptions `json:"vmxOptions,omitempty"`
  ...
}

type CustomVMXOptions struct {
  EnableSwap bool `json:"enableSwap,omitempty"`
  ...                                            // More options can be added as needed
}

type CustomVMXOption string

const (
  EnableSwap CustomVMXOption = "sched.swap.vmxSwapEnabled"
)
```

And the clusterconfig:

```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: VSphereMachineConfig
metadata:
   name: my-cluster-machines
spec:
  diskGiB: 25
  datastore: "Datastore"
  folder: "Folder"
  vmxOptions:
    enableSwap: false
...
```

Pros:

* Users won't have to worry about knowing the exact VMX keys and values for a specific setting like `"sched.swap.vmxSwapEnabled": "FALSE"`. Instead, they can configure it using user-friendly fields like `enableSwap: false`.
* It makes validations a lot easier since EKS-A will know what VMX options and values to expect.

Cons: 

* Limits the number of options users will be allowed to configure since 
* Adding new options will require code changes.


#### Option 2: Generic Map approach

In this approach, EKS-A will expose VMX options as a generic map of strings (`map[string]string`). Users will be expected to pass the exact config option that vSphere supports. 

This is what the VSphereMachineConfigSpec API will look like with this approach:
```go
type VSphereMachineConfigSpec struct {
  VMXOptions map[string]string `json:"vmxOptions,omitempty"`
  ...
}
```

And the clusterconfig:

```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: VSphereMachineConfig
metadata:
   name: my-cluster-machines
spec:
  diskGiB: 25
  datastore: "Datastore"
  folder: "Folder"
  vmxOptions:
    "sched.swap.vmxSwapEnabled": "FALSE"
...
```

Pros:

* Allows users to configure any settings they want.
* EKS-A doesn't have to manage types for different keys and values for VMX settings.

Cons:

* Validations are going to be hard since we won't be able to validate every possible setting.
* Misconfigured settings would cause VM failures that can't be caught during validations.
