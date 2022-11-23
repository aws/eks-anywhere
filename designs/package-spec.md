# Package Controller Configuration

## Introduction

**Problem:** currently the ability to configure the package controller is very limited and requires code changes

Code changes currently have to be made to configure the package controller during cluster creation or using the install command. Changes need to be coordinated between the CLI and the package controller helm chart to alter configuration. This makes it more difficult to support things like various private registry configurations.

### Tenets

* ***Flexible:*** reduce dependency of the CLI on helm chart.

### Goals and Objectives

As an EKS-A cluster administrator I want to configure the package controller for:

* Harbor using proxy cache
* Harbor using replication
* Harbor as a private registry using the import images command
* Configure ECR token refresher
* Configure for kubelet private registry authentication
* Support proxy configuration
* Support package controller installation from alternate registries
* Disable installation of package controller
* Allow for future configuration options

### Statement of Scope

**In scope**
* All of the above

## Current state
The package controller can be used currently with:
1. Proxy configurations
1. ECR token refresher enabled/disabled
1. Installation of the package controller from alternate registries

## Overview of Solution
Add a section to the cluster specification to enable/disable the package controller installation and pass configuration values.

## Solution Details

Add a packageController section to the cluster specification for example to use development/staging builds:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: billy
spec:
  packageController:
    enable: true
    values: |-
      defaultRegistry: public.ecr.aws/l0g8r8j6
      defaultImageRegistry: 857151390494.dkr.ecr.us-west-2.amazonaws.com
```

These values will be passed to the helm chart during installation just like `--set defaultRegistry=public.ecr.aws/l0g8r8j6` for example.

## Implementation

The proposed design:
```
// PackageController defines configuration of the package controller.
type PackageController struct {
    // +kubebuilder:validation:Optional
    // +kubebuilder:default:="true"
    // Enable installation of the package controller.
    Enable bool `json:"enable,omitempty"`

    // +kubebuilder:validation:Optional
    // Values to pass into package controller installation.
    Values string `json:"values,omitempty"`
}
```

## Testing

We should add at least one E2E to test disabled package controller, configuration of values will be used to test staging builds.
