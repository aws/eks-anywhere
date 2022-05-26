package config

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
}
