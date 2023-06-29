package git

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/eks-anywhere/pkg/constants"
)

// IsCodeCommitURL check if repo url is code commit url and returns user from url.
func IsCodeCommitURL(repoURL string) (bool, *url.Userinfo, error) {
	parsedRepoURL, err := url.Parse(repoURL)
	if err != nil {
		return false, nil, fmt.Errorf("parsing repository URL %s: %v", repoURL, err)
	}
	if strings.Contains(parsedRepoURL.Hostname(), constants.CodeCommitSubHost) && strings.Contains(parsedRepoURL.Hostname(), constants.AWSSubHost) {
		return true, parsedRepoURL.User, nil
	}
	return false, nil, nil
}
