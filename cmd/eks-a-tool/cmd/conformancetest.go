package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/internal/pkg/conformance"
)

var conformanceTestCmd = &cobra.Command{
	Use:   "test <cluster-context>",
	Short: "Conformance test command",
	Long:  "This command run the conformance tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			log.Fatalf("Error running sonobuoy: cluster context not provided")
		}
		results, err := conformance.RunTests(cmd.Context(), args[0])
		if err != nil {
			log.Fatalf("Error running sonobuoy: %v", err)
		}
		log.Printf("Conformance Test results:\n %v", results)
		return nil
	},
}

func init() {
	conformanceCmd.AddCommand(conformanceTestCmd)
	err := viper.BindPFlags(conformanceTestCmd.Flags())
	if err != nil {
		log.Fatalf("Error initializing flags: %v", err)
	}
}
