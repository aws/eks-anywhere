package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/aws/eks-anywhere/pkg/logger"
)

func CreateInstance(session *session.Session, amiId, key, tag, instanceProfileName, subnetId, name string) (string, error) {
	service := ec2.New(session)

	logger.V(2).Info("Creating instance", "name", name)
	result, err := service.RunInstances(&ec2.RunInstancesInput{
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
					{
						Key:   aws.String("Name"),
						Value: aws.String(name),
					},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("error creating instance: %v", err)
	}

	logger.V(2).Info("Waiting until the instance starts running")
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			result.Instances[0].InstanceId,
		},
	}
	err = service.WaitUntilInstanceRunning(input)
	if err != nil {
		return "", fmt.Errorf("error waiting for instance: %v", err)
	}
	logger.V(2).Info("Instance is running")

	return *result.Instances[0].InstanceId, nil
}
