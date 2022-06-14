package curatedpackages

import (
	"context"
	"fmt"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/utils/urls"
)

type Installer struct {
	chartInstaller   ChartInstaller
	kubectlRunner    KubectlRunner
	spec             *cluster.Spec
	packagesLocation string
}

func NewInstaller(installer ChartInstaller, runner KubectlRunner,
	spec *cluster.Spec, packagesLocation string) *Installer {

	return &Installer{
		chartInstaller:   installer,
		kubectlRunner:    runner,
		spec:             spec,
		packagesLocation: packagesLocation,
	}
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
	kubeConfig := kubeconfig.FromClusterName(pi.spec.Cluster.Name)

	versionsBundle := pi.spec.VersionsBundle
	if versionsBundle == nil {
		return fmt.Errorf("unknown release bundle")
	}
	chart := versionsBundle.PackageController.HelmChart
	imageUrl := urls.ReplaceHost(chart.Image(), pi.spec.Cluster.RegistryMirror())
	pc := NewPackageControllerClient(pi.chartInstaller, pi.kubectlRunner, kubeConfig, imageUrl, chart.Name, chart.Tag())
	err := pc.InstallController(ctx)
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
