# Supporting Cloudstack clusters across endpoints

## Introduction

**Problem:** 

The mangement API endpoint for Apache Cloudstack (ACS) is a singe point of failure. If the endpoint goes down, then control of all of all VM's, networks, zones, accounts, domains, and everything else on Cloudstack goes down. So, we want to spread our clusters across many Cloudstack endpoints and hosts to protect against that.

For scalability, multiple Cloudstack endpoints will likely be required for storage and API endpoint throughput. Just one cluster creation triggers as many as a thousand API calls to ACS (estimated). There are many ways to support this scale, but adding more Cloudstack hosts and endpoints is a fairly foolproof way to do so. Then, there’s the size and performance of the underlying database that each Cloudstack instance runs on.

In CAPC, we are considering addressing the problem by extending our use of the concept of [Failure Domains](https://cluster-api.sigs.k8s.io/developer/providers/v1alpha2-to-v1alpha3.html?highlight=failure%20domain#optional-support-failure-domains) and distributing a cluster across the given ones. However, instead of a failure domain consisting of a zone on a single Cloudstack endpoint, we will redefine it to consist of the unique combination of a Cloudstack zone, api endpoint, account, and domain. In order to support this functionality in EKS-A, we need to have a similar breakdown where an EKS-A cluster can span across multiple endpoints and Failure Domains.

### Tenets

* ****Simple:**** simple to use, simple to understand, simple to maintain
* ****Declarative:**** intent based system, as close to a Kubernetes native experience as possible

### Goals and Objectives

As a Kubernetes administrator I want to:

* Perform preflight checks when creating/upgrading clusters which span across multiple failure domains
* Create EKS Anywhere clusters which span across multiple failure domains 
* Upgrade EKS Anywhere clusters which span across multiple failure domains
* Delete EKS Anywhere clusters which span across multiple failure domains

### Statement of Scope

**In scope**

* Add support for create/upgrade/delete of EKS-A clusters across multiple Cloudstack API endpoints
* Add test environment for CI/CD e2e tests which can be used as a second Cloudstack API endpoint 

**Not in scope**

* 

**Future scope**

* Multiple network support to handle IP address exhaustion within a zone

## Overview of Solution

We propose to take the least invasive solution of repurposing the CloudstackDataCenterConfig to point to multiple failure domains, each of which contains the necessary
information for interacting with a Cloudstack failure domain. The assumption is that the necessary Cloudstack resources (i.e. image, computeOffering, ISOAttachment, etc.)
will be available on *all* the Cloudstack API endpoints. 

## Solution Details

### Interface changes
Currently, the CloudstackDataCenterConfig spec contains:
```
// Domain contains a grouping of accounts. Domains usually contain multiple accounts that have some logical relationship to each other and a set of delegated administrators with some authority over the domain and its subdomains
Domain string `json:"domain"`
// Zones is a list of one or more zones that are managed by a single CloudStack management endpoint.
Zones []CloudStackZone `json:"zones"`
// Account typically represents a customer of the service provider or a department in a large organization. Multiple users can exist in an account, and all CloudStack resources belong to an account. Accounts have users and users have credentials to operate on resources within that account. If an account name is provided, a domain must also be provided.
Account string `json:"account,omitempty"`
// CloudStack Management API endpoint's IP. It is added to VM's noproxy list
ManagementApiEndpoint string `json:"managementApiEndpoint"`
```

We would instead propose to remove all the existing attributes and instead, simply include a list of FailureDomain objects, where each FailureDomain object looks like
```
// Domain contains a grouping of accounts. Domains usually contain multiple accounts that have some logical relationship to each other and a set of delegated administrators with some authority over the domain and its subdomains
// This field is considered as a fully qualified domain name which is the same as the domain path without "ROOT/" prefix. For example, if "foo" is specified then a domain with "ROOT/foo" domain path is picked.
// The value "ROOT" is a special case that points to "the" ROOT domain of the CloudStack. That is, a domain with a path "ROOT/ROOT" is not allowed.
Domain string `json:"domain"`
// Zones is a list of one or more zones that are managed by a single CloudStack management endpoint.
Zone CloudStackZone `json:"zone"`
// Account typically represents a customer of the service provider or a department in a large organization. Multiple users can exist in an account, and all CloudStack resources belong to an account. Accounts have users and users have credentials to operate on resources within that account. If an account name is provided, a domain must also be provided.
Account string `json:"account,omitempty"`
// CloudStack Management API endpoint's IP. It is added to VM's noproxy list
ManagementApiEndpoint string `json:"managementApiEndpoint"`
```

and we would parse these resources and pass them into CAPC by modifying the templates we have currently implemented. We can then use this new model to read in credentials, perform pre-flight checks, plumb data to CAPC, and support upgrades in the controller. The goal would be to make these new resources backwards compatible via code

### Failure Domain

A failure domain is a CAPI concept which serves to improve HA and availability by destributing machines across "failure domains", as discussed [here](https://cluster-api.sigs.k8s.io/developer/providers/v1alpha2-to-v1alpha3.html?highlight=domain#optional-support-failure-domains). 
CAPC currently utilizes them to distribute machines across CloudStack Zones. However, we now want to go a step further and consider the following unique combination to be a failure domain:

1. Cloudstack endpoint
2. Cloudstack domain
3. Cloudstack zone
4. Cloudstack account

You can find more information about these Cloudstack resources [here](http://docs.cloudstack.apache.org/en/latest/conceptsandterminology/concepts.html#cloudstack-terminology)

### `CloudstackDatacenterConfig` Validation

With the multi-endpoint system for the Cloudstack provider, users reference a CloudstackMachineConfig and it's created across multiple failure domains. The implication
is that all the Cloudstack resources such as image, ComputeOffering, ISOAttachment, etc. must be available in *all* the failure domains, or all the Cloudstack endpoints,
and these resources must be referenced by name, not unique ID. This would mean that for each CloudstackMachineConfig, we have to make sure that all the prerequisite
Cloudstack resources are available in all the Cloudstack API endpoints. 

In practice, the pseudocode would look like:

for failureDomain in failureDomains:
  for machineConfig in machineConfigs:
    validate resource presence with the failureDomain's configuration of the CloudMonkey executable

### Cloudstack credentials


In a multi-endpoint Cloudstack cluster, each endpoint may have its own credentials. We propose that Cloudstack credentials will be passed in via environment variable in the same way as they are currently,
only as a list corresponding to failure domains. Currently, these credentials are passed in via environment variable, which contains a base64 encoded .ini file that looks like

```
[Global]
api-key    = redacted
secret-key = redacted
api-url    = http://172.16.0.1:8080/client/api
```

We would propose an extension of the above input mechanism so the user could provide credentials across multiple Cloudstack API endpoints like

```
[Global]
api-key    = redacted
secret-key = redacted
api-url    = http://172.16.0.1:8080/client/api

[FailureDomain2]
api-key    = redacted
secret-key = redacted
api-url    = http://172.16.0.2:8080/client/api

[FailureDomain3]
api-key    = redacted
secret-key = redacted
api-url    = http://172.16.0.3:8080/client/api

...
```

We are also exploring converting the ini file to a yaml input file which contains a list of credentials and their associated endpoints. Either way, this environment variable would
be passed along to CAPC and used by the CAPC controller just like it is currently.

### Backwards Compatibility

In order to support backwards compatibility in the CloudstackDatacenterConfig resource for users with existing clusters, we can
1. Make all the fields optional and see if the user has the old fields set or the new ones, then write a transformer to set the new fields and clean up the old ones
2. Introduce an eks-a version bump with conversion webhooks

Between these two approaches, I would take the first and then deprecate the legacy fields in a subsequent release to simplify the code paths.

However, given that the Cloudstack credentials are persisted in a write-once secret on the cluster, upgrading existing clusters may not be feasible unless CAPC supports overwriting that secret.

## User Experience


## Security

The main change regarding security is the additional credential management. Otherwise, we are doing exactly the same operations - preflight check with cloudmonkey,
create yaml templates and apply them, and then read/write eks-a resources in the eks-a controller. The corresponding change is an extension of an existing mechanism
and there should not be any new surface area for risks than there was previously.

## Testing

The new code will be covered by unit and e2e tests, and the e2e framework will be extended to support cluster creation across multiple Cloudstack API endpoints.

The following e2e test will be added:

simple flow cluster creation/scaleu/deletion across multiple Cloudstack API endpoints:

* create a management+workload cluster spanning multiple Cloudstack API endpoints
* scale the size of the management+workload cluster so that we touch multiple Cloudstack API endpoints
* delete cluster

## Other approaches explored

1. Another direction we can go to support this feature is to refactor the entire EKS-A codebase so that instead of all the failure domains existing inside the CloudstackDatacenterConfig object, each CloudstackDatacenterConfig itself corresponds with a single failure domain. Then, the top level EKS-A Cluster object could be refactored to have a list of DatacenterRefs instead of a single one. However, this approach feels extremely invasive to the product and does not provide tangible value to the other providers.
2. Lastly, we can consider the option of introducing a new DatacenterConfig object which represents not one Cloudstack "Availability Zone", but multiple. However, the issue here is that the CloudstackDatacenterConfig already has a list of CloudstackZone objects, so we're essentially already supporting multiple of what could be interpreted as Availability Zones. Adding additional attributes to that concept is a more natural extension of the API, instead of defining a new type
