package curatedpackages

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/logger"
)

type PackageController interface {
	InstallController(ctx context.Context) error
	IsInstalled(ctx context.Context) bool
}

type PackageHandler interface {
	CreatePackages(ctx context.Context, fileName string, kubeConfig string) error
}

type Installer struct {
	packageController PackageController
	spec              *cluster.Spec
	packageClient     PackageHandler
	packagesLocation  string
	kubectl           KubectlRunner
}

func NewInstaller(runner KubectlRunner, pc PackageHandler, pcc PackageController, spec *cluster.Spec, packagesLocation string) *Installer {
	return &Installer{
		spec:              spec,
		packagesLocation:  packagesLocation,
		packageController: pcc,
		packageClient:     pc,
		kubectl:           runner,
	}
}

func (pi *Installer) InstallCuratedPackages(ctx context.Context) error {
	PrintLicense()

	kubeConfig := kubeconfig.FromClusterName(pi.spec.Cluster.Name)

	err := VerifyCertManagerExists(ctx, pi.kubectl, kubeConfig)
	if err != nil {
		return err
	}

	err = pi.installPackagesController(ctx)
	if err != nil {
		logger.MarkFail("Error when installing curated packages on workload cluster; please install through eksctl anywhere install packagecontroller command", "error", err)
		return err
	}

	err = pi.installPackages(ctx)
	if err != nil {
		logger.MarkFail("Error when installing curated packages on workload cluster; please install through eksctl anywhere create packages command", "error", err)
		return err
	}

	return nil
}

func (pi *Installer) installPackagesController(ctx context.Context) error {
	logger.Info("Installing curated packages controller on management cluster")
	err := pi.packageController.InstallController(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (pi *Installer) installPackages(ctx context.Context) error {
	if pi.packagesLocation == "" {
		return nil
	}
	kubeConfig := kubeconfig.FromClusterName(pi.spec.Cluster.Name)
	err := pi.packageClient.CreatePackages(ctx, pi.packagesLocation, kubeConfig)
	if err != nil {
		return err
	}
	return nil
}
