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
)

const etcdadmProjectPath = "projects/kubernetes-sigs/etcdadm"

// GetEtcdadmAssets returns the eks-a artifacts for etcdadm
func (r *ReleaseConfig) GetEtcdadmAssets() ([]Artifact, error) {
	os := "linux"
	arch := "amd64"
	gitTag, err := r.readGitTag(etcdadmProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}

	var sourceS3Key string
	var sourceS3Prefix string
	var releaseS3Path string
	var releaseName string
	sourcedFromBranch := r.BuildRepoBranchName
	latestPath := getLatestUploadDestination(sourcedFromBranch)

	if r.DevRelease || r.ReleaseEnvironment == "development" {
		sourceS3Key = fmt.Sprintf("etcdadm-%s-%s-%s.tar.gz", os, arch, gitTag)
		sourceS3Prefix = fmt.Sprintf("%s/%s", etcdadmProjectPath, latestPath)
	} else {
		sourceS3Key = fmt.Sprintf("etcdadm-%s-%s.tar.gz", os, arch)
		sourceS3Prefix = fmt.Sprintf("releases/bundles/%d/artifacts/etcdadm/%s", r.BundleNumber, gitTag)
	}

	if r.DevRelease {
		releaseName = fmt.Sprintf("etcdadm-%s-%s-%s.tar.gz", r.ReleaseVersion, os, arch)
		releaseS3Path = fmt.Sprintf("artifacts/%s/etcdadm/%s", r.DevReleaseUriVersion, gitTag)
	} else {
		releaseName = fmt.Sprintf("etcdadm-%s-%s.tar.gz", os, arch)
		releaseS3Path = fmt.Sprintf("releases/bundles/%d/artifacts/etcdadm/%s", r.BundleNumber, gitTag)
	}

	cdnURI, err := r.GetURI(filepath.Join(releaseS3Path, releaseName))
	if err != nil {
		return nil, errors.Cause(err)
	}

	archiveArtifact := &ArchiveArtifact{
		SourceS3Key:       sourceS3Key,
		SourceS3Prefix:    sourceS3Prefix,
		ArtifactPath:      filepath.Join(r.ArtifactDir, "etcdadm", r.BuildRepoHead),
		ReleaseName:       releaseName,
		ReleaseS3Path:     releaseS3Path,
		ReleaseCdnURI:     cdnURI,
		OS:                os,
		Arch:              []string{arch},
		GitTag:            gitTag,
		ProjectPath:       etcdadmProjectPath,
		SourcedFromBranch: sourcedFromBranch,
	}
	artifacts := []Artifact{{Archive: archiveArtifact}}

	return artifacts, nil
}
