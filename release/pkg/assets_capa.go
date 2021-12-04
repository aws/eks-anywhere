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

const capaProjectPath = "projects/kubernetes-sigs/cluster-api-provider-aws"

// GetCapaAssets returns the eks-a artifacts for CAPA
func (r *ReleaseConfig) GetCapaAssets() ([]Artifact, error) {
	gitTag, err := r.readGitTag(capaProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}

	capaImages := []string{
		"cluster-api-aws-controller",
		"eks-bootstrap-controller",
		"eks-control-plane-controller",
	}

	var imageTagOverrides []ImageTagOverride
	artifacts := []Artifact{}

	for _, image := range capaImages {
		var imageTagOverride ImageTagOverride
		repoName := fmt.Sprintf("kubernetes-sigs/cluster-api-provider-aws/%s", image)
		tagOptions := map[string]string{
			"gitTag":      gitTag,
			"projectPath": capaProjectPath,
		}

		sourceImageUri, err := r.GetSourceImageURI(image, repoName, tagOptions)
		if err != nil {
			return nil, errors.Cause(err)
		}
		releaseImageUri, err := r.GetReleaseImageURI(image, repoName, tagOptions)
		if err != nil {
			return nil, errors.Cause(err)
		}

		imageArtifact := &ImageArtifact{
			AssetName:       image,
			SourceImageURI:  sourceImageUri,
			ReleaseImageURI: releaseImageUri,
			Arch:            []string{"amd64"},
			OS:              "linux",
			GitTag:          gitTag,
			ProjectPath:     capaProjectPath,
		}

		artifacts = append(artifacts, Artifact{Image: imageArtifact})

		if image == "cluster-api-aws-controller" {
			imageTagOverride = ImageTagOverride{
				Repository: repoName,
				ReleaseUri: imageArtifact.ReleaseImageURI,
			}
			imageTagOverrides = append(imageTagOverrides, imageTagOverride)
		}
	}

	kubeRbacProxyImageTagOverride, err := r.GetKubeRbacProxyImageTagOverride()
	if err != nil {
		return nil, errors.Cause(err)
	}

	imageTagOverrides = append(imageTagOverrides, kubeRbacProxyImageTagOverride)

	manifestList := []string{
		"infrastructure-components.yaml",
		"cluster-template.yaml",
		"metadata.yaml",
	}

	for _, manifest := range manifestList {
		var sourceS3Prefix string
		var releaseS3Path string
		latestPath := r.getLatestUploadDestination()

		if r.DevRelease || r.ReleaseEnvironment == "development" {
			sourceS3Prefix = fmt.Sprintf("%s/%s/manifests/infrastructure-aws/%s", capaProjectPath, latestPath, gitTag)
		} else {
			sourceS3Prefix = fmt.Sprintf("releases/bundles/%d/artifacts/cluster-api-provider-aws/manifests/infrastructure-aws/%s", r.BundleNumber, gitTag)
		}

		if r.DevRelease {
			releaseS3Path = fmt.Sprintf("artifacts/%s/cluster-api-provider-aws/manifests/infrastructure-aws/%s", r.DevReleaseUriVersion, gitTag)
		} else {
			releaseS3Path = fmt.Sprintf("releases/bundles/%d/artifacts/cluster-api-provider-aws/manifests/infrastructure-aws/%s", r.BundleNumber, gitTag)
		}

		cdnURI, err := r.GetURI(filepath.Join(
			releaseS3Path,
			manifest))
		if err != nil {
			return nil, errors.Cause(err)
		}

		manifestArtifact := &ManifestArtifact{
			SourceS3Key:       manifest,
			SourceS3Prefix:    sourceS3Prefix,
			ArtifactPath:      filepath.Join(r.ArtifactDir, "capa-manifests", r.BuildRepoHead),
			ReleaseName:       manifest,
			ReleaseS3Path:     releaseS3Path,
			ReleaseCdnURI:     cdnURI,
			ImageTagOverrides: imageTagOverrides,
			GitTag:            gitTag,
			ProjectPath:       capaProjectPath,
		}
		artifacts = append(artifacts, Artifact{Manifest: manifestArtifact})
	}

	return artifacts, nil
}

func (r *ReleaseConfig) GetAwsBundle(imageDigests map[string]string) (anywherev1alpha1.AwsBundle, error) {
	awsBundleArtifacts := map[string][]Artifact{
		"cluster-api-provider-aws": r.BundleArtifactsTable["cluster-api-provider-aws"],
		"kube-rbac-proxy":          r.BundleArtifactsTable["kube-rbac-proxy"],
	}

	version, err := BuildComponentVersion(
		newVersionerWithGITTAG(filepath.Join(r.BuildRepoSource, capaProjectPath)),
	)
	if err != nil {
		return anywherev1alpha1.AwsBundle{}, errors.Wrapf(err, "Error getting version for cluster-api-provider-aws")
	}
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}

	for _, artifacts := range awsBundleArtifacts {
		for _, artifact := range artifacts {
			if artifact.Image != nil {
				imageArtifact := artifact.Image

				bundleImageArtifact := anywherev1alpha1.Image{
					Name:        imageArtifact.AssetName,
					Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
					OS:          imageArtifact.OS,
					Arch:        imageArtifact.Arch,
					URI:         imageArtifact.ReleaseImageURI,
					ImageDigest: imageDigests[imageArtifact.ReleaseImageURI],
				}
				bundleImageArtifacts[imageArtifact.AssetName] = bundleImageArtifact
			}

			if artifact.Manifest != nil {
				manifestArtifact := artifact.Manifest
				bundleManifestArtifact := anywherev1alpha1.Manifest{
					URI: manifestArtifact.ReleaseCdnURI,
				}

				bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact
			}
		}
	}

	bundle := anywherev1alpha1.AwsBundle{
		Version:         version,
		Controller:      bundleImageArtifacts["cluster-api-aws-controller"],
		KubeProxy:       bundleImageArtifacts["kube-rbac-proxy"],
		Components:      bundleManifestArtifacts["infrastructure-components.yaml"],
		ClusterTemplate: bundleManifestArtifacts["cluster-template.yaml"],
		Metadata:        bundleManifestArtifacts["metadata.yaml"],
	}

	return bundle, nil
}
