package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/internal/test/e2e"
	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	amiIdFlagName              = "ami-id"
	storageBucketFlagName      = "storage-bucket"
	jobIdFlagName              = "job-id"
	instanceProfileFlagName    = "instance-profile-name"
	subnetIdFlagName           = "subnet-id"
	regexFlagName              = "regex"
	maxInstancesFlagName       = "max-instances"
	maxConcurrentTestsFlagName = "max-concurrent-tests"
	skipFlagName               = "skip"
	bundlesOverrideFlagName    = "bundles-override"
	cleanupVmsFlagName         = "cleanup-vms"
	testReportFolderFlagName   = "test-report-folder"
	branchNameFlagName         = "branch-name"
)

var runE2ECmd = &cobra.Command{
	Use:          "run",
	Short:        "Run E2E",
	Long:         "Run end to end tests",
	SilenceUsage: true,
	PreRun:       preRunSetup,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := runE2E(cmd.Context())
		if err != nil {
			logger.Fatal(err, "Failed to run e2e test")
		}
		return nil
	},
}

var requiredFlags = []string{amiIdFlagName, storageBucketFlagName, jobIdFlagName, instanceProfileFlagName}

func preRunSetup(cmd *cobra.Command, args []string) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
}

func init() {
	integrationTestCmd.AddCommand(runE2ECmd)
	runE2ECmd.Flags().StringP(amiIdFlagName, "a", "", "Ami id")
	runE2ECmd.Flags().StringP(storageBucketFlagName, "s", "", "S3 bucket name to store eks-a binary")
	runE2ECmd.Flags().StringP(jobIdFlagName, "j", "", "Id of the job being run")
	runE2ECmd.Flags().StringP(instanceProfileFlagName, "i", "", "IAM instance profile name to attach to ec2 instances")
	runE2ECmd.Flags().StringP(subnetIdFlagName, "n", "", "EC2 subnet ID")
	runE2ECmd.Flags().StringP(regexFlagName, "r", "", "Run only those tests and examples matching the regular expression. Equivalent to go test -run")
	runE2ECmd.Flags().IntP(maxInstancesFlagName, "m", 1, "Run tests in parallel on same instance within the max EC2 instance count")
	runE2ECmd.Flags().IntP(maxConcurrentTestsFlagName, "c", 1, "Maximum number of parallel tests that can be run at a time")
	runE2ECmd.Flags().StringSlice(skipFlagName, nil, "List of tests to skip")
	runE2ECmd.Flags().Bool(bundlesOverrideFlagName, false, "Flag to indicate if the tests should run with a bundles override")
	runE2ECmd.Flags().Bool(cleanupVmsFlagName, false, "Flag to indicate if VSphere VMs should be cleaned up automatically as tests complete")
	runE2ECmd.Flags().String(testReportFolderFlagName, "", "Folder destination fo JUnit tests reports")
	runE2ECmd.Flags().String(branchNameFlagName, "main", "EKS-A origin branch from where the tests are being run")

	for _, flag := range requiredFlags {
		if err := runE2ECmd.MarkFlagRequired(flag); err != nil {
			log.Fatalf("Error marking flag %s as required: %v", flag, err)
		}
	}
}

func runE2E(ctx context.Context) error {
	amiId := viper.GetString(amiIdFlagName)
	storageBucket := viper.GetString(storageBucketFlagName)
	jobId := viper.GetString(jobIdFlagName)
	instanceProfileName := viper.GetString(instanceProfileFlagName)
	subnetId := viper.GetString(subnetIdFlagName)
	testRegex := viper.GetString(regexFlagName)
	maxInstances := viper.GetInt(maxInstancesFlagName)
	maxConcurrentTests := viper.GetInt(maxConcurrentTestsFlagName)
	testsToSkip := viper.GetStringSlice(skipFlagName)
	bundlesOverride := viper.GetBool(bundlesOverrideFlagName)
	cleanupVms := viper.GetBool(cleanupVmsFlagName)
	testReportFolder := viper.GetString(testReportFolderFlagName)
	branchName := viper.GetString(branchNameFlagName)

	runConf := e2e.ParallelRunConf{
		MaxInstances:        maxInstances,
		MaxConcurrentTests:  maxConcurrentTests,
		AmiId:               amiId,
		InstanceProfileName: instanceProfileName,
		StorageBucket:       storageBucket,
		JobId:               jobId,
		SubnetId:            subnetId,
		Regex:               testRegex,
		TestsToSkip:         testsToSkip,
		BundlesOverride:     bundlesOverride,
		CleanupVms:          cleanupVms,
		TestReportFolder:    testReportFolder,
		BranchName:          branchName,
	}

	err := e2e.RunTestsInParallel(runConf)
	if err != nil {
		return fmt.Errorf("running e2e tests: %v", err)
	}

	return nil
}
