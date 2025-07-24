package certificates

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
	kubemocks "github.com/aws/eks-anywhere/pkg/clients/kubernetes/mocks"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	clusterLabel = "demo"
	cpIP         = "10.0.0.1"
	etcdIP       = "10.0.0.2"
	namespace    = constants.EksaSystemNamespace
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
			err := ValidateConfig(tt.config, "")
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

			config, err := ParseConfig(tmpfile.Name())
			if err != nil && !tt.expectError {
				t.Errorf("unexpected error parsing config: %v", err)
				return
			}

			if err == nil {
				err = ValidateConfig(config, tt.component)
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

func Test_getControlPlaneIPs_Success(t *testing.T) {
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
		DoAndReturn(func(_ context.Context, l interface{}, _ ...interface{}) error {
			*l.(*clusterv1.MachineList) = *machines
			return nil
		})

	if got, err := getControlPlaneIPs(context.Background(), k, &types.Cluster{Name: clusterLabel}); err != nil {
		t.Fatalf("getControlPlaneIPs() expected no error, got: %v", err)
	} else if len(got) != 1 || got[0] != cpIP {
		t.Fatalf("getControlPlaneIPs() expected [%s], got: %v", cpIP, got)
	}
}

func Test_getEtcdIPs_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	k := kubemocks.NewMockClient(ctrl)

	machines := &clusterv1.MachineList{
		Items: []clusterv1.Machine{
			buildMachine(map[string]string{
				clusterNameLabel:  clusterLabel,
				externalEtcdLabel: clusterLabel + "-etcd",
			}, etcdIP),
		},
	}

	k.EXPECT().
		List(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, l interface{}, _ ...interface{}) error {
			*l.(*clusterv1.MachineList) = *machines
			return nil
		})

	if got, err := getEtcdIPs(context.Background(), k, &types.Cluster{Name: clusterLabel}); err != nil {
		t.Fatalf("getEtcdIPs() expected no error, got: %v", err)
	} else if len(got) != 1 || got[0] != etcdIP {
		t.Fatalf("getEtcdIPs() expected [%s], got: %v", etcdIP, got)
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

	cfg := &RenewalConfig{
		ClusterName: clusterLabel,
		OS:          string(v1alpha1.Ubuntu),
		ControlPlane: NodeConfig{
			SSH: SSHConfig{User: "ec2-user", KeyPath: "/test"},
		},
	}

	if err := PopulateConfig(context.Background(), cfg, k, &types.Cluster{Name: clusterLabel}); err == nil {
		t.Fatalf("PopulateConfig() expected error, got nil")
	}
}

func Test_getControlPlaneIPs_ListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	k := kubemocks.NewMockClient(ctrl)

	k.EXPECT().
		List(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("api server unavailable"))

	if _, err := getControlPlaneIPs(context.Background(), k, &types.Cluster{Name: clusterLabel}); err == nil {
		t.Fatalf("getControlPlaneIPs() expected error, got nil")
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

	cfg := &RenewalConfig{
		ClusterName: clusterLabel,
		OS:          string(v1alpha1.Ubuntu),
		ControlPlane: NodeConfig{
			SSH: SSHConfig{User: "ec2-user", KeyPath: "/test"},
		},
	}

	if err := PopulateConfig(context.Background(), cfg, k, &types.Cluster{Name: clusterLabel}); err != nil {
		t.Fatalf("PopulateConfig() expected no error, got: %v", err)
	}
	if len(cfg.ControlPlane.Nodes) != 1 || cfg.ControlPlane.Nodes[0] != cpIP {
		t.Fatalf("PopulateConfig() expected ControlPlane.Nodes=[%s], got: %v", cpIP, cfg.ControlPlane.Nodes)
	}
	if len(cfg.Etcd.Nodes) != 1 || cfg.Etcd.Nodes[0] != etcdIP {
		t.Fatalf("PopulateConfig() expected Etcd.Nodes=[%s], got: %v", etcdIP, cfg.Etcd.Nodes)
	}
}

func TestPopulateConfig_ControlPlaneListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	k := kubemocks.NewMockClient(ctrl)

	k.EXPECT().
		List(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("api down"))

	cfg := &RenewalConfig{
		ClusterName: clusterLabel,
		OS:          string(v1alpha1.Ubuntu),
		ControlPlane: NodeConfig{
			SSH: SSHConfig{User: "ec2-user", KeyPath: "/test"},
		},
	}

	if err := PopulateConfig(context.Background(), cfg, k, &types.Cluster{Name: clusterLabel}); err == nil {
		t.Fatalf("PopulateConfig() expected error, got nil")
	}
}

func TestParseConfig_InvalidYAML(t *testing.T) {
	bad := "clusterName: foo: bar"
	file, cleanup := createConfigFileFromYAML(t, bad)
	defer cleanup()

	if _, err := ParseConfig(file); err == nil {
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

	if cfg, err := ParseConfig(file); err != nil {
		t.Fatalf("unexpected: %v", err)
	} else if cfg.ControlPlane.SSH.Password != "pass-cp" || cfg.Etcd.SSH.Password != "pass-etcd" {
		t.Fatalf("env passphrase not injected into cfg: %#v", cfg)
	}
}

func TestValidateConfig_MissingOS(t *testing.T) {
	key := "/tmp/key-missing-os"
	if err := os.WriteFile(key, []byte("k"), 0o600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}
	defer os.Remove(key)

	if err := ValidateConfig(&RenewalConfig{
		ClusterName:  "c",
		ControlPlane: NodeConfig{Nodes: []string{"n"}, SSH: SSHConfig{User: "u", KeyPath: key}},
	}, ""); err == nil {
		t.Fatalf("want missing os error, got %v", err)
	}
}

func TestValidateConfig_EtcdMissingSSHUser(t *testing.T) {
	key := "/tmp/key-bad-etcd"
	if err := os.WriteFile(key, []byte("k"), 0o600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}
	defer os.Remove(key)

	cfg := &RenewalConfig{
		ClusterName: "c", OS: "ubuntu",
		ControlPlane: NodeConfig{Nodes: []string{"n"}, SSH: SSHConfig{User: "u", KeyPath: key}},
		Etcd:         NodeConfig{Nodes: []string{"e"}, SSH: SSHConfig{KeyPath: key}},
	}
	if err := ValidateConfig(cfg, ""); err == nil {
		t.Fatalf("want nested etcd validation error, got %v", err)
	}
}

func TestValidateNodeConfig_MissingSSHUser(t *testing.T) {
	key := "/tmp/key-no-user"
	if err := os.WriteFile(key, []byte("k"), 0o600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}
	defer os.Remove(key)

	nc := &NodeConfig{
		Nodes: []string{"1.1.1.1"},
		SSH:   SSHConfig{KeyPath: key},
	}
	if err := validateNodeConfig(nc); err == nil {
		t.Fatalf("want sshUser required error, got %v", err)
	}
}

func TestValidateNodeConfig_MissingKeyPath(t *testing.T) {
	nc := &NodeConfig{
		Nodes: []string{"1.1.1.1"},
		SSH:   SSHConfig{User: "test"},
	}
	if err := validateNodeConfig(nc); err == nil {
		t.Fatalf("want sshKey required error, got %v", err)
	}
}

func TestValidateComponentWithConfig_ValidControlPlane(t *testing.T) {
	cfg := &RenewalConfig{
		ClusterName:  "test",
		OS:           "ubuntu",
		ControlPlane: NodeConfig{Nodes: []string{"1.1.1.1"}},
	}

	if err := ValidateComponentWithConfig("control-plane", cfg); err != nil {
		t.Fatalf("ValidateComponentWithConfig() expected no error for control-plane component, got: %v", err)
	}
}

func TestValidateComponentWithConfig_ValidEtcdWithNodes(t *testing.T) {
	cfg := &RenewalConfig{
		ClusterName:  "test",
		OS:           "ubuntu",
		ControlPlane: NodeConfig{Nodes: []string{"1.1.1.1"}},
		Etcd:         NodeConfig{Nodes: []string{"2.2.2.2"}},
	}

	if err := ValidateComponentWithConfig("etcd", cfg); err != nil {
		t.Fatalf("ValidateComponentWithConfig() expected no error for etcd component with nodes, got: %v", err)
	}
}

func TestValidateComponentWithConfig_EmptyComponent(t *testing.T) {
	cfg := &RenewalConfig{
		ClusterName:  "test",
		OS:           "ubuntu",
		ControlPlane: NodeConfig{Nodes: []string{"1.1.1.1"}},
	}

	if err := ValidateComponentWithConfig("", cfg); err != nil {
		t.Fatalf("ValidateComponentWithConfig() expected no error for empty component, got: %v", err)
	}
}

func TestPopulateConfig_ExistingNodesEarlyReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	k := kubemocks.NewMockClient(ctrl)

	cfg := &RenewalConfig{
		ClusterName: clusterLabel,
		OS:          string(v1alpha1.Ubuntu),
		ControlPlane: NodeConfig{
			Nodes: []string{cpIP},
			SSH:   SSHConfig{User: "ec2-user", KeyPath: "/test"},
		},
	}

	if err := PopulateConfig(context.Background(), cfg, k, &types.Cluster{Name: clusterLabel}); err != nil {
		t.Fatalf("PopulateConfig() expected no error for early return, got: %v", err)
	}
	if len(cfg.ControlPlane.Nodes) != 1 || cfg.ControlPlane.Nodes[0] != cpIP {
		t.Fatalf("PopulateConfig() should preserve existing ControlPlane.Nodes")
	}
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
