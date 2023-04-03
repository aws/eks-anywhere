package curatedpackages

import (
	"context"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
)

type PackageController interface {
	// Enable curated packages support.
	Enable(ctx context.Context) error
	IsInstalled(ctx context.Context) bool
}

type PackageHandler interface {
	CreatePackages(ctx context.Context, fileName string, kubeConfig string) error
}

type Installer struct {
	packageController PackageController
	spec              *cluster.Spec
	packageClient     PackageHandler
	kubectl           KubectlRunner
	packagesLocation  string
	mgmtKubeconfig    string
}

// IsPackageControllerDisabled detect if the package controller is disabled.
func IsPackageControllerDisabled(cluster *anywherev1.Cluster) bool {
	return cluster != nil && cluster.Spec.Packages != nil && cluster.Spec.Packages.Disable
}

// NewInstaller installs packageController and packages during cluster creation.
func NewInstaller(runner KubectlRunner, pc PackageHandler, pcc PackageController, spec *cluster.Spec, packagesLocation, mgmtKubeconfig string) *Installer {
	return &Installer{
		spec:              spec,
		packagesLocation:  packagesLocation,
		packageController: pcc,
		packageClient:     pc,
		kubectl:           runner,
		mgmtKubeconfig:    mgmtKubeconfig,
	}
}

// InstallCuratedPackages installs curated packages as part of the cluster creation.
func (pi *Installer) InstallCuratedPackages(ctx context.Context) {
	if IsPackageControllerDisabled(pi.spec.Cluster) {
		logger.Info("  Package controller disabled")
		return
	}
	PrintLicense()
	err := pi.installPackagesController(ctx)
	// There is an ask from customers to avoid considering the failure of installing curated packages
	// controller as an error but rather a warning
	if err != nil {
		logger.MarkWarning("  Failed to install the optional EKS-A Curated Package Controller. Please try installation again through eksctl after the cluster creation succeeds", "warning", err)
		return
	}

	// There is an ask from customers to avoid considering the failure of the installation of curated packages
	// as an error but rather a warning
	err = pi.installPackages(ctx)
	if err != nil {
		logger.MarkWarning("  Failed installing curated packages on the cluster; please install through eksctl anywhere create packages command after the cluster creation succeeds", "error", err)
	}
}

func (pi *Installer) installPackagesController(ctx context.Context) error {
	logger.Info("Enabling curated packages on the cluster")
	err := pi.packageController.Enable(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (pi *Installer) installPackages(ctx context.Context) error {
	if pi.packagesLocation == "" {
		return nil
	}
	err := pi.packageClient.CreatePackages(ctx, pi.packagesLocation, pi.mgmtKubeconfig)
	if err != nil {
		return err
	}
	return nil
}
