package certificates

import (
	"os"
	"strings"
	"testing"
)

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// TestParseConfigFileNotFound tests the ParseConfig function with a non-existent file.
func TestParseConfigFileNotFound(t *testing.T) {
	_, err := ParseConfig("non-existent-file.yaml")
	if err == nil {
		t.Error("expected error for non-existent file but got none")
	}
	if !strings.Contains(err.Error(), "reading config file") {
		t.Errorf("expected error message to contain 'reading config file', got %q", err.Error())
	}
}

// setupSSHKeyForTest creates a temporary SSH key file for testing.
func setupSSHKeyForTest(t *testing.T, path string) func() {
	if path == "" {
		return func() {}
	}

	if err := os.WriteFile(path, []byte("test-key"), 0o600); err != nil {
		t.Fatal(err)
	}
	return func() { os.Remove(path) }
}

// validateConfigTestCase defines a test case for validateConfig.
type validateConfigTestCase struct {
	name        string
	config      *RenewalConfig
	expectError bool
	errorMsg    string
}

// runValidateConfigTest runs a single validateConfig test case.
func runValidateConfigTest(t *testing.T, tt validateConfigTestCase) {
	cleanup := setupSSHKeyForTest(t, tt.config.ControlPlane.SSHKey)
	defer cleanup()

	err := validateConfig(tt.config)
	if tt.expectError && err == nil {
		t.Error("expected error but got none")
	}
	if !tt.expectError && err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if tt.expectError && err != nil && tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
		t.Errorf("expected error message to contain %q, got %q", tt.errorMsg, err.Error())
	}
}

// TestValidateConfig tests the validateConfig function directly.
func TestValidateConfig(t *testing.T) {
	tests := []validateConfigTestCase{
		{
			name: "valid config",
			config: &RenewalConfig{
				ClusterName: "test-cluster",
				ControlPlane: NodeConfig{
					Nodes:   []string{"192.168.1.10"},
					OS:      "ubuntu",
					SSHKey:  "/tmp/test-key",
					SSHUser: "ec2-user",
				},
			},
			expectError: false,
		},
		{
			name: "missing cluster name",
			config: &RenewalConfig{
				ControlPlane: NodeConfig{
					Nodes:   []string{"192.168.1.10"},
					OS:      "ubuntu",
					SSHKey:  "/tmp/test-key",
					SSHUser: "ec2-user",
				},
			},
			expectError: true,
			errorMsg:    "cluster name is required",
		},
		{
			name: "missing control plane nodes",
			config: &RenewalConfig{
				ClusterName: "test-cluster",
				ControlPlane: NodeConfig{
					OS:      "ubuntu",
					SSHKey:  "/tmp/test-key",
					SSHUser: "ec2-user",
				},
			},
			expectError: true,
			errorMsg:    "at least one node is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runValidateConfigTest(t, tt)
		})
	}
}

// TestSSHKeyNotFound tests the validateNodeConfig function with a non-existent SSH key file.
func TestSSHKeyNotFound(t *testing.T) {
	config := NodeConfig{
		Nodes:   []string{"192.168.1.10"},
		OS:      "ubuntu",
		SSHKey:  "/tmp/non-existent-key",
		SSHUser: "ec2-user",
	}

	err := validateNodeConfig(&config)
	if err == nil {
		t.Error("expected error for non-existent SSH key file but got none")
	}
	if !strings.Contains(err.Error(), "retrieving SSH key file information") {
		t.Errorf("expected error message to contain 'retrieving SSH key file information', got %q", err.Error())
	}
}

// parseConfigTestCase defines a test case for ParseConfig.
type parseConfigTestCase struct {
	name        string
	configYaml  string
	expectError bool
	errorMsg    string
}

// createTempConfigFile creates a temporary config file with the given YAML content.
func createTempConfigFile(t *testing.T, content string) string {
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	return tmpfile.Name()
}

// runParseConfigTest runs a single ParseConfig test case.
func runParseConfigTest(t *testing.T, tt parseConfigTestCase) {
	// Create temporary config file
	configFile := createTempConfigFile(t, tt.configYaml)
	defer os.Remove(configFile)

	// Create temporary SSH key file
	keyFile := "/tmp/test-key"
	cleanup := setupSSHKeyForTest(t, keyFile)
	defer cleanup()

	// Test config parsing
	_, err := ParseConfig(configFile)
	if tt.expectError && err == nil {
		t.Error("expected error but got none")
	}
	if !tt.expectError && err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if tt.expectError && err != nil && tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
		t.Errorf("expected error message to contain %q, got %q", tt.errorMsg, err.Error())
	}
}

// getParseConfigTestCases returns test cases for ParseConfig.
func getParseConfigTestCases() []parseConfigTestCase {
	return []parseConfigTestCase{
		{
			name: "valid config with both etcd and control plane",
			configYaml: `
clusterName: test-cluster
controlPlane:
  nodes:
  - 192.168.1.10
  os: ubuntu
  sshKey: /tmp/test-key
  sshUser: ec2-user
etcd:
  nodes:
  - 192.168.1.20
  os: ubuntu
  sshKey: /tmp/test-key
  sshUser: ec2-user
`,
			expectError: false,
		},
		{
			name: "valid config without etcd (embedded)",
			configYaml: `
clusterName: test-cluster
controlPlane:
  nodes:
  - 192.168.1.10
  os: ubuntu
  sshKey: /tmp/test-key
  sshUser: ec2-user
`,
			expectError: false,
		},
		{
			name: "invalid config - missing cluster name",
			configYaml: `
controlPlane:
  nodes:
  - 192.168.1.10
  os: ubuntu
  sshKey: /tmp/test-key
  sshUser: ec2-user
`,
			expectError: true,
		},
		{
			name: "invalid config - missing control plane nodes",
			configYaml: `
clusterName: test-cluster
controlPlane:
  os: ubuntu
  sshKey: /tmp/test-key
  sshUser: ec2-user
`,
			expectError: true,
		},
		{
			name: "invalid config - unsupported OS",
			configYaml: `
clusterName: test-cluster
controlPlane:
  nodes:
  - 192.168.1.10
  os: windows
  sshKey: /tmp/test-key
  sshUser: ec2-user
`,
			expectError: true,
			errorMsg:    "unsupported OS",
		},
		{
			name: "valid config with SSH password",
			configYaml: `
clusterName: test-cluster
controlPlane:
  nodes:
  - 192.168.1.10
  os: ubuntu
  sshKey: /tmp/test-key
  sshUser: ec2-user
  sshPasswd: password123
`,
			expectError: false,
		},
		{
			name: "valid config with redhat OS",
			configYaml: `
clusterName: test-cluster
controlPlane:
  nodes:
  - 192.168.1.10
  os: redhat
  sshKey: /tmp/test-key
  sshUser: ec2-user
`,
			expectError: false,
		},
		{
			name: "valid config with bottlerocket OS",
			configYaml: `
clusterName: test-cluster
controlPlane:
  nodes:
  - 192.168.1.10
  os: bottlerocket
  sshKey: /tmp/test-key
  sshUser: ec2-user
`,
			expectError: false,
		},
		{
			name: "valid config with different OS types",
			configYaml: `
clusterName: test-cluster
controlPlane:
  nodes:
  - 192.168.1.10
  os: ubuntu
  sshKey: /tmp/test-key
  sshUser: ec2-user
etcd:
  nodes:
  - 192.168.1.20
  os: bottlerocket
  sshKey: /tmp/test-key
  sshUser: ec2-user
`,
			expectError: false,
		},
	}
}

// TestParseConfig tests the ParseConfig function.
func TestParseConfig(t *testing.T) {
	tests := getParseConfigTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runParseConfigTest(t, tt)
		})
	}
}

// validateNodeConfigTestCase defines a test case for validateNodeConfig.
type validateNodeConfigTestCase struct {
	name        string
	config      NodeConfig
	component   string
	expectError bool
}

// runValidateNodeConfigTest runs a single validateNodeConfig test case.
func runValidateNodeConfigTest(t *testing.T, tt validateNodeConfigTestCase) {
	cleanup := setupSSHKeyForTest(t, tt.config.SSHKey)
	defer cleanup()

	err := validateNodeConfig(&tt.config)
	if tt.expectError && err == nil {
		t.Error("expected error but got none")
	}
	if !tt.expectError && err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// getValidateNodeConfigTestCases returns test cases for validateNodeConfig.
func getValidateNodeConfigTestCases() []validateNodeConfigTestCase {
	return []validateNodeConfigTestCase{
		{
			name: "valid ubuntu config",
			config: NodeConfig{
				Nodes:   []string{"192.168.1.10"},
				OS:      "ubuntu",
				SSHKey:  "/tmp/test-key",
				SSHUser: "ec2-user",
			},
			component:   "control plane",
			expectError: false,
		},
		{
			name: "valid redhat config",
			config: NodeConfig{
				Nodes:   []string{"192.168.1.10"},
				OS:      "redhat",
				SSHKey:  "/tmp/test-key",
				SSHUser: "ec2-user",
			},
			component:   "etcd",
			expectError: false,
		},
		{
			name: "valid bottlerocket config",
			config: NodeConfig{
				Nodes:   []string{"192.168.1.10"},
				OS:      "bottlerocket",
				SSHKey:  "/tmp/test-key",
				SSHUser: "ec2-user",
			},
			component:   "control plane",
			expectError: false,
		},
		{
			name: "invalid - missing nodes",
			config: NodeConfig{
				OS:      "ubuntu",
				SSHKey:  "/tmp/test-key",
				SSHUser: "ec2-user",
			},
			component:   "control plane",
			expectError: true,
		},
		{
			name: "invalid - unsupported OS",
			config: NodeConfig{
				Nodes:   []string{"192.168.1.10"},
				OS:      "windows",
				SSHKey:  "/tmp/test-key",
				SSHUser: "ec2-user",
			},
			component:   "control plane",
			expectError: true,
		},
		{
			name: "invalid - missing SSH key",
			config: NodeConfig{
				Nodes:   []string{"192.168.1.10"},
				OS:      "ubuntu",
				SSHUser: "ec2-user",
			},
			component:   "control plane",
			expectError: true,
		},
		{
			name: "invalid - missing SSH user",
			config: NodeConfig{
				Nodes:  []string{"192.168.1.10"},
				OS:     "ubuntu",
				SSHKey: "/tmp/test-key",
			},
			component:   "control plane",
			expectError: true,
		},
	}
}

// TestValidateNodeConfig tests the validateNodeConfig function.
func TestValidateNodeConfig(t *testing.T) {
	tests := getValidateNodeConfigTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runValidateNodeConfigTest(t, tt)
		})
	}
}
