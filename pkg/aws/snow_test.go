package aws_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/aws"
)

func TestLoadConfigWithSnow(t *testing.T) {
	tests := []struct {
		name          string
		certsFilePath string
		wantErr       string
	}{
		{
			name:          "validate certs",
			certsFilePath: "testdata/valid_certificates",
			wantErr:       "",
		},
		{
			name:          "invalidate certs",
			certsFilePath: "testdata/invalid_certificates",
			wantErr:       "failed to load custom CA bundle PEM file",
		},
		{
			name:          "certs not exists",
			certsFilePath: "testdata/nonexists_certificates",
			wantErr:       "no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newAwsTest(t)
			_, err := aws.LoadConfig(g.ctx, aws.WithSnowEndpointAccess("", tt.certsFilePath, ""))
			if tt.wantErr == "" {
				g.Expect(err).To(Succeed())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}

func TestBuildSnowAwsClientMap(t *testing.T) {
	tests := []struct {
		name          string
		credsFilePath string
		certsFilePath string
		wantErr       string
	}{
		{
			name:          "valid",
			credsFilePath: credentialsFile,
			certsFilePath: certificatesFile,
			wantErr:       "",
		},
		{
			name:          "nonexistent creds",
			credsFilePath: "testdata/nonexistent",
			certsFilePath: certificatesFile,
			wantErr:       "fetching aws credentials from env: file 'testdata/nonexistent' does not exist",
		},
		{
			name:          "nonexistent certs",
			credsFilePath: credentialsFile,
			certsFilePath: "testdata/nonexistent",
			wantErr:       "fetching aws CA bundles from env: file 'testdata/nonexistent' does not exist",
		},
		{
			name:          "invalid ips in creds",
			credsFilePath: "testdata/invalid_credentials_no_ips",
			certsFilePath: certificatesFile,
			wantErr:       "getting device ips from aws credentials: no ip address profile found in content",
		},
		{
			name:          "missing access key in creds",
			credsFilePath: "testdata/invalid_credentials_no_access_key",
			certsFilePath: certificatesFile,
			wantErr:       "setting up aws client: setting aws config:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newAwsTest(t)
			ctx := context.Background()
			t.Setenv(aws.EksaAwsCredentialsFileKey, tt.credsFilePath)
			t.Setenv(aws.EksaAwsCABundlesFileKey, tt.certsFilePath)

			_, err := aws.BuildClients(ctx)
			if tt.wantErr == "" {
				g.Expect(err).To(Succeed())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}
