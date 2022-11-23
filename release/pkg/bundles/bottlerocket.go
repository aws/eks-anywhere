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
	"github.com/aws/eks-anywhere/release/pkg/filereader"
	releasetypes "github.com/aws/eks-anywhere/release/pkg/types"
)

func GetBottlerocketBundle(r *releasetypes.ReleaseConfig) (anywherev1alpha1.BottlerocketBundle, error) {
	adminArtifact, err := bottlerocketBundleImageArtifact(r, "BOTTLEROCKET_ADMIN_CONTAINER_METADATA", "bottlerocket-admin")
	if err != nil {
		return anywherev1alpha1.BottlerocketBundle{}, errors.Cause(err)
	}

	controlArtifact, err := bottlerocketBundleImageArtifact(r, "BOTTLEROCKET_CONTROL_CONTAINER_METADATA", "bottlerocket-control")
	if err != nil {
		return anywherev1alpha1.BottlerocketBundle{}, errors.Cause(err)
	}

	bundle := anywherev1alpha1.BottlerocketBundle{
		Admin:   *adminArtifact,
		Control: *controlArtifact,
	}

	return bundle, nil
}

func bottlerocketBundleImageArtifact(r *releasetypes.ReleaseConfig, metadataFile, imageName string) (*anywherev1alpha1.Image, error) {
	bottlerocketContainerRegistry := "public.ecr.aws/bottlerocket"
	tag, imageDigest, err := filereader.GetBottlerocketContainerMetadata(r, metadataFile)
	if err != nil {
		return nil, errors.Cause(err)
	}

	return &anywherev1alpha1.Image{
		Name:        imageName,
		Description: fmt.Sprintf("Container image for %s image", imageName),
		OS:          "linux",
		Arch:        []string{"amd64"},
		URI:         fmt.Sprintf("%s/%s:%s", bottlerocketContainerRegistry, imageName, tag),
		ImageDigest: imageDigest,
	}, nil
}
