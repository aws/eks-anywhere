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
	"fmt"
	"os/exec"

	commandutils "github.com/aws/eks-anywhere/release/cli/pkg/util/command"
)

func CloneRepo(cloneUrl, destination string) (string, error) {
	cloneRepoCommandSequence := fmt.Sprintf("git clone --depth 1 %s %[2]s; cd %[2]s; git config --unset-all remote.origin.fetch; git config --add remote.origin.fetch '+refs/heads/*:refs/remotes/origin/*'; git fetch --unshallow; git pull --all", cloneUrl, destination)
	cmd := exec.Command("bash", "-c", cloneRepoCommandSequence)
	return commandutils.ExecCommand(cmd)
}

func CheckoutRepo(gitRoot, branch string) (string, error) {
	cmd := exec.Command("git", "-C", gitRoot, "checkout", branch)
	return commandutils.ExecCommand(cmd)
}

func DescribeTag(gitRoot string) (string, error) {
	cmd := exec.Command("git", "-C", gitRoot, "describe", "--tag")
	return commandutils.ExecCommand(cmd)
}

func GetRepoTagsDescending(gitRoot string) (string, error) {
	cmd := exec.Command("git", "-C", gitRoot, "tag", "-l", "v*", "--sort", "-v:refname")
	return commandutils.ExecCommand(cmd)
}

func GetHead(gitRoot string) (string, error) {
	cmd := exec.Command("git", "-C", gitRoot, "rev-parse", "HEAD")
	return commandutils.ExecCommand(cmd)
}

func GetRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	return commandutils.ExecCommand(cmd)
}

func GetCurrentBranch(gitRoot string) (string, error) {
	cmd := exec.Command("git", "-C", gitRoot, "branch", "--show-current")
	return commandutils.ExecCommand(cmd)
}
