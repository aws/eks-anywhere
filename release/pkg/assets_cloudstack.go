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

// TODO: needs to confirm vgg@ which one to use manager or cloudApiController
func (r *ReleaseConfig) GetCloudStackBundle(eksDReleaseChannel string, imageDigests map[string]string) (anywherev1alpha1.CloudStackBundle, error) {
	cloudstackBundleArtifactsFuncs := map[string]func() ([]Artifact, error){
		"cluster-api-provider-cloudstack": r.GetCapcAssets,
		"kube-proxy":                      r.GetKubeRbacProxyAssets,
	}

	version, err := r.GenerateComponentBundleVersion(
		newVersionerWithGITTAG(filepath.Join(r.BuildRepoSource, "projects/kubernetes-sigs/cluster-api-provider-cloudstack")),
	)
	if err != nil {
		return anywherev1alpha1.CloudStackBundle{}, errors.Wrapf(err, "Error getting version for cluster-api-provider-cloudstack")
	}
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	for componentName, artifactFunc := range cloudstackBundleArtifactsFuncs {
		artifacts, err := artifactFunc()
		if err != nil {
			return anywherev1alpha1.CloudStackBundle{}, errors.Wrapf(err, "Error getting artifact information for %s", componentName)
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

	cloudStackCloudProviderArtifacts, err := r.GetCloudStackCloudProviderAssets(eksDReleaseChannel)
	if err != nil {
		return anywherev1alpha1.CloudStackBundle{}, errors.Wrapf(err, "Error getting artifact information for cloud-provider-cloudstack for channel %s", eksDReleaseChannel)
	}

	for _, artifact := range cloudStackCloudProviderArtifacts {
		imageArtifact := artifact.Image

		bundleArtifact := anywherev1alpha1.Image{
			Name:        imageArtifact.AssetName,
			Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
			OS:          imageArtifact.OS,
			Arch:        imageArtifact.Arch,
			URI:         imageArtifact.ReleaseImageURI,
			ImageDigest: imageDigests[imageArtifact.ReleaseImageURI],
		}
		bundleImageArtifacts[imageArtifact.AssetName] = bundleArtifact
	}

	bundle := anywherev1alpha1.CloudStackBundle{
		Version: version,
		// ClusterAPIController: bundleImageArtifacts["cluster-api-cloudstack-controller"],
		KubeProxy:       bundleImageArtifacts["kube-rbac-proxy"],
		Manager:         bundleImageArtifacts["cloud-provider-cloudstack"],
		Components:      bundleManifestArtifacts["infrastructure-components.yaml"],
		ClusterTemplate: bundleManifestArtifacts["cluster-template.yaml"],
		Metadata:        bundleManifestArtifacts["metadata.yaml"],
	}

	return bundle, nil
}
