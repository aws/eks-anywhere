package dependencies

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/aws"
	"github.com/aws/eks-anywhere/pkg/awsiamauth"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cli"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	cliconfig "github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/diagnostics"
	"github.com/aws/eks-anywhere/pkg/eksd"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/executables/cmk"
	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	gitfactory "github.com/aws/eks-anywhere/pkg/git/factory"
	"github.com/aws/eks-anywhere/pkg/gitops/flux"
	"github.com/aws/eks-anywhere/pkg/govmomi"
	"github.com/aws/eks-anywhere/pkg/helm"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/manifests"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/networking/cilium"
	"github.com/aws/eks-anywhere/pkg/networking/kindnetd"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
	"github.com/aws/eks-anywhere/pkg/providers/nutanix"
	"github.com/aws/eks-anywhere/pkg/providers/snow"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/providers/validator"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/version"
	"github.com/aws/eks-anywhere/pkg/workflow/task/workload"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

type Dependencies struct {
	Logger                      logr.Logger
	Provider                    providers.Provider
	ClusterAwsCli               *executables.Clusterawsadm
	DockerClient                *executables.Docker
	Kubectl                     *executables.Kubectl
	Govc                        *executables.Govc
	CloudStackValidatorRegistry cloudstack.ValidatorRegistry
	SnowAwsClientRegistry       *snow.AwsClientRegistry
	SnowConfigManager           *snow.ConfigManager
	Writer                      filewriter.FileWriter
	Kind                        *executables.Kind
	Clusterctl                  *executables.Clusterctl
	Flux                        *executables.Flux
	Troubleshoot                *executables.Troubleshoot
	Helm                        *executables.Helm
	UnAuthKubeClient            *kubernetes.UnAuthClient
	Networking                  clustermanager.Networking
	CNIInstaller                workload.CNIInstaller
	CiliumTemplater             *cilium.Templater
	AwsIamAuth                  *awsiamauth.Installer
	ClusterManager              *clustermanager.ClusterManager
	KubernetesRetrierClient     *clustermanager.KubernetesRetrierClient
	Bootstrapper                *bootstrapper.Bootstrapper
	GitOpsFlux                  *flux.Flux
	Git                         *gitfactory.GitTools
	EksdInstaller               *eksd.Installer
	EksdUpgrader                *eksd.Upgrader
	ClusterApplier              clustermanager.Applier
	AnalyzerFactory             diagnostics.AnalyzerFactory
	CollectorFactory            diagnostics.CollectorFactory
	DignosticCollectorFactory   diagnostics.DiagnosticBundleFactory
	CAPIManager                 *clusterapi.Manager
	FileReader                  *files.Reader
	ManifestReader              *manifests.Reader
	closers                     []types.Closer
	CliConfig                   *cliconfig.CliConfig
	CreateCliConfig             *cliconfig.CreateClusterCLIConfig
	PackageInstaller            interfaces.PackageInstaller
	BundleRegistry              curatedpackages.BundleRegistry
	PackageControllerClient     *curatedpackages.PackageControllerClient
	PackageClient               curatedpackages.PackageHandler
	VSphereValidator            *vsphere.Validator
	VSphereDefaulter            *vsphere.Defaulter
	NutanixClientCache          *nutanix.ClientCache
	NutanixDefaulter            *nutanix.Defaulter
	NutanixValidator            *nutanix.Validator
	SnowValidator               *snow.Validator
	IPValidator                 *validator.IPValidator
	UnAuthKubectlClient         KubeClients
	HelmEnvClientFactory        *helm.EnvClientFactory
	ExecutableBuilder           *executables.ExecutablesBuilder
	CreateClusterDefaulter      cli.CreateClusterDefaulter
	UpgradeClusterDefaulter     cli.UpgradeClusterDefaulter
	KubeconfigWriter            kubeconfig.Writer
	ClusterCreator              *clustermanager.ClusterCreator
	EksaInstaller               *clustermanager.EKSAInstaller
	DeleteClusterDefaulter      cli.DeleteClusterDefaulter
	ClusterDeleter              clustermanager.Deleter
}

// KubeClients defines super struct that exposes all behavior.
type KubeClients struct {
	*executables.Kubectl
	*kubernetes.UnAuthClient
}

func (d *Dependencies) Close(ctx context.Context) error {
	// Reverse the loop so we close like LIFO
	for i := len(d.closers) - 1; i >= 0; i-- {
		if err := d.closers[i].Close(ctx); err != nil {
			return err
		}
	}

	return nil
}

// ForSpec constructs a Factory using the bundle referenced by clusterSpec.
func ForSpec(clusterSpec *cluster.Spec) *Factory {
	versionsBundle := clusterSpec.RootVersionsBundle()
	eksaToolsImage := versionsBundle.Eksa.CliTools
	return NewFactory().
		UseExecutableImage(eksaToolsImage.VersionedImage()).
		WithRegistryMirror(registrymirror.FromCluster(clusterSpec.Cluster)).
		UseProxyConfiguration(clusterSpec.Cluster.ProxyConfiguration()).
		WithWriterFolder(clusterSpec.Cluster.Name).
		WithDiagnosticCollectorImage(versionsBundle.Eksa.DiagnosticCollector.VersionedImage())
}

// Factory helps initialization.
type Factory struct {
	executablesConfig        *executablesConfig
	config                   config
	registryMirror           *registrymirror.RegistryMirror
	proxyConfiguration       map[string]string
	writerFolder             string
	diagnosticCollectorImage string
	buildSteps               []buildStep
	dependencies             Dependencies
}

type executablesConfig struct {
	builder            *executables.ExecutablesBuilder
	image              string
	useDockerContainer bool
	dockerClient       executables.DockerClient
	mountDirs          []string
}

type config struct {
	bundlesOverride string
	noTimeouts      bool
}

type buildStep func(ctx context.Context) error

func NewFactory() *Factory {
	return &Factory{
		writerFolder: "./",
		executablesConfig: &executablesConfig{
			useDockerContainer: executables.ExecutablesInDocker(),
		},
		buildSteps: make([]buildStep, 0),
	}
}

func (f *Factory) Build(ctx context.Context) (*Dependencies, error) {
	for _, step := range f.buildSteps {
		if err := step(ctx); err != nil {
			return nil, err
		}
	}

	// clean up stack
	f.buildSteps = make([]buildStep, 0)

	// Make copy of dependencies since its attributes are public
	d := f.dependencies

	return &d, nil
}

func (f *Factory) WithWriterFolder(folder string) *Factory {
	f.writerFolder = folder
	return f
}

// WithRegistryMirror configures the factory to use registry mirror wherever applicable.
func (f *Factory) WithRegistryMirror(registryMirror *registrymirror.RegistryMirror) *Factory {
	f.registryMirror = registryMirror

	return f
}

func (f *Factory) UseProxyConfiguration(proxyConfig map[string]string) *Factory {
	f.proxyConfiguration = proxyConfig
	return f
}

func (f *Factory) GetProxyConfiguration() map[string]string {
	return f.proxyConfiguration
}

func (f *Factory) WithProxyConfiguration() *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.proxyConfiguration == nil {
			proxyConfig := cliconfig.GetProxyConfigFromEnv()
			f.UseProxyConfiguration(proxyConfig)
		}
		return nil
	},
	)

	return f
}

func (f *Factory) UseExecutableImage(image string) *Factory {
	f.executablesConfig.image = image
	return f
}

// WithExecutableImage sets the right cli tools image for the executable builder, reading
// from the Bundle and using the first VersionsBundle
// This is just the default for when there is not an specific kubernetes version available
// For commands that receive a cluster config file or a kubernetes version directly as input,
// use UseExecutableImage to specify the image directly.
func (f *Factory) WithExecutableImage() *Factory {
	f.WithManifestReader()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.executablesConfig.image != "" {
			return nil
		}

		if f.config.bundlesOverride != "" {
			image, err := f.selectImageFromBundleOverride(f.config.bundlesOverride)
			if err != nil {
				return err
			}
			f.executablesConfig.image = image
			return nil
		}
		bundles, err := f.dependencies.ManifestReader.ReadBundlesForVersion(version.Get().GitVersion)
		if err != nil {
			return fmt.Errorf("retrieving executable tools image from bundle in dependency factory: %v", err)
		}

		f.executablesConfig.image = bundles.DefaultEksAToolsImage().VersionedImage()
		return nil
	})

	return f
}

// selectImageFromBundleOverride retrieves an image from a bundles override.
//
// Handles cases where the bundle is configured with an override.
func (f *Factory) selectImageFromBundleOverride(bundlesOverride string) (string, error) {
	releaseBundles, err := bundles.Read(f.dependencies.ManifestReader, bundlesOverride)
	if err != nil {
		return "", fmt.Errorf("retrieving executable tools image from overridden bundle in dependency factory %v", err)
	}
	// Note: Currently using the first available version of the cli tools
	// This is because the binaries bundled are all the same version hence no compatibility concerns
	// In case, there is a change to this behavior, there might be a need to reassess this item
	return releaseBundles.DefaultEksAToolsImage().VersionedImage(), nil
}

// WithCustomBundles allows configuring a bundle override.
func (f *Factory) WithCustomBundles(bundlesOverride string) *Factory {
	if bundlesOverride == "" {
		return f
	}
	f.config.bundlesOverride = bundlesOverride
	f.WithExecutableImage()
	return f
}

func (f *Factory) WithExecutableMountDirs(mountDirs ...string) *Factory {
	f.executablesConfig.mountDirs = mountDirs
	return f
}

func (f *Factory) WithLocalExecutables() *Factory {
	f.executablesConfig.useDockerContainer = false
	return f
}

// UseExecutablesDockerClient forces a specific DockerClient to build
// Executables as opposed to follow the normal building flow
// This is only for testing.
func (f *Factory) UseExecutablesDockerClient(client executables.DockerClient) *Factory {
	f.executablesConfig.dockerClient = client
	return f
}

// dockerLogin performs a docker login with the ENV VARS.
func dockerLogin(ctx context.Context, registry string, docker executables.DockerClient) error {
	username, password, _ := cliconfig.ReadCredentials()
	err := docker.Login(ctx, registry, username, password)
	if err != nil {
		return err
	}
	return nil
}

// WithDockerLogin adds a docker login to the build steps.
func (f *Factory) WithDockerLogin() *Factory {
	f.WithDocker()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.registryMirror != nil {
			err := dockerLogin(ctx, f.registryMirror.BaseRegistry, f.executablesConfig.dockerClient)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return f
}

func (f *Factory) WithExecutableBuilder() *Factory {
	if f.executablesConfig.useDockerContainer {
		f.WithExecutableImage().WithDocker()
		if f.registryMirror != nil && f.registryMirror.Auth {
			f.WithDockerLogin()
		}
	}

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.executablesConfig.builder != nil {
			return nil
		}

		if f.executablesConfig.useDockerContainer {
			image := f.executablesConfig.image
			if f.registryMirror != nil {
				image = f.registryMirror.ReplaceRegistry(image)
			}
			b, err := executables.NewInDockerExecutablesBuilder(
				f.executablesConfig.dockerClient,
				image,
				f.executablesConfig.mountDirs...,
			)
			if err != nil {
				return err
			}
			f.executablesConfig.builder = b
		} else {
			f.executablesConfig.builder = executables.NewLocalExecutablesBuilder()
		}

		f.dependencies.ExecutableBuilder = f.executablesConfig.builder

		closer, err := f.executablesConfig.builder.Init(ctx)
		if err != nil {
			return err
		}
		if f.registryMirror != nil && f.registryMirror.Auth {
			docker := f.executablesConfig.builder.BuildDockerExecutable()
			err := dockerLogin(ctx, f.registryMirror.BaseRegistry, docker)
			if err != nil {
				return err
			}
		}
		f.dependencies.closers = append(f.dependencies.closers, closer)

		return nil
	})

	return f
}

// WithHelmExecutableBuilder adds a build step to initializes the helm.ExecutableBuilder dependency.
func (f *Factory) WithHelmExecutableBuilder() *Factory {
	f.WithExecutableBuilder()
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.ExecutableBuilder != nil {
			return nil
		}

		f.dependencies.ExecutableBuilder = f.executablesConfig.builder
		return nil
	})

	return f
}

// ProviderOptions contains per provider options.
type ProviderOptions struct {
	// Tinkerbell contains Tinkerbell specific options.
	Tinkerbell *TinkerbellOptions
}

// TinkerbellOptions contains Tinkerbell specific options.
type TinkerbellOptions struct {
	// BMCOptions contains options for configuring BMC interactions.
	BMCOptions *hardware.BMCOptions
}

// WithProvider initializes the provider dependency and adds to the build steps.
func (f *Factory) WithProvider(clusterConfigFile string, clusterConfig *v1alpha1.Cluster, skipIPCheck bool, hardwareCSVPath string, force bool, tinkerbellBootstrapIP string, skippedValidations map[string]bool, opts *ProviderOptions) *Factory { // nolint:gocyclo
	switch clusterConfig.Spec.DatacenterRef.Kind {
	case v1alpha1.VSphereDatacenterKind:
		f.WithKubectl().WithGovc().WithWriter().WithIPValidator()
	case v1alpha1.CloudStackDatacenterKind:
		f.WithKubectl().WithCloudStackValidatorRegistry(skipIPCheck).WithWriter()
	case v1alpha1.DockerDatacenterKind:
		f.WithDocker().WithKubectl()
	case v1alpha1.TinkerbellDatacenterKind:
		if clusterConfig.Spec.RegistryMirrorConfiguration != nil {
			f.WithDocker().WithKubectl().WithWriter().WithHelm(helm.WithInsecure())
		} else {
			f.WithDocker().WithKubectl().WithWriter().WithHelm()
		}
	case v1alpha1.SnowDatacenterKind:
		f.WithUnAuthKubeClient().WithSnowConfigManager()
	case v1alpha1.NutanixDatacenterKind:
		f.WithKubectl().WithNutanixClientCache().WithNutanixDefaulter().WithNutanixValidator().WithIPValidator()
	}

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Provider != nil {
			return nil
		}

		switch clusterConfig.Spec.DatacenterRef.Kind {
		case v1alpha1.VSphereDatacenterKind:
			datacenterConfig, err := v1alpha1.GetVSphereDatacenterConfig(clusterConfigFile)
			if err != nil {
				return fmt.Errorf("unable to get datacenter config from file %s: %v", clusterConfigFile, err)
			}

			f.dependencies.Provider = vsphere.NewProvider(
				datacenterConfig,
				clusterConfig,
				f.dependencies.Govc,
				f.dependencies.Kubectl,
				f.dependencies.Writer,
				f.dependencies.IPValidator,
				time.Now,
				skipIPCheck,
				skippedValidations,
			)

		case v1alpha1.CloudStackDatacenterKind:
			datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(clusterConfigFile)
			if err != nil {
				return fmt.Errorf("unable to get datacenter config from file %s: %v", clusterConfigFile, err)
			}

			execConfig, err := decoder.ParseCloudStackCredsFromEnv()
			if err != nil {
				return fmt.Errorf("parsing CloudStack credentials: %v", err)
			}
			validator, err := f.dependencies.CloudStackValidatorRegistry.Get(execConfig)
			if err != nil {
				return fmt.Errorf("building validator from exec config: %v", err)
			}
			f.dependencies.Provider = cloudstack.NewProvider(datacenterConfig, clusterConfig, f.dependencies.Kubectl, validator, f.dependencies.Writer, time.Now, logger.Get())

		case v1alpha1.SnowDatacenterKind:
			f.dependencies.Provider = snow.NewProvider(
				f.dependencies.UnAuthKubeClient,
				f.dependencies.SnowConfigManager,
				skipIPCheck,
			)

		case v1alpha1.TinkerbellDatacenterKind:
			datacenterConfig, err := v1alpha1.GetTinkerbellDatacenterConfig(clusterConfigFile)
			if err != nil {
				return fmt.Errorf("unable to get datacenter config from file %s: %v", clusterConfigFile, err)
			}

			config, err := cluster.ParseConfigFromFile(clusterConfigFile)
			if err != nil {
				return fmt.Errorf("unable to get machine config from file %s: %v", clusterConfigFile, err)
			}

			machineConfigs := config.TinkerbellMachineConfigs
			tinkerbellIP := tinkerbellBootstrapIP
			if tinkerbellIP == "" {
				logger.V(4).Info("Inferring local Tinkerbell Bootstrap IP from environment")
				localIp, err := networkutils.GetLocalIP()
				if err != nil {
					return err
				}
				tinkerbellIP = localIp.String()
			}
			logger.V(4).Info("Tinkerbell IP", "tinkerbell-ip", tinkerbellIP)

			provider, err := tinkerbell.NewProvider(
				datacenterConfig,
				machineConfigs,
				clusterConfig,
				hardwareCSVPath,
				f.dependencies.Writer,
				f.dependencies.DockerClient,
				f.dependencies.Helm,
				f.dependencies.Kubectl,
				tinkerbellIP,
				time.Now,
				force,
				skipIPCheck,
			)
			if err != nil {
				return err
			}
			if opts != nil && opts.Tinkerbell != nil && opts.Tinkerbell.BMCOptions != nil {
				provider.BMCOptions = opts.Tinkerbell.BMCOptions
			}

			f.dependencies.Provider = provider

		case v1alpha1.DockerDatacenterKind:
			datacenterConfig, err := v1alpha1.GetDockerDatacenterConfig(clusterConfigFile)
			if err != nil {
				return fmt.Errorf("unable to get datacenter config from file %s: %v", clusterConfigFile, err)
			}

			f.dependencies.Provider = docker.NewProvider(
				datacenterConfig,
				f.dependencies.DockerClient,
				f.dependencies.Kubectl,
				time.Now,
			)
		case v1alpha1.NutanixDatacenterKind:
			datacenterConfig, err := v1alpha1.GetNutanixDatacenterConfig(clusterConfigFile)
			if err != nil {
				return fmt.Errorf("unable to get datacenter config from file %s: %v", clusterConfigFile, err)
			}

			config, err := cluster.ParseConfigFromFile(clusterConfigFile)
			if err != nil {
				return fmt.Errorf("unable to get machine config from file %s: %v", clusterConfigFile, err)
			}

			machineConfigs := config.NutanixMachineConfigs
			skipVerifyTransport := http.DefaultTransport.(*http.Transport).Clone()
			skipVerifyTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			httpClient := &http.Client{Transport: skipVerifyTransport}
			provider := nutanix.NewProvider(
				datacenterConfig,
				machineConfigs,
				clusterConfig,
				f.dependencies.Kubectl,
				f.dependencies.Writer,
				f.dependencies.NutanixClientCache,
				f.dependencies.IPValidator,
				crypto.NewTlsValidator(),
				httpClient,
				time.Now,
				skipIPCheck,
			)
			f.dependencies.Provider = provider
		default:
			return fmt.Errorf("no provider support for datacenter kind: %s", clusterConfig.Spec.DatacenterRef.Kind)
		}

		return nil
	})

	return f
}

// WithKubeconfigWriter adds the KubeconfigReader dependency depending on the provider.
func (f *Factory) WithKubeconfigWriter(clusterConfig *v1alpha1.Cluster) *Factory {
	f.WithUnAuthKubeClient()
	if clusterConfig.Spec.DatacenterRef.Kind == v1alpha1.DockerDatacenterKind {
		f.WithDocker()
	}

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.KubeconfigWriter != nil {
			return nil
		}
		writer := kubeconfig.NewClusterAPIKubeconfigSecretWriter(f.dependencies.UnAuthKubeClient)
		switch clusterConfig.Spec.DatacenterRef.Kind {
		case v1alpha1.DockerDatacenterKind:
			f.dependencies.KubeconfigWriter = docker.NewKubeconfigWriter(f.dependencies.DockerClient, writer)
		default:
			f.dependencies.KubeconfigWriter = writer
		}
		return nil
	})

	return f
}

// WithClusterCreator adds the ClusterCreator dependency.
func (f *Factory) WithClusterCreator(clusterConfig *v1alpha1.Cluster) *Factory {
	f.WithClusterApplier().WithWriter().WithKubeconfigWriter(clusterConfig)
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.ClusterCreator != nil {
			return nil
		}

		f.dependencies.ClusterCreator = clustermanager.NewClusterCreator(f.dependencies.ClusterApplier, f.dependencies.KubeconfigWriter, f.dependencies.Writer)
		return nil
	})

	return f
}

func (f *Factory) WithDocker() *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.DockerClient != nil {
			return nil
		}

		f.dependencies.DockerClient = executables.BuildDockerExecutable()
		if f.executablesConfig.dockerClient == nil {
			f.executablesConfig.dockerClient = f.dependencies.DockerClient
		}

		return nil
	})

	return f
}

func (f *Factory) WithKubectl() *Factory {
	f.WithExecutableBuilder()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Kubectl != nil {
			return nil
		}

		f.dependencies.Kubectl = f.executablesConfig.builder.BuildKubectlExecutable()
		return nil
	})

	return f
}

func (f *Factory) WithGovc() *Factory {
	f.WithExecutableBuilder().WithWriter()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Govc != nil {
			return nil
		}

		f.dependencies.Govc = f.executablesConfig.builder.BuildGovcExecutable(f.dependencies.Writer)
		f.dependencies.closers = append(f.dependencies.closers, f.dependencies.Govc)

		return nil
	})

	return f
}

// WithCloudStackValidatorRegistry initializes the CloudStack validator for the object being constructed to make it available in the constructor.
func (f *Factory) WithCloudStackValidatorRegistry(skipIPCheck bool) *Factory {
	f.WithExecutableBuilder().WithWriter()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.CloudStackValidatorRegistry != nil {
			return nil
		}

		cmkBuilder := cmk.NewCmkBuilder(f.executablesConfig.builder)
		f.dependencies.CloudStackValidatorRegistry = cloudstack.NewValidatorFactory(cmkBuilder, f.dependencies.Writer, skipIPCheck)

		return nil
	})

	return f
}

func (f *Factory) WithSnowConfigManager() *Factory {
	f.WithAwsSnow().WithWriter()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.SnowConfigManager != nil {
			return nil
		}

		client := aws.NewClient()
		if err := client.BuildIMDS(ctx); err != nil {
			return err
		}

		validator := snow.NewValidator(f.dependencies.SnowAwsClientRegistry, snow.WithIMDS(client))
		defaulters := snow.NewDefaulters(f.dependencies.SnowAwsClientRegistry, f.dependencies.Writer)

		f.dependencies.SnowConfigManager = snow.NewConfigManager(defaulters, validator)

		return nil
	})

	return f
}

func (f *Factory) WithAwsSnow() *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.SnowAwsClientRegistry != nil {
			return nil
		}

		clientRegistry := snow.NewAwsClientRegistry()
		err := clientRegistry.Build(ctx)
		if err != nil {
			return err
		}
		f.dependencies.SnowAwsClientRegistry = clientRegistry

		return nil
	})

	return f
}

func (f *Factory) WithWriter() *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Writer != nil {
			return nil
		}

		var err error
		f.dependencies.Writer, err = filewriter.NewWriter(f.writerFolder)
		if err != nil {
			return err
		}

		return nil
	})

	return f
}

func (f *Factory) WithKind() *Factory {
	f.WithExecutableBuilder().WithWriter()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Kind != nil {
			return nil
		}

		f.dependencies.Kind = f.executablesConfig.builder.BuildKindExecutable(f.dependencies.Writer)
		return nil
	})

	return f
}

func (f *Factory) WithClusterctl() *Factory {
	f.WithExecutableBuilder().WithWriter().WithFileReader()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Clusterctl != nil {
			return nil
		}

		f.dependencies.Clusterctl = f.executablesConfig.builder.BuildClusterCtlExecutable(
			f.dependencies.Writer,
			f.dependencies.FileReader,
		)
		return nil
	})

	return f
}

func (f *Factory) WithFlux() *Factory {
	f.WithExecutableBuilder()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Flux != nil {
			return nil
		}

		f.dependencies.Flux = f.executablesConfig.builder.BuildFluxExecutable()
		return nil
	})

	return f
}

func (f *Factory) WithTroubleshoot() *Factory {
	f.WithExecutableBuilder()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Troubleshoot != nil {
			return nil
		}

		f.dependencies.Troubleshoot = f.executablesConfig.builder.BuildTroubleshootExecutable()
		return nil
	})

	return f
}

// WithHelm initializes a new Helm executable as a factory dependency.
func (f *Factory) WithHelm(opts ...helm.Opt) *Factory {
	f.WithExecutableBuilder().WithProxyConfiguration()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.registryMirror != nil {
			opts = append(opts, helm.WithRegistryMirror(f.registryMirror))
		}

		if f.proxyConfiguration != nil {
			opts = append(opts, helm.WithProxyConfig(f.proxyConfiguration))
		}

		f.dependencies.Helm = f.executablesConfig.builder.BuildHelmExecutable(opts...)
		return nil
	})

	return f
}

// WithHelmEnvClientFactory configures the HelmEnvClientFactory dependency with a helm.EnvClientFactory.
func (f *Factory) WithHelmEnvClientFactory(opts ...helm.Opt) *Factory {
	f.WithExecutableBuilder().WithProxyConfiguration()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.HelmEnvClientFactory != nil {
			return nil
		}

		if f.proxyConfiguration != nil {
			opts = append(opts, helm.WithProxyConfig(f.proxyConfiguration))
		}

		envClientFactory := helm.NewEnvClientFactory(f.executablesConfig.builder)
		err := envClientFactory.Init(ctx, f.registryMirror, opts...)
		if err != nil {
			return fmt.Errorf("building helm env client factory: %v", err)
		}

		f.dependencies.HelmEnvClientFactory = envClientFactory
		return nil
	})

	return f
}

// WithNetworking builds a Networking.
func (f *Factory) WithNetworking(clusterConfig *v1alpha1.Cluster) *Factory {
	var networkingBuilder func() clustermanager.Networking
	if clusterConfig.Spec.ClusterNetwork.CNIConfig.Kindnetd != nil {
		f.WithKubectl().WithFileReader()
		networkingBuilder = func() clustermanager.Networking {
			return kindnetd.NewKindnetd(f.dependencies.Kubectl, f.dependencies.FileReader)
		}
	} else {
		f.WithKubectl().WithCiliumTemplater()

		networkingBuilder = func() clustermanager.Networking {
			var opts []cilium.RetrierClientOpt
			if f.config.noTimeouts {
				opts = append(opts, cilium.RetrierClientRetrier(retrier.NewWithNoTimeout()))
			}

			c := cilium.NewCilium(
				cilium.NewRetrier(f.dependencies.Kubectl, opts...),
				f.dependencies.CiliumTemplater,
			)
			c.SetSkipUpgrade(!clusterConfig.Spec.ClusterNetwork.CNIConfig.Cilium.IsManaged())
			return c
		}
	}

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Networking != nil {
			return nil
		}
		f.dependencies.Networking = networkingBuilder()

		return nil
	})

	return f
}

// WithCNIInstaller builds a CNI installer for the given cluster.
func (f *Factory) WithCNIInstaller(spec *cluster.Spec, provider providers.Provider) *Factory {
	if spec.Cluster.Spec.ClusterNetwork.CNIConfig.Kindnetd != nil {
		f.WithKubectl().WithFileReader()
	} else {
		f.WithKubectl().WithCiliumTemplater()
	}

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.CNIInstaller != nil {
			return nil
		}

		if spec.Cluster.Spec.ClusterNetwork.CNIConfig.Kindnetd != nil {
			f.dependencies.CNIInstaller = kindnetd.NewInstallerForSpec(
				f.dependencies.Kubectl,
				f.dependencies.FileReader,
				spec,
			)
		} else {
			f.dependencies.CNIInstaller = cilium.NewInstallerForSpec(
				cilium.NewRetrier(f.dependencies.Kubectl),
				f.dependencies.CiliumTemplater,
				cilium.Config{
					Spec:              spec,
					AllowedNamespaces: maps.Keys(provider.GetDeployments()),
				},
			)
		}

		return nil
	})

	return f
}

func (f *Factory) WithCiliumTemplater() *Factory {
	f.WithHelmEnvClientFactory(helm.WithInsecure())

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.CiliumTemplater != nil {
			return nil
		}
		f.dependencies.CiliumTemplater = cilium.NewTemplater(f.dependencies.HelmEnvClientFactory)

		return nil
	})

	return f
}

func (f *Factory) WithAwsIamAuth() *Factory {
	f.WithKubectl().WithWriter()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.AwsIamAuth != nil {
			return nil
		}
		certgen := crypto.NewCertificateGenerator()
		clusterId := uuid.New()

		var opts []awsiamauth.RetrierClientOpt
		if f.config.noTimeouts {
			opts = append(opts, awsiamauth.RetrierClientRetrier(*retrier.NewWithNoTimeout()))
		}

		f.dependencies.AwsIamAuth = awsiamauth.NewInstaller(
			certgen,
			clusterId,
			awsiamauth.NewRetrierClient(f.dependencies.Kubectl, opts...),
			f.dependencies.Writer,
		)
		return nil
	})

	return f
}

// WithIPValidator builds the IPValidator for the given cluster.
func (f *Factory) WithIPValidator() *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.IPValidator != nil {
			return nil
		}
		f.dependencies.IPValidator = validator.NewIPValidator()
		return nil
	})
	return f
}

type bootstrapperClient struct {
	*executables.Kind
	*executables.Kubectl
}

func (f *Factory) WithBootstrapper() *Factory {
	f.WithKind().WithKubectl()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Bootstrapper != nil {
			return nil
		}

		var opts []bootstrapper.RetrierClientOpt
		if f.config.noTimeouts {
			opts = append(opts,
				bootstrapper.WithRetrierClientRetrier(
					*retrier.NewWithNoTimeout(),
				),
			)
		}

		f.dependencies.Bootstrapper = bootstrapper.New(
			bootstrapper.NewRetrierClient(
				f.dependencies.Kind,
				f.dependencies.Kubectl,
				opts...,
			),
		)
		return nil
	})

	return f
}

type clusterManagerClient struct {
	*executables.Clusterctl
	*clustermanager.KubernetesRetrierClient
}

// ClusterManagerTimeoutOptions maintains the timeout options for cluster manager.
type ClusterManagerTimeoutOptions struct {
	NoTimeouts bool

	ControlPlaneWait, ExternalEtcdWait, MachineWait, UnhealthyMachineWait, NodeStartupWait time.Duration
}

func (f *Factory) eksaInstallerOpts() []clustermanager.EKSAInstallerOpt {
	var opts []clustermanager.EKSAInstallerOpt
	if f.config.noTimeouts {
		opts = append(opts, clustermanager.WithEKSAInstallerNoTimeouts())
	}
	return opts
}

func (f *Factory) clusterManagerOpts(timeoutOpts *ClusterManagerTimeoutOptions) []clustermanager.ClusterManagerOpt {
	if timeoutOpts == nil {
		return nil
	}

	o := []clustermanager.ClusterManagerOpt{
		clustermanager.WithControlPlaneWaitTimeout(timeoutOpts.ControlPlaneWait),
		clustermanager.WithExternalEtcdWaitTimeout(timeoutOpts.ExternalEtcdWait),
		clustermanager.WithMachineMaxWait(timeoutOpts.MachineWait),
		clustermanager.WithUnhealthyMachineTimeout(timeoutOpts.UnhealthyMachineWait),
		clustermanager.WithNodeStartupTimeout(timeoutOpts.NodeStartupWait),
	}

	if f.config.noTimeouts {
		o = append(o, clustermanager.WithNoTimeouts())
	}

	return o
}

// WithClusterManager builds a cluster manager based on the cluster config and timeout options.
func (f *Factory) WithClusterManager(clusterConfig *v1alpha1.Cluster, timeoutOpts *ClusterManagerTimeoutOptions) *Factory {
	f.WithClusterctl().WithNetworking(clusterConfig).WithWriter().WithDiagnosticBundleFactory().WithAwsIamAuth().WithFileReader().WithUnAuthKubeClient().WithKubernetesRetrierClient().WithEKSAInstaller()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.ClusterManager != nil {
			return nil
		}

		client := clusterManagerClient{
			f.dependencies.Clusterctl,
			f.dependencies.KubernetesRetrierClient,
		}

		f.dependencies.ClusterManager = clustermanager.New(
			f.dependencies.UnAuthKubeClient,
			client,
			f.dependencies.Networking,
			f.dependencies.Writer,
			f.dependencies.DignosticCollectorFactory,
			f.dependencies.AwsIamAuth,
			f.dependencies.EksaInstaller,
			f.clusterManagerOpts(timeoutOpts)...,
		)
		return nil
	})

	return f
}

// WithEKSAInstaller builds a cluster manager based on the cluster config and timeout options.
func (f *Factory) WithEKSAInstaller() *Factory {
	f.WithFileReader().WithKubernetesRetrierClient()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.EksaInstaller != nil {
			return nil
		}

		installer := clustermanager.NewEKSAInstaller(f.dependencies.KubernetesRetrierClient, f.dependencies.FileReader, f.eksaInstallerOpts()...)

		f.dependencies.EksaInstaller = installer
		return nil
	})

	return f
}

// WithKubernetesRetrierClient builds a cluster manager based on the cluster config and timeout options.
func (f *Factory) WithKubernetesRetrierClient() *Factory {
	f.WithKubectl()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.KubernetesRetrierClient != nil {
			return nil
		}
		var r *retrier.Retrier
		if f.config.noTimeouts {
			r = retrier.NewWithNoTimeout()
		} else {
			r = clustermanager.DefaultRetrier()
		}

		retrierClient := clustermanager.NewRetrierClient(
			f.dependencies.Kubectl,
			r,
		)

		f.dependencies.KubernetesRetrierClient = retrierClient

		return nil
	})

	return f
}

// WithNoTimeouts injects no timeouts to all the dependencies with configurable timeout.
// Calling this method sets no timeout for the waits and retries in all the
// cluster operations, i.e. cluster manager, eksa installer, networking installer.
// Instead of passing the option to each dependency's constructor, use this
// method to pass no timeouts to new dependency.
func (f *Factory) WithNoTimeouts() *Factory {
	f.config.noTimeouts = true
	return f
}

// WithCliConfig builds a cli config.
func (f *Factory) WithCliConfig(cliConfig *cliconfig.CliConfig) *Factory {
	f.dependencies.CliConfig = cliConfig
	return f
}

// WithCreateClusterDefaulter builds a create cluster defaulter that builds defaulter dependencies specific to the create cluster command. The defaulter is then run once the factory is built in the create cluster command.
func (f *Factory) WithCreateClusterDefaulter(createCliConfig *cliconfig.CreateClusterCLIConfig) *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		controlPlaneIPCheckAnnotationDefaulter := cluster.NewControlPlaneIPCheckAnnotationDefaulter(createCliConfig.SkipCPIPCheck)
		machineHealthCheckDefaulter := cluster.NewMachineHealthCheckDefaulter(createCliConfig.NodeStartupTimeout, createCliConfig.UnhealthyMachineTimeout)

		createClusterDefaulter := cli.NewCreateClusterDefaulter(controlPlaneIPCheckAnnotationDefaulter, machineHealthCheckDefaulter)

		f.dependencies.CreateClusterDefaulter = createClusterDefaulter

		return nil
	})

	return f
}

// WithUpgradeClusterDefaulter builds a create cluster defaulter that builds defaulter dependencies specific to the create cluster command. The defaulter is then run once the factory is built in the create cluster command.
func (f *Factory) WithUpgradeClusterDefaulter(upgradeCliConfig *cliconfig.UpgradeClusterCLIConfig) *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		machineHealthCheckDefaulter := cluster.NewMachineHealthCheckDefaulter(upgradeCliConfig.NodeStartupTimeout, upgradeCliConfig.UnhealthyMachineTimeout)

		upgradeClusterDefaulter := cli.NewUpgradeClusterDefaulter(machineHealthCheckDefaulter)

		f.dependencies.UpgradeClusterDefaulter = upgradeClusterDefaulter

		return nil
	})

	return f
}

// WithDeleteClusterDefaulter builds a delete cluster defaulter that builds defaulter dependencies specific to the delete cluster command. The defaulter is then run once the factory is built in the delete cluster command.
func (f *Factory) WithDeleteClusterDefaulter(deleteCliConfig *cliconfig.DeleteClusterCLIConfig) *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		nsDefaulter := cluster.NewNamespaceDefaulter(deleteCliConfig.ClusterNamespace)
		deleteClusterDefaulter := cli.NewDeleteClusterDefaulter(nsDefaulter)

		f.dependencies.DeleteClusterDefaulter = deleteClusterDefaulter

		return nil
	})

	return f
}

type eksdInstallerClient struct {
	*executables.Kubectl
}

func (f *Factory) WithEksdInstaller() *Factory {
	f.WithKubectl().WithFileReader()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.EksdInstaller != nil {
			return nil
		}

		var opts []eksd.InstallerOpt
		if f.config.noTimeouts {
			opts = append(opts, eksd.WithRetrier(retrier.NewWithNoTimeout()))
		}

		f.dependencies.EksdInstaller = eksd.NewEksdInstaller(
			&eksdInstallerClient{
				f.dependencies.Kubectl,
			},
			f.dependencies.FileReader,
			opts...,
		)
		return nil
	})

	return f
}

func (f *Factory) WithEksdUpgrader() *Factory {
	f.WithKubectl().WithFileReader()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.EksdUpgrader != nil {
			return nil
		}

		var opts []eksd.InstallerOpt
		if f.config.noTimeouts {
			opts = append(opts, eksd.WithRetrier(retrier.NewWithNoTimeout()))
		}

		f.dependencies.EksdUpgrader = eksd.NewUpgrader(
			&eksdInstallerClient{
				f.dependencies.Kubectl,
			},
			f.dependencies.FileReader,
			opts...,
		)
		return nil
	})

	return f
}

// WithClusterApplier builds a cluster applier.
func (f *Factory) WithClusterApplier() *Factory {
	f.WithLogger().WithUnAuthKubeClient().WithLogger()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		var opts []clustermanager.ApplierOpt
		if f.config.noTimeouts {
			// opts = append(opts, clustermanager.ManagementUpgraderRetrier(*retrier.NewWithNoTimeout()))
			opts = append(opts, clustermanager.WithApplierNoTimeouts())
		}

		f.dependencies.ClusterApplier = clustermanager.NewApplier(
			f.dependencies.Logger,
			f.dependencies.UnAuthKubeClient,
			opts...,
		)
		return nil
	})
	return f
}

// WithClusterDeleter builds a cluster deleter.
func (f *Factory) WithClusterDeleter() *Factory {
	f.WithLogger().WithUnAuthKubeClient().WithLogger()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		var opts []clustermanager.DeleterOpt
		if f.config.noTimeouts {
			opts = append(opts, clustermanager.WithDeleterNoTimeouts())
		}

		f.dependencies.ClusterDeleter = clustermanager.NewDeleter(
			f.dependencies.Logger,
			f.dependencies.UnAuthKubeClient,
			opts...,
		)
		return nil
	})
	return f
}

// WithValidatorClients builds KubeClients.
func (f *Factory) WithValidatorClients() *Factory {
	f.WithKubectl().WithUnAuthKubeClient()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		f.dependencies.UnAuthKubectlClient = KubeClients{
			Kubectl:      f.dependencies.Kubectl,
			UnAuthClient: f.dependencies.UnAuthKubeClient,
		}

		return nil
	})

	return f
}

// WithLogger setups a logger to be injected in constructors. It uses the logger
// package level logger.
func (f *Factory) WithLogger() *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		f.dependencies.Logger = logger.Get()
		return nil
	})
	return f
}

func (f *Factory) WithGit(clusterConfig *v1alpha1.Cluster, fluxConfig *v1alpha1.FluxConfig) *Factory {
	f.WithWriter()
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.Git != nil {
			return nil
		}

		if fluxConfig == nil {
			return nil
		}

		tools, err := gitfactory.Build(ctx, clusterConfig, fluxConfig, f.dependencies.Writer)
		if err != nil {
			return fmt.Errorf("creating Git provider: %v", err)
		}

		if fluxConfig.Spec.Git != nil {
			err = tools.Client.ValidateRemoteExists(ctx)
			if err != nil {
				return err
			}
		}

		if tools.Provider != nil {
			err = tools.Provider.Validate(ctx)
			if err != nil {
				return fmt.Errorf("validating provider: %v", err)
			}
		}

		f.dependencies.Git = tools
		return nil
	})
	return f
}

// WithGitOpsFlux builds a gitops flux.
func (f *Factory) WithGitOpsFlux(clusterConfig *v1alpha1.Cluster, fluxConfig *v1alpha1.FluxConfig, cliConfig *cliconfig.CliConfig) *Factory {
	f.WithWriter().WithFlux().WithKubectl().WithGit(clusterConfig, fluxConfig)

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.GitOpsFlux != nil {
			return nil
		}

		f.dependencies.GitOpsFlux = flux.NewFlux(f.dependencies.Flux, f.dependencies.Kubectl, f.dependencies.Git, cliConfig)

		return nil
	})

	return f
}

func (f *Factory) WithPackageInstaller(spec *cluster.Spec, packagesLocation, kubeConfig string) *Factory {
	f.WithKubectl().WithPackageControllerClient(spec, kubeConfig).WithPackageClient()
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.PackageInstaller != nil {
			return nil
		}
		managementClusterName := getManagementClusterName(spec)
		mgmtKubeConfig := kubeconfig.ResolveFilename(kubeConfig, managementClusterName)

		f.dependencies.PackageInstaller = curatedpackages.NewInstaller(
			f.dependencies.Kubectl,
			f.dependencies.PackageClient,
			f.dependencies.PackageControllerClient,
			spec,
			packagesLocation,
			mgmtKubeConfig,
		)
		return nil
	})
	return f
}

func (f *Factory) WithPackageControllerClient(spec *cluster.Spec, kubeConfig string) *Factory {
	f.WithHelm(helm.WithInsecure()).WithKubectl()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.PackageControllerClient != nil || spec == nil {
			return nil
		}
		managementClusterName := getManagementClusterName(spec)
		mgmtKubeConfig := kubeconfig.ResolveFilename(kubeConfig, managementClusterName)

		httpProxy, httpsProxy, noProxy := getProxyConfiguration(spec)
		eksaAccessKeyID, eksaSecretKey, eksaRegion := os.Getenv(cliconfig.EksaAccessKeyIdEnv), os.Getenv(cliconfig.EksaSecretAccessKeyEnv), os.Getenv(cliconfig.EksaRegionEnv)

		eksaAwsConfig := ""
		p := os.Getenv(cliconfig.EksaAwsConfigFileEnv)
		if p != "" {
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			eksaAwsConfig = string(b)
		}
		writer, err := filewriter.NewWriter(spec.Cluster.Name)
		if err != nil {
			return err
		}
		bundle := spec.RootVersionsBundle()
		if bundle == nil {
			return fmt.Errorf("could not find VersionsBundle")
		}
		f.dependencies.PackageControllerClient = curatedpackages.NewPackageControllerClient(
			f.dependencies.Helm,
			f.dependencies.Kubectl,
			spec.Cluster.Name,
			mgmtKubeConfig,
			&bundle.PackageController.HelmChart,
			f.registryMirror,
			curatedpackages.WithEksaAccessKeyId(eksaAccessKeyID),
			curatedpackages.WithEksaSecretAccessKey(eksaSecretKey),
			curatedpackages.WithEksaRegion(eksaRegion),
			curatedpackages.WithEksaAwsConfig(eksaAwsConfig),
			curatedpackages.WithHTTPProxy(httpProxy),
			curatedpackages.WithHTTPSProxy(httpsProxy),
			curatedpackages.WithNoProxy(noProxy),
			curatedpackages.WithManagementClusterName(managementClusterName),
			curatedpackages.WithValuesFileWriter(writer),
			curatedpackages.WithClusterSpec(spec),
		)
		return nil
	})

	return f
}

func (f *Factory) WithPackageClient() *Factory {
	f.WithKubectl()
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.PackageClient != nil {
			return nil
		}

		f.dependencies.PackageClient = curatedpackages.NewPackageClient(
			f.dependencies.Kubectl,
		)
		return nil
	})
	return f
}

func (f *Factory) WithCuratedPackagesRegistry(registryName, kubeVersion string, version version.Info) *Factory {
	if registryName != "" {
		f.WithHelm(helm.WithInsecure())
	} else {
		f.WithManifestReader()
	}

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.BundleRegistry != nil {
			return nil
		}

		if registryName != "" {
			f.dependencies.BundleRegistry = curatedpackages.NewCustomRegistry(
				f.dependencies.Helm,
				registryName,
			)
		} else {
			f.dependencies.BundleRegistry = curatedpackages.NewDefaultRegistry(
				f.dependencies.ManifestReader,
				kubeVersion,
				version,
			)
		}
		return nil
	})
	return f
}

func (f *Factory) WithDiagnosticBundleFactory() *Factory {
	f.WithWriter().WithTroubleshoot().WithCollectorFactory().WithAnalyzerFactory().WithKubectl()
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.DignosticCollectorFactory != nil {
			return nil
		}

		opts := diagnostics.EksaDiagnosticBundleFactoryOpts{
			AnalyzerFactory:  f.dependencies.AnalyzerFactory,
			Client:           f.dependencies.Troubleshoot,
			CollectorFactory: f.dependencies.CollectorFactory,
			Kubectl:          f.dependencies.Kubectl,
			Writer:           f.dependencies.Writer,
		}

		f.dependencies.DignosticCollectorFactory = diagnostics.NewFactory(opts)
		return nil
	})

	return f
}

func (f *Factory) WithAnalyzerFactory() *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.AnalyzerFactory != nil {
			return nil
		}

		f.dependencies.AnalyzerFactory = diagnostics.NewAnalyzerFactory()
		return nil
	})

	return f
}

func (f *Factory) WithDiagnosticCollectorImage(diagnosticCollectorImage string) *Factory {
	f.diagnosticCollectorImage = diagnosticCollectorImage
	return f
}

func (f *Factory) WithCollectorFactory() *Factory {
	f.WithFileReader()
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.CollectorFactory != nil {
			return nil
		}

		if f.diagnosticCollectorImage == "" {
			f.dependencies.CollectorFactory = diagnostics.NewDefaultCollectorFactory(f.dependencies.FileReader)
		} else {
			f.dependencies.CollectorFactory = diagnostics.NewCollectorFactory(f.diagnosticCollectorImage, f.dependencies.FileReader)
		}
		return nil
	})

	return f
}

func (f *Factory) WithCAPIManager() *Factory {
	f.WithClusterctl()
	f.WithKubectl()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.CAPIManager != nil {
			return nil
		}

		f.dependencies.CAPIManager = clusterapi.NewManager(f.dependencies.Clusterctl, f.dependencies.Kubectl)
		return nil
	})

	return f
}

func (f *Factory) WithFileReader() *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.FileReader != nil {
			return nil
		}

		f.dependencies.FileReader = files.NewReader(files.WithEKSAUserAgent("cli", version.Get().GitVersion))
		return nil
	})

	return f
}

func (f *Factory) WithManifestReader() *Factory {
	f.WithFileReader()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.ManifestReader != nil {
			return nil
		}

		f.dependencies.ManifestReader = manifests.NewReader(f.dependencies.FileReader)
		return nil
	})

	return f
}

func (f *Factory) WithUnAuthKubeClient() *Factory {
	f.WithKubectl()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.UnAuthKubeClient != nil {
			return nil
		}

		f.dependencies.UnAuthKubeClient = kubernetes.NewUnAuthClient(f.dependencies.Kubectl)
		if err := f.dependencies.UnAuthKubeClient.Init(); err != nil {
			return fmt.Errorf("building unauth kube client: %v", err)
		}

		return nil
	})

	return f
}

func (f *Factory) WithVSphereValidator() *Factory {
	f.WithGovc()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.VSphereValidator != nil {
			return nil
		}
		vcb := govmomi.NewVMOMIClientBuilder()
		v := vsphere.NewValidator(
			f.dependencies.Govc,
			vcb,
		)
		f.dependencies.VSphereValidator = v

		return nil
	})

	return f
}

func (f *Factory) WithVSphereDefaulter() *Factory {
	f.WithGovc()

	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.VSphereDefaulter != nil {
			return nil
		}

		f.dependencies.VSphereDefaulter = vsphere.NewDefaulter(f.dependencies.Govc)

		return nil
	})

	return f
}

// WithNutanixDefaulter adds a new NutanixDefaulter to the factory.
func (f *Factory) WithNutanixDefaulter() *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.NutanixDefaulter != nil {
			return nil
		}

		f.dependencies.NutanixDefaulter = nutanix.NewDefaulter()

		return nil
	})

	return f
}

// WithNutanixValidator adds a new NutanixValidator to the factory.
func (f *Factory) WithNutanixValidator() *Factory {
	f.WithNutanixClientCache()
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.NutanixValidator != nil {
			return nil
		}
		skipVerifyTransport := http.DefaultTransport.(*http.Transport).Clone()
		skipVerifyTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		v := nutanix.NewValidator(
			f.dependencies.NutanixClientCache,
			crypto.NewTlsValidator(),
			&http.Client{Transport: skipVerifyTransport},
		)
		f.dependencies.NutanixValidator = v

		return nil
	})

	return f
}

// WithNutanixClientCache adds a new NutanixClientCache to the factory.
func (f *Factory) WithNutanixClientCache() *Factory {
	f.buildSteps = append(f.buildSteps, func(ctx context.Context) error {
		if f.dependencies.NutanixClientCache != nil {
			return nil
		}

		f.dependencies.NutanixClientCache = nutanix.NewClientCache()

		return nil
	})

	return f
}

func getProxyConfiguration(clusterSpec *cluster.Spec) (httpProxy, httpsProxy string, noProxy []string) {
	proxyConfiguration := clusterSpec.Cluster.Spec.ProxyConfiguration
	if proxyConfiguration != nil {
		return proxyConfiguration.HttpProxy, proxyConfiguration.HttpsProxy, proxyConfiguration.NoProxy
	}
	return "", "", nil
}

func getManagementClusterName(clusterSpec *cluster.Spec) string {
	if clusterSpec.Cluster.Spec.ManagementCluster.Name != "" {
		return clusterSpec.Cluster.Spec.ManagementCluster.Name
	}
	return clusterSpec.Cluster.Name
}
