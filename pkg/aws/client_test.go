package aws_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/aws"
)

type awsTest struct {
	*WithT
	ctx context.Context
}

func newAwsTest(t *testing.T) *awsTest {
	return &awsTest{
		WithT: NewWithT(t),
		ctx:   context.Background(),
	}
}

func TestLoadConfig(t *testing.T) {
	tt := newAwsTest(t)
	_, err := aws.LoadConfig(tt.ctx)
	tt.Expect(err).To(Succeed())
}

func TestLoadConfigSnow(t *testing.T) {
	tt := newAwsTest(t)
	config, err := aws.LoadConfig(tt.ctx, aws.WithSnowEndpointAccess("1.2.3.4", certificatesFile, credentialsFile))
	tt.Expect(err).To(Succeed())
	snowballDeviceEndpoint, err := config.EndpointResolverWithOptions.ResolveEndpoint("Snowball Device", "snow")
	tt.Expect(snowballDeviceEndpoint.URL).To(Equal("https://1.2.3.4:9092"))
	tt.Expect(err).To(Succeed())
	ec2Endpoint, err := config.EndpointResolverWithOptions.ResolveEndpoint("EC2", "snow")
	tt.Expect(ec2Endpoint.URL).To(Equal("https://1.2.3.4:8243"))
	tt.Expect(err).To(Succeed())
}
