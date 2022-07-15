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

func GetBottlerocketAdminBundle(r *releasetypes.ReleaseConfig) (anywherev1alpha1.BottlerocketAdminBundle, error) {
	bottleAdminContainerRegistry := "public.ecr.aws/bottlerocket"
	tag, imageDigest, err := filereader.GetBottlerocketAdminContainerMetadata(r)
	if err != nil {
		return anywherev1alpha1.BottlerocketAdminBundle{}, errors.Cause(err)
	}

	name := "bottlerocket-admin"
	bundleImageArtifact := anywherev1alpha1.Image{
		Name:        name,
		Description: fmt.Sprintf("Container image for %s image", name),
		OS:          "linux",
		Arch:        []string{"amd64"},
		URI:         fmt.Sprintf("%s/%s:%s", bottleAdminContainerRegistry, name, tag),
		ImageDigest: imageDigest,
	}

	bundle := anywherev1alpha1.BottlerocketAdminBundle{
		Admin: bundleImageArtifact,
	}

	return bundle, nil
}
