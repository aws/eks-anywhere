// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pkg

import (
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const etcdadmControllerProjectPath = "projects/aws/etcdadm-controller"

// GetEtcdadmControllerAssets returns the eks-a artifacts for etcdadm controller
func (r *ReleaseConfig) GetEtcdadmControllerAssets() ([]Artifact, error) {
	gitTag, err := r.readGitTag(etcdadmControllerProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}

	name := "etcdadm-controller"
	repoName := fmt.Sprintf("aws/%s", name)
	tagOptions := map[string]string{
		"gitTag":      gitTag,
		"projectPath": etcdadmControllerProjectPath,
	}

	sourceImageUri, sourcedFromBranch, err := r.GetSourceImageURI(name, repoName, tagOptions)
	if err != nil {
		return nil, errors.Cause(err)
	}
	releaseImageUri, err := r.GetReleaseImageURI(name, repoName, tagOptions)
	if err != nil {
		return nil, errors.Cause(err)
	}

	imageArtifact := &ImageArtifact{
		AssetName:         name,
		SourceImageURI:    sourceImageUri,
		ReleaseImageURI:   releaseImageUri,
		Arch:              []string{"amd64"},
		OS:                "linux",
		GitTag:            gitTag,
		ProjectPath:       etcdadmControllerProjectPath,
		SourcedFromBranch: sourcedFromBranch,
	}
	artifacts := []Artifact{Artifact{Image: imageArtifact}}

	var imageTagOverrides []ImageTagOverride

	kubeRbacProxyImageTagOverride, err := r.GetKubeRbacProxyImageTagOverride()
	if err != nil {
		return nil, errors.Cause(err)
	}

	imageTagOverride := ImageTagOverride{
		Repository: repoName,
		ReleaseUri: imageArtifact.ReleaseImageURI,
	}
	imageTagOverrides = append(imageTagOverrides, imageTagOverride, kubeRbacProxyImageTagOverride)

	manifestList := []string{
		"bootstrap-components.yaml",
		"metadata.yaml",
	}

	for _, manifest := range manifestList {
		var sourceS3Prefix string
		var releaseS3Path string
		latestPath := getLatestUploadDestination(sourcedFromBranch)

		if r.DevRelease || r.ReleaseEnvironment == "development" {
			sourceS3Prefix = fmt.Sprintf("%s/%s/manifests/bootstrap-etcdadm-controller/%s", etcdadmControllerProjectPath, latestPath, gitTag)
		} else {
			sourceS3Prefix = fmt.Sprintf("releases/bundles/%d/artifacts/etcdadm-controller/manifests/bootstrap-etcdadm-controller/%s", r.BundleNumber, gitTag)
		}

		if r.DevRelease {
			releaseS3Path = fmt.Sprintf("artifacts/%s/etcdadm-controller/manifests/bootstrap-etcdadm-controller/%s", r.DevReleaseUriVersion, gitTag)
		} else {
			releaseS3Path = fmt.Sprintf("releases/bundles/%d/artifacts/etcdadm-controller/manifests/bootstrap-etcdadm-controller/%s", r.BundleNumber, gitTag)
		}

		cdnURI, err := r.GetURI(filepath.Join(releaseS3Path, manifest))
		if err != nil {
			return nil, errors.Cause(err)
		}

		manifestArtifact := &ManifestArtifact{
			SourceS3Key:       manifest,
			SourceS3Prefix:    sourceS3Prefix,
			ArtifactPath:      filepath.Join(r.ArtifactDir, "etcdadm-controller-manifests", r.BuildRepoHead),
			ReleaseName:       manifest,
			ReleaseS3Path:     releaseS3Path,
			ReleaseCdnURI:     cdnURI,
			ImageTagOverrides: imageTagOverrides,
			GitTag:            gitTag,
			ProjectPath:       etcdadmControllerProjectPath,
		}
		artifacts = append(artifacts, Artifact{Manifest: manifestArtifact})
	}

	return artifacts, nil
}

func (r *ReleaseConfig) GetEtcdadmControllerBundle(imageDigests map[string]string) (anywherev1alpha1.EtcdadmControllerBundle, error) {
	etcdadmControllerBundleArtifacts := map[string][]Artifact{
		"etcdadm-controller": r.BundleArtifactsTable["etcdadm-controller"],
		"kube-rbac-proxy":    r.BundleArtifactsTable["kube-rbac-proxy"],
	}
	sortedComponentNames := sortArtifactsMap(etcdadmControllerBundleArtifacts)

	var sourceBranch string
	var componentChecksum string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	artifactHashes := []string{}

	for _, componentName := range sortedComponentNames {
		for _, artifact := range etcdadmControllerBundleArtifacts[componentName] {
			if artifact.Image != nil {
				imageArtifact := artifact.Image
				if componentName == "etcdadm-controller" {
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

				manifestHash, err := r.GenerateManifestHash(manifestArtifact)
				if err != nil {
					return anywherev1alpha1.EtcdadmControllerBundle{}, err
				}
				artifactHashes = append(artifactHashes, manifestHash)
			}
		}
	}

	if r.DryRun {
		componentChecksum = fakeComponentChecksum
	} else {
		componentChecksum = generateComponentHash(artifactHashes)
	}
	version, err := BuildComponentVersion(
		newVersionerWithGITTAG(r.BuildRepoSource, etcdadmControllerProjectPath, sourceBranch, r),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.EtcdadmControllerBundle{}, errors.Wrapf(err, "Error getting version for etcdadm-controller")
	}

	bundle := anywherev1alpha1.EtcdadmControllerBundle{
		Version:    version,
		Controller: bundleImageArtifacts["etcdadm-controller"],
		KubeProxy:  bundleImageArtifacts["kube-rbac-proxy"],
		Components: bundleManifestArtifacts["bootstrap-components.yaml"],
		Metadata:   bundleManifestArtifacts["metadata.yaml"],
	}

	return bundle, nil
}
