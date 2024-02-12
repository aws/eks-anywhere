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

package release

import (
	"fmt"
	"io"

	"sigs.k8s.io/yaml"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/cli/pkg/aws/s3"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
)

type EksAReleases []anywherev1alpha1.EksARelease

func GetPreviousReleaseIfExists(r *releasetypes.ReleaseConfig) (*anywherev1alpha1.Release, error) {
	emptyRelease := &anywherev1alpha1.Release{
		Spec: anywherev1alpha1.ReleaseSpec{
			Releases: []anywherev1alpha1.EksARelease{},
		},
	}
	if r.DryRun {
		return emptyRelease, nil
	}

	release := &anywherev1alpha1.Release{}
	eksAReleaseManifestKey := r.ReleaseManifestFilepath()

	if !s3.KeyExists(r.ReleaseBucket, eksAReleaseManifestKey) {
		return emptyRelease, nil
	}

	content, err := s3.Read(r.ReleaseBucket, eksAReleaseManifestKey)
	if err != nil {
		return nil, fmt.Errorf("reading releases manifest from S3: %v", err)
	}
	defer content.Close()

	contents, err := io.ReadAll(content)
	if err != nil {
		return nil, fmt.Errorf("reading releases manifest response from S3: %v", err)
	}

	if err = yaml.Unmarshal(contents, release); err != nil {
		return nil, fmt.Errorf("unmarshaling releases manifest from [%s]: %v", eksAReleaseManifestKey, err)
	}

	return release, nil
}

// AppendOrUpdateRelease appends a new release to the manifest if it does not exist, or updates the existing release.
func (releases EksAReleases) AppendOrUpdateRelease(r anywherev1alpha1.EksARelease) EksAReleases {
	for i, release := range releases {
		if r.Version == release.Version {
			releases[i] = r
			fmt.Println("Updating existing release in releases manifest")
			return releases
		}
	}
	releases = append(releases, r)
	fmt.Println("Adding new release to releases manifest")
	return releases
}

// Trim removes the oldest releases if the manifest size exceeds the maxSize.
// If maxSize is -1, no releases are removed.
func Trim(releases EksAReleases, maxSize int) EksAReleases {
	if maxSize == -1 || len(releases) <= maxSize {
		return releases
	}
	return releases[len(releases)-maxSize:]
}
