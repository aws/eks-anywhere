package stack

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/helm"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	args              = "args"
	createNamespace   = "createNamespace"
	deploy            = "deploy"
	env               = "env"
	hostPortEnabled   = "hostPortEnabled"
	image             = "image"
	namespace         = "namespace"
	overridesFileName = "tinkerbell-chart-overrides.yaml"
	port              = "port"

	boots          = "boots"
	hegel          = "hegel"
	tinkController = "tinkController"
	tinkServer     = "tinkServer"
	rufio          = "rufio"
	grpcPort       = "42113"
	kubevip        = "kubevip"
	envoy          = "envoy"
)

type Docker interface {
	CheckContainerExistence(ctx context.Context, name string) (bool, error)
	ForceRemove(ctx context.Context, name string) error
	Run(ctx context.Context, image string, name string, cmd []string, flags ...string) error
}

type Helm interface {
	RegistryLogin(ctx context.Context, endpoint, username, password string) error
	InstallChartWithValuesFile(ctx context.Context, chart, ociURI, version, kubeconfigFilePath, valuesFilePath string) error
	UpgradeChartWithValuesFile(ctx context.Context, chart, ociURI, version, kubeconfigFilePath, valuesFilePath string, opts ...helm.Opt) error
}

// StackInstaller deploys a Tinkerbell stack.
//
//nolint:revive // Stutter and the interface shouldn't exist. Will clean up (chrisdoherty4)
type StackInstaller interface {
	CleanupLocalBoots(ctx context.Context, forceCleanup bool) error
	Install(ctx context.Context, bundle releasev1alpha1.TinkerbellBundle, tinkerbellIP, kubeconfig, hookOverride string, opts ...InstallOption) error
	UninstallLocal(ctx context.Context) error
	Upgrade(_ context.Context, _ releasev1alpha1.TinkerbellBundle, tinkerbellIP, kubeconfig, hookOverride string, opts ...InstallOption) error
	AddNoProxyIP(IP string)
	GetNamespace() string
}

type Installer struct {
	docker          Docker
	filewriter      filewriter.FileWriter
	helm            Helm
	podCidrRange    string
	registryMirror  *registrymirror.RegistryMirror
	proxyConfig     *v1alpha1.ProxyConfiguration
	namespace       string
	createNamespace bool
	bootsOnDocker   bool
	hostPort        bool
	loadBalancer    bool
	envoy           bool
}

type InstallOption func(s *Installer)

// WithNamespaceCreate is an InstallOption is lets you specify whether to create the namespace needed for Tinkerbell stack.
func WithNamespaceCreate(create bool) InstallOption {
	return func(s *Installer) {
		s.createNamespace = create
	}
}

// WithBootsOnDocker is an InstallOption to run Boots as a Docker container.
func WithBootsOnDocker() InstallOption {
	return func(s *Installer) {
		s.bootsOnDocker = true
	}
}

// WithBootsOnKubernetes is an InstallOption to run Boots as a Kubernetes deployment.
func WithBootsOnKubernetes() InstallOption {
	return func(s *Installer) {
		s.bootsOnDocker = false
	}
}

// WithHostPortEnabled is an InstallOption that allows you to enable/disable host port for Tinkerbell deployments.
func WithHostPortEnabled(enabled bool) InstallOption {
	return func(s *Installer) {
		s.hostPort = enabled
	}
}

func WithEnvoyEnabled(enabled bool) InstallOption {
	return func(s *Installer) {
		s.envoy = enabled
	}
}

// WithLoadBalancer is an InstallOption that allows you to setup a LoadBalancer to expose hegel and tink-server.
func WithLoadBalancerEnabled(enabled bool) InstallOption {
	return func(s *Installer) {
		s.loadBalancer = enabled
	}
}

// AddNoProxyIP is for workload cluster upgrade, we have to pass
// controlPlaneEndpoint IP of managemement cluster if proxy is configured.
func (s *Installer) AddNoProxyIP(IP string) {
	s.proxyConfig.NoProxy = append(s.proxyConfig.NoProxy, IP)
}

// NewInstaller returns a Tinkerbell StackInstaller which can be used to install or uninstall the Tinkerbell stack.
func NewInstaller(docker Docker, filewriter filewriter.FileWriter, helm Helm, namespace string, podCidrRange string, registryMirror *registrymirror.RegistryMirror, proxyConfig *v1alpha1.ProxyConfiguration) StackInstaller {
	return &Installer{
		docker:         docker,
		filewriter:     filewriter,
		helm:           helm,
		registryMirror: registryMirror,
		proxyConfig:    proxyConfig,
		namespace:      namespace,
		podCidrRange:   podCidrRange,
	}
}

// Install installs the Tinkerbell stack on a target cluster using a helm chart and providing the necessary values overrides.
func (s *Installer) Install(ctx context.Context, bundle releasev1alpha1.TinkerbellBundle, tinkerbellIP, kubeconfig, hookOverride string, opts ...InstallOption) error {
	logger.V(6).Info("Installing Tinkerbell helm chart")

	for _, option := range opts {
		option(s)
	}

	bootEnv := s.getBootsEnv(bundle.TinkerbellStack, tinkerbellIP)

	osiePath, err := getURIDir(bundle.TinkerbellStack.Hook.Initramfs.Amd.URI)
	if err != nil {
		return fmt.Errorf("getting directory path from hook uri: %v", err)
	}

	if hookOverride != "" {
		osiePath = hookOverride
	}

	valuesMap := map[string]interface{}{
		namespace:       s.namespace,
		createNamespace: s.createNamespace,
		tinkController: map[string]interface{}{
			image: bundle.TinkerbellStack.Tink.TinkController.URI,
		},
		tinkServer: map[string]interface{}{
			image: bundle.TinkerbellStack.Tink.TinkServer.URI,
			args:  []string{},
			port: map[string]bool{
				hostPortEnabled: s.hostPort,
			},
		},
		hegel: map[string]interface{}{
			image: bundle.TinkerbellStack.Hegel.URI,
			port: map[string]bool{
				hostPortEnabled: s.hostPort,
			},
			env: []map[string]string{
				{
					"name":  "HEGEL_TRUSTED_PROXIES",
					"value": s.podCidrRange,
				},
			},
		},
		boots: map[string]interface{}{
			deploy: !s.bootsOnDocker,
			image:  bundle.TinkerbellStack.Boots.URI,
			env:    bootEnv,
			args: []string{
				"-dhcp-addr=0.0.0.0:67",
				fmt.Sprintf("-osie-path-override=%s", osiePath),
			},
		},
		rufio: map[string]interface{}{
			image: bundle.TinkerbellStack.Rufio.URI,
		},
		kubevip: map[string]interface{}{
			image:  bundle.KubeVip.URI,
			deploy: s.loadBalancer,
		},
		envoy: map[string]interface{}{
			image:        bundle.Envoy.URI,
			deploy:       s.envoy,
			"externalIp": tinkerbellIP,
		},
	}

	values, err := yaml.Marshal(valuesMap)
	if err != nil {
		return fmt.Errorf("marshalling values override for Tinkerbell Installer helm chart: %s", err)
	}

	valuesPath, err := s.filewriter.Write(overridesFileName, values)
	if err != nil {
		return fmt.Errorf("writing values override for Tinkerbell Installer helm chart: %s", err)
	}

	if err := s.authenticateHelmRegistry(ctx); err != nil {
		return err
	}

	err = s.helm.InstallChartWithValuesFile(
		ctx,
		bundle.TinkerbellStack.TinkebellChart.Name,
		fmt.Sprintf("oci://%s", s.localRegistryURL(bundle.TinkerbellStack.TinkebellChart.Image())),
		bundle.TinkerbellStack.TinkebellChart.Tag(),
		kubeconfig,
		valuesPath,
	)
	if err != nil {
		return fmt.Errorf("installing Tinkerbell helm chart: %v", err)
	}

	return s.installBootsOnDocker(ctx, bundle.TinkerbellStack, tinkerbellIP, kubeconfig, hookOverride)
}

func (s *Installer) installBootsOnDocker(ctx context.Context, bundle releasev1alpha1.TinkerbellStackBundle, tinkServerIP, kubeconfig, hookOverride string) error {
	if !s.bootsOnDocker {
		return nil
	}

	kubeconfig, err := filepath.Abs(kubeconfig)
	if err != nil {
		return fmt.Errorf("getting absolute path for kubeconfig: %v", err)
	}

	flags := []string{
		"-v", fmt.Sprintf("%s:/kubeconfig", kubeconfig),
		"--network", "host",
		"-e", fmt.Sprintf("PUBLIC_IP=%s", tinkServerIP),
		"-e", fmt.Sprintf("PUBLIC_SYSLOG_IP=%s", tinkServerIP),
		"-e", fmt.Sprintf("BOOTS_KUBE_NAMESPACE=%v", s.namespace),
	}

	for _, e := range s.getBootsEnv(bundle, tinkServerIP) {
		flags = append(flags, "-e", fmt.Sprintf("%s=%s", e["name"], e["value"]))
	}

	osiePath, err := getURIDir(bundle.Hook.Initramfs.Amd.URI)
	if err != nil {
		return fmt.Errorf("getting directory path from hook uri: %v", err)
	}

	if hookOverride != "" {
		osiePath = hookOverride
	}

	cmd := []string{
		"-kubeconfig", "/kubeconfig",
		"-dhcp-addr", "0.0.0.0:67",
		"-osie-path-override", osiePath,
	}
	if err := s.docker.Run(ctx, s.localRegistryURL(bundle.Boots.URI), boots, cmd, flags...); err != nil {
		return fmt.Errorf("running boots with docker: %v", err)
	}

	return nil
}

func (s *Installer) getBootsEnv(bundle releasev1alpha1.TinkerbellStackBundle, tinkServerIP string) []map[string]string {
	env := []map[string]string{
		toEnvEntry("DATA_MODEL_VERSION", "kubernetes"),
		toEnvEntry("TINKERBELL_TLS", "false"),
		toEnvEntry("TINKERBELL_GRPC_AUTHORITY", fmt.Sprintf("%s:%s", tinkServerIP, grpcPort)),
	}

	extraKernelArgs := fmt.Sprintf("tink_worker_image=%s", s.localRegistryURL(bundle.Tink.TinkWorker.URI))

	if s.registryMirror != nil {
		localRegistry := s.registryMirror.BaseRegistry
		extraKernelArgs = fmt.Sprintf("%s insecure_registries=%s", extraKernelArgs, localRegistry)
		if s.registryMirror.Auth {
			username, password, _ := config.ReadCredentials()
			env = append(env,
				toEnvEntry("REGISTRY_USERNAME", username),
				toEnvEntry("REGISTRY_PASSWORD", password))
		}
	}

	if s.proxyConfig != nil {
		noProxy := strings.Join(s.proxyConfig.NoProxy, ",")
		extraKernelArgs = fmt.Sprintf("%s HTTP_PROXY=%s HTTPS_PROXY=%s NO_PROXY=%s", extraKernelArgs, s.proxyConfig.HttpProxy, s.proxyConfig.HttpsProxy, noProxy)
	}

	return append(env, toEnvEntry("BOOTS_EXTRA_KERNEL_ARGS", extraKernelArgs))
}

func toEnvEntry(k, v string) map[string]string {
	return map[string]string{
		"name":  k,
		"value": v,
	}
}

// UninstallLocal currently removes local docker container running Boots.
func (s *Installer) UninstallLocal(ctx context.Context) error {
	return s.uninstallBootsFromDocker(ctx)
}

func (s *Installer) uninstallBootsFromDocker(ctx context.Context) error {
	logger.V(4).Info("Removing local boots container")
	if err := s.docker.ForceRemove(ctx, boots); err != nil {
		return fmt.Errorf("removing local boots container: %v", err)
	}

	return nil
}

func getURIDir(uri string) (string, error) {
	index := strings.LastIndex(uri, "/")
	if index == -1 {
		return "", fmt.Errorf("uri is invalid: %s", uri)
	}
	return uri[:index], nil
}

// CleanupLocalBoots determines whether Boots is already running locally
// and either cleans it up or errors out depending on the `remove` flag.
func (s *Installer) CleanupLocalBoots(ctx context.Context, remove bool) error {
	exists, err := s.docker.CheckContainerExistence(ctx, boots)
	// return error if the docker call failed
	if err != nil {
		return fmt.Errorf("checking boots container existence: %v", err)
	}

	// return nil if boots container doesn't exist
	if !exists {
		return nil
	}

	// if remove is set, try to delete boots
	if remove {
		return s.uninstallBootsFromDocker(ctx)
	}

	// finally, return an "already exists" error if boots exists and forceCleanup is not set
	return errors.New("boots container already exists, delete the container manually")
}

func (s *Installer) localRegistryURL(originalURL string) string {
	return s.registryMirror.ReplaceRegistry(originalURL)
}

func (s *Installer) authenticateHelmRegistry(ctx context.Context) error {
	if s.registryMirror != nil && s.registryMirror.Auth {
		username, password, err := config.ReadCredentials()
		if err != nil {
			return err
		}
		endpoint := s.registryMirror.BaseRegistry
		if err := s.helm.RegistryLogin(ctx, endpoint, username, password); err != nil {
			return err
		}
	}
	return nil
}

// Upgrade the Tinkerbell stack using images specified in bundle.
func (s *Installer) Upgrade(ctx context.Context, bundle releasev1alpha1.TinkerbellBundle, tinkerbellIP, kubeconfig string, hookOverride string, opts ...InstallOption) error {
	logger.V(6).Info("Upgrading Tinkerbell helm chart")

	for _, option := range opts {
		option(s)
	}

	bootEnv := s.getBootsEnv(bundle.TinkerbellStack, tinkerbellIP)

	osiePath, err := getURIDir(bundle.TinkerbellStack.Hook.Initramfs.Amd.URI)
	if err != nil {
		return fmt.Errorf("getting directory path from hook uri: %v", err)
	}
	if hookOverride != "" {
		osiePath = hookOverride
	}
	valuesMap := map[string]interface{}{
		namespace:       s.namespace,
		createNamespace: false,
		tinkController: map[string]interface{}{
			image: bundle.TinkerbellStack.Tink.TinkController.URI,
		},
		tinkServer: map[string]interface{}{
			image: bundle.TinkerbellStack.Tink.TinkServer.URI,
			args:  []string{},
		},
		hegel: map[string]interface{}{
			image: bundle.TinkerbellStack.Hegel.URI,
			env: []map[string]string{
				{
					"name":  "HEGEL_TRUSTED_PROXIES",
					"value": s.podCidrRange,
				},
			},
		},
		boots: map[string]interface{}{
			image: bundle.TinkerbellStack.Boots.URI,
			env:   bootEnv,
			args: []string{
				"-dhcp-addr=0.0.0.0:67",
				fmt.Sprintf("-osie-path-override=%s", osiePath),
			},
		},
		rufio: map[string]interface{}{
			image: bundle.TinkerbellStack.Rufio.URI,
		},
		kubevip: map[string]interface{}{
			image:  bundle.KubeVip.URI,
			deploy: s.loadBalancer,
		},
		envoy: map[string]interface{}{
			image: bundle.Envoy.URI,
		},
	}

	values, err := yaml.Marshal(valuesMap)
	if err != nil {
		return fmt.Errorf("marshalling values override for Tinkerbell Installer helm chart: %s", err)
	}

	valuesPath, err := s.filewriter.Write(overridesFileName, values)
	if err != nil {
		return fmt.Errorf("writing values override for Tinkerbell Installer helm chart: %s", err)
	}

	if err := s.authenticateHelmRegistry(ctx); err != nil {
		return err
	}

	envMap := map[string]string{}
	if s.proxyConfig != nil {
		envMap["NO_PROXY"] = strings.Join(s.proxyConfig.NoProxy, ",")
	}
	return s.helm.UpgradeChartWithValuesFile(
		ctx,
		bundle.TinkerbellStack.TinkebellChart.Name,
		fmt.Sprintf("oci://%s", s.localRegistryURL(bundle.TinkerbellStack.TinkebellChart.Image())),
		bundle.TinkerbellStack.TinkebellChart.Tag(),
		kubeconfig,
		valuesPath,
		helm.WithProxyConfig(envMap),
	)
}

// GetNamespace retrieves the namespace the installer is using for stack deployment.
func (s *Installer) GetNamespace() string {
	return s.namespace
}
