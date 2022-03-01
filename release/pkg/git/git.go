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

package git

import (
	"os/exec"

	"github.com/aws/eks-anywhere/release/pkg/utils"
)

func CloneRepo(cloneUrl, destination string) (string, error) {
	cmd := exec.Command("git", "clone", cloneUrl, destination)
	return utils.ExecCommand(cmd)
}

func CheckoutRepo(gitRoot, branch string) (string, error) {
	cmd := exec.Command("git", "-C", gitRoot, "checkout", branch)
	return utils.ExecCommand(cmd)
}

func DescribeTag(gitRoot string) (string, error) {
	cmd := exec.Command("git", "-C", gitRoot, "describe", "--tag")
	return utils.ExecCommand(cmd)
}

func GetRepoTagsDescending(gitRoot string) (string, error) {
	cmd := exec.Command("git", "-C", gitRoot, "tag", "-l", "--sort", "-v:refname")
	return utils.ExecCommand(cmd)
}

func GetLatestCommitForPath(gitRoot, path string) (string, error) {
	cmd := exec.Command("git", "-C", gitRoot, "log", "--pretty=format:%h", "-n1", path)
	return utils.ExecCommand(cmd)
}

func GetHead(gitRoot string) (string, error) {
	cmd := exec.Command("git", "-C", gitRoot, "rev-parse", "HEAD")
	return utils.ExecCommand(cmd)
}
