package curatedpackages

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/utils/urls"
)

type PackageController interface {
	InstallController(ctx context.Context) error
}

type PackageHandler interface {
	CreatePackages(ctx context.Context, fileName string, kubeConfig string) error
}

type Installer struct {
	packageController PackageController
	spec              *cluster.Spec
	packageClient     PackageHandler
	packagesLocation  string
}

func NewInstaller(installer ChartInstaller, runner KubectlRunner, spec *cluster.Spec, packagesLocation string) *Installer {
	pcc := newPackageController(installer, runner, spec)

	pc := NewPackageClient(
		runner,
	)

	return &Installer{
		spec:              spec,
		packagesLocation:  packagesLocation,
		packageController: pcc,
		packageClient:     pc,
	}
}

func newPackageController(installer ChartInstaller, runner KubectlRunner, spec *cluster.Spec) *PackageControllerClient {
	kubeConfig := kubeconfig.FromClusterName(spec.Cluster.Name)

	chart := spec.VersionsBundle.PackageController.HelmChart
	imageUrl := urls.ReplaceHost(chart.Image(), spec.Cluster.RegistryMirror())
	return NewPackageControllerClient(installer, runner, kubeConfig, imageUrl, chart.Name, chart.Tag())
}

func (pi *Installer) InstallCuratedPackages(ctx context.Context) error {
	PrintLicense()
	err := pi.installPackagesController(ctx)
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
