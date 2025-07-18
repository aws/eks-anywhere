package certificates

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/aws/eks-anywhere/pkg/certificates/mocks"
	kubemocks "github.com/aws/eks-anywhere/pkg/clients/kubernetes/mocks"
)

func TestNewRenewerSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(OSTypeLinux),
		ControlPlane: NodeConfig{
			Nodes: []string{
				"0.0.0.0",
			},
			SSH: SSHConfig{
				User:    "XXXX",
				KeyPath: "test",
			},
		},
	}

	osRenewer := BuildOSRenewer(cfg.OS, t.TempDir())

	sshEtcd := mocks.NewMockSSHRunner(ctrl)
	sshCP := mocks.NewMockSSHRunner(ctrl)
	kubeClient := kubemocks.NewMockClient(ctrl)

	sshCP.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", nil).
		AnyTimes()

	sshEtcd.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", nil).
		AnyTimes()

	kubeClient.EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(apierrors.NewNotFound(schema.GroupResource{}, "")).
		AnyTimes()

	renewer := &Renewer{
		backupDir:       t.TempDir(),
		kubectl:         kubeClient,
		os:              osRenewer,
		sshEtcd:         sshEtcd,
		sshControlPlane: sshCP,
	}

	if err := renewer.RenewCertificates(context.Background(), cfg, "etcd"); err != nil {
		t.Fatalf("RenewCertificates() expected no error, got: %v", err)
	}
}

func TestRenewEtcdCerts_BackupError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(OSTypeLinux),
		Etcd: NodeConfig{
			Nodes: []string{"etcd-1"},
		},
		ControlPlane: NodeConfig{
			Nodes: []string{
				"0.0.0.0",
			},
			SSH: SSHConfig{
				User:    "XXXX",
				KeyPath: "test",
			},
		},
	}

	osRenewer := BuildOSRenewer(cfg.OS, t.TempDir())

	sshEtcd := mocks.NewMockSSHRunner(ctrl)
	sshEtcd.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", fmt.Errorf("backing up etcd certs error")).
		AnyTimes()

	sshCP := mocks.NewMockSSHRunner(ctrl)
	kubeClient := kubemocks.NewMockClient(ctrl)

	renewer := &Renewer{
		backupDir:       t.TempDir(),
		kubectl:         kubeClient,
		os:              osRenewer,
		sshEtcd:         sshEtcd,
		sshControlPlane: sshCP,
	}
	if err := renewer.renewEtcdCerts(context.Background(), cfg); err == nil {
		t.Fatalf("renewEtcdCerts() expected error, got nil")
	}
}

func TestRenewEtcdCerts_RenewError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(OSTypeLinux),
		Etcd: NodeConfig{
			Nodes: []string{"etcd-1"},
		},
	}

	osRenewer := BuildOSRenewer(cfg.OS, t.TempDir())

	sshEtcd := mocks.NewMockSSHRunner(ctrl)
	sshEtcd.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", fmt.Errorf("renew etcd error")).
		AnyTimes()

	renewer := &Renewer{
		backupDir: t.TempDir(),
		os:        osRenewer,
		sshEtcd:   sshEtcd,
	}

	if err := renewer.renewEtcdCerts(context.Background(), cfg); err == nil {
		t.Fatalf("renewEtcdCerts() expected error, got nil")
	}
}

func TestRenewEtcdCerts_SuccessfulRenewal(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(OSTypeLinux),
		Etcd: NodeConfig{
			Nodes: []string{"etcd-1"},
		},
	}

	osRenewer := BuildOSRenewer(cfg.OS, t.TempDir())

	sshEtcd := mocks.NewMockSSHRunner(ctrl)
	sshEtcd.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ string) (string, error) {
			return "dummy-cert-content", nil
		}).
		AnyTimes()

	renewer := &Renewer{
		backupDir: t.TempDir(),
		os:        osRenewer,
		sshEtcd:   sshEtcd,
	}

	if err := renewer.renewEtcdCerts(context.Background(), cfg); err != nil {
		t.Fatalf("renewEtcdCerts() expected no error, got: %v", err)
	}
}

func TestRenewControlPlaneCerts_SuccessfulRenewal(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(OSTypeLinux),
		ControlPlane: NodeConfig{
			Nodes: []string{"cp-1"},
		},
	}

	osRenewer := BuildOSRenewer(cfg.OS, t.TempDir())

	sshCP := mocks.NewMockSSHRunner(ctrl)
	sshCP.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", nil).
		AnyTimes()

	renewer := &Renewer{
		backupDir:       t.TempDir(),
		os:              osRenewer,
		sshControlPlane: sshCP,
	}

	if err := renewer.renewControlPlaneCerts(context.Background(), cfg, "control-plane"); err != nil {
		t.Fatalf("renewControlPlaneCerts() expected no error, got: %v", err)
	}
}

func TestRenewControlPlaneCerts_RenewError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(OSTypeLinux),
		ControlPlane: NodeConfig{
			Nodes: []string{"cp-1"},
		},
	}

	osRenewer := BuildOSRenewer(cfg.OS, t.TempDir())

	sshCP := mocks.NewMockSSHRunner(ctrl)
	sshCP.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", fmt.Errorf("renew control plane error")).
		AnyTimes()

	renewer := &Renewer{
		backupDir:       t.TempDir(),
		os:              osRenewer,
		sshControlPlane: sshCP,
	}

	if err := renewer.renewControlPlaneCerts(context.Background(), cfg, "control-plane"); err == nil {
		t.Fatalf("renewControlPlaneCerts() expected error, got nil")
	}
}

func TestRenewCertificates_RenewControlPlaneCertsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(OSTypeLinux),
		ControlPlane: NodeConfig{
			Nodes: []string{
				"0.0.0.0",
			},
			SSH: SSHConfig{
				User:    "XXXX",
				KeyPath: "test",
			},
		},
	}

	osRenewer := BuildOSRenewer(cfg.OS, t.TempDir())

	sshEtcd := mocks.NewMockSSHRunner(ctrl)
	sshCP := mocks.NewMockSSHRunner(ctrl)
	kubeClient := kubemocks.NewMockClient(ctrl)

	sshCP.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", fmt.Errorf("backing up control plane certs error")).
		AnyTimes()

	renewer := &Renewer{
		backupDir:       t.TempDir(),
		kubectl:         kubeClient,
		os:              osRenewer,
		sshEtcd:         sshEtcd,
		sshControlPlane: sshCP,
	}

	if err := renewer.RenewCertificates(context.Background(), cfg, "control-plane"); err == nil {
		t.Fatalf("RenewCertificates() expected error, got nil")
	}
}

func TestRenewCertificates_UpdateAPIServerEtcdClientSecretError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(OSTypeLinux),
		Etcd: NodeConfig{
			Nodes: []string{
				"etcd-1",
			},
		},
		ControlPlane: NodeConfig{
			Nodes: []string{
				"0.0.0.0",
			},
			SSH: SSHConfig{
				User:    "XXXX",
				KeyPath: "test",
			},
		},
	}

	osRenewer := BuildOSRenewer(cfg.OS, t.TempDir())

	sshEtcd := mocks.NewMockSSHRunner(ctrl)
	sshCP := mocks.NewMockSSHRunner(ctrl)
	kubeClient := kubemocks.NewMockClient(ctrl)

	sshEtcd.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ string) (string, error) {
			return "certificate or key content", nil
		}).
		AnyTimes()

	sshCP.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", nil).
		AnyTimes()

	kubeClient.EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ string, obj interface{}) error {
			secret := obj.(*corev1.Secret)
			secret.Data = map[string][]byte{}
			return nil
		}).
		AnyTimes()

	kubeClient.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(fmt.Errorf("updating secret error")).
		AnyTimes()

	renewer := &Renewer{
		backupDir:       t.TempDir(),
		kubectl:         kubeClient,
		os:              osRenewer,
		sshEtcd:         sshEtcd,
		sshControlPlane: sshCP,
	}

	writeDummyEtcdCerts(t, renewer.backupDir)

	if err := renewer.RenewCertificates(context.Background(), cfg, "etcd"); err == nil {
		t.Fatalf("RenewCertificates() expected error, got nil")
	}
}

func TestRenewCertificates_CopyEtcdCertsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(OSTypeLinux),
		Etcd: NodeConfig{
			Nodes: []string{
				"etcd-1",
			},
		},
		ControlPlane: NodeConfig{
			Nodes: []string{
				"0.0.0.0",
			},
			SSH: SSHConfig{
				User:    "XXXX",
				KeyPath: "test",
			},
		},
	}

	osRenewer := BuildOSRenewer(cfg.OS, t.TempDir())

	sshEtcd := mocks.NewMockSSHRunner(ctrl)
	sshCP := mocks.NewMockSSHRunner(ctrl)
	kubeClient := kubemocks.NewMockClient(ctrl)

	firstCall := sshEtcd.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", nil)

	secondCall := sshEtcd.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", nil).
		After(firstCall)

	thirdCall := sshEtcd.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", nil).
		After(secondCall)

	sshEtcd.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", fmt.Errorf("copy etcd certs error")).
		After(thirdCall)

	renewer := &Renewer{
		backupDir:       t.TempDir(),
		kubectl:         kubeClient,
		os:              osRenewer,
		sshEtcd:         sshEtcd,
		sshControlPlane: sshCP,
	}

	if err := renewer.RenewCertificates(context.Background(), cfg, "etcd"); err == nil {
		t.Fatalf("RenewCertificates() expected error, got nil")
	}
}

func TestUpdateAPIServerEtcdClientSecret_ReadCertificateFileError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	kubeClient := kubemocks.NewMockClient(ctrl)

	renewer := &Renewer{
		backupDir: t.TempDir(),
		kubectl:   kubeClient,
	}

	if err := renewer.updateAPIServerEtcdClientSecret(context.Background(), "test-cluster"); err == nil {
		t.Fatalf("updateAPIServerEtcdClientSecret() expected error, got nil")
	}
}

func TestUpdateAPIServerEtcdClientSecret_SecretNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	kubeClient := kubemocks.NewMockClient(ctrl)

	kubeClient.EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(apierrors.NewNotFound(schema.GroupResource{}, "")).
		AnyTimes()

	renewer := &Renewer{
		backupDir: t.TempDir(),
		kubectl:   kubeClient,
	}

	writeDummyEtcdCerts(t, renewer.backupDir)

	if err := renewer.updateAPIServerEtcdClientSecret(context.Background(), "test-cluster"); err != nil {
		t.Fatalf("updateAPIServerEtcdClientSecret() expected no error, got: %v", err)
	}
}

func TestUpdateAPIServerEtcdClientSecret_GetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	kubeClient := kubemocks.NewMockClient(ctrl)

	kubeClient.EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(fmt.Errorf("get error")).
		AnyTimes()

	renewer := &Renewer{
		backupDir: t.TempDir(),
		kubectl:   kubeClient,
	}

	writeDummyEtcdCerts(t, renewer.backupDir)

	if err := renewer.updateAPIServerEtcdClientSecret(context.Background(), "test-cluster"); err != nil {
		t.Fatalf("updateAPIServerEtcdClientSecret() expected no error, got: %v", err)
	}
}

func TestUpdateAPIServerEtcdClientSecret_ReadKeyFileError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	kubeClient := kubemocks.NewMockClient(ctrl)

	renewer := &Renewer{
		backupDir: t.TempDir(),
		kubectl:   kubeClient,
	}

	dir := filepath.Join(renewer.backupDir, tempLocalEtcdCertsDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("creating directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "apiserver-etcd-client.crt"), []byte("cert"), 0o644); err != nil {
		t.Fatalf("writing crt: %v", err)
	}

	if err := renewer.updateAPIServerEtcdClientSecret(context.Background(), "test-cluster"); err == nil {
		t.Fatalf("updateAPIServerEtcdClientSecret() expected error, got nil")
	}
}

func TestUpdateAPIServerEtcdClientSecret_SuccessfulUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	kubeClient := kubemocks.NewMockClient(ctrl)

	kubeClient.EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ string, obj interface{}) error {
			secret := obj.(*corev1.Secret)
			secret.Data = map[string][]byte{}
			return nil
		})

	kubeClient.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(nil)

	renewer := &Renewer{
		backupDir: t.TempDir(),
		kubectl:   kubeClient,
	}

	writeDummyEtcdCerts(t, renewer.backupDir)

	if err := renewer.updateAPIServerEtcdClientSecret(context.Background(), "test-cluster"); err != nil {
		t.Fatalf("updateAPIServerEtcdClientSecret() expected no error, got: %v", err)
	}
}

func TestCleanup_NonExistentDirectoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	renewer := &Renewer{
		backupDir: "/non-existent-dir",
	}

	renewer.cleanup()
}

func writeDummyEtcdCerts(t *testing.T, backupDir string) {
	t.Helper()

	dir := filepath.Join(backupDir, tempLocalEtcdCertsDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("creating directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "apiserver-etcd-client.crt"), []byte("cert"), 0o644); err != nil {
		t.Fatalf("writing crt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "apiserver-etcd-client.key"), []byte("key"), 0o644); err != nil {
		t.Fatalf("writing key: %v", err)
	}
}

func TestValidateRenewalConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *RenewalConfig
		component string
		wantEtcd  bool
		wantCP    bool
		wantErr   string
	}{
		{
			name:      "valid etcd",
			component: "etcd",
			config: &RenewalConfig{
				Etcd: NodeConfig{Nodes: []string{"etcd-1"}},
			},
			wantEtcd: true,
			wantCP:   true,
		},
		{
			name:      "valid control-plane",
			component: "control-plane",
			config: &RenewalConfig{
				ControlPlane: NodeConfig{Nodes: []string{"cp-1"}},
			},
			wantEtcd: false,
			wantCP:   true,
		},
		{
			name:      "no etcd nodes",
			component: "etcd",
			config:    &RenewalConfig{},
			wantEtcd:  false,
			wantCP:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renewer := &Renewer{}
			processEtcd, processControlPlane, err := renewer.validateRenewalConfig(tt.config, tt.component)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if processEtcd != tt.wantEtcd {
				t.Fatalf("expected processEtcd=%v, got: %v", tt.wantEtcd, processEtcd)
			}
			if processControlPlane != tt.wantCP {
				t.Fatalf("expected processControlPlane=%v, got: %v", tt.wantCP, processControlPlane)
			}
		})
	}
}
