package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/aflag"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows/management"
)

type upgradeManagementComponentsOptions struct {
	clusterOptions
}

var umco = &upgradeManagementComponentsOptions{}

func init() {
	flagSet := upgradeManagementComponentsCmd.Flags()
	aflag.String(aflag.ClusterConfig, &umco.fileName, flagSet)
	aflag.String(aflag.BundleOverride, &umco.bundlesOverride, flagSet)
}

var upgradeManagementComponentsCmd = &cobra.Command{
	Use:          "management-components",
	Short:        "Upgrade management components in a management cluster",
	Long:         "The term 'management components' encompasses all Kubernetes controllers and their CRDs present in the management cluster that are responsible for reconciling your EKS Anywhere (EKS-A) cluster. This command is specifically designed to facilitate the upgrade of these management components. Post this upgrade, the cluster itself can be upgraded by updating the 'eksaRelease' field in your eksa cluster object.",
	PreRunE:      bindFlagsToViper,
	SilenceUsage: true,
	Args:         cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		clusterSpec, err := newClusterSpec(clusterOptions{
			fileName: umco.fileName,
		})
		if err != nil {
			return err
		}
		if !clusterSpec.Cluster.IsSelfManaged() {
			return fmt.Errorf("cluster %s doesn't contain management components to be upgraded", clusterSpec.Cluster.Name)
		}

		cliConfig := buildCliConfig(clusterSpec)
		dirs, err := uc.directoriesToMount(clusterSpec, cliConfig)
		if err != nil {
			return err
		}

		factory := dependencies.ForSpec(clusterSpec).WithExecutableMountDirs(dirs...).
			WithBootstrapper().
			WithCliConfig(cliConfig).
			WithClusterManager(clusterSpec.Cluster, nil).
			WithClusterApplier().
			WithProvider(umco.fileName, clusterSpec.Cluster, false, "", false, "", nil, nil).
			WithGitOpsFlux(clusterSpec.Cluster, clusterSpec.FluxConfig, cliConfig).
			WithWriter().
			WithCAPIManager().
			WithEksdUpgrader().
			WithEksdInstaller().
			WithKubectl().
			WithValidatorClients()

		deps, err := factory.Build(ctx)
		if err != nil {
			return err
		}
		defer close(cmd.Context(), deps)

		runner := management.NewUpgradeManagementComponentsRunner(
			deps.UnAuthKubeClient,
			deps.Provider,
			deps.CAPIManager,
			deps.ClusterManager,
			deps.GitOpsFlux,
			deps.Writer,
			deps.EksdUpgrader,
			deps.EksdInstaller,
		)

		managementCluster := &types.Cluster{
			Name:           clusterSpec.Cluster.Name,
			KubeconfigFile: kubeconfig.FromClusterName(clusterSpec.Cluster.Name),
		}

		validator := management.NewUMCValidator(managementCluster, deps.Kubectl)
		return runner.Run(ctx, clusterSpec, managementCluster, validator)
	},
}

func init() {
	upgradeCmd.AddCommand(upgradeManagementComponentsCmd)
}
