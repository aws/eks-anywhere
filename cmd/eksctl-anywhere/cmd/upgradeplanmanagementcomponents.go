package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/cluster"
	capiupgrader "github.com/aws/eks-anywhere/pkg/clusterapi"
	eksaupgrader "github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	fluxupgrader "github.com/aws/eks-anywhere/pkg/gitops/flux"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

var upgradePlanManagementComponentsCmd = &cobra.Command{
	Use:          "management-components",
	Short:        "Lists the current and target versions for upgrading the management components in a management cluster",
	Long:         "Provides a list of current and target versions for upgrading the management components in a management cluster. The term 'management components' encompasses all Kubernetes controllers and their CRDs present in the management cluster that are responsible for reconciling your EKS Anywhere (EKS-A) cluster.",
	PreRunE:      bindFlagsToViper,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := uc.upgradePlanManagementComponents(cmd.Context()); err != nil {
			return fmt.Errorf("failed to display upgrade plan: %v", err)
		}
		return nil
	},
}

func init() {
	upgradePlanCmd.AddCommand(upgradePlanManagementComponentsCmd)
	upgradePlanManagementComponentsCmd.Flags().StringVarP(&uc.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	upgradePlanManagementComponentsCmd.Flags().StringVar(&uc.bundlesOverride, "bundles-override", "", "Override default Bundles manifest (not recommended)")
	upgradePlanManagementComponentsCmd.Flags().StringVarP(&output, outputFlagName, "o", outputDefault, "Output format: text|json")
	upgradePlanManagementComponentsCmd.Flags().StringVar(&uc.managementKubeconfig, "kubeconfig", "", "Management cluster kubeconfig file")
	err := upgradePlanManagementComponentsCmd.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func (uc *upgradeClusterOptions) upgradePlanManagementComponents(ctx context.Context) error {
	if _, err := uc.commonValidations(ctx); err != nil {
		return fmt.Errorf("common validations failed due to: %v", err)
	}

	newClusterSpec, err := newClusterSpec(uc.clusterOptions)
	if err != nil {
		return err
	}

	deps, err := dependencies.ForSpec(newClusterSpec).
		WithClusterManager(newClusterSpec.Cluster, nil).
		WithProvider(uc.fileName, newClusterSpec.Cluster, false, uc.hardwareCSVPath, uc.forceClean, uc.tinkerbellBootstrapIP, map[string]bool{}, uc.providerOptions).
		WithGitOpsFlux(newClusterSpec.Cluster, newClusterSpec.FluxConfig, nil).
		WithCAPIManager().
		Build(ctx)
	if err != nil {
		return err
	}

	managementCluster := &types.Cluster{
		Name:           newClusterSpec.Cluster.Name,
		KubeconfigFile: getKubeconfigPath(newClusterSpec.Cluster.Name, uc.wConfig),
	}

	if newClusterSpec.ManagementCluster != nil {
		managementCluster = newClusterSpec.ManagementCluster
	}

	logger.V(0).Info("Checking new release availability...")
	currentSpec, err := deps.ClusterManager.GetCurrentClusterSpec(ctx, managementCluster, newClusterSpec.Cluster.Name)
	if err != nil {
		return err
	}

	if !newClusterSpec.Cluster.IsSelfManaged() {
		logger.V(0).Info(fmt.Sprintf("No management components to plan. Cluster %s is not a self-managed cluster.", newClusterSpec.Cluster.Name))
		return nil
	}

	var componentChangeDiffs *types.ChangeDiff
	componentChangeDiffs, err = getManagementComponentsChangeDiffs(ctx, deps.UnAuthKubeClient, managementCluster, currentSpec, newClusterSpec, deps.Provider)
	if err != nil {
		return err
	}

	serializedDiff, err := serialize(componentChangeDiffs, output)
	if err != nil {
		return err
	}

	logger.V(0).Info(serializedDiff)

	return nil
}

func getManagementComponentsChangeDiffs(ctx context.Context, clientFactory interfaces.ClientFactory, managementCluster *types.Cluster, currentSpec *cluster.Spec, newClusterSpec *cluster.Spec, provider providers.Provider) (*types.ChangeDiff, error) {
	client, err := clientFactory.BuildClientFromKubeconfig(managementCluster.KubeconfigFile)
	if err != nil {
		return nil, err
	}

	currentManagementComponents, err := cluster.GetManagementComponents(ctx, client, currentSpec.Cluster)
	if err != nil {
		return nil, err
	}

	componentChangeDiffs := &types.ChangeDiff{}
	newManagementComponents := cluster.ManagementComponentsFromBundles(newClusterSpec.Bundles)
	componentChangeDiffs.Append(eksaupgrader.EksaChangeDiff(currentManagementComponents, newManagementComponents))
	componentChangeDiffs.Append(fluxupgrader.ChangeDiff(currentManagementComponents, newManagementComponents, currentSpec, newClusterSpec))
	componentChangeDiffs.Append(capiupgrader.ChangeDiff(currentManagementComponents, newManagementComponents, provider))

	return componentChangeDiffs, nil
}
