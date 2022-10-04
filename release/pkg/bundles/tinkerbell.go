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
	"strings"

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/pkg/constants"
	"github.com/aws/eks-anywhere/release/pkg/helm"
	releasetypes "github.com/aws/eks-anywhere/release/pkg/types"
	bundleutils "github.com/aws/eks-anywhere/release/pkg/util/bundles"
	"github.com/aws/eks-anywhere/release/pkg/version"
)

func GetTinkerbellBundle(r *releasetypes.ReleaseConfig, imageDigests map[string]string) (anywherev1alpha1.TinkerbellBundle, error) {
	tinkerbellBundleArtifacts := map[string][]releasetypes.Artifact{
		"cluster-api-provider-tinkerbell": r.BundleArtifactsTable["cluster-api-provider-tinkerbell"],
		"kube-vip":                        r.BundleArtifactsTable["kube-vip"],
		"envoy":                           r.BundleArtifactsTable["envoy"],
		"tink":                            r.BundleArtifactsTable["tink"],
		"hegel":                           r.BundleArtifactsTable["hegel"],
		"cfssl":                           r.BundleArtifactsTable["cfssl"],
		"boots":                           r.BundleArtifactsTable["boots"],
		"hub":                             r.BundleArtifactsTable["hub"],
		"hook":                            r.BundleArtifactsTable["hook"],
		"rufio":                           r.BundleArtifactsTable["rufio"],
		"tinkerbell-chart":                r.BundleArtifactsTable["tinkerbell-chart"],
	}
	sortedComponentNames := bundleutils.SortArtifactsMap(tinkerbellBundleArtifacts)

	var helmdir string
	var URI string

	// Find the source of the Helm chart prior to the initial loop.
	for _, componentName := range sortedComponentNames {
		for _, artifact := range tinkerbellBundleArtifacts[componentName] {
			if artifact.Image != nil && strings.HasSuffix(artifact.Image.AssetName, "chart") {
				URI = artifact.Image.SourceImageURI
			}
		}
	}
	driver, err := helm.NewHelm()
	if err != nil {
		return anywherev1alpha1.TinkerbellBundle{}, fmt.Errorf("creating helm client: %w", err)
	}

	if !r.DryRun {
		helmdir, err = helm.GetHelmDest(driver, URI, "tinkerbell-chart")
		if err != nil {
			return anywherev1alpha1.TinkerbellBundle{}, errors.Wrap(err, "Error GetHelmDest")
		}
	}

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
				bundleImageArtifact := anywherev1alpha1.Image{}
				if componentName == "cluster-api-provider-tinkerbell" {
					sourceBranch = imageArtifact.SourcedFromBranch
				}
				if strings.HasSuffix(imageArtifact.AssetName, "chart") {
					if !r.DryRun {
						err := helm.ModifyChartYaml(*imageArtifact, r, driver, helmdir)
						if err != nil {
							return anywherev1alpha1.TinkerbellBundle{}, errors.Wrap(err, "Error modifying and pushing helm Chart.yaml")
						}
					}
					bundleImageArtifact = anywherev1alpha1.Image{
						Name:        imageArtifact.AssetName,
						Description: fmt.Sprintf("Helm chart for %s", imageArtifact.AssetName),
						URI:         imageArtifact.ReleaseImageURI,
						ImageDigest: imageDigests[imageArtifact.ReleaseImageURI],
					}
				}
				bundleImageArtifact = anywherev1alpha1.Image{
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
		version.NewVersionerWithGITTAG(r.BuildRepoSource, constants.CaptProjectPath, sourceBranch, r),
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
