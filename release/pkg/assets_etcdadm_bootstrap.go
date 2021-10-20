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

// GetEtcdadmBootstrapAssets returns the eks-a artifacts for etcdadm bootstrap provider
func (r *ReleaseConfig) GetEtcdadmBootstrapAssets() ([]Artifact, error) {
	gitTag, err := r.getEtcdadmBootstrapProviderGitTag()
	if err != nil {
		return nil, errors.Cause(err)
	}

	name := "etcdadm-bootstrap-provider"
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

		if r.DevRelease || r.ReleaseEnvironment == "development" {
			sourceS3Prefix = fmt.Sprintf("projects/mrajashree/etcdadm-bootstrap-provider/latest/manifests/bootstrap-etcdadm-bootstrap/%s", gitTag)
		} else {
			sourceS3Prefix = fmt.Sprintf("releases/bundles/%d/artifacts/etcdadm-bootstrap-provider/manifests/bootstrap-etcdadm-bootstrap/%s", r.BundleNumber, gitTag)
		}

		if r.DevRelease {
			releaseS3Path = fmt.Sprintf("artifacts/%s/etcdadm-bootstrap-provider/manifests/bootstrap-etcdadm-bootstrap/%s", r.DevReleaseUriVersion, gitTag)
		} else {
			releaseS3Path = fmt.Sprintf("releases/bundles/%d/artifacts/etcdadm-bootstrap-provider/manifests/bootstrap-etcdadm-bootstrap/%s", r.BundleNumber, gitTag)
		}

		cdnURI, err := r.GetURI(filepath.Join(releaseS3Path, manifest))
		if err != nil {
			return nil, errors.Cause(err)
		}

		manifestArtifact := &ManifestArtifact{
			SourceS3Key:       manifest,
			SourceS3Prefix:    sourceS3Prefix,
			ArtifactPath:      filepath.Join(r.ArtifactDir, "etcdadm-bootstrap-provider-manifests", r.BuildRepoHead),
			ReleaseName:       manifest,
			ReleaseS3Path:     releaseS3Path,
			ReleaseCdnURI:     cdnURI,
			ImageTagOverrides: imageTagOverrides,
		}
		artifacts = append(artifacts, Artifact{Manifest: manifestArtifact})
	}

	return artifacts, nil
}

func (r *ReleaseConfig) GetEtcdadmBootstrapBundle(imageDigests map[string]string) (anywherev1alpha1.EtcdadmBootstrapBundle, error) {
	etcdadmBootstrapBundleArtifactsFuncs := map[string]func() ([]Artifact, error){
		"etcdadm-bootstrap-provider": r.GetEtcdadmBootstrapAssets,
		"kube-proxy":                 r.GetKubeRbacProxyAssets,
	}

	version, err := r.GenerateComponentBundleVersion(
		newVersionerWithGITTAG(filepath.Join(r.BuildRepoSource, "projects/mrajashree/etcdadm-bootstrap-provider")),
	)
	if err != nil {
		return anywherev1alpha1.EtcdadmBootstrapBundle{}, errors.Wrapf(err, "Error getting version for etcdadm-bootstrap-provider")
	}
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}

	for componentName, artifactFunc := range etcdadmBootstrapBundleArtifactsFuncs {
		artifacts, err := artifactFunc()
		if err != nil {
			return anywherev1alpha1.EtcdadmBootstrapBundle{}, errors.Wrapf(err, "Error getting artifact information for %s", componentName)
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

	bundle := anywherev1alpha1.EtcdadmBootstrapBundle{
		Version:    version,
		Controller: bundleImageArtifacts["etcdadm-bootstrap-provider"],
		KubeProxy:  bundleImageArtifacts["kube-rbac-proxy"],
		Components: bundleManifestArtifacts["bootstrap-components.yaml"],
		Metadata:   bundleManifestArtifacts["metadata.yaml"],
	}

	return bundle, nil
}

func (r *ReleaseConfig) getEtcdadmBootstrapProviderGitTag() (string, error) {
	projectSource := "projects/mrajashree/etcdadm-bootstrap-provider"
	tagFile := filepath.Join(r.BuildRepoSource, projectSource, "GIT_TAG")
	gitTag, err := readFile(tagFile)
	if err != nil {
		return "", errors.Cause(err)
	}

	return gitTag, nil
}
