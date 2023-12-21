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

package archives

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/pkg/errors"

	assettypes "github.com/aws/eks-anywhere/release/cli/pkg/assets/types"
	"github.com/aws/eks-anywhere/release/cli/pkg/filereader"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
	artifactutils "github.com/aws/eks-anywhere/release/cli/pkg/util/artifacts"
)

func EksDistroArtifactPathGetter(rc *releasetypes.ReleaseConfig, archive *assettypes.Archive, projectPath, gitTag, eksDReleaseChannel, eksDReleaseNumber, kubeVersion, latestPath, arch string) (string, string, string, string, error) {
	var sourceS3Key string
	var sourceS3Prefix string
	var releaseS3Path string
	var releaseName string
	bottlerocketSupportedK8sVersions, err := filereader.GetBottlerocketSupportedK8sVersionsByFormat(rc, archive.Format)
	if err != nil {
		return "", "", "", "", errors.Cause(err)
	}
	if archive.OSName == "bottlerocket" && !slices.Contains(bottlerocketSupportedK8sVersions, eksDReleaseChannel) {
		return "", "", "", "", nil
	}

	imageExtensions := map[string]string{
		"ami": "gz",
		"ova": "ova",
		"raw": "gz",
	}
	imageExtension := imageExtensions[archive.Format]
	if archive.OSName == "bottlerocket" && (archive.Format == "ami" || archive.Format == "raw") {
		imageExtension = "img.gz"
	}

	if rc.DevRelease || rc.ReleaseEnvironment == "development" {
		sourceS3Key = fmt.Sprintf("%s.%s", archive.OSName, imageExtension)
		sourceS3Prefix = fmt.Sprintf("%s/%s/%s/%s/%s/%s", projectPath, eksDReleaseChannel, archive.Format, archive.OSName, archive.OSVersion, latestPath)
	} else {
		sourceS3Key = fmt.Sprintf("%s-%s-eks-d-%s-%s-eks-a-%d-%s.%s",
			archive.OSName,
			kubeVersion,
			eksDReleaseChannel,
			eksDReleaseNumber,
			rc.BundleNumber,
			arch,
			imageExtension,
		)
		sourceS3Prefix = fmt.Sprintf("releases/bundles/%d/artifacts/%s/%s", rc.BundleNumber, archive.Format, eksDReleaseChannel)
	}

	if rc.DevRelease {
		releaseName = fmt.Sprintf("%s-%s-eks-d-%s-%s-eks-a-%s-%s.%s",
			archive.OSName,
			kubeVersion,
			eksDReleaseChannel,
			eksDReleaseNumber,
			rc.DevReleaseUriVersion,
			arch,
			imageExtension,
		)
		releaseS3Path = fmt.Sprintf("artifacts/%s/eks-distro/%s/%s/%s-%s",
			rc.DevReleaseUriVersion,
			archive.Format,
			eksDReleaseChannel,
			eksDReleaseChannel,
			eksDReleaseNumber,
		)
	} else {
		releaseName = fmt.Sprintf("%s-%s-eks-d-%s-%s-eks-a-%d-%s.%s",
			archive.OSName,
			kubeVersion,
			eksDReleaseChannel,
			eksDReleaseNumber,
			rc.BundleNumber,
			arch,
			imageExtension,
		)
		releaseS3Path = fmt.Sprintf("releases/bundles/%d/artifacts/%s/%s", rc.BundleNumber, archive.Format, eksDReleaseChannel)
	}

	return sourceS3Key, sourceS3Prefix, releaseName, releaseS3Path, nil
}

func TarballArtifactPathGetter(rc *releasetypes.ReleaseConfig, archive *assettypes.Archive, projectPath, gitTag, eksDReleaseChannel, eksDReleaseNumber, kubeVersion, latestPath, arch string) (string, string, string, string, error) {
	os := "linux"
	var sourceS3Key string
	var sourceS3Prefix string
	var releaseS3Path string
	var releaseName string

	if rc.DevRelease || rc.ReleaseEnvironment == "development" {
		sourceS3Key = fmt.Sprintf("%s-%s-%s-%s.tar.gz", archive.Name, os, arch, gitTag)
		sourceS3Prefix = fmt.Sprintf("%s/%s", projectPath, latestPath)
	} else {
		sourceS3Key = fmt.Sprintf("%s-%s-%s.tar.gz", archive.Name, os, arch)
		sourceS3Prefix = fmt.Sprintf("releases/bundles/%d/artifacts/%s/%s", rc.BundleNumber, archive.Name, gitTag)
	}

	if rc.DevRelease {
		releaseName = fmt.Sprintf("%s-%s-%s-%s.tar.gz", archive.Name, rc.DevReleaseUriVersion, os, arch)
		releaseS3Path = fmt.Sprintf("artifacts/%s/%s/%s", rc.DevReleaseUriVersion, archive.Name, gitTag)
	} else {
		releaseName = fmt.Sprintf("%s-%s-%s.tar.gz", archive.Name, os, arch)
		releaseS3Path = fmt.Sprintf("releases/bundles/%d/artifacts/%s/%s", rc.BundleNumber, archive.Name, gitTag)
	}

	return sourceS3Key, sourceS3Prefix, releaseName, releaseS3Path, nil
}

func KernelArtifactPathGetter(rc *releasetypes.ReleaseConfig, archive *assettypes.Archive, projectPath, gitTag, eksDReleaseChannel, eksDReleaseNumber, kubeVersion, latestPath, arch string) (string, string, string, string, error) {
	var sourceS3Prefix string
	var releaseS3Path string
	sourceS3Key, releaseName := archive.Name, archive.Name

	if rc.DevRelease || rc.ReleaseEnvironment == "development" {
		sourceS3Prefix = fmt.Sprintf("%s/%s/%s", projectPath, latestPath, gitTag)
	} else {
		sourceS3Prefix = fmt.Sprintf("releases/bundles/%d/artifacts/hook/%s", rc.BundleNumber, gitTag)
	}

	if rc.DevRelease {
		releaseS3Path = fmt.Sprintf("artifacts/%s/hook/%s", rc.DevReleaseUriVersion, gitTag)
	} else {
		releaseS3Path = fmt.Sprintf("releases/bundles/%d/artifacts/hook/%s", rc.BundleNumber, gitTag)
	}

	return sourceS3Key, sourceS3Prefix, releaseName, releaseS3Path, nil
}

func GetArchiveAssets(rc *releasetypes.ReleaseConfig, archive *assettypes.Archive, projectPath, gitTag, eksDReleaseChannel, eksDReleaseNumber, kubeVersion string) (*releasetypes.ArchiveArtifact, error) {
	os := "linux"
	arch := "amd64"
	if archive.ArchitectureOverride != "" {
		arch = archive.ArchitectureOverride
	}
	sourcedFromBranch := rc.BuildRepoBranchName
	latestPath := artifactutils.GetLatestUploadDestination(sourcedFromBranch)

	sourceS3Key, sourceS3Prefix, releaseName, releaseS3Path, err := getArtifactPathGenerator(archive)(rc, archive, projectPath, gitTag, eksDReleaseChannel, eksDReleaseNumber, kubeVersion, latestPath, arch)
	if err != nil {
		return nil, errors.Cause(err)
	}
	if sourceS3Key == "" && err == nil {
		return nil, nil
	}

	cdnURI, err := artifactutils.GetURI(rc.CDN, filepath.Join(releaseS3Path, releaseName))
	if err != nil {
		return nil, errors.Cause(err)
	}

	archiveArtifact := &releasetypes.ArchiveArtifact{
		SourceS3Key:       sourceS3Key,
		SourceS3Prefix:    sourceS3Prefix,
		ArtifactPath:      filepath.Join(rc.ArtifactDir, fmt.Sprintf("%s-%s", archive.Name, archive.Format), eksDReleaseChannel, rc.BuildRepoHead),
		ReleaseName:       releaseName,
		ReleaseS3Path:     releaseS3Path,
		ReleaseCdnURI:     cdnURI,
		OS:                os,
		OSName:            archive.OSName,
		Arch:              []string{arch},
		GitTag:            gitTag,
		ProjectPath:       projectPath,
		SourcedFromBranch: sourcedFromBranch,
		ImageFormat:       archive.Format,
	}

	return archiveArtifact, nil
}

func getArtifactPathGenerator(archive *assettypes.Archive) assettypes.ArchiveS3PathGenerator {
	if archive.ArchiveS3PathGetter != nil {
		return assettypes.ArchiveS3PathGenerator(archive.ArchiveS3PathGetter)
	}
	return assettypes.ArchiveS3PathGenerator(TarballArtifactPathGetter)
}
