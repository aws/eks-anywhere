package aws_test

import (
	"os"
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/aws"
)

const (
	credentialsFile  = "testdata/valid_credentials"
	certificatesFile = "testdata/valid_certificates"
)

func TestAwsCredentialsFile(t *testing.T) {
	tt := newAwsTest(t)
	t.Setenv(aws.EksaAwsCredentialsFileKey, credentialsFile)
	_, err := aws.AwsCredentialsFile()
	tt.Expect(err).To(Succeed())
}

func TestAwsCredentialsFileEnvNotSet(t *testing.T) {
	tt := newAwsTest(t)
	os.Unsetenv(aws.EksaAwsCredentialsFileKey)
	_, err := aws.AwsCredentialsFile()
	tt.Expect(err).To((MatchError(ContainSubstring("env 'EKSA_AWS_CREDENTIALS_FILE' is not set or is empty"))))
}

func TestAwsCredentialsFileNotExists(t *testing.T) {
	tt := newAwsTest(t)
	t.Setenv(aws.EksaAwsCredentialsFileKey, "testdata/not_exists_credentials")
	_, err := aws.AwsCredentialsFile()
	tt.Expect(err).To((MatchError(ContainSubstring("file 'testdata/not_exists_credentials' does not exist"))))
}

func TestAwsCABundlesFile(t *testing.T) {
	tt := newAwsTest(t)
	t.Setenv(aws.EksaAwsCABundlesFileKey, certificatesFile)
	_, err := aws.AwsCABundlesFile()
	tt.Expect(err).To(Succeed())
}

func TestAwsCABundlesFileEnvNotSet(t *testing.T) {
	tt := newAwsTest(t)
	os.Unsetenv(aws.EksaAwsCABundlesFileKey)
	_, err := aws.AwsCABundlesFile()
	tt.Expect(err).To((MatchError(ContainSubstring("env 'EKSA_AWS_CA_BUNDLES_FILE' is not set or is empty"))))
}

func TestAwsCABundlesFileNotExists(t *testing.T) {
	tt := newAwsTest(t)
	t.Setenv(aws.EksaAwsCABundlesFileKey, "testdata/not_exists_certificates")
	_, err := aws.AwsCABundlesFile()
	tt.Expect(err).To((MatchError(ContainSubstring("file 'testdata/not_exists_certificates' does not exist"))))
}

func TestEncodeFileFromEnv(t *testing.T) {
	tt := newAwsTest(t)
	t.Setenv(aws.EksaAwsCredentialsFileKey, credentialsFile)
	strB64, err := aws.EncodeFileFromEnv(aws.EksaAwsCredentialsFileKey)
	tt.Expect(err).To(Succeed())
	tt.Expect(strB64).To(Equal("WzEuMi4zLjRdCmF3c19hY2Nlc3Nfa2V5X2lkID0gQUJDREVGR0hJSktMTU5PUFFSMlQKYXdzX3NlY3JldF9hY2Nlc3Nfa2V5ID0gQWZTRDdzWXovVEJadHprUmVCbDZQdXVJU3pKMld0TmtlZVB3K25OekoKcmVnaW9uID0gc25vdwoKWzEuMi4zLjVdCmF3c19hY2Nlc3Nfa2V5X2lkID0gQUJDREVGR0hJSktMTU5PUFFSMlQKYXdzX3NlY3JldF9hY2Nlc3Nfa2V5ID0gQWZTRDdzWXovVEJadHprUmVCbDZQdXVJU3pKMld0TmtlZVB3K25OekoKcmVnaW9uID0gc25vdw=="))
}

func TestParseDeviceIPsFromFile(t *testing.T) {
	tests := []struct {
		name    string
		creds   string
		want    []string
		wantErr string
	}{
		{
			name: "validate creds",
			creds: `[1.2.3.4]
aws_access_key_id = ABCDEFGHIJKLMNOPQR2T
aws_secret_access_key = AfSD7sYz/TBZtzkReBl6PuuISzJ2WtNkeePw+nNzJ
region = snow

[1.2.3.5]
aws_access_key_id = ABCDEFGHIJKLMNOPQR2T
aws_secret_access_key = AfSD7sYz/TBZtzkReBl6PuuISzJ2WtNkeePw+nNzJ
region = snow`,
			want: []string{
				"1.2.3.4",
				"1.2.3.5",
			},
			wantErr: "",
		},
		{
			name: "no ip in profile",
			creds: `[invalid profile]
aws_access_key_id = ABCDEFGHIJKLMNOPQR2T
aws_secret_access_key = AfSD7sYz/TBZtzkReBl6PuuISzJ2WtNkeePw+nNzJ
region = snow`,
			wantErr: "no ip address profile found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newAwsTest(t)

			got, err := aws.ParseDeviceIPs(strings.NewReader(tt.creds))

			if tt.wantErr == "" {
				g.Expect(err).To(Succeed())
				g.Expect(got).To(Equal(tt.want))
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}
