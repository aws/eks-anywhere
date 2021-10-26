---
title: "Upgrade core components"
linkTitle: "Upgrade core components"
weight: 20
date: 2017-01-05
description: >
How to upgrade core components in your cluster.
---
EKS Anywhere `upgrade` also supports upgrading the following core components:

* Core CAPI
* CAPI providers
* Cert-manager
* Etcdadm CAPI provider
* EKS Anywhere controllers and CRDs
* Flux

The latest versions of these core EKS Anywhere components are embedded into a bundles manifest that the CLI uses to fetch the latest versions and image builds needed for each component upgrade.

---
### Performing a core components upgrade
To upgrade to the latest version of the core EKS Anywhere components, run the same upgrade command as you would when upgrading the cluster:
```
eksctl anywhere upgrade cluster -f ./cluster.yaml
```
In addition to upgrading other cluster attributes that you may have modified in your cluster specification, any core components with newer versions will also be upgraded.
You can see which components were upgraded by running `kubectl get pods -A` and observing the updated `Age` for the upgraded components.