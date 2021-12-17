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

package utils

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func ExecCommand(cmd *exec.Cmd) (string, error) {
	stdout, err := cmd.Output()
	if err != nil {
		return "", errors.Cause(err)
	}
	return string(stdout), nil
}

func SliceContains(s []string, str string) bool {
	for _, elem := range s {
		if elem == str {
			return true
		}
	}
	return false
}

func IsObjectNotFoundError(err error) bool {
	return err.Error() == "Requested object not found"
}

func IsImageNotFoundError(err error) bool {
	return err.Error() == "Requested image not found"
}

func SplitImageUri(imageUri, imageContainerRegistry string) (string, string) {
	imageUriSplit := strings.Split(imageUri, ":")
	imageRepository := strings.Replace(imageUriSplit[0], imageContainerRegistry+"/", "", -1)
	imageTag := imageUriSplit[1]

	return imageRepository, imageTag
}

func GetManifestFilepaths(devRelease bool, bundleNumber int, kind, branch string) string {
	var manifestFilepath string
	switch kind {
	case anywherev1alpha1.BundlesKind:
		if devRelease {
			if branch != "main" {
				manifestFilepath = fmt.Sprintf("%s/bundle-release.yaml", branch)
			} else {
				manifestFilepath = "bundle-release.yaml"
			}
		} else {
			manifestFilepath = fmt.Sprintf("releases/bundles/%d/manifest.yaml", bundleNumber)
		}
	case anywherev1alpha1.ReleaseKind:
		if devRelease {
			if branch != "main" {
				manifestFilepath = fmt.Sprintf("%s/eks-a-release.yaml", branch)
			} else {
				manifestFilepath = "eks-a-release.yaml"
			}
		} else {
			manifestFilepath = "releases/eks-a/manifest.yaml"
		}
	}
	return manifestFilepath
}
