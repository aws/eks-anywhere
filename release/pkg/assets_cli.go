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
	"strings"

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/pkg/utils"
)

// GetCliArtifacts returns the artifacts for eksctl-anywhere cli
func (r *ReleaseConfig) GetEksACliArtifacts() ([]Artifact, error) {
	osList := []string{"linux", "darwin"}
	arch := "amd64"

	artifacts := []Artifact{}
	for _, os := range osList {
		releaseName := fmt.Sprintf("eksctl-anywhere-%s-%s-%s.tar.gz", r.ReleaseVersion, os, arch)
		releaseName = strings.ReplaceAll(releaseName, "+", "-")

		var sourceS3Key string
		var sourceS3Prefix string
		var releaseS3Path string
		sourcedFromBranch := r.CliRepoBranchName
		latestPath := getLatestUploadDestination(sourcedFromBranch)
		if r.DevRelease || r.ReleaseEnvironment == "development" {
			sourceS3Key = fmt.Sprintf("eksctl-anywhere-%s-%s.tar.gz", os, arch)
			sourceS3Prefix = fmt.Sprintf("eks-a-cli/%s/%s/%s", latestPath, os, arch)
		} else {
			sourceS3Key = fmt.Sprintf("eksctl-anywhere-%s-%s-%s.tar.gz", r.ReleaseVersion, os, arch)
			sourceS3Prefix = fmt.Sprintf("releases/eks-a/%d/artifacts/eks-a/%s/%s/%s", r.ReleaseNumber, r.ReleaseVersion, os, arch)
		}

		if r.DevRelease {
			releaseS3Path = fmt.Sprintf("eks-anywhere/%s/eks-a-cli/%s/%s", r.DevReleaseUriVersion, os, arch)
		} else {
			releaseS3Path = fmt.Sprintf("releases/eks-a/%d/artifacts/eks-a/%s/%s/%s", r.ReleaseNumber, r.ReleaseVersion, os, arch)
		}

		cdnURI, err := r.GetURI(filepath.Join(releaseS3Path, releaseName))
		if err != nil {
			return nil, errors.Cause(err)
		}

		archiveArtifact := &ArchiveArtifact{
			SourceS3Key:    sourceS3Key,
			SourceS3Prefix: sourceS3Prefix,
			ArtifactPath:   filepath.Join(r.ArtifactDir, "eks-a", r.CliRepoHead),
			ReleaseName:    releaseName,
			ReleaseS3Path:  releaseS3Path,
			ReleaseCdnURI:  cdnURI,
			OS:             os,
			Arch:           []string{arch},
		}

		artifacts = append(artifacts, Artifact{Archive: archiveArtifact})
	}
	return artifacts, nil
}

func (r *ReleaseConfig) GetEksARelease() (anywherev1alpha1.EksARelease, error) {
	fmt.Println("\n==========================================================")
	fmt.Println("               EKS-A Release Spec Generation")
	fmt.Println("==========================================================")

	artifacts := r.EksAArtifactsTable["eks-a-cli"]

	bundleManifestFilePath := utils.GetManifestFilepaths(r.DevRelease, r.BundleNumber, anywherev1alpha1.BundlesKind, r.BuildRepoBranchName)
	bundleManifestUrl, err := r.GetURI(bundleManifestFilePath)
	if err != nil {
		return anywherev1alpha1.EksARelease{}, errors.Cause(err)
	}
	bundleArchiveArtifacts := map[string]anywherev1alpha1.Archive{}

	for _, artifact := range artifacts {
		archiveArtifact := artifact.Archive

		tarfile := filepath.Join(archiveArtifact.ArtifactPath, archiveArtifact.ReleaseName)
		sha256, sha512, err := r.readShaSums(tarfile)
		if err != nil {
			return anywherev1alpha1.EksARelease{}, errors.Cause(err)
		}

		bundleArchiveArtifact := anywherev1alpha1.Archive{
			Name:        fmt.Sprintf("eksctl-anywhere-%s", archiveArtifact.OS),
			Description: fmt.Sprintf("EKS Anywhere %s CLI", strings.Title(archiveArtifact.OS)),
			OS:          archiveArtifact.OS,
			Arch:        archiveArtifact.Arch,
			URI:         archiveArtifact.ReleaseCdnURI,
			SHA256:      sha256,
			SHA512:      sha512,
		}

		bundleArchiveArtifacts[fmt.Sprintf("eksctl-anywhere-%s", archiveArtifact.OS)] = bundleArchiveArtifact
	}

	eksARelease := anywherev1alpha1.EksARelease{
		Date:      r.ReleaseDate.String(),
		Version:   r.ReleaseVersion,
		Number:    r.ReleaseNumber,
		GitCommit: r.CliRepoHead,
		GitTag:    r.ReleaseVersion,
		EksABinary: anywherev1alpha1.BinaryBundle{
			LinuxBinary:  bundleArchiveArtifacts["eksctl-anywhere-linux"],
			DarwinBinary: bundleArchiveArtifacts["eksctl-anywhere-darwin"],
		},
		BundleManifestUrl: bundleManifestUrl,
	}

	fmt.Printf("%s Successfully generated EKS-A release spec\n", SuccessIcon)

	return eksARelease, nil
}
