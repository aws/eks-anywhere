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

package manifests

import (
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"

	assettypes "github.com/aws/eks-anywhere/release/cli/pkg/assets/types"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
	artifactutils "github.com/aws/eks-anywhere/release/cli/pkg/util/artifacts"
)

func GetManifestAssets(rc *releasetypes.ReleaseConfig, manifestComponent *assettypes.ManifestComponent, manifestFile, projectName, projectPath, gitTag, sourcedFromBranch string, imageTagOverrides []releasetypes.ImageTagOverride) (*releasetypes.ManifestArtifact, error) {
	componentName := manifestComponent.Name
	var sourceS3Prefix string
	var releaseS3Path string
	latestPath := artifactutils.GetLatestUploadDestination(sourcedFromBranch)

	manifestPrefixFolder := projectName
	if manifestComponent.ReleaseManifestPrefix != "" {
		manifestPrefixFolder = manifestComponent.ReleaseManifestPrefix
	}

	if rc.DevRelease || rc.ReleaseEnvironment == "development" {
		sourceS3Prefix = fmt.Sprintf("%s/%s/manifests/%s", projectPath, latestPath, componentName)
		if !manifestComponent.NoVersionSuffix {
			sourceS3Prefix = fmt.Sprintf("%s/%s", sourceS3Prefix, gitTag)
		}
	} else {
		sourceS3Prefix = fmt.Sprintf("releases/bundles/%d/artifacts/%s/manifests/%s/%s", rc.BundleNumber, manifestPrefixFolder, componentName, gitTag)
	}

	if rc.DevRelease {
		releaseS3Path = fmt.Sprintf("artifacts/%s/%s/manifests/%s/%s", rc.DevReleaseUriVersion, manifestPrefixFolder, componentName, gitTag)
	} else {
		releaseS3Path = fmt.Sprintf("releases/bundles/%d/artifacts/%s/manifests/%s/%s", rc.BundleNumber, manifestPrefixFolder, componentName, gitTag)
	}

	cdnURI, err := artifactutils.GetURI(rc.CDN, filepath.Join(releaseS3Path, manifestFile))
	if err != nil {
		return nil, errors.Cause(err)
	}

	manifestArtifact := &releasetypes.ManifestArtifact{
		SourceS3Key:       manifestFile,
		SourceS3Prefix:    sourceS3Prefix,
		ArtifactPath:      filepath.Join(rc.ArtifactDir, fmt.Sprintf("%s-manifests", componentName), rc.BuildRepoHead),
		ReleaseName:       manifestFile,
		ReleaseS3Path:     releaseS3Path,
		ReleaseCdnURI:     cdnURI,
		ImageTagOverrides: imageTagOverrides,
		GitTag:            gitTag,
		ProjectPath:       projectPath,
		SourcedFromBranch: sourcedFromBranch,
		Component:         componentName,
	}

	return manifestArtifact, nil
}
