package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	fluxupgrader "github.com/aws/eks-anywhere/pkg/addonmanager/addonclients"
	eksaupgrader "github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
)

var upgradePlanCmd = &cobra.Command{
	Use:          "plan",
	Short:        "Provides recommended versions for upgrade",
	Long:         "Provide a list of recommended target versions for upgrading the core components in the workload cluster",
	PreRunE:      preRunUpgradePlanCluster,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := uc.upgradePlanCluster(cmd.Context()); err != nil {
			return fmt.Errorf("failed to display upgrade plan: %v", err)
		}
		return nil
	},
}

func preRunUpgradePlanCluster(cmd *cobra.Command, args []string) error {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
	return nil
}

func init() {
	upgradeCmd.AddCommand(upgradePlanCmd)
	upgradePlanCmd.Flags().StringVarP(&uc.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	upgradePlanCmd.Flags().StringVar(&uc.bundlesOverride, "bundles-override", "", "Override default Bundles manifest (not recommended)")
	err := upgradePlanCmd.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func (uc *upgradeClusterOptions) upgradePlanCluster(ctx context.Context) error {
	if _, err := uc.commonValidations(ctx); err != nil {
		return fmt.Errorf("common validations failed due to: %v", err)
	}

	newClusterSpec, err := newClusterSpec(uc.clusterOptions)
	if err != nil {
		return err
	}
	deps, err := dependencies.ForSpec(ctx, newClusterSpec).
		WithBootstrapper().
		WithClusterManager().
		WithFluxAddonClient(ctx, newClusterSpec.Cluster, newClusterSpec.GitOpsConfig).
		WithCAPIManager().
		Build(ctx)
	if err != nil {
		return err
	}

	workloadCluster := &types.Cluster{
		Name:           newClusterSpec.Name,
		KubeconfigFile: uc.kubeConfig(newClusterSpec.Name),
	}

	logger.V(0).Info("Checking new release availability...")
	currentSpec, err := deps.ClusterManager.GetCurrentClusterSpec(ctx, workloadCluster, newClusterSpec.Name)
	if err != nil {
		return err
	}

	eksaChangeDiff := eksaupgrader.ChangeDiff(currentSpec, newClusterSpec)
	fluxChangeDiff := fluxupgrader.FluxChangeDiff(currentSpec, newClusterSpec)
	w := tabwriter.NewWriter(os.Stdout, 10, 4, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tCURRENT VERSION\tNEXT VERSION")
	fmt.Fprintf(w, "%s\t%s\t%s\n", eksaChangeDiff.ComponentName, eksaChangeDiff.NewVersion, eksaChangeDiff.OldVersion)
	fmt.Fprintf(w, "%s\t%s\t%s\n", fluxChangeDiff.ComponentName, fluxChangeDiff.NewVersion, fluxChangeDiff.OldVersion)

	if err := w.Flush(); err != nil {
		fmt.Printf("Error %v", err)
	}
	fmt.Println("")

	return nil
}
