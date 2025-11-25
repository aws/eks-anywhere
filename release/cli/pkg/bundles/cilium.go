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

const (
	ciliumImageName         = "cilium"
	ciliumOperatorImageName = "operator-generic"
	ciliumHelmChartName     = "cilium-chart"
	ciliumHelmChart         = "cilium"
	ciliumImage             = "cilium"
	ciliumOperatorImage     = "operator-generic"
)

func GetCiliumBundle(r *releasetypes.ReleaseConfig) (anywherev1alpha1.CiliumBundle, error) {
	ciliumArtifacts, err := r.BundleArtifactsTable.Load("cilium")
	if err != nil {
		return anywherev1alpha1.CiliumBundle{}, fmt.Errorf("artifacts for project cilium not found in bundle artifacts table")
	}

	ciliumContainerRegistry := "public.ecr.aws/isovalent"
	ciliumGitTag, err := filereader.ReadGitTag(constants.CiliumProjectPath, r.BuildRepoSource, r.BuildRepoBranchName)
	if err != nil {
		return anywherev1alpha1.CiliumBundle{}, errors.Cause(err)
	}
	ciliumImages := []imageDefinition{
		containerImage(ciliumImageName, ciliumImage, ciliumContainerRegistry, ciliumGitTag),
		containerImage(ciliumOperatorImageName, ciliumOperatorImage, ciliumContainerRegistry, ciliumGitTag),
		// Helm charts are in the same repository and have the same
		// sem version as the corresponding container image but omiting the initial "v"
		chart(ciliumHelmChartName, ciliumHelmChart, ciliumContainerRegistry, strings.TrimPrefix(ciliumGitTag, "v")),
	}

	bundleImageArtifacts := map[string]anywherev1alpha1.Image{}
	bundleManifestArtifacts := map[string]anywherev1alpha1.Manifest{}

	for _, imageDef := range ciliumImages {
		imageDigest, err := getCiliumImageDigest(r.BuildRepoSource, imageDef.name)
		if err != nil {
			return anywherev1alpha1.CiliumBundle{}, errors.Cause(err)
		}

		bundleImageArtifacts[imageDef.name] = imageDef.builder(imageDigest)
	}

	for _, artifact := range ciliumArtifacts {
		if artifact.Manifest != nil {
			manifestArtifact := artifact.Manifest
			bundleManifestArtifact := anywherev1alpha1.Manifest{
				URI: manifestArtifact.ReleaseCdnURI,
			}

			bundleManifestArtifacts[manifestArtifact.ReleaseName] = bundleManifestArtifact
		}
	}

	bundle := anywherev1alpha1.CiliumBundle{
		Version:   ciliumGitTag,
		Cilium:    bundleImageArtifacts[ciliumImageName],
		Operator:  bundleImageArtifacts[ciliumOperatorImageName],
		Manifest:  bundleManifestArtifacts["cilium.yaml"],
		HelmChart: bundleImageArtifacts[ciliumHelmChartName],
	}

	return bundle, nil
}

func getCiliumImageDigest(gitRootPath, imageName string) (string, error) {
	projectSource := "projects/cilium/cilium"
	imageDigestFileName := fmt.Sprintf("images/%s/IMAGE_DIGEST", imageName)
	imageDigestFile := filepath.Join(gitRootPath, projectSource, imageDigestFileName)
	imageDigest, err := filereader.ReadFileContentsTrimmed(imageDigestFile)
	if err != nil {
		return "", errors.Cause(err)
	}

	return imageDigest, nil
}

type imageDefinition struct {
	name, image, registry, tag string
	builder                    imageBuilder
}

type imageBuilder func(digest string) anywherev1alpha1.Image

func containerImage(name, image, registry, tag string) imageDefinition {
	return imageDefinition{
		name:     name,
		image:    image,
		registry: registry,
		tag:      tag,
		builder: func(digest string) anywherev1alpha1.Image {
			return anywherev1alpha1.Image{
				Name:        name,
				Description: fmt.Sprintf("Container image for %s image", name),
				OS:          "linux",
				Arch:        []string{"amd64"},
				URI:         fmt.Sprintf("%s/%s:%s", registry, image, tag),
				ImageDigest: digest,
			}
		},
	}
}

func chart(name, image, registry, tag string) imageDefinition {
	return imageDefinition{
		name:     name,
		image:    image,
		registry: registry,
		tag:      tag,
		builder: func(digest string) anywherev1alpha1.Image {
			return anywherev1alpha1.Image{
				Name:        name,
				Description: fmt.Sprintf("Helm chart for %s", name),
				URI:         fmt.Sprintf("%s/%s:%s", registry, image, tag),
				ImageDigest: digest,
			}
		},
	}
}
