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

package artifacts

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
)

func IsObjectNotFoundError(err error) bool {
	return err.Error() == "Requested object not found"
}

func IsImageNotFoundError(err error) bool {
	return err.Error() == "Requested image not found"
}

func GetFakeSHA(hashType int) (string, error) {
	if (hashType != 256) && (hashType != 512) {
		return "", fmt.Errorf("unsupported hash algorithm: %d", hashType)
	}

	var shaSum string
	if hashType == 256 {
		shaSum = strings.Repeat(constants.HexAlphabet, 4)
	} else {
		shaSum = strings.Repeat(constants.HexAlphabet, 8)
	}
	return shaSum, nil
}

func GetLatestUploadDestination(sourcedFromBranch string) string {
	if sourcedFromBranch == "main" {
		return "latest"
	} else {
		return sourcedFromBranch
	}
}

// GetURI returns an full URL for the given path.
func GetURI(cdn, path string) (string, error) {
	uri, err := url.Parse(cdn)
	if err != nil {
		return "", err
	}
	uri.Path = path
	return uri.String(), nil
}

func SplitImageUri(imageUri, imageContainerRegistry string) (string, string) {
	imageUriSplit := strings.Split(imageUri, ":")
	imageRepository := strings.Replace(imageUriSplit[0], imageContainerRegistry+"/", "", -1)
	imageTag := imageUriSplit[1]

	return imageRepository, imageTag
}
