---
title: "OIDC configuration"
linkTitle: "OIDC"
weight: 30
description: >
  EKS Anywhere cluster yaml specification OIDC reference
---


## OIDC support (optional)
EKS Anywhere can create clusters that support API server OIDC authentication.
In order to add OIDC support, you need to configure your cluster by updating the configuration file before creating the cluster.
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
List of identity providers you want configured for the cluster.
Right now, only 1 provider of type `OIDCConfig` is supported.
This would include a reference to the `OIDCConfig` object with the configuration below.

### clientId (required)
* Description: Client ID for the OpenID Connect client.
* Type: string
### groupsClaim (optional)
* Description: Name of a custom OpenID Connect claim for specifying user groups.
* Type: string
### groupsPrefix (optional)
* Description: String to be prefixed to all groups to prevent conflicts with other authentication strategies.
* Type: string
### issuerUrl (required)
* Description: URL of the OpenID issuer. Only HTTPS scheme will be accepted.
* Type: string
### requiredClaims (optional)
List of RequiredClaim objects listed below. 
Only one is supported at this time.

### requiredClaims[0] (optional)
* Description: A key=value pair that describes a required claim in the ID Token.
  * claim
    * type: string
  * value
    * type: string
* Type: object
### usernameClaim (optional)
* Description: OpenID claim to use as the username.
Note that claims other than the default ('sub') are not guaranteed to be unique and immutable.
* Type: string
### usernamePrefix (optional)
* Description: UsernamePrefix defines a string to be prefixed to all usernames.
If not provided, username claims other than 'email' are prefixed by the issuer URL to avoid clashes.
To skip any prefixing, provide the value '-'.
* Type: string

