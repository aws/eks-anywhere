package framework

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/networking/cilium"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

// WithSkipCiliumUpgrade returns an E2E test option that configures the Cluster object to
// skip Cilium upgrades.
func WithSkipCiliumUpgrade() ClusterE2ETestOpt {
	return WithClusterFiller(func(cluster *v1alpha1.Cluster) {
		cluster.Spec.ClusterNetwork.CNIConfig.Cilium.SkipUpgrade = ptr.Bool(true)
	})
}

// UninstallCilium uninstalls the workload clusters Cilium.
func (e *ClusterE2ETest) UninstallCilium() {
	e.ValidateCiliumCLIAvailable()

	cmd := exec.Command("cilium", "uninstall")
	cmd.Env = append(cmd.Env, "KUBECONFIG="+e.KubeconfigFilePath())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	e.T.Log("Uninstalling Cilium using Cilium CLI")
	if err := cmd.Run(); err != nil {
		e.T.Fatal(err)
	}
}

// ValidateCiliumCLIAvailable ensures the Cilium CLI can be found on the PATH.
func (e *ClusterE2ETest) ValidateCiliumCLIAvailable() {
	if _, err := exec.LookPath("cilium"); err != nil {
		e.T.Fatal("Cilium CLI is required to run these tests (https://github.com/cilium/cilium-cli).")
	}
}

// InstallOSSCilium installs an open source version of Cilium. The version is dependent on the
// Cilium CLI version available on the PATH.
func (e *ClusterE2ETest) InstallOSSCilium() {
	e.ValidateCiliumCLIAvailable()

	cmd := exec.Command("cilium", "install")
	cmd.Env = append(cmd.Env, "KUBECONFIG="+e.KubeconfigFilePath())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	e.T.Log("Installing OSS Cilium using Cilium CLI")
	if err := cmd.Run(); err != nil {
		e.T.Fatal(err)
	}
}

// ReplaceCiliumWithOSSCilium replaces the current Cilium installation in the workload cluster
// with an open source version. See InstallOSSCilium().
func (e *ClusterE2ETest) ReplaceCiliumWithOSSCilium() {
	e.UninstallCilium()
	e.InstallOSSCilium()
}

// ValidateEKSACiliumNotInstalled inspects the workload cluster for an EKSA Cilium installation
// erroring if one is found.
func (e *ClusterE2ETest) ValidateEKSACiliumNotInstalled() {
	client, err := buildClusterClient(e.KubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("Constructing client: %v", err)
	}

	install, err := cilium.GetInstallation(context.Background(), client)
	if err != nil {
		e.T.Fatalf("Getting Cilium installation: %v", err)
	}

	if install.Installed() {
		e.T.Fatal("Unexpected Cilium install found in the workload cluster")
	}
}

// ValidateEKSACiliumInstalled inspects the workload cluster for an EKSA Cilium installation
// erroring if one is not found.
func (e *ClusterE2ETest) ValidateEKSACiliumInstalled() {
	e.T.Logf("Checking for EKSA Cilium installation with %v", e.KubeconfigFilePath())
	client, err := buildClusterClient(e.KubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("Constructing client: %v", err)
	}

	install, err := cilium.GetInstallation(context.Background(), client)
	if err != nil {
		e.T.Fatalf("Getting Cilium installation: %v", err)
	}

	if !install.Installed() {
		e.T.Fatal("Expected EKSA Cilium to be installed but found nothing")
	}
}

// AwaitCiliumDaemonSetReady awaits the Cilium daemonset to be ready in the cluster represented by client.
// It is ready when the DaemonSet's .Status.NumberUnavailable is 0.
func AwaitCiliumDaemonSetReady(ctx context.Context, client client.Client, retries int, timeout time.Duration) error {
	return retrier.Retry(12, timeout, func() error {
		installation, err := cilium.GetInstallation(ctx, client)
		if err != nil {
			return err
		}

		if installation.DaemonSet == nil {
			return errors.New("cilium DaemonSet not found")
		}

		if installation.DaemonSet.Status.NumberUnavailable != 0 {
			return errors.New("DaemonSet not ready")
		}

		return nil
	})
}
