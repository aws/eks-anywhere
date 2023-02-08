# Package Controller Configuration

## Introduction

**Problem:** currently the ability to configure the package controller is very limited and requires code changes

The preference in the CLI is to validate the values passed to the package controller, so the EKS Anywhere cluster specification will include all the necessary configuration values for the controller.

### Tenets

* ***Validation:*** allow the customer the flexibility to configure the package controller and validate those values.

### Goals and Objectives

As an EKS-A cluster administrator I want to configure the package controller for:

* Allow packages only to be configured from a private registry
* Harbor using proxy cache
* Harbor using replication
* Harbor as a private registry using the import images command
* Configure ECR token refresher
* Configure for kubelet private registry authentication
* Support proxy configuration
* Support package controller installation from alternate registries
* Disable installation of package controller

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

A sample package controller section in the cluster specification. The values specified here:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: billy
spec:
  packages:
    disable: false
    controller:
      # -- Controller repository name.
      repository: "eks-anywhere-packages"
      # -- Controller image tag
      tag: "{{eks-anywhere-packages-tag}}"
      # -- Controller image digest
      digest: "{{eks-anywhere-packages}}"
      # -- Whether to turn on Webhooks for the controller image
      enableWebhooks: "true"
      # -- Additional environment variables for the controller pod.
      # - name: EKSA_PUBLIC_KEY
      #   value: ""
      env:
        - name: EKSA_PUBLIC_KEY
          value: "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEnP0Yo+ZxzPUEfohcG3bbJ8987UT4f0tj+XVBjS/s35wkfjrxTKrVZQpz3ta3zi5ZlgXzd7a20B1U1Py/TtPsxw=="
      resources:
        requests:
          cpu: 100m
          memory: 50Mi
        limits:
          cpu: 750m
          memory: 450Mi
    # Additional Variables to configure cronjob
    cronjob:
      name: ecr-refresher
      # -- ECR refresher repository name.
      repository: "ecr-token-refresher"
      # -- ECR refresher tag
      tag: "{{ecr-token-refresher-tag}}"
      # -- ECR refresher digest
      digest: "{{ecr-token-refresher}}"
      suspend: false
    registryMirrorSecret:
      endpoint: ""
      cacertcontent: ""
      insecure: "ZmFsc2UK"
```

These values will be passed to the helm chart during installation using a values.yaml file to protect the secrets.
A run-time validation will be performed to verify a controller section is not included for a workload cluster.

## Testing

We should add at least one E2E to test disabled package controller, configuration of values will be used to test staging builds.
