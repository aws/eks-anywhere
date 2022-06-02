# CAPI Resources in Custom Namespace

## Introduction

**Problem:** Customers managing large scale EKS-A clusters (hundreds or thousands of workload clusters for a given management cluster) expect to have difficulty managing all these resources if they are all in the same namespace 

This document aims to solve that, by allowing customers to specify a new flag on the cluster spec. If set, the underlying CAPI/CAPX/etcd resources will be created under the Cluster's namespace, rather than all dumped into `eksa-system`. 
This change would make it so customers would be able to put these resources virtually in any namespace, despite the fact that they may be considered eks-a internal components.

### Tenets

****Simple:**** make everything simple and secure by default, but give options to change those defaults if there is a need

### Goals and Objectives

As a EKS-A cluster administrator I want to organize/partition each cluster into a dedicated namespace so that I can more easily manage my workload clusters

### Statement of Scope

**In scope**
* Allow users to indicate whether they'd like all underlying cluster resources to fall into the Cluster's namespace

**Not in scope**
* Support for customizing which resources are going into which namespace at different layers (e.g. CAPC resources into ns1, CAPI resources into ns2, eks-a resources into ns3, etc.)

**Future scope**
* N/A
 
## Current state
1. Top level eks-a resources (e.g. Cluster, *DatacenterConfig, *MachineConfig) all go into the "default" namespace by default but can be customized to any namespace. Ideally, customers should only ever interact with these top-level eks-a resources
2. Underlying CAPI/CAPX and etcd resources are created in the eksa-system, to be hidden from customers as implementation details

## Overview of Solution
1. Introduce a new boolean flag in the EKS-A Cluster CRD spec which allows users to indicate whether they would like the underlying resources to be created in the same namespace as the eks-a Cluster resource

## Implementation

Currently, the `eksa-system` namespace is hardcoded throughout the entire codebase for underlying resources. It's used in
1. providers' template generation logic, both by the CLI and the EKS-A controller
2. kubectl and clusterctl executables for interacting with the cluster and moving resources between clusters
3. eks-a controller for retrieving capi resources with FetchObjectByName method

We would need to implement the ability to make namespace configurable in all of these usages, to be extracted from the top level EKS-A Cluster object and checked against the boolean

## Unknowns

1. Where else is the `eksa-system` namespace hardcoded and assumed to be present?
2. Would we expect any permissions issues with these resources existing in a custom namespace?

## Testing

We should add at least one E2E test for the whole flow:
1. Specify custom namespace for eks-a Cluster, with flag set to indicate underlying resources to be created in the same namespace
2. Create/delete succeeds