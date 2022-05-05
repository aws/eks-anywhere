package curatedpackages

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere/pkg/constants"
)

type PackageControllerClient struct {
	kubeConfig     string
	ociUri         string
	chartName      string
	chartVersion   string
	chartInstaller ChartInstaller
	kubectl        KubectlRunner
}

type ChartInstaller interface {
	InstallChartFromName(ctx context.Context, ociURI, kubeConfig, name, version string) error
}

func NewPackageControllerClient(chartInstaller ChartInstaller, kubectl KubectlRunner, kubeConfig, ociUri, chartName, chartVersion string) *PackageControllerClient {
	return &PackageControllerClient{
		kubeConfig:     kubeConfig,
		ociUri:         ociUri,
		chartName:      chartName,
		chartVersion:   chartVersion,
		chartInstaller: chartInstaller,
		kubectl:        kubectl,
	}
}

func (pc *PackageControllerClient) InstallController(ctx context.Context) error {
	uri := fmt.Sprintf("%s%s", "oci://", pc.ociUri)
	return pc.chartInstaller.InstallChartFromName(ctx, uri, pc.kubeConfig, pc.chartName, pc.chartVersion)
}

func (pc *PackageControllerClient) ControllerExists(ctx context.Context) bool {
	err := pc.GetActiveController(ctx)
	if err != nil {
		return false
	}
	return true
}

func (pc *PackageControllerClient) GetActiveController(ctx context.Context) error {
	params := []string{"get", "packageBundleController", "--kubeconfig", pc.kubeConfig, "--namespace", constants.EksaPackagesName, bundle.PackageBundleControllerName}
	_, err := pc.kubectl.ExecuteCommand(ctx, params...)
	if err != nil {
		return err
	}
	return nil
}
