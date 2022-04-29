package v1alpha1

import (
	"errors"
	"fmt"
	"net/url"
)

const (
	FluxConfigKind = "FluxConfig"
)

func GetAndValidateFluxConfig(fileName string, refName string, clusterConfig *Cluster) (*FluxConfig, error) {
	config, err := getFluxConfig(fileName)
	if err != nil {
		return nil, err
	}
	if err = validateFluxConfig(config); err != nil {
		return nil, err
	}
	if err = validateFluxRefName(config, refName); err != nil {
		return nil, err
	}
	if err = validateFluxNamespace(config, clusterConfig); err != nil {
		return nil, err
	}
	return config, nil
}

func getFluxConfig(fileName string) (*FluxConfig, error) {
	var config FluxConfig
	err := ParseClusterConfig(fileName, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func validateFluxConfig(config *FluxConfig) error {
	if config.Spec.Git != nil && config.Spec.Github != nil {
		return errors.New("must specify only one provider")
	}
	if config.Spec.Git == nil && config.Spec.Github == nil {
		return errors.New("must specify a provider. Valid options are git and github")
	}
	if config.Spec.Github != nil {
		err := validateGithubProviderConfig(*config.Spec.Github)
		if err != nil {
			return err
		}
	}
	if config.Spec.Git != nil {
		err := validateGitProviderConfig(*config.Spec.Git)
		if err != nil {
			return err
		}
	}

	if len(config.Spec.Branch) > 0 {
		err := validateGitBranchName(config.Spec.Branch)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateGitProviderConfig(config GitProviderConfig) error {
	if len(config.RepositoryUrl) <= 0 {
		return errors.New("'repositoryUrl' is not set or empty in gitProviderConfig; repositoryUrl is a required field")
	}

	return validateRepositoryUrl(config.RepositoryUrl)
}

func validateGithubProviderConfig(config GithubProviderConfig) error {
	if len(config.Owner) <= 0 {
		return errors.New("'owner' is not set or empty in githubProviderConfig; owner is a required field")
	}
	if len(config.Repository) <= 0 {
		return errors.New("'repository' is not set or empty in githubProviderConfig; repository is a required field")
	}
	err := validateGitRepoName(config.Repository)
	if err != nil {
		return err
	}
	return nil
}

func validateRepositoryUrl(repositoryUrl string) error {
	url, err := url.Parse(repositoryUrl)
	if err != nil {
		return fmt.Errorf("unable to parse repository url: %v", err)
	}
	if url.Scheme != "https" && url.Scheme != "ssh" {
		return fmt.Errorf("invalid repository url scheme: %v", err)
	}
	return nil
}

func validateFluxRefName(config *FluxConfig, refName string) error {
	if config == nil {
		return nil
	}
	if config.Name != refName {
		return fmt.Errorf("FluxConfig retrieved with name %s does not match name (%s) specified in "+
			"gitOpsRef", config.Name, refName)
	}
	return nil
}

func validateFluxNamespace(config *FluxConfig, clusterConfig *Cluster) error {
	if config == nil {
		return nil
	}

	if config.Namespace != clusterConfig.Namespace {
		return errors.New("FluxConfig and Cluster objects must have the same namespace specified")
	}
	return nil
}

func setFluxConfigDefaults(flux *FluxConfig) {
	if flux == nil {
		return
	}

	c := &flux.Spec
	if len(c.SystemNamespace) == 0 {
		c.SystemNamespace = FluxDefaultNamespace
	}

	if len(c.Branch) == 0 {
		c.Branch = FluxDefaultBranch
	}
}
