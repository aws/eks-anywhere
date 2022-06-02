package config

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	EksaGitPassphraseTokenEnv   = "EKSA_GIT_SSH_KEY_PASSPHRASE"
	EksaGitPrivateKeyTokenEnv   = "EKSA_GIT_PRIVATE_KEY"
	EksaGitKnownHostsFileEnv    = "EKSA_GIT_KNOWN_HOSTS"
	SshKnownHostsEnv            = "SSH_KNOWN_HOSTS"
	EksaReplicasReadyTimeoutEnv = "EKSA_REPLICAS_READY_TIMEOUT"
)

type CliConfig struct {
	GitSshKeyPassphrase string
	GitPrivateKeyFile   string
	GitKnownHostsFile   string
	MaxWaitPerMachine   time.Duration
}

const defaultMaxWaitPerMachine = 10 * time.Minute

func GetMaxWaitPerMachine() time.Duration {
	if env, found := os.LookupEnv(EksaReplicasReadyTimeoutEnv); found {
		if duration, err := time.ParseDuration(env); err == nil {
			return duration
		} else {
			logger.V(3).Info(fmt.Sprintf("Invalid EKSA_REPLICAS_READY_TIMEOUT value: %s Use the default timeout: %s",
				env, defaultMaxWaitPerMachine.String()))
		}
	}
	return defaultMaxWaitPerMachine
}
