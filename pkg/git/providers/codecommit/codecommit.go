package codecommit

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const (
	// DefaultSSHAuthUser for client auth.
	DefaultSSHAuthUser = "git"
	codeCommitSubHost  = "git-codecommit"
	awsSubHost         = "amazonaws.com"
)

// IsCodeCommitURL check if repo url is code commit url and returns user from url.
func IsCodeCommitURL(repoURL string) (string, error) {
	parsedRepoURL, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("parsing repository URL %s: %v", repoURL, err)
	}
	if strings.Contains(parsedRepoURL.Hostname(), codeCommitSubHost) && strings.Contains(parsedRepoURL.Hostname(), awsSubHost) {
		codeCommitRegex := regexp.MustCompile(`^ssh:\/\/[a-zA-Z0-9]+@git-codecommit\.[a-z0-9-]+\.amazonaws\.com\/v1\/repos\/[a-zA-Z0-9\\\-_\.]+$`)
		if !codeCommitRegex.MatchString(repoURL) {
			return "", fmt.Errorf("invalid AWS CodeCommit url: url should be in format ssh://<SSH-Key-ID>@git-codecommit.<region>.amazonaws.com/v1/repos/<repository>")
		}
		if parsedRepoURL.User.Username() == DefaultSSHAuthUser {
			return "", fmt.Errorf("invalid AWS CodeCommit url: ssh username should be the SSH key ID for the provided private key")
		}
		return parsedRepoURL.User.Username(), nil
	}
	return "", nil
}
