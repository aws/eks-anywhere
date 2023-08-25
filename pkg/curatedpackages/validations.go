package curatedpackages

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

func ValidateKubeVersion(kubeVersion string, clusterName string) error {
	if len(clusterName) > 0 {
		if len(kubeVersion) > 0 {
			return fmt.Errorf("please specify either kube-version or cluster name not both")
		}
		return nil
	}

	if len(kubeVersion) > 0 {
		versionSplit := strings.Split(kubeVersion, ".")
		if len(versionSplit) != 2 {
			return fmt.Errorf("please specify kube-version as <major>.<minor>")
		}
		return nil
	}
	return fmt.Errorf("please specify kube-version or cluster name")
}

// ValidateAWSCreds checks if the aws credentials has access to pull images from a ECR url.
func ValidateAWSCreds(awsClientID, awsClientSecret, registryURL string) error {
	creds := credentials.NewStaticCredentials(awsClientID, awsClientSecret, "")

	region, err := getRegistryRegion(registryURL)
	if err != nil {
		return err
	}

	s, err := session.NewSession(&aws.Config{Credentials: creds, Region: &region})
	if err != nil {
		return err
	}
	ecrClient := ecr.New(s)
	resp, err := ecrClient.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return err
	}
	authToken := resp.AuthorizationData[0].AuthorizationToken

	return ValidateECRAuthToken(http.DefaultClient, *authToken, registryURL)
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ValidateECRAuthToken checks if Authorization header basic authToken is valid for pull images from a ECR.
func ValidateECRAuthToken(client httpClient, authToken string, registryURL string) error {
	// we can also check "/v2/eks-anywhere-packages/blobs/latest" as well for ecr:GetDownloadUrlForLayer policy
	// but manifest check should be enough for now
	manifestPath := "/v2/eks-anywhere-packages/manifests/latest"

	req, err := http.NewRequest("GET", registryURL+manifestPath, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Basic "+authToken)

	resp2, err := client.Do(req)
	if err != nil {
		return err
	}

	bodyBytes, err := io.ReadAll(resp2.Body)
	if resp2.StatusCode != 200 {
		return fmt.Errorf("%s\n, %v", string(bodyBytes), err)
	}

	return nil
}

func getRegistryRegion(registryURL string) (string, error) {
	// registryURL is in the format of https://<account_id>.dkr.ecr.<region>.amazonaws.com
	// so we can just split on "." and take the 3rd element
	split := strings.Split(registryURL, ".")
	if len(split) < 4 {
		return "", fmt.Errorf("registryURL is invalid")
	}

	return split[3], nil
}
