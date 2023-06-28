package config

const (
	EksaGitPassphraseTokenEnv = "EKSA_GIT_SSH_KEY_PASSPHRASE"
	EksaGitPrivateKeyTokenEnv = "EKSA_GIT_PRIVATE_KEY"
	EksaGitKnownHostsFileEnv  = "EKSA_GIT_KNOWN_HOSTS"
	SshKnownHostsEnv          = "SSH_KNOWN_HOSTS"
	EksaAccessKeyIdEnv        = "EKSA_AWS_ACCESS_KEY_ID"
	EksaSecretAccessKeyEnv    = "EKSA_AWS_SECRET_ACCESS_KEY"
	AwsAccessKeyIdEnv         = "AWS_ACCESS_KEY_ID"
	AwsSecretAccessKeyEnv     = "AWS_SECRET_ACCESS_KEY"
	EksaRegionEnv             = "EKSA_AWS_REGION"
	DefaultSSHAuthUser        = "git"
)

type CliConfig struct {
	GitSshKeyPassphrase string
	GitPrivateKeyFile   string
	GitKnownHostsFile   string
}
