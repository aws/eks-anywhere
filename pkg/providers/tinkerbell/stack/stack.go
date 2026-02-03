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
	overridesFileName = "tinkerbell-chart-overrides.yaml"

	boots        = "boots"
	smeeHTTPPort = "7171"
	grpcPort     = "42113"

	// localTinkWorkerImage is the path to the tink-worker image embedded in HookOS.
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

	osiePath, err := getURIDir(bundle.Hook.Initramfs.Amd.URI)
	if err != nil {
		return fmt.Errorf("getting directory path from hook uri: %v", err)
	}

	if hookOverride != "" {
		osiePath = hookOverride
	}

	extraKernelArgList := s.getSmeeKernelArgs(bundle)
	extraKernelArgList = append(extraKernelArgList, fmt.Sprintf("tink_worker_image=%s", localTinkWorkerImage))
	extraKernelArgs := strings.Join(extraKernelArgList, " ")

	// Mono-repo tinkerbell binary uses TINKERBELL_ prefix for env vars
	flags := []string{
		"-v", fmt.Sprintf("%s:/kubeconfig", kubeconfig),
		"--network", "host",
		"-e", "TINKERBELL_BACKEND=kube",
		"-e", "TINKERBELL_BACKEND_KUBE_CONFIG=/kubeconfig",
		"-e", fmt.Sprintf("TINKERBELL_BACKEND_KUBE_NAMESPACE=%s", s.namespace),
		"-e", fmt.Sprintf("TINKERBELL_PUBLIC_IPV4=%s", tinkServerIP),
		"-e", fmt.Sprintf("TINKERBELL_TRUSTED_PROXIES=%s", s.podCidrRange),
		"-e", "TINKERBELL_ENABLE_SMEE=true",
		"-e", "TINKERBELL_ENABLE_TOOTLES=false",
		"-e", "TINKERBELL_ENABLE_TINK_SERVER=false",
		"-e", "TINKERBELL_ENABLE_TINK_CONTROLLER=false",
		"-e", "TINKERBELL_ENABLE_RUFIO_CONTROLLER=false",
		"-e", "TINKERBELL_ENABLE_SECONDSTAR=false",
		"-e", "TINKERBELL_ENABLE_CRD_MIGRATIONS=false",
		"-e", "TINKERBELL_DHCP_ENABLED=true",
		"-e", "TINKERBELL_DHCP_MODE=reservation",
		"-e", fmt.Sprintf("TINKERBELL_DHCP_IP_FOR_PACKET=%s", tinkServerIP),
		"-e", fmt.Sprintf("TINKERBELL_DHCP_SYSLOG_IP=%s", tinkServerIP),
		"-e", fmt.Sprintf("TINKERBELL_DHCP_TFTP_IP=%s", tinkServerIP),
		"-e", fmt.Sprintf("TINKERBELL_DHCP_IPXE_HTTP_BINARY_HOST=%s", tinkServerIP),
		"-e", fmt.Sprintf("TINKERBELL_DHCP_IPXE_HTTP_BINARY_PORT=%s", smeeHTTPPort),
		"-e", fmt.Sprintf("TINKERBELL_DHCP_IPXE_HTTP_SCRIPT_HOST=%s", tinkServerIP),
		"-e", fmt.Sprintf("TINKERBELL_DHCP_IPXE_HTTP_SCRIPT_PORT=%s", smeeHTTPPort),
		"-e", fmt.Sprintf("TINKERBELL_IPXE_HTTP_SCRIPT_BIND_PORT=%s", smeeHTTPPort),
		"-e", fmt.Sprintf("TINKERBELL_IPXE_HTTP_SCRIPT_OSIE_URL=%s", osiePath),
		"-e", fmt.Sprintf("TINKERBELL_IPXE_HTTP_SCRIPT_EXTRA_KERNEL_ARGS=%s", extraKernelArgs),
		"-e", fmt.Sprintf("TINKERBELL_IPXE_SCRIPT_TINK_SERVER_ADDR_PORT=%s:%s", tinkServerIP, grpcPort),
		"-e", "TINKERBELL_IPXE_SCRIPT_TINK_SERVER_INSECURE_TLS=true",
		"-e", "TINKERBELL_ISO_ENABLED=true",
		"-e", "TINKERBELL_ISO_STATIC_IPAM_ENABLED=true",
		"-e", fmt.Sprintf("TINKERBELL_ISO_UPSTREAM_URL=%s", isoOverride),
		"-e", "TINKERBELL_TFTP_SERVER_ENABLED=true",
		"-e", "TINKERBELL_SYSLOG_ENABLED=true",
	}

	// Mono-repo binary uses env vars only, no command line args
	cmd := []string{}
	if err := s.docker.Run(ctx, s.localRegistryURL(bundle.Boots.URI), boots, cmd, flags...); err != nil {
		return fmt.Errorf("running boots with docker: %v", err)
	}

	return nil
}

func (s *Installer) getSmeeKernelArgs(_ releasev1alpha1.TinkerbellStackBundle) []string {
	extraKernelArgs := []string{}

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

// createValuesOverride generates the values override file for the mono-repo tinkerbell helm chart.
// The mono-repo chart uses a single deployment with all services (smee, tootles, tink-server,
// tink-controller, rufio) running in one pod, controlled via enable flags.
func (s *Installer) createValuesOverride(bundle releasev1alpha1.TinkerbellBundle, bootEnv []string, tinkerbellIP, loadBalancerInterface, isoURL string, osiePath *url.URL) map[string]any {
	// Build OSIE URL from parsed path
	osieURL := fmt.Sprintf("%s://%s", osiePath.Scheme, osiePath.Host)
	if osiePath.Path != "" {
		osieURL = osieURL + osiePath.Path
	}

	tinkerbellImageURI := bundle.TinkerbellStack.Boots.URI
	kubevipImageURI := bundle.KubeVip.URI
	relayInitImageURI := bundle.TinkerbellStack.Tink.TinkRelayInit.URI

	tinkerbellImage, tinkerbellTag := parseImageURI(tinkerbellImageURI)

	valuesMap := map[string]any{
		"name":     "tinkerbell",
		"publicIP": tinkerbellIP,
		"trustedProxies": []string{
			s.podCidrRange,
		},
		"deployment": map[string]any{
			"image":         tinkerbellImage,
			"imageTag":      tinkerbellTag,
			"agentImage":    localTinkWorkerImage,
			"agentImageTag": "",
			"hostNetwork":   s.hostNetwork,
			"tolerations": []map[string]any{
				{
					"key":      "node-role.kubernetes.io/control-plane",
					"operator": "Exists",
					"effect":   "NoSchedule",
				},
			},
			"affinity": map[string]any{
				"nodeAffinity": map[string]any{
					"preferredDuringSchedulingIgnoredDuringExecution": []map[string]any{
						{
							"weight": 1,
							"preference": map[string]any{
								"matchExpressions": []map[string]any{
									{
										"key":      "node-role.kubernetes.io/control-plane",
										"operator": "DoesNotExist",
									},
								},
							},
						},
					},
				},
			},
			"init": map[string]any{
				"enabled":       s.dhcpRelay,
				"image":         relayInitImageURI,
				"interfaceMode": "macvlan",
			},
			"envs": map[string]any{
				"globals": map[string]any{
					"backend":               "kube",
					"backendKubeNamespace":  s.namespace,
					"enableSmee":            !s.bootsOnDocker,
					"enableTootles":         true,
					"enableTinkServer":      true,
					"enableTinkController":  true,
					"enableRufioController": true,
					"enableSecondstar":      false,
					"enableCRDMigrations":   false,
				},
				"smee": map[string]any{
					"dhcpEnabled":                     true,
					"dhcpMode":                        "reservation",
					"dhcpIPForPacket":                 tinkerbellIP,
					"dhcpSyslogIP":                    tinkerbellIP,
					"dhcpTftpIP":                      tinkerbellIP,
					"dhcpIpxeHttpBinaryHost":          tinkerbellIP,
					"dhcpIpxeHttpBinaryPort":          7171,
					"dhcpIpxeHttpScriptHost":          tinkerbellIP,
					"dhcpIpxeHttpScriptPort":          7171,
					"ipxeHttpScriptBindPort":          7171,
					"ipxeHttpScriptOsieURL":           osieURL,
					"ipxeHttpScriptExtraKernelArgs":   bootEnv,
					"ipxeScriptTinkServerAddrPort":    fmt.Sprintf("%s:%s", tinkerbellIP, grpcPort),
					"ipxeScriptTinkServerInsecureTLS": true,
					"isoEnabled":                      true,
					"isoStaticIPAMEnabled":            true,
					"isoUpstreamURL":                  isoURL,
					"tftpServerEnabled":               true,
					"syslogEnabled":                   true,
				},
				"tootles": map[string]any{
					"bindPort": 7172,
				},
				"tinkServer": map[string]any{
					"bindPort": 42113,
				},
				"tinkController": map[string]any{
					"enableLeaderElection": true,
				},
				"rufio": map[string]any{
					"enableLeaderElection": true,
				},
			},
		},
		"service": map[string]any{
			"type":           "LoadBalancer",
			"loadBalancerIP": tinkerbellIP,
			"lbClass":        "kube-vip.io/kube-vip-class",
		},
		"rbac": map[string]any{
			"type": "Role",
		},
		"optional": map[string]any{
			"hookos": map[string]any{
				"enabled": false,
			},
			"kubevip": map[string]any{
				"enabled": s.loadBalancer,
				"image":   kubevipImageURI,
				"additionalEnv": []map[string]string{
					{
						"name":  "prometheus_server",
						"value": ":2213",
					},
					{
						"name":  "lb_class_only",
						"value": "true",
					},
				},
			},
		},
	}

	// Set load balancer interface if specified
	if loadBalancerInterface != "" {
		valuesMap["optional"].(map[string]any)["kubevip"].(map[string]any)["interface"] = loadBalancerInterface
		valuesMap["deployment"].(map[string]any)["init"].(map[string]any)["sourceInterface"] = loadBalancerInterface
	}

	return valuesMap
}

// parseImageURI splits an image URI into image and tag components.
// Example: "public.ecr.aws/eks-anywhere/tinkerbell:v0.1.0" returns ("public.ecr.aws/eks-anywhere/tinkerbell", "v0.1.0").
func parseImageURI(uri string) (string, string) {
	// Handle both tag (:) and digest (@) separators
	if idx := strings.LastIndex(uri, "@"); idx != -1 {
		return uri[:idx], uri[idx+1:]
	}
	if idx := strings.LastIndex(uri, ":"); idx != -1 {
		// Make sure we're not splitting on the port in the registry URL
		// by checking if there's a / after the :
		afterColon := uri[idx+1:]
		if !strings.Contains(afterColon, "/") {
			return uri[:idx], afterColon
		}
	}
	return uri, "latest"
}
