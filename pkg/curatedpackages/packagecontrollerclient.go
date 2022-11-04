package curatedpackages

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/templater"
)

//go:embed config/awssecret.yaml
var awsSecretYaml string

//go:embed config/packagebundlecontroller.yaml
var packageBundleControllerYaml string

const (
	eksaDefaultRegion = "us-west-2"
	cronJobName       = "cronjob/cron-ecr-renew"
	jobName           = "eksa-auth-refresher"
	packagesNamespace = "eksa-packages"
)

type PackageControllerClientOpt func(client *PackageControllerClient)

type PackageControllerClient struct {
	kubeConfig            string
	uri                   string
	chartName             string
	chartVersion          string
	chartInstaller        ChartInstaller
	clusterName           string
	managementClusterName string
	kubectl               KubectlRunner
	eksaAccessKeyId       string
	eksaSecretAccessKey   string
	eksaRegion            string
	httpProxy             string
	httpsProxy            string
	noProxy               []string
	// activeBundleTimeout is the timeout to activate a bundle on installation.
	activeBundleTimeout time.Duration
}

type ChartInstaller interface {
	InstallChart(ctx context.Context, chart, ociURI, version, kubeconfigFilePath, namespace string, values []string) error
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

// EnableCuratedPackages enables curated packages in a cluster
// In case the cluster is management cluster, it performs the following actions:
//   - Installation of Package Controller through helm chart installation
//   - Creation of secret credentials
//   - Creation of a single run of a cron job refresher
//   - Activation of a curated packages bundle
//
// In case the cluster is a workload cluster, it performs the following actions:
//   - Creation of package bundle controller (PBC) custom resource in management cluster
func (pc *PackageControllerClient) EnableCuratedPackages(ctx context.Context) error {
	// When the management cluster name and current cluster name are different,
	// it indicates that we are trying to install the controller on a workload cluster.
	// Instead of installing the controller, install packagebundlecontroller resource.
	if pc.managementClusterName != pc.clusterName {
		return pc.InstallPBCResources(ctx)
	}
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

	err := pc.chartInstaller.InstallChart(ctx, pc.chartName, ociUri, pc.chartVersion, pc.kubeConfig, "", values)
	if err != nil {
		return err
	}

	if err = pc.ApplySecret(ctx); err != nil {
		logger.Info("Warning: No AWS key/license provided. Please be aware this might prevent the package controller from installing curated packages.")
	}

	if err = pc.CreateCronJob(ctx); err != nil {
		logger.Info("Warning: not able to trigger cron job, please be aware this will prevent the package controller from installing curated packages.")
	}

	if err := pc.waitForActiveBundle(ctx); err != nil {
		return err
	}

	return nil
}

// InstallPBCResources installs Curated Packages Bundle Controller Custom Resource
// This method is used only for Workload clusters
// Please refer to this documentation: https://github.com/aws/eks-anywhere-packages/blob/main/docs/design/remote-management.md
func (pc *PackageControllerClient) InstallPBCResources(ctx context.Context) error {
	templateValues := map[string]string{
		"clusterName": pc.clusterName,
		"namespace":   constants.EksaPackagesName,
	}

	result, err := templater.Execute(packageBundleControllerYaml, templateValues)
	if err != nil {
		return fmt.Errorf("replacing template values %v", err)
	}

	params := []string{"create", "-f", "-", "--kubeconfig", pc.kubeConfig}
	stdOut, err := pc.kubectl.ExecuteFromYaml(ctx, result, params...)
	if err != nil {
		return fmt.Errorf("creating package bundle controller custom resource%v", err)
	}

	fmt.Print(&stdOut)
	return nil
}

// packageBundleControllerResource is the name of the package bundle controller
// resource in the API.
const packageBundleControllerResource string = "packageBundleController"

// waitForActiveBundle polls the package bundle controller for its active bundle.
//
// It returns nil on success. Success is defined as receiving a valid package
// bundle controller from the API with a non-empty active bundle.
//
// If no timeout is specified, a default of 1 minute is used.
func (pc *PackageControllerClient) waitForActiveBundle(ctx context.Context) error {
	timeout := time.Minute
	if pc.activeBundleTimeout > 0 {
		timeout = pc.activeBundleTimeout
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error)
	go func() {
		defer close(done)
		pbc := &packagesv1.PackageBundleController{}
		for {
			err := pc.kubectl.GetObject(timeoutCtx, packageBundleControllerResource, pc.clusterName,
				packagesv1.PackageNamespace, pc.kubeConfig, pbc)
			if err != nil {
				done <- fmt.Errorf("getting package bundle controller: %w", err)
				return
			}

			if pbc.Spec.ActiveBundle != "" {
				logger.V(6).Info("found packages bundle controller active bundle",
					"name", pbc.Spec.ActiveBundle)
				return
			}

			logger.V(6).Info("waiting for package bundle controller to activate a bundle",
				"clusterName", pc.clusterName)
			// TODO read a polling interval value from the context, falling
			// back to this as a default.
			time.Sleep(time.Second)
		}
	}()

	select {
	case <-timeoutCtx.Done():
		return timeoutCtx.Err()
	case err := <-done:
		if err != nil {
			return err
		}

		return nil
	}
}

// IsInstalled checks if a package controller custom resource exists.
func (pc *PackageControllerClient) IsInstalled(ctx context.Context) bool {
	bool, err := pc.kubectl.HasResource(ctx, packageBundleControllerResource, pc.clusterName, pc.kubeConfig, constants.EksaPackagesName)
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

func WithActiveBundleTimeout(timeout time.Duration) func(client *PackageControllerClient) {
	return func(config *PackageControllerClient) {
		config.activeBundleTimeout = timeout
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

func WithManagementClusterName(managementClusterName string) func(client *PackageControllerClient) {
	return func(config *PackageControllerClient) {
		config.managementClusterName = managementClusterName
	}
}
