package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/validations"
)

var vsphereRmVmsCmd = &cobra.Command{
	Use:   "vms <cluster-name>",
	Short: "VSphere rmvms command",
	Long:  "This command removes vms associated with a cluster name",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		var clusterName string

		clusterName, err = validations.ValidateClusterNameArg(args)
		if err != nil {
			return err
		}

		err = vsphereRmVms(cmd.Context(), clusterName, viper.GetBool("dry-run"))
		if err != nil {
			log.Fatalf("Error removing vms: %v", err)
		}
		return nil
	},
}

func init() {
	var err error

	vsphereRmCmd.AddCommand(vsphereRmVmsCmd)
	vsphereRmVmsCmd.Flags().Bool("dry-run", false, "Dry run flag")
	err = viper.BindPFlags(vsphereRmVmsCmd.Flags())
	if err != nil {
		log.Fatalf("Error initializing flags: %v", err)
	}
}

func vsphereRmVms(ctx context.Context, clusterName string, dryRun bool) error {
	executableBuilder, close, err := executables.InitInDockerExecutablesBuilder(ctx, executables.DefaultEksaImage())
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	defer close.CheckErr(ctx)
	tmpWriter, _ := filewriter.NewWriter("rmvms")
	govc := executableBuilder.BuildGovcExecutable(tmpWriter)
	defer govc.Close(ctx)

	return govc.CleanupVms(ctx, clusterName, dryRun)
}
