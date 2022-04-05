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

func withCustomCABundleFile(certsFile string) config.LoadOptionsFunc {
	return func(o *config.LoadOptions) error {
		caPEM, err := ioutil.ReadFile(certsFile)
		if err != nil {
			return fmt.Errorf("reading aws CA bundles file: %v", err)
		}
		o.CustomCABundle = bytes.NewReader(caPEM)
		return nil
	}
}

func WithSnow(deviceIP string, certsFile string) awsConfigOpts {
	return []func(*config.LoadOptions) error{
		withCustomCABundleFile(certsFile),
		config.WithSharedConfigProfile(deviceIP),
		config.WithEndpointResolverWithOptions(SnowEndpointResolver(deviceIP)),
	}
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
