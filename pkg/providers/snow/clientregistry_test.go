package snow_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/aws"
	"github.com/aws/eks-anywhere/pkg/providers/snow"
)

func TestGetSnowAwsClientMapSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	clientBuilder := snow.NewAwsClientRegistry()
	t.Setenv(aws.EksaAwsCredentialsFileKey, credsFilePath)
	t.Setenv(aws.EksaAwsCABundlesFileKey, certsFilePath)

	err := clientBuilder.Build(ctx)
	g.Expect(err).To(Succeed())

	clientMap, err := clientBuilder.Get(ctx)
	g.Expect(err).To(Succeed())
	g.Expect(clientMap).NotTo(BeNil())
}

func TestBuildSnowAwsClientMapFailure(t *testing.T) {
	g := NewWithT(t)
	t.Setenv(credsFileEnvVar, "")
	ctx := context.Background()
	clientBuilder := snow.NewAwsClientRegistry()

	err := clientBuilder.Build(ctx)
	g.Expect(err).To(MatchError(ContainSubstring("fetching aws credentials from env")))
}

func TestGetSnowAwsClientMapFailure(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	clientBuilder := snow.NewAwsClientRegistry()

	_, err := clientBuilder.Get(ctx)
	g.Expect(err).To(MatchError(ContainSubstring("aws clients for snow not initialized")))
}
