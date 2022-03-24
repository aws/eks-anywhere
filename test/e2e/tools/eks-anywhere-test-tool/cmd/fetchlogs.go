package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere-test-tool/pkg/awsprofiles"
	"github.com/aws/eks-anywhere-test-tool/pkg/cloudwatch"
	"github.com/aws/eks-anywhere-test-tool/pkg/codebuild"
	"github.com/aws/eks-anywhere-test-tool/pkg/logfetcher"
)

const logToStdout = "stdout"

var fl = &e2eFetchOptions{}

var e2eFetchLogsCommand = &cobra.Command{
	Use:   "logs",
	Short: "fetch logs",
	Long:  "This command fetches the Cloudwatch logs associated with a given test execution",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Let's fetch some logs! \U0001FAB5")
		buildAccountCodebuild, err := codebuild.New(awsprofiles.BuildAccount)
		if err != nil {
			return fmt.Errorf("creating codebuild client: %v", err)
		}

		buildAccountCw, err := cloudwatch.New(awsprofiles.BuildAccount)
		if err != nil {
			return fmt.Errorf("creating cloudwatch logs client: %v", err)
		}

		testAccountCw, err := cloudwatch.New(awsprofiles.TestAccount)
		if err != nil {
			return fmt.Errorf("instantiating CW profile: %v", err)
		}

		var fetcherOpts []logfetcher.LogFetcherOpt
		if fl.tests != nil {
			fetcherOpts = append(fetcherOpts, logfetcher.WithTestFilterByName(fl.tests))
		}

		if fl.logTo == logToStdout {
			fetcherOpts = append(fetcherOpts, logfetcher.WithLogStdout())
		}

		fetcher := logfetcher.New(buildAccountCw, testAccountCw, buildAccountCodebuild, fetcherOpts...)

		var opts []logfetcher.FetchLogsOpt
		if fl.forBuildId != "" {
			opts = append(opts, logfetcher.WithCodebuildBuild(fl.forBuildId))
		}

		if fl.forProject != "" {
			opts = append(opts, logfetcher.WithCodebuildProject(fl.forProject))
		}
		return fetcher.FetchLogs(opts...)
	},
}

func init() {
	e2eFetchCommand.AddCommand(e2eFetchLogsCommand)
	e2eFetchLogsCommand.Flags().StringVar(&fl.forBuildId, "buildId", "", "Build ID to fetch logs for")
	e2eFetchLogsCommand.Flags().StringVar(&fl.forProject, "project", "", "Project to fetch builds from")
	e2eFetchLogsCommand.Flags().StringSliceVar(&fl.tests, "tests", nil, "Filter tests by name")
	e2eFetchLogsCommand.Flags().StringVar(&fl.logTo, "log-to", "", "Log output to")
	err := viper.BindPFlags(e2eFetchLogsCommand.Flags())
	if err != nil {
		log.Fatalf("Error initializing flags: %v", err)
	}
}
