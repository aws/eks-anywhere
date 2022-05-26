package tinkerbell

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
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

	if p.setupTinkerbell {
		opts = append(opts, bootstrapper.WithExtraPortMappings(tinkerbellStackPorts))
	}

	return opts, nil
}

func (p *Provider) PreCAPIInstallOnBootstrap(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	if p.setupTinkerbell {
		logger.V(4).Info("Installing Tinkerbell stack on the bootstrap cluster")
		if err := p.InstallTinkerbellStack(ctx, cluster, clusterSpec, true); err != nil {
			return fmt.Errorf("installing tinkerbell stack on the bootstrap cluster: %v", err)
		}
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
	if p.setupTinkerbell {

		logger.V(4).Info("Installing Tinkerbell stack on the workload cluster")
		if err := p.InstallTinkerbellStack(ctx, cluster, clusterSpec, false); err != nil {
			return fmt.Errorf("install tinkerbell stack on workload cluster: %v", err)
		}

		logger.V(4).Info("Removing local boots container")
		if err := p.docker.ForceRemove(ctx, bootsContainerName); err != nil {
			return fmt.Errorf("remove local boots container: %v", err)
		}
	}
	return nil
}

func (p *Provider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	logger.Info("Warning: The tinkerbell infrastructure provider is still in development and should not be used in production")

	// TODO(chrisdoherty4) Extract to a defaulting construct and add associated validations to ensure
	// there is always a user with ssh key configured.
	if err := p.configureSshKeys(); err != nil {
		return err
	}

	spec := NewClusterSpec(clusterSpec, p.machineConfigs, p.datacenterConfig)

	// TODO(chrisdoherty4) Look to inject the validator. Possibly look to use a builder for
	// constructing the validations rather than injecting flags into the provider.
	validator := NewClusterSpecValidator(
		NewMinimumHardwareAvailableAssertion(p.catalogue),
	)

	if !p.skipIpCheck {
		validator.Register(NewIPNotInUseAssertion(p.netClient))
	}

	writer := hardware.NewMachineCatalogueWriter(p.catalogue)
	machineValidator := hardware.NewDefaultMachineValidator()

	// Translate all Machine instances from the p.machines source into Kubernetes object types.
	// The PostBootstrapSetup() call invoked elsewhere in the program serializes the catalogue
	// and submits it to the clsuter.
	if err := hardware.TranslateAll(p.machines, writer, machineValidator); err != nil {
		return err
	}

	// Validate must happen last beacuse we depend on the catalogue entries for some checks.
	if err := validator.Validate(spec); err != nil {
		return err
	}

	return nil
}
