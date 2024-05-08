---
title: "MachineHealthCheck"
linkTitle: "MachineHealthCheck"
weight: 40
aliases:
    /docs/reference/clusterspec/optional/healthchecks/
description: >
  EKS Anywhere cluster yaml specification for MachineHealthCheck configuration
---

## MachineHealthCheck Support

#### Provider support details
|                | vSphere | Bare Metal | Nutanix | CloudStack | Snow |
|:--------------:|:-------:|:----------:|:-------:|:----------:|:----:|
| **Supported?** |   ✓	    |     ✓      |   	 ✓   |     ✓      |  ✓   |

You can configure EKS Anywhere to specify timeouts and `maxUnhealthy` values for machine health checks.

A MachineHealthCheck (MHC) is a resource in Cluster API which allows users to define conditions under which Machines within a Cluster should be considered unhealthy. A MachineHealthCheck is defined on a management cluster and scoped to a particular workload cluster.

Note: Even though the MachineHealthCheck configuration in the EKS-A spec is optional, MachineHealthChecks are still installed for all clusters using the default values mentioned below.

EKS Anywhere allows users to have granular control over MachineHealthChecks in their cluster configuration, with default values (derived from Cluster API) being applied if the MHC is not configured in the spec. The top-level `machineHealthCheck` field governs the global MachineHealthCheck settings for all Machines (control-plane and worker). These global settings can be overridden through the nested `machineHealthCheck` field in the control plane configuration and each worker node configuration. If the nested MHC fields are not configured, then the top-level settings are applied to the respective Machines.

The following cluster spec shows an example of how to configure health check timeouts and `maxUnhealthy`:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
   name: my-cluster-name
spec:
   ...
  machineHealthCheck:               # Top-level MachineHealthCheck configuration
    maxUnhealthy: "60%"
    nodeStartupTimeout: "10m0s"
    unhealthyMachineTimeout: "5m0s"
   ...
 controlPlaneConfiguration:         # MachineHealthCheck configuration for Control plane
    machineHealthCheck:
      maxUnhealthy: 100%
      nodeStartupTimeout: "15m0s"
      unhealthyMachineTimeout: 10m
   ...
  workerNodeGroupConfigurations:
  - count: 1
    name: md-0
    machineHealthCheck:             # MachineHealthCheck configuration for Worker Node Group 0
      maxUnhealthy: 100%
      nodeStartupTimeout: "10m0s"
      unhealthyMachineTimeout: 20m
  - count: 1
    name: md-1
    machineHealthCheck:             # MachineHealthCheck configuration for Worker Node Group 1
      maxUnhealthy: 100%
      nodeStartupTimeout: "10m0s"
      unhealthyMachineTimeout: 20m
   ...
```
## MachineHealthCheck Spec Details
### __machineHealthCheck__ (optional)
* __Description__: top-level key; required to configure global MachineHealthCheck timeouts and `maxUnhealthy`.
* __Type__: object

### __machineHealthCheck.maxUnhealthy__ (optional)
* __Description__: determines the maximum permissible number or percentage of unhealthy Machines in a cluster before further remediation is prevented. This ensures that MachineHealthChecks only remediate Machines when the cluster is healthy.
* __Default__: ```100%``` for control plane machines, ```40%``` for worker nodes (Cluster API defaults).
* __Type__: integer (count) or string (percentage)

### __machineHealthCheck.nodeStartupTimeout__ (optional)
* __Description__: determines how long a MachineHealthCheck should wait for a Node to join the cluster, before considering a Machine unhealthy.
* __Default__: ```20m0s``` for Tinkerbell provider, ```10m0s``` for all other providers.
* __Minimum Value (If configured)__: ```30s```
* __Type__: string

### __machineHealthCheck.unhealthyMachineTimeout__ (optional)
* __Description__: determines how long the unhealthy Node conditions (e.g., `Ready=False`, `Ready=Unknown`) should be matched for, before considering a Machine unhealthy.
* __Default__: ```5m0s```
* __Type__: string

### __controlPlaneConfiguration.machineHealthCheck__ (optional)
* __Description__: Control plane level configuration for MachineHealthCheck timeouts and `maxUnhealthy` values.
* __Type__: object

### __controlPlaneConfiguration.machineHealthCheck.maxUnhealthy__ (optional)
* __Description__: determines the maximum permissible number or percentage of unhealthy control plane Machines in a cluster before further remediation is prevented. This ensures that MachineHealthChecks only remediate Machines when the cluster is healthy.
* __Default__: Top-level MHC `maxUnhealthy` if set or ```100%``` otherwise.
* __Type__: integer (count) or string (percentage)

### __controlPlaneConfiguration.machineHealthCheck.nodeStartupTimeout__ (optional)
* __Description__: determines how long a MachineHealthCheck should wait for a control plane Node to join the cluster, before considering the Machine unhealthy.
* __Default__: Top-level MHC `nodeStartupTimeout` if set or ```20m0s``` for Tinkerbell provider, ```10m0s``` for all other providers otherwise.
* __Minimum Value (if configured)__: ```30s```
* __Type__: string

### __controlPlaneConfiguration.machineHealthCheck.unhealthyMachineTimeout__ (optional)
* __Description__: determines how long the unhealthy conditions (e.g., `Ready=False`, `Ready=Unknown`) should be matched for a control plane Node, before considering the Machine unhealthy.
* __Default__: Top-level MHC `nodeStartupTimeout` if set or ```5m0s``` otherwise.
* __Type__: string

### __workerNodeGroupConfigurations.machineHealthCheck__ (optional)
* __Description__: Worker node level configuration for MachineHealthCheck timeouts and `maxUnhealthy` values.
* __Type__: object

### __workerNodeGroupConfigurations.machineHealthCheck.maxUnhealthy__ (optional)
* __Description__: determines the maximum permissible number or percentage of unhealthy worker Machines in a cluster before further remediation is prevented. This ensures that MachineHealthChecks only remediate Machines when the cluster is healthy.
* __Default__: Top-level MHC `maxUnhealthy` if set or ```40%``` otherwise.
* __Type__: integer (count) or string (percentage)

### __workerNodeGroupConfigurations.machineHealthCheck.nodeStartupTimeout__ (optional)
* __Description__: determines how long a MachineHealthCheck should wait for a worker Node to join the cluster, before considering the Machine unhealthy.
* __Default__: Top-level MHC `nodeStartupTimeout` if set or ```20m0s``` for Tinkerbell provider, ```10m0s``` for all other providers otherwise.
* __Minimum Value (if configured)__: ```30s```
* __Type__: string

### __workerNodeGroupConfigurations.machineHealthCheck.unhealthyMachineTimeout__ (optional)
* __Description__: determines how long the unhealthy conditions (e.g., `Ready=False`, `Ready=Unknown`) should be matched for a worker Node, before considering the Machine unhealthy.
* __Default__: Top-level MHC `nodeStartupTimeout` if set or ```5m0s``` otherwise.
* __Type__: string
