package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/logger"
)

var dockerIntegrationTestCmd = &cobra.Command{
	Use:    "docker",
	Short:  "End to end test with docker",
	Long:   "Run end to end test with docker",
	PreRun: preRunSetup,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := runDockerIntegrationTest(cmd.Context())
		if err != nil {
			logger.Fatal(err, "Failed to run docker e2e test")
		}
		return nil
	},
}

func preRunSetup(cmd *cobra.Command, args []string) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
}

func init() {
	integrationTestCmd.AddCommand(dockerIntegrationTestCmd)
	dockerIntegrationTestCmd.Flags().StringP("ami-id", "a", "", "Ami id")
	dockerIntegrationTestCmd.Flags().StringP("storage-bucket", "s", "", "S3 bucket name to store eks-a binary")
	dockerIntegrationTestCmd.Flags().StringP("job-id", "j", "", "Id of the job being run")
	dockerIntegrationTestCmd.Flags().StringP("instance-profile-name", "i", "", "IAM instance profile name to attach to ec2 instances")
	dockerIntegrationTestCmd.Flags().StringP("subnet-id", "n", "", "EC2 subnet ID")
	dockerIntegrationTestCmd.Flags().StringP("dry-run", "d", "false", "Dry run flag")
}

func runDockerIntegrationTest(ctx context.Context) error {
	fmt.Println("Running e2e test with docker")
	amiId := viper.GetString("ami-id")
	if amiId == "" {
		log.Fatal(errors.New("no ami-id was provided"))
	}
	storageBucket := viper.GetString("storage-bucket")
	if storageBucket == "" {
		log.Fatal(errors.New("no storage bucket name was  provided"))
	}
	jobId := viper.GetString("job-id")

	if jobId == "" {
		log.Fatal(errors.New("no job id was provided"))
	}

	instanceProfileName := viper.GetString("instance-profile-name")

	if instanceProfileName == "" {
		log.Fatal(errors.New("no instance profile name was provided"))
	}

	subnetId := viper.GetString("subnet-id")
	if subnetId == "" {
		log.Fatal(errors.New("no subnet id  was provided"))
	}

	dryRun := viper.GetString("dry-run")

	if dryRun == "true" {
		return nil
	}

	session := session.Must(session.NewSession())
	// Sending subnet id as an empty string for now. Will send an actual id when we create
	// a subnet for this
	instanceId := createTestSetup(session, amiId, instanceProfileName, storageBucket, jobId, "")
	createCluster(session, instanceId)
	deleteCluster(session, instanceId)

	key := "Integration-Test-Done"
	value := "TRUE"
	tagEc2Instance(session, instanceId, key, value)

	fmt.Println("Test ran successfully")
	return nil
}
