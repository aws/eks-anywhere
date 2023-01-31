package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/internal/test/cleanup"
	"github.com/aws/eks-anywhere/pkg/validations"
)

const (
	endpointFlag     = "endpoint"
	portFlag         = "port"
	insecureFlag     = "insecure"
	ignoreErrorsFlag = "ignoreErrors"
)

var nutanixRmVmsCmd = &cobra.Command{
	Use:    "vms <cluster-name>",
	PreRun: prerunCmdBindFlags,
	Short:  "Nutanix rmvms command",
	Long:   "This command removes vms associated with a cluster name",
	RunE: func(cmd *cobra.Command, args []string) error {
		clusterName, err := validations.ValidateClusterNameArg(args)
		if err != nil {
			return err
		}
		insecure := false
		if viper.IsSet(insecureFlag) {
			insecure = true
		}
		err = cleanup.NutanixTestResourcesCleanup(cmd.Context(), clusterName, viper.GetString(endpointFlag), viper.GetString(portFlag), insecure, viper.GetBool(ignoreErrorsFlag))
		if err != nil {
			log.Fatalf("Error removing vms: %v", err)
		}
		return nil
	},
}

func init() {
	nutanixRmCmd.AddCommand(nutanixRmVmsCmd)

	nutanixRmVmsCmd.Flags().StringP(endpointFlag, "e", "", "specify Nutanix Prism endpoint (REQUIRED)")
	nutanixRmVmsCmd.Flags().StringP(portFlag, "p", "9440", "specify Nutanix Prism port (default: 9440)")
	nutanixRmVmsCmd.Flags().StringP(insecureFlag, "k", "false", "skip TLS when contacting Prism APIs (default: false)")
	nutanixRmVmsCmd.Flags().String(ignoreErrorsFlag, "true", "ignore APIs errors when deleting VMs (default: true)")

	if err := nutanixRmVmsCmd.MarkFlagRequired(endpointFlag); err != nil {
		log.Fatalf("Marking flag '%s' as required", endpointFlag)
	}
}
