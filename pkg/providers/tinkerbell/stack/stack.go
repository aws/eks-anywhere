package stack

import (
	"context"
	"errors"
	"fmt"
	"net/url"
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
	deploy            = "deploy"
	additionalEnv     = "additionalEnv"
	hostPortEnabled   = "hostPortEnabled"
	image             = "image"
	namespace         = "namespace"
	overridesFileName = "tinkerbell-chart-overrides.yaml"
	port              = "port"
	addr              = "addr"
	enabled           = "enabled"
	kubevipInterface  = "interface"

	boots                        = "boots"
	smee                         = "smee"
	hegel                        = "hegel"
	tink                         = "tink"
	controller                   = "controller"
	server                       = "server"
	rufio                        = "rufio"
	rufioMaxConcurrentReconciles = 10
	grpcPort                     = "42113"
	kubevip                      = "kubevip"
	stack                        = "stack"
	hook                         = "hook"
	service                      = "service"
	relay                        = "relay"
	smeeHTTPPort                 = "7171"

	// localTinkWorkerImage is the path to the tink-worker image in the hook-OS.
	localTinkWorkerImage = "127.0.0.1/embedded/tink-worker"
)

type Docker interface {
	CheckContainerExistence(ctx context.Context, name string) (bool, error)
	ForceRemove(ctx context.Context, name string) error
	Run(ctx context.Context, image string, name string, cmd []string, flags ...string) error
}

type Helm interface {
	RegistryLogin(ctx context.Context, endpoint, username, password string) error
	UpgradeInstallChartWithValuesFile(ctx context.Context, chart, ociURI, version, kubeconfigFilePath, namespace, valuesFilePath string, opts ...helm.Opt) error
	Uninstall(ctx context.Context, chart, kubeconfigFilePath, namespace string, opts ...helm.Opt) error
	ListCharts(ctx context.Context, kubeconfigFilePath, filter string) ([]string, error)
}

// StackInstaller deploys a Tinkerbell stack.
//
//nolint:revive // Stutter and the interface shouldn't exist. Will clean up (chrisdoherty4)
type StackInstaller interface {
	CleanupLocalBoots(ctx context.Context, forceCleanup bool) error
	Install(ctx context.Context, bundle releasev1alpha1.TinkerbellBundle, tinkerbellIP, kubeconfig, hookOverride string, opts ...InstallOption) error
	UninstallLocal(ctx context.Context) error
	Uninstall(ctx context.Context, bundle releasev1alpha1.TinkerbellBundle, kubeconfig string) error
	Upgrade(_ context.Context, _ releasev1alpha1.TinkerbellBundle, tinkerbellIP, kubeconfig, hookOverride string, opts ...InstallOption) error
	UpgradeInstallCRDs(ctx context.Context, bundle releasev1alpha1.TinkerbellBundle, kubeconfig string, opts ...InstallOption) error
	UpgradeLegacy(ctx context.Context, bundle releasev1alpha1.TinkerbellBundle, kubeconfig string, opts ...InstallOption) error
	AddNoProxyIP(IP string)
	GetNamespace() string
	HasLegacyChart(ctx context.Context, bundle releasev1alpha1.TinkerbellBundle, kubeconfig string) (bool, error)
}

type Installer struct {
	docker                Docker
	filewriter            filewriter.FileWriter
	helm                  Helm
	podCidrRange          string
	registryMirror        *registrymirror.RegistryMirror
	proxyConfig           *v1alpha1.ProxyConfiguration
	namespace             string
	loadBalancerInterface string
	hookIsoURL            string
	bootsOnDocker         bool
	hostNetwork           bool
	loadBalancer          bool
	stackService          bool
	dhcpRelay             bool
}

type InstallOption func(s *Installer)

// WithLoadBalancerInterface is an InstallOption that allows you to configure load balancer interface for the tinkerbell stack.
func WithLoadBalancerInterface(loadBalancerInterface string) InstallOption {
	return func(s *Installer) {
		s.loadBalancerInterface = loadBalancerInterface
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

// WithHostNetworkEnabled is an InstallOption that allows you to enable/disable host network for Tinkerbell deployments.
func WithHostNetworkEnabled(enabled bool) InstallOption {
	return func(s *Installer) {
		s.hostNetwork = enabled
	}
}

// WithLoadBalancerEnabled is an InstallOption that allows you to setup a LoadBalancer to expose hegel and tink-server.
func WithLoadBalancerEnabled(enabled bool) InstallOption {
	return func(s *Installer) {
		s.loadBalancer = enabled
	}
}

// WithStackServiceEnabled is an InstallOption that allows you to enable the nginx service as a reverse proxy.
func WithStackServiceEnabled(enabled bool) InstallOption {
	return func(s *Installer) {
		s.stackService = enabled
	}
}

// WithDHCPRelayEnabled is an InstallOption that allows you to enable DHCP Relay.
func WithDHCPRelayEnabled(enabled bool) InstallOption {
	return func(s *Installer) {
		s.dhcpRelay = enabled
	}
}

// WithHookIsoOverride is an InstallOption allows you to set a URL of the HookOS ISO image.
func WithHookIsoOverride(url string) InstallOption {
	return func(s *Installer) {
		s.hookIsoURL = url
	}
}

// AddNoProxyIP is for workload cluster upgrade, we have to pass
// controlPlaneEndpoint IP of managemement cluster if proxy is configured.
func (s *Installer) AddNoProxyIP(IP string) {
	s.proxyConfig.NoProxy = append(s.proxyConfig.NoProxy, IP)
}

// NewInstaller returns a Tinkerbell StackInstaller which can be used to install or uninstall the Tinkerbell stack.
func NewInstaller(docker Docker, filewriter filewriter.FileWriter, helm Helm, hookIsoURL, namespace, podCidrRange string, registryMirror *registrymirror.RegistryMirror, proxyConfig *v1alpha1.ProxyConfiguration) StackInstaller {
	return &Installer{
		docker:         docker,
		filewriter:     filewriter,
		helm:           helm,
		registryMirror: registryMirror,
		proxyConfig:    proxyConfig,
		hookIsoURL:     hookIsoURL,
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

	bootEnv := s.getSmeeKernelArgs(bundle.TinkerbellStack)

	osieURI, err := getURIDir(bundle.TinkerbellStack.Hook.Initramfs.Amd.URI)
	if err != nil {
		return fmt.Errorf("getting directory path from hook uri: %v", err)
	}

	if hookOverride != "" {
		osieURI = hookOverride
	}

	osiePath, err := url.ParseRequestURI(osieURI)
	if err != nil {
		return fmt.Errorf("parsing hookOverride: %v", err)
	}

	if s.hookIsoURL == "" {
		s.hookIsoURL = bundle.TinkerbellStack.Hook.ISO.Amd.URI
	}

	valuesMap := s.createValuesOverride(bundle, bootEnv, tinkerbellIP, s.loadBalancerInterface, s.hookIsoURL, osiePath)

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

	additionalArgs := []string{
		"--skip-crds",
	}

	err = s.helm.UpgradeInstallChartWithValuesFile(
		ctx,
		bundle.TinkerbellStack.Stack.Name,
		fmt.Sprintf("oci://%s", s.localRegistryURL(bundle.TinkerbellStack.Stack.Image())),
		bundle.TinkerbellStack.Stack.Tag(),
		kubeconfig,
		s.namespace,
		valuesPath,
		helm.WithExtraFlags(additionalArgs),
	)
	if err != nil {
		return fmt.Errorf("installing Tinkerbell helm chart: %v", err)
	}

	return s.installBootsOnDocker(ctx, bundle.TinkerbellStack, tinkerbellIP, kubeconfig, hookOverride, s.hookIsoURL)
}

func (s *Installer) installBootsOnDocker(ctx context.Context, bundle releasev1alpha1.TinkerbellStackBundle, tinkServerIP, kubeconfig, hookOverride, isoOverride string) error {
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
		"-e", fmt.Sprintf("SMEE_DHCP_SYSLOG_IP=%s", tinkServerIP),
		"-e", fmt.Sprintf("SMEE_DHCP_IP_FOR_PACKET=%s", tinkServerIP),
		"-e", fmt.Sprintf("SMEE_DHCP_TFTP_IP=%s", tinkServerIP),
		"-e", fmt.Sprintf("SMEE_DHCP_HTTP_IPXE_BINARY_HOST=%s", tinkServerIP),
		"-e", fmt.Sprintf("SMEE_DHCP_HTTP_IPXE_BINARY_PORT=%s", smeeHTTPPort),
		"-e", fmt.Sprintf("SMEE_DHCP_HTTP_IPXE_SCRIPT_HOST=%s", tinkServerIP),
		"-e", fmt.Sprintf("SMEE_DHCP_HTTP_IPXE_SCRIPT_PORT=%s", smeeHTTPPort),
		"-e", fmt.Sprintf("SMEE_HTTP_PORT=%s", smeeHTTPPort),
		"-e", fmt.Sprintf("SMEE_BACKEND_KUBE_NAMESPACE=%v", s.namespace),
		"-e", fmt.Sprintf("SMEE_ISO_ENABLED=%v", true),
		"-e", fmt.Sprintf("SMEE_ISO_STATIC_IPAM_ENABLED=%v", true),
		"-e", fmt.Sprintf("SMEE_ISO_URL=%v", isoOverride),
	}

	extraKernelArgList := s.getSmeeKernelArgs(bundle)
	extraKernelArgs := strings.Join(extraKernelArgList, " ")
	flags = append(flags, "-e", fmt.Sprintf("%s=%s", "SMEE_EXTRA_KERNEL_ARGS", extraKernelArgs))

	osiePath, err := getURIDir(bundle.Hook.Initramfs.Amd.URI)
	if err != nil {
		return fmt.Errorf("getting directory path from hook uri: %v", err)
	}

	if hookOverride != "" {
		osiePath = hookOverride
	}

	cmd := []string{
		"-backend-kube-config", "/kubeconfig",
		"-dhcp-addr", "0.0.0.0:67",
		"-osie-url", osiePath,
		"-tink-server", fmt.Sprintf("%s:%s", tinkServerIP, grpcPort),
		"-syslog-addr", tinkServerIP,
		"-tftp-addr", tinkServerIP,
		"-http-addr", tinkServerIP,
	}
	if err := s.docker.Run(ctx, s.localRegistryURL(bundle.Boots.URI), boots, cmd, flags...); err != nil {
		return fmt.Errorf("running boots with docker: %v", err)
	}

	return nil
}

func (s *Installer) getSmeeKernelArgs(bundle releasev1alpha1.TinkerbellStackBundle) []string {
	extraKernelArgs := []string{}
	if s.bootsOnDocker {
		extraKernelArgs = append(extraKernelArgs, fmt.Sprintf("tink_worker_image=%s", localTinkWorkerImage))
	}

	if s.registryMirror != nil {
		localRegistry := s.registryMirror.BaseRegistry
		extraKernelArgs = append(extraKernelArgs, fmt.Sprintf("insecure_registries=%s", localRegistry))
		if s.registryMirror.Auth {
			username, password, _ := config.ReadCredentials()
			username = fmt.Sprintf("registry_username=%s", username)
			password = fmt.Sprintf("registry_password=%s", password)
			extraKernelArgs = append(extraKernelArgs, username, password)
		}
	}

	if s.proxyConfig != nil {
		noProxy := strings.Join(s.proxyConfig.NoProxy, ",")
		httpProxy := fmt.Sprintf("HTTP_PROXY=%s", s.proxyConfig.HttpProxy)
		httpsProxy := fmt.Sprintf("HTTPS_PROXY=%s", s.proxyConfig.HttpsProxy)
		noProxy = fmt.Sprintf("NO_PROXY=%s", noProxy)
		extraKernelArgs = append(extraKernelArgs, httpProxy, httpsProxy, noProxy)
	}

	return extraKernelArgs
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
func (s *Installer) Upgrade(ctx context.Context, bundle releasev1alpha1.TinkerbellBundle, tinkerbellIP, kubeconfig, hookOverride string, opts ...InstallOption) error {
	logger.V(6).Info("Upgrading Tinkerbell helm chart")

	for _, option := range opts {
		option(s)
	}

	bootEnv := s.getSmeeKernelArgs(bundle.TinkerbellStack)

	osieURI, err := getURIDir(bundle.TinkerbellStack.Hook.Initramfs.Amd.URI)
	if err != nil {
		return fmt.Errorf("getting directory path from hook uri: %v", err)
	}

	if hookOverride != "" {
		osieURI = hookOverride
	}

	osiePath, err := url.ParseRequestURI(osieURI)
	if err != nil {
		return fmt.Errorf("parsing hookOverride: %v", err)
	}

	if s.hookIsoURL == "" {
		s.hookIsoURL = bundle.TinkerbellStack.Hook.ISO.Amd.URI
	}

	valuesMap := s.createValuesOverride(bundle, bootEnv, tinkerbellIP, s.loadBalancerInterface, s.hookIsoURL, osiePath)

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

	additionalArgs := []string{
		"--skip-crds",
	}
	return s.helm.UpgradeInstallChartWithValuesFile(
		ctx,
		bundle.TinkerbellStack.Stack.Name,
		fmt.Sprintf("oci://%s", s.localRegistryURL(bundle.TinkerbellStack.Stack.Image())),
		bundle.TinkerbellStack.Stack.Tag(),
		kubeconfig,
		s.namespace,
		valuesPath,
		helm.WithProxyConfig(envMap),
		helm.WithExtraFlags(additionalArgs),
	)
}

// UpgradeInstallCRDs the Tinkerbell CRDs using images specified in bundle.
func (s *Installer) UpgradeInstallCRDs(ctx context.Context, bundle releasev1alpha1.TinkerbellBundle, kubeconfig string, opts ...InstallOption) error {
	logger.V(6).Info("Upgrading Tinkerbell CRDs helm chart")

	for _, option := range opts {
		option(s)
	}

	if err := s.authenticateHelmRegistry(ctx); err != nil {
		return err
	}

	envMap := map[string]string{}
	if s.proxyConfig != nil {
		envMap["NO_PROXY"] = strings.Join(s.proxyConfig.NoProxy, ",")
	}

	return s.helm.UpgradeInstallChartWithValuesFile(
		ctx,
		bundle.TinkerbellStack.TinkerbellCrds.Name,
		fmt.Sprintf("oci://%s", s.localRegistryURL(bundle.TinkerbellStack.TinkerbellCrds.Image())),
		bundle.TinkerbellStack.TinkerbellCrds.Tag(),
		kubeconfig,
		s.namespace,
		"",
		helm.WithProxyConfig(envMap),
		helm.WithExtraFlags([]string{}),
	)
}

// Uninstall uninstalls a tinkerbell chart of a certain name.
func (s *Installer) Uninstall(ctx context.Context, bundle releasev1alpha1.TinkerbellBundle, kubeconfig string) error {
	logger.V(6).Info("Uninstalling old Tinkerbell helm chart")

	additionalArgs := []string{
		"--ignore-not-found",
	}

	return s.helm.Uninstall(
		ctx,
		bundle.TinkerbellStack.TinkebellChart.Name,
		kubeconfig,
		"",
		helm.WithExtraFlags(additionalArgs),
	)
}

// GetNamespace retrieves the namespace the installer is using for stack deployment.
func (s *Installer) GetNamespace() string {
	return s.namespace
}

// UpgradeLegacy upgrades the legacy Tinkerbell stack using images specified in bundle.
func (s *Installer) UpgradeLegacy(ctx context.Context, bundle releasev1alpha1.TinkerbellBundle, kubeconfig string, opts ...InstallOption) error {
	logger.V(6).Info("Upgrading legacy Tinkerbell helm chart")

	for _, option := range opts {
		option(s)
	}

	if err := s.authenticateHelmRegistry(ctx); err != nil {
		return err
	}

	envMap := map[string]string{}
	if s.proxyConfig != nil {
		envMap["NO_PROXY"] = strings.Join(s.proxyConfig.NoProxy, ",")
	}

	return s.helm.UpgradeInstallChartWithValuesFile(
		ctx,
		bundle.TinkerbellStack.TinkebellChart.Name,
		fmt.Sprintf("oci://%s", s.localRegistryURL(bundle.TinkerbellStack.TinkebellChart.Image())),
		bundle.TinkerbellStack.TinkebellChart.Tag(),
		kubeconfig,
		"",
		"",
		helm.WithProxyConfig(envMap),
	)
}

// HasLegacyChart returns whether or not the legacy tinkerbell-chart exists on the cluster.
func (s *Installer) HasLegacyChart(ctx context.Context, bundle releasev1alpha1.TinkerbellBundle, kubeconfig string) (bool, error) {
	logger.V(6).Info("Checking if legacy chart exists")

	charts, err := s.helm.ListCharts(ctx, kubeconfig, bundle.TinkerbellStack.TinkebellChart.Name)
	if err != nil {
		return false, err
	}

	if len(charts) > 0 {
		return true, nil
	}

	return false, nil
}

// createValuesOverride generates the values override file to send to helm.
func (s *Installer) createValuesOverride(bundle releasev1alpha1.TinkerbellBundle, bootEnv []string, tinkerbellIP, loadBalancerInterface, isoURL string, osiePath *url.URL) map[string]interface{} {
	valuesMap := map[string]interface{}{
		tink: map[string]interface{}{
			controller: map[string]interface{}{
				image: bundle.TinkerbellStack.Tink.TinkController.URI,
				"singleNodeClusterConfig": map[string]interface{}{
					"controlPlaneTolerationsEnabled": true,
					"nodeAffinityWeight":             1,
				},
			},
			server: map[string]interface{}{
				image: bundle.TinkerbellStack.Tink.TinkServer.URI,
				"singleNodeClusterConfig": map[string]interface{}{
					"controlPlaneTolerationsEnabled": true,
					"nodeAffinityWeight":             1,
				},
			},
		},
		hegel: map[string]interface{}{
			image: bundle.TinkerbellStack.Hegel.URI,
			"trustedProxies": []string{
				s.podCidrRange,
			},
			"singleNodeClusterConfig": map[string]interface{}{
				"controlPlaneTolerationsEnabled": true,
				"nodeAffinityWeight":             1,
			},
		},
		smee: map[string]interface{}{
			deploy:     !s.bootsOnDocker,
			image:      bundle.TinkerbellStack.Boots.URI,
			"publicIP": tinkerbellIP,
			"trustedProxies": []string{
				s.podCidrRange,
			},
			"http": map[string]interface{}{
				"tinkServer": map[string]interface{}{
					"ip":          tinkerbellIP,
					port:          grpcPort,
					"insecureTLS": true,
				},
				"osieUrl": map[string]interface{}{
					"scheme": osiePath.Scheme,
					"host":   osiePath.Hostname(),
					"port":   osiePath.Port(),
					"path":   osiePath.Path,
				},
				"additionalKernelArgs": bootEnv,
			},
			"tinkWorkerImage": localTinkWorkerImage,
			"iso": map[string]interface{}{
				// it's safe to populate the URL and default to true as rufio jobs for mounting and booting
				// from iso happens only when bootmode is set to iso on tinkerbellmachinetemplate
				enabled:             true,
				"staticIPAMEnabled": true,
				"url":               isoURL,
			},
			"singleNodeClusterConfig": map[string]interface{}{
				"controlPlaneTolerationsEnabled": true,
				"nodeAffinityWeight":             1,
			},
		},
		rufio: map[string]interface{}{
			image: bundle.TinkerbellStack.Rufio.URI,
			"additionalArgs": []string{
				"-metrics-bind-address=127.0.0.1:8080",
				fmt.Sprintf("-max-concurrent-reconciles=%v", rufioMaxConcurrentReconciles),
			},
			"singleNodeClusterConfig": map[string]interface{}{
				"controlPlaneTolerationsEnabled": true,
				"nodeAffinityWeight":             1,
			},
		},
		stack: map[string]interface{}{
			image: bundle.TinkerbellStack.Tink.Nginx.URI,
			"singleNodeClusterConfig": map[string]interface{}{
				"controlPlaneTolerationsEnabled": true,
				"nodeAffinityWeight":             1,
			},
			kubevip: map[string]interface{}{
				image:   bundle.KubeVip.URI,
				enabled: s.loadBalancer,
				additionalEnv: []map[string]string{
					{
						"name":  "prometheus_server",
						"value": ":2213",
					},
					// The Tinkerbell stack needs a load balancer to work properly.
					// We bundle Kube-vip in, as the load balancer, when we deploy the stack.
					// We don't want this load balancer to be used by any other workloads.
					// It allows us greater confidence in successful lifecycle events for the Tinkerbell stack, amongst other things.
					// Also, the user should be free from Tinkerbell stack constraints
					// and free to deploy a load balancer of their choosing and not be coupled to ours.
					// setting lb_class_only=true means that k8s services must explicitly populate
					// the kube-vip loadBalancerClass with the kube-vip value for kube-vip to serve an IP.
					{
						"name":  "lb_class_only",
						"value": "true",
					},
				},
			},
			hook: map[string]interface{}{
				enabled: false,
			},
			service: map[string]interface{}{
				enabled: s.stackService,
			},
			relay: map[string]interface{}{
				enabled:     s.dhcpRelay,
				image:       bundle.TinkerbellStack.Tink.TinkRelay.URI,
				"initImage": bundle.TinkerbellStack.Tink.TinkRelayInit.URI,
			},
			"loadBalancerIP": tinkerbellIP,
			"hostNetwork":    s.hostNetwork,
		},
	}

	if loadBalancerInterface != "" {
		valuesMap[stack].(map[string]interface{})[kubevip].(map[string]interface{})[kubevipInterface] = loadBalancerInterface
	}

	return valuesMap
}
