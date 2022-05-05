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

func (pc *PackageControllerClient) ValidateControllerExists(ctx context.Context) error {
	params := []string{"get", "packageBundleController", "--kubeconfig", pc.kubeConfig, "--namespace", constants.EksaPackagesName, bundle.PackageBundleControllerName}
	stdOut, err := pc.kubectl.ExecuteCommand(ctx, params...)
	if err != nil {
		return err
	}
	if len(stdOut.Bytes()) != 0 {
		return errors.New("curated Packages Controller Exists in the current Cluster")
	}
	return nil
}
