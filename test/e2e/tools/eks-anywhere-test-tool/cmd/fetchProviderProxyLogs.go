package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere-test-tool/pkg/awsprofiles"
	"github.com/aws/eks-anywhere-test-tool/pkg/cloudwatch"
	"github.com/aws/eks-anywhere-test-tool/pkg/codebuild"
	"github.com/aws/eks-anywhere-test-tool/pkg/providerProxy"
)

type proxyLogsFetchOptions struct {
	forBuildId string
	forProject string
	logTo      string
}

var fs = &proxyLogsFetchOptions{}

var e2eFetchProxyLogsCommand = &cobra.Command{
	Use:   "providerProxyLogs",
	Short: "fetch provider proxy logs associated with a give build execution",
	Long:  "This command fetches proxy logs which capture wire communication between our test clusters and the EKS A vsphere endpoint.",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		var fetcherOpts []providerProxy.ProxyFetcherOpt

		if fs.logTo == logToStdout {
			fetcherOpts = append(fetcherOpts, providerProxy.WithLogStdout())
		}

		fetcher := providerProxy.New(buildAccountCw, testAccountCw, buildAccountCodebuild, fetcherOpts...)

		var opts []providerProxy.FetchSessionOpts
		if fs.forBuildId != "" {
			opts = append(opts, providerProxy.WithCodebuildBuild(fs.forBuildId))
		}

		if fs.forProject != "" {
			opts = append(opts, providerProxy.WithCodebuildProject(fs.forProject))
		}

		return fetcher.FetchProviderProxyLogs(opts...)
	},
}

func init() {
	e2eFetchCommand.AddCommand(e2eFetchProxyLogsCommand)
	e2eFetchProxyLogsCommand.Flags().StringVar(&fs.forBuildId, "buildId", "", "Build ID to fetch logs for")
	e2eFetchProxyLogsCommand.Flags().StringVar(&fs.forProject, "project", "", "Project to fetch builds from")
	e2eFetchProxyLogsCommand.Flags().StringVar(&fs.logTo, "log-to", "", "Log output to")
	err := viper.BindPFlags(e2eFetchProxyLogsCommand.Flags())
	if err != nil {
		log.Fatalf("Error initializing flags: %v", err)
	}
}
