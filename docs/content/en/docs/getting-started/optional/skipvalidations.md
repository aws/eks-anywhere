---
title: "Skipping validations configuration"
linkTitle: "Skipping Validations"
weight: 45
aliases:
    /docs/reference/clusterspec/optional/skip-validations/
description: >
 EKS Anywhere cluster annotations to skip validations
---

EKS Anywhere runs a set of validations while performing cluster operations. Some of these validations can be chosen to be skipped.

One such validation EKS Anywhere runs is a check for whether cluster's control plane ip is in use or not.
- If a cluster is being created using the EKS Anywhere cli, this validation can be skipped by using the `--skip-ip-check` flag or by setting the below annotation on the `Cluster` object.
- If a workload cluster is being created using tools like `kubectl` or `GitOps`, the validation can only be skipped by adding the below annotation.


Configure an EKS Anywhere cluster to skip the validation for checking the uniqueness of the control plane IP by using the `anywhere.eks.amazonaws.com/skip-ip-check` annotation and setting the value to `true` like shown below.

```yaml
    apiVersion: anywhere.eks.amazonaws.com/v1alpha1
    kind: Cluster
    metadata:
      annotations:
        anywhere.eks.amazonaws.com/skip-ip-check: "true"
      name: my-cluster-name
    spec:
    .
    .
    .

```

Note that this annotation is also automatically set if you use the `--skip-ip-check` flag while running the EKS Anywhere create cluster command.

