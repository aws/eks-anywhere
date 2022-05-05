package e2e

import (
	"fmt"
	"os"
	"regexp"

	"github.com/aws/eks-anywhere/internal/pkg/ssm"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/logger"
	e2etests "github.com/aws/eks-anywhere/test/framework"
)

type s3Files struct {
	key, dstPath string
	permission   int
}

func (e *E2ESession) setupFluxGitEnv(testRegex string) error {
	re := regexp.MustCompile(`^.*GitFlux.*$`)
	if !re.MatchString(testRegex) {
		logger.V(2).Info("Not running Flux Generic Git Provider tests, skipping environment setup")
		return nil
	}

	requiredEnvVars := e2etests.RequiredFluxGitEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}

	for _, file := range buildFluxGitFiles(e.testEnvVars) {
		if err := e.downloadFileInInstance(file); err != nil {
			return fmt.Errorf("downloading flux-git file to instance: %v", err)
		}
	}

	return nil
}

func buildFluxGitFiles(envVars map[string]string) []s3Files {
	return []s3Files{
		{
			key:        "git-flux/known_hosts",
			dstPath:    envVars[config.EksaGitKnownHostsFileEnv],
			permission: 600,
		},
		{
			key:        "git-flux/private-key",
			dstPath:    envVars[config.EksaGitPrivateKeyTokenEnv],
			permission: 600,
		},
	}
}

func (e *E2ESession) downloadFileInInstance(file s3Files) error {
	logger.V(1).Info("Downloading from s3 in instance", "file", file.key)

	command := fmt.Sprintf("aws s3 cp s3://%s/%s %s && chmod %d %[3]s", e.storageBucket, file.key, file.dstPath, file.permission)
	if err := ssm.Run(e.session, e.instanceId, command); err != nil {
		return fmt.Errorf("downloading file in instance: %v", err)
	}
	logger.V(1).Info("Successfully downloaded file", "file", file.key)

	return nil
}
