package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/aflag"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/upgradevalidations"
	"github.com/aws/eks-anywhere/pkg/workflows/management"
	"github.com/aws/eks-anywhere/pkg/workflows/workload"
)

type upgradeClusterOptions struct {
	clusterOptions
	timeoutOptions
	wConfig               string
	forceClean            bool
	hardwareCSVPath       string
	tinkerbellBootstrapIP string
	skipValidations       []string
	providerOptions       *dependencies.ProviderOptions
}

var uc = &upgradeClusterOptions{
	providerOptions: &dependencies.ProviderOptions{
		Tinkerbell: &dependencies.TinkerbellOptions{
			BMCOptions: &hardware.BMCOptions{
				RPC: &hardware.RPCOpts{},
			},
		},
	},
}

var upgradeClusterCmd = &cobra.Command{
	Use:          "cluster",
	Short:        "Upgrade workload cluster",
	Long:         "This command is used to upgrade workload clusters",
	PreRunE:      bindFlagsToViper,
	SilenceUsage: true,
	Args:         cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if uc.forceClean {
			logger.MarkFail(forceCleanupDeprecationMessageForUpgrade)
			return errors.New("please remove the --force-cleanup flag")
		}

		if err := uc.upgradeCluster(cmd, args); err != nil {
			return fmt.Errorf("failed to upgrade cluster: %v", err)
		}
		return nil
	},
}

func init() {
	upgradeCmd.AddCommand(upgradeClusterCmd)
	applyClusterOptionFlags(upgradeClusterCmd.Flags(), &uc.clusterOptions)
	applyTimeoutFlags(upgradeClusterCmd.Flags(), &uc.timeoutOptions)
	applyTinkerbellHardwareFlag(upgradeClusterCmd.Flags(), &uc.hardwareCSVPath)
	upgradeClusterCmd.Flags().StringVarP(&uc.wConfig, "w-config", "w", "", "Kubeconfig file to use when upgrading a workload cluster")
	upgradeClusterCmd.Flags().BoolVar(&uc.forceClean, "force-cleanup", false, "Force deletion of previously created bootstrap cluster")
	hideForceCleanup(upgradeClusterCmd.Flags())
	upgradeClusterCmd.Flags().StringArrayVar(&uc.skipValidations, "skip-validations", []string{}, fmt.Sprintf("Bypass upgrade validations by name. Valid arguments you can pass are --skip-validations=%s", strings.Join(upgradevalidations.SkippableValidations[:], ",")))

	aflag.MarkRequired(createClusterCmd.Flags(), aflag.ClusterConfig.Name)
	tinkerbellFlags(upgradeClusterCmd.Flags(), uc.providerOptions.Tinkerbell.BMCOptions.RPC)
}

// nolint:gocyclo
func (uc *upgradeClusterOptions) upgradeCluster(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	clusterConfigFileExist := validations.FileExists(uc.fileName)
	if !clusterConfigFileExist {
		return fmt.Errorf("the cluster config file %s does not exist", uc.fileName)
	}

	clusterConfig, err := v1alpha1.GetAndValidateClusterConfig(uc.fileName)
	if err != nil {
		return fmt.Errorf("the cluster config file provided is invalid: %v", err)
	}

	if clusterConfig.Spec.DatacenterRef.Kind == v1alpha1.TinkerbellDatacenterKind {
		if err := checkTinkerbellFlags(cmd.Flags(), uc.hardwareCSVPath, Upgrade); err != nil {
			return err
		}
	}

	if clusterConfig.Spec.EtcdEncryption != nil && clusterConfig.Spec.DatacenterRef.Kind != v1alpha1.CloudStackDatacenterKind && clusterConfig.Spec.DatacenterRef.Kind != v1alpha1.VSphereDatacenterKind {
		return fmt.Errorf("etcdEncryption is currently not supported for the current provider: %s", clusterConfig.Spec.DatacenterRef.Kind)
	}

	if err := v1alpha1.ValidateEtcdEncryptionConfig(clusterConfig.Spec.EtcdEncryption); err != nil {
		return err
	}

	if _, err := uc.commonValidations(ctx); err != nil {
		return fmt.Errorf("common validations failed due to: %v", err)
	}

	if err := validations.ValidateClusterNameFromCommandAndConfig(args, clusterConfig.Name); err != nil {
		return err
	}
	clusterSpec, err := newClusterSpec(uc.clusterOptions)
	if err != nil {
		return err
	}

	if err := validations.ValidateAuthenticationForRegistryMirror(clusterSpec); err != nil {
		return err
	}

	cliConfig := buildCliConfig(clusterSpec)
	dirs, err := uc.directoriesToMount(clusterSpec, cliConfig)
	if err != nil {
		return err
	}

	upgradeCLIConfig, err := buildUpgradeCliConfig(uc)
	if err != nil {
		return err
	}

	clusterManagerTimeoutOpts, err := buildClusterManagerOpts(uc.timeoutOptions, clusterSpec.Cluster.Spec.DatacenterRef.Kind)
	if err != nil {
		return fmt.Errorf("failed to build cluster manager opts: %v", err)
	}

	var skippedValidations map[string]bool
	if len(uc.skipValidations) != 0 {
		skippedValidations, err = validations.ValidateSkippableValidation(uc.skipValidations, upgradevalidations.SkippableValidations)
		if err != nil {
			return err
		}
	}

	factory := dependencies.ForSpec(clusterSpec).WithExecutableMountDirs(dirs...).
		WithBootstrapper().
		WithCliConfig(cliConfig).
		WithClusterManager(clusterSpec.Cluster, clusterManagerTimeoutOpts).
		WithClusterApplier().
		WithProvider(uc.fileName, clusterSpec.Cluster, cc.skipIpCheck, uc.hardwareCSVPath, uc.forceClean, uc.tinkerbellBootstrapIP, skippedValidations, uc.providerOptions).
		WithGitOpsFlux(clusterSpec.Cluster, clusterSpec.FluxConfig, cliConfig).
		WithWriter().
		WithCAPIManager().
		WithEksdUpgrader().
		WithEksdInstaller().
		WithKubectl().
		WithValidatorClients().
		WithUpgradeClusterDefaulter(upgradeCLIConfig)

	if uc.timeoutOptions.noTimeouts {
		factory.WithNoTimeouts()
	}

	deps, err := factory.Build(ctx)
	if err != nil {
		return err
	}
	defer close(ctx, deps)

	clusterSpec, err = deps.UpgradeClusterDefaulter.Run(ctx, clusterSpec)
	if err != nil {
		return err
	}

	workloadCluster := &types.Cluster{
		Name:           clusterSpec.Cluster.Name,
		KubeconfigFile: getKubeconfigPath(clusterSpec.Cluster.Name, uc.wConfig),
	}

	var managementCluster *types.Cluster
	if clusterSpec.ManagementCluster == nil {
		managementCluster = workloadCluster
	} else {
		managementCluster = clusterSpec.ManagementCluster
	}

	validationOpts := &validations.Opts{
		Kubectl:            deps.UnAuthKubectlClient,
		Spec:               clusterSpec,
		WorkloadCluster:    workloadCluster,
		ManagementCluster:  managementCluster,
		Provider:           deps.Provider,
		CliConfig:          cliConfig,
		SkippedValidations: skippedValidations,
		KubeClient:         deps.UnAuthKubeClient.KubeconfigClient(managementCluster.KubeconfigFile),
	}

	upgradeValidations := upgradevalidations.New(validationOpts)

	if clusterConfig.IsSelfManaged() {
		upgrade := management.NewUpgrade(
			deps.UnAuthKubeClient,
			deps.Provider,
			deps.CAPIManager,
			deps.ClusterManager,
			deps.GitOpsFlux,
			deps.Writer,
			deps.EksdUpgrader,
			deps.EksdInstaller,
			deps.ClusterApplier,
		)

		err = upgrade.Run(ctx, clusterSpec, managementCluster, upgradeValidations)

	} else {
		upgradeWorkloadCluster := workload.NewUpgrade(
			deps.Provider,
			deps.ClusterManager,
			deps.GitOpsFlux,
			deps.Writer,
			deps.ClusterApplier,
			deps.EksdInstaller,
			deps.PackageInstaller,
		)
		err = upgradeWorkloadCluster.Run(ctx, workloadCluster, clusterSpec, upgradeValidations)
	}

	cleanup(deps, &err)
	return err
}

func (uc *upgradeClusterOptions) commonValidations(ctx context.Context) (cluster *v1alpha1.Cluster, err error) {
	clusterConfig, err := commonValidation(ctx, uc.fileName)
	if err != nil {
		return nil, err
	}

	if uc.wConfig == "" && uc.managementKubeconfig != "" && clusterConfig.IsSelfManaged() {
		uc.wConfig = uc.managementKubeconfig
		uc.managementKubeconfig = ""
	}

	kubeconfigPath := getKubeconfigPath(clusterConfig.Name, uc.wConfig)
	if err := kubeconfig.ValidateFilename(kubeconfigPath); err != nil {
		return nil, err
	}

	return clusterConfig, nil
}
