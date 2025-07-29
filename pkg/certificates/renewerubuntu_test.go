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
)

var errString = fmt.Errorf("error")

func TestLinuxRenewer_RenewControlPlaneCerts_RenewError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(certificates.OSTypeLinux),
		ControlPlane: certificates.NodeConfig{
			Nodes: []string{"cp-1"},
		},
	}

	ssh := mocks.NewMockSSHRunner(ctrl)
	r := certificates.NewLinuxRenewer(t.TempDir())

	ssh.EXPECT().
		RunCommand(gomock.Any(), "cp-1", gomock.Any(), gomock.Any()).
		Return("", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), "cp-1", gomock.Any(), gomock.Any()).
		Return("", errString)

	if err := r.RenewControlPlaneCerts(context.Background(), "cp-1", cfg, "", ssh); err == nil {
		t.Fatalf("RenewControlPlaneCerts() expected error, got nil")
	}
}

func TestLinuxRenewer_RenewControlPlaneCerts_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &certificates.RenewalConfig{
		ClusterName: "test-cluster",
		OS:          string(certificates.OSTypeLinux),
		ControlPlane: certificates.NodeConfig{
			Nodes: []string{"cp-1"},
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
		Return("", errString)

	if err := r.RenewControlPlaneCerts(context.Background(), "cp-1", cfg, "", ssh); err == nil {
		t.Fatalf("RenewControlPlaneCerts() expected error, got nil")
	}
}

func TestLinuxRenewer_RenewControlPlaneCerts_RestartPodsError(t *testing.T) {
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
		Return("", errString)

	if err := r.RenewControlPlaneCerts(context.Background(), "cp-1", cfg, "", ssh); err == nil {
		t.Fatalf("RenewControlPlaneCerts() expected error, got nil")
	}
}

func TestLinuxRenewer_RenewControlPlaneCerts_ExternalEtcdTransferError(t *testing.T) {
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
		Return("", errString)

	if err := r.RenewControlPlaneCerts(context.Background(), "cp-1", cfg, "", ssh); err == nil {
		t.Fatalf("RenewControlPlaneCerts() expected error, got nil")
	}
}

func TestLinuxRenewer_RenewEtcdCerts_JoinError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ssh := mocks.NewMockSSHRunner(ctrl)
	r := certificates.NewLinuxRenewer(t.TempDir())

	ssh.EXPECT().
		RunCommand(gomock.Any(), "etcd-1", gomock.Any(), gomock.Any()).
		Return("", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), "etcd-1", gomock.Any(), gomock.Any()).
		Return("", errString)

	if err := r.RenewEtcdCerts(context.Background(), "etcd-1", ssh); err == nil {
		t.Fatalf("RenewEtcdCerts() expected error, got nil")
	}
}

func TestLinuxRenewer_RenewEtcdCerts_ValidateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ssh := mocks.NewMockSSHRunner(ctrl)
	r := certificates.NewLinuxRenewer(t.TempDir())

	ssh.EXPECT().
		RunCommand(gomock.Any(), "etcd-1", gomock.Any(), gomock.Any()).
		Return("", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), "etcd-1", gomock.Any(), gomock.Any()).
		Return("", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), "etcd-1", gomock.Any(), gomock.Any()).
		Return("", errString)

	if err := r.RenewEtcdCerts(context.Background(), "etcd-1", ssh); err == nil {
		t.Fatalf("RenewEtcdCerts() expected error, got nil")
	}
}

func TestLinuxRenewer_CopyEtcdCertsToLocal_EmptyCert(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ssh := mocks.NewMockSSHRunner(ctrl)
	r := certificates.NewLinuxRenewer(t.TempDir())

	ssh.EXPECT().
		RunCommand(gomock.Any(), "etcd-1", gomock.Any(), gomock.Any()).
		Return("", nil)

	if err := r.CopyEtcdCertsToLocal(context.Background(), "etcd-1", ssh); err == nil {
		t.Fatalf("CopyEtcdCertsToLocal() expected error, got nil")
	}
}

func TestLinuxRenewer_CopyEtcdCertsToLocal_EmptyKey(t *testing.T) {
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

	if err := r.CopyEtcdCertsToLocal(context.Background(), "etcd-1", ssh); err == nil {
		t.Fatalf("CopyEtcdCertsToLocal() expected error, got nil")
	}
}

func TestLinuxRenewer_CopyEtcdCertsToLocal_ReadKeyError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ssh := mocks.NewMockSSHRunner(ctrl)
	r := certificates.NewLinuxRenewer(t.TempDir())

	ssh.EXPECT().
		RunCommand(gomock.Any(), "etcd-1", gomock.Any(), gomock.Any()).
		Return("cert", nil)

	ssh.EXPECT().
		RunCommand(gomock.Any(), "etcd-1", gomock.Any(), gomock.Any()).
		Return("", errString)

	if err := r.CopyEtcdCertsToLocal(context.Background(), "etcd-1", ssh); err == nil {
		t.Fatalf("CopyEtcdCertsToLocal() expected error, got nil")
	}
}

func TestLinuxRenewer_CopyEtcdCertsToLocal_MkdirError(t *testing.T) {
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

	if err := r.CopyEtcdCertsToLocal(context.Background(), "etcd-1", ssh); err == nil {
		t.Fatalf("CopyEtcdCertsToLocal() expected error, got nil")
	}
}

func TestLinuxRenewer_CopyEtcdCertsToLocal_WriteKeyError(t *testing.T) {
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

	if err := r.CopyEtcdCertsToLocal(context.Background(), "etcd-1", ssh); err == nil {
		t.Fatalf("CopyEtcdCertsToLocal() expected error, got nil")
	}
}

func TestLinuxRenewer_TransferCertsToControlPlaneFromLocal_ReadCertError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ssh := mocks.NewMockSSHRunner(ctrl)
	r := certificates.NewLinuxRenewer(t.TempDir())

	if err := r.TransferCertsToControlPlaneFromLocal(context.Background(), "cp-1", ssh); err == nil {
		t.Fatalf("TransferCertsToControlPlaneFromLocal() expected error, got nil")
	}
}

func TestLinuxRenewer_TransferCertsToControlPlaneFromLocal_CopyCertError(t *testing.T) {
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
		Return("", errString)

	if err := r.TransferCertsToControlPlaneFromLocal(context.Background(), "cp-1", ssh); err == nil {
		t.Fatalf("TransferCertsToControlPlaneFromLocal() expected error, got nil")
	}
}

func TestLinuxRenewer_TransferCertsToControlPlaneFromLocal_ReadKeyError(t *testing.T) {
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

	if err := r.TransferCertsToControlPlaneFromLocal(context.Background(), "cp-1", ssh); err == nil {
		t.Fatalf("TransferCertsToControlPlaneFromLocal() expected error, got nil")
	}
}

func TestLinuxRenewer_TransferCertsToControlPlaneFromLocal_CopyKeyError(t *testing.T) {
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
		Return("", errString)

	if err := r.TransferCertsToControlPlaneFromLocal(context.Background(), "cp-1", ssh); err == nil {
		t.Fatalf("TransferCertsToControlPlaneFromLocal() expected error, got nil")
	}
}
