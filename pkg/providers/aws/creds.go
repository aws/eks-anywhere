package aws

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	"github.com/aws/eks-anywhere/pkg/templater"
)

//go:embed config/aws_user_config.yaml
var awsUserConfig string

func (p *provider) setupAWSCredentials(ctx context.Context) error {
	configFile, err := p.createIAMConfigFile()
	if err != nil {
		return err
	}

	envMap, err := p.EnvMap()
	if err != nil {
		return err
	}

	err = p.providerClient.BootstrapIam(ctx, envMap, configFile)
	if err != nil {
		return fmt.Errorf("error bootstrapping iam: %v", err)
	}
	bootstrapCredentials, err := p.providerClient.BootstrapCreds(ctx, envMap)
	if err != nil {
		return fmt.Errorf("error bootstrapping aws creds")
	}
	os.Setenv("AWS_B64ENCODED_CREDENTIALS", bootstrapCredentials)
	return nil
}

func (p *provider) createIAMConfigFile() (filePath string, err error) {
	t := templater.New(p.writer)
	values := map[string]string{
		"clusterName": p.clusterName,
	}

	fileName := fmt.Sprintf("eksa-%s-iam.config", p.clusterName)
	configFile, err := t.WriteToFile(awsUserConfig, values, fileName)
	if err != nil {
		return "", fmt.Errorf("error creating aws iam user config file: %v", err)
	}

	return configFile, nil
}
