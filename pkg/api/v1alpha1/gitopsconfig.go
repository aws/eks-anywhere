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
