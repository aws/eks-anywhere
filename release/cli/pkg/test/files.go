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

package test

import (
	"os"
	"os/exec"
	"testing"

	commandutils "github.com/aws/eks-anywhere/release/cli/pkg/util/command"
)

func CheckFilesEquals(t *testing.T, actualPath, expectedPath string, update bool) {
	t.Helper()
	actualContent, err := readFile(actualPath)
	if err != nil {
		t.Fatalf("Error reading actual path %s:\n%v", actualPath, err)
	}

	if update {
		err = os.WriteFile(expectedPath, []byte(actualContent), 0o644)
		if err != nil {
			t.Fatalf("Error updating testdata bundle: %v\n", err)
		}
	}

	expectedContent, err := readFile(expectedPath)
	if err != nil {
		t.Fatalf("Error reading expected path %s:\n%v", expectedPath, err)
	}

	if actualContent != expectedContent {
		diffCmd := exec.Command("diff", expectedPath, actualPath)
		diff, err := commandutils.ExecCommand(diffCmd)
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				if exitError.ExitCode() == 1 {
					t.Fatalf("Actual file differs from expected:\n%s", string(diff))
				}
			}
		}
		t.Fatalf("Actual and expected files are different, actual =\n  %s \n expected =\n %s\n%s", actualContent, expectedContent, err)
	}
}

func readFile(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
