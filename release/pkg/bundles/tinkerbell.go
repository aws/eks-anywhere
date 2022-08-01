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

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	releasetypes "github.com/aws/eks-anywhere/release/pkg/types"
	bundleutils "github.com/aws/eks-anywhere/release/pkg/util/bundles"
	"github.com/aws/eks-anywhere/release/pkg/version"
)

const (
	captProjectPath = "projects/tinkerbell/cluster-api-provider-tinkerbell"
	HookProjectPath = "projects/tinkerbell/hook"
)

func GetTinkerbellBundle(r *releasetypes.ReleaseConfig, imageDigests map[string]string) (anywherev1alpha1.TinkerbellBundle, error) {
	tinkerbellBundleArtifacts := map[string][]releasetypes.Artifact{
		"cluster-api-provider-tinkerbell": r.BundleArtifactsTable["cluster-api-provider-tinkerbell"],
		"kube-vip":                        r.BundleArtifactsTable["kube-vip"],
		"tink":                            r.BundleArtifactsTable["tink"],
		"hegel":                           r.BundleArtifactsTable["hegel"],
		"cfssl":                           r.BundleArtifactsTable["cfssl"],
		"pbnj":                            r.BundleArtifactsTable["pbnj"],
		"boots":                           r.BundleArtifactsTable["boots"],
		"hub":                             r.BundleArtifactsTable["hub"],
		"hook":                            r.BundleArtifactsTable["hook"],
		"rufio":                           r.BundleArtifactsTable["rufio"],
		"tinkerbell-chart":                r.BundleArtifactsTable["tinkerbell-chart"],
	}
	sortedComponentNames := bundleutils.SortArtifactsMap(tinkerbellBundleArtifacts)

	var sourceBranch string
	var componentChecksum string
	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}
	bundleArchiveArtifacts := map[string]anywherev1alpha1.Archive{}
	artifactHashes := []string{}

	for _, componentName := range sortedComponentNames {
		for _, artifact := range tinkerbellBundleArtifacts[componentName] {
			if artifact.Image != nil {
				imageArtifact := artifact.Image
				if componentName == "cluster-api-provider-tinkerbell" {
					sourceBranch = imageArtifact.SourcedFromBranch
				}
				bundleImageArtifact := anywherev1alpha1.Image{
					Name:        imageArtifact.AssetName,
					Description: fmt.Sprintf("Container image for %s image", imageArtifact.AssetName),
					OS:          imageArtifact.OS,
					Arch:        imageArtifact.Arch,
					URI:         imageArtifact.ReleaseImageURI,
					ImageDigest: imageDigests[imageArtifact.ReleaseImageURI],
				}
				bundleImageArtifacts[imageArtifact.AssetName] = bundleImageArtifact
				artifactHashes = append(artifactHashes, bundleImageArtifact.ImageDigest)
			}

			if artifact.Manifest != nil {
				manifestArtifact := artifact.Manifest
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
	version, err := version.BuildComponentVersion(
		version.NewVersionerWithGITTAG(r.BuildRepoSource, captProjectPath, sourceBranch, r),
		componentChecksum,
	)
	if err != nil {
		return anywherev1alpha1.TinkerbellBundle{}, errors.Wrapf(err, "Error getting version for cluster-api-provider-tinkerbell")
	}

	// TODO: remove these 2 lines when CAPT releases a new git tag
	_ = version
	version = "v0.1.0"

	bundle := anywherev1alpha1.TinkerbellBundle{
		Version:              version,
		ClusterAPIController: bundleImageArtifacts["cluster-api-provider-tinkerbell"],
		KubeVip:              bundleImageArtifacts["kube-vip"],
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
			Boots: bundleImageArtifacts["boots"],
			Cfssl: bundleImageArtifacts["cfssl"],
			Hegel: bundleImageArtifacts["hegel"],
			Hook: anywherev1alpha1.HookBundle{
				Bootkit: bundleImageArtifacts["hook-bootkit"],
				Docker:  bundleImageArtifacts["hook-docker"],
				Kernel:  bundleImageArtifacts["hook-kernel"],
				Initramfs: anywherev1alpha1.HookArch{
					Arm: bundleArchiveArtifacts["initramfs-aarch64"],
					Amd: bundleArchiveArtifacts["initramfs-x86_64"],
				},
				Vmlinuz: anywherev1alpha1.HookArch{
					Arm: bundleArchiveArtifacts["vmlinuz-aarch64"],
					Amd: bundleArchiveArtifacts["vmlinuz-x86_64"],
				},
			},
			Rufio: bundleImageArtifacts["rufio"],
			Tink: anywherev1alpha1.TinkBundle{
				TinkController: bundleImageArtifacts["tink-controller"],
				TinkServer:     bundleImageArtifacts["tink-server"],
				TinkWorker:     bundleImageArtifacts["tink-worker"],
			},
			TinkebellChart: bundleImageArtifacts["tinkerbell-chart"],
		},
	}

	return bundle, nil
}
