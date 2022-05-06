package curatedpackages

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere/pkg/constants"
)

type PackageControllerClient struct {
	kubeConfig     string
	uri            string
	chartName      string
	chartVersion   string
	chartInstaller ChartInstaller
	kubectl        KubectlRunner
}

type ChartInstaller interface {
	InstallChartFromName(ctx context.Context, ociURI, kubeConfig, name, version string) error
}

func NewPackageControllerClient(chartInstaller ChartInstaller, kubectl KubectlRunner, kubeConfig, uri, chartName, chartVersion string) *PackageControllerClient {
	return &PackageControllerClient{
		kubeConfig:     kubeConfig,
		uri:            uri,
		chartName:      chartName,
		chartVersion:   chartVersion,
		chartInstaller: chartInstaller,
		kubectl:        kubectl,
	}
}

func (pc *PackageControllerClient) InstallController(ctx context.Context) error {
	PrintLicense()
	ociUri := fmt.Sprintf("%s%s", "oci://", pc.uri)
	return pc.chartInstaller.InstallChartFromName(ctx, ociUri, pc.kubeConfig, pc.chartName, pc.chartVersion)
}

func (pc *PackageControllerClient) ValidateControllerDoesNotExist(ctx context.Context) error {
	found, _ := pc.kubectl.GetResource(ctx, "packageBundleController", bundle.PackageBundleControllerName, pc.kubeConfig, constants.EksaPackagesName)
	if found {
		return errors.New("curated Packages controller exists in the current cluster")
	}
	return nil
}
