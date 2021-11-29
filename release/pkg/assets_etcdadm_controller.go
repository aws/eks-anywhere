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

// GetEtcdadmControllerAssets returns the eks-a artifacts for etcdadm controller
func (r *ReleaseConfig) GetEtcdadmControllerAssets() ([]Artifact, error) {
	gitTag, err := r.getEtcdadmControllerGitTag()
	if err != nil {
		return nil, errors.Cause(err)
	}

	name := "etcdadm-controller"
	repoName := fmt.Sprintf("mrajashree/%s", name)
	tagOptions := map[string]string{
		"gitTag": gitTag,
	}
	artifacts := []Artifact{}

	imageArtifact := &ImageArtifact{
		AssetName:       name,
		SourceImageURI:  r.GetSourceImageURI(name, repoName, tagOptions),
		ReleaseImageURI: r.GetReleaseImageURI(name, repoName, tagOptions),
		Arch:            []string{"amd64"},
		OS:              "linux",
		GitTag:          gitTag,
	}

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

	artifacts = append(artifacts, Artifact{Image: imageArtifact})

	manifestList := []string{
		"bootstrap-components.yaml",
		"metadata.yaml",
	}

	for _, manifest := range manifestList {
		var sourceS3Prefix string
		var releaseS3Path string
		latestPath := r.getLatestUploadDestination()

		if r.DevRelease || r.ReleaseEnvironment == "development" {
			sourceS3Prefix = fmt.Sprintf("projects/mrajashree/etcdadm-controller/%s/manifests/bootstrap-etcdadm-controller/%s", latestPath, gitTag)
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

		sourceS3URI, err := r.GetSourceManifestURI(filepath.Join(r.SourceBucket, sourceS3Prefix, manifest))
		if err != nil {
			return nil, errors.Cause(err)
		}

		manifestArtifact := &ManifestArtifact{
			SourceS3Key:       manifest,
			SourceS3Prefix:    sourceS3Prefix,
			SourceS3URI:       sourceS3URI,
			ArtifactPath:      filepath.Join(r.ArtifactDir, "etcdadm-controller-manifests", r.BuildRepoHead),
			ReleaseName:       manifest,
			ReleaseS3Path:     releaseS3Path,
			ReleaseCdnURI:     cdnURI,
			ImageTagOverrides: imageTagOverrides,
		}
		artifacts = append(artifacts, Artifact{Manifest: manifestArtifact})
	}

	return artifacts, nil
}

func (r *ReleaseConfig) GetEtcdadmControllerBundle(imageDigests map[string]string) (anywherev1alpha1.EtcdadmControllerBundle, error) {
	etcdadmControllerBundleArtifactsFuncs := map[string]func() ([]Artifact, error){
		"etcdadm-controller": r.GetEtcdadmControllerAssets,
		"kube-proxy":         r.GetKubeRbacProxyAssets,
	}
	components := SortArtifactsFuncMap(etcdadmControllerBundleArtifactsFuncs)

	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	bundleObjects := []string{}

	for _, componentName := range components {
		artifactFunc := etcdadmControllerBundleArtifactsFuncs[componentName]
		artifacts, err := artifactFunc()
		if err != nil {
			return anywherev1alpha1.EtcdadmControllerBundle{}, errors.Wrapf(err, "Error getting artifact information for %s", componentName)
		}

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
				bundleObjects = append(bundleObjects, bundleImageArtifact.ImageDigest)
			}

			if artifact.Manifest != nil {
				manifestArtifact := artifact.Manifest
				bundleManifestArtifact := anywherev1alpha1.Manifest{
					URI: manifestArtifact.ReleaseCdnURI,
				}

				bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact

				manifestContents, err := ReadHttpFile(manifestArtifact.SourceS3URI)
				if err != nil {
					return anywherev1alpha1.EtcdadmControllerBundle{}, err
				}
				bundleObjects = append(bundleObjects, string(manifestContents[:]))
			}
		}
	}

	componentChecksum := GenerateComponentChecksum(bundleObjects)
	version, err := BuildComponentVersion(
		newVersionerWithGITTAG(filepath.Join(r.BuildRepoSource, "projects/mrajashree/etcdadm-controller")),
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

func (r *ReleaseConfig) getEtcdadmControllerGitTag() (string, error) {
	projectSource := "projects/mrajashree/etcdadm-controller"
	tagFile := filepath.Join(r.BuildRepoSource, projectSource, "GIT_TAG")
	gitTag, err := readFile(tagFile)
	if err != nil {
		return "", errors.Cause(err)
	}

	return gitTag, nil
}
