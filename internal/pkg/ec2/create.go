package ec2

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

func CreateInstance(session *session.Session, amiId, key, tag, instanceProfileName, subnetId, name string) (string, error) {
	r := retrier.New(180*time.Minute, retrier.WithRetryPolicy(func(totalRetries int, err error) (retry bool, wait time.Duration) {
		// EC2 Request token bucket has a refill rate of 2 request tokens
		// per second, so 10 seconds wait per retry should be sufficient
		if request.IsErrorThrottle(err) && totalRetries < 50 {
			return true, 10 * time.Second
		}
		return false, 0
	}))

	service := ec2.New(session)
	var result *ec2.Reservation

	err := r.Retry(func() error {
		var err error
		logger.V(2).Info("Creating instance", "name", name)
		result, err = service.RunInstances(&ec2.RunInstancesInput{
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
			return fmt.Errorf("error creating instance: %v", err)
		}

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("retries exhausted when trying to create instances: %v", err)
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
