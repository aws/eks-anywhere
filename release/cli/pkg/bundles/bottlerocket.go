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

func bottlerocketKubeadmBootstrapArtifact(r *releasetypes.ReleaseConfig, eksDReleaseChannel string, imageDigests releasetypes.ImageDigestsTable) (anywherev1alpha1.Image, error) {
	bottlerocketBootstrapArtifacts, err := r.BundleArtifactsTable.Load(fmt.Sprintf("bottlerocket-bootstrap-%s", eksDReleaseChannel))
	if err != nil {
		return anywherev1alpha1.Image{}, fmt.Errorf("artifacts for project bottlerocket-bootstrap-%s not found in bundle artifacts table", eksDReleaseChannel)
	}

	bundleArtifacts := map[string]anywherev1alpha1.Image{}

	for _, artifact := range bottlerocketBootstrapArtifacts {
		imageArtifact := artifact.Image
		imageDigest, err := imageDigests.Load(imageArtifact.ReleaseImageURI)
		if err != nil {
			return anywherev1alpha1.Image{}, fmt.Errorf("loading digest from image digests table: %v", err)
		}
		bottlerocketBootstrapImage := anywherev1alpha1.Image{
			Name:        imageArtifact.AssetName,
			Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
			OS:          imageArtifact.OS,
			Arch:        imageArtifact.Arch,
			URI:         imageArtifact.ReleaseImageURI,
			ImageDigest: imageDigest,
		}
		bundleArtifacts[imageArtifact.AssetName] = bottlerocketBootstrapImage
	}

	return bundleArtifacts["bottlerocket-bootstrap"], nil
}
