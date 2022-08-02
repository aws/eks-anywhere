package curatedpackages

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/templater"
)

//go:embed config/awssecret.yaml
var awsSecretYaml string

const (
	eksaDefaultRegion = "us-west-2"
	cronJobName       = "cronjob/ecr-refresher"
	jobName           = "eksa-auth-refresher"
)

type PackageControllerClientOpt func(*PackageControllerClient)

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

func NewPackageControllerClient(chartInstaller ChartInstaller, kubectl KubectlRunner, kubeConfig, uri, chartName, chartVersion string, options ...PackageControllerClientOpt) *PackageControllerClient {
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
	err := pc.chartInstaller.InstallChart(ctx, pc.chartName, ociUri, pc.chartVersion, pc.kubeConfig, values)
	if err != nil {
		return err
	}

	err = pc.ApplySecret(ctx)
	if err != nil {
		logger.Info("Warning: not able to create secret. Package installation might fail.", "error", err)
	}

	err = pc.TriggerCronJob(ctx)
	if err != nil {
		logger.Info("Warning: not able to trigger cron job. Package installation might fail.", "error", err)
	}
	return nil
}

func (pc *PackageControllerClient) ValidateControllerDoesNotExist(ctx context.Context) error {
	found, _ := pc.kubectl.GetResource(ctx, "packageBundleController", packagesv1.PackageBundleControllerName, pc.kubeConfig, constants.EksaPackagesName)
	if found {
		return errors.New("curated Packages controller exists in the current cluster")
	}
	return nil
}

func (pc *PackageControllerClient) ApplySecret(ctx context.Context) error {
	eksaAccessKeyId, eksaSecretAccessKey := os.Getenv(config.EksaAccessKeyIdEnv), os.Getenv(config.EksaSecretAcessKeyEnv)
	eksaRegion := eksaDefaultRegion
	if region, found := os.LookupEnv(config.EksaRegionEnv); found {
		eksaRegion = region
	}
	templateValues := map[string]string{
		"eksaAccessKeyId":     eksaAccessKeyId,
		"eksaSecretAccessKey": eksaSecretAccessKey,
		"eksaRegion":          eksaRegion,
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

func (pc *PackageControllerClient) TriggerCronJob(ctx context.Context) error {
	params := []string{"create", "job", jobName, "--from=" + cronJobName, "--kubeconfig", pc.kubeConfig, "--namespace", constants.EksaPackagesName}
	stdOut, err := pc.kubectl.ExecuteCommand(ctx, params...)
	if err != nil {
		return fmt.Errorf("executing cron job %v", err)
	}
	fmt.Println(stdOut)
	return nil
}
