package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere-test-tool/pkg/awsprofiles"
	"github.com/aws/eks-anywhere-test-tool/pkg/cloudwatch"
	"github.com/aws/eks-anywhere-test-tool/pkg/codebuild"
	"github.com/aws/eks-anywhere-test-tool/pkg/filewriter"
	"github.com/aws/eks-anywhere-test-tool/pkg/logfetcher"
)

var fl = &e2eFetchOptions{}

var e2eFetchLogsCommand = &cobra.Command{
	Use:   "logs",
	Short: "fetch logs",
	Long:  "This command fetches the Cloudwatch logs associated with a given test execution",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Let's fetch some logs! \U0001FAB5")
		buildAccountCodebuild, err := codebuild.New(awsprofiles.BuildAccount)
		if err != nil {
			return fmt.Errorf("error when creating codebuild client: %v", err)
		}

		buildAccountCw, err := cloudwatch.New(awsprofiles.BuildAccount)
		if err != nil {
			return fmt.Errorf("error when creating cloudwatch logs client: %v", err)
		}

		testAccountCw, err := cloudwatch.New(awsprofiles.TestAccount)
		if err != nil {
			return fmt.Errorf("error when instantiating CW profile: %v", err)
		}

		now := time.Now().Format(time.RFC3339 + "-logs")
		writer, err := filewriter.NewWriter(now)
		if err != nil {
			return fmt.Errorf("error when setting up writer: %v", err)
		}

		var opts []logfetcher.FetchLogsOpt
		if fl.forBuildId != "" {
			opts = append(opts, logfetcher.WithCodebuildBuild(fl.forBuildId))
		}

		if fl.forProject != "" {
			opts = append(opts, logfetcher.WithCodebuildProject(fl.forProject))
		}

		fetcher := logfetcher.New(buildAccountCw, testAccountCw, buildAccountCodebuild, writer)
		return fetcher.FetchLogs(opts...)
	},
}

func init() {
	e2eFetchCommand.AddCommand(e2eFetchLogsCommand)
	e2eFetchLogsCommand.Flags().StringVar(&fl.forBuildId, "buildId", "", "Build ID to fetch logs for")
	e2eFetchLogsCommand.Flags().StringVar(&fl.forProject, "project", "", "Project to fetch builds from")
	err := viper.BindPFlags(e2eFetchLogsCommand.Flags())
	if err != nil {
		log.Fatalf("Error initializing flags: %v", err)
	}
}
