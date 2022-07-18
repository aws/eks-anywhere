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

const (
	imageBuilderProjectPath  = "projects/kubernetes-sigs/image-builder"
	kindProjectPath          = "projects/kubernetes-sigs/kind"
	releasePath              = "release"
	eksDReleaseComponentsUrl = "https://distro.eks.amazonaws.com/crds/releases.distro.eks.amazonaws.com-v1alpha1.yaml"
	fakeGitCommit            = "0123456789abcdef0123456789abcdef01234567"
)

// GetEksDChannelAssets returns the eks-d artifacts including OVAs and kind node image
func (r *ReleaseConfig) GetEksDChannelAssets(eksDReleaseChannel, kubeVer, eksDReleaseNumber string) ([]Artifact, error) {
	// Ova artifacts
	os := "linux"
	arch := "amd64"
	osNames := []string{"ubuntu", "bottlerocket"}
	imageFormats := []string{"ova", "raw"}
	imageExtensions := map[string]string{
		"ova": "ova",
		"raw": "gz",
	}
	artifacts := []Artifact{}
	imageBuilderGitTag, err := r.readGitTag(imageBuilderProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}

	for _, osName := range osNames {
		for _, imageFormat := range imageFormats {
			bottlerocketSupportedK8sVersions, err := getBottlerocketSupportedK8sVersionsByFormat(r, imageFormat)
			if err != nil {
				return nil, errors.Cause(err)
			}
			if osName == "bottlerocket" && !utils.SliceContains(bottlerocketSupportedK8sVersions, eksDReleaseChannel) {
				continue
			}
			var sourceS3Key string
			var sourceS3Prefix string
			var releaseS3Path string
			var releaseName string
			sourcedFromBranch := r.BuildRepoBranchName
			latestPath := getLatestUploadDestination(sourcedFromBranch)
			imageExtension := imageExtensions[imageFormat]
			if osName == "bottlerocket" && imageFormat == "raw" {
				imageExtension = "img.gz"
			}

			if r.DevRelease || r.ReleaseEnvironment == "development" {
				sourceS3Key = fmt.Sprintf("%s.%s", osName, imageExtension)
				sourceS3Prefix = fmt.Sprintf("%s/%s/%s/%s/%s", imageBuilderProjectPath, eksDReleaseChannel, imageFormat, osName, latestPath)
			} else {
				sourceS3Key = fmt.Sprintf("%s-%s-eks-d-%s-%s-eks-a-%d-%s.%s",
					osName,
					kubeVer,
					eksDReleaseChannel,
					eksDReleaseNumber,
					r.BundleNumber,
					arch,
					imageExtension,
				)
				sourceS3Prefix = fmt.Sprintf("releases/bundles/%d/artifacts/%s/%s", r.BundleNumber, imageFormat, eksDReleaseChannel)
			}

			if r.DevRelease {
				releaseName = fmt.Sprintf("%s-%s-eks-d-%s-%s-eks-a-%s-%s.%s",
					osName,
					kubeVer,
					eksDReleaseChannel,
					eksDReleaseNumber,
					r.DevReleaseUriVersion,
					arch,
					imageExtension,
				)
				releaseS3Path = fmt.Sprintf("artifacts/%s/eks-distro/%s/%s/%s-%s",
					r.DevReleaseUriVersion,
					imageFormat,
					eksDReleaseChannel,
					eksDReleaseChannel,
					eksDReleaseNumber,
				)
			} else {
				releaseName = fmt.Sprintf("%s-%s-eks-d-%s-%s-eks-a-%d-%s.%s",
					osName,
					kubeVer,
					eksDReleaseChannel,
					eksDReleaseNumber,
					r.BundleNumber,
					arch,
					imageExtension,
				)
				releaseS3Path = fmt.Sprintf("releases/bundles/%d/artifacts/%s/%s", r.BundleNumber, imageFormat, eksDReleaseChannel)
			}

			cdnURI, err := r.GetURI(filepath.Join(releaseS3Path, releaseName))
			if err != nil {
				return nil, errors.Cause(err)
			}

			archiveArtifact := &ArchiveArtifact{
				SourceS3Key:       sourceS3Key,
				SourceS3Prefix:    sourceS3Prefix,
				ArtifactPath:      filepath.Join(r.ArtifactDir, fmt.Sprintf("eks-d-%s", imageFormat), eksDReleaseChannel, r.BuildRepoHead),
				ReleaseName:       releaseName,
				ReleaseS3Path:     releaseS3Path,
				ReleaseCdnURI:     cdnURI,
				OS:                os,
				OSName:            osName,
				Arch:              []string{arch},
				GitTag:            imageBuilderGitTag,
				ProjectPath:       imageBuilderProjectPath,
				SourcedFromBranch: sourcedFromBranch,
				ImageFormat:       imageFormat,
			}

			artifacts = append(artifacts, Artifact{Archive: archiveArtifact})
		}
	}

	kindGitTag, err := r.readGitTag(kindProjectPath, r.BuildRepoBranchName)
	if err != nil {
		return nil, errors.Cause(err)
	}

	name := "kind-node"
	repoName := "kubernetes-sigs/kind/node"
	tagOptions := map[string]string{
		"eksDReleaseChannel": eksDReleaseChannel,
		"eksDReleaseNumber":  eksDReleaseNumber,
		"kubeVersion":        kubeVer,
		"projectPath":        kindProjectPath,
		"gitTag":             kindGitTag,
	}

	sourceImageUri, sourcedFromBranch, err := r.GetSourceImageURI(name, repoName, tagOptions)
	if err != nil {
		return nil, errors.Cause(err)
	}
	if sourcedFromBranch != r.BuildRepoBranchName {
		kindGitTag, err = r.readGitTag(eksAToolsProjectPath, sourcedFromBranch)
		if err != nil {
			return nil, errors.Cause(err)
		}
		tagOptions["gitTag"] = kindGitTag
	}
	releaseImageUri, err := r.GetReleaseImageURI(name, repoName, tagOptions)
	if err != nil {
		return nil, errors.Cause(err)
	}

	imageArtifact := &ImageArtifact{
		AssetName:         name,
		SourceImageURI:    sourceImageUri,
		ReleaseImageURI:   releaseImageUri,
		Arch:              []string{"amd64"},
		OS:                "linux",
		GitTag:            kindGitTag,
		ProjectPath:       kindProjectPath,
		SourcedFromBranch: sourcedFromBranch,
	}

	artifacts = append(artifacts, Artifact{Image: imageArtifact})

	return artifacts, nil
}

func (r *ReleaseConfig) GetEksDReleaseBundle(eksDReleaseChannel, kubeVer, eksDReleaseNumber string, imageDigests map[string]string, dev bool) (anywherev1alpha1.EksDRelease, error) {
	artifacts := r.BundleArtifactsTable[fmt.Sprintf("eks-d-%s", eksDReleaseChannel)]

	tarballArtifacts := map[string][]Artifact{
		"cri-tools": r.BundleArtifactsTable["cri-tools"],
		"etcdadm":   r.BundleArtifactsTable["etcdadm"],
	}

	bundleArchiveArtifacts := map[string]anywherev1alpha1.Archive{}
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}

	eksDManifestUrl := GetEksDReleaseManifestUrl(eksDReleaseChannel, eksDReleaseNumber, dev)
	for _, artifact := range artifacts {
		if artifact.Archive != nil {
			archiveArtifact := artifact.Archive
			osName := archiveArtifact.OSName
			imageFormat := archiveArtifact.ImageFormat

			tarfile := filepath.Join(archiveArtifact.ArtifactPath, archiveArtifact.ReleaseName)
			sha256, sha512, err := r.readShaSums(tarfile)
			if err != nil {
				return anywherev1alpha1.EksDRelease{}, errors.Cause(err)
			}

			bundleArchiveArtifact := anywherev1alpha1.Archive{
				Name:        archiveArtifact.ReleaseName,
				Description: fmt.Sprintf("%s %s image for EKS-D %s-%s release", strings.Title(archiveArtifact.OSName), strings.Title(archiveArtifact.ImageFormat), eksDReleaseChannel, eksDReleaseNumber),
				OS:          archiveArtifact.OS,
				OSName:      archiveArtifact.OSName,
				Arch:        archiveArtifact.Arch,
				URI:         archiveArtifact.ReleaseCdnURI,
				SHA256:      sha256,
				SHA512:      sha512,
			}

			bundleArchiveArtifacts[fmt.Sprintf("%s-%s", osName, imageFormat)] = bundleArchiveArtifact
		}

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

			bundleImageArtifacts["kind-node"] = bundleImageArtifact
		}
	}

	for componentName, artifacts := range tarballArtifacts {
		for _, artifact := range artifacts {
			if artifact.Archive != nil {
				archiveArtifact := artifact.Archive

				tarfile := filepath.Join(archiveArtifact.ArtifactPath, archiveArtifact.ReleaseName)
				sha256, sha512, err := r.readShaSums(tarfile)
				if err != nil {
					return anywherev1alpha1.EksDRelease{}, errors.Cause(err)
				}

				bundleArchiveArtifact := anywherev1alpha1.Archive{
					Name:        archiveArtifact.ReleaseName,
					Description: fmt.Sprintf("%s tarball for %s/%s", componentName, archiveArtifact.OS, archiveArtifact.Arch[0]),
					OS:          archiveArtifact.OS,
					Arch:        archiveArtifact.Arch,
					URI:         archiveArtifact.ReleaseCdnURI,
					SHA256:      sha256,
					SHA512:      sha512,
				}

				bundleArchiveArtifacts[componentName] = bundleArchiveArtifact
			}
		}
	}

	eksdRelease, err := getEksdRelease(eksDManifestUrl)
	if err != nil {
		return anywherev1alpha1.EksDRelease{}, err
	}

	gitCommit := r.BuildRepoHead
	if r.DryRun {
		gitCommit = fakeGitCommit
	}

	bundle := anywherev1alpha1.EksDRelease{
		Name:           eksdRelease.Name,
		ReleaseChannel: eksDReleaseChannel,
		KubeVersion:    kubeVer,
		EksDReleaseUrl: eksDManifestUrl,
		GitCommit:      gitCommit,
		KindNode:       bundleImageArtifacts["kind-node"],
		Ova: anywherev1alpha1.OSImageBundle{
			Bottlerocket: anywherev1alpha1.OSImage{
				Archive: bundleArchiveArtifacts["bottlerocket-ova"],
			},
			Ubuntu: anywherev1alpha1.OSImage{
				Archive: bundleArchiveArtifacts["ubuntu-ova"],
				Etcdadm: bundleArchiveArtifacts["etcdadm"],
				Crictl:  bundleArchiveArtifacts["cri-tools"],
			},
		},
		Raw: anywherev1alpha1.OSImageBundle{
			Bottlerocket: anywherev1alpha1.OSImage{
				Archive: bundleArchiveArtifacts["bottlerocket-raw"],
			},
			Ubuntu: anywherev1alpha1.OSImage{
				Archive: bundleArchiveArtifacts["ubuntu-raw"],
				Etcdadm: bundleArchiveArtifacts["etcdadm"],
				Crictl:  bundleArchiveArtifacts["cri-tools"],
			},
		},
		Components: eksDReleaseComponentsUrl,
	}

	return bundle, nil
}
