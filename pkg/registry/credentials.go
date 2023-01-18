package registry

import (
	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/config/credentials"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// CredentialStore for registry credentials such as ~/.docker/config.json.
type CredentialStore struct {
	directory  string
	configFile *configfile.ConfigFile
}

// NewCredentialStore create a credential store.
func NewCredentialStore() *CredentialStore {
	return &CredentialStore{
		directory: config.Dir(),
	}
}

// SetDirectory override default directory.
func (cs *CredentialStore) SetDirectory(directory string) {
	cs.directory = directory
}

// Init initialize a credential store.
func (cs *CredentialStore) Init() (err error) {
	cs.configFile, err = config.Load(cs.directory)
	if err != nil {
		return err
	}
	if !cs.configFile.ContainsAuth() {
		cs.configFile.CredentialsStore = credentials.DetectDefaultStore(cs.configFile.CredentialsStore)
	}
	return nil
}

// Credential get an authentication credential for a given registry.
func (cs *CredentialStore) Credential(registry string) (auth.Credential, error) {
	authConf, err := cs.configFile.GetCredentialsStore(registry).Get(registry)
	if err != nil {
		return auth.EmptyCredential, err
	}
	cred := auth.Credential{
		Username:     authConf.Username,
		Password:     authConf.Password,
		AccessToken:  authConf.RegistryToken,
		RefreshToken: authConf.IdentityToken,
	}
	return cred, nil
}
