package bundles

import (
	"fmt"

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
	bundleutils "github.com/aws/eks-anywhere/release/cli/pkg/util/bundles"
	"github.com/aws/eks-anywhere/release/cli/pkg/version"
)

// GetNutanixBundle returns the bundle for Nutanix.
func GetNutanixBundle(r *releasetypes.ReleaseConfig, imageDigests map[string]string) (anywherev1alpha1.NutanixBundle, error) {
	nutanixBundleArtifacts := map[string][]releasetypes.Artifact{
		"cluster-api-provider-nutanix": r.BundleArtifactsTable["cluster-api-provider-nutanix"],
		"kube-rbac-proxy":              r.BundleArtifactsTable["kube-rbac-proxy"],
		"kube-vip":                     r.BundleArtifactsTable["kube-vip"],
	}
	sortedComponentNames := bundleutils.SortArtifactsMap(nutanixBundleArtifacts)

	var sourceBranch string
	var componentChecksum string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	artifactHashes := make([]string, 0)
	for _, componentName := range sortedComponentNames {
		for _, artifact := range nutanixBundleArtifacts[componentName] {
			if artifact.Image != nil {
				imageArtifact := artifact.Image
				if componentName == "cluster-api-provider-nutanix" {
					sourceBranch = imageArtifact.SourcedFromBranch
				}
				bundleImageArtifact := anywherev1alpha1.Image{
					Name:        imageArtifact.AssetName,
					Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
					OS:          imageArtifact.OS,
					Arch:        imageArtifact.Arch,
					URI:         imageArtifact.ReleaseImageURI,
					ImageDigest: imageDigests[imageArtifact.ReleaseImageURI],
				}
				bundleImageArtifacts[imageArtifact.AssetName] = bundleImageArtifact
				artifactHashes = append(artifactHashes, bundleImageArtifact.ImageDigest)
			}

			if artifact.Manifest != nil {
				manifestArtifact := artifact.Manifest
				bundleManifestArtifact := anywherev1alpha1.Manifest{
					URI: manifestArtifact.ReleaseCdnURI,
				}

				bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact

				manifestHash, err := version.GenerateManifestHash(r, manifestArtifact)
				if err != nil {
					return anywherev1alpha1.NutanixBundle{}, err
				}
				artifactHashes = append(artifactHashes, manifestHash)
			}
		}
	}

	if r.DryRun {
		componentChecksum = version.FakeComponentChecksum
	} else {
		componentChecksum = version.GenerateComponentHash(artifactHashes, r.DryRun)
	}
	capxVersion, err := version.BuildComponentVersion(
		version.NewVersionerWithGITTAG(r.BuildRepoSource, constants.CapxProjectPath, sourceBranch, r),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.NutanixBundle{}, errors.Wrapf(err, "Error getting version for cluster-api-provider-nutanix")
	}

	bundle := anywherev1alpha1.NutanixBundle{
		Version:              capxVersion,
		ClusterAPIController: bundleImageArtifacts["cluster-api-provider-nutanix"],
		KubeVip:              bundleImageArtifacts["kube-vip"],
		Components:           bundleManifestArtifacts["infrastructure-components.yaml"],
		ClusterTemplate:      bundleManifestArtifacts["cluster-template.yaml"],
		Metadata:             bundleManifestArtifacts["metadata.yaml"],
	}

	return bundle, nil
}
