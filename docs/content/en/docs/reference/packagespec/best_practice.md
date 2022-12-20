---
title: "Configuration Best Practice"
linkTitle: "Best Practice"
weight: 15
description: >
---


### Best Practice
Any package configuration options listed under `Reference/Packages` should be modified through package yaml files (with `kind: Package`) through command `eksctl anywhere apply package -f packageFileName`. Modifying objects outside of package yaml files may lead to unpredictable behaviors.

For automatic namespace (targetNamespace) creation, see `createNamespace` field: [PackagebundleController.spec]({{< ref "packages.md/#packagebundlecontrollerspec" >}}) 
