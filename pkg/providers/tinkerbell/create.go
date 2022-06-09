package tinkerbell

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
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

	localIP, err := networkutils.GetLocalIP()
	if err != nil {
		return err
	}

	err = p.stackInstaller.Install(
		ctx,
		clusterSpec.VersionsBundle.Tinkerbell.TinkerbellStack,
		localIP.String(),
		cluster.KubeconfigFile,
		stack.WithNamespaceCreate(false),
		stack.WithBootsOnDocker(),
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
		clusterSpec.VersionsBundle.Tinkerbell.TinkerbellStack,
		p.templateBuilder.datacenterSpec.TinkerbellIP,
		cluster.KubeconfigFile,
		stack.WithNamespaceCreate(true),
		stack.WithBootsOnKubernetes(),
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
		NewCreateMinimumHardwareAvailableAssertion(p.catalogue),
	)

	if !p.skipIpCheck {
		clusterSpecValidator.Register(NewIPNotInUseAssertion(p.netClient))
	}

	// Validate must happen last beacuse we depend on the catalogue entries for some checks.
	if err := clusterSpecValidator.Validate(spec); err != nil {
		return err
	}

	return nil
}

func (p *Provider) readCSVToCatalogue() error {
	catalogueWriter := hardware.NewMachineCatalogueWriter(p.catalogue)

	writer := hardware.MultiMachineWriter(catalogueWriter, &p.diskExtractor)

	machineValidator := hardware.NewDefaultMachineValidator()

	// TODO(chrisdoherty4) Build the selectors slice using the selectors from TinkerbellMachineConfig's
	var selectors []v1alpha1.HardwareSelector
	machineValidator.Register(hardware.MatchingDisksForSelectors(selectors))

	// Translate all Machine instances from the p.machines source into Kubernetes object types.
	// The PostBootstrapSetup() call invoked elsewhere in the program serializes the catalogue
	// and submits it to the clsuter.
	machines, err := hardware.NewCSVReaderFromFile(p.hardwareCSVFile)
	if err != nil {
		return err
	}

	return hardware.TranslateAll(machines, writer, machineValidator)
}
