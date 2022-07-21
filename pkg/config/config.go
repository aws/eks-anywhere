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
	ExternalEtcdTimeoutEnv      = "EKSA_EXTERNAL_ETCD_TIMEOUT"
)

type CliConfig struct {
	GitSshKeyPassphrase string
	GitPrivateKeyFile   string
	GitKnownHostsFile   string
	MaxWaitPerMachine   time.Duration
}

const (
	DefaultMaxWaitPerMachine = 10 * time.Minute
	DefaultEtcdWaitStr       = "60m"
)

func GetMaxWaitPerMachine() time.Duration {
	if env, found := os.LookupEnv(EksaReplicasReadyTimeoutEnv); found {
		if duration, err := time.ParseDuration(env); err == nil {
			return duration
		} else {
			logger.V(3).Info(fmt.Sprintf("Invalid EKSA_REPLICAS_READY_TIMEOUT value: %s Use the default timeout: %s",
				env, DefaultMaxWaitPerMachine.String()))
		}
	}
	return DefaultMaxWaitPerMachine
}

func GetExternalEtcdTimeout() string {
	if env, found := os.LookupEnv(ExternalEtcdTimeoutEnv); found {
		if _, err := time.ParseDuration(env); err == nil {
			return env
		}
		logger.V(3).Info(fmt.Sprintf("Invalid %s value: %s Use the default timeout: %s", ExternalEtcdTimeoutEnv, env, ExternalEtcdTimeoutEnv))
	}
	return DefaultEtcdWaitStr
}
