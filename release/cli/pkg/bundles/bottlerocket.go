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

package bundles

import (
	"fmt"

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/cli/pkg/filereader"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
)

func GetBottlerocketHostContainersBundle(r *releasetypes.ReleaseConfig, eksDReleaseChannel string, imageDigests releasetypes.ImageDigestsTable) (anywherev1alpha1.BottlerocketHostContainersBundle, error) {
	adminArtifact, err := bottlerocketDefaultArtifact(r, "BOTTLEROCKET_ADMIN_CONTAINER_METADATA", "bottlerocket-admin")
	if err != nil {
		return anywherev1alpha1.BottlerocketHostContainersBundle{}, errors.Cause(err)
	}

	controlArtifact, err := bottlerocketDefaultArtifact(r, "BOTTLEROCKET_CONTROL_CONTAINER_METADATA", "bottlerocket-control")
	if err != nil {
		return anywherev1alpha1.BottlerocketHostContainersBundle{}, errors.Cause(err)
	}

	kubeadmBootstrapArtifact, err := bottlerocketKubeadmBootstrapArtifact(r, eksDReleaseChannel, imageDigests)
	if err != nil {
		return anywherev1alpha1.BottlerocketHostContainersBundle{}, errors.Cause(err)
	}

	bundle := anywherev1alpha1.BottlerocketHostContainersBundle{
		Admin:            adminArtifact,
		Control:          controlArtifact,
		KubeadmBootstrap: kubeadmBootstrapArtifact,
	}

	return bundle, nil
}

func bottlerocketDefaultArtifact(r *releasetypes.ReleaseConfig, metadataFile, imageName string) (anywherev1alpha1.Image, error) {
	bottlerocketContainerRegistry := "public.ecr.aws/bottlerocket"
	tag, imageDigest, err := filereader.GetBottlerocketContainerMetadata(r, metadataFile)
	if err != nil {
		return anywherev1alpha1.Image{}, errors.Cause(err)
	}

	return anywherev1alpha1.Image{
		Name:        imageName,
		Description: fmt.Sprintf("Container image for %s image", imageName),
		OS:          "linux",
		Arch:        []string{"amd64"},
		URI:         fmt.Sprintf("%s/%s:%s", bottlerocketContainerRegistry, imageName, tag),
		ImageDigest: imageDigest,
	}, nil
}

// getBottlerocketBootstrapArtifact is a shared helper function that retrieves a specific
// bottlerocket bootstrap artifact by asset name from the bundle artifacts table.
func getBottlerocketBootstrapArtifact(r *releasetypes.ReleaseConfig, eksDReleaseChannel string, imageDigests releasetypes.ImageDigestsTable, assetName string) (anywherev1alpha1.Image, error) {
	bottlerocketBootstrapArtifacts, err := r.BundleArtifactsTable.Load(fmt.Sprintf("bottlerocket-bootstrap-%s", eksDReleaseChannel))
	if err != nil {
		return anywherev1alpha1.Image{}, fmt.Errorf("artifacts for project bottlerocket-bootstrap-%s not found in bundle artifacts table", eksDReleaseChannel)
	}

	for _, artifact := range bottlerocketBootstrapArtifacts {
		imageArtifact := artifact.Image
		if imageArtifact.AssetName == assetName {
			imageDigest, err := imageDigests.Load(imageArtifact.ReleaseImageURI)
			if err != nil {
				return anywherev1alpha1.Image{}, fmt.Errorf("loading digest from image digests table: %v", err)
			}

			return anywherev1alpha1.Image{
				Name:        imageArtifact.AssetName,
				Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
				OS:          imageArtifact.OS,
				Arch:        imageArtifact.Arch,
				URI:         imageArtifact.ReleaseImageURI,
				ImageDigest: imageDigest,
			}, nil
		}
	}

	return anywherev1alpha1.Image{}, fmt.Errorf("%s artifact not found", assetName)
}

func bottlerocketKubeadmBootstrapArtifact(r *releasetypes.ReleaseConfig, eksDReleaseChannel string, imageDigests releasetypes.ImageDigestsTable) (anywherev1alpha1.Image, error) {
	return getBottlerocketBootstrapArtifact(r, eksDReleaseChannel, imageDigests, "bottlerocket-bootstrap")
}

func GetBottlerocketBootstrapContainersBundle(r *releasetypes.ReleaseConfig, eksDReleaseChannel string, imageDigests releasetypes.ImageDigestsTable) (anywherev1alpha1.BottlerocketBootstrapContainersBundle, error) {
	bundle := anywherev1alpha1.BottlerocketBootstrapContainersBundle{}

	// VSphere multi-network bootstrap container (optional)
	if vsphereMultiNetworkArtifact, err := bottlerocketVsphereMultiNetworkArtifact(r, eksDReleaseChannel, imageDigests); err == nil {
		bundle.VsphereMultiNetworkBootstrap = vsphereMultiNetworkArtifact
	}
	// Note: We don't return an error if the artifact is not found since bootstrap containers are optional
	// and may not be available for all release channels or configurations.

	return bundle, nil
}

func bottlerocketVsphereMultiNetworkArtifact(r *releasetypes.ReleaseConfig, eksDReleaseChannel string, imageDigests releasetypes.ImageDigestsTable) (anywherev1alpha1.Image, error) {
	return getBottlerocketBootstrapArtifact(r, eksDReleaseChannel, imageDigests, "bottlerocket-bootstrap-vsphere-multi-network")
}
