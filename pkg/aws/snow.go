package aws

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

const (
	snowEC2Port        = 8243
	snowballDevicePort = 9092
)

func BuildClients(ctx context.Context) (Clients, error) {
	credsFile, err := AwsCredentialsFile()
	if err != nil {
		return nil, fmt.Errorf("fetching aws credentials from env: %v", err)
	}
	certsFile, err := AwsCABundlesFile()
	if err != nil {
		return nil, fmt.Errorf("fetching aws CA bundles from env: %v", err)
	}

	deviceIps, err := ParseDeviceIPsFromFile(credsFile)
	if err != nil {
		return nil, fmt.Errorf("getting device ips from aws credentials: %v", err)
	}

	deviceClientMap := make(Clients, len(deviceIps))

	for _, ip := range deviceIps {
		config, err := LoadConfig(ctx, WithSnowEndpointAccess(ip, certsFile, credsFile))
		if err != nil {
			return nil, fmt.Errorf("setting up aws client: %v", err)
		}
		deviceClientMap[ip] = NewClientFromConfig(config)
	}
	return deviceClientMap, nil
}

func snowEndpoints(deviceIP string) []ServiceEndpoint {
	return []ServiceEndpoint{
		{
			ServiceID:     "EC2",
			SigningRegion: "snow",
			URL:           fmt.Sprintf("https://%s:%d", deviceIP, snowEC2Port),
		},
		{
			ServiceID:     "Snowball Device",
			SigningRegion: "snow",
			URL:           fmt.Sprintf("https://%s:%d", deviceIP, snowballDevicePort),
		},
	}
}

// WithCustomCABundleFile is a helper function to construct functional options
// that reads an aws certificates file and sets CustomCABundle on config's LoadOptions.
func WithCustomCABundleFile(certsFile string) AwsConfigOpt {
	return func(opts *config.LoadOptions) error {
		caPEM, err := os.Open(certsFile)
		if err != nil {
			return fmt.Errorf("reading aws certificates file: %w", err)
		}

		customBundleOpt := config.WithCustomCABundle(bufio.NewReader(caPEM))
		if err := customBundleOpt(opts); err != nil {
			return err
		}

		return nil
	}
}

// WithSnowEndpointAccess gathers all the config's LoadOptions for snow,
// which includes snowball ec2 endpoint, snow credentials for a specific profile,
// and CA bundles for accessing the https endpoint.
func WithSnowEndpointAccess(deviceIP string, certsFile, credsFile string) AwsConfigOpt {
	return AwsConfigOptSet(
		WithCustomCABundleFile(certsFile),
		config.WithSharedCredentialsFiles([]string{credsFile}),
		config.WithSharedConfigProfile(deviceIP),
		config.WithEndpointResolverWithOptions(SnowEndpointResolver(deviceIP)),
	)
}

func SnowEndpointResolver(deviceIP string) aws.EndpointResolverWithOptionsFunc {
	return aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		for _, endpoint := range snowEndpoints(deviceIP) {
			if service == endpoint.ServiceID {
				return aws.Endpoint{
					URL:           endpoint.URL,
					SigningRegion: endpoint.SigningRegion,
				}, nil
			}
		}
		// returning EndpointNotFoundError allows the service to fallback to it's default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})
}
