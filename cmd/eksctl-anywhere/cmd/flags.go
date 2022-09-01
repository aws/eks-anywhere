package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func bindFlagsToViper(cmd *cobra.Command, args []string) error {
	var err error
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if err != nil {
			return
		}
		err = viper.BindPFlag(flag.Name, flag)
	})
	return err
}
