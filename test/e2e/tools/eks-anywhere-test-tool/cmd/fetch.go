package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type e2eFetchOptions struct {
	forBuildId string
	forProject string
	tests      []string
	logTo      string
	fetchAll   bool
}

var e2eFetchCommand = &cobra.Command{
	Use:   "fetch",
	Short: "e2e fetch command",
	Long:  "This command fetches various artifacts and logs from the e2e tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("sup it's the command fetch")
		return nil
	},
}

func init() {
	e2eCmd.AddCommand(e2eFetchCommand)
	err := viper.BindPFlags(e2eFetchCommand.Flags())
	if err != nil {
		log.Fatalf("Error initializing flags: %v", err)
	}
}
