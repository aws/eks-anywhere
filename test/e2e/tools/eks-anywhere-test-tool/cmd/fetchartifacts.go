package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere-test-tool/pkg/artifacts"
	"github.com/aws/eks-anywhere-test-tool/pkg/awsprofiles"
	"github.com/aws/eks-anywhere-test-tool/pkg/cloudwatch"
	"github.com/aws/eks-anywhere-test-tool/pkg/codebuild"
	"github.com/aws/eks-anywhere-test-tool/pkg/constants"
	"github.com/aws/eks-anywhere-test-tool/pkg/fileutils"
	"github.com/aws/eks-anywhere-test-tool/pkg/filewriter"
	"github.com/aws/eks-anywhere-test-tool/pkg/s3"
	"github.com/aws/eks-anywhere/pkg/logger"
)

var fa = &e2eFetchOptions{}

var e2eFetchArtifactsCommand = &cobra.Command{
	Use:   "artifacts",
	Short: "fetch artifacts",
	Long:  "This command fetches the artifacts associated with a given test execution",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Let's fetch some artifacts! \U0001FAA3")

		_, present := os.LookupEnv(constants.E2eArtifactsBucketEnvVar)
		if !present {
			logger.MarkFail("E2E Test artifact bucket env var is not set!", "var", constants.E2eArtifactsBucketEnvVar)
			return fmt.Errorf("no e2e bucket env var set")
		}

		buildAccountCodebuild, err := codebuild.New(awsprofiles.BuildAccount)
		if err != nil {
			return fmt.Errorf("creating codebuild client: %v", err)
		}

		testAccountS3, err := s3.New(awsprofiles.TestAccount)
		if err != nil {
			return fmt.Errorf("creating s3 client: %v", err)
		}

		dir := fileutils.GenOutputDirName("artifacts")
		writer := filewriter.NewWriter(dir)
		if err != nil {
			return fmt.Errorf("setting up writer: %v", err)
		}

		buildAccountCw, err := cloudwatch.New(awsprofiles.BuildAccount)
		if err != nil {
			return fmt.Errorf("creating cloudwatch logs client: %v", err)
		}

		artifactFetcher := artifacts.New(testAccountS3, buildAccountCodebuild, writer, buildAccountCw)

		var opts []artifacts.FetchArtifactsOpt
		if fa.forBuildId != "" {
			opts = append(opts, artifacts.WithCodebuildBuild(fa.forBuildId))
		}

		if fa.forProject != "" {
			opts = append(opts, artifacts.WithCodebuildProject(fa.forProject))
		}

		if fa.fetchAll {
			opts = append(opts, artifacts.WithAllArtifacts())
		}

		return artifactFetcher.FetchArtifacts(opts...)
	},
}

func init() {
	e2eFetchCommand.AddCommand(e2eFetchArtifactsCommand)
	e2eFetchArtifactsCommand.Flags().StringVar(&fa.forBuildId, "buildId", "", "Build ID to fetch artifacts for")
	e2eFetchArtifactsCommand.Flags().StringVar(&fa.forProject, "project", "", "Project to fetch builds from")
	e2eFetchArtifactsCommand.Flags().BoolVar(&fa.fetchAll, "all", false, "Fetch all artifacts")
	err := viper.BindPFlags(e2eFetchArtifactsCommand.Flags())
	if err != nil {
		log.Fatalf("Error initializing flags: %v", err)
	}
}
