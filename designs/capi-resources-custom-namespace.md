# CAPI Resources in Custom Namespace

## Introduction

**Problem:** Customers managing large scale EKS-A clusters (hundreds or thousands of workload clusters for a given management cluster) expect to have difficulty managing all these resources if they are all in the same namespace. Specifically, they expect to encounter two problems:

1. Difficulty troubleshooting issues for a single cluster when there are so many resources to sift through. 
2. High levels of load on a bootstrap cluster when executing an upgrade, since the clusterctl move operation will take all the CAPI components inside a namespace and instatiate them on the bootstrap cluster. This can result in heavy and unnecessary load on the bootstrap cluster

This document aims to solve that, by allowing customers to instantiate the underlying CAPI/CAPX/etcd resources in a newly defined namespace, rather than all being dumped into `eksa-system`. 
This change would make it so customers would be able to put these resources virtually in any namespace, despite the fact that they may be considered eks-a internal components.

### Tenets

* ****Secure By Default:**** make everything simple and secure by default, but give options to change those defaults if there is a need

### Goals and Objectives

* As a EKS-A cluster administrator I want to quickly and easily query resources for a given cluster so that I can more easily troubleshooting cluster issues with many workload clusters present
* As an EKS-A cluster administrator, I want to be able to perform cluster upgrades without having to move all the cluster resources onto the bootstrap cluster
* As a cluster administrator I'd like to be able to specify exactly which namespace the components go in (and have EKS-A stop me if I try to put it in default or kube-system)

### Statement of Scope

**In scope**
* Allow users to indicate which namespace they'd like to use for their cluster resources

**Not in scope**
* Support for customizing which resources are going into which namespace at different layers (e.g. CAPC resources into ns1, CAPI resources into ns2, eks-a resources into ns3, etc.)

**Future scope**
* N/A
 
## Current state
1. Top level eks-a resources (e.g. Cluster, *DatacenterConfig, *MachineConfig) all go into the "default" namespace by default but can be customized to any namespace. Ideally, customers should only ever interact with these top-level eks-a resources
2. Underlying CAPI/CAPX and etcd resources are created in the eksa-system, to be hidden from customers as implementation details

## Overview of Solution
1. Introduce a attribute in the EKS-A Cluster CRD spec which allows users to indicate which namespace they would like the underlying resources for their new cluster to be created in

## Implementation

Currently, the `eksa-system` namespace is hardcoded throughout the entire codebase for underlying resources. It's used in
1. providers' template generation logic, both by the CLI and the EKS-A controller
2. kubectl and clusterctl executables for interacting with the cluster and moving resources between clusters
3. eks-a controller for retrieving capi resources with FetchObjectByName method

We would need to implement the ability to make namespace configurable in all of these usages, to be extracted from the top level EKS-A Cluster object

## Unknowns

1. Where else is the `eksa-system` namespace hardcoded and assumed to be present?
2. Would we expect any permissions issues with these resources existing in a custom namespace?

## Testing

We should add at least one E2E test for the whole flow:
1. Specify custom namespace for eks-a Cluster, with all underlying resources being created in that same namespace for the cluster
2. Create/upgrade/delete succeeds

## Documentation

We will need to add documentation instructing users how to use this new feature, as well as the security implications it may have relating to rbac and resource accessibility.
