package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	fluxupgrader "github.com/aws/eks-anywhere/pkg/addonmanager/addonclients"
	capiupgrader "github.com/aws/eks-anywhere/pkg/clusterapi"
	eksaupgrader "github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networking/cilium"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	outputFlagName = "output"
	outputDefault  = outputText
	outputText     = "text"
	outputJson     = "json"
)

var output string

var upgradePlanClusterCmd = &cobra.Command{
	Use:          "cluster",
	Short:        "Provides new release versions for the next cluster upgrade",
	Long:         "Provides a list of target versions for upgrading the core components in the workload cluster",
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
	upgradePlanCmd.AddCommand(upgradePlanClusterCmd)
	upgradePlanClusterCmd.Flags().StringVarP(&uc.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	upgradePlanClusterCmd.Flags().StringVar(&uc.bundlesOverride, "bundles-override", "", "Override default Bundles manifest (not recommended)")
	upgradePlanClusterCmd.Flags().StringVarP(&output, outputFlagName, "o", outputDefault, "Output format: text|json")
	err := upgradePlanClusterCmd.MarkFlagRequired("filename")
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
		WithClusterManager(newClusterSpec.Cluster).
		WithProvider(uc.fileName, newClusterSpec.Cluster, cc.skipIpCheck, uc.hardwareFileName, cc.skipPowerActions).
		WithFluxAddonClient(ctx, newClusterSpec.Cluster, newClusterSpec.GitOpsConfig).
		WithCAPIManager().
		Build(ctx)
	if err != nil {
		return err
	}

	workloadCluster := &types.Cluster{
		Name:           newClusterSpec.Cluster.Name,
		KubeconfigFile: getKubeconfigPath(newClusterSpec.Cluster.Name, uc.wConfig),
	}

	logger.V(0).Info("Checking new release availability...")
	currentSpec, err := deps.ClusterManager.GetCurrentClusterSpec(ctx, workloadCluster, newClusterSpec.Cluster.Name)
	if err != nil {
		return err
	}

	componentChangeDiffs := eksaupgrader.EksaChangeDiff(currentSpec, newClusterSpec)
	componentChangeDiffs.Append(fluxupgrader.FluxChangeDiff(currentSpec, newClusterSpec))
	componentChangeDiffs.Append(capiupgrader.CapiChangeDiff(currentSpec, newClusterSpec, deps.Provider))
	componentChangeDiffs.Append(cilium.ChangeDiff(currentSpec, newClusterSpec))

	serializedDiff, err := serialize(componentChangeDiffs, output)
	if err != nil {
		return err
	}

	fmt.Print(serializedDiff)

	return nil
}

func serialize(componentChangeDiffs *types.ChangeDiff, outputFormat string) (string, error) {
	switch outputFormat {
	case outputText:
		return serializeToText(componentChangeDiffs)
	case outputJson:
		return serializeToJson(componentChangeDiffs)
	default:
		return "", fmt.Errorf("invalid output format [%s]", outputFormat)
	}
}

func serializeToText(componentChangeDiffs *types.ChangeDiff) (string, error) {
	if componentChangeDiffs == nil {
		return "All the components are up to date with the latest versions", nil
	}

	buffer := bytes.Buffer{}
	w := tabwriter.NewWriter(&buffer, 10, 4, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tCURRENT VERSION\tNEXT VERSION")
	for i := range componentChangeDiffs.ComponentReports {
		fmt.Fprintf(w, "%s\t%s\t%s\n", componentChangeDiffs.ComponentReports[i].ComponentName, componentChangeDiffs.ComponentReports[i].OldVersion, componentChangeDiffs.ComponentReports[i].NewVersion)
	}
	if err := w.Flush(); err != nil {
		return "", fmt.Errorf("failed flushing table writer: %v", err)
	}

	return buffer.String(), nil
}

func serializeToJson(componentChangeDiffs *types.ChangeDiff) (string, error) {
	if componentChangeDiffs == nil {
		componentChangeDiffs = &types.ChangeDiff{ComponentReports: []types.ComponentChangeDiff{}}
	}

	jsonDiff, err := json.Marshal(componentChangeDiffs)
	if err != nil {
		return "", fmt.Errorf("failed serializing the components diff to json: %v", err)
	}

	return string(jsonDiff), nil
}
