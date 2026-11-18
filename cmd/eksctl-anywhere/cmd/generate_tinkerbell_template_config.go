package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/aflag"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/utils/yaml"
)

const shortGenerateTinkerbellTemplateConfigHelp = "Generate TinkerbellTemplateConfig objects"

const longGenerateTinkerbellTemplateConfigHelp = `Generate TinkerbellTemplateConfig objects for your cluster specification.

The TinkerbellTemplateConfig is part of an EKS Anywhere bare metal cluster 
specification. When no template config is specified on TinkerbellMachineConfig
objects, EKS Anywhere generates the template config internally. The template 
config defines the actions for provisioning a bare metal host such as streaming 
an OS image to disk. Actions vary based on the OS - see the EKS Anywhere 
documentation for more details on the individual actions.

The template config include it in your bare metal cluster specification and
reference it in the TinkerbellMachineConfig object using the .spec.templateRef
field.
`

// NewGenerateTinkerbellTemplateConfig creates a command that will generate a TinkerbellTemplateConfig
// using the cluster configuration.
func NewGenerateTinkerbellTemplateConfig() *cobra.Command {
	var opts struct {
		clusterOptions
		BootstrapTinkerbellIP string
	}

	// Configure the flagset. Some of these flags are duplicated from other parts of the cmd code
	// for consistency but their descriptions may vary because of the commands use-case.
	flgs := pflag.NewFlagSet("", pflag.ContinueOnError)
	aflag.String(aflag.ClusterConfig, &opts.fileName, flgs)
	aflag.String(aflag.BundleOverride, &opts.bundlesOverride, flgs)
	aflag.String(aflag.TinkerbellBootstrapIP, &opts.BootstrapTinkerbellIP, flgs)

	cmd := &cobra.Command{
		Use:     "tinkerbelltemplateconfig",
		Short:   shortGenerateTinkerbellTemplateConfigHelp,
		Long:    longGenerateTinkerbellTemplateConfigHelp,
		Aliases: []string{"ttc"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validation logic called by newClusterSpec arbitrarily logs warnings. Until it can
			// be refactored we need to redirect logging such that the output is swallowed.
			if err := logger.Init(logger.Options{
				Level: -1,
			}); err != nil {
				return err
			}

			cs, err := newClusterSpec(opts.clusterOptions)

			// Reset the logger before evaluating anything in-case logic higher up the call path
			// needs the logger.
			if err := logger.Init(logger.Options{
				Level: viper.GetInt("verbosity"),
			}); err != nil {
				return err
			}

			// Handle the newClusterSpec() error.
			if err != nil {
				return err
			}

			cfg, err := generateTinkerbellTemplateConfig(cs, opts.BootstrapTinkerbellIP, &networkutils.DefaultNetClient{})
			if err != nil {
				return err
			}

			return yaml.NewK8sEncoder(os.Stdout).Encode(cfg)
		},
	}

	// Configure the commands flags.
	cmd.Flags().AddFlagSet(flgs)

	return cmd
}

func generateTinkerbellTemplateConfig(cs *cluster.Spec, bootstrapOverride string, client networkutils.NetClient) (*v1alpha1.TinkerbellTemplateConfig, error) {
	tinkerbellIP := cs.TinkerbellDatacenter.Spec.TinkerbellIP
	bootstrapIP, err := networkutils.ResolveLocalIP(client, tinkerbellIP, bootstrapOverride)
	if err != nil {
		return nil, fmt.Errorf("tinkerbell bootstrap ip: %w", err)
	}

	// Validated clusters use one OS family for their control plane and workers.
	controlPlaneMachineConfigName := cs.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	osFamily := cs.TinkerbellMachineConfigs[controlPlaneMachineConfigName].OSFamily()
	return v1alpha1.NewDefaultTinkerbellTemplateConfigCreate(cs.Cluster,
		cs.TinkerbellDatacenter.Spec.OSImageURL, bootstrapIP.String(), tinkerbellIP, osFamily), nil
}
