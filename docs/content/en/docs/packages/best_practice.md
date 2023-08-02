---
title: "Configuration best practices"
linkTitle: "Best practices"
weight: 6
aliases:
    /docs/reference/packagespec/best_practice/
description: >
  Best practices with curated packages
---


### Best Practice

Any supported EKS Anywhere curated package should be modified through package yaml files (with `kind: Package`) and applied through the command `eksctl anywhere apply package -f packageFileName`. Modifying objects outside of package yaml files may lead to unpredictable behaviors.

For automatic namespace (targetNamespace) creation, see `createNamespace` field: [PackagebundleController.spec]({{< ref "packages.md/#packagebundlecontrollerspec" >}})
