---
title: "OIDC"
linkTitle: "OIDC"
weight: 30
aliases:
    /docs/reference/clusterspec/optional/oidc/
description: >
  EKS Anywhere cluster yaml specification OIDC reference
---


## OIDC support (optional)
EKS Anywhere can create clusters that support api server OIDC authentication.

#### Provider support details
|                | vSphere | Bare Metal | Nutanix | CloudStack | Snow |
|:--------------:|:-------:|:----------:|:-------:|:----------:|:----:|
| **Supported?** |   ✓	    |     ✓      |   	 ✓   |     ✓      |  ✓   |

In order to add OIDC support, you need to configure your cluster by updating the configuration file to include the details below. The OIDC configuration can be added at cluster creation time, or introduced via a cluster upgrade in VMware and CloudStack.

This is a generic template with detailed descriptions below for reference:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
   name: my-cluster-name
spec:
   ...
   # OIDC support
   identityProviderRefs:
      - kind: OIDCConfig
        name: my-cluster-name
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: OIDCConfig
metadata:
   name: my-cluster-name
spec:
    clientId: ""
    groupsClaim: ""
    groupsPrefix: ""
    issuerUrl: "https://x"
    requiredClaims:
      - claim: ""
        value: ""
    usernameClaim: ""
    usernamePrefix: ""
```
### identityProviderRefs (Under Cluster)
List of identity providers you want configured for the Cluster.
This would include a reference to the `OIDCConfig` object with the configuration below.

### clientId (required)
* Description: ClientId defines the client ID for the OpenID Connect client
* Type: string
### groupsClaim (optional)
* Description: GroupsClaim defines the name of a custom OpenID Connect claim for specifying user groups
* Type: string
### groupsPrefix (optional)
* Description: GroupsPrefix defines a string to be prefixed to all groups to prevent conflicts with other authentication strategies
* Type: string
### issuerUrl (required)
* Description: IssuerUrl defines the URL of the OpenID issuer, only HTTPS scheme will be accepted
* Type: string
### requiredClaims (optional)
List of RequiredClaim objects listed below. 
Only one is supported at this time.

### requiredClaims[0] (optional)
* Description: RequiredClaim defines a key=value pair that describes a required claim in the ID Token
  * claim
    * type: string
  * value
    * type: string
* Type: object
### usernameClaim (optional)
* Description: UsernameClaim defines the OpenID claim to use as the user name.
Note that claims other than the default ('sub') is not guaranteed to be unique and immutable
* Type: string
### usernamePrefix (optional)
* Description: UsernamePrefix defines a string to be prefixed to all usernames.
If not provided, username claims other than 'email' are prefixed by the issuer URL to avoid clashes.
To skip any prefixing, provide the value '-'.
* Type: string

