package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/certificates"
	"github.com/aws/eks-anywhere/pkg/constants"
)

// Test YAML configurations.
var (
	validConfigYaml = `
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
`

	validConfigYamlNoEtcd = `
clusterName: test-cluster
os: ubuntu
controlPlane:
  nodes:
  - 192.168.1.10
  ssh:
    sshUser: ec2-user
    sshKey: /tmp/test-key
`

	invalidConfigYaml = `
invalid: yaml: :
  - this is not valid yaml
`
)

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

// TestValidateComponent tests the validateComponent function.
func TestValidateComponent(t *testing.T) {
	testConfig := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          "ubuntu",
		ControlPlane: certificates.NodeConfig{
			Nodes: []string{"192.168.1.10"},
			SSH: certificates.SSHConfig{
				User:    "ec2-user",
				KeyPath: "/tmp/test-key",
			},
		},
		Etcd: certificates.NodeConfig{
			Nodes: []string{"192.168.1.20"},
			SSH: certificates.SSHConfig{
				User:    "ec2-user",
				KeyPath: "/tmp/test-key",
			},
		},
	}

	tests := []struct {
		name        string
		component   string
		config      *certificates.RenewalConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid etcd component",
			component:   constants.EtcdComponent,
			config:      testConfig,
			expectError: false,
		},
		{
			name:        "valid control-plane component",
			component:   constants.ControlPlaneComponent,
			config:      testConfig,
			expectError: false,
		},
		{
			name:        "empty component",
			component:   "",
			config:      testConfig,
			expectError: false,
		},
		{
			name:        "invalid component",
			component:   "invalid",
			config:      testConfig,
			expectError: true,
			errorMsg:    "invalid component",
		},
		{
			name:      "etcd component with no etcd nodes",
			component: constants.EtcdComponent,
			config: &certificates.RenewalConfig{
				ClusterName: "test-cluster",
				OS:          "ubuntu",
				ControlPlane: certificates.NodeConfig{
					Nodes: []string{"192.168.1.10"},
					SSH: certificates.SSHConfig{
						User:    "ec2-user",
						KeyPath: "/tmp/test-key",
					},
				},
				Etcd: certificates.NodeConfig{},
			},
			expectError: true,
			errorMsg:    "no external etcd nodes defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := certificates.ValidateComponentWithConfig(tt.component, tt.config)
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
			errorMsg:    "reading config file",
		},
		{
			name:        "invalid config yaml",
			configYaml:  invalidConfigYaml,
			expectError: true,
			errorMsg:    "parsing config file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configFile string
			var fileCleanup func()

			if tt.configYaml != "" {
				configFile, fileCleanup = createConfigFileFromYAML(t, tt.configYaml)
				defer fileCleanup()
			} else {
				configFile = tt.configFile
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

// testRenewCertificates is a test-only version that skips dependency initialization.
func testRenewCertificates(configFile, component string) error {
	cfg, err := certificates.ParseConfig(configFile)
	if err != nil {
		return err
	}
	return certificates.ValidateComponentWithConfig(component, cfg)
}

// TestRenewCertificates tests the renewCertificates method.
func TestRenewCertificates(t *testing.T) {
	// Setup SSH key file once for all tests
	keyFile := "/tmp/test-key"
	if err := os.WriteFile(keyFile, []byte("test-key"), 0o600); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(keyFile)

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
			configYaml:  validConfigYaml,
		},
		{
			name:        "valid config with control-plane component",
			component:   constants.ControlPlaneComponent,
			expectError: false,
			configYaml:  validConfigYaml,
		},
		{
			name:        "valid config with empty component",
			component:   "",
			expectError: false,
			configYaml:  validConfigYaml,
		},
		{
			name:        "invalid component",
			component:   "invalid",
			expectError: true,
			errorMsg:    "invalid component",
			configYaml:  validConfigYamlNoEtcd,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare config file
			configFile, fileCleanup := createConfigFileFromYAML(t, tt.configYaml)
			defer fileCleanup()

			// Use test-specific function instead of actual method
			err := testRenewCertificates(configFile, tt.component)

			// Check for expected errors
			checkTestError(t, err, tt.expectError, tt.errorMsg)
		})
	}
}
