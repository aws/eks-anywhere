package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/logger"
)

var vsphereIntegrationTestCmd = &cobra.Command{
	Use:    "vsphere",
	Short:  "End to end test with vsphere",
	Long:   "Run end to end test with vsphere",
	PreRun: preRunSetup,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := runvsphereIntegrationTest(cmd.Context())
		if err != nil {
			logger.Fatal(err, "Failed to run vsphere e2e test")
		}
		return nil
	},
}

func init() {
	integrationTestCmd.AddCommand(vsphereIntegrationTestCmd)
	vsphereIntegrationTestCmd.Flags().StringP("ami-id", "a", "", "Ami id")
	vsphereIntegrationTestCmd.Flags().StringP("storage-bucket", "s", "", "S3 bucket name to store eks-a binary")
	vsphereIntegrationTestCmd.Flags().StringP("job-id", "j", "", "Id of the job being run")
	vsphereIntegrationTestCmd.Flags().StringP("instance-profile-name", "i", "", "IAM instance profile name to attach to ec2 instances")
	vsphereIntegrationTestCmd.Flags().StringP("subnet-id", "n", "", "EC2 subnet ID")
	vsphereIntegrationTestCmd.Flags().StringP("dry-run", "d", "false", "Dry run flag")
}

func runvsphereIntegrationTest(ctx context.Context) error {
	fmt.Println("Running e2e test with vsphere")
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
	dryRun := viper.GetString("dry-run")

	if dryRun == "true" {
		return nil
	}

	session := session.Must(session.NewSession())
	fmt.Println("Upload eks-a binary to s3 bucket")
	file := "bin/linux/eks-a"
	key := jobId + "/eks-a"
	uploadToS3Bucket(session, file, key, storageBucket)
	fmt.Println("Upload Completed")

	fmt.Println("Upload eks-a-tool binary to s3 bucket")
	file = "bin/eks-a-tool"
	key = jobId + "/eks-a-tool"
	uploadToS3Bucket(session, file, key, storageBucket)
	fmt.Println("Upload Completed")

	fmt.Println("Upload cluster config to s3 bucket")
	file = "e2e-test"
	key = jobId + "/e2e-test"
	uploadToS3Bucket(session, file, key, storageBucket)
	fmt.Println("Upload Completed")

	fmt.Println("Create an ec2 instance with an instance profile and tag it")
	key = "Integration-Test"
	tag := "EKSA-VSPHERE"
	instanceId := createEc2Instance(session, amiId, key, tag, instanceProfileName, subnetId)
	fmt.Println("InstanceId: ", *instanceId)

	var err error

	command := "ls"
	retryCount := 0
	fmt.Println("Run ssm sendCommand to execute test command: ", command)

	for retryCount <= 9 {
		err = runAndValidateSsmCommand(session, instanceId, command)
		if err == nil {
			break
		}
		fmt.Println("Retrying ssm send to make sure that the ec2 instance is registered with ssm")
		retryCount = retryCount + 1
		time.Sleep(20 * time.Second)
	}

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully executed test command")

	fmt.Println("Run ssm sendCommand to download eks-a, eks-a-tools binary and cluster config from the s3 bucket")
	command = "aws s3 cp s3://" + storageBucket + "/" + jobId + "/ . --recursive; chmod 645 eks-a; chmod 645 eks-a-tool"
	fmt.Println(command)
	err = runAndValidateSsmCommand(session, instanceId, command)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully downloaded resources from S3 bucket")

	fmt.Println("Run ssm sendCommand to get a unique IP for Control plane IP and set CONTROL_PLANE_ENDPOINT_IP")
	command = getVSphereEnvVarCommand()
	command += "export CONTROL_PLANE_ENDPOINT_IP=$(./eks-a-tool unique-ip -c 198.18.0.0/16); "
	command += "echo $CONTROL_PLANE_ENDPOINT_IP; "
	command += "./eks-a-tool vsphere autofill -f e2e-test; cat e2e-test;"
	fmt.Println(command)
	err = runAndValidateSsmCommand(session, instanceId, command)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully generated unique IP and set CONTROL_PLANE_ENDPOINT_IP")

	fmt.Println("Run ssm sendCommand to rename the cluster config")
	command = "mv e2e-test e2e-test.yaml;"
	fmt.Println(command)
	err = runAndValidateSsmCommand(session, instanceId, command)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully renamed cluster config")

	fmt.Println("Run ssm sendCommand to create eks-a cluster on vSphere")
	command = getVSphereEnvVarCommand()
	command += "./eks-a create cluster -v 4 -f e2e-test.yaml"
	fmt.Println(command)
	err = runAndValidateSsmCommand(session, instanceId, command)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully created eks-a cluster")

	fmt.Println("Run ssm sendCommand to delete eks-a cluster on vSphere")
	command = getVSphereEnvVarCommand()
	command += "./eks-a delete cluster -v 4 -f e2e-test.yaml"
	fmt.Println(command)
	err = runAndValidateSsmCommand(session, instanceId, command)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully deleted eks-a cluster")

	fmt.Println("Test ran Successfully")
	return nil
}
