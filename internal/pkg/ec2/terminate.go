package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TerminateEc2Instances(session *session.Session, instances []*string) error {
	service := ec2.New(session)
	input := &ec2.TerminateInstancesInput{
		InstanceIds: instances,
	}

	_, err := service.TerminateInstances(input)
	if err != nil {
		return fmt.Errorf("terminating EC2 instances: %v", err)
	}

	return nil
}
