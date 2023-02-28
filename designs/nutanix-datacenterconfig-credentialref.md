# Nutanix Datacenter Config Credential Reference

## Introduction

**Problem:** When a management cluster and a workload cluster are present on two different Nutanix Prism Central instances,
using same credentials to authenticate to both instances is not possible. To remedy this, we need to associate credentials
with corresponding Nutanix Datacenter Configs.

## Overview of Solution
With this feature, a user can specify credentials reference (i.e. Name and Namespace of the secret containing the credentials)
in the Nutanix Datacenter Config. This info will be fetched and used to authenticate to the Nutanix Prism Central instance
when creating the workload cluster.

### Solution Details
Extend the `NutanixDatacenterConfig` CRD to include a new field `credentialRef` storing name, namespace, and kind (similar to how
`credentialsRef` is defined in the `NutanixCluster` CRD).

```go
// NutanixDatacenterConfigSpec defines the desired state of NutanixDatacenterConfig.
type NutanixDatacenterConfigSpec struct {
    ...
    // CredentialRef is the reference to the secret that contains the credentials
    // for the Nutanix Prism Central
    // +optional
    CredentialRef *NutanixCredentialReference `json:"credentialRef,omitempty"`
}

// NutanixCredentialKind is the kind of the Nutanix credential
type NutanixCredentialKind string

// NutanixCredentialReference is the reference to the secret that contains the credentials
// +kubebuilder:object:generate=true
type NutanixCredentialReference struct {
    // Kind of the Nutanix credential
    // +kubebuilder:validation:Enum=Secret
    Kind NutanixCredentialKind `json:"kind"`
    // Name of the credential.
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:MinLength=1
    Name string `json:"name"`
    // namespace of the credential.
    // +optional
    Namespace string `json:"namespace"`
}
```

When creating a workload cluster, the `NutanixDatacenterConfig` referenced in the `Cluster` CRD will be fetched and
the `credentialRef` field will be used to fetch the secret to get the credentials needed to authenticate with Prism Central
to create resources needed for the cluster. If the `credentialRef` field is not present, the credentials from the
cluster (i.e. eksa-system/cluster-name) will be used as default to preserve existing behaviour.

