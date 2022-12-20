package validations

import (
	"errors"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
)

func ValidateAuthenticationForGitProvider(clusterSpec *cluster.Spec, cliConfig *config.CliConfig) error {
	if clusterSpec.FluxConfig == nil || clusterSpec.FluxConfig.Spec.Git == nil {
		return nil
	}

	if cliConfig == nil {
		return nil
	}

	if cliConfig.GitPrivateKeyFile == "" {
		return errors.New("provide a path to a private key file via the EKSA_GIT_PRIVATE_KEY in order to use the generic git Flux provider")
	}

	if !FileExistsAndIsNotEmpty(cliConfig.GitPrivateKeyFile) {
		return fmt.Errorf("private key file does not exist at %s or is empty", cliConfig.GitPrivateKeyFile)
	}

	if cliConfig.GitKnownHostsFile == "" {
		return errors.New("provide a path to an SSH Known Hosts file which contains a valid entry associate with the given private key via the EKSA_GIT_SSH_KNOWN_HOSTS environment variable")
	}

	if !FileExistsAndIsNotEmpty(cliConfig.GitKnownHostsFile) {
		return fmt.Errorf("SSH known hosts file does not exist at %v or is empty", cliConfig.GitKnownHostsFile)
	}

	return nil
}
