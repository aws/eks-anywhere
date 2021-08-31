package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func uploadToS3Bucket(session *session.Session, file string, key string, bucket string) {
	s3Uploader := s3manager.NewUploader(session)
	fileBody, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	_, err = s3Uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   fileBody,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func createEc2Instance(session *session.Session, amiId string, key string, tag string, instanceProfileName string, subnetId string) *string {
	service := ec2.New(session)

	input := ec2.RunInstancesInput{
		ImageId:      aws.String(amiId),
		InstanceType: aws.String("t3.2xlarge"),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/xvda"),
				Ebs: &ec2.EbsBlockDevice{
					VolumeSize: aws.Int64(100),
				},
			},
		},
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Name: aws.String(instanceProfileName),
		},
		SubnetId: aws.String(subnetId),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String(key),
						Value: aws.String(tag),
					},
				},
			},
		},
	}
	if subnetId != "" {
		input.SubnetId = aws.String(subnetId)
	}
	result, err := service.RunInstances(&input)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Instance created and Tagged")

	fmt.Println("Wait till the instance starts running")
	instanceInput := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			result.Instances[0].InstanceId,
		},
	}
	err = service.WaitUntilInstanceRunning(instanceInput)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Instance is running")

	return result.Instances[0].InstanceId
}

func tagEc2Instance(session *session.Session, instanceId *string, key string, value string) {
	fmt.Println("Tag ec2 instance")
	service := ec2.New(session)
	_, err := service.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{instanceId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String(key),
				Value: aws.String(value),
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully tagged ec2 instance")
}

func runAndValidateSsmCommand(session *session.Session, instanceId *string, command string) error {
	service := ssm.New(session)
	fmt.Println("Running ssm command")
	result, err := service.SendCommand(
		&ssm.SendCommandInput{
			DocumentName: aws.String("AWS-RunShellScript"),
			InstanceIds:  []*string{instanceId},
			Parameters:   map[string][]*string{"commands": {aws.String(command)}},
		},
	)
	if err != nil {
		return err
	}

	// Make sure ssm send command is registered
	retryCount := 0
	registrationErr := err
	commandStatus := ""
	fmt.Println("Wait till the ssm command is registered")
	for retryCount <= 9 {
		status, registrationErr := service.GetCommandInvocation(
			&ssm.GetCommandInvocationInput{
				CommandId:  result.Command.CommandId,
				InstanceId: instanceId,
			},
		)
		if registrationErr != nil {
			time.Sleep(20 * time.Second)
			retryCount = retryCount + 1
			continue
		} else {
			commandStatus = *status.Status
			fmt.Println("Command status: ", commandStatus)
			if commandStatus != "Pending" && commandStatus != "InProgress" && commandStatus != "Delayed" {
				fmt.Println("Command stdout: ", *status.StandardOutputContent)
				fmt.Println("Command stderr: ", *status.StandardErrorContent)
				fmt.Println("Command statusDetails: ", *status.StatusDetails)
			}
			break
		}
	}
	if registrationErr != nil {
		return registrationErr
	}
	if commandStatus == "Success" {
		return nil
	}

	fmt.Println("Checking command status")
	retryCount = 0
	for retryCount <= 60 {
		status, err := service.GetCommandInvocation(
			&ssm.GetCommandInvocationInput{
				CommandId:  result.Command.CommandId,
				InstanceId: instanceId,
			},
		)
		if err != nil {
			return nil
		}
		commandStatus = *status.Status
		fmt.Println("Command status: ", commandStatus)
		if commandStatus != "Pending" && commandStatus != "InProgress" && commandStatus != "Delayed" {
			fmt.Println("Command stdout: ", *status.StandardOutputContent)
			fmt.Println("Command stderr: ", *status.StandardErrorContent)
			fmt.Println("Command statusDetails: ", *status.StatusDetails)
			break
		}
		retryCount = retryCount + 1
		time.Sleep(20 * time.Second)
	}
	if commandStatus != "Success" {
		return errors.New("failed to execute ssm command")
	}
	return nil
}

func getVSphereEnvVarCommand() string {
	command := ""
	if vSphereUsername, ok := os.LookupEnv("VSPHERE_USERNAME"); ok && len(vSphereUsername) > 0 {
		command += "export VSPHERE_USERNAME=" + vSphereUsername + ";"
	}
	if vSpherePassword, ok := os.LookupEnv("VSPHERE_PASSWORD"); ok && len(vSpherePassword) > 0 {
		command += "export VSPHERE_PASSWORD=" + vSpherePassword + ";"
	}
	if govcURL, ok := os.LookupEnv("GOVC_URL"); ok && len(govcURL) > 0 {
		command += "export GOVC_URL=" + govcURL + ";"
	}
	if govcInsecure, ok := os.LookupEnv("GOVC_INSECURE"); ok && len(govcInsecure) > 0 {
		command += "export GOVC_INSECURE=" + govcInsecure + ";"
	}
	return command
}

func createTestSetup(session *session.Session, amiId string, instanceProfileName string, storageBucket string, jobId string, subnetId string) *string {
	fmt.Println("Upload eks-a binary to s3 bucket")
	file := "bin/linux/eks-a"
	key := jobId + "/eks-a"
	uploadToS3Bucket(session, file, key, storageBucket)
	fmt.Println("Upload Completed")

	fmt.Println("Create an ec2 instance with an instance profile and tag it")
	key = "Integration-Test"
	tag := "EKSA-DOCKER"
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

	fmt.Println("Run ssm sendCommand to download eks-a binary from the s3 bucket")
	command = "aws s3 cp s3://" + storageBucket + "/" + jobId + "/eks-a .; chmod 645 eks-a"
	fmt.Println(command)
	err = runAndValidateSsmCommand(session, instanceId, command)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully downloaded eks-a binary")
	return instanceId
}

func createCluster(session *session.Session, instanceId *string) {
	fmt.Println("Run ssm sendCommand to generate cluster config")
	command := "./eks-a generate clusterconfig -p docker e2e-docker > eksa-cluster.yaml"
	fmt.Println(command)
	err := runAndValidateSsmCommand(session, instanceId, command)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully generated cluster config")

	fmt.Println("Run ssm sendCommand to print the cluster config")
	command = "cat eksa-cluster.yaml"
	fmt.Println(command)
	err = runAndValidateSsmCommand(session, instanceId, command)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully printed cluster config")

	fmt.Println("Run ssm sendCommand to create eks-a cluster")
	command = "./eks-a create cluster -v 6 -f eksa-cluster.yaml"
	fmt.Println(command)
	err = runAndValidateSsmCommand(session, instanceId, command)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully created eks-a cluster")
}

func deleteCluster(session *session.Session, instanceId *string) {
	fmt.Println("Run ssm sendCommand to delete eks-a cluster")
	command := "./eks-a delete cluster -v 6 -f eksa-cluster.yaml"
	fmt.Println(command)
	err := runAndValidateSsmCommand(session, instanceId, command)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully deleted eks-a cluster")
}
