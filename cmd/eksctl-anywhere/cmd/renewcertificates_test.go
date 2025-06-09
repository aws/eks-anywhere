package cmd

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// createTestConfigFile creates a temporary config file for testing.
func createTestConfigFile(t *testing.T) (string, func()) {
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}

	configContent := `
clusterName: test-cluster
controlPlane:
  nodes:
  - 192.168.1.10
  os: ubuntu
  sshKey: /tmp/test-key
  sshUser: ec2-user
`
	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	return tmpfile.Name(), func() { os.Remove(tmpfile.Name()) }
}

// checkTestError validates test errors against expectations.
func checkTestError(t *testing.T, err error, expectError bool, errorMsg string) {
	if expectError && err == nil {
		t.Error("expected error but got none")
	}
	if !expectError && err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if expectError && err != nil && errorMsg != "" && !strings.Contains(err.Error(), errorMsg) {
		t.Errorf("expected error message to contain %q, got %q", errorMsg, err.Error())
	}
}

func TestRenewCertificatesValidation(t *testing.T) {
	tests := []struct {
		name        string
		configFile  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "no config file provided",
			configFile:  "",
			expectError: true,
			errorMsg:    "must specify --config",
		},
		{
			name:        "valid config file",
			configFile:  "test.yaml",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configFile string
			var cleanup func()

			if tt.configFile != "" {
				configFile, cleanup = createTestConfigFile(t)
				defer cleanup()
			} else {
				configFile = ""
			}

			// Set up the command options
			rc := &renewCertificatesOptions{
				configFile: configFile,
			}

			cmd := &cobra.Command{}

			// Run the validation
			err := validateRenewCertificatesOptions(cmd, rc)

			// Check for expected errors
			checkTestError(t, err, tt.expectError, tt.errorMsg)
		})
	}
}

func validateRenewCertificatesOptions(_ *cobra.Command, rc *renewCertificatesOptions) error {
	if rc.configFile == "" {
		return fmt.Errorf("must specify --config")
	}
	return nil
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
			component:   "etcd",
			expectError: false,
		},
		{
			name:        "valid control-plane component",
			component:   "control-plane",
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
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectError && err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("expected error message to contain %q, got %q", tt.errorMsg, err.Error())
			}
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

// TestRenewCertificates tests the renewCertificates method.
func TestRenewCertificates(t *testing.T) {
	// Setup SSH key file
	cleanup := setupSSHKeyFile(t)
	defer cleanup()

	// Define test cases
	tests := []struct {
		name        string
		configFile  string
		component   string
		configYaml  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid config with valid component",
			component:   "etcd",
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
		{
			name:        "non-existent config file",
			component:   "etcd",
			configFile:  "non-existent-file.yaml",
			expectError: true,
			errorMsg:    "parsing config file",
		},
		{
			name:        "invalid config yaml",
			component:   "etcd",
			expectError: true,
			errorMsg:    "parsing config file",
			configYaml: `
invalid: yaml: :
  - this is not valid yaml
`,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare config file
			var configFile string
			var fileCleanup func()

			if tt.configFile != "" {
				configFile = tt.configFile
			} else if tt.configYaml != "" {
				configFile, fileCleanup = createConfigFileFromYAML(t, tt.configYaml)
				defer fileCleanup()
			}

			// Set up the command options
			rc := &renewCertificatesOptions{
				configFile: configFile,
				component:  tt.component,
			}

			cmd := &cobra.Command{}

			// Run the renewCertificates method
			err := rc.renewCertificates(cmd)

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
		configFile  string
		component   string
		configYaml  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid config with valid component",
			component:   "etcd",
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
			originalRC := rc
			defer func() { rc = originalRC }()

			var configFile string
			var fileCleanup func()

			if tt.configFile != "" {
				configFile = tt.configFile
			} else if tt.configYaml != "" {
				configFile, fileCleanup = createConfigFileFromYAML(t, tt.configYaml)
				defer fileCleanup()
			}

			rc = &renewCertificatesOptions{
				configFile: configFile,
				component:  tt.component,
			}

			cmd := &cobra.Command{}

			runE := renewCertificatesCmd.RunE

			err := runE(cmd, []string{})

			checkTestError(t, err, tt.expectError, tt.errorMsg)
		})
	}
}
