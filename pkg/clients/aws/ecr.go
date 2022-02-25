package aws

import (
	"encoding/base64"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/service/ecr"

	"github.com/aws/eks-anywhere/pkg/constants"
)

func (a *AwsClient) GetEcrLoginPassword() (string, error) {
	svc := ecr.New(a.session)

	input := &ecr.GetAuthorizationTokenInput{}
	output, err := svc.GetAuthorizationToken(input)
	if err != nil {
		return "", fmt.Errorf("failed getting ecr authorization token: %v", err)
	}

	authB64 := output.AuthorizationData[0].AuthorizationToken
	auth, err := base64.StdEncoding.DecodeString(*authB64)
	if err != nil {
		return "", fmt.Errorf("failed decoding authorization token: %v", err)
	}

	re := regexp.MustCompile(`^` + constants.EcrRegistryUserName + `:`)
	return re.ReplaceAllString(string(auth), ""), nil
}
