package config

const (
	EksaGitPasswordTokenEnv   = "EKSA_GIT_PASSWORD"
	EksaGitPrivateKeyTokenEnv = "EKSA_GIT_PRIVATE_KEY"
)

type CliConfig struct {
	GitPassword       string
	GitPrivateKeyFile string
}
