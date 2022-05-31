package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/internal/test/e2e"
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
		err = e2e.CleanUpCloudstackTestResources(cmd.Context(), clusterName, viper.GetBool(dryRunFlag))
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
