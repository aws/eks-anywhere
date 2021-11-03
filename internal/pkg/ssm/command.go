package ssm

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

const ssmLogGroup = "/eks-anywhere/test/e2e"

var initE2EDirCommand = "mkdir -p /home/e2e/bin && cd /home/e2e"

func WaitForSSMReady(session *session.Session, instanceId string) error {
	err := retrier.Retry(10, 20*time.Second, func() error {
		_, err := Run(session, instanceId, "ls")
		return err
	})
	if err != nil {
		return fmt.Errorf("error waiting for ssm to be ready: %v", err)
	}

	return nil
}

type CommandOpt func(c *ssm.SendCommandInput)

func WithOutputToS3(bucket, dir string) CommandOpt {
	return func(c *ssm.SendCommandInput) {
		c.OutputS3BucketName = aws.String(bucket)
		c.OutputS3KeyPrefix = aws.String(dir)
	}
}

func WithOutputToCloudwatch() CommandOpt {
	return func(c *ssm.SendCommandInput) {
		cwEnabled := true
		logGroup := ssmLogGroup
		cw := ssm.CloudWatchOutputConfig{
			CloudWatchLogGroupName:  &logGroup,
			CloudWatchOutputEnabled: &cwEnabled,
		}
		c.CloudWatchOutputConfig = &cw
	}
}

var nonFinalStatuses = map[string]struct{}{
	ssm.CommandInvocationStatusInProgress: {}, ssm.CommandInvocationStatusDelayed: {}, ssm.CommandInvocationStatusPending: {},
}

// TODO: cleanup this method
func Run(session *session.Session, instanceId, command string, opts ...CommandOpt) (commandId string, err error) {
	service := ssm.New(session)

	c := &ssm.SendCommandInput{
		DocumentName: aws.String("AWS-RunShellScript"),
		InstanceIds:  []*string{aws.String(instanceId)},
		Parameters:   map[string][]*string{"commands": {aws.String(initE2EDirCommand), aws.String(command)}, "executionTimeout": {aws.String("10800")}},
	}

	for _, opt := range opts {
		opt(c)
	}
	var result *ssm.SendCommandOutput
	r := retrier.New(180*time.Minute, retrier.WithRetryPolicy(func(totalRetries int, err error) (retry bool, wait time.Duration) {
		if request.IsErrorThrottle(err) && totalRetries < 60 {
			return true, 60 * time.Second
		}
		return false, 0
	}))
	err = r.Retry(func() error {
		var err error
		logger.V(2).Info("Running ssm command", "cmd", command)
		result, err = service.SendCommand(c)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("retries exhausted sending ssm command: %v", err)
	}

	logger.V(2).Info("SSM command started", "commandId", result.Command.CommandId)
	if c.OutputS3BucketName != nil {
		logger.V(4).Info(
			"SSM command output to S3", "url",
			fmt.Sprintf("s3://%s/%s/%s/%s/awsrunShellScript/0.awsrunShellScript/stderr", *c.OutputS3BucketName, *c.OutputS3KeyPrefix, *result.Command.CommandId, instanceId),
		)
	}

	commandId = *result.Command.CommandId
	commandIn := &ssm.GetCommandInvocationInput{
		CommandId:  &commandId,
		InstanceId: aws.String(instanceId),
	}

	// Make sure ssm send command is registered
	logger.V(2).Info("Waiting for ssm command to be registered")
	err = retrier.Retry(10, 5*time.Second, func() error {
		_, err := service.GetCommandInvocation(commandIn)
		if err != nil {
			return fmt.Errorf("error getting ssm command invocation: %v", err)
		}

		return nil
	})
	if err != nil {
		return commandId, fmt.Errorf("error waiting for ssm command to be registered: %v", err)
	}

	logger.V(2).Info("Waiting for ssm command to finish")
	var commandOut *ssm.GetCommandInvocationOutput
	r = retrier.New(180*time.Minute, retrier.WithMaxRetries(2160, 60*time.Second))
	err = r.Retry(func() error {
		var err error
		commandOut, err = service.GetCommandInvocation(commandIn)
		if err != nil {
			return err
		}

		status := *commandOut.Status
		if isFinalStatus(status) {
			logger.V(2).Info("SSM command finished", "status", status, "commandId", result.Command.CommandId)
			// TODO: these outputs might be truncated (8000 chars max). Get the logs from s3 with StandardErrorUrl and StandardOutputContent instead
			fmt.Println("Command stdout:")
			fmt.Println(*commandOut.StandardOutputContent)
			printDivider()
			fmt.Println("Command stderr")
			fmt.Println(*commandOut.StandardErrorContent)
			printDivider()

			return nil
		}

		return fmt.Errorf("command still running with status %s", status)
	})
	if err != nil {
		return commandId, fmt.Errorf("retries exhausted running ssm command: %v", err)
	}

	if *commandOut.Status != ssm.CommandInvocationStatusSuccess {
		return commandId, errors.New("failed to execute ssm command")
	}

	return commandId, nil
}

func isFinalStatus(status string) bool {
	_, nonFinalStatus := nonFinalStatuses[status]
	return !nonFinalStatus
}

func printDivider() {
	fmt.Println("------")
}
