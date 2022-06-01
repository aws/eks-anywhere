package ssm

import (
	"fmt"

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

func DeregisterInstance(session *session.Session, id string) (*ssm.DeregisterManagedInstanceOutput, error) {
	s := ssm.New(session)
	input := ssm.DeregisterManagedInstanceInput{
		InstanceId: &id,
	}

	output, err := s.DeregisterManagedInstance(&input)
	if err != nil {
		return nil, fmt.Errorf("failed to deregister ssm instance %s: %v", id, err)
	}

	return output, nil
}
