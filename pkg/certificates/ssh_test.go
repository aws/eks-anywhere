package certificates

import (
	"os"
	"strings"
	"testing"
)

// contains checks if a string contains another string.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// sshConfigTestCase defines a test case for initSSHConfig.
type sshConfigTestCase struct {
	name        string
	user        string
	keyPath     string
	passwd      string
	keyContent  string
	expectError bool
	errorMsg    string
}

// setupSSHKeyFile creates a temporary SSH key file with the given content.
func setupSSHKeyFile(t *testing.T, path, content string) func() {
	if content == "" {
		return func() {}
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return func() { os.Remove(path) }
}

// verifySSHConfig verifies that the SSH config was set correctly.
func verifySSHConfig(t *testing.T, r *Renewer, tt sshConfigTestCase) {
	if !tt.expectError {
		if r.sshConfig == nil {
			t.Error("SSH config was not set")
		}
		if r.sshConfig.User != tt.user {
			t.Errorf("expected SSH user %q, got %q", tt.user, r.sshConfig.User)
		}
		if r.sshKeyPath != tt.keyPath {
			t.Errorf("expected SSH key path %q, got %q", tt.keyPath, r.sshKeyPath)
		}
	}
}

// runSSHConfigTest runs a single initSSHConfig test case.
func runSSHConfigTest(t *testing.T, tt sshConfigTestCase) {
	cleanup := setupSSHKeyFile(t, tt.keyPath, tt.keyContent)
	defer cleanup()

	r, err := NewRenewer()
	if err != nil {
		t.Fatalf("failed to create renewer: %v", err)
	}

	err = r.initSSHConfig(tt.user, tt.keyPath, tt.passwd)
	if tt.expectError && err == nil {
		t.Error("expected error but got none")
	}
	if !tt.expectError && err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if tt.expectError && err != nil && tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
		t.Errorf("expected error message to contain %q, got %q", tt.errorMsg, err.Error())
	}

	verifySSHConfig(t, r, tt)
}

// getSSHConfigTestCases returns test cases for initSSHConfig.
func getSSHConfigTestCases() []sshConfigTestCase {
	return []sshConfigTestCase{
		{
			name:        "valid SSH key without passphrase",
			user:        "ec2-user",
			keyPath:     "/tmp/test-key-1",
			passwd:      "",
			keyContent:  testPrivateKey,
			expectError: false,
		},
		{
			name:        "invalid SSH key",
			user:        "ec2-user",
			keyPath:     "/tmp/test-key-2",
			passwd:      "",
			keyContent:  "invalid-key",
			expectError: true,
			errorMsg:    "parsing SSH key",
		},
		{
			name:        "non-existent SSH key file",
			user:        "ec2-user",
			keyPath:     "/tmp/non-existent-key",
			passwd:      "",
			keyContent:  "",
			expectError: true,
			errorMsg:    "reading SSH key",
		},
	}
}

// TestInitSSHConfig tests the initSSHConfig function.
func TestInitSSHConfig(t *testing.T) {
	tests := getSSHConfigTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runSSHConfigTest(t, tt)
		})
	}
}

// Test private key for SSH tests.
// This is a test key, not used for anything real.
var testPrivateKey = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACBsETg9gZQ5dSy+4qy7Cg4Zx7bE+KFi0xQyNKTJiM4YHwAAAJg2zz0UNs89
FAAAAAtzc2gtZWQyNTUxOQAAACBsETg9gZQ5dSy+4qy7Cg4Zx7bE+KFi0xQyNKTJiM4YHw
AAAEAIUUzgh0BfSZJ1JJ0NqQwO8FnIQgYyVFtZ3wYQEIQQoGwROD2BlDl1LL7irLsKDhnH
tsT4oWLTFDI0pMmIzhgfAAAAEHRlc3RAZXhhbXBsZS5jb20BAgMEBQ==
-----END OPENSSH PRIVATE KEY-----`
