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
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	utilversion "k8s.io/apimachinery/pkg/util/version"
	"sigs.k8s.io/yaml"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	"github.com/aws/eks-anywhere/release/cli/pkg/filereader"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
	bundleutils "github.com/aws/eks-anywhere/release/cli/pkg/util/bundles"
	"github.com/aws/eks-anywhere/release/cli/pkg/version"
)

func GetTinkerbellBundle(r *releasetypes.ReleaseConfig, imageDigests releasetypes.ImageDigestsTable) (anywherev1alpha1.TinkerbellBundle, error) {
	projectsInBundle := []string{"cluster-api-provider-tinkerbell", "kube-vip", "envoy", "tinkerbell", "tink", "hub", "hook", "stack", "tinkerbell-chart", "tinkerbell-crds"}
	tinkerbellBundleArtifacts := map[string][]releasetypes.Artifact{}
	for _, project := range projectsInBundle {
		projectArtifacts, err := r.BundleArtifactsTable.Load(project)
		if err != nil {
			return anywherev1alpha1.TinkerbellBundle{}, fmt.Errorf("artifacts for project %s not found in bundle artifacts table", project)
		}
		tinkerbellBundleArtifacts[project] = projectArtifacts
	}
	sortedComponentNames := bundleutils.SortArtifactsMap(tinkerbellBundleArtifacts)

	var sourceBranch string
	var captMetadata *releasetypes.ManifestArtifact
	var componentChecksum string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	bundleArchiveArtifacts := map[string]anywherev1alpha1.Archive{}
	artifactHashes := []string{}

	for _, componentName := range sortedComponentNames {
		for _, artifact := range tinkerbellBundleArtifacts[componentName] {
			if artifact.Image != nil {
				imageArtifact := artifact.Image
				bundleImageArtifact := anywherev1alpha1.Image{}
				if componentName == "cluster-api-provider-tinkerbell" {
					sourceBranch = imageArtifact.SourcedFromBranch
				}
				imageDigest, err := imageDigests.Load(imageArtifact.ReleaseImageURI)
				if err != nil {
					return anywherev1alpha1.TinkerbellBundle{}, fmt.Errorf("loading digest from image digests table: %v", err)
				}
				// If the artifact is a helm chart, handle differently. Helm charts can have both suffix "-chart" and "-helm",
				// so we need to handle both cases.
				if strings.HasSuffix(imageArtifact.AssetName, "chart") || strings.HasSuffix(imageArtifact.AssetName, "helm") {
					assetName := strings.TrimSuffix(imageArtifact.AssetName, "-helm")
					bundleImageArtifact = anywherev1alpha1.Image{
						Name:        assetName,
						Description: fmt.Sprintf("Helm chart for %s", assetName),
						URI:         imageArtifact.ReleaseImageURI,
						ImageDigest: imageDigest,
					}
				} else {
					bundleImageArtifact = anywherev1alpha1.Image{
						Name:        imageArtifact.AssetName,
						Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
						OS:          imageArtifact.OS,
						Arch:        imageArtifact.Arch,
						URI:         imageArtifact.ReleaseImageURI,
						ImageDigest: imageDigest,
					}
				}
				bundleImageArtifacts[imageArtifact.AssetName] = bundleImageArtifact
				artifactHashes = append(artifactHashes, bundleImageArtifact.ImageDigest)
			}

			if artifact.Manifest != nil {
				manifestArtifact := artifact.Manifest
				if componentName == "cluster-api-provider-tinkerbell" && manifestArtifact.ReleaseName == "metadata.yaml" {
					captMetadata = manifestArtifact
				}
				bundleManifestArtifact := anywherev1alpha1.Manifest{
					URI: manifestArtifact.ReleaseCdnURI,
				}

				bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact

				manifestHash, err := version.GenerateManifestHash(r, manifestArtifact)
				if err != nil {
					return anywherev1alpha1.TinkerbellBundle{}, err
				}

				artifactHashes = append(artifactHashes, manifestHash)
			}

			if artifact.Archive != nil {
				archiveArtifact := artifact.Archive
				bundleArchiveArtifact := anywherev1alpha1.Archive{
					Name:        archiveArtifact.ReleaseName,
					Description: "Tinkerbell operating system installation environment (osie) component",
					URI:         archiveArtifact.ReleaseCdnURI,
				}
				bundleArchiveArtifacts[archiveArtifact.ReleaseName] = bundleArchiveArtifact
			}
		}
	}

	if r.DryRun {
		componentChecksum = version.FakeComponentChecksum
	} else {
		componentChecksum = version.GenerateComponentHash(artifactHashes, r.DryRun)
	}
	captGitTag, err := filereader.ReadGitTag(constants.CaptProjectPath, r.BuildRepoSource, sourceBranch)
	if err != nil {
		return anywherev1alpha1.TinkerbellBundle{}, errors.Wrap(err, "Error getting version for cluster-api-provider-tinkerbell")
	}
	providerVersion, err := tinkerbellProviderVersion(captGitTag, captMetadata, componentChecksum, r.DryRun)
	if err != nil {
		return anywherev1alpha1.TinkerbellBundle{}, errors.Wrapf(err, "Error getting version for cluster-api-provider-tinkerbell")
	}

	bundle := anywherev1alpha1.TinkerbellBundle{
		Version:              providerVersion,
		ClusterAPIController: bundleImageArtifacts["cluster-api-provider-tinkerbell"],
		KubeVip:              bundleImageArtifacts["kube-vip"],
		Envoy:                bundleImageArtifacts["envoy"],
		Components:           bundleManifestArtifacts["infrastructure-components.yaml"],
		Metadata:             bundleManifestArtifacts["metadata.yaml"],
		ClusterTemplate:      bundleManifestArtifacts["cluster-template.yaml"],
		TinkerbellStack: anywherev1alpha1.TinkerbellStackBundle{
			Actions: anywherev1alpha1.ActionsBundle{
				Cexec:       bundleImageArtifacts["cexec"],
				Kexec:       bundleImageArtifacts["kexec"],
				ImageToDisk: bundleImageArtifacts["image2disk"],
				OciToDisk:   bundleImageArtifacts["oci2disk"],
				Reboot:      bundleImageArtifacts["reboot"],
				WriteFile:   bundleImageArtifacts["writefile"],
			},
			Boots: bundleImageArtifacts["tinkerbell"],
			Hook: anywherev1alpha1.HookBundle{
				Bootkit: bundleImageArtifacts["hook-bootkit"],
				Docker:  bundleImageArtifacts["hook-docker"],
				Kernel:  bundleImageArtifacts["hook-kernel"],
				Initramfs: anywherev1alpha1.HookArch{
					Arm: bundleArchiveArtifacts["initramfs-aarch64"],
					Amd: bundleArchiveArtifacts["initramfs-x86_64"],
				},
				ISO: anywherev1alpha1.HookArch{
					Arm: bundleArchiveArtifacts["hook-aarch64-efi-initrd.iso"],
					Amd: bundleArchiveArtifacts["hook-x86_64-efi-initrd.iso"],
				},
				Vmlinuz: anywherev1alpha1.HookArch{
					Arm: bundleArchiveArtifacts["vmlinuz-aarch64"],
					Amd: bundleArchiveArtifacts["vmlinuz-x86_64"],
				},
			},
			Stack: bundleImageArtifacts["tinkerbell-helm"],
			Tink: anywherev1alpha1.TinkBundle{
				TinkRelayInit: bundleImageArtifacts["tink-relay-init"],
			},
			TinkebellChart: bundleImageArtifacts["tinkerbell-chart"],
			TinkerbellCrds: bundleImageArtifacts["tinkerbell-crds-helm"],
		},
	}

	return bundle, nil
}

func tinkerbellProviderVersion(gitTag string, metadataArtifact *releasetypes.ManifestArtifact, checksum string, dryRun bool) (string, error) {
	if _, err := utilversion.ParseSemantic(gitTag); err == nil {
		return fmt.Sprintf("%s+%s", gitTag, checksum), nil
	}

	if sha, err := hex.DecodeString(gitTag); err != nil || len(sha) != 20 {
		return "", fmt.Errorf("CAPT source version %q must be a semantic version or a full commit SHA", gitTag)
	}
	if dryRun {
		// Dry-run generation must work without downloaded metadata.
		return fmt.Sprintf("v0.0.0-dev-%s+%s", gitTag, checksum), nil
	}
	if metadataArtifact == nil {
		return "", fmt.Errorf("CAPT metadata is required for commit %s", gitTag)
	}
	metadataPath := filepath.Join(metadataArtifact.ArtifactPath, metadataArtifact.ReleaseName)
	contents, err := os.ReadFile(metadataPath)
	if err != nil {
		return "", errors.Wrap(err, "reading CAPT metadata")
	}
	var metadata struct {
		ReleaseSeries []struct {
			Major uint `json:"major"`
			Minor uint `json:"minor"`
		} `json:"releaseSeries"`
	}
	if err := yaml.Unmarshal(contents, &metadata); err != nil {
		return "", errors.Wrap(err, "parsing CAPT metadata")
	}
	if len(metadata.ReleaseSeries) == 0 {
		return "", fmt.Errorf("CAPT metadata has no release series")
	}

	// Commit pins use the newest release series declared by the built provider.
	series := metadata.ReleaseSeries[0]
	for _, candidate := range metadata.ReleaseSeries[1:] {
		if candidate.Major > series.Major || (candidate.Major == series.Major && candidate.Minor > series.Minor) {
			series = candidate
		}
	}
	return fmt.Sprintf("v%d.%d.0-dev-%s+%s", series.Major, series.Minor, gitTag, checksum), nil
}
