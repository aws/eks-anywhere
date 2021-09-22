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

// GetClusterControllerAssets returns the artifacts for eks-a cluster controller
func (r *ReleaseConfig) GetClusterControllerAssets() ([]Artifact, error) {
	projectSource := "projects/aws/eks-anywhere"
	tagFile := filepath.Join(r.BuildRepoSource, projectSource, "GIT_TAG")
	gitTag, err := readFile(tagFile)
	if err != nil {
		return nil, errors.Cause(err)
	}

	name := "eks-anywhere-cluster-controller"

	var sourceRepoName string
	var releaseRepoName string
	if r.DevRelease || r.ReleaseEnvironment == "development" {
		sourceRepoName = "eks-anywhere-cluster-controller"
	} else {
		sourceRepoName = "cluster-controller"
	}

	if r.DevRelease {
		releaseRepoName = "eks-anywhere-cluster-controller"
	} else {
		releaseRepoName = "cluster-controller"
	}

	tagOptions := map[string]string{}
	artifacts := []Artifact{}

	imageArtifact := &ImageArtifact{
		AssetName:       name,
		SourceImageURI:  r.GetSourceImageURI(name, sourceRepoName, tagOptions),
		ReleaseImageURI: r.GetReleaseImageURI(name, releaseRepoName, tagOptions),
		Arch:            []string{"amd64"},
		OS:              "linux",
		GitTag:          gitTag,
	}

	var imageTagOverrides []ImageTagOverride

	kubeRbacProxyImageTagOverride, err := r.GetKubeRbacProxyImageTagOverride()
	if err != nil {
		return nil, errors.Cause(err)
	}

	imageTagOverride := ImageTagOverride{
		Repository: sourceRepoName,
		ReleaseUri: imageArtifact.ReleaseImageURI,
	}
	imageTagOverrides = append(imageTagOverrides, imageTagOverride, kubeRbacProxyImageTagOverride)

	artifacts = append(artifacts, Artifact{Image: imageArtifact})

	manifest := "eksa-components.yaml"

	var sourceS3Prefix string
	var releaseS3Path string

	if r.DevRelease || r.ReleaseEnvironment == "development" {
		sourceS3Prefix = "projects/aws/eks-anywhere/latest/manifests/cluster-controller"
	} else {
		sourceS3Prefix = fmt.Sprintf("releases/bundles/%s/artifacts/eks-anywhere-cluster-controller/manifests/%s", r.BundleNumber, gitTag)
	}

	if r.DevRelease {
		releaseS3Path = fmt.Sprintf("artifacts/%s/eks-anywhere/manifests/cluster-controller/%s", r.DevReleaseUriVersion, gitTag)
	} else {
		releaseS3Path = fmt.Sprintf("releases/bundles/%s/artifacts/eks-anywhere-cluster-controller/manifests/%s", r.BundleNumber, gitTag)
	}

	cdnURI, err := r.GetURI(filepath.Join(releaseS3Path, manifest))
	if err != nil {
		return nil, errors.Cause(err)
	}

	manifestArtifact := &ManifestArtifact{
		SourceS3Key:       manifest,
		SourceS3Prefix:    sourceS3Prefix,
		ArtifactPath:      filepath.Join(r.ArtifactDir, "cluster-controller-manifests", r.BuildRepoHead),
		ReleaseName:       manifest,
		ReleaseS3Path:     releaseS3Path,
		ReleaseCdnURI:     cdnURI,
		ImageTagOverrides: imageTagOverrides,
	}
	artifacts = append(artifacts, Artifact{Manifest: manifestArtifact})

	return artifacts, nil
}
