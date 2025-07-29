package certificates_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	gomock "github.com/golang/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/certificates"
	kubemocks "github.com/aws/eks-anywhere/pkg/clients/kubernetes/mocks"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	clusterLabel      = "demo"
	cpIP              = "10.0.0.1"
	etcdIP            = "10.0.0.2"
	namespace         = constants.EksaSystemNamespace
	clusterNameLabel  = "cluster.x-k8s.io/cluster-name"
	controlPlaneLabel = "cluster.x-k8s.io/control-plane"
	externalEtcdLabel = "cluster.x-k8s.io/etcd-cluster"
)

// helper to generate a Machine with the given labels and external IP.
func buildMachine(labels map[string]string, ip string) clusterv1.Machine {
	return clusterv1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Status: clusterv1.MachineStatus{
			Addresses: []clusterv1.MachineAddress{
				{Type: clusterv1.MachineExternalIP, Address: ip},
			},
		},
	}
}

// setupSSHKeyForTest creates a temporary SSH key file for testing.
func setupSSHKeyForTest(t *testing.T, path string) func() {
	t.Helper()
	if err := os.WriteFile(path, []byte("test-key"), 0o600); err != nil {
		t.Fatalf("setupSSHKeyForTest() failed to create key file: %v", err)
	}
	return func() { os.Remove(path) }
}

func createConfigFileFromYAML(t *testing.T, yamlContent string) (string, func()) {
	t.Helper()

	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("create tmp: %v", err)
	}

	if _, err := tmpfile.Write([]byte(yamlContent)); err != nil {
		tmpfile.Close()
		t.Fatalf("write tmp: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("close tmp: %v", err)
	}

	return tmpfile.Name(), func() { os.Remove(tmpfile.Name()) }
}

// TestParseConfigFileNotFound tests the ParseConfig function with a non-existent file.
func TestParseConfigFileNotFound(t *testing.T) {
	_, err := certificates.ParseConfig("non-existent-file.yaml")
	if err == nil {
		t.Error("expected error for non-existent file but got none")
	}
}

func TestParseConfig_InvalidYAML(t *testing.T) {
	bad := "clusterName: foo: bar"
	file, cleanup := createConfigFileFromYAML(t, bad)
	defer cleanup()

	if _, err := certificates.ParseConfig(file); err == nil {
		t.Fatalf("ParseConfig(): want YAML error, got %v", err)
	}
}

func TestParseConfig_EnvPasswordsInjected(t *testing.T) {
	key := "/tmp/test-key-pass"
	if err := os.WriteFile(key, []byte("k"), 0o600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}
	defer os.Remove(key)

	yml := fmt.Sprintf(
		"clusterName: demo\n"+
			"os: ubuntu\n"+
			"controlPlane:\n"+
			"  nodes:\n"+
			"    - 1.2.3.4\n"+
			"  ssh:\n"+
			"    sshUser: u\n"+
			"    sshKey: %s\n"+
			"etcd:\n"+
			"  nodes:\n"+
			"    - 5.6.7.8\n"+
			"  ssh:\n"+
			"    sshUser: u\n"+
			"    sshKey: %s\n",
		key, key)

	os.Setenv("EKSA_SSH_KEY_PASSPHRASE_CP", "pass-cp")
	os.Setenv("EKSA_SSH_KEY_PASSPHRASE_ETCD", "pass-etcd")
	defer func() {
		os.Unsetenv("EKSA_SSH_KEY_PASSPHRASE_CP")
		os.Unsetenv("EKSA_SSH_KEY_PASSPHRASE_ETCD")
	}()

	file, cleanup := createConfigFileFromYAML(t, yml)
	defer cleanup()

	if cfg, err := certificates.ParseConfig(file); err != nil {
		t.Fatalf("unexpected: %v", err)
	} else if cfg.ControlPlane.SSH.Password != "pass-cp" || cfg.Etcd.SSH.Password != "pass-etcd" {
		t.Fatalf("env passphrase not injected into cfg: %#v", cfg)
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
		component   string
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
			component:   "",
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
			component:   "",
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
			component:   "",
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
			component:   "",
			expectError: true,
		},
		{
			name: "invalid component - etcd with no etcd nodes",
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
			component:   "etcd",
			expectError: true,
		},
		{
			name: "invalid component - unknown component",
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
			component:   "unknown",
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

			config, err := certificates.ParseConfig(tmpfile.Name())
			if err != nil && !tt.expectError {
				t.Errorf("unexpected error parsing config: %v", err)
				return
			}

			if err == nil {
				err = certificates.ValidateConfig(config, tt.component)
				if tt.expectError && err == nil {
					t.Error("expected validation error but got none")
				}
				if !tt.expectError && err != nil {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

// Testcertificates.ValidateConfig tests the certificates.ValidateConfig function directly.
func TestValidateConfig(t *testing.T) {
	// Setup SSH key once for all tests
	keyFile := "/tmp/test-key"
	cleanup := setupSSHKeyForTest(t, keyFile)
	defer cleanup()

	tests := []struct {
		name        string
		config      *certificates.RenewalConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &certificates.RenewalConfig{
				ClusterName: "test-cluster",
				OS:          "ubuntu",
				ControlPlane: certificates.NodeConfig{
					Nodes: []string{"192.168.1.10"},
					SSH: certificates.SSHConfig{
						User:    "ec2-user",
						KeyPath: keyFile,
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing cluster name",
			config: &certificates.RenewalConfig{
				OS: "ubuntu",
				ControlPlane: certificates.NodeConfig{
					Nodes: []string{"192.168.1.10"},
					SSH: certificates.SSHConfig{
						User:    "ec2-user",
						KeyPath: keyFile,
					},
				},
			},
			expectError: true,
		},
		{
			name: "missing control plane nodes",
			config: &certificates.RenewalConfig{
				ClusterName: "test-cluster",
				OS:          "ubuntu",
				ControlPlane: certificates.NodeConfig{
					SSH: certificates.SSHConfig{
						User:    "ec2-user",
						KeyPath: keyFile,
					},
				},
			},
			expectError: true,
		},
		{
			name: "non-existent SSH key file",
			config: &certificates.RenewalConfig{
				ClusterName: "test-cluster",
				OS:          "ubuntu",
				ControlPlane: certificates.NodeConfig{
					Nodes: []string{"192.168.1.10"},
					SSH: certificates.SSHConfig{
						User:    "ec2-user",
						KeyPath: "/tmp/non-existent-key",
					},
				},
			},
			expectError: true,
		},
		{
			name: "unsupported OS",
			config: &certificates.RenewalConfig{
				ClusterName: "test-cluster",
				OS:          "windows",
				ControlPlane: certificates.NodeConfig{
					Nodes: []string{"192.168.1.10"},
					SSH: certificates.SSHConfig{
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
			err := certificates.ValidateConfig(tt.config, "")
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateConfig_MissingOS(t *testing.T) {
	key := "/tmp/key-missing-os"
	if err := os.WriteFile(key, []byte("k"), 0o600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}
	defer os.Remove(key)

	if err := certificates.ValidateConfig(&certificates.RenewalConfig{
		ClusterName:  "c",
		ControlPlane: certificates.NodeConfig{Nodes: []string{"n"}, SSH: certificates.SSHConfig{User: "u", KeyPath: key}},
	}, ""); err == nil {
		t.Fatalf("want missing os error, got %v", err)
	}
}

func TestValidateConfig_MissingSSHUser(t *testing.T) {
	key := "/tmp/key-no-user"
	if err := os.WriteFile(key, []byte("k"), 0o600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}
	defer os.Remove(key)

	cfg := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(v1alpha1.Ubuntu),
		ControlPlane: certificates.NodeConfig{
			Nodes: []string{"1.1.1.1"},
			SSH:   certificates.SSHConfig{KeyPath: key},
		},
	}

	if err := certificates.ValidateConfig(cfg, ""); err == nil {
		t.Fatalf("want sshUser required error, got nil")
	}
}

func TestValidateConfig_MissingKeyPath(t *testing.T) {
	cfg := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(v1alpha1.Ubuntu),
		ControlPlane: certificates.NodeConfig{
			Nodes: []string{"1.1.1.1"},
			SSH:   certificates.SSHConfig{User: "test"},
		},
	}

	if err := certificates.ValidateConfig(cfg, ""); err == nil {
		t.Fatalf("want sshKey required error, got nil")
	}
}

func TestValidateConfig_EtcdMissingSSHUser(t *testing.T) {
	key := "/tmp/key-bad-etcd"
	if err := os.WriteFile(key, []byte("k"), 0o600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}
	defer os.Remove(key)

	cfg := &certificates.RenewalConfig{
		ClusterName: "c", OS: "ubuntu",
		ControlPlane: certificates.NodeConfig{Nodes: []string{"n"}, SSH: certificates.SSHConfig{User: "u", KeyPath: key}},
		Etcd:         certificates.NodeConfig{Nodes: []string{"e"}, SSH: certificates.SSHConfig{KeyPath: key}},
	}
	if err := certificates.ValidateConfig(cfg, ""); err == nil {
		t.Fatalf("want nested etcd validation error, got %v", err)
	}
}

func TestValidateComponentWithConfig_ValidControlPlane(t *testing.T) {
	cfg := &certificates.RenewalConfig{
		ClusterName:  "test",
		OS:           "ubuntu",
		ControlPlane: certificates.NodeConfig{Nodes: []string{"1.1.1.1"}},
	}

	if err := certificates.ValidateComponentWithConfig("control-plane", cfg); err != nil {
		t.Fatalf("ValidateComponentWithConfig() expected no error for control-plane component, got: %v", err)
	}
}

func TestValidateComponentWithConfig_ValidEtcdWithNodes(t *testing.T) {
	cfg := &certificates.RenewalConfig{
		ClusterName:  "test",
		OS:           "ubuntu",
		ControlPlane: certificates.NodeConfig{Nodes: []string{"1.1.1.1"}},
		Etcd:         certificates.NodeConfig{Nodes: []string{"2.2.2.2"}},
	}

	if err := certificates.ValidateComponentWithConfig("etcd", cfg); err != nil {
		t.Fatalf("ValidateComponentWithConfig() expected no error for etcd component with nodes, got: %v", err)
	}
}

func TestValidateComponentWithConfig_EmptyComponent(t *testing.T) {
	cfg := &certificates.RenewalConfig{
		ClusterName:  "test",
		OS:           "ubuntu",
		ControlPlane: certificates.NodeConfig{Nodes: []string{"1.1.1.1"}},
	}

	if err := certificates.ValidateComponentWithConfig("", cfg); err != nil {
		t.Fatalf("ValidateComponentWithConfig() expected no error for empty component, got: %v", err)
	}
}

func TestPopulateConfig_ControlPlaneSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	k := kubemocks.NewMockClient(ctrl)

	machines := &clusterv1.MachineList{
		Items: []clusterv1.Machine{
			buildMachine(map[string]string{
				clusterNameLabel:  clusterLabel,
				controlPlaneLabel: "",
			}, cpIP),
		},
	}

	k.EXPECT().
		List(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, l any, _ ...any) error {
			*l.(*clusterv1.MachineList) = *machines
			return nil
		})

	k.EXPECT().
		List(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, l any, _ ...any) error {
			*l.(*clusterv1.MachineList) = *machines
			return nil
		})

	cfg := &certificates.RenewalConfig{
		ClusterName: clusterLabel,
		OS:          string(v1alpha1.Ubuntu),
		ControlPlane: certificates.NodeConfig{
			SSH: certificates.SSHConfig{User: "ec2-user", KeyPath: "/test"},
		},
	}

	if err := certificates.PopulateConfig(context.Background(), cfg, k, &types.Cluster{Name: clusterLabel}); err != nil {
		t.Fatalf("PopulateConfig() expected no error, got: %v", err)
	}
}

func TestPopulateConfig_EtcdSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	k := kubemocks.NewMockClient(ctrl)

	cpMachines := &clusterv1.MachineList{
		Items: []clusterv1.Machine{},
	}

	etcdMachines := &clusterv1.MachineList{
		Items: []clusterv1.Machine{
			buildMachine(map[string]string{
				clusterNameLabel:  clusterLabel,
				externalEtcdLabel: clusterLabel + "-etcd",
			}, etcdIP),
		},
	}

	k.EXPECT().
		List(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, l any, _ ...any) error {
			*l.(*clusterv1.MachineList) = *cpMachines
			return nil
		})

	k.EXPECT().
		List(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, l any, _ ...any) error {
			*l.(*clusterv1.MachineList) = *etcdMachines
			return nil
		})

	cfg := &certificates.RenewalConfig{
		ClusterName: clusterLabel,
		OS:          string(v1alpha1.Ubuntu),
		ControlPlane: certificates.NodeConfig{
			SSH: certificates.SSHConfig{User: "ec2-user", KeyPath: "/test"},
		},
	}

	if err := certificates.PopulateConfig(context.Background(), cfg, k, &types.Cluster{Name: clusterLabel}); err != nil {
		t.Fatalf("PopulateConfig() expected no error, got: %v", err)
	}
}

func TestPopulateConfig_EtcdListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	k := kubemocks.NewMockClient(ctrl)

	cpMachines := &clusterv1.MachineList{
		Items: []clusterv1.Machine{
			buildMachine(map[string]string{
				clusterNameLabel:  clusterLabel,
				controlPlaneLabel: "",
			}, cpIP),
		},
	}

	k.EXPECT().
		List(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, l interface{}, _ ...interface{}) error {
			*l.(*clusterv1.MachineList) = *cpMachines
			return nil
		})

	k.EXPECT().
		List(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("etcd list error"))

	cfg := &certificates.RenewalConfig{
		ClusterName: clusterLabel,
		OS:          string(v1alpha1.Ubuntu),
		ControlPlane: certificates.NodeConfig{
			SSH: certificates.SSHConfig{User: "ec2-user", KeyPath: "/test"},
		},
	}

	if err := certificates.PopulateConfig(context.Background(), cfg, k, &types.Cluster{Name: clusterLabel}); err == nil {
		t.Fatalf("PopulateConfig() expected error, got nil")
	}
}

func TestPopulateConfig_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	k := kubemocks.NewMockClient(ctrl)

	all := &clusterv1.MachineList{
		Items: []clusterv1.Machine{
			buildMachine(map[string]string{
				clusterNameLabel:  clusterLabel,
				controlPlaneLabel: "",
			}, cpIP),
			buildMachine(map[string]string{
				clusterNameLabel:  clusterLabel,
				externalEtcdLabel: clusterLabel + "-etcd",
			}, etcdIP),
		},
	}

	k.EXPECT().
		List(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, l interface{}, _ ...interface{}) error {
			*l.(*clusterv1.MachineList) = *all
			return nil
		}).Times(2)

	cfg := &certificates.RenewalConfig{
		ClusterName: clusterLabel,
		OS:          string(v1alpha1.Ubuntu),
		ControlPlane: certificates.NodeConfig{
			SSH: certificates.SSHConfig{User: "ec2-user", KeyPath: "/test"},
		},
	}

	if err := certificates.PopulateConfig(context.Background(), cfg, k, &types.Cluster{Name: clusterLabel}); err != nil {
		t.Fatalf("PopulateConfig() expected no error, got: %v", err)
	}
}

func TestPopulateConfig_ControlPlaneListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	k := kubemocks.NewMockClient(ctrl)

	k.EXPECT().
		List(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("api down"))

	cfg := &certificates.RenewalConfig{
		ClusterName: clusterLabel,
		OS:          string(v1alpha1.Ubuntu),
		ControlPlane: certificates.NodeConfig{
			SSH: certificates.SSHConfig{User: "ec2-user", KeyPath: "/test"},
		},
	}

	if err := certificates.PopulateConfig(context.Background(), cfg, k, &types.Cluster{Name: clusterLabel}); err == nil {
		t.Fatalf("PopulateConfig() expected error, got nil")
	}
}

func TestPopulateConfig_ExistingNodesEarlyReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	k := kubemocks.NewMockClient(ctrl)

	cfg := &certificates.RenewalConfig{
		ClusterName: clusterLabel,
		OS:          string(v1alpha1.Ubuntu),
		ControlPlane: certificates.NodeConfig{
			Nodes: []string{cpIP},
			SSH:   certificates.SSHConfig{User: "ec2-user", KeyPath: "/test"},
		},
	}

	if err := certificates.PopulateConfig(context.Background(), cfg, k, &types.Cluster{Name: clusterLabel}); err != nil {
		t.Fatalf("PopulateConfig() expected no error for early return, got: %v", err)
	}
}
