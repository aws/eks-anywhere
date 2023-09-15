package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/createcluster"
	"github.com/aws/eks-anywhere/pkg/validations/createvalidations"
	"github.com/aws/eks-anywhere/pkg/version"
)

type validateOptions struct {
	clusterOptions
	tinkerbell tinkerbell.Config
}

var valOpt = &validateOptions{}

var validateCreateClusterCmd = &cobra.Command{
	Use:          "cluster -f <cluster-config-file> [flags]",
	Short:        "Validate create cluster",
	Long:         "Use eksctl anywhere validate create cluster to validate the create cluster action",
	PreRunE:      bindFlagsToViper,
	SilenceUsage: true,
	RunE:         valOpt.validateCreateCluster,
}

func init() {
	validateCreateCmd.AddCommand(validateCreateClusterCmd)
	applyTinkerbellHardwareFlag(validateCreateClusterCmd.Flags(), &valOpt.tinkerbell.HardwareFile)
	validateCreateClusterCmd.Flags().StringVarP(&valOpt.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	validateCreateClusterCmd.Flags().StringVar(&valOpt.tinkerbell.IP, "tinkerbell-bootstrap-ip", "", "Override the local tinkerbell IP in the bootstrap cluster")

	if err := validateCreateClusterCmd.MarkFlagRequired("filename"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}

	validateCreateCmd.Flags().StringVar(&cc.tinkerbellConfig.Rufio.WebhookSecret, "webhook-secrets", "", "Comma separated list of secrets for use with the bare metal webhook provider")
	markFlagHidden(validateCreateCmd.Flags(), "webhook-secrets")
	validateCreateCmd.Flags().StringVar(&cc.tinkerbellConfig.Rufio.ConsumerURL, "consumer-url", "", "URL for the bare metal webhook consumer")
	markFlagHidden(validateCreateCmd.Flags(), "webhook-url")
}

func (valOpt *validateOptions) validateCreateCluster(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	clusterSpec, err := readClusterSpec(valOpt.fileName, version.Get())
	if err != nil {
		return err
	}

	if clusterSpec.Config.Cluster.Spec.DatacenterRef.Kind == v1alpha1.TinkerbellDatacenterKind {
		if err := checkTinkerbellFlags(cmd.Flags(), valOpt.tinkerbell.HardwareFile, 0); err != nil {
			return err
		}
	}

	cliConfig := buildCliConfig(clusterSpec)
	dirs, err := valOpt.directoriesToMount(clusterSpec, cliConfig)
	if err != nil {
		return err
	}

	tmpPath, err := os.MkdirTemp("./", "tmpValidate")
	if err != nil {
		return err
	}
	deps, err := dependencies.ForSpec(ctx, clusterSpec).
		WithExecutableMountDirs(dirs...).
		WithWriterFolder(tmpPath).
		WithDocker().
		WithKubectl().
		WithTinkerbellConfig(valOpt.tinkerbell).
		WithProvider(valOpt.fileName, clusterSpec.Cluster, false, true, map[string]bool{}).
		WithGitOpsFlux(clusterSpec.Cluster, clusterSpec.FluxConfig, cliConfig).
		WithUnAuthKubeClient().
		WithValidatorClients().
		Build(ctx)
	if err != nil {
		cleanupDirectory(tmpPath)
		return err
	}
	defer close(ctx, deps)

	validationOpts := &validations.Opts{
		Kubectl: deps.UnAuthKubectlClient,
		Spec:    clusterSpec,
		WorkloadCluster: &types.Cluster{
			Name:           clusterSpec.Cluster.Name,
			KubeconfigFile: kubeconfig.FromClusterName(clusterSpec.Cluster.Name),
		},
		ManagementCluster: getManagementCluster(clusterSpec),
		Provider:          deps.Provider,
		CliConfig:         cliConfig,
	}

	createValidations := createvalidations.New(validationOpts)

	commandVal := createcluster.NewValidations(clusterSpec, deps.Provider, deps.GitOpsFlux, createValidations, deps.DockerClient)
	err = commandVal.Validate(ctx)

	cleanupDirectory(tmpPath)
	return err
}
