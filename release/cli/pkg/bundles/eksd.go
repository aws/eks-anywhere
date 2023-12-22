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
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	"github.com/aws/eks-anywhere/release/cli/pkg/filereader"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
)

func GetEksDReleaseBundle(r *releasetypes.ReleaseConfig, eksDReleaseChannel, kubeVer, eksDReleaseNumber string, imageDigests map[string]string, dev bool) (anywherev1alpha1.EksDRelease, error) {
	artifacts := r.BundleArtifactsTable[fmt.Sprintf("image-builder-%s", eksDReleaseChannel)]
	artifacts = append(artifacts, r.BundleArtifactsTable[fmt.Sprintf("kind-%s", eksDReleaseChannel)]...)

	tarballArtifacts := map[string][]releasetypes.Artifact{
		"containerd":    r.BundleArtifactsTable["containerd"],
		"cri-tools":     r.BundleArtifactsTable["cri-tools"],
		"etcdadm":       r.BundleArtifactsTable["etcdadm"],
		"image-builder": r.BundleArtifactsTable["image-builder"],
	}

	bundleArchiveArtifacts := map[string]anywherev1alpha1.Archive{}
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}

	eksDManifestUrl := filereader.GetEksDReleaseManifestUrl(eksDReleaseChannel, eksDReleaseNumber, dev)
	for _, artifact := range artifacts {
		if artifact.Archive != nil {
			archiveArtifact := artifact.Archive
			osName := archiveArtifact.OSName
			imageFormat := archiveArtifact.ImageFormat

			tarfile := filepath.Join(archiveArtifact.ArtifactPath, archiveArtifact.ReleaseName)
			sha256, sha512, err := filereader.ReadShaSums(tarfile, r)
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
				sha256, sha512, err := filereader.ReadShaSums(tarfile, r)
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

	eksdRelease, err := filereader.GetEksdRelease(eksDManifestUrl)
	if err != nil {
		return anywherev1alpha1.EksDRelease{}, err
	}

	gitCommit := r.BuildRepoHead
	if r.DryRun {
		gitCommit = constants.FakeGitCommit
	}

	bundle := anywherev1alpha1.EksDRelease{
		Name:           eksdRelease.Name,
		ReleaseChannel: eksDReleaseChannel,
		KubeVersion:    kubeVer,
		EksDReleaseUrl: eksDManifestUrl,
		GitCommit:      gitCommit,
		KindNode:       bundleImageArtifacts["kind-node"],
		Etcdadm:        bundleArchiveArtifacts["etcdadm"],
		Crictl:         bundleArchiveArtifacts["cri-tools"],
		Containerd:     bundleArchiveArtifacts["containerd"],
		ImageBuilder:   bundleArchiveArtifacts["image-builder"],
		Ami: anywherev1alpha1.OSImageBundle{
			Bottlerocket: bundleArchiveArtifacts["bottlerocket-ami"],
		},
		Ova: anywherev1alpha1.OSImageBundle{
			Bottlerocket: bundleArchiveArtifacts["bottlerocket-ova"],
		},
		Raw: anywherev1alpha1.OSImageBundle{
			Bottlerocket: bundleArchiveArtifacts["bottlerocket-raw"],
		},
		Components: constants.EksDReleaseComponentsUrl,
	}

	return bundle, nil
}
