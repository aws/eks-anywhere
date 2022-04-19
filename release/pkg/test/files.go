package test

import (
	"io/ioutil"
	"os/exec"
	"strings"
	"testing"

	"github.com/aws/eks-anywhere/release/pkg/utils"
)

func CheckFilesEquals(t *testing.T, actualPath, expectedPath string) {
	t.Helper()
	actualContent, err := readFile(actualPath)
	if err != nil {
		t.Fatalf("Error reading actual path %s:\n%v", actualPath, err)
	}

	expectedContent, err := readFile(expectedPath)
	if err != nil {
		t.Fatalf("Error reading expected path %s:\n%v", expectedPath, err)
	}

	if actualContent != expectedContent {
		diffCmd := exec.Command("diff", expectedPath, actualPath)
		diff, err := utils.ExecCommand(diffCmd)
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
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
