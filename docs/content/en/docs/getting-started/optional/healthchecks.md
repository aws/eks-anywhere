---
title: "Machine Health Checks"
linkTitle: "Machine Health Checks"
weight: 40
aliases:
    /docs/reference/clusterspec/optional/healthchecks/
description: >
  EKS Anywhere cluster yaml specification for machine health check configuration
---

## Machine Health Checks Support 

#### Provider support details
|                | vSphere | Bare Metal | Nutanix | CloudStack | Snow |
|:--------------:|:-------:|:----------:|:-------:|:----------:|:----:|
| **Supported?** |   ✓	    |     ✓      |   	 ✓   |     ✓      |  ✓   |

You can configure EKS Anywhere to specify timeouts for machine health checks.
A Machine Health Check is a resource which allows users to define conditions under which Machines within a Cluster should be considered unhealthy. A Machine Health Check is defined on a management cluster and scoped to a particular workload cluster. If not configured in the spec, the default values are used to configure the machine health checks. 

Note: Even though the configuration on machine health check timeouts in the EKSA spec is optional, machine health checks are still installed for all clusters using the default timeout values mentioned below.

The following cluster spec shows an example of how to configure health check timeouts:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
   name: my-cluster-name
spec:
   ...
  machineHealthCheck:
    nodeStartupTimeout: "10m0s"
    unhealthyMachineTimeout: "5m0s"
```
## Machine Health Check Spec Details
### __machineHealthCheck__ (optional)
* __Description__: top level key; required to configure machine health check timeouts.
* __Type__: object

### __nodeStartupTimeout__ (optional)
* __Description__: determines how long a Machine Health Check should wait for a Node to join the cluster, before considering a Machine unhealthy.
* __Default__: ```20m0s``` for Tinkerbell provider, ```10m0s``` for all other providers.
* __Minimum Value (If configured)__: ```30s```
* __Type__: string

### __unhealthyMachineTimeout__ (optional)
* __Description__: if the unhealthy condition is matched for the duration of this timeout, the Machine is considered unhealthy.
* __Default__: ```5m0s```
* __Type__: string
