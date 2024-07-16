package ssm

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func GetInstanceByActivationId(session *session.Session, id string) (*ssm.InstanceInformation, error) {
	s := ssm.New(session)
	instanceActivationIdKey := "ActivationIds"
	input := ssm.DescribeInstanceInformationInput{
		Filters: []*ssm.InstanceInformationStringFilter{
			{Key: &instanceActivationIdKey, Values: []*string{&id}},
		},
	}

	output, err := s.DescribeInstanceInformation(&input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe ssm instance %s: %v", id, err)
	}

	infoList := output.InstanceInformationList

	if len(infoList) == 0 {
		return nil, fmt.Errorf("no ssm instance with name %s: %v", id, err)
	}

	return infoList[0], nil
}

// DeregisterInstances deregisters SSM instances.
func DeregisterInstances(session *session.Session, ids ...string) ([]*ssm.DeregisterManagedInstanceOutput, error) {
	s := ssm.New(session)
	var outputs []*ssm.DeregisterManagedInstanceOutput
	for _, id := range ids {
		input := ssm.DeregisterManagedInstanceInput{
			InstanceId: &id,
		}

		output, err := s.DeregisterManagedInstance(&input)
		if err != nil {
			return nil, fmt.Errorf("failed to deregister ssm instance %s: %v", id, err)
		}

		outputs = append(outputs, output)
	}

	return outputs, nil
}

func ListInstancesByTags(ctx context.Context, session *session.Session, tags ...Tag) ([]*ssm.InstanceInformation, error) {
	s := ssm.New(session)
	input := ssm.DescribeInstanceInformationInput{
		Filters: make([]*ssm.InstanceInformationStringFilter, 0, len(tags)),
	}

	for _, tag := range tags {
		input.Filters = append(input.Filters, &ssm.InstanceInformationStringFilter{
			Key:    aws.String("tag:" + tag.Key),
			Values: aws.StringSlice([]string{tag.Value}),
		})
	}

	output, err := s.DescribeInstanceInformation(&input)
	if err != nil {
		return nil, fmt.Errorf("listing ssm instances by tags: %v", err)
	}

	return output.InstanceInformationList, nil
}
