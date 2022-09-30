package curatedpackages

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/templater"
)

//go:embed config/awssecret.yaml
var awsSecretYaml string

const (
	eksaDefaultRegion = "us-west-2"
	cronJobName       = "cronjob/cron-ecr-renew"
	jobName           = "eksa-auth-refresher"
)

type PackageControllerClientOpt func(client *PackageControllerClient)

type PackageControllerClient struct {
	kubeConfig          string
	uri                 string
	chartName           string
	chartVersion        string
	chartInstaller      ChartInstaller
	clusterName         string
	kubectl             KubectlRunner
	eksaAccessKeyId     string
	eksaSecretAccessKey string
	eksaRegion          string
	httpProxy           string
	httpsProxy          string
	noProxy             []string
}

type ChartInstaller interface {
	InstallChart(ctx context.Context, chart, ociURI, version, kubeconfigFilePath string, values []string) error
}

func NewPackageControllerClient(chartInstaller ChartInstaller, kubectl KubectlRunner, clusterName, kubeConfig, uri, chartName, chartVersion string, options ...PackageControllerClientOpt) *PackageControllerClient {
	pcc := &PackageControllerClient{
		kubeConfig:     kubeConfig,
		clusterName:    clusterName,
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
	clusterName := fmt.Sprintf("clusterName=%s", pc.clusterName)
	values := []string{sourceRegistry, clusterName}

	// Provide proxy details for curated packages helm chart when proxy details provided
	if pc.httpProxy != "" {
		httpProxy := fmt.Sprintf("proxy.HTTP_PROXY=%s", pc.httpProxy)
		httpsProxy := fmt.Sprintf("proxy.HTTPS_PROXY=%s", pc.httpsProxy)

		// Helm requires commas to be escaped: https://github.com/rancher/rancher/issues/16195
		noProxy := fmt.Sprintf("proxy.NO_PROXY=%s", strings.Join(pc.noProxy, "\\,"))
		values = append(values, httpProxy, httpsProxy, noProxy)
	}

	err := pc.chartInstaller.InstallChart(ctx, pc.chartName, ociUri, pc.chartVersion, pc.kubeConfig, values)
	if err != nil {
		return err
	}

	if err = pc.ApplySecret(ctx); err != nil {
		logger.Info("Warning: No AWS key/license provided. Please be aware this will prevent the package controller from installing curated packages.")
	}

	if err = pc.CreateCronJob(ctx); err != nil {
		logger.Info("Warning: not able to trigger cron job, please be aware this will prevent the package controller from installing curated packages.")
	}
	return nil
}

// IsInstalled checks if a package controller custom resource exists.
func (pc *PackageControllerClient) IsInstalled(ctx context.Context) bool {
	bool, err := pc.kubectl.GetResource(ctx, "packageBundleController", pc.clusterName, pc.kubeConfig, constants.EksaPackagesName)
	return bool && err == nil
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
	stdOut, err := pc.kubectl.ExecuteFromYaml(ctx, result, params...)
	if err != nil {
		return fmt.Errorf("creating secret %v", err)
	}

	fmt.Print(&stdOut)
	return nil
}

func (pc *PackageControllerClient) CreateCronJob(ctx context.Context) error {
	params := []string{"create", "job", jobName, "--from=" + cronJobName, "--kubeconfig", pc.kubeConfig, "--namespace", constants.EksaPackagesName}
	stdOut, err := pc.kubectl.ExecuteCommand(ctx, params...)
	if err != nil {
		return fmt.Errorf("executing cron job %v", err)
	}
	fmt.Print(&stdOut)
	return nil
}

func WithEksaAccessKeyId(eksaAccessKeyId string) func(client *PackageControllerClient) {
	return func(config *PackageControllerClient) {
		config.eksaAccessKeyId = eksaAccessKeyId
	}
}

func WithEksaSecretAccessKey(eksaSecretAccessKey string) func(client *PackageControllerClient) {
	return func(config *PackageControllerClient) {
		config.eksaSecretAccessKey = eksaSecretAccessKey
	}
}

func WithEksaRegion(eksaRegion string) func(client *PackageControllerClient) {
	return func(config *PackageControllerClient) {
		if eksaRegion != "" {
			config.eksaRegion = eksaRegion
		} else {
			config.eksaRegion = eksaDefaultRegion
		}
	}
}

func WithHTTPProxy(httpProxy string) func(client *PackageControllerClient) {
	return func(config *PackageControllerClient) {
		config.httpProxy = httpProxy
	}
}

func WithHTTPSProxy(httpsProxy string) func(client *PackageControllerClient) {
	return func(config *PackageControllerClient) {
		config.httpsProxy = httpsProxy
	}
}

func WithNoProxy(noProxy []string) func(client *PackageControllerClient) {
	return func(config *PackageControllerClient) {
		if noProxy != nil {
			config.noProxy = noProxy
		}
	}
}
