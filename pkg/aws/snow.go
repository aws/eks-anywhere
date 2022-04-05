package aws

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

const (
	snowEC2Port = 8243
)

func snowEndpoint(deviceIP string) *ServiceEndpoint {
	return &ServiceEndpoint{
		ServiceID:     "EC2",
		SigningRegion: "snow",
		URL:           fmt.Sprintf("https://%s:%d", deviceIP, snowEC2Port),
	}
}

func WithCustomCABundleFile(certsFile string) config.LoadOptionsFunc {
	caPEM, err := ioutil.ReadFile(certsFile)
	if err != nil {
		return func(*config.LoadOptions) error {
			return fmt.Errorf("reading aws certificates file: %v", err)
		}
	}

	return config.WithCustomCABundle(bytes.NewReader(caPEM))
}

func WithSnow(deviceIP string, certsFile, credsFile string) AwsConfigOpt {
	return AwsConfigOptSet(
		WithCustomCABundleFile(certsFile),
		config.WithSharedCredentialsFiles([]string{credsFile}),
		config.WithSharedConfigProfile(deviceIP),
		config.WithEndpointResolverWithOptions(SnowEndpointResolver(deviceIP)),
	)
}

func SnowEndpointResolver(deviceIP string) aws.EndpointResolverWithOptionsFunc {
	endpoint := snowEndpoint(deviceIP)
	return aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == endpoint.ServiceID {
			return aws.Endpoint{
				URL:           endpoint.URL,
				SigningRegion: endpoint.SigningRegion,
			}, nil
		}
		// returning EndpointNotFoundError allows the service to fallback to it's default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})
}
