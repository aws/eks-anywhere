# Allow users to configure kube-apiserver flags

## Problem Statement

A customer is currently using OIDC for authenticating the kubernetes service accounts(KSA) and they need some mechanism to configure the kube-apiserver flags for their usecase. The main issue that we are addressing in this document is how we want to allow users to be able to configure these flags.

## Overview of Solution

Allow users to configure the flags by exposing a map in the cluster spec yaml

**Schema:**

```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: mgmt-cluster
spec:
  ...
  controlPlaneConfiguration:
    ...
    # More control plane components can be added here in the future
    apiServerExtraArgs:
      ...
      "service-account-issuer": "https://{my-service-account-issuer-url}"
      "service-account-jwks-uri": "https://{my-service-account-issuer-url}/openid/v1/jwks"
      "service-account-signing-key-file": "/etc/kubernetes/pki/sa.key"
      "service-account-key-file": "/etc/kubernetes/pki/sa.pub"
```

**Validations:**

* Validate that oidc flags are not configured in apiServerExtraArgs if OIDCConfig identity provider is already configured in the spec
* Validate that the feature flag is enabled for configuring apiServerExtraArgs

**Pros:**

* Creates a standard way of exposing any flag for the control plane components
* Gives more flexibility to the users in terms of validating the flag values for the api-server

**Cons:**

* Does not enforce OIDC compliance or any other validations on the allowed values for the flags

## Alternate Solutions

Allow users to configure the flags as a struct field in the cluster spec yaml

**Schema:**

```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: mgmt-cluster
spec:
  ...
  controlPlaneConfiguration:
    ...
    apiServerConfiguration:
      ...
      serviceAccountIssuer: 
      - "https://{my-service-account-issuer-url}"
      serviceAccountJwksUri: "https://{my-service-account-issuer-url}/openid/v1/jwks"
      serviceAccountSigningKeyFile: "/etc/kubernetes/pki/sa.key"
      serviceAccountKeyFile: "/etc/kubernetes/pki/sa.pub"
```

**Validations:**

* Validate that both serviceAccountIssuer and serviceAccountJwksUri have same domain and use https scheme
* Additional set of validations specific to each of the flags

**Pros:**

* Fails fast if any of the flags are misconfigured with invalid values
* Allows enforcing OIDC compliance for the service account flags of the api-server

**Cons:**

* Gives less flexibility to the users for configuring the flags in terms of number of validations
* Does not provide a standard way to configure the flags
* Difficult to validate each and every flag and debug any issues with apiserver

## Implementation Details

```
apiServerExtraArgs:
  "service-account-issuer": "https://{my-service-account-issuer-url}"
  "service-account-jwks-uri": "https://{my-service-account-issuer-url}/openid/v1/jwks"
```

https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/control-plane-flags/#apiserver-flags  
These flags will be fetched from the cluster spec and added to the apiServerExtraArgs in the ClusterConfiguration object during create and upgrade operations for generating control plane CAPI spec

Users need to enable the feature flag `API_SERVER_EXTRA_ARGS_ENABLED=true` to configure the api server flags in the cluster spec. If it's not enabled, then it will throw an error when validating the cluster spec. This is done in order to expose this functionality for now before we determine to support it officially with some more robust validations.

The `service-account-issuer` flag can be configured for both podIamConfig as well as controlPlaneConfiguration to enable both features. If both are configured, the podIamConfig url will be appended to the controlPlaneConfiguration url.

If OIDCConfig is specified in the identityProviderRefs within the spec, then oidc flags cannot be configured in the apiServerExtraArgs and the CLI will throw an error.

## Documentation

We would have to add `controlPlaneConfiguration.apiServerExtraArgs` as an optional configuration for the cluster spec in our EKS-A docs

## Migration plan for existing flags

* Phase 1: We can add more flags to the above options and have validations for the existing flags configured in some other fields to make sure that there is no conflict between them and allow only one of them to be configured
* Phase 2: We can decide on the priority among the existing conflicting fields and if the flags are configured in multiple fields, the one with higher priority will have precedence and will be used in the cluster
* Phase 3: We can deprecate all the lower priority conflicting fields for the existing flags and have only one standardized way of configuring all the flags

## References

* https://github.com/kubernetes/enhancements/tree/master/keps/sig-auth/1393-oidc-discovery
* https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#service-account-issuer-discovery
* https://openid.net/developers/how-connect-works/
* https://auth0.com/docs/get-started/authentication-and-authorization-flow/authorization-code-flow

