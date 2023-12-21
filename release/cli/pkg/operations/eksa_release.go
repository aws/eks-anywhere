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

package operations

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/release/cli/pkg/bundles"
	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
)

func GenerateEksAArtifactsTable(r *releasetypes.ReleaseConfig) (releasetypes.ArtifactsTable, error) {
	fmt.Println("\n==========================================================")
	fmt.Println("                 EKS-A Artifacts Table Generation")
	fmt.Println("==========================================================")

	var artifactsTable releasetypes.ArtifactsTable
	artifacts, err := bundles.GetEksACliArtifacts(r)
	if err != nil {
		return releasetypes.ArtifactsTable{}, errors.Wrapf(err, "Error getting artifact information for EKS-A CLI")
	}

	artifactsTable.Store("eks-a-cli", artifacts)

	fmt.Printf("%s Successfully generated EKS-A artifacts table\n", constants.SuccessIcon)

	return artifactsTable, nil
}

func EksAArtifactsRelease(r *releasetypes.ReleaseConfig) error {
	fmt.Println("\n==========================================================")
	fmt.Println("                 EKS-A CLI Artifacts Release")
	fmt.Println("==========================================================")
	err := DownloadArtifacts(context.Background(), r, r.EksAArtifactsTable)
	if err != nil {
		return errors.Cause(err)
	}

	err = RenameArtifacts(context.Background(), r, r.EksAArtifactsTable)
	if err != nil {
		return errors.Cause(err)
	}

	err = UploadArtifacts(context.Background(), r, r.EksAArtifactsTable)
	if err != nil {
		return errors.Cause(err)
	}

	return nil
}
