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
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	"github.com/aws/eks-anywhere/release/cli/pkg/git"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
)

func SetRepoHeads(r *releasetypes.ReleaseConfig) error {
	fmt.Println("\n==========================================================")
	fmt.Println("                    Local Repository Setup")
	fmt.Println("==========================================================")

	// Get the repos from env var
	if r.CliRepoUrl == "" || r.BuildRepoUrl == "" {
		return fmt.Errorf("One or both clone URLs are empty")
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Cause(err)
	}
	parentSourceDir := filepath.Join(homeDir, "eks-a-source")

	// Clone the CLI repository
	fmt.Println("Cloning CLI repository")
	r.CliRepoSource = filepath.Join(parentSourceDir, "eks-a-cli")
	out, err := git.CloneRepo(r.CliRepoUrl, r.CliRepoSource)
	fmt.Println(out)
	if err != nil {
		return errors.Cause(err)
	}

	// Clone the build-tooling repository
	fmt.Println("Cloning build-tooling repository")
	r.BuildRepoSource = filepath.Join(parentSourceDir, "eks-a-build")
	out, err = git.CloneRepo(r.BuildRepoUrl, r.BuildRepoSource)
	fmt.Println(out)
	if err != nil {
		return errors.Cause(err)
	}

	if r.BuildRepoBranchName != "main" {
		fmt.Printf("Checking out build-tooling repo at branch %s\n", r.BuildRepoBranchName)
		out, err = git.CheckoutRepo(r.BuildRepoSource, r.BuildRepoBranchName)
		fmt.Println(out)
		if err != nil {
			return errors.Cause(err)
		}
	}

	if r.CliRepoBranchName != "main" {
		fmt.Printf("Checking out CLI repo at branch %s\n", r.CliRepoBranchName)
		out, err = git.CheckoutRepo(r.CliRepoSource, r.CliRepoBranchName)
		fmt.Println(out)
		if err != nil {
			return errors.Cause(err)
		}
	}

	// Set HEADs of the repos
	r.CliRepoHead, err = git.GetHead(r.CliRepoSource)
	if err != nil {
		return errors.Cause(err)
	}
	fmt.Printf("Head of cli repo: %s\n", r.CliRepoHead)

	r.BuildRepoHead, err = git.GetHead(r.BuildRepoSource)
	if err != nil {
		return errors.Cause(err)
	}
	fmt.Printf("Head of build repo: %s\n", r.BuildRepoHead)

	fmt.Printf("%s Successfully completed local repository setup\n", constants.SuccessIcon)

	return nil
}
