package reconciler

import (
	"bytes"
	"context"
	"net"

	awstypes "github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awscredentials "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/aws"
	"github.com/aws/eks-anywhere/pkg/providers/snow"
)

type AwsClientBuilder struct {
	client client.Client
}

func NewAwsClientBuilder(client client.Client) *AwsClientBuilder {
	return &AwsClientBuilder{
		client: client,
	}
}

func (b *AwsClientBuilder) Get(ctx context.Context) (snow.AwsClientMap, error) {
	// Setting the aws client map in validator on every reconcile based on the secrets at that point of time
	credentials, certificates, err := getSnowCredentials(ctx, b.client)
	if err != nil {
		return nil, errors.Wrap(err, "getting snow credentials")
	}

	clients, err := createAwsClients(ctx, credentials, certificates)
	if err != nil {
		return nil, err
	}
	return snow.NewAwsClientMap(clients), nil
}

type credentialConfiguration struct {
	AccessKey string `ini:"aws_access_key_id"`
	SecretKey string `ini:"aws_secret_access_key"`
	Region    string `ini:"region"`
}

func createAwsClients(ctx context.Context, credentials []byte, certificates []byte) (aws.Clients, error) {
	var deviceIps []string
	credsCfg, err := ini.Load(credentials)
	if err != nil {
		return nil, errors.Wrap(err, "loading values from credentials")
	}
	for _, ip := range credsCfg.SectionStrings() {
		if net.ParseIP(ip) != nil {
			deviceIps = append(deviceIps, ip)
		}
	}

	deviceClientMap := make(aws.Clients, len(deviceIps))
	for _, ip := range deviceIps {
		ipCfg, err := parseIpConfiguration(credsCfg, ip)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing configuration for %v", ip)
		}
		clientCfg, err := awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithCustomCABundle(bytes.NewReader(certificates)),
			awsconfig.WithRegion(ipCfg.Region),
			awsconfig.WithCredentialsProvider(awscredentials.StaticCredentialsProvider{
				Value: awstypes.Credentials{
					AccessKeyID:     ipCfg.AccessKey,
					SecretAccessKey: ipCfg.SecretKey,
				},
			}),
			awsconfig.WithEndpointResolverWithOptions(aws.SnowEndpointResolver(ip)),
		)
		if err != nil {
			return nil, errors.Wrap(err, "setting up aws client")
		}
		deviceClientMap[ip] = aws.NewClientFromConfig(clientCfg)
	}

	return deviceClientMap, nil
}

func parseIpConfiguration(credsCfg *ini.File, ip string) (*credentialConfiguration, error) {
	var config credentialConfiguration
	err := credsCfg.Section(ip).StrictMapTo(&config)
	if err != nil {
		return nil, err
	}
	if len(config.AccessKey) == 0 {
		return nil, errors.New("unable to set aws_access_key_id")
	}
	if len(config.SecretKey) == 0 {
		return nil, errors.New("unable to set aws_secret_access_key")
	}
	if len(config.Region) == 0 {
		return nil, errors.New("unable to set region")
	}
	return &config, nil
}
