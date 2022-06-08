package curatedpackages

import (
	"context"
	"errors"
	"fmt"
	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
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
	InstallChart(ctx context.Context, chart, ociURI, version, kubeconfigFilePath string, values []string) error
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
	ociUri := fmt.Sprintf("%s%s", "oci://", pc.uri)
	registry := GetRegistry(pc.uri)
	sourceRegistry := fmt.Sprintf("sourceRegistry=%s", registry)
	values := []string{sourceRegistry}
	return pc.chartInstaller.InstallChart(ctx, pc.chartName, ociUri, pc.chartVersion, pc.kubeConfig, values)
}

func (pc *PackageControllerClient) ValidateControllerDoesNotExist(ctx context.Context) error {
	found, _ := pc.kubectl.GetResource(ctx, "packageBundleController", packagesv1.PackageBundleControllerName, pc.kubeConfig, constants.EksaPackagesName)
	if found {
		return errors.New("curated Packages controller exists in the current cluster")
	}
	return nil
}
