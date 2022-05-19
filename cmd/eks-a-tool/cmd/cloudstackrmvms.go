package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
	"github.com/aws/eks-anywhere/pkg/validations"
)

const dryRunFlag = "dry-run"

var cloudstackRmVmsCmd = &cobra.Command{
	Use:    "vms <cluster-name>",
	PreRun: prerunCmdBindFlags,
	Short:  "CloudStack rmvms command",
	Long:   "This command removes vms associated with a cluster name",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		var clusterName string

		clusterName, err = validations.ValidateClusterNameArg(args)
		if err != nil {
			return err
		}
		err = cloudstackRmVms(cmd.Context(), clusterName, viper.GetBool(dryRunFlag))
		if err != nil {
			log.Fatalf("Error removing vms: %v", err)
		}
		return nil
	},
}

func init() {
	var err error

	cloudstackRmCmd.AddCommand(cloudstackRmVmsCmd)
	cloudstackRmVmsCmd.Flags().Bool(dryRunFlag, false, "Dry run flag")
	err = viper.BindPFlags(cloudstackRmVmsCmd.Flags())
	if err != nil {
		log.Fatalf("Error initializing flags: %v", err)
	}
}

func cloudstackRmVms(ctx context.Context, clusterName string, dryRun bool) error {
	executableBuilder, close, err := executables.NewExecutableBuilder(ctx, executables.DefaultEksaImage())
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	defer close.CheckErr(ctx)
	tmpWriter, err := filewriter.NewWriter("rmvms")
	if err != nil {
		return fmt.Errorf("creating filewriter for directory rmvms: %v", err)
	}
	execConfig, err := decoder.ParseCloudStackSecret()
	if err != nil {
		return fmt.Errorf("building cmk executable: %v", err)
	}
	cmk := executableBuilder.BuildCmkExecutable(tmpWriter, *execConfig)
	defer cmk.Close(ctx)

	if err := cmk.ValidateCloudStackConnection(ctx); err != nil {
		return fmt.Errorf("validating cloudstack connection with cloudmonkey: %v", err)
	}

	return cmk.CleanupVms(ctx, clusterName, dryRun)
}
