package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/upgradevalidations"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type upgradeClusterOptions struct {
	clusterOptions
	timeoutOptions
	wConfig               string
	forceClean            bool
	hardwareCSVPath       string
	tinkerbellBootstrapIP string
	skipValidations       []string
}

var uc = &upgradeClusterOptions{}

var upgradeClusterCmd = &cobra.Command{
	Use:          "cluster",
	Short:        "Upgrade workload cluster",
	Long:         "This command is used to upgrade workload clusters",
	PreRunE:      bindFlagsToViper,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := uc.upgradeCluster(cmd); err != nil {
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
	upgradeClusterCmd.Flags().StringArrayVar(&uc.skipValidations, "skip-validations", []string{}, fmt.Sprintf("Bypass upgrade validations by name. Valid arguments you can pass are --skip-validations=%s", strings.Join(upgradevalidations.SkippableValidations[:], ",")))

	if err := upgradeClusterCmd.MarkFlagRequired("filename"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func (uc *upgradeClusterOptions) upgradeCluster(cmd *cobra.Command) error {
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

	if _, err := uc.commonValidations(ctx); err != nil {
		return fmt.Errorf("common validations failed due to: %v", err)
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

	factory := dependencies.ForSpec(ctx, clusterSpec).WithExecutableMountDirs(dirs...).
		WithBootstrapper().
		WithCliConfig(cliConfig).
		WithClusterManager(clusterSpec.Cluster, clusterManagerTimeoutOpts).
		WithProvider(uc.fileName, clusterSpec.Cluster, cc.skipIpCheck, uc.hardwareCSVPath, uc.forceClean, uc.tinkerbellBootstrapIP, skippedValidations).
		WithGitOpsFlux(clusterSpec.Cluster, clusterSpec.FluxConfig, cliConfig).
		WithWriter().
		WithCAPIManager().
		WithEksdUpgrader().
		WithEksdInstaller().
		WithKubectl().
		WithValidatorClients()

	if uc.timeoutOptions.noTimeouts {
		factory.WithNoTimeouts()
	}

	deps, err := factory.Build(ctx)
	if err != nil {
		return err
	}
	defer close(ctx, deps)

	upgradeCluster := workflows.NewUpgrade(
		deps.Bootstrapper,
		deps.Provider,
		deps.CAPIManager,
		deps.ClusterManager,
		deps.GitOpsFlux,
		deps.Writer,
		deps.EksdUpgrader,
		deps.EksdInstaller,
	)

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
	}

	upgradeValidations := upgradevalidations.New(validationOpts)

	err = upgradeCluster.Run(ctx, clusterSpec, managementCluster, workloadCluster, upgradeValidations, uc.forceClean)
	cleanup(deps, &err)
	return err
}

func (uc *upgradeClusterOptions) commonValidations(ctx context.Context) (cluster *v1alpha1.Cluster, err error) {
	clusterConfig, err := commonValidation(ctx, uc.fileName)
	if err != nil {
		return nil, err
	}

	kubeconfigPath := getKubeconfigPath(clusterConfig.Name, uc.wConfig)
	if err := kubeconfig.ValidateFilename(kubeconfigPath); err != nil {
		return nil, err
	}

	return clusterConfig, nil
}
