package ec2

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/aws/eks-anywhere/pkg/logger"
)

func ListInstances(session *session.Session, key string, value string, maxAge float64) ([]*string, error) {
	service := ec2.New(session)
	var instanceList []*string

	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:" + key),
				Values: []*string{
					aws.String(value),
				},
			},
			{
				Name: aws.String("instance-state-name"),
				Values: []*string{
					aws.String("running"),
					aws.String("pending"),
					aws.String("stopping"),
					aws.String("stopped"),
				},
			},
		},
	}

	for {
		result, err := service.DescribeInstances(input)
		if err != nil {
			return nil, fmt.Errorf("describing EC2 instances: %v", err)
		}
		reservations := result.Reservations
		for _, reservation := range reservations {
			instances := reservation.Instances

			for _, instance := range instances {
				instanceAge := time.Since(*(instance.LaunchTime)).Seconds()
				logger.V(4).Info("EC2", "instance_id", *(instance.InstanceId), "age_seconds", instanceAge)
				failed := true
				for _, tag := range instance.Tags {
					if *(tag.Key) == "Integration-Test-Done" && *(tag.Value) == "TRUE" {
						logger.V(4).Info("Tests ran successfully on this instance")
						failed = false
					}
				}
				if instanceAge > maxAge || !failed {
					logger.V(4).Info("Adding the instance in the deletion list")
					instanceList = append(instanceList, instance.InstanceId)
				} else {
					logger.V(4).Info("Not adding the instance in the deletion list")
				}
			}
		}
		if result.NextToken != nil {
			input.NextToken = result.NextToken
		} else {
			break
		}
	}
	return instanceList, nil
}
