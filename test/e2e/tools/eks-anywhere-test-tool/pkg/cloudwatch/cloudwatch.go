package cloudwatch

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"

	"github.com/aws/eks-anywhere-test-tool/pkg/awsprofiles"
	"github.com/aws/eks-anywhere-test-tool/pkg/constants"
	"github.com/aws/eks-anywhere-test-tool/pkg/logger"
)

type Cloudwatch struct {
	session *session.Session
	svc     *cloudwatchlogs.CloudWatchLogs
}

func New(account awsprofiles.EksAccount) (*Cloudwatch, error) {
	logger.V(2).Info("creating Cloudwatch client")
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: account.ProfileName(),
		Config:  aws.Config{Region: aws.String(constants.AwsAccountRegion)},
	})
	if err != nil {
		fmt.Printf("Got error when setting up session: %v", err)
		os.Exit(1)
	}

	svc := cloudwatchlogs.New(sess)

	logger.V(2).Info("created Cloudwatch client")
	return &Cloudwatch{
		session: sess,
		svc:     svc,
	}, nil
}

func (c *Cloudwatch) GetLogs(logGroupName string, logStreamName string) ([]*cloudwatchlogs.OutputLogEvent, error) {
	var nextToken *string
	var output []*cloudwatchlogs.OutputLogEvent

	for {
		l, err := c.getLogs(logGroupName, logStreamName, nextToken)
		if err != nil {
			logger.Info("error fetching cloudwatch logs", "group", logGroupName, "stream", logStreamName, "err", err)
			return nil, err
		}
		if l.NextForwardToken == nil || nextToken != nil && *nextToken == *l.NextForwardToken {
			logger.Info("finished fetching logs", "logGroup", logGroupName, "logStream", logStreamName)
			logger.V(3).Info("token comparison", "nextToken", nextToken, "nextForwardToken", l.NextForwardToken)
			break
		}
		nextToken = l.NextForwardToken
		logger.Info("fetched logs", "logGroup", logGroupName, "logStream", logStreamName, "events", len(l.Events))
		logger.V(3).Info("token comparison", "nextToken", nextToken, "nextForwardToken", l.NextForwardToken)
		output = append(output, l.Events...)
	}
	return output, nil
}

func (c Cloudwatch) getLogs(logGroupName string, logStreamName string, nextToken *string) (*cloudwatchlogs.GetLogEventsOutput, error) {
	return c.svc.GetLogEvents(&cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
		NextToken:     nextToken,
	})
}
