package curatedpackages

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"github.com/aws/eks-anywhere/pkg/templater"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

//go:embed config/secrets.yaml
var secretsValueYaml string

const (
	eksaDefaultRegion = "us-west-2"
	valueFileName     = "values.yaml"
)

type PackageControllerClientOpt func(client *PackageControllerClient)

type PackageControllerClient struct {
	kubeConfig string
	chart      *releasev1.Image
	// chartManager installs and deletes helm charts.
	chartManager          ChartManager
	clusterName           string
	clusterSpec           *anywherev1.ClusterSpec
	managementClusterName string
	kubectl               KubectlRunner
	eksaAccessKeyID       string
	eksaSecretAccessKey   string
	eksaRegion            string
	eksaAwsConfig         string
	httpProxy             string
	httpsProxy            string
	noProxy               []string
	registryMirror        *registrymirror.RegistryMirror
	// activeBundleTimeout is the timeout to activate a bundle on installation.
	activeBundleTimeout time.Duration
	valuesFileWriter    filewriter.FileWriter
	// skipWaitForPackageBundle indicates whether the installer should wait
	// until a package bundle is activated.
	//
	// Skipping the wait is desirable for full cluster lifecycle use cases,
	// where resource creation and error reporting are asynchronous in nature.
	skipWaitForPackageBundle bool
	// tracker creates k8s clients for workload clusters managed via full
	// cluster lifecycle API.
	clientBuilder ClientBuilder

	// mu provides some thread-safety.
	mu sync.Mutex

	// registryAccessTester test if the aws credential has access to registry
	registryAccessTester RegistryAccessTester
}

// ClientBuilder returns a k8s client for the specified cluster.
type ClientBuilder interface {
	GetClient(context.Context, types.NamespacedName) (client.Client, error)
}

type ChartInstaller interface {
	InstallChart(ctx context.Context, chart, ociURI, version, kubeconfigFilePath, namespace, valueFilePath string, skipCRDs bool, values []string) error
}

// ChartUninstaller handles deleting chart installations.
type ChartUninstaller interface {
	Delete(ctx context.Context, kubeconfigFilePath, installName, namespace string) error
}

// ChartManager installs and uninstalls helm charts.
type ChartManager interface {
	ChartInstaller
	ChartUninstaller
}

// NewPackageControllerClientFullLifecycle creates a PackageControllerClient
// for the Full Cluster Lifecycle controller.
//
// It differs because the CLI use case has far more information available at
// instantiation, while the FCL use case has less information at
// instantiation, and the rest when cluster creation is triggered.
func NewPackageControllerClientFullLifecycle(logger logr.Logger, chartManager ChartManager, kubectl KubectlRunner, clientBuilder ClientBuilder) *PackageControllerClient {
	return &PackageControllerClient{
		chartManager:             chartManager,
		kubectl:                  kubectl,
		skipWaitForPackageBundle: true,
		eksaRegion:               eksaDefaultRegion,
		clientBuilder:            clientBuilder,
		registryAccessTester:     &DefaultRegistryAccessTester{},
	}
}

// EnableFullLifecycle wraps Enable to handle run-time arguments.
//
// This method fills in the gaps between the original CLI use case, where all
// information is known at PackageControllerClient initialization, and the
// Full Cluster Lifecycle use case, where there's limited information at
// initialization. Basically any parameter here isn't known at instantiation
// of the PackageControllerClient during full cluster lifecycle usage, hence
// why this method exists.
func (pc *PackageControllerClient) EnableFullLifecycle(ctx context.Context, log logr.Logger, clusterName, kubeConfig string, chart *releasev1.Image, registryMirror *registrymirror.RegistryMirror, options ...PackageControllerClientOpt) (err error) {
	log.V(6).Info("enabling curated packages full lifecycle")
	defer func(err *error) {
		if err != nil && *err != nil {
			log.Error(*err, "Enabling curated packages full lifecycle", "clusterName", clusterName)
		} else {
			log.Info("Successfully enabled curated packages full lifecycle")
		}
	}(&err)
	pc.mu.Lock()
	// This anonymous function ensures that the pc.mu is unlocked before
	// Enable is called, preventing deadlocks in the event that Enable tries
	// to acquire pc.mu.
	err = func() error {
		defer pc.mu.Unlock()
		pc.skipWaitForPackageBundle = true
		pc.clusterName = clusterName
		pc.kubeConfig = kubeConfig
		pc.chart = chart
		pc.registryMirror = registryMirror
		writer, err := filewriter.NewWriter(clusterName)
		if err != nil {
			return fmt.Errorf("creating file writer for helm values: %w", err)
		}
		options = append(options, WithValuesFileWriter(writer))
		for _, o := range options {
			o(pc)
		}
		return nil
	}()
	if err != nil {
		return err
	}

	return pc.Enable(ctx)
}

// NewPackageControllerClient instantiates a new instance of PackageControllerClient.
func NewPackageControllerClient(chartManager ChartManager, kubectl KubectlRunner, clusterName, kubeConfig string, chart *releasev1.Image, registryMirror *registrymirror.RegistryMirror, options ...PackageControllerClientOpt) *PackageControllerClient {
	pcc := &PackageControllerClient{
		kubeConfig:           kubeConfig,
		clusterName:          clusterName,
		chart:                chart,
		chartManager:         chartManager,
		kubectl:              kubectl,
		registryMirror:       registryMirror,
		eksaRegion:           eksaDefaultRegion,
		registryAccessTester: &DefaultRegistryAccessTester{},
	}

	for _, o := range options {
		o(pcc)
	}
	return pcc
}

// Enable curated packages in a cluster
//
// In case the cluster is management cluster, it performs the following actions:
//   - Installation of Package Controller through helm chart installation
//   - Creation of secret credentials
//   - Creation of a single run of a cron job refresher
//   - Activation of a curated packages bundle
//
// In case the cluster is a workload cluster, it performs the following actions:
//   - Creation of package bundle controller (PBC) custom resource in management cluster
func (pc *PackageControllerClient) Enable(ctx context.Context) error {
	ociURI := fmt.Sprintf("%s%s", "oci://", pc.registryMirror.ReplaceRegistry(pc.chart.Image()))
	clusterName := fmt.Sprintf("clusterName=%s", pc.clusterName)
	sourceRegistry, defaultRegistry, defaultImageRegistry := pc.GetCuratedPackagesRegistries(ctx)
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

	skipCRDs := false
	chartName := pc.chart.Name
	if pc.managementClusterName != pc.clusterName {
		values = append(values, "workloadPackageOnly=true")
		values = append(values, "managementClusterName="+pc.managementClusterName)
		chartName = chartName + "-" + pc.clusterName
		skipCRDs = true
	}

	if err := pc.chartManager.InstallChart(ctx, chartName, ociURI, pc.chart.Tag(), pc.kubeConfig, constants.EksaPackagesName, valueFilePath, skipCRDs, values); err != nil {
		return err
	}

	if !pc.skipWaitForPackageBundle {
		return pc.waitForActiveBundle(ctx)
	}

	return nil
}

// GetCuratedPackagesRegistries gets value for configurable registries from PBC.
func (pc *PackageControllerClient) GetCuratedPackagesRegistries(ctx context.Context) (sourceRegistry, defaultRegistry, defaultImageRegistry string) {
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
		sourceRegistry = publicStagingECR
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
		if pc.eksaRegion != eksaDefaultRegion {
			defaultImageRegistry = strings.ReplaceAll(defaultImageRegistry, eksaDefaultRegion, pc.eksaRegion)
		}

		regionalRegistry := GetRegionalRegistry(defaultRegistry, pc.eksaRegion)
		if err := pc.registryAccessTester.Test(ctx, pc.eksaAccessKeyID, pc.eksaSecretAccessKey, pc.eksaRegion, pc.eksaAwsConfig, regionalRegistry); err == nil {
			// use regional registry when the above credential is good
			logger.V(6).Info("Using regional registry")
			// In the dev case, we use a separate public ECR registry in the
			// beta packages account to source the packages controller and
			// credential provider package
			if regionalRegistry == devRegionalECR {
				sourceRegistry = devRegionalPublicECR
			}
			defaultRegistry = regionalRegistry
			defaultImageRegistry = regionalRegistry
		} else {
			logger.V(6).Info("Error pulling from regional registry", "Registry", regionalRegistry, "RegionalRegistryAccessIssue", err)
			logger.V(6).Info("Using fallback registry", "Registry", defaultRegistry)
		}
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
		"eksaAwsConfig":       base64.StdEncoding.EncodeToString([]byte(pc.eksaAwsConfig)),
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

	values, err := pc.GetPackageControllerConfiguration()
	return []byte(values + string(result)), err
}

// packageBundleControllerResource is the name of the package bundle controller
// resource in the API.
const packageBundleControllerResource string = "packageBundleController"

// waitForActiveBundle polls the package bundle controller for its active bundle.
//
// It returns nil on success. Success is defined as receiving a valid package
// bundle controller from the API with a non-empty active bundle.
//
// If no timeout is specified, a default of 3 minutes is used.
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
			if err != nil && !apierrors.IsNotFound(err) {
				done <- fmt.Errorf("getting package bundle controller: %w", err)
				return
			}

			if pbc != nil && pbc.Spec.ActiveBundle != "" {
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
	hasResource, err := pc.kubectl.HasResource(ctx, packageBundleControllerResource, pc.clusterName, pc.kubeConfig, constants.EksaPackagesName)
	return hasResource && err == nil
}

func formatYamlLine(space, key, value string) string {
	if value == "" {
		return ""
	}
	return space + key + ": " + value + "\n"
}

func formatImageResource(resource *anywherev1.ImageResource, name string) (result string) {
	if resource.CPU != "" || resource.Memory != "" {
		result = "    " + name + ":\n"
		result += formatYamlLine("      ", "cpu", resource.CPU)
		result += formatYamlLine("      ", "memory", resource.Memory)
	}
	return result
}

func formatCronJob(cronJob *anywherev1.PackageControllerCronJob) (result string) {
	if cronJob != nil {
		result += "cronjob:\n"
		result += formatYamlLine("  ", "digest", cronJob.Digest)
		result += formatYamlLine("  ", "repository", cronJob.Repository)
		result += formatYamlLine("  ", "suspend", strconv.FormatBool(cronJob.Disable))
		result += formatYamlLine("  ", "tag", cronJob.Tag)
	}
	return result
}

func formatResources(resources *anywherev1.PackageControllerResources) (result string) {
	if resources.Limits.CPU != "" || resources.Limits.Memory != "" ||
		resources.Requests.CPU != "" || resources.Requests.Memory != "" {
		result += "  resources:\n"
		result += formatImageResource(&resources.Limits, "limits")
		result += formatImageResource(&resources.Requests, "requests")
	}
	return result
}

// GetPackageControllerConfiguration returns the default kubernetes version for a Cluster.
func (pc *PackageControllerClient) GetPackageControllerConfiguration() (result string, err error) {
	clusterSpec := pc.clusterSpec
	if clusterSpec == nil || clusterSpec.Packages == nil {
		return "", nil
	}

	if clusterSpec.Packages.Controller != nil {
		result += "controller:\n"
		result += formatYamlLine("  ", "digest", clusterSpec.Packages.Controller.Digest)
		result += formatYamlLine("  ", "enableWebhooks", strconv.FormatBool(!clusterSpec.Packages.Controller.DisableWebhooks))
		result += formatYamlLine("  ", "repository", clusterSpec.Packages.Controller.Repository)
		result += formatYamlLine("  ", "tag", clusterSpec.Packages.Controller.Tag)
		result += formatResources(&clusterSpec.Packages.Controller.Resources)
		if len(clusterSpec.Packages.Controller.Env) > 0 {
			result += "  env:\n"
			for _, kvp := range clusterSpec.Packages.Controller.Env {
				results := strings.SplitN(kvp, "=", 2)
				if len(results) != 2 {
					err = fmt.Errorf("invalid environment in specification <%s>", kvp)
					continue
				}
				result += "  - name: " + results[0] + "\n"
				result += "    value: " + results[1] + "\n"
			}
		}
	}
	result += formatCronJob(clusterSpec.Packages.CronJob)

	return result, err
}

// Reconcile installs resources when a full cluster lifecycle cluster is created.
func (pc *PackageControllerClient) Reconcile(ctx context.Context, logger logr.Logger, client client.Client, cluster *anywherev1.Cluster) error {
	image, err := pc.getBundleFromCluster(ctx, client, cluster)
	if err != nil {
		return err
	}

	registry := registrymirror.FromCluster(cluster)

	// No Kubeconfig is passed. This is intentional. The helm executable will
	// get that configuration from its environment.
	if err := pc.EnableFullLifecycle(ctx, logger, cluster.Name, "", image, registry,
		WithManagementClusterName(cluster.ManagedBy())); err != nil {
		return fmt.Errorf("packages client error: %w", err)
	}

	return nil
}

// getBundleFromCluster based on the cluster's k8s version.
func (pc *PackageControllerClient) getBundleFromCluster(ctx context.Context, client client.Client, clusterObj *anywherev1.Cluster) (*releasev1.Image, error) {
	bundles, err := cluster.BundlesForCluster(ctx, clientutil.NewKubeClient(client), clusterObj)
	if err != nil {
		return nil, err
	}

	verBundle, err := cluster.GetVersionsBundle(clusterObj.Spec.KubernetesVersion, bundles)
	if err != nil {
		return nil, err
	}

	return &verBundle.PackageController.HelmChart, nil
}

// KubeDeleter abstracts client.Client so mocks can be substituted in tests.
type KubeDeleter interface {
	Delete(context.Context, client.Object, ...client.DeleteOption) error
}

// ReconcileDelete removes resources after a full cluster lifecycle cluster is
// deleted.
func (pc *PackageControllerClient) ReconcileDelete(ctx context.Context, logger logr.Logger, client KubeDeleter, cluster *anywherev1.Cluster) error {
	namespace := "eksa-packages-" + cluster.Name
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}
	if err := client.Delete(ctx, ns); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("deleting workload cluster curated packages namespace %q %w", namespace, err)
		}
		logger.V(6).Info("not found", "namespace", namespace)
	}

	name := "eks-anywhere-packages-" + pc.clusterName
	if err := pc.chartManager.Delete(ctx, pc.kubeConfig, name, constants.EksaPackagesName); err != nil {
		if !strings.Contains(err.Error(), "release: not found") {
			return err
		}
		logger.V(6).Info("not found", "release", name)
	}

	logger.Info("Removed curated packages installation", "clusterName")

	return nil
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
		}
	}
}

// WithEksaAwsConfig set the eksaAwsConfig field.
func WithEksaAwsConfig(eksaAwsConfig string) func(client *PackageControllerClient) {
	return func(config *PackageControllerClient) {
		if eksaAwsConfig != "" {
			config.eksaAwsConfig = eksaAwsConfig
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

// WithClusterSpec sets the cluster spec.
func WithClusterSpec(clusterSpec *cluster.Spec) func(client *PackageControllerClient) {
	return func(config *PackageControllerClient) {
		config.clusterSpec = &clusterSpec.Cluster.Spec
	}
}

// WithRegistryAccessTester sets the registryTester.
func WithRegistryAccessTester(registryTester RegistryAccessTester) func(client *PackageControllerClient) {
	return func(config *PackageControllerClient) {
		config.registryAccessTester = registryTester
	}
}
