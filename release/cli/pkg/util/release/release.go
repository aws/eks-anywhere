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
	"strings"

	"sigs.k8s.io/yaml"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/cli/pkg/aws/s3"
	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	"github.com/aws/eks-anywhere/release/cli/pkg/filereader"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
	artifactutils "github.com/aws/eks-anywhere/release/cli/pkg/util/artifacts"
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
	eksAReleaseManifestKey := artifactutils.GetManifestFilepaths(r.DevRelease, r.Weekly, r.BundleNumber, constants.ReleaseKind, r.BuildRepoBranchName, r.ReleaseDate)
	eksAReleaseManifestUrl := fmt.Sprintf("%s/%s", r.CDN, eksAReleaseManifestKey)

	if s3.KeyExists(r.ReleaseBucket, eksAReleaseManifestKey) {
		contents, err := filereader.ReadHttpFile(eksAReleaseManifestUrl)
		if err != nil {
			return nil, fmt.Errorf("Error reading releases manifest from S3: %v", err)
		}

		if err = yaml.Unmarshal(contents, release); err != nil {
			return nil, fmt.Errorf("Error unmarshaling releases manifest from [%s]: %v", eksAReleaseManifestUrl, err)
		}

		return release, nil
	}

	return emptyRelease, nil
}

func (releases EksAReleases) AppendOrUpdateRelease(r anywherev1alpha1.EksARelease) EksAReleases {
	currentReleaseSemver := strings.Split(r.Version, "+")[0]
	for i, release := range releases {
		existingReleaseSemver := strings.Split(release.Version, "+")[0]
		if currentReleaseSemver == existingReleaseSemver {
			releases[i] = r
			fmt.Println("Updating existing release in releases manifest")
			return releases
		}
	}
	releases = append(releases, r)
	fmt.Println("Adding new release to releases manifest")
	return releases
}
