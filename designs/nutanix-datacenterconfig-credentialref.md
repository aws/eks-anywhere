# Nutanix Datacenter Config Credential Reference

## Introduction

**Problem:** When a management cluster and a workload cluster are present on two different Nutanix Prism Central instances,
using same credentials to authenticate to both instances is not possible. To remedy this, we need to associate credentials
with corresponding Nutanix Datacenter Configs.

## Overview of Solution
With this feature, a user can specify credentials reference (i.e. Name of the secret in eksa-system namespace containing the Prism credentials)
in the Nutanix Datacenter Config. This info will be fetched and used to authenticate to the Nutanix Prism Central instance
when creating the workload cluster.

### Solution Details
Extend the `NutanixDatacenterConfig` CRD to include a new field `credentialRef` storing name (name of kubernetes secret object) and
kind ("Secret"). This is similar to how `credentialsRef` is defined in the `NutanixCluster` CRD minus the namespace. The namespace
of the secret will always be `eksa-system`. Keeping the namespace constant will allow us to avoid granting EKS-A controller access to
secrets in other namespaces. The `Ref` type used in Snow provider is reused here for this purpose.

```go
// NutanixDatacenterConfigSpec defines the desired state of NutanixDatacenterConfig.
type NutanixDatacenterConfigSpec struct {
    ...
    // CredentialRef is the reference to the secret name that contains the credentials
    // for the Nutanix Prism Central. The namespace for the secret is assumed to be a constant i.e. eksa-system.
    // +optional
    CredentialRef *Ref `json:"credentialRef,omitempty"`
}

func (in *NutanixDatacenterConfig) Validate() error {
    ...
    if in.Spec.CredentialRef != nil && in.Spec.CredentialRef.Kind != RefKindSecret {
        return errors.New("only Secret is supported as a credential kind")
    }
    ...
}
```

When creating a workload cluster, the `NutanixDatacenterConfig` referenced in the `Cluster` CRD will be fetched and
the `credentialRef` field will be used to fetch the secret to get the credentials needed to authenticate with Prism Central
to create resources needed for the cluster.

When creating a NutanixDatacenterConfig, if the `credentialRef` field is not present the validation webhook will return
a validation failure. When updating an existing NutanixDatacenterConfig the validation webhook will only return a validation
failure if the `credentialRef` field is being removed and set to nil.

During management cluster creation, we can fetch the secrets from the environment, create the secret manifests in `eksa-system`
as we do now, and add the `credentialRef` to the corresponding `NutanixDatacenterConfig` if not present. This will allow
us to assume that `credentialRef` will always be set in the `NutanixDatacenterConfig` in all cases.

The NutanixDatacenterReconciler and the NutanixClusterReconciler can ensure the credentialRef is set in the NutanixDatacenterConfig.
If the `credentialRef` is not set, the reconcilers can set them and return and the future iterations will have the credentialRef set.

We will continue to create a cluster-specific secret for CAPX provider named after the cluster name as we have been doing so far. This way
from a users perspective secrets can be shared across clusters at a EKS-A level, but at a CAPX level, each cluster will have
its own decoupled secret. This will allow us to handle the format changes (if any) for the CAPX secret in the future. In future,
we will also add support for watching the referenced secrets and use `OwnerReferences` to link back secrets to Clusters.
