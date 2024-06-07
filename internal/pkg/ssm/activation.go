package ssm

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type ActivationInfo struct {
	ActivationCode string
	ActivationID   string
}

// Tag is an SSM tag.
type Tag struct {
	Key   string
	Value string
}

// CreateActivation creates an SSM Hybrid activation.
func CreateActivation(session *session.Session, instanceName, role string, tags ...Tag) (*ActivationInfo, error) {
	s := ssm.New(session)

	request := ssm.CreateActivationInput{
		DefaultInstanceName: &instanceName,
		Description:         &instanceName,
		IamRole:             &role,
	}

	for _, tag := range tags {
		request.Tags = append(request.Tags,
			&ssm.Tag{Key: aws.String(tag.Key), Value: aws.String(tag.Value)},
		)
	}

	result, err := s.CreateActivation(&request)
	if err != nil {
		return nil, fmt.Errorf("failed to activate ssm instance %s: %v", instanceName, err)
	}

	return &ActivationInfo{ActivationCode: *result.ActivationCode, ActivationID: *result.ActivationId}, nil
}

// DeleteActivations deletes SSM activations.
func DeleteActivations(session *session.Session, ids ...string) ([]*ssm.DeleteActivationOutput, error) {
	s := ssm.New(session)
	var outputs []*ssm.DeleteActivationOutput
	for _, id := range ids {
		request := ssm.DeleteActivationInput{
			ActivationId: &id,
		}

		result, err := s.DeleteActivation(&request)
		if err != nil {
			return nil, fmt.Errorf("failed to delete ssm activation: %v", err)
		}

		outputs = append(outputs, result)
	}

	return outputs, nil
}
