package config

const (
	EksaGitPasswordTokenEnv   = "EKSA_GIT_PASSWORD"
	EksaGitPrivateKeyTokenEnv = "EKSA_GIT_PRIVATE_KEY"
	EksaGitKnownHostsFileEnv  = "EKSA_GIT_KNOWN_HOSTS"
	SshKnownHostsEnv          = "SSH_KNOWN_HOSTS"
)

type CliConfig struct {
	GitPassword       string
	GitPrivateKeyFile string
	GitKnownHostsFile string
}
