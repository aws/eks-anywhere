package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/release/pkg/pull_request"
	"github.com/aws/eks-anywhere/release/pkg/pull_request/types"
)

const (
	ghUserFlagName             = "gh-user"
	branchFlagName             = "base-branch"
	releaseTypeFlagName        = "release-type"
	releaseEnvironmentFlagName = "release-environment"
	bundleNumberFlagName       = "bundle-number"
	cliMinVersionFlagName      = "cli-min-version"
	cliMaxVersionFlagName      = "cli-max-version"
	releaseNumberFlagName      = "release-number"
	releaseVersionFlagName     = "release-version"
	dryRunFlagName             = "dry-run"
)

var requiredFlags = []string{ghUserFlagName, releaseTypeFlagName, releaseEnvironmentFlagName}

// pullRequestCmd represents the pull-request command
var pullRequestCmd = &cobra.Command{
	Use:   "pull-request",
	Short: "Create a pull request for eks-anywhere release",
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		if err != nil {
			fmt.Printf("Error initializing flags: %v\n", err)
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		ghUser := viper.GetString(ghUserFlagName)
		branch := viper.GetString(branchFlagName)
		releaseType := viper.GetString(releaseTypeFlagName)
		releaseEnvironment := viper.GetString(releaseEnvironmentFlagName)
		bundleNumber := viper.GetString(bundleNumberFlagName)
		cliMinVersion := viper.GetString(cliMinVersionFlagName)
		cliMaxVersion := viper.GetString(cliMaxVersionFlagName)
		releaseNumber := viper.GetString(releaseNumberFlagName)
		releaseVersion := viper.GetString(releaseVersionFlagName)
		dryRun := viper.GetBool(dryRunFlagName)

		if releaseType == "bundle" && (bundleNumber == "" || cliMinVersion == "" || cliMaxVersion == "") {
			fmt.Printf("No value provided for one or more of bundle-number, min-version and max-version.\n")
			os.Exit(1)
		}

		if releaseType == "eks-a" && (releaseNumber == "" || releaseVersion == "") {
			fmt.Printf("No value provided for one or more of release-number and release-version.\n")
			os.Exit(1)
		}

		pullRequestConfig := &types.PullRequestConfig{
			GithubUser:         ghUser,
			BaseBranch:         branch,
			ReleaseType:        releaseType,
			ReleaseEnvironment: releaseEnvironment,
			BundleNumber:       bundleNumber,
			CliMinVersion:      cliMinVersion,
			CliMaxVersion:      cliMaxVersion,
			ReleaseNumber:      releaseNumber,
			ReleaseVersion:     releaseVersion,
			DryRun:             dryRun,
		}

		err := pull_request.CreatePR(pullRequestConfig)
		if err != nil {
			fmt.Printf("Error creating release PR: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(pullRequestCmd)

	pullRequestCmd.Flags().String(ghUserFlagName, "", "The GitHub Username against whom the PR will be created")
	pullRequestCmd.Flags().String(branchFlagName, "main", "The branch to cut PRs against")
	pullRequestCmd.Flags().String(releaseTypeFlagName, "", "The type of EKS-D release (bundle/eks-a)")
	pullRequestCmd.Flags().String(releaseEnvironmentFlagName, "", "The release stage (development/production)")
	pullRequestCmd.Flags().String(bundleNumberFlagName, "", "The cardinal release number for this versioned bundle")
	pullRequestCmd.Flags().String(cliMinVersionFlagName, "", "The minimum version of EKS Anywhere supported by this bundle")
	pullRequestCmd.Flags().String(cliMaxVersionFlagName, "", "The maximum version of EKS Anywhere supported by this bundle")
	pullRequestCmd.Flags().String(releaseNumberFlagName, "", "The cardinal release number of EKS Anywhere")
	pullRequestCmd.Flags().String(releaseVersionFlagName, "", "The semantic release version of EKS Anywhere")
	pullRequestCmd.Flags().Bool(dryRunFlagName, true, "Flag to perform the pull request creation as a dry run")

	for _, flag := range requiredFlags {
		err := pullRequestCmd.MarkFlagRequired(flag)
		if err != nil {
			fmt.Printf("Error marking flag %s as required: %v", flag, err)
		}
	}
}
