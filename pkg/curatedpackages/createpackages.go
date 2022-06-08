package curatedpackages

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/utils/urls"
)

type PackageInstaller struct {
	chartInstaller   ChartInstaller
	kubectlRunner    KubectlRunner
	spec             *cluster.Spec
	packagesLocation string
}

func NewPackageInstaller(installer ChartInstaller, runner KubectlRunner,
	spec *cluster.Spec, packagesLocation string) *PackageInstaller {
	return &PackageInstaller{
		chartInstaller:   installer,
		kubectlRunner:    runner,
		spec:             spec,
		packagesLocation: packagesLocation,
	}
}

func (pi *PackageInstaller) InstallCuratedPackages(ctx context.Context) error {
	PrintLicense()
	err := pi.installPackagesController(ctx)
	if err != nil {
		logger.MarkFail("Error when installing curated packages on workload cluster; please install through eksctl anywhere install packagecontroller command", "error", err)
		return nil
	}

	if pi.packagesLocation != "" {
		err = pi.installPackages(ctx)
	}

	if err != nil {
		logger.MarkFail("Error when installing curated packages on workload cluster; please install through eksctl anywhere create packages command", "error", err)
	}
	return nil
}

func (pi *PackageInstaller) installPackagesController(ctx context.Context) error {
	logger.Info("Installing curated packages controller on workload cluster")
	kubeConfig := kubeconfig.FromClusterName(pi.spec.Cluster.Name)

	chart := pi.spec.VersionsBundle.VersionsBundle.PackageController.HelmChart
	imageUrl := urls.ReplaceHost(chart.Image(), pi.spec.Cluster.RegistryMirror())
	pc := NewPackageControllerClient(pi.chartInstaller, pi.kubectlRunner, kubeConfig, imageUrl, chart.Name, chart.Tag())
	err := pc.InstallController(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (pi *PackageInstaller) installPackages(ctx context.Context) error {
	kubeConfig := kubeconfig.FromClusterName(pi.spec.Cluster.Name)
	packageClient := NewPackageClient(
		nil,
		pi.kubectlRunner,
	)
	err := packageClient.CreatePackages(ctx, pi.packagesLocation, kubeConfig)
	if err != nil {
		return err
	}
	return nil
}
