package tinkerbell

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack"
	"github.com/aws/eks-anywhere/pkg/types"
)

func (p *Provider) BootstrapClusterOpts(_ *cluster.Spec) ([]bootstrapper.BootstrapClusterOption, error) {
	opts, err := common.BootstrapClusterOpts(p.clusterConfig, p.datacenterConfig.Spec.TinkerbellIP)
	if err != nil {
		return nil, err
	}
	opts = append(opts, bootstrapper.WithExtraPortMappings(tinkerbellStackPorts))

	return opts, nil
}

func (p *Provider) PreCAPIInstallOnBootstrap(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	logger.V(4).Info("Installing Tinkerbell stack on bootstrap cluster")

	versionsBundle := clusterSpec.RootVersionsBundle()

	err := p.stackInstaller.Install(
		ctx,
		versionsBundle.Tinkerbell,
		p.tinkerbellIP,
		cluster.KubeconfigFile,
		p.datacenterConfig.Spec.HookImagesURLPath,
		stack.WithBootsOnDocker(),
		stack.WithHostPortEnabled(true), // enable host port on bootstrap cluster
	)
	if err != nil {
		return fmt.Errorf("install Tinkerbell stack on bootstrap cluster: %v", err)
	}

	return nil
}

func (p *Provider) PostBootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	return p.applyHardware(ctx, cluster)
}

// ApplyHardwareToCluster adds all the hardwares to the cluster.
func (p *Provider) applyHardware(ctx context.Context, cluster *types.Cluster) error {
	hardwareSpec, err := hardware.MarshalCatalogue(p.catalogue)
	if err != nil {
		return fmt.Errorf("failed marshalling resources for hardware spec: %v", err)
	}
	err = p.providerKubectlClient.ApplyKubeSpecFromBytesForce(ctx, cluster, hardwareSpec)
	if err != nil {
		return fmt.Errorf("applying hardware yaml: %v", err)
	}
	if len(p.catalogue.AllBMCs()) > 0 {
		err = p.providerKubectlClient.WaitForRufioMachines(ctx, cluster, "5m", "Contactable", constants.EksaSystemNamespace)
		if err != nil {
			return fmt.Errorf("waiting for baseboard management to be contactable: %v", err)
		}
	}
	return nil
}

func (p *Provider) PostWorkloadInit(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	logger.V(4).Info("Installing Tinkerbell stack on workload cluster")

	if p.datacenterConfig.Spec.SkipLoadBalancerDeployment {
		logger.Info("Warning: Skipping load balancer deployment. Please install and configure a load balancer once the cluster is created.")
	}

	versionsBundle := clusterSpec.RootVersionsBundle()

	err := p.stackInstaller.Install(
		ctx,
		versionsBundle.Tinkerbell,
		p.templateBuilder.datacenterSpec.TinkerbellIP,
		cluster.KubeconfigFile,
		p.datacenterConfig.Spec.HookImagesURLPath,
		stack.WithBootsOnKubernetes(),
		stack.WithHostPortEnabled(false), // disable host port on workload cluster
		stack.WithEnvoyEnabled(true),     // use envoy on workload cluster
		stack.WithLoadBalancerEnabled(
			len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations) != 0 && // load balancer is handled by kube-vip in control plane nodes
				!p.datacenterConfig.Spec.SkipLoadBalancerDeployment), // configure load balancer based on datacenterConfig.Spec.SkipLoadBalancerDeployment
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
		return errExternalEtcdUnsupported
	}

	if err := p.stackInstaller.CleanupLocalBoots(ctx, p.forceCleanup); err != nil {
		return err
	}

	// TODO(chrisdoherty4) Extract to a defaulting construct and add associated validations to ensure
	// there is always a user with ssh key configured.
	if err := p.configureSshKeys(); err != nil {
		return err
	}
	clusterSpec.TinkerbellMachineConfigs = p.machineConfigs
	if p.hardwareCSVIsProvided() {
		if err := p.readCSVToCatalogue(); err != nil {
			return err
		}
	}

	spec := NewClusterSpec(clusterSpec, p.machineConfigs, p.datacenterConfig)

	if p.clusterConfig.IsManaged() {
		for _, mc := range p.MachineConfigs(clusterSpec) {
			em, err := p.providerKubectlClient.SearchTinkerbellMachineConfig(ctx, mc.GetName(), clusterSpec.ManagementCluster.KubeconfigFile, mc.GetNamespace())
			if err != nil {
				return err
			}
			if len(em) > 0 {
				return fmt.Errorf("TinkerbellMachineConfig %s already exists", mc.GetName())
			}
		}
		existingDatacenter, err := p.providerKubectlClient.SearchTinkerbellDatacenterConfig(ctx, p.datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, clusterSpec.Cluster.Namespace)
		if err != nil {
			return err
		}
		if len(existingDatacenter) > 0 {
			return fmt.Errorf("TinkerbellDatacenterConfig %s already exists", p.datacenterConfig.Name)
		}

		if err := p.getHardwareFromManagementCluster(ctx, clusterSpec.ManagementCluster); err != nil {
			return err
		}

		// for workload cluster use tinkerbell IP of the management cluster
		managementClusterSpec, err := p.providerKubectlClient.GetEksaCluster(ctx, clusterSpec.ManagementCluster, clusterSpec.ManagementCluster.Name)
		if err != nil {
			return err
		}

		managementDatacenterConfig, err := p.providerKubectlClient.GetEksaTinkerbellDatacenterConfig(ctx, managementClusterSpec.Spec.DatacenterRef.Name, clusterSpec.ManagementCluster.KubeconfigFile, managementClusterSpec.Namespace)
		if err != nil {
			return fmt.Errorf("getting TinkerbellIP of management cluster: %s", err)
		}
		// Checking for empty first as that returns a different error in the datacenter config validate method below
		if p.datacenterConfig.Spec.TinkerbellIP != "" && p.datacenterConfig.Spec.TinkerbellIP != managementDatacenterConfig.Spec.TinkerbellIP {
			return fmt.Errorf("TinkerbellIP %v does not match management cluster ip %v", p.datacenterConfig.Spec.TinkerbellIP, managementDatacenterConfig.Spec.TinkerbellIP)
		}
	}
	// TODO(chrisdoherty4) Look to inject the validator. Possibly look to use a builder for
	// constructing the validations rather than injecting flags into the provider.
	clusterSpecValidator := NewClusterSpecValidator(
		MinimumHardwareAvailableAssertionForCreate(p.catalogue),
		HardwareSatisfiesOnlyOneSelectorAssertion(p.catalogue),
	)

	clusterSpecValidator.Register(AssertPortsNotInUse(p.netClient))

	if !p.skipIpCheck {
		clusterSpecValidator.Register(NewIPNotInUseAssertion(p.netClient))
		if !p.clusterConfig.IsManaged() {
			clusterSpecValidator.Register(AssertTinkerbellIPNotInUse(p.netClient))
		}
	}
	// Validate must happen last beacuse we depend on the catalogue entries for some checks.
	if err := clusterSpecValidator.Validate(spec); err != nil {
		return err
	}

	if p.clusterConfig.IsManaged() {
		return p.applyHardware(ctx, clusterSpec.ManagementCluster)
	}

	return nil
}

func (p *Provider) getHardwareFromManagementCluster(ctx context.Context, cluster *types.Cluster) error {
	// Retrieve all unprovisioned hardware from the management cluster and populate the catalogue so
	// it can be considered for the workload creation.
	hardware, err := p.providerKubectlClient.GetUnprovisionedTinkerbellHardware(
		ctx,
		cluster.KubeconfigFile,
		constants.EksaSystemNamespace,
	)
	if err != nil {
		return fmt.Errorf("retrieving unprovisioned hardware: %v", err)
	}
	for i := range hardware {
		if err := p.catalogue.InsertHardware(&hardware[i]); err != nil {
			return err
		}
	}

	// Retrieve all provisioned hardware from the management cluster and populate diskExtractors's
	// disksProvisionedHardware map for use during workload creation
	hardware, err = p.providerKubectlClient.GetProvisionedTinkerbellHardware(
		ctx,
		cluster.KubeconfigFile,
		constants.EksaSystemNamespace,
	)
	if err != nil {
		return fmt.Errorf("retrieving provisioned hardware: %v", err)
	}

	// Remove all the provisioned hardware from the existing cluster if repeated from the hardware csv input.
	if err := p.catalogue.RemoveHardwares(hardware); err != nil {
		return err
	}

	return nil
}

func (p *Provider) readCSVToCatalogue() error {
	// Create a catalogue writer used to write hardware to the catalogue.
	catalogueWriter := hardware.NewMachineCatalogueWriter(p.catalogue)

	machineValidator := hardware.NewDefaultMachineValidator()

	// Translate all Machine instances from the p.machines source into Kubernetes object types.
	// The PostBootstrapSetup() call invoked elsewhere in the program serializes the catalogue
	// and submits it to the clsuter.
	machines, err := hardware.NewNormalizedCSVReaderFromFile(p.hardwareCSVFile, p.BMCOptions)
	if err != nil {
		return err
	}

	return hardware.TranslateAll(machines, catalogueWriter, machineValidator)
}
