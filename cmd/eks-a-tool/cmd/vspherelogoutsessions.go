package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
)

var sessions []string

var vsphereSessionRmCommand = &cobra.Command{
	Use:    "sessions",
	Short:  "vsphere logout sessions command",
	Long:   "This command logs out all of the provided VSphere user sessions ",
	PreRun: prerunSessionLogoutCmd,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := vsphereLogoutSessions(cmd.Context(), sessions)
		if err != nil {
			log.Fatalf("Error removing sessions: %v", err)
		}
		return nil
	},
}

func init() {
	vsphereRmCmd.AddCommand(vsphereSessionRmCommand)
	vsphereSessionRmCommand.Flags().StringSliceVarP(&sessions, "sessionTokens", "s", []string{}, "sessions to logout")
	err := vsphereSessionRmCommand.MarkFlagRequired("sessionTokens")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func prerunSessionLogoutCmd(cmd *cobra.Command, args []string) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
}

func vsphereLogoutSessions(ctx context.Context, sessions []string) error {
	tmpWriter, _ := filewriter.NewWriter("sessionlogout")
	executableBuilder, close, err := executables.NewExecutableBuilder(ctx, executables.DefaultEksaImage(), "./sessionlogout")
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	defer close.CheckErr(ctx)
	govc := executableBuilder.BuildGovcExecutable(tmpWriter)
	defer govc.Close(ctx)

	tmpWriter.CleanUp()
	return govc.Logout(ctx)
}
