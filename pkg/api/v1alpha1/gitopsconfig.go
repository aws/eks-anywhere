package v1alpha1

import (
	"errors"
	"fmt"
	"regexp"
)

const (
	GitOpsConfigKind     = "GitOpsConfig"
	FluxDefaultNamespace = "flux-system"
	FluxDefaultBranch    = "main"
)

func GetAndValidateGitOpsConfig(fileName string, refName string, clusterConfig *Cluster) (*GitOpsConfig, error) {
	config, err := getGitOpsConfig(fileName)
	if err != nil {
		return nil, err
	}
	if err = validateGitOpsConfig(config); err != nil {
		return nil, err
	}
	if err = validateGitOpsRefName(config, refName); err != nil {
		return nil, err
	}
	if err = validateGitOpsNamespace(config, clusterConfig); err != nil {
		return nil, err
	}

	return config, nil
}

func getGitOpsConfig(fileName string) (*GitOpsConfig, error) {
	var config GitOpsConfig
	err := ParseClusterConfig(fileName, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func validateGitOpsConfig(config *GitOpsConfig) error {
	if config == nil {
		return errors.New("gitOpsRef is specified but GitOpsConfig is not specified")
	}

	flux := config.Spec.Flux

	if len(flux.Github.Owner) <= 0 {
		return errors.New("'owner' is not set or empty in gitOps.flux; owner is a required field")
	}
	if len(flux.Github.Repository) <= 0 {
		return errors.New("'repository' is not set or empty in gitOps.flux; repository is a required field")
	}
	err := validateGitRepoName(flux.Github.Repository)
	if err != nil {
		return err
	}
	if len(flux.Github.Branch) > 0 {
		err := validateGitBranchName(config.Spec.Flux.Github.Branch)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateGitBranchName(branchName string) error {
	allowedGitBranchNameRegex := regexp.MustCompile(`^([0-9A-Za-z\_\+,]+)\.?\/?([0-9A-Za-z\-\_\+,]+)$`)

	if !allowedGitBranchNameRegex.MatchString(branchName) {
		return fmt.Errorf("%s is not a valid git branch name, please check with this documentation https://git-scm.com/docs/git-check-ref-format for valid git branch names", branchName)
	}
	return nil
}

func validateGitRepoName(repoName string) error {
	allowedGitRepoName := regexp.MustCompile(`^([0-9A-Za-z-_.]+)$`)
	if !allowedGitRepoName.MatchString(repoName) {
		return fmt.Errorf("%s is not a valid git repository name, name can contain only letters, digits, '_', '-' and '.'", repoName)
	}
	return nil
}

func validateGitOpsRefName(config *GitOpsConfig, refName string) error {
	if config == nil {
		return nil
	}

	if config.Name != refName {
		return fmt.Errorf("GitOpsConfig retrieved with name %s does not match name (%s) specified in "+
			"identityProviderRefs", config.Name, refName)
	}

	return nil
}

func validateGitOpsNamespace(config *GitOpsConfig, clusterConfig *Cluster) error {
	if config == nil {
		return nil
	}

	if config.Namespace != clusterConfig.Namespace {
		return fmt.Errorf("GitOpsConfig and Cluster objects must have the same namespace specified")
	}

	return nil
}

func setGitOpsConfigDefaults(gitops *GitOpsConfig) {
	if gitops == nil {
		return
	}

	c := &gitops.Spec.Flux
	if len(c.Github.FluxSystemNamespace) == 0 {
		c.Github.FluxSystemNamespace = FluxDefaultNamespace
	}

	if len(c.Github.Branch) == 0 {
		c.Github.Branch = FluxDefaultBranch
	}
}
