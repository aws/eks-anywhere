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

func (r *ReleaseConfig) GetVsphereBundle(eksDReleaseChannel string, imageDigests map[string]string) (anywherev1alpha1.VSphereBundle, error) {
	vsphereBundleArtifactsFuncs := map[string]func() ([]Artifact, error){
		"cluster-api-provider-vsphere": r.GetCapvAssets,
		"kube-proxy":                   r.GetKubeRbacProxyAssets,
		"kube-vip":                     r.GetKubeVipAssets,
		"vsphere-csi-driver":           r.GetVsphereCsiAssets,
	}

	version, err := BuildComponentVersion(
		newVersionerWithGITTAG(filepath.Join(r.BuildRepoSource, "projects/kubernetes-sigs/cluster-api-provider-vsphere")),
	)
	if err != nil {
		return anywherev1alpha1.VSphereBundle{}, errors.Wrapf(err, "Error getting version for cluster-api-provider-sphere")
	}
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	for componentName, artifactFunc := range vsphereBundleArtifactsFuncs {
		artifacts, err := artifactFunc()
		if err != nil {
			return anywherev1alpha1.VSphereBundle{}, errors.Wrapf(err, "Error getting artifact information for %s", componentName)
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

	vSphereCloudProviderArtifacts, err := r.GetVsphereCloudProviderAssets(eksDReleaseChannel)
	if err != nil {
		return anywherev1alpha1.VSphereBundle{}, errors.Wrapf(err, "Error getting artifact information for cloud-provider-vsphere for channel %s", eksDReleaseChannel)
	}

	for _, artifact := range vSphereCloudProviderArtifacts {
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

	bundle := anywherev1alpha1.VSphereBundle{
		Version:              version,
		ClusterAPIController: bundleImageArtifacts["cluster-api-vsphere-controller"],
		KubeProxy:            bundleImageArtifacts["kube-rbac-proxy"],
		Manager:              bundleImageArtifacts["cloud-provider-vsphere"],
		KubeVip:              bundleImageArtifacts["kube-vip"],
		Driver:               bundleImageArtifacts["vsphere-csi-driver"],
		Syncer:               bundleImageArtifacts["vsphere-csi-syncer"],
		Components:           bundleManifestArtifacts["infrastructure-components.yaml"],
		ClusterTemplate:      bundleManifestArtifacts["cluster-template.yaml"],
		Metadata:             bundleManifestArtifacts["metadata.yaml"],
	}

	return bundle, nil
}
