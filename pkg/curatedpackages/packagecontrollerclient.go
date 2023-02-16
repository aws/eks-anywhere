package curatedpackages

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

//go:embed config/secrets.yaml
var secretsValueYaml string

const (
	eksaDefaultRegion = "us-west-2"
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
	valuesFileWriter    filewriter.FileWriter
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
	ociURI := fmt.Sprintf("%s%s", "oci://", pc.registryMirror.ReplaceRegistry(pc.chart.Image()))
	clusterName := fmt.Sprintf("clusterName=%s", pc.clusterName)
	sourceRegistry, defaultRegistry, defaultImageRegistry := pc.GetCuratedPackagesRegistries()
	sourceRegistry = fmt.Sprintf("sourceRegistry=%s", sourceRegistry)
	defaultRegistry = fmt.Sprintf("defaultRegistry=%s", defaultRegistry)
	defaultImageRegistry = fmt.Sprintf("defaultImageRegistry=%s", defaultImageRegistry)
	values := []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}

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

	var err error
	var valueFilePath string
	if valueFilePath, _, err = pc.CreateHelmOverrideValuesYaml(); err != nil {
		return err
	}

	chartName := pc.chart.Name
	if pc.managementClusterName != pc.clusterName {
		values = append(values, "workloadOnly=true")
		chartName = chartName + "-" + pc.clusterName
	}

	if err := pc.chartInstaller.InstallChart(ctx, chartName, ociURI, pc.chart.Tag(), pc.kubeConfig, "", valueFilePath, values); err != nil {
		return err
	}

	return pc.waitForActiveBundle(ctx)
}

// GetCuratedPackagesRegistries gets value for configurable registries from PBC.
func (pc *PackageControllerClient) GetCuratedPackagesRegistries() (sourceRegistry, defaultRegistry, defaultImageRegistry string) {
	sourceRegistry = publicProdECR
	defaultImageRegistry = packageProdDomain
	accountName := prodAccount
	if strings.Contains(pc.chart.Image(), devAccount) {
		accountName = devAccount
		defaultImageRegistry = packageDevDomain
		sourceRegistry = publicDevECR
	}
	if strings.Contains(pc.chart.Image(), stagingAccount) {
		accountName = stagingAccount
		defaultImageRegistry = packageProdDomain
		sourceRegistry = stagingDevECR
	}
	defaultRegistry = sourceRegistry

	if pc.registryMirror != nil {
		// account is added as part of registry name in package controller helm chart
		// https://github.com/aws/eks-anywhere-packages/blob/main/charts/eks-anywhere-packages/values.yaml#L15-L18
		sourceRegistry = fmt.Sprintf("%s/%s", pc.registryMirror.CoreEKSAMirror(), accountName)
		defaultRegistry = fmt.Sprintf("%s/%s", pc.registryMirror.CoreEKSAMirror(), accountName)
		if gatedOCINamespace := pc.registryMirror.CuratedPackagesMirror(); gatedOCINamespace != "" {
			defaultImageRegistry = gatedOCINamespace
		}
	} else {
		defaultImageRegistry = strings.ReplaceAll(defaultImageRegistry, defaultRegion, pc.eksaRegion)
	}
	return sourceRegistry, defaultRegistry, defaultImageRegistry
}

// CreateHelmOverrideValuesYaml creates a temp file to override certain values in package controller helm install.
func (pc *PackageControllerClient) CreateHelmOverrideValuesYaml() (string, []byte, error) {
	content, err := pc.generateHelmOverrideValues()
	if err != nil {
		return "", nil, err
	}
	if pc.valuesFileWriter == nil {
		return "", content, fmt.Errorf("valuesFileWriter is nil")
	}
	filePath, err := pc.valuesFileWriter.Write(valueFileName, content)
	if err != nil {
		return "", content, err
	}
	return filePath, content, nil
}

func (pc *PackageControllerClient) generateHelmOverrideValues() ([]byte, error) {
	var err error
	endpoint, username, password, caCertContent, insecureSkipVerify := "", "", "", "", "false"
	if pc.registryMirror != nil {
		endpoint = pc.registryMirror.BaseRegistry
		username, password, err = config.ReadCredentials()
		if err != nil {
			return []byte{}, err
		}
		caCertContent = pc.registryMirror.CACertContent
		if pc.registryMirror.InsecureSkipVerify {
			insecureSkipVerify = "true"
		}
	}
	templateValues := map[string]interface{}{
		"eksaAccessKeyId":     base64.StdEncoding.EncodeToString([]byte(pc.eksaAccessKeyID)),
		"eksaSecretAccessKey": base64.StdEncoding.EncodeToString([]byte(pc.eksaSecretAccessKey)),
		"eksaRegion":          base64.StdEncoding.EncodeToString([]byte(pc.eksaRegion)),
		"mirrorEndpoint":      base64.StdEncoding.EncodeToString([]byte(endpoint)),
		"mirrorUsername":      base64.StdEncoding.EncodeToString([]byte(username)),
		"mirrorPassword":      base64.StdEncoding.EncodeToString([]byte(password)),
		"mirrorCACertContent": base64.StdEncoding.EncodeToString([]byte(caCertContent)),
		"insecureSkipVerify":  base64.StdEncoding.EncodeToString([]byte(insecureSkipVerify)),
	}
	result, err := templater.Execute(secretsValueYaml, templateValues)
	if err != nil {
		return []byte{}, err
	}
	return result, nil
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
	timeout := 3 * time.Minute
	if pc.activeBundleTimeout > 0 {
		timeout = pc.activeBundleTimeout
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	targetNs := constants.EksaPackagesName + "-" + pc.clusterName
	done := make(chan error)
	go func() {
		defer close(done)
		pbc := &packagesv1.PackageBundleController{}
		for {
			readyCnt := 0
			err := pc.kubectl.GetObject(timeoutCtx, packageBundleControllerResource, pc.clusterName,
				packagesv1.PackageNamespace, pc.kubeConfig, pbc)
			if err != nil {
				done <- fmt.Errorf("getting package bundle controller: %w", err)
				return
			}

			if pbc.Spec.ActiveBundle != "" {
				logger.V(6).Info("found packages bundle controller active bundle",
					"name", pbc.Spec.ActiveBundle)
				readyCnt++
			} else {
				logger.V(6).Info("waiting for package bundle controller to activate a bundle",
					"clusterName", pc.clusterName)
			}

			found, _ := pc.kubectl.HasResource(timeoutCtx, "namespace", targetNs, pc.kubeConfig, "default")

			if found {
				logger.V(6).Info("found namespace", "namespace", targetNs)
				readyCnt++
			} else {
				logger.V(6).Info("waiting for namespace", "namespace", targetNs)
			}

			if readyCnt == 2 {
				return
			}

			// TODO read a polling interval value from the context, falling
			// back to this as a default.
			time.Sleep(time.Second)
		}
	}()

	select {
	case <-timeoutCtx.Done():
		return fmt.Errorf("timed out finding an active package bundle / %s namespace for the current cluster: %v", targetNs, timeoutCtx.Err())
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

// WithValuesFileWriter sets up a writer to generate temporary values.yaml to
// override some values in package controller helm chart.
func WithValuesFileWriter(writer filewriter.FileWriter) func(client *PackageControllerClient) {
	return func(config *PackageControllerClient) {
		config.valuesFileWriter = writer
	}
}
