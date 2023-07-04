package codecommit

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestIsCodeCommitURL(t *testing.T) {
	tests := []struct {
		testName     string
		url          string
		wantErr      string
		isCodeCommit bool
	}{
		{
			"github url",
			"ssh://git@github.com/jeff/test-repo.git",
			"",
			false,
		},
		{
			"valid CodeCommit url",
			"ssh://TESTSSHKEYID@git-codecommit.us-west-1.amazonaws.com/v1/repos/test-repo",
			"",
			true,
		},
		{
			"invalid url - empty ssh key id",
			"://git@git-codecommit.us-west-1.amazonaws.com/v1/repos/test-repo",
			"parsing repository URL",
			false,
		},
		{
			"invalid url - no ssh key id",
			"ssh://git-codecommit.us-west-1.amazonaws.com/v1/repos/test-repo",
			"invalid AWS CodeCommit url: url should be in format ssh://<SSH-Key-ID>@git-codecommit.<region>.amazonaws.com/v1/repos/<repository>",
			false,
		},
		{
			"invalid url - wrong host",
			"ssh://TESTSSHKEYID@git-codecommits.us-west-1.amazonaws.com/v1/repos/test-repo",
			"invalid AWS CodeCommit url: url should be in format ssh://<SSH-Key-ID>@git-codecommit.<region>.amazonaws.com/v1/repos/<repository>",
			false,
		},
		{
			"invalid url - invalid user",
			"ssh://git@git-codecommit.us-west-1.amazonaws.com/v1/repos/test-repo",
			"invalid AWS CodeCommit url: ssh username should be the SSH key ID for the provided private key",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			user, err := IsCodeCommitURL(tt.url)
			if err == nil {
				g.Expect(user != "").To(Equal(tt.isCodeCommit))
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}
