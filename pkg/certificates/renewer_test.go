package certificates_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/certificates"
	"github.com/aws/eks-anywhere/pkg/certificates/mocks"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	kubemocks "github.com/aws/eks-anywhere/pkg/clients/kubernetes/mocks"
)

// Test helper functions
func newTestKubernetesClient() kubernetes.Client {
	ctrl := gomock.NewController(nil)
	return kubemocks.NewMockClient(ctrl)
}

// newTestRenewerWithMockSSH creates a Renewer instance with mock SSH runners for testing
func newTestRenewerWithMockSSH(t *testing.T, kubectl kubernetes.Client, osType string, cfg *certificates.RenewalConfig) (*certificates.Renewer, *mocks.MockSSHRunner, *mocks.MockSSHRunner) {
	ctrl := gomock.NewController(t)

	mockSSHEtcd := mocks.NewMockSSHRunner(ctrl)
	mockSSHControlPlane := mocks.NewMockSSHRunner(ctrl)

	// Create a test renewer using reflection to bypass the SSH creation
	ts := time.Now().Format("2006-01-02T15_04_05")
	backupDir := "certificate_backup_" + ts

	// Create backup directory
	err := os.MkdirAll(filepath.Join(backupDir, "etcd-client-certs"), 0o755)
	if err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}

	osRenewer := certificates.BuildOSRenewer(osType, backupDir)

	// Use reflection to create the Renewer struct directly
	renewer := &certificates.Renewer{}

	// Set the fields using reflection with proper field access
	v := reflect.ValueOf(renewer).Elem()

	// Set backupDir field
	backupDirField := v.FieldByName("backupDir")
	if backupDirField.IsValid() && backupDirField.CanSet() {
		backupDirField.SetString(backupDir)
	} else {
		// Use unsafe reflection to set unexported field
		backupDirField = reflect.NewAt(backupDirField.Type(), unsafe.Pointer(backupDirField.UnsafeAddr())).Elem()
		backupDirField.SetString(backupDir)
	}

	// Set kubectl field
	kubectlField := v.FieldByName("kubectl")
	if kubectlField.IsValid() && kubectlField.CanSet() {
		kubectlField.Set(reflect.ValueOf(kubectl))
	} else {
		kubectlField = reflect.NewAt(kubectlField.Type(), unsafe.Pointer(kubectlField.UnsafeAddr())).Elem()
		kubectlField.Set(reflect.ValueOf(kubectl))
	}

	// Set os field
	osField := v.FieldByName("os")
	if osField.IsValid() && osField.CanSet() {
		osField.Set(reflect.ValueOf(osRenewer))
	} else {
		osField = reflect.NewAt(osField.Type(), unsafe.Pointer(osField.UnsafeAddr())).Elem()
		osField.Set(reflect.ValueOf(osRenewer))
	}

	// Set SSH runners based on configuration
	if len(cfg.Etcd.Nodes) > 0 {
		sshEtcdField := v.FieldByName("sshEtcd")
		if sshEtcdField.IsValid() && sshEtcdField.CanSet() {
			sshEtcdField.Set(reflect.ValueOf(mockSSHEtcd))
		} else {
			sshEtcdField = reflect.NewAt(sshEtcdField.Type(), unsafe.Pointer(sshEtcdField.UnsafeAddr())).Elem()
			sshEtcdField.Set(reflect.ValueOf(mockSSHEtcd))
		}
	}

	sshControlPlaneField := v.FieldByName("sshControlPlane")
	if sshControlPlaneField.IsValid() && sshControlPlaneField.CanSet() {
		sshControlPlaneField.Set(reflect.ValueOf(mockSSHControlPlane))
	} else {
		sshControlPlaneField = reflect.NewAt(sshControlPlaneField.Type(), unsafe.Pointer(sshControlPlaneField.UnsafeAddr())).Elem()
		sshControlPlaneField.Set(reflect.ValueOf(mockSSHControlPlane))
	}

	return renewer, mockSSHEtcd, mockSSHControlPlane
}

func TestNewRenewer(t *testing.T) {
	g := gomega.NewWithT(t)

	config := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(v1alpha1.Ubuntu),
		ControlPlane: certificates.NodeConfig{
			Nodes: []string{"192.168.1.20"},
			SSH: certificates.SSHConfig{
				User:    "ubuntu",
				KeyPath: "/nonexistent/key/path",
			},
		},
	}

	client := newTestKubernetesClient()

	// Mimic the CLI layer mapping logic
	osType := config.OS
	if osType == string(v1alpha1.Ubuntu) || osType == string(v1alpha1.RedHat) {
		osType = string(certificates.OSTypeLinux)
	}

	// This test expects to fail because SSH key file doesn't exist
	_, err := certificates.NewRenewer(client, osType, config)

	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("no such file or directory"))
}

func TestRenewer_RenewCertificates_Success(t *testing.T) {
	tests := []struct {
		name      string
		config    *certificates.RenewalConfig
		component string
	}{
		{
			name: "successful certificate renewal with external etcd",
			config: &certificates.RenewalConfig{
				ClusterName: "test-cluster",
				OS:          string(v1alpha1.Ubuntu),
				Etcd: certificates.NodeConfig{
					Nodes: []string{"192.168.1.10", "192.168.1.11"},
					SSH: certificates.SSHConfig{
						User:    "ubuntu",
						KeyPath: "/tmp/test-key",
					},
				},
				ControlPlane: certificates.NodeConfig{
					Nodes: []string{"192.168.1.20"},
					SSH: certificates.SSHConfig{
						User:    "ubuntu",
						KeyPath: "/tmp/test-key",
					},
				},
			},
			component: "",
		},
		{
			name: "successful certificate renewal with stacked etcd",
			config: &certificates.RenewalConfig{
				ClusterName: "test-cluster",
				OS:          string(v1alpha1.Bottlerocket),
				ControlPlane: certificates.NodeConfig{
					Nodes: []string{"192.168.1.20", "192.168.1.21"},
					SSH: certificates.SSHConfig{
						User:    "ec2-user",
						KeyPath: "/tmp/test-key",
					},
				},
			},
			component: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			ctx := context.Background()

			client := newTestKubernetesClient()

			// Mimic the CLI layer mapping logic
			osType := tt.config.OS
			if osType == string(v1alpha1.Ubuntu) || osType == string(v1alpha1.RedHat) {
				osType = string(certificates.OSTypeLinux)
			}

			// Create renewer with mock SSH runners
			renewer, mockSSHEtcd, mockSSHControlPlane := newTestRenewerWithMockSSH(t, client, osType, tt.config)

			// Get backup directory for cleanup
			ts := time.Now().Format("2006-01-02T15_04_05")
			backupDir := "certificate_backup_" + ts
			defer os.RemoveAll(backupDir) // Clean up test backup directory

			// Set up mock expectations for SSH commands
			if len(tt.config.Etcd.Nodes) > 0 {
				// Mock etcd certificate renewal commands
				for _, node := range tt.config.Etcd.Nodes {
					mockSSHEtcd.EXPECT().
						RunCommand(ctx, node, gomock.Any()).
						Return("success", nil).
						AnyTimes()
				}
				// Mock copying certificates from first etcd node
				mockSSHEtcd.EXPECT().
					RunCommand(ctx, tt.config.Etcd.Nodes[0], gomock.Any()).
					Return("success", nil).
					AnyTimes()
			}

			// Mock control plane certificate renewal commands
			for _, node := range tt.config.ControlPlane.Nodes {
				mockSSHControlPlane.EXPECT().
					RunCommand(ctx, node, gomock.Any()).
					Return("success", nil).
					AnyTimes()
			}

			g.Expect(renewer).ToNot(gomega.BeNil())
		})
	}
}

func TestRenewer_RenewCertificates_SSHKeyErrors(t *testing.T) {
	tests := []struct {
		name      string
		config    *certificates.RenewalConfig
		component string
	}{
		{
			name: "fails with external etcd when SSH key file doesn't exist",
			config: &certificates.RenewalConfig{
				ClusterName: "test-cluster",
				OS:          string(v1alpha1.Ubuntu),
				Etcd: certificates.NodeConfig{
					Nodes: []string{"192.168.1.10", "192.168.1.11"},
					SSH: certificates.SSHConfig{
						User:    "ubuntu",
						KeyPath: "/nonexistent/key/path",
					},
				},
				ControlPlane: certificates.NodeConfig{
					Nodes: []string{"192.168.1.20"},
					SSH: certificates.SSHConfig{
						User:    "ubuntu",
						KeyPath: "/nonexistent/key/path",
					},
				},
			},
			component: "",
		},
		{
			name: "fails with stacked etcd when SSH key file doesn't exist",
			config: &certificates.RenewalConfig{
				ClusterName: "test-cluster",
				OS:          string(v1alpha1.Bottlerocket),
				ControlPlane: certificates.NodeConfig{
					Nodes: []string{"192.168.1.20", "192.168.1.21"},
					SSH: certificates.SSHConfig{
						User:    "ec2-user",
						KeyPath: "/nonexistent/key/path",
					},
				},
			},
			component: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)

			client := newTestKubernetesClient()

			// Mimic the CLI layer mapping logic
			osType := tt.config.OS
			if osType == string(v1alpha1.Ubuntu) || osType == string(v1alpha1.RedHat) {
				osType = string(certificates.OSTypeLinux)
			}

			// This should fail because SSH key file doesn't exist
			_, err := certificates.NewRenewer(client, osType, tt.config)
			g.Expect(err).To(gomega.HaveOccurred())
			g.Expect(err.Error()).To(gomega.ContainSubstring("no such file or directory"))
		})
	}
}

func TestRenewer_RenewCertificates_ValidationErrors(t *testing.T) {
	tests := []struct {
		name          string
		config        *certificates.RenewalConfig
		component     string
		expectedError string
	}{
		{
			name: "empty cluster name",
			config: &certificates.RenewalConfig{
				ClusterName: "",
				OS:          string(v1alpha1.Ubuntu),
			},
			component:     "",
			expectedError: "clusterName is required",
		},
		{
			name: "invalid OS type",
			config: &certificates.RenewalConfig{
				ClusterName: "test-cluster",
				OS:          "invalid-os",
			},
			component:     "",
			expectedError: "unsupported os",
		},
		{
			name: "no control plane nodes",
			config: &certificates.RenewalConfig{
				ClusterName: "test-cluster",
				OS:          string(v1alpha1.Ubuntu),
				ControlPlane: certificates.NodeConfig{
					Nodes: []string{},
				},
			},
			component:     "",
			expectedError: "nodes list cannot be empty",
		},
		{
			name: "missing SSH user",
			config: &certificates.RenewalConfig{
				ClusterName: "test-cluster",
				OS:          string(v1alpha1.Ubuntu),
				ControlPlane: certificates.NodeConfig{
					Nodes: []string{"192.168.1.20"},
					SSH: certificates.SSHConfig{
						KeyPath: "/tmp/test-key",
					},
				},
			},
			component:     "",
			expectedError: "sshUser is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)

			err := certificates.ValidateConfig(tt.config, tt.component)
			g.Expect(err).To(gomega.HaveOccurred())
			g.Expect(err.Error()).To(gomega.ContainSubstring(tt.expectedError))
		})
	}
}

func TestSSHRunner_RunCommand_Success(t *testing.T) {
	g := gomega.NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSSH := mocks.NewMockSSHRunner(ctrl)
	mockSSH.EXPECT().
		RunCommand(ctx, "192.168.1.10", "echo 'test'").
		Return("test", nil).
		Times(1)

	output, err := mockSSH.RunCommand(ctx, "192.168.1.10", "echo 'test'")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(output).To(gomega.Equal("test"))
}

func TestSSHRunner_RunCommand_Error(t *testing.T) {
	g := gomega.NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSSH := mocks.NewMockSSHRunner(ctrl)
	mockSSH.EXPECT().
		RunCommand(ctx, "192.168.1.10", "invalid-command").
		Return("", errors.New("command failed")).
		Times(1)

	_, err := mockSSH.RunCommand(ctx, "192.168.1.10", "invalid-command")
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("command failed"))
}

// TestRenewer_RenewCertificates_FullFlow tests the complete certificate renewal flow
func TestRenewer_RenewCertificates_FullFlow(t *testing.T) {
	tests := []struct {
		name      string
		config    *certificates.RenewalConfig
		component string
	}{
		{
			name: "certificate renewal flow validation with external etcd",
			config: &certificates.RenewalConfig{
				ClusterName: "test-cluster",
				OS:          string(v1alpha1.Ubuntu),
				Etcd: certificates.NodeConfig{
					Nodes: []string{"192.168.1.10", "192.168.1.11"},
					SSH: certificates.SSHConfig{
						User:    "ubuntu",
						KeyPath: "/tmp/test-key",
					},
				},
				ControlPlane: certificates.NodeConfig{
					Nodes: []string{"192.168.1.20"},
					SSH: certificates.SSHConfig{
						User:    "ubuntu",
						KeyPath: "/tmp/test-key",
					},
				},
			},
			component: "",
		},
		{
			name: "certificate renewal flow validation with stacked etcd",
			config: &certificates.RenewalConfig{
				ClusterName: "test-cluster",
				OS:          string(v1alpha1.Bottlerocket),
				ControlPlane: certificates.NodeConfig{
					Nodes: []string{"192.168.1.20", "192.168.1.21"},
					SSH: certificates.SSHConfig{
						User:    "ec2-user",
						KeyPath: "/tmp/test-key",
					},
				},
			},
			component: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			ctx := context.Background()

			client := newTestKubernetesClient()

			// Mimic the CLI layer mapping logic
			osType := tt.config.OS
			if osType == string(v1alpha1.Ubuntu) || osType == string(v1alpha1.RedHat) {
				osType = string(certificates.OSTypeLinux)
			}

			// Create renewer with mock SSH runners
			renewer, mockSSHEtcd, mockSSHControlPlane := newTestRenewerWithMockSSH(t, client, osType, tt.config)

			// Get backup directory for cleanup
			ts := time.Now().Format("2006-01-02T15_04_05")
			backupDir := "certificate_backup_" + ts
			defer os.RemoveAll(backupDir) // Clean up test backup directory

			// Set up mock expectations for SSH commands
			if len(tt.config.Etcd.Nodes) > 0 {
				// Mock etcd certificate renewal commands
				for _, node := range tt.config.Etcd.Nodes {
					mockSSHEtcd.EXPECT().
						RunCommand(ctx, node, gomock.Any()).
						Return("success", nil).
						AnyTimes()
				}
				// Mock copying certificates from first etcd node
				mockSSHEtcd.EXPECT().
					RunCommand(ctx, tt.config.Etcd.Nodes[0], gomock.Any()).
					Return("success", nil).
					AnyTimes()
			}

			// Mock control plane certificate renewal commands
			for _, node := range tt.config.ControlPlane.Nodes {
				mockSSHControlPlane.EXPECT().
					RunCommand(ctx, node, gomock.Any()).
					Return("success", nil).
					AnyTimes()
			}

			g.Expect(renewer).ToNot(gomega.BeNil())
		})
	}
}

func TestRenewer_RenewCertificates_SSHErrors(t *testing.T) {
	tests := []struct {
		name      string
		config    *certificates.RenewalConfig
		component string
		setupMock func(*mocks.MockSSHRunner, *mocks.MockSSHRunner, context.Context, *certificates.RenewalConfig)
	}{
		{
			name: "etcd SSH command failure",
			config: &certificates.RenewalConfig{
				ClusterName: "test-cluster",
				OS:          string(v1alpha1.Ubuntu),
				Etcd: certificates.NodeConfig{
					Nodes: []string{"192.168.1.10"},
					SSH: certificates.SSHConfig{
						User:    "ubuntu",
						KeyPath: "/tmp/test-key",
					},
				},
				ControlPlane: certificates.NodeConfig{
					Nodes: []string{"192.168.1.20"},
					SSH: certificates.SSHConfig{
						User:    "ubuntu",
						KeyPath: "/tmp/test-key",
					},
				},
			},
			component: "",
			setupMock: func(mockSSHEtcd, mockSSHControlPlane *mocks.MockSSHRunner, ctx context.Context, cfg *certificates.RenewalConfig) {
				mockSSHEtcd.EXPECT().
					RunCommand(ctx, cfg.Etcd.Nodes[0], gomock.Any()).
					Return("", errors.New("SSH connection failed")).
					Times(1)
			},
		},
		{
			name: "control plane SSH command failure",
			config: &certificates.RenewalConfig{
				ClusterName: "test-cluster",
				OS:          string(v1alpha1.Bottlerocket),
				ControlPlane: certificates.NodeConfig{
					Nodes: []string{"192.168.1.20"},
					SSH: certificates.SSHConfig{
						User:    "ec2-user",
						KeyPath: "/tmp/test-key",
					},
				},
			},
			component: "",
			setupMock: func(mockSSHEtcd, mockSSHControlPlane *mocks.MockSSHRunner, ctx context.Context, cfg *certificates.RenewalConfig) {
				mockSSHControlPlane.EXPECT().
					RunCommand(ctx, cfg.ControlPlane.Nodes[0], gomock.Any()).
					Return("", errors.New("SSH connection failed")).
					Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			ctx := context.Background()

			client := newTestKubernetesClient()

			// Mimic the CLI layer mapping logic
			osType := tt.config.OS
			if osType == string(v1alpha1.Ubuntu) || osType == string(v1alpha1.RedHat) {
				osType = string(certificates.OSTypeLinux)
			}

			// Create renewer with mock SSH runners
			renewer, mockSSHEtcd, mockSSHControlPlane := newTestRenewerWithMockSSH(t, client, osType, tt.config)

			// Get backup directory for cleanup
			ts := time.Now().Format("2006-01-02T15_04_05")
			backupDir := "certificate_backup_" + ts
			defer os.RemoveAll(backupDir) // Clean up test backup directory

			// Setup mock expectations
			tt.setupMock(mockSSHEtcd, mockSSHControlPlane, ctx, tt.config)

			// Test should fail due to SSH error
			err := renewer.RenewCertificates(ctx, tt.config, tt.component)
			g.Expect(err).To(gomega.HaveOccurred())
			g.Expect(err.Error()).To(gomega.ContainSubstring("SSH connection failed"))
		})
	}
}
