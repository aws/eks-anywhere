package config

import "time"

const (
	EksaGitPassphraseTokenEnv = "EKSA_GIT_SSH_KEY_PASSPHRASE"
	EksaGitPrivateKeyTokenEnv = "EKSA_GIT_PRIVATE_KEY"
	EksaGitKnownHostsFileEnv  = "EKSA_GIT_KNOWN_HOSTS"
	SshKnownHostsEnv          = "SSH_KNOWN_HOSTS"
	EksaAccessKeyIdEnv        = "EKSA_AWS_ACCESS_KEY_ID"
	EksaSecretAccessKeyEnv    = "EKSA_AWS_SECRET_ACCESS_KEY"
	AwsAccessKeyIdEnv         = "AWS_ACCESS_KEY_ID"
	AwsSecretAccessKeyEnv     = "AWS_SECRET_ACCESS_KEY"
	EksaAwsConfigFileEnv      = "EKSA_AWS_CONFIG_FILE"
	EksaRegionEnv             = "EKSA_AWS_REGION"
)

type CliConfig struct {
	GitSshKeyPassphrase string
	GitPrivateKeyFile   string
	GitKnownHostsFile   string
}

// CreateClusterCLIConfig is the config we use for create cluster specific configurations.
type CreateClusterCLIConfig struct {
	SkipCPIPCheck           bool
	NodeStartupTimeout      time.Duration
	UnhealthyMachineTimeout time.Duration
}

// UpgradeClusterCLIConfig is the config we use for create cluster specific configurations.
type UpgradeClusterCLIConfig struct {
	NodeStartupTimeout      time.Duration
	UnhealthyMachineTimeout time.Duration
}

// DeleteClusterCLIConfig is the config we use for delete cluster specific configurations.
type DeleteClusterCLIConfig struct {
	ClusterNamespace string
}
