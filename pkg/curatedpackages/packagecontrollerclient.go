package curatedpackages

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

//go:embed config/secrets.yaml
var secretsValueYaml string

//go:embed config/packagebundlecontroller.yaml
var packageBundleControllerYaml string

const (
	eksaDefaultRegion = "us-west-2"
	cronJobName       = "cronjob/cron-ecr-renew"
	jobName           = "eksa-auth-refresher"
	valueFileName     = "values.yaml"
)

type PackageControllerClientOpt func(client *PackageControllerClient)

type PackageControllerClient struct {
	kubeConfig            string
	chart                 *v1alpha1.Image
	chartInstaller        ChartInstaller
	clusterName           string
	managementClusterName string
	kubectl               KubectlRunner
	eksaAccessKeyID       string
	eksaSecretAccessKey   string
	eksaRegion            string
	httpProxy             string
	httpsProxy            string
	noProxy               []string
	registryMirror        *registrymirror.RegistryMirror
	// activeBundleTimeout is the timeout to activate a bundle on installation.
	activeBundleTimeout time.Duration
}

type ChartInstaller interface {
	InstallChart(ctx context.Context, chart, ociURI, version, kubeconfigFilePath, namespace, valueFilePath string, values []string) error
}

// NewPackageControllerClient instantiates a new instance of PackageControllerClient.
func NewPackageControllerClient(chartInstaller ChartInstaller, kubectl KubectlRunner, clusterName, kubeConfig string, chart *v1alpha1.Image, registryMirror *registrymirror.RegistryMirror, options ...PackageControllerClientOpt) *PackageControllerClient {
	pcc := &PackageControllerClient{
		kubeConfig:     kubeConfig,
		clusterName:    clusterName,
		chart:          chart,
		chartInstaller: chartInstaller,
		kubectl:        kubectl,
		registryMirror: registryMirror,
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
	ociURI := fmt.Sprintf("%s%s", "oci://", pc.registryMirror.ReplaceRegistry(pc.chart.Image()))
	var values []string
	clusterName := fmt.Sprintf("clusterName=%s", pc.clusterName)
	if pc.registryMirror != nil {
		// account is added as part of registry name in package controller helm chart
		// https://github.com/aws/eks-anywhere-packages/blob/main/charts/eks-anywhere-packages/values.yaml#L15-L18
		accountName := "eks-anywhere"
		if strings.Contains(ociURI, "l0g8r8j6") {
			accountName = "l0g8r8j6"
		}
		sourceRegistry := fmt.Sprintf("sourceRegistry=%s/%s", pc.registryMirror.CoreEKSAMirror(), accountName)
		defaultRegistry := fmt.Sprintf("defaultRegistry=%s/%s", pc.registryMirror.CoreEKSAMirror(), accountName)
		if gatedOCINamespace := pc.registryMirror.CuratedPackagesMirror(); gatedOCINamespace == "" {
			// no registry mirror for curated packages
			values = []string{sourceRegistry, defaultRegistry, clusterName}
		} else {
			defaultImageRegistry := fmt.Sprintf("defaultImageRegistry=%s", gatedOCINamespace)
			values = []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
		}
	} else {
		sourceRegistry := fmt.Sprintf("sourceRegistry=%s", GetRegistry(pc.chart.Image()))
		defaultImageRegistry := fmt.Sprintf("defaultImageRegistry=%s", strings.ReplaceAll(constants.DefaultCuratedPackagesRegistryRegex, "*", pc.eksaRegion))
		values = []string{sourceRegistry, defaultImageRegistry, clusterName}
	}

	// Provide proxy details for curated packages helm chart when proxy details provided
	if pc.httpProxy != "" {
		httpProxy := fmt.Sprintf("proxy.HTTP_PROXY=%s", pc.httpProxy)
		httpsProxy := fmt.Sprintf("proxy.HTTPS_PROXY=%s", pc.httpsProxy)

		// Helm requires commas to be escaped: https://github.com/rancher/rancher/issues/16195
		noProxy := fmt.Sprintf("proxy.NO_PROXY=%s", strings.Join(pc.noProxy, "\\,"))
		values = append(values, httpProxy, httpsProxy, noProxy)
	}
	if (pc.eksaSecretAccessKey == "" || pc.eksaAccessKeyID == "") && pc.registryMirror == nil {
		values = append(values, "cronjob.suspend=true")
	}

	writer, err := filewriter.NewWriter(pc.clusterName)
	defer writer.CleanUpTemp()
	if err != nil {
		return err
	}
	var valueFilePath string
	if valueFilePath, err = pc.CreateHelmOverrideValuesYaml(writer); err != nil {
		return err
	}

	if err := pc.chartInstaller.InstallChart(ctx, pc.chart.Name, ociURI, pc.chart.Tag(), pc.kubeConfig, "", valueFilePath, values); err != nil {
		return err
	}

	return pc.waitForActiveBundle(ctx)
}

// CreateHelmOverrideValuesYaml creates a temp file to override certain values in package controller helm install.
func (pc *PackageControllerClient) CreateHelmOverrideValuesYaml(writer filewriter.FileWriter) (string, error) {
	content, err := pc.GenerateHelmOverrideValues()
	if err != nil {
		return "", err
	}
	filePath, err := writer.Write(valueFileName, content)
	if err != nil {
		return "", err
	}
	return filePath, nil
}

// GenerateHelmOverrideValues generates override values.
func (pc *PackageControllerClient) GenerateHelmOverrideValues() ([]byte, error) {
	var err error
	endpoint, username, password, caCertContent := "", "", "", ""
	if pc.registryMirror != nil {
		endpoint = pc.registryMirror.BaseRegistry
		username, password, err = pc.registryMirror.Credentials()
		if err != nil {
			return []byte{}, err
		}
		caCertContent = pc.registryMirror.CACertContent
	}
	templateValues := map[string]interface{}{
		"eksaAccessKeyId":     base64.StdEncoding.EncodeToString([]byte(pc.eksaAccessKeyID)),
		"eksaSecretAccessKey": base64.StdEncoding.EncodeToString([]byte(pc.eksaSecretAccessKey)),
		"eksaRegion":          base64.StdEncoding.EncodeToString([]byte(pc.eksaRegion)),
		"mirrorEndpoint":      base64.StdEncoding.EncodeToString([]byte(endpoint)),
		"mirrorUsername":      base64.StdEncoding.EncodeToString([]byte(username)),
		"mirrorPassword":      base64.StdEncoding.EncodeToString([]byte(password)),
		"mirrorCACertContent": base64.StdEncoding.EncodeToString([]byte(caCertContent)),
	}
	result, err := templater.Execute(secretsValueYaml, templateValues)
	if err != nil {
		return []byte{}, err
	}
	return result, nil
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
		return fmt.Errorf("timed out finding an active package bundle for the current cluster: %v", timeoutCtx.Err())
	case err := <-done:
		if err != nil {
			return fmt.Errorf("couldn't find an active package bundle for the current cluster: %v", err)
		}

		return nil
	}
}

// IsInstalled checks if a package controller custom resource exists.
func (pc *PackageControllerClient) IsInstalled(ctx context.Context) bool {
	bool, err := pc.kubectl.HasResource(ctx, packageBundleControllerResource, pc.clusterName, pc.kubeConfig, constants.EksaPackagesName)
	return bool && err == nil
}

func WithEksaAccessKeyId(eksaAccessKeyId string) func(client *PackageControllerClient) {
	return func(config *PackageControllerClient) {
		config.eksaAccessKeyID = eksaAccessKeyId
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
