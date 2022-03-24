package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TagInstance(session *session.Session, instanceId, key, value string) error {
	service := ec2.New(session)
	_, err := service.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{&instanceId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String(key),
				Value: aws.String(value),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("tagging instance: %v", err)
	}

	return nil
}
