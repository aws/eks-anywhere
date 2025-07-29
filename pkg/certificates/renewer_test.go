package certificates_test

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

	"github.com/aws/eks-anywhere/pkg/certificates"
	"github.com/aws/eks-anywhere/pkg/certificates/mocks"
	kubemocks "github.com/aws/eks-anywhere/pkg/clients/kubernetes/mocks"
)

const (
	tempLocalEtcdCertsDir = "etcd-client-certs"
	backupDirStr          = "certificate_backup_"
)

func TestNewRenewerSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(certificates.OSTypeLinux),
		ControlPlane: certificates.NodeConfig{
			Nodes: []string{
				"0.0.0.0",
			},
			SSH: certificates.SSHConfig{
				User:    "XXXX",
				KeyPath: "test",
			},
		},
	}

	osRenewer := certificates.BuildOSRenewer(cfg.OS, t.TempDir())

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

	renewer := &certificates.Renewer{
		BackupDir:       t.TempDir(),
		Kubectl:         kubeClient,
		Os:              osRenewer,
		SshEtcd:         sshEtcd,
		SshControlPlane: sshCP,
	}

	if err := renewer.RenewCertificates(context.Background(), cfg, "etcd"); err != nil {
		t.Fatalf("RenewCertificates() expected no error, got: %v", err)
	}
}

func TestRenewEtcdCerts_BackupError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(certificates.OSTypeLinux),
		Etcd: certificates.NodeConfig{
			Nodes: []string{"etcd-1"},
		},
		ControlPlane: certificates.NodeConfig{
			Nodes: []string{
				"0.0.0.0",
			},
			SSH: certificates.SSHConfig{
				User:    "XXXX",
				KeyPath: "test",
			},
		},
	}

	osRenewer := certificates.BuildOSRenewer(cfg.OS, t.TempDir())

	sshEtcd := mocks.NewMockSSHRunner(ctrl)
	sshEtcd.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", fmt.Errorf("backing up etcd certs error")).
		AnyTimes()

	sshCP := mocks.NewMockSSHRunner(ctrl)
	kubeClient := kubemocks.NewMockClient(ctrl)

	renewer := &certificates.Renewer{
		BackupDir:       t.TempDir(),
		Kubectl:         kubeClient,
		Os:              osRenewer,
		SshEtcd:         sshEtcd,
		SshControlPlane: sshCP,
	}
	if err := renewer.RenewCertificates(context.Background(), cfg, ""); err == nil {
		t.Fatalf("renewEtcdCerts() expected error, got nil")
	}
}

func TestRenewEtcdCerts_RenewError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(certificates.OSTypeLinux),
		Etcd: certificates.NodeConfig{
			Nodes: []string{"etcd-1"},
		},
	}

	osRenewer := certificates.BuildOSRenewer(cfg.OS, t.TempDir())

	sshEtcd := mocks.NewMockSSHRunner(ctrl)
	sshEtcd.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", fmt.Errorf("renew etcd error")).
		AnyTimes()

	renewer := &certificates.Renewer{
		BackupDir: t.TempDir(),
		Os:        osRenewer,
		SshEtcd:   sshEtcd,
	}

	if err := renewer.RenewCertificates(context.Background(), cfg, ""); err == nil {
		t.Fatalf("renewEtcdCerts() expected error, got nil")
	}
}

func TestRenewEtcdCerts_SuccessfulRenewal(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(certificates.OSTypeLinux),
		Etcd: certificates.NodeConfig{
			Nodes: []string{"etcd-1"},
		},
	}

	osRenewer := certificates.BuildOSRenewer(cfg.OS, t.TempDir())

	sshEtcd := mocks.NewMockSSHRunner(ctrl)
	sshEtcd.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ string, opts ...certificates.SSHOption) (string, error) {
			return "dummy-cert-content", nil
		}).
		AnyTimes()

	kubeClient := kubemocks.NewMockClient(ctrl)
	kubeClient.EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(apierrors.NewNotFound(schema.GroupResource{}, "")).
		AnyTimes()

	renewer := &certificates.Renewer{
		BackupDir: t.TempDir(),
		Os:        osRenewer,
		SshEtcd:   sshEtcd,
		Kubectl:   kubeClient,
	}

	writeDummyEtcdCerts(t, renewer.BackupDir)

	if err := renewer.RenewCertificates(context.Background(), cfg, ""); err != nil {
		t.Fatalf("RenewCertificates() expected no error, got: %v", err)
	}
}

func TestRenewCertificates_RenewControlPlaneCertsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(certificates.OSTypeLinux),
		ControlPlane: certificates.NodeConfig{
			Nodes: []string{
				"0.0.0.0",
			},
			SSH: certificates.SSHConfig{
				User:    "XXXX",
				KeyPath: "test",
			},
		},
	}

	osRenewer := certificates.BuildOSRenewer(cfg.OS, t.TempDir())

	sshEtcd := mocks.NewMockSSHRunner(ctrl)
	sshCP := mocks.NewMockSSHRunner(ctrl)
	kubeClient := kubemocks.NewMockClient(ctrl)

	sshCP.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", fmt.Errorf("backing up control plane certs error")).
		AnyTimes()

	renewer := &certificates.Renewer{
		BackupDir:       t.TempDir(),
		Kubectl:         kubeClient,
		Os:              osRenewer,
		SshEtcd:         sshEtcd,
		SshControlPlane: sshCP,
	}

	if err := renewer.RenewCertificates(context.Background(), cfg, "control-plane"); err == nil {
		t.Fatalf("RenewCertificates() expected error, got nil")
	}
}

func TestRenewCertificates_UpdateAPIServerEtcdClientSecretError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(certificates.OSTypeLinux),
		Etcd: certificates.NodeConfig{
			Nodes: []string{
				"etcd-1",
			},
		},
		ControlPlane: certificates.NodeConfig{
			Nodes: []string{
				"0.0.0.0",
			},
			SSH: certificates.SSHConfig{
				User:    "XXXX",
				KeyPath: "test",
			},
		},
	}

	osRenewer := certificates.BuildOSRenewer(cfg.OS, t.TempDir())

	sshEtcd := mocks.NewMockSSHRunner(ctrl)
	sshCP := mocks.NewMockSSHRunner(ctrl)
	kubeClient := kubemocks.NewMockClient(ctrl)

	sshEtcd.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ string, opts ...certificates.SSHOption) (string, error) {
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

	renewer := &certificates.Renewer{
		BackupDir:       t.TempDir(),
		Kubectl:         kubeClient,
		Os:              osRenewer,
		SshEtcd:         sshEtcd,
		SshControlPlane: sshCP,
	}

	writeDummyEtcdCerts(t, renewer.BackupDir)

	if err := renewer.RenewCertificates(context.Background(), cfg, "etcd"); err == nil {
		t.Fatalf("RenewCertificates() expected error, got nil")
	}
}

func TestRenewCertificates_CopyEtcdCertsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(certificates.OSTypeLinux),
		Etcd: certificates.NodeConfig{
			Nodes: []string{
				"etcd-1",
			},
		},
		ControlPlane: certificates.NodeConfig{
			Nodes: []string{
				"0.0.0.0",
			},
			SSH: certificates.SSHConfig{
				User:    "XXXX",
				KeyPath: "test",
			},
		},
	}

	osRenewer := certificates.BuildOSRenewer(cfg.OS, t.TempDir())

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

	renewer := &certificates.Renewer{
		BackupDir:       t.TempDir(),
		Kubectl:         kubeClient,
		Os:              osRenewer,
		SshEtcd:         sshEtcd,
		SshControlPlane: sshCP,
	}

	if err := renewer.RenewCertificates(context.Background(), cfg, "etcd"); err == nil {
		t.Fatalf("RenewCertificates() expected error, got nil")
	}
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

func TestRenewCertificates_ReadCertificateFileError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(certificates.OSTypeLinux),
		Etcd: certificates.NodeConfig{
			Nodes: []string{"etcd-1"},
		},
	}

	osRenewer := certificates.BuildOSRenewer(cfg.OS, t.TempDir())
	sshEtcd := mocks.NewMockSSHRunner(ctrl)
	kubeClient := kubemocks.NewMockClient(ctrl)

	sshEtcd.EXPECT().RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _, cmd string, opts ...certificates.SSHOption) (string, error) {
			return "certificate-content", nil
		}).AnyTimes()

	renewer := &certificates.Renewer{
		BackupDir: t.TempDir(),
		Os:        osRenewer,
		SshEtcd:   sshEtcd,
		Kubectl:   kubeClient,
	}

	if err := renewer.RenewCertificates(context.Background(), cfg, ""); err == nil {
		t.Fatalf("RenewCertificates() expected read certificate file error, got nil")
	}
}

func TestRenewCertificates_ReadKeyFileError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(certificates.OSTypeLinux),
		Etcd: certificates.NodeConfig{
			Nodes: []string{"etcd-1"},
		},
	}

	osRenewer := certificates.BuildOSRenewer(cfg.OS, t.TempDir())
	sshEtcd := mocks.NewMockSSHRunner(ctrl)
	kubeClient := kubemocks.NewMockClient(ctrl)

	sshEtcd.EXPECT().RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _, cmd string, opts ...certificates.SSHOption) (string, error) {
			return "certificate-content", nil
		}).AnyTimes()

	renewer := &certificates.Renewer{
		BackupDir: t.TempDir(),
		Os:        osRenewer,
		SshEtcd:   sshEtcd,
		Kubectl:   kubeClient,
	}

	dir := filepath.Join(renewer.BackupDir, tempLocalEtcdCertsDir)
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "apiserver-etcd-client.crt"), []byte("cert"), 0o644)

	if err := renewer.RenewCertificates(context.Background(), cfg, ""); err == nil {
		t.Fatalf("RenewCertificates() expected read key file error, got nil")
	}
}

func TestRenewCertificates_KubernetesAPIUnavailable(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(certificates.OSTypeLinux),
		Etcd: certificates.NodeConfig{
			Nodes: []string{"etcd-1"},
		},
	}

	osRenewer := certificates.BuildOSRenewer(cfg.OS, t.TempDir())
	sshEtcd := mocks.NewMockSSHRunner(ctrl)
	kubeClient := kubemocks.NewMockClient(ctrl)

	sshEtcd.EXPECT().RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ string, opts ...certificates.SSHOption) (string, error) {
			return "certificate-content", nil
		}).AnyTimes()

	kubeClient.EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(fmt.Errorf("connection refused")).
		AnyTimes()

	renewer := &certificates.Renewer{
		BackupDir: t.TempDir(),
		Os:        osRenewer,
		SshEtcd:   sshEtcd,
		Kubectl:   kubeClient,
	}

	writeDummyEtcdCerts(t, renewer.BackupDir)

	if err := renewer.RenewCertificates(context.Background(), cfg, ""); err != nil {
		t.Fatalf("RenewCertificates() expected no error when API unavailable, got: %v", err)
	}
}

func TestRenewCertificates_SuccessfulSecretUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(certificates.OSTypeLinux),
		Etcd: certificates.NodeConfig{
			Nodes: []string{"etcd-1"},
		},
	}

	osRenewer := certificates.BuildOSRenewer(cfg.OS, t.TempDir())
	sshEtcd := mocks.NewMockSSHRunner(ctrl)
	kubeClient := kubemocks.NewMockClient(ctrl)

	sshEtcd.EXPECT().RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ string, opts ...certificates.SSHOption) (string, error) {
			return "certificate-content", nil
		}).AnyTimes()

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

	renewer := &certificates.Renewer{
		BackupDir: t.TempDir(),
		Os:        osRenewer,
		SshEtcd:   sshEtcd,
		Kubectl:   kubeClient,
	}

	writeDummyEtcdCerts(t, renewer.BackupDir)

	if err := renewer.RenewCertificates(context.Background(), cfg, ""); err != nil {
		t.Fatalf("RenewCertificates() expected no error, got: %v", err)
	}
}

func TestRenewCertificates_CleanupError(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	cfg := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(certificates.OSTypeLinux),
		ControlPlane: certificates.NodeConfig{
			Nodes: []string{"cp-1"},
		},
	}

	osRenewer := certificates.BuildOSRenewer(cfg.OS, t.TempDir())
	sshCP := mocks.NewMockSSHRunner(ctrl)

	sshCP.EXPECT().RunCommand(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()

	renewer := &certificates.Renewer{
		BackupDir:       "/non-existent-path/cannot-delete",
		Os:              osRenewer,
		SshControlPlane: sshCP,
	}

	if err := renewer.RenewCertificates(context.Background(), cfg, ""); err != nil {
		t.Fatalf("RenewCertificates() expected no error even with cleanup failure, got: %v", err)
	}
}

func TestNewRenewerEtcdSSHError(t *testing.T) {
	cfg := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(certificates.OSTypeLinux),
		Etcd: certificates.NodeConfig{
			Nodes: []string{"etcd-1"},
			SSH: certificates.SSHConfig{
				User:    "",
				KeyPath: "/non-existent-path",
			},
		},
		ControlPlane: certificates.NodeConfig{
			Nodes: []string{"cp-1"},
			SSH: certificates.SSHConfig{
				User:    "user",
				KeyPath: "key-path",
			},
		},
	}

	kubeClient := kubemocks.NewMockClient(nil)

	_, err := certificates.NewRenewer(kubeClient, cfg.OS, cfg)
	if err == nil {
		t.Fatal("NewRenewer() expected error, got nil")
	}
}

func TestNewRenewerControlPlaneSSHError(t *testing.T) {
	cfg := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(certificates.OSTypeLinux),
		Etcd: certificates.NodeConfig{
			Nodes: []string{},
		},
		ControlPlane: certificates.NodeConfig{
			Nodes: []string{"cp-1"},
			SSH: certificates.SSHConfig{
				User:    "",
				KeyPath: "/non-existent-path",
			},
		},
	}

	kubeClient := kubemocks.NewMockClient(nil)

	_, err := certificates.NewRenewer(kubeClient, cfg.OS, cfg)
	if err == nil {
		t.Fatal("NewRenewer() expected error, got nil")
	}
}
