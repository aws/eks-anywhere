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

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// GetEksAToolsAssets returns the eks-a artifacts for eks-a-tools image
func (r *ReleaseConfig) GetEksAToolsAssets() ([]Artifact, error) {
	projectSource := "projects/aws/eks-anywhere-build-tooling"
	tagFile := filepath.Join(r.BuildRepoSource, projectSource, "GIT_TAG")
	gitTag, err := readFile(tagFile)
	if err != nil {
		return nil, errors.Cause(err)
	}
	name := "eks-anywhere-cli-tools"

	var sourceRepoName string
	var releaseRepoName string
	if r.DevRelease || r.ReleaseEnvironment == "development" {
		sourceRepoName = "eks-anywhere-cli-tools"
	} else {
		sourceRepoName = "cli-tools"
	}

	if r.DevRelease {
		releaseRepoName = "eks-anywhere-cli-tools"
	} else {
		releaseRepoName = "cli-tools"
	}

	tagOptions := map[string]string{
		"gitTag": gitTag,
	}
	imageArtifact := &ImageArtifact{
		AssetName:       name,
		SourceImageURI:  r.GetSourceImageURI(name, sourceRepoName, tagOptions),
		ReleaseImageURI: r.GetReleaseImageURI(name, releaseRepoName, tagOptions),
		Arch:            []string{"amd64"},
		OS:              "linux",
	}

	artifact := Artifact{Image: imageArtifact}

	return []Artifact{artifact}, nil
}

func (r *ReleaseConfig) GetEksaBundle(imageDigests map[string]string) (anywherev1alpha1.EksaBundle, error) {
	eksABundleArtifactsFuncs := map[string]func() ([]Artifact, error){
		"eks-a-tools":           r.GetEksAToolsAssets,
		"cluster-controller":    r.GetClusterControllerAssets,
		"diagnostic-collector:": r.GetDiagnosticCollectorAssets,
	}
	components := SortArtifactsFuncMap(eksABundleArtifactsFuncs)

	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	bundleObjects := []string{}

	for _, componentName := range components {
		artifactFunc := eksABundleArtifactsFuncs[componentName]
		artifacts, err := artifactFunc()
		if err != nil {
			return anywherev1alpha1.EksaBundle{}, errors.Wrapf(err, "Error getting artifact information for %s", componentName)
		}

		for _, artifact := range artifacts {
			if artifact.Image != nil {
				imageArtifact := artifact.Image

				bundleImageArtifact := anywherev1alpha1.Image{
					Name:        imageArtifact.AssetName,
					Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
					OS:          imageArtifact.OS,
					Arch:        imageArtifact.Arch,
					URI:         imageArtifact.ReleaseImageURI,
					ImageDigest: imageDigests[imageArtifact.ReleaseImageURI],
				}
				bundleImageArtifacts[imageArtifact.AssetName] = bundleImageArtifact
				bundleObjects = append(bundleObjects, bundleImageArtifact.ImageDigest)
			}

			if artifact.Manifest != nil {
				manifestArtifact := artifact.Manifest
				bundleManifestArtifact := anywherev1alpha1.Manifest{
					URI: manifestArtifact.ReleaseCdnURI,
				}

				bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact
			}
		}
	}

	componentChecksum := GenerateComponentChecksum(bundleObjects)
	version, err := BuildComponentVersion(newCliVersioner(r.ReleaseVersion, r.CliRepoSource), componentChecksum)
	if err != nil {
		return anywherev1alpha1.EksaBundle{}, errors.Wrapf(err, "failed generating version for eksa bundle")
	}

	bundle := anywherev1alpha1.EksaBundle{
		Version:             version,
		CliTools:            bundleImageArtifacts["eks-anywhere-cli-tools"],
		Components:          bundleManifestArtifacts["eksa-components.yaml"],
		ClusterController:   bundleImageArtifacts["eks-anywhere-cluster-controller"],
		DiagnosticCollector: bundleImageArtifacts["eks-anywhere-diagnostic-collector"],
	}

	return bundle, nil
}
