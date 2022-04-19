package aws_test

import (
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
			_, err := aws.LoadConfig(g.ctx, aws.WithSnowEndpointAccess("device-ip", tt.certsFilePath, ""))
			if tt.wantErr == "" {
				g.Expect(err).To(Succeed())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}
