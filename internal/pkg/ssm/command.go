package ssm

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere/pkg/retrier"
)

const ssmLogGroup = "/eks-anywhere/test/e2e"

var initE2EDirCommand = "mkdir -p /home/e2e/bin && cd /home/e2e"

// WaitForSSMReady waits for the SSM command to be ready.
func WaitForSSMReady(session *session.Session, instanceID string, timeout time.Duration) error {
	err := retrier.Retry(10, 20*time.Second, func() error {
		return Run(session, logr.Discard(), instanceID, "ls", timeout)
	})
	if err != nil {
		return fmt.Errorf("waiting for ssm to be ready: %v", err)
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

// Run runs the command using SSM on the instance corresponding to the instanceID.
func Run(session *session.Session, logger logr.Logger, instanceID, command string, timeout time.Duration, opts ...CommandOpt) error {
	o, err := RunCommand(session, logger, instanceID, command, timeout, opts...)
	if err != nil {
		return err
	}
	if !o.Successful() {
		return fmt.Errorf("ssm command returned not successful result %s: %s", *o.commandOut.Status, string(o.StdErr))
	}

	return nil
}

// RunCommand runs the command using SSM on the instance corresponding to the instanceID.
func RunCommand(session *session.Session, logger logr.Logger, instanceID, command string, timeout time.Duration, opts ...CommandOpt) (*RunOutput, error) {
	service := ssm.New(session)

	result, err := sendCommand(service, logger, instanceID, command, timeout, opts...)
	if err != nil {
		return nil, err
	}

	commandIn := &ssm.GetCommandInvocationInput{
		CommandId:  result.Command.CommandId,
		InstanceId: aws.String(instanceID),
	}

	// Make sure ssm send command is registered
	logger.Info("Waiting for ssm command to be registered", "commandId", commandIn.CommandId, "instanceID", commandIn.InstanceId)
	err = retrier.Retry(10, 5*time.Second, func() error {
		_, err := service.GetCommandInvocation(commandIn)
		if err != nil {
			return fmt.Errorf("getting ssm command invocation: %v", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("waiting for ssm command to be registered: %v", err)
	}

	logger.Info("Waiting for ssm command to finish")
	var commandOut *ssm.GetCommandInvocationOutput

	// Making the retrier wait for longer than the provided SSM timeout to make sure
	// we always get the output results.
	r := retrier.New(timeout+5*time.Minute, retrier.WithMaxRetries(math.MaxInt, 60*time.Second))
	err = r.Retry(func() error {
		var err error
		commandOut, err = service.GetCommandInvocation(commandIn)
		if err != nil {
			return err
		}

		status := *commandOut.Status
		if isFinalStatus(status) {
			logger.V(4).Info("SSM command finished", "status", status, "commandId", *result.Command.CommandId)
			return nil
		}

		return fmt.Errorf("command still running with status %s", status)
	})
	if err != nil {
		return nil, fmt.Errorf("retries exhausted running ssm command: %v", err)
	}

	return buildRunOutput(commandOut), nil
}

func sendCommand(service *ssm.SSM, logger logr.Logger, instanceID, command string, timeout time.Duration, opts ...CommandOpt) (*ssm.SendCommandOutput, error) {
	in := &ssm.SendCommandInput{
		DocumentName: aws.String("AWS-RunShellScript"),
		InstanceIds:  []*string{aws.String(instanceID)},
		Parameters:   map[string][]*string{"commands": {aws.String(initE2EDirCommand), aws.String(command)}, "executionTimeout": {aws.String(strconv.FormatFloat(timeout.Seconds(), 'f', 0, 64))}},
	}

	for _, opt := range opts {
		opt(in)
	}

	var result *ssm.SendCommandOutput
	r := retrier.New(timeout, retrier.WithRetryPolicy(func(totalRetries int, err error) (retry bool, wait time.Duration) {
		if request.IsErrorThrottle(err) && totalRetries < 60 {
			return true, 60 * time.Second
		}
		return false, 0
	}))
	err := r.Retry(func() error {
		var err error
		logger.V(4).Info("Running ssm command", "cmd", command)
		result, err = service.SendCommand(in)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("retries exhausted sending ssm command: %v", err)
	}

	logger.V(4).Info("SSM command started", "commandId", result.Command.CommandId)
	if in.OutputS3BucketName != nil {
		logger.V(4).Info(
			"SSM command output to S3", "url",
			fmt.Sprintf("s3://%s/%s/%s/%s/awsrunShellScript/0.awsrunShellScript/stderr", *in.OutputS3BucketName, *in.OutputS3KeyPrefix, *result.Command.CommandId, instanceID),
		)
	}

	return result, nil
}

func isFinalStatus(status string) bool {
	_, nonFinalStatus := nonFinalStatuses[status]
	return !nonFinalStatus
}
