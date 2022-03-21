package aws

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/ecr"

	"github.com/aws/eks-anywhere/pkg/types"
)

func (a *AwsClient) GetEcrCredentials() (*types.DockerCredentials, error) {
	svc := ecr.New(a.session)

	input := &ecr.GetAuthorizationTokenInput{}
	output, err := svc.GetAuthorizationToken(input)
	if err != nil {
		return nil, fmt.Errorf("failed getting ecr authorization token: %v", err)
	}

	authB64 := output.AuthorizationData[0].AuthorizationToken
	auth, err := base64.StdEncoding.DecodeString(*authB64)
	if err != nil {
		return nil, fmt.Errorf("failed decoding authorization token: %v", err)
	}

	authSplit := strings.Split(string(auth), ":")
	creds := &types.DockerCredentials{
		Username: authSplit[0],
		Password: authSplit[1],
	}
	return creds, nil
}
