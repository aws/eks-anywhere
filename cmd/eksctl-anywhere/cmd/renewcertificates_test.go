package cmd

import (
	"os"
	"testing"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/constants"
)

// checkTestError validates test errors against expectations.
func checkTestError(t *testing.T, err error, expectError bool, errorMsg string) {
	if expectError && err == nil {
		t.Error("expected error but got none")
	}
	if !expectError && err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if expectError && err != nil && errorMsg != "" && !contains(err.Error(), errorMsg) {
		t.Errorf("expected error message to contain %q, got %q", errorMsg, err.Error())
	}
}

// contains checks if a string contains another string.
func contains(s, substr string) bool {
	if s == "" || substr == "" {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestValidateComponent tests the validateComponent function.
func TestValidateComponent(t *testing.T) {
	tests := []struct {
		name        string
		component   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid etcd component",
			component:   constants.EtcdComponent,
			expectError: false,
		},
		{
			name:        "valid control-plane component",
			component:   constants.ControlPlaneComponent,
			expectError: false,
		},
		{
			name:        "empty component",
			component:   "",
			expectError: false,
		},
		{
			name:        "invalid component",
			component:   "invalid",
			expectError: true,
			errorMsg:    "invalid component",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateComponent(tt.component)
			checkTestError(t, err, tt.expectError, tt.errorMsg)
		})
	}
}

// createConfigFileFromYAML creates a temporary config file with the given YAML content.
func createConfigFileFromYAML(t *testing.T, yamlContent string) (string, func()) {
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := tmpfile.Write([]byte(yamlContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	return tmpfile.Name(), func() { os.Remove(tmpfile.Name()) }
}

// setupSSHKeyFile creates a temporary SSH key file for testing.
func setupSSHKeyFile(t *testing.T) func() {
	keyFile := "/tmp/test-key"
	if err := os.WriteFile(keyFile, []byte("test-key"), 0o600); err != nil {
		t.Fatal(err)
	}
	return func() { os.Remove(keyFile) }
}

// TestConfigFileValidation tests the config file validation separately.
func TestConfigFileValidation(t *testing.T) {
	tests := []struct {
		name        string
		configFile  string
		configYaml  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "non-existent config file",
			configFile:  "non-existent-file.yaml",
			expectError: true,
			errorMsg:    "parsing config file",
		},
		{
			name: "invalid config yaml",
			configYaml: `
invalid: yaml: :
  - this is not valid yaml
`,
			expectError: true,
			errorMsg:    "parsing config file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configFile string
			var fileCleanup func()

			if tt.configFile != "" {
				configFile = tt.configFile
			} else if tt.configYaml != "" {
				configFile, fileCleanup = createConfigFileFromYAML(t, tt.configYaml)
				defer fileCleanup()
			}

			rc := &renewCertificatesOptions{
				configFile: configFile,
			}

			cmd := &cobra.Command{}
			err := rc.renewCertificates(cmd, []string{})
			checkTestError(t, err, tt.expectError, tt.errorMsg)
		})
	}
}

// TestRenewCertificates tests the renewCertificates method.
func TestRenewCertificates(t *testing.T) {
	// Setup SSH key file once for all tests
	cleanup := setupSSHKeyFile(t)
	defer cleanup()

	// Define test cases
	tests := []struct {
		name        string
		component   string
		configYaml  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid config with etcd component",
			component:   constants.EtcdComponent,
			expectError: false,
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
		},
		{
			name:        "valid config with control-plane component",
			component:   constants.ControlPlaneComponent,
			expectError: false,
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
		},
		{
			name:        "valid config with empty component",
			component:   "",
			expectError: false,
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
		},
		{
			name:        "invalid component",
			component:   "invalid",
			expectError: true,
			errorMsg:    "invalid component",
			configYaml: `
clusterName: test-cluster
controlPlane:
  nodes:
  - 192.168.1.10
  os: ubuntu
  sshKey: /tmp/test-key
  sshUser: ec2-user
`,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare config file
			configFile, fileCleanup := createConfigFileFromYAML(t, tt.configYaml)
			defer fileCleanup()

			// Set up the command options
			rc := &renewCertificatesOptions{
				configFile: configFile,
				component:  tt.component,
			}

			cmd := &cobra.Command{}

			// Run the renewCertificates method
			err := rc.renewCertificates(cmd, []string{})

			// Check for expected errors
			checkTestError(t, err, tt.expectError, tt.errorMsg)
		})
	}
}

// TestRenewCertificatesRunE tests the RunE function of the renewCertificatesCmd.
func TestRenewCertificatesRunE(t *testing.T) {
	// Setup SSH key file
	cleanup := setupSSHKeyFile(t)
	defer cleanup()

	// Define test cases
	tests := []struct {
		name        string
		component   string
		configYaml  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid config with valid component",
			component:   constants.EtcdComponent,
			expectError: false,
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
		},
		{
			name:        "invalid component",
			component:   "invalid",
			expectError: true,
			errorMsg:    "invalid component",
			configYaml: `
clusterName: test-cluster
controlPlane:
  nodes:
  - 192.168.1.10
  os: ubuntu
  sshKey: /tmp/test-key
  sshUser: ec2-user
`,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, fileCleanup := createConfigFileFromYAML(t, tt.configYaml)
			defer fileCleanup()

			testRC := &renewCertificatesOptions{
				configFile: configFile,
				component:  tt.component,
			}

			cmd := &cobra.Command{}
			err := testRC.renewCertificates(cmd, []string{})
			checkTestError(t, err, tt.expectError, tt.errorMsg)
		})
	}
}
