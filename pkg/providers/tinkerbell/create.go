package tinkerbell

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack"
	"github.com/aws/eks-anywhere/pkg/types"
)

func (p *Provider) BootstrapClusterOpts() ([]bootstrapper.BootstrapClusterOption, error) {
	var opts []bootstrapper.BootstrapClusterOption

	if p.clusterConfig.Spec.ProxyConfiguration != nil {
		// +2 for control plane endpoint and tinkerbell IP.
		noProxyAddresses := make([]string, 0, len(p.clusterConfig.Spec.ProxyConfiguration.NoProxy)+2)
		noProxyAddresses = append(
			noProxyAddresses,
			p.clusterConfig.Spec.ControlPlaneConfiguration.Endpoint.Host,
			p.datacenterConfig.Spec.TinkerbellIP,
		)
		noProxyAddresses = append(noProxyAddresses, p.clusterConfig.Spec.ProxyConfiguration.NoProxy...)

		env := make(map[string]string, 3)
		env["HTTP_PROXY"] = p.clusterConfig.Spec.ProxyConfiguration.HttpProxy
		env["HTTPS_PROXY"] = p.clusterConfig.Spec.ProxyConfiguration.HttpsProxy
		env["NO_PROXY"] = strings.Join(noProxyAddresses, ",")

		opts = append(opts, bootstrapper.WithEnv(env))
	}

	opts = append(opts, bootstrapper.WithExtraPortMappings(tinkerbellStackPorts))

	return opts, nil
}

func (p *Provider) PreCAPIInstallOnBootstrap(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	logger.V(4).Info("Installing Tinkerbell stack on bootstrap cluster")

	err := p.stackInstaller.Install(
		ctx,
		clusterSpec.VersionsBundle.Tinkerbell,
		p.tinkerbellIp,
		cluster.KubeconfigFile,
		p.datacenterConfig.Spec.HookImagesURLPath,
		stack.WithNamespaceCreate(false),
		stack.WithBootsOnDocker(),
		stack.WithHostPortEnabled(true),
	)
	if err != nil {
		return fmt.Errorf("install Tinkerbell stack on bootstrap cluster: %v", err)
	}

	return nil
}

func (p *Provider) PostBootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	hardwareSpec, err := hardware.MarshalCatalogue(p.catalogue)
	if err != nil {
		return fmt.Errorf("failed marshalling resources for hardware spec: %v", err)
	}
	err = p.providerKubectlClient.ApplyKubeSpecFromBytesForce(ctx, cluster, hardwareSpec)
	if err != nil {
		return fmt.Errorf("applying hardware yaml: %v", err)
	}
	return nil
}

func (p *Provider) PostWorkloadInit(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	logger.V(4).Info("Installing Tinkerbell stack on workload cluster")

	err := p.stackInstaller.Install(
		ctx,
		clusterSpec.VersionsBundle.Tinkerbell,
		p.templateBuilder.datacenterSpec.TinkerbellIP,
		cluster.KubeconfigFile,
		p.datacenterConfig.Spec.HookImagesURLPath,
		stack.WithNamespaceCreate(true),
		stack.WithBootsOnKubernetes(),
		stack.WithHostPortEnabled(false),
		stack.WithLoadBalancer(),
	)
	if err != nil {
		return fmt.Errorf("installing stack on workload cluster: %v", err)
	}

	if err := p.stackInstaller.UninstallLocal(ctx); err != nil {
		return err
	}

	return nil
}

func (p *Provider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		return ErrExternalEtcdUnsupported
	}

	if err := p.stackInstaller.CleanupLocalBoots(ctx, p.forceCleanup); err != nil {
		return err
	}

	// TODO(chrisdoherty4) Extract to a defaulting construct and add associated validations to ensure
	// there is always a user with ssh key configured.
	if err := p.configureSshKeys(); err != nil {
		return err
	}

	if err := p.readCSVToCatalogue(); err != nil {
		return err
	}

	spec := NewClusterSpec(clusterSpec, p.machineConfigs, p.datacenterConfig)

	// TODO(chrisdoherty4) Look to inject the validator. Possibly look to use a builder for
	// constructing the validations rather than injecting flags into the provider.
	clusterSpecValidator := NewClusterSpecValidator(
		MinimumHardwareAvailableAssertionForCreate(p.catalogue),
		HardwareSatisfiesOnlyOneSelectorAssertion(p.catalogue),
	)

	if !p.skipIpCheck {
		clusterSpecValidator.Register(NewIPNotInUseAssertion(p.netClient))
	}

	if p.datacenterConfig.Spec.OSImageURL != "" {
		if _, err := url.ParseRequestURI(p.datacenterConfig.Spec.OSImageURL); err != nil {
			return fmt.Errorf("parsing osImageOverride: %v", err)
		}
	}

	if p.datacenterConfig.Spec.HookImagesURLPath != "" {
		if _, err := url.ParseRequestURI(p.datacenterConfig.Spec.HookImagesURLPath); err != nil {
			return fmt.Errorf("parsing hookOverride: %v", err)
		}
		logger.Info("hook path override set", "path", p.datacenterConfig.Spec.HookImagesURLPath)
	}

	// Validate must happen last beacuse we depend on the catalogue entries for some checks.
	if err := clusterSpecValidator.Validate(spec); err != nil {
		return err
	}

	return nil
}

func (p *Provider) readCSVToCatalogue() error {
	// Create a catalogue writer used to write hardware to the catalogue.
	catalogueWriter := hardware.NewMachineCatalogueWriter(p.catalogue)

	// Combine disk extraction with catalogue writing. Disk extraction will be used for rendering
	// templates.
	writer := hardware.MultiMachineWriter(catalogueWriter, &p.diskExtractor)

	machineValidator := hardware.NewDefaultMachineValidator()

	// Build a set of selectors from machine configs.
	selectors := selectorsFromMachineConfigs(p.machineConfigs)
	machineValidator.Register(hardware.MatchingDisksForSelectors(selectors))

	// Translate all Machine instances from the p.machines source into Kubernetes object types.
	// The PostBootstrapSetup() call invoked elsewhere in the program serializes the catalogue
	// and submits it to the clsuter.
	machines, err := hardware.NewNormalizedCSVReaderFromFile(p.hardwareCSVFile)
	if err != nil {
		return err
	}

	return hardware.TranslateAll(machines, writer, machineValidator)
}

// selectorsFromMachineConfigs extracts all selectors from TinkerbellMachineConfigs returning them
// as a slice. It doesn't need the map, it only accepts that for ease as that's how we manage them
// in the provider construct.
func selectorsFromMachineConfigs(configs map[string]*v1alpha1.TinkerbellMachineConfig) []v1alpha1.HardwareSelector {
	selectors := make([]v1alpha1.HardwareSelector, 0, len(configs))
	for _, s := range configs {
		selectors = append(selectors, s.Spec.HardwareSelector)
	}
	return selectors
}
