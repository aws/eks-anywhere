package ssm

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type ActivationInfo struct {
	ActivationCode string
	ActivationID   string
}

func CreateActivation(session *session.Session, instanceName, role string) (*ActivationInfo, error) {
	s := ssm.New(session)

	request := ssm.CreateActivationInput{
		DefaultInstanceName: &instanceName,
		Description:         &instanceName,
		IamRole:             &role,
	}

	result, err := s.CreateActivation(&request)
	if err != nil {
		return nil, fmt.Errorf("failed to activate ssm instance %s: %v", instanceName, err)
	}

	return &ActivationInfo{ActivationCode: *result.ActivationCode, ActivationID: *result.ActivationId}, nil
}

func DeleteActivation(session *session.Session, activationId string) (*ssm.DeleteActivationOutput, error) {
	s := ssm.New(session)

	request := ssm.DeleteActivationInput{
		ActivationId: &activationId,
	}

	result, err := s.DeleteActivation(&request)
	if err != nil {
		return nil, fmt.Errorf("failed to delete ssm activation: %v", err)
	}

	return result, nil
}
