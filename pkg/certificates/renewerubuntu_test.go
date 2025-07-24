package certificates_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/pkg/certificates"
	"github.com/aws/eks-anywhere/pkg/certificates/mocks"
	kubemocks "github.com/aws/eks-anywhere/pkg/clients/kubernetes/mocks"
)

var errBoomTest = fmt.Errorf("boom-test")

func TestNewRenewerWithEtcdSSHError(t *testing.T) {
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

func TestNewRenewerWithControlPlaneSSHError(t *testing.T) {
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

func TestRenewControlPlaneCertsWithRenewError(t *testing.T) {
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
	ssh := mocks.NewMockSSHRunner(ctrl)
	kubeClient := kubemocks.NewMockClient(ctrl)

	ssh.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", errBoomTest)

	renewer := &certificates.Renewer{
		BackupDir:       t.TempDir(),
		Kubectl:         kubeClient,
		OS:              osRenewer,
		SSHControlPlane: ssh,
	}

	if err := renewer.RenewCertificates(context.Background(), cfg, "control-plane"); err == nil {
		t.Fatalf("RenewCertificates() expected error, got nil")
	}
}

func TestRenewControlPlaneCertsWithExternalEtcdValidationError(t *testing.T) {
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
	ssh := mocks.NewMockSSHRunner(ctrl)
	kubeClient := kubemocks.NewMockClient(ctrl)

	ssh.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", errBoomTest)

	renewer := &certificates.Renewer{
		BackupDir:       t.TempDir(),
		Kubectl:         kubeClient,
		OS:              osRenewer,
		SSHControlPlane: ssh,
	}

	if err := renewer.RenewCertificates(context.Background(), cfg, "control-plane"); err == nil {
		t.Fatalf("RenewCertificates() expected error, got nil")
	}
}

func TestRenewEtcdCertsWithJoinError(t *testing.T) {
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
	ssh := mocks.NewMockSSHRunner(ctrl)
	kubeClient := kubemocks.NewMockClient(ctrl)

	ssh.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", errBoomTest)

	renewer := &certificates.Renewer{
		BackupDir: t.TempDir(),
		Kubectl:   kubeClient,
		OS:        osRenewer,
		SSHEtcd:   ssh,
	}

	if err := renewer.RenewCertificates(context.Background(), cfg, ""); err == nil {
		t.Fatalf("RenewCertificates() expected error, got nil")
	}
}

func TestRenewEtcdCertsWithValidateError(t *testing.T) {
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
	ssh := mocks.NewMockSSHRunner(ctrl)
	kubeClient := kubemocks.NewMockClient(ctrl)

	ssh.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", errBoomTest)

	renewer := &certificates.Renewer{
		BackupDir: t.TempDir(),
		Kubectl:   kubeClient,
		OS:        osRenewer,
		SSHEtcd:   ssh,
	}

	if err := renewer.RenewCertificates(context.Background(), cfg, ""); err == nil {
		t.Fatalf("RenewCertificates() expected error, got nil")
	}
}

func TestRenewControlPlaneCertsWithExternalEtcdTransferError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &certificates.RenewalConfig{
		Etcd: certificates.NodeConfig{
			Nodes: []string{"etcd-1"},
		},
	}

	ssh := mocks.NewMockSSHRunner(ctrl)
	r := certificates.NewLinuxRenewer(t.TempDir())

	ssh.EXPECT().
		RunCommand(gomock.Any(), "cp-1", gomock.Any(), gomock.Any()).
		Return("", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), "cp-1", gomock.Any(), gomock.Any()).
		Return("", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), "cp-1", gomock.Any(), gomock.Any()).
		Return("", errBoomTest)

	err := r.RenewControlPlaneCerts(context.Background(), "cp-1", cfg, "", ssh)
	if err == nil {
		t.Fatalf("RenewControlPlaneCerts() expected error, got nil")
	}
}

func TestCopyEtcdCertsToLocalWithEmptyCert(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ssh := mocks.NewMockSSHRunner(ctrl)
	r := certificates.NewLinuxRenewer(t.TempDir())

	ssh.EXPECT().
		RunCommand(gomock.Any(), "etcd-1", gomock.Any(), gomock.Any()).
		Return("", nil)

	err := r.CopyEtcdCertsToLocal(context.Background(), "etcd-1", ssh)
	if err == nil {
		t.Fatalf("CopyEtcdCertsToLocal() expected error, got nil")
	}
}

func TestCopyEtcdCertsToLocalWithEmptyKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ssh := mocks.NewMockSSHRunner(ctrl)
	r := certificates.NewLinuxRenewer(t.TempDir())

	ssh.EXPECT().
		RunCommand(gomock.Any(), "etcd-1", gomock.Any(), gomock.Any()).
		Return("cert", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), "etcd-1", gomock.Any(), gomock.Any()).
		Return("", nil)

	err := r.CopyEtcdCertsToLocal(context.Background(), "etcd-1", ssh)
	if err == nil {
		t.Fatalf("CopyEtcdCertsToLocal() expected error, got nil")
	}
}

func TestTransferCertsToControlPlaneWithReadCertError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ssh := mocks.NewMockSSHRunner(ctrl)
	r := certificates.NewLinuxRenewer(t.TempDir())

	err := r.TransferCertsToControlPlaneFromLocal(context.Background(), "cp-1", ssh)
	if err == nil {
		t.Fatalf("TransferCertsToControlPlaneFromLocal() expected error, got nil")
	}
}

func TestTransferCertsToControlPlaneWithCopyCertError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tmp := t.TempDir()
	localDir := filepath.Join(tmp, tempLocalEtcdCertsDir)
	if err := os.MkdirAll(localDir, 0o700); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localDir, "apiserver-etcd-client.crt"), []byte("crt"), 0o600); err != nil {
		t.Fatalf("failed to write certificate file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localDir, "apiserver-etcd-client.key"), []byte("key"), 0o600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	ssh := mocks.NewMockSSHRunner(ctrl)
	r := certificates.NewLinuxRenewer(tmp)

	ssh.EXPECT().
		RunCommand(gomock.Any(), "cp-1", gomock.Any(), gomock.Any()).
		Return("", errBoomTest)

	err := r.TransferCertsToControlPlaneFromLocal(context.Background(), "cp-1", ssh)
	if err == nil {
		t.Fatalf("TransferCertsToControlPlaneFromLocal() expected error, got nil")
	}
}

func TestRenewControlPlaneCertsWithRestartPodsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &certificates.RenewalConfig{}

	ssh := mocks.NewMockSSHRunner(ctrl)
	r := certificates.NewLinuxRenewer(t.TempDir())

	ssh.EXPECT().
		RunCommand(gomock.Any(), "cp-1", gomock.Any(), gomock.Any()).
		Return("", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), "cp-1", gomock.Any(), gomock.Any()).
		Return("", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), "cp-1", gomock.Any(), gomock.Any()).
		Return("", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), "cp-1", gomock.Any(), gomock.Any()).
		Return("", errBoomTest)

	err := r.RenewControlPlaneCerts(context.Background(), "cp-1", cfg, "", ssh)
	if err == nil {
		t.Fatalf("RenewControlPlaneCerts() expected error, got nil")
	}
}

func TestCopyEtcdCertsToLocalWithReadKeyError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ssh := mocks.NewMockSSHRunner(ctrl)
	r := certificates.NewLinuxRenewer(t.TempDir())

	ssh.EXPECT().
		RunCommand(gomock.Any(), "etcd-1", gomock.Any(), gomock.Any()).
		Return("cert", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), "etcd-1", gomock.Any(), gomock.Any()).
		Return("", errBoomTest)

	err := r.CopyEtcdCertsToLocal(context.Background(), "etcd-1", ssh)
	if err == nil {
		t.Fatalf("CopyEtcdCertsToLocal() expected error, got nil")
	}
}

func TestCopyEtcdCertsToLocalWithMkdirError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tmp := t.TempDir()
	badDir := filepath.Join(tmp, tempLocalEtcdCertsDir)
	if err := os.WriteFile(badDir, []byte("x"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	ssh := mocks.NewMockSSHRunner(ctrl)
	r := certificates.NewLinuxRenewer(tmp)

	ssh.EXPECT().
		RunCommand(gomock.Any(), "etcd-1", gomock.Any(), gomock.Any()).
		Return("cert", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), "etcd-1", gomock.Any(), gomock.Any()).
		Return("key", nil)

	err := r.CopyEtcdCertsToLocal(context.Background(), "etcd-1", ssh)
	if err == nil {
		t.Fatalf("CopyEtcdCertsToLocal() expected error, got nil")
	}
}

func TestTransferCertsToControlPlaneWithReadKeyError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tmp := t.TempDir()
	localDir := filepath.Join(tmp, tempLocalEtcdCertsDir)
	if err := os.MkdirAll(localDir, 0o700); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(filepath.Join(localDir, "apiserver-etcd-client.crt"), []byte("crt"), 0o600); err != nil {
		t.Fatalf("failed to write certificate file: %v", err)
	}

	ssh := mocks.NewMockSSHRunner(ctrl)
	r := certificates.NewLinuxRenewer(tmp)

	err := r.TransferCertsToControlPlaneFromLocal(context.Background(), "cp-1", ssh)
	if err == nil {
		t.Fatalf("TransferCertsToControlPlaneFromLocal() expected error, got nil")
	}
}

func TestTransferCertsToControlPlaneWithCopyKeyError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tmp := t.TempDir()
	writeDummyEtcdCerts(t, tmp)

	ssh := mocks.NewMockSSHRunner(ctrl)
	r := certificates.NewLinuxRenewer(tmp)

	ssh.EXPECT().
		RunCommand(gomock.Any(), "cp-1", gomock.Any(), gomock.Any()).
		Return("", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), "cp-1", gomock.Any(), gomock.Any()).
		Return("", errBoomTest)

	err := r.TransferCertsToControlPlaneFromLocal(context.Background(), "cp-1", ssh)
	if err == nil {
		t.Fatalf("TransferCertsToControlPlaneFromLocal() expected error, got nil")
	}
}

func TestCopyEtcdCertsToLocalWithWriteKeyError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tmp := t.TempDir()
	localDir := filepath.Join(tmp, tempLocalEtcdCertsDir)
	if err := os.MkdirAll(localDir, 0o700); err != nil {
		t.Fatalf("setup: %v", err)
	}

	certPath := filepath.Join(localDir, "apiserver-etcd-client.crt")
	if err := os.WriteFile(certPath, []byte("cert"), 0o600); err != nil {
		t.Fatalf("failed to write certificate file: %v", err)
	}

	keyPath := filepath.Join(localDir, "apiserver-etcd-client.key")
	if err := os.Mkdir(keyPath, 0o755); err != nil {
		t.Fatalf("setup key conflict: %v", err)
	}

	ssh := mocks.NewMockSSHRunner(ctrl)
	r := certificates.NewLinuxRenewer(tmp)

	ssh.EXPECT().
		RunCommand(gomock.Any(), "etcd-1", gomock.Any(), gomock.Any()).
		Return("cert", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), "etcd-1", gomock.Any(), gomock.Any()).
		Return("key", nil)

	err := r.CopyEtcdCertsToLocal(context.Background(), "etcd-1", ssh)
	if err == nil {
		t.Fatalf("CopyEtcdCertsToLocal() expected error, got nil")
	}
}
