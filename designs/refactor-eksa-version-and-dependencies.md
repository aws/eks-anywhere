# Refactor Cluster Version and Dependencies

This doc is targeted to addressing ways to get the EKS-A CLI version that was last used to create/upgrade a cluster.

## Context

We want to add a preflight validation to prevent customers from skipping minor versions of EKS-A when upgrading a cluster. Currently, there is no trivial way to check the cli version that was last used to manage the cluster. 

## Proposed Solutions

The preferred solution is to add a new field indicating the EKS-A version to the ClusterSpec and update it whenever the CLI creates or upgrades the cluster. This new field can be called EKSAVersion and it would make BundlesRef redundant as the bundle can instead be derived from the version. Therefore, we can deprecate the BundlesRef field and add a way to map versions to bundles.

An alternative solution is to improve the relationship between the bundle and CLI to map from bundle to CLI. Currently, one bundle can have a relation to many CLI versions. We would need a one-to-one relationship between the bundle and CLI in order to map from bundle to CLI. Once that relationship is enforced, the bundle mentioned in the cluster’s BundlesRef can be used to find the CLI version. The first solution is preferred as it provides a direct and elementary way to get the current EKS-A version to use in validations. 

## Scenarios to handle

The preferred solution would require us to handle differences between existing clusters and new clusters regarding the changes to ClusterSpec. 

1. For new clusters, we would need a webhook validation that only admits EKSAVersion while preventing the user from passing the deprecated BundlesRef.
2. For existing clusters where the BundleRef isn’t changing, we can continue to use BundlesRef to get the bundle since these clusters won’t have EKSAVersion.
3. For existing clusters that users try to upgrade with a new bundle, the BundlesRef should be nullified and EKSAVersion should be added. The CLI must ensure EKSAVersion is passed when BundlesRef is nullified. Any change to BundlesRef other than nullification should be rejected.

In any of the three scenarios, only one of the two fields should be passed. There should never be a case where both are passed.

## Implementation Details

Add a new string type and field called EKSAVersion. The new field should go under ClusterSpec. Mark the BundlesRef field as deprecated and mention the newly added field as its replacement. 

To map the versions to bundles, we can introduce a new CRD called EKSARelease. For each EKS-A version, there should be an EKSARelease object in the cluster. The EKSARelease object should include the version scheme in its name (e.g. eksa-v0-15-3) to make it easy to derive the EKSARelease object from the EKSAVersion field. There should be a function, which handles this conversion from EKSAVersion to EKSARelease.ObjectMeta.Name.

The EKSARelease CRD should have a spec containing a BundlesRef and EksARelease object. From this point, the BundlesRef can be used to get the Bundle.

Instances of EKSARelease can be created from the releases manifest. Rather than making a get call to download the releases manifest, the manifest can be added to the EKS-A repo and embedded into the CLI at compile time.  The release process should be modified to dynamically update the file within the EKS-A repo. 

`type EKSAVersion string`

```
type EKSARelease struct {
  metav1.TypeMeta
  metav1.ObjectMeta
  Spec EKSAReleaseSpec
}
```

```
type EKSAReleaseSpec struct {
  *BundlesRef
  EksARelease
}              
```