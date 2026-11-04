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
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCloneRepo_Success(t *testing.T) {
	srcDir := t.TempDir()
	setupLocalRepoWithMultipleCommits(t, srcDir, 3)

	destDir := filepath.Join(t.TempDir(), "cloned")

	output, err := CloneRepo("file://"+srcDir, destDir)
	if err != nil {
		t.Fatalf("CloneRepo failed: %v\nOutput: %s", err, output)
	}

	headOut, err := GetHead(destDir)
	if err != nil {
		t.Fatalf("GetHead on cloned repo failed: %v", err)
	}
	if len(headOut) < 7 {
		t.Fatalf("expected valid commit SHA, got: %q", headOut)
	}
}

func TestCloneRepo_CommandInjection_ShellMetacharacters(t *testing.T) {
	// This test verifies the security fix: shell metacharacters in the clone URL
	// must NOT be interpreted. With the old bash -c approach, a URL containing
	// "; malicious_command" would execute that command. The fixed version passes
	// args directly to exec, so metacharacters are treated as literal strings.
	//
	// We test with a URL that contains a semicolon followed by a command that
	// would create a marker file. If the marker file exists after CloneRepo,
	// the injection succeeded (test should fail).

	markerDir := t.TempDir()
	markerFile := filepath.Join(markerDir, "injected")

	// Craft a malicious URL: if shell-interpreted, "touch <markerFile>" would run
	maliciousUrl := "https://example.com/repo.git; touch " + markerFile

	destDir := filepath.Join(t.TempDir(), "dest")

	// CloneRepo will fail because the URL is invalid, but the important thing
	// is that the injected command does NOT execute
	_, _ = CloneRepo(maliciousUrl, destDir)

	if _, err := os.Stat(markerFile); err == nil {
		t.Fatal("SECURITY: command injection succeeded — marker file was created")
	}
}

func TestCloneRepo_CommandInjection_Backticks(t *testing.T) {
	markerDir := t.TempDir()
	markerFile := filepath.Join(markerDir, "injected_backtick")

	maliciousUrl := "`touch " + markerFile + "`"

	destDir := filepath.Join(t.TempDir(), "dest")
	_, _ = CloneRepo(maliciousUrl, destDir)

	if _, err := os.Stat(markerFile); err == nil {
		t.Fatal("SECURITY: backtick command injection succeeded — marker file was created")
	}
}

func TestCloneRepo_CommandInjection_DollarSubstitution(t *testing.T) {
	markerDir := t.TempDir()
	markerFile := filepath.Join(markerDir, "injected_dollar")

	maliciousUrl := "$(touch " + markerFile + ")"

	destDir := filepath.Join(t.TempDir(), "dest")
	_, _ = CloneRepo(maliciousUrl, destDir)

	if _, err := os.Stat(markerFile); err == nil {
		t.Fatal("SECURITY: $() command injection succeeded — marker file was created")
	}
}

func TestCloneRepo_CommandInjection_PipeRedirect(t *testing.T) {
	markerDir := t.TempDir()
	markerFile := filepath.Join(markerDir, "injected_pipe")

	maliciousUrl := "https://example.com/repo.git | touch " + markerFile

	destDir := filepath.Join(t.TempDir(), "dest")
	_, _ = CloneRepo(maliciousUrl, destDir)

	if _, err := os.Stat(markerFile); err == nil {
		t.Fatal("SECURITY: pipe command injection succeeded — marker file was created")
	}
}

func TestCloneRepo_CommandInjection_DestinationParam(t *testing.T) {
	// Also test that shell metacharacters in the destination parameter are safe
	markerDir := t.TempDir()
	markerFile := filepath.Join(markerDir, "injected_dest")

	srcDir := t.TempDir()
	setupLocalRepo(t, srcDir)

	maliciousDest := "/tmp/foo; touch " + markerFile

	_, _ = CloneRepo("file://"+srcDir, maliciousDest)

	if _, err := os.Stat(markerFile); err == nil {
		t.Fatal("SECURITY: command injection via destination parameter succeeded")
	}
}

func TestCloneRepo_InvalidUrl_ReturnsError(t *testing.T) {
	destDir := filepath.Join(t.TempDir(), "dest")

	_, err := CloneRepo("not-a-valid-url", destDir)
	if err == nil {
		t.Fatal("expected error for invalid clone URL, got nil")
	}
}

func TestCloneRepo_DestinationAlreadyExists_ReturnsError(t *testing.T) {
	srcDir := t.TempDir()
	setupLocalRepo(t, srcDir)

	destDir := t.TempDir() // already exists as non-empty dir
	// Create a file inside to make git clone fail
	if err := os.WriteFile(filepath.Join(destDir, "blocker"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := CloneRepo("file://"+srcDir, destDir)
	if err == nil {
		t.Fatal("expected error when destination already exists with content, got nil")
	}
}

func TestCloneRepo_OutputContainsContent(t *testing.T) {
	srcDir := t.TempDir()
	setupLocalRepoWithMultipleCommits(t, srcDir, 3)

	destDir := filepath.Join(t.TempDir(), "cloned")

	output, err := CloneRepo("file://"+srcDir, destDir)
	if err != nil {
		t.Fatalf("CloneRepo failed: %v\nOutput: %s", err, output)
	}
	// Output should be non-empty (git pull --all typically outputs something)
	// but we mainly verify it doesn't panic or lose output
	t.Logf("CloneRepo output: %q", output)
}

func TestCloneRepo_RemoteOriginFetchConfigured(t *testing.T) {
	srcDir := t.TempDir()
	setupLocalRepoWithMultipleCommits(t, srcDir, 3)

	destDir := filepath.Join(t.TempDir(), "cloned")

	_, err := CloneRepo("file://"+srcDir, destDir)
	if err != nil {
		t.Fatalf("CloneRepo failed: %v", err)
	}

	// Verify the fetch refspec was set correctly
	cmd := exec.Command("git", "-C", destDir, "config", "--get-all", "remote.origin.fetch")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to get remote.origin.fetch: %v, output: %s", err, out)
	}

	expected := "+refs/heads/*:refs/remotes/origin/*"
	if got := string(out); got == "" {
		t.Fatal("remote.origin.fetch is empty")
	} else if !contains(got, expected) {
		t.Fatalf("expected remote.origin.fetch to contain %q, got: %q", expected, got)
	}
}

func TestCloneRepo_RepoIsUnshallowed(t *testing.T) {
	srcDir := t.TempDir()
	setupLocalRepoWithMultipleCommits(t, srcDir, 3)

	destDir := filepath.Join(t.TempDir(), "cloned")

	_, err := CloneRepo("file://"+srcDir, destDir)
	if err != nil {
		t.Fatalf("CloneRepo failed: %v", err)
	}

	// A shallow repo has .git/shallow; after unshallow it should be gone
	shallowFile := filepath.Join(destDir, ".git", "shallow")
	if _, err := os.Stat(shallowFile); err == nil {
		t.Fatal("repo is still shallow after CloneRepo — .git/shallow exists")
	}
}

func TestCheckoutRepo_Success(t *testing.T) {
	srcDir := t.TempDir()
	setupLocalRepoWithBranch(t, srcDir, "test-branch")

	_, err := CheckoutRepo(srcDir, "test-branch")
	if err != nil {
		t.Fatalf("CheckoutRepo failed: %v", err)
	}

	branch, err := GetCurrentBranch(srcDir)
	if err != nil {
		t.Fatalf("GetCurrentBranch failed: %v", err)
	}
	if branch != "test-branch" {
		t.Fatalf("expected branch 'test-branch', got %q", branch)
	}
}

func TestCheckoutRepo_InvalidBranch(t *testing.T) {
	srcDir := t.TempDir()
	setupLocalRepo(t, srcDir)

	_, err := CheckoutRepo(srcDir, "nonexistent-branch")
	if err == nil {
		t.Fatal("expected error for nonexistent branch, got nil")
	}
}

func TestGetHead_ValidRepo(t *testing.T) {
	srcDir := t.TempDir()
	setupLocalRepo(t, srcDir)

	head, err := GetHead(srcDir)
	if err != nil {
		t.Fatalf("GetHead failed: %v", err)
	}
	if len(head) != 40 {
		t.Fatalf("expected 40-char SHA, got %d chars: %q", len(head), head)
	}
}

func TestGetHead_InvalidDir(t *testing.T) {
	_, err := GetHead("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent path, got nil")
	}
}

func TestDescribeTag_NoTags(t *testing.T) {
	srcDir := t.TempDir()
	setupLocalRepo(t, srcDir)

	_, err := DescribeTag(srcDir)
	if err == nil {
		t.Fatal("expected error when no tags exist, got nil")
	}
}

func TestDescribeTag_WithTag(t *testing.T) {
	srcDir := t.TempDir()
	setupLocalRepo(t, srcDir)

	run(t, srcDir, "git", "tag", "v1.0.0")

	tag, err := DescribeTag(srcDir)
	if err != nil {
		t.Fatalf("DescribeTag failed: %v", err)
	}
	if tag != "v1.0.0" {
		t.Fatalf("expected tag 'v1.0.0', got %q", tag)
	}
}

func TestGetRepoTagsDescending(t *testing.T) {
	srcDir := t.TempDir()
	setupLocalRepoWithMultipleCommits(t, srcDir, 3)

	run(t, srcDir, "git", "tag", "v1.0.0", "HEAD~2")
	run(t, srcDir, "git", "tag", "v2.0.0", "HEAD~1")
	run(t, srcDir, "git", "tag", "v3.0.0", "HEAD")

	tags, err := GetRepoTagsDescending(srcDir)
	if err != nil {
		t.Fatalf("GetRepoTagsDescending failed: %v", err)
	}

	// Should be descending: v3, v2, v1
	if !contains(tags, "v3.0.0") || !contains(tags, "v1.0.0") {
		t.Fatalf("expected all tags, got: %q", tags)
	}

	// Check ordering
	v3Pos := indexOf(tags, "v3.0.0")
	v1Pos := indexOf(tags, "v1.0.0")
	if v3Pos > v1Pos {
		t.Fatalf("expected v3.0.0 before v1.0.0 (descending), got: %q", tags)
	}
}

func TestGetCurrentBranch(t *testing.T) {
	srcDir := t.TempDir()
	setupLocalRepo(t, srcDir)

	branch, err := GetCurrentBranch(srcDir)
	if err != nil {
		t.Fatalf("GetCurrentBranch failed: %v", err)
	}
	// Default branch after git init with initial commit
	if branch == "" {
		t.Fatal("expected non-empty branch name")
	}
}

// --- Helpers ---

func setupLocalRepo(t *testing.T, dir string) {
	t.Helper()
	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
	writeFile(t, filepath.Join(dir, "README.md"), "initial")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "initial commit")
}

func setupLocalRepoWithMultipleCommits(t *testing.T, dir string, n int) {
	t.Helper()
	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
	for i := range n {
		writeFile(t, filepath.Join(dir, "file.txt"), string(rune('a'+i)))
		run(t, dir, "git", "add", ".")
		run(t, dir, "git", "commit", "-m", "commit "+string(rune('0'+i)))
	}
}

func setupLocalRepoWithBranch(t *testing.T, dir, branch string) {
	t.Helper()
	setupLocalRepo(t, dir)
	run(t, dir, "git", "branch", branch)
}

func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command %q failed in %s: %v\nOutput: %s", append([]string{name}, args...), dir, err, out)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && indexOfStr(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	return indexOfStr(s, substr)
}

func indexOfStr(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
