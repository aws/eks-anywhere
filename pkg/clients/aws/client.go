package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
)

type AwsClient struct {
	session *session.Session
}

func NewClient() (*AwsClient, error) {
	session, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("error creating aws session: %v", err)
	}
	a := &AwsClient{
		session: session,
	}
	return a, nil
}
