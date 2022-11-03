package curatedpackages

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
)

type PackageController interface {
	EnableCuratedPackages(ctx context.Context) error
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

func (pi *Installer) InstallCuratedPackages(ctx context.Context) {
	PrintLicense()
	err := pi.installPackagesController(ctx)

	// There is an ask from customers to avoid considering the installation of curated packages
	// controller as an error but rather a warning
	if err != nil {
		logger.MarkWarning("Failed enabling curated packages on the cluster; please install through eksctl anywhere install packagecontroller command", "error", err)
		return
	}

	// There is an ask from customers to avoid considering the installation of curated packages
	// as an error but rather a warning
	err = pi.installPackages(ctx)
	if err != nil {
		logger.MarkWarning("Failed installing curated packages on the cluster; please install through eksctl anywhere create packages command", "error", err)
	}
}

func (pi *Installer) installPackagesController(ctx context.Context) error {
	logger.Info("Installing curated packages controller on management cluster")
	err := pi.packageController.EnableCuratedPackages(ctx)
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
