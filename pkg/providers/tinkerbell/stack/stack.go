package stack

import (
	"context"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/utils/urls"
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
}

type Installer struct {
	docker          Docker
	filewriter      filewriter.FileWriter
	helm            Helm
	podCidrRange    string
	registryMirror  *v1alpha1.RegistryMirrorConfiguration
	namespace       string
	createNamespace bool
	bootsOnDocker   bool
	hostPort        bool
	loadBalancer    bool
	envoy           bool
}

type InstallOption func(s *Installer)

type StackInstaller interface {
	CleanupLocalBoots(ctx context.Context, forceCleanup bool) error
	Install(ctx context.Context, bundle releasev1alpha1.TinkerbellBundle, tinkerbellIP, kubeconfig, hookOverride string, opts ...InstallOption) error
	UninstallLocal(ctx context.Context) error
}

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

// NewInstaller returns a Tinkerbell StackInstaller which can be used to install or uninstall the Tinkerbell stack.
func NewInstaller(docker Docker, filewriter filewriter.FileWriter, helm Helm, namespace string, podCidrRange string, registryMirror *v1alpha1.RegistryMirrorConfiguration) StackInstaller {
	return &Installer{
		docker:         docker,
		filewriter:     filewriter,
		helm:           helm,
		registryMirror: registryMirror,
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

	bootEnv := []map[string]string{}
	for k, v := range s.getBootsEnv(bundle.TinkerbellStack, tinkerbellIP) {
		bootEnv = append(bootEnv, map[string]string{
			"name":  k,
			"value": v,
		})
	}

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
			deploy: true,
			image:  bundle.TinkerbellStack.Tink.TinkController.URI,
		},
		tinkServer: map[string]interface{}{
			deploy: true,
			image:  bundle.TinkerbellStack.Tink.TinkServer.URI,
			args:   []string{"--tls=false"},
			port: map[string]bool{
				hostPortEnabled: s.hostPort,
			},
		},
		hegel: map[string]interface{}{
			deploy: true,
			image:  bundle.TinkerbellStack.Hegel.URI,
			args:   []string{"--grpc-use-tls=false"},
			port: map[string]bool{
				hostPortEnabled: s.hostPort,
			},
			env: []map[string]string{
				{
					"name":  "TRUSTED_PROXIES",
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
			deploy: true,
			image:  bundle.TinkerbellStack.Rufio.URI,
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

	if s.registryMirror != nil && s.registryMirror.Authenticate {
		username, password, err := config.ReadCredentials()
		if err != nil {
			return err
		}
		endpoint := net.JoinHostPort(s.registryMirror.Endpoint, s.registryMirror.Port)
		if err := s.helm.RegistryLogin(ctx, endpoint, username, password); err != nil {
			return err
		}
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
	}

	for name, value := range s.getBootsEnv(bundle, tinkServerIP) {
		flags = append(flags, "-e", fmt.Sprintf("%s=%s", name, value))
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

func (s *Installer) getBootsEnv(bundle releasev1alpha1.TinkerbellStackBundle, tinkServerIP string) map[string]string {
	bootsEnv := map[string]string{
		"DATA_MODEL_VERSION":        "kubernetes",
		"TINKERBELL_TLS":            "false",
		"TINKERBELL_GRPC_AUTHORITY": fmt.Sprintf("%s:%s", tinkServerIP, grpcPort),
	}

	extraKernelArgs := fmt.Sprintf("tink_worker_image=%s", s.localRegistryURL(bundle.Tink.TinkWorker.URI))
	if s.registryMirror != nil {
		localRegistry := net.JoinHostPort(s.registryMirror.Endpoint, s.registryMirror.Port)
		extraKernelArgs = fmt.Sprintf("%s insecure_registries=%s", extraKernelArgs, localRegistry)
		if s.registryMirror.Authenticate {
			username, password, _ := config.ReadCredentials()
			bootsEnv["REGISTRY_USERNAME"] = username
			bootsEnv["REGISTRY_PASSWORD"] = password
		}
	}
	bootsEnv["BOOTS_EXTRA_KERNEL_ARGS"] = extraKernelArgs

	return bootsEnv
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
	return errors.New("boots container already exists, delete the container manually or re-run the command with --force-cleanup")
}

func (s *Installer) localRegistryURL(originalURL string) string {
	if s.registryMirror != nil {
		localRegistry := net.JoinHostPort(s.registryMirror.Endpoint, s.registryMirror.Port)
		return urls.ReplaceHost(originalURL, localRegistry)
	}
	return originalURL
}
