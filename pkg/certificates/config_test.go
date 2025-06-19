package certificates

import (
	"os"
	"testing"
)

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

// TestParseConfigFileNotFound tests the ParseConfig function with a non-existent file.
func TestParseConfigFileNotFound(t *testing.T) {
	_, err := ParseConfig("non-existent-file.yaml")
	if err == nil {
		t.Error("expected error for non-existent file but got none")
	}
}

// TestValidateConfig tests the validateConfig function directly.
func TestValidateConfig(t *testing.T) {
	// Setup SSH key once for all tests
	keyFile := "/tmp/test-key"
	cleanup := setupSSHKeyForTest(t, keyFile)
	defer cleanup()

	tests := []struct {
		name        string
		config      *RenewalConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &RenewalConfig{
				ClusterName: "test-cluster",
				OS:          "ubuntu",
				ControlPlane: NodeConfig{
					Nodes: []string{"192.168.1.10"},
					SSH: SSHConfig{
						User:    "ec2-user",
						KeyPath: keyFile,
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing cluster name",
			config: &RenewalConfig{
				OS: "ubuntu",
				ControlPlane: NodeConfig{
					Nodes: []string{"192.168.1.10"},
					SSH: SSHConfig{
						User:    "ec2-user",
						KeyPath: keyFile,
					},
				},
			},
			expectError: true,
		},
		{
			name: "missing control plane nodes",
			config: &RenewalConfig{
				ClusterName: "test-cluster",
				OS:          "ubuntu",
				ControlPlane: NodeConfig{
					SSH: SSHConfig{
						User:    "ec2-user",
						KeyPath: keyFile,
					},
				},
			},
			expectError: true,
		},
		{
			name: "non-existent SSH key file",
			config: &RenewalConfig{
				ClusterName: "test-cluster",
				OS:          "ubuntu",
				ControlPlane: NodeConfig{
					Nodes: []string{"192.168.1.10"},
					SSH: SSHConfig{
						User:    "ec2-user",
						KeyPath: "/tmp/non-existent-key",
					},
				},
			},
			expectError: true,
		},
		{
			name: "unsupported OS",
			config: &RenewalConfig{
				ClusterName: "test-cluster",
				OS:          "windows",
				ControlPlane: NodeConfig{
					Nodes: []string{"192.168.1.10"},
					SSH: SSHConfig{
						User:    "ec2-user",
						KeyPath: keyFile,
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestParseConfig tests the ParseConfig function.
func TestParseConfig(t *testing.T) {
	// Setup SSH key once for all tests
	keyFile := "/tmp/test-key"
	cleanup := setupSSHKeyForTest(t, keyFile)
	defer cleanup()

	tests := []struct {
		name        string
		configYaml  string
		expectError bool
	}{
		{
			name: "valid config with both etcd and control plane",
			configYaml: `
clusterName: test-cluster
os: ubuntu
controlPlane:
  nodes:
  - 192.168.1.10
  ssh:
    sshUser: ec2-user
    sshKey: /tmp/test-key
etcd:
  nodes:
  - 192.168.1.20
  ssh:
    sshUser: ec2-user
    sshKey: /tmp/test-key
`,
			expectError: false,
		},
		{
			name: "valid config without etcd (embedded)",
			configYaml: `
clusterName: test-cluster
os: ubuntu
controlPlane:
  nodes:
  - 192.168.1.10
  ssh:
    sshUser: ec2-user
    sshKey: /tmp/test-key
`,
			expectError: false,
		},
		{
			name: "invalid config - missing cluster name",
			configYaml: `
os: ubuntu
controlPlane:
  nodes:
  - 192.168.1.10
  ssh:
    sshUser: ec2-user
    sshKey: /tmp/test-key
`,
			expectError: true,
		},
		{
			name: "invalid config - unsupported OS",
			configYaml: `
clusterName: test-cluster
os: windows
controlPlane:
  nodes:
  - 192.168.1.10
  ssh:
    sshUser: ec2-user
    sshKey: /tmp/test-key
`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpfile, err := os.CreateTemp("", "config-*.yaml")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.configYaml)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			// Test config parsing
			_, err = ParseConfig(tmpfile.Name())
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
