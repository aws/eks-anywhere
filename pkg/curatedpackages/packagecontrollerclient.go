package curatedpackages

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/aws/eks-anywhere/pkg/templater"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

//go:embed config/awssecret.yaml
var awsSecretYaml string

const (
	eksaDefaultRegion = "us-west-2"
)

type PackageControllerClientOpt func(*PackageControllerClient)

type PackageControllerClient struct {
	kubeConfig          string
	uri                 string
	chartName           string
	chartVersion        string
	chartInstaller      ChartInstaller
	kubectl             KubectlRunner
	eksaAccessKeyId     string
	eksaSecretAccessKey string
	eksaRegion          string
}

type ChartInstaller interface {
	InstallChart(ctx context.Context, chart, ociURI, version, kubeconfigFilePath string, values []string) error
}

func NewPackageControllerClient(chartInstaller ChartInstaller, kubectl KubectlRunner, kubeConfig, uri, chartName, chartVersion string, options ...PackageControllerClientOpt) *PackageControllerClient {
	pcc := &PackageControllerClient{
		kubeConfig:     kubeConfig,
		uri:            uri,
		chartName:      chartName,
		chartVersion:   chartVersion,
		chartInstaller: chartInstaller,
		kubectl:        kubectl,
	}
	for _, o := range options {
		o(pcc)
	}
	return pcc
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

func (pc *PackageControllerClient) ApplySecret(ctx context.Context) error {
	templateValues := map[string]string{
		"eksaAccessKeyId":     pc.eksaAccessKeyId,
		"eksaSecretAccessKey": pc.eksaSecretAccessKey,
		"eksaRegion":          pc.eksaRegion,
	}

	result, err := templater.Execute(awsSecretYaml, templateValues)
	if err != nil {
		return fmt.Errorf("replacing template values %v", err)
	}

	params := []string{"create", "-f", "-", "--kubeconfig", pc.kubeConfig}
	stdOut, err := pc.kubectl.CreateFromYaml(ctx, result, params...)
	if err != nil {
		return fmt.Errorf("creating secret %v", err)
	}

	fmt.Print(&stdOut)
	return nil
}

func WithEKSAAccessKeyId(eksaAccessKeyId string) func(client *PackageControllerClient) {
	return func(config *PackageControllerClient) {
		config.eksaAccessKeyId = eksaAccessKeyId
	}
}

func WithEKSASecretAccessKey(eksaSecretAccessKey string) func(client *PackageControllerClient) {
	return func(config *PackageControllerClient) {
		config.eksaSecretAccessKey = eksaSecretAccessKey
	}
}

func WithEksaRegion(eksaRegion string) func(client *PackageControllerClient) {
	return func(config *PackageControllerClient) {
		if eksaRegion == "" {
			config.eksaRegion = eksaDefaultRegion
		} else {
			config.eksaRegion = eksaRegion
		}
	}
}
