package gitfactory

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	gogitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"golang.org/x/crypto/ssh"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/git/gitclient"
	"github.com/aws/eks-anywhere/pkg/git/gogithub"
	"github.com/aws/eks-anywhere/pkg/git/providers/codecommit"
	"github.com/aws/eks-anywhere/pkg/git/providers/github"
)

type GitTools struct {
	Provider            git.ProviderClient
	Client              git.Client
	Writer              filewriter.FileWriter
	RepositoryDirectory string
}

type GitToolsOpt func(opts *GitTools)

func Build(ctx context.Context, cluster *v1alpha1.Cluster, fluxConfig *v1alpha1.FluxConfig, writer filewriter.FileWriter, opts ...GitToolsOpt) (*GitTools, error) {
	var repo string
	var repoUrl string
	var gitAuth transport.AuthMethod
	var err error
	var tools GitTools

	switch {
	case fluxConfig.Spec.Github != nil:
		githubToken, err := github.GetGithubAccessTokenFromEnv()
		if err != nil {
			return nil, err
		}

		tools.Provider, err = buildGithubProvider(ctx, githubToken, fluxConfig.Spec.Github)
		if err != nil {
			return nil, fmt.Errorf("building github provider: %v", err)
		}

		gitAuth = &http.BasicAuth{Password: githubToken, Username: fluxConfig.Spec.Github.Owner}
		repo = fluxConfig.Spec.Github.Repository
		repoUrl = github.RepoUrl(fluxConfig.Spec.Github.Owner, repo)
	case fluxConfig.Spec.Git != nil:
		privateKeyFile := os.Getenv(config.EksaGitPrivateKeyTokenEnv)
		privateKeyPassphrase := os.Getenv(config.EksaGitPassphraseTokenEnv)
		gitKnownHosts := os.Getenv(config.EksaGitKnownHostsFileEnv)
		if err = os.Setenv(config.SshKnownHostsEnv, gitKnownHosts); err != nil {
			return nil, fmt.Errorf("unable to set %s: %v", config.SshKnownHostsEnv, err)
		}
		user := codecommit.DefaultSSHAuthUser
		repoUrl = fluxConfig.Spec.Git.RepositoryUrl
		codeCommitUserName, err := codecommit.IsCodeCommitURL(repoUrl)
		if err != nil {
			return nil, err
		}
		if codeCommitUserName != "" {
			user = codeCommitUserName
		}
		gitAuth, err = getSSHAuthFromPrivateKey(privateKeyFile, privateKeyPassphrase, user)
		if err != nil {
			return nil, err
		}
		repo = path.Base(strings.TrimSuffix(repoUrl, filepath.Ext(repoUrl)))
	default:
		return nil, fmt.Errorf("no valid git provider in FluxConfigSpec. Spec: %v", fluxConfig)
	}

	tools.RepositoryDirectory = filepath.Join(cluster.Name, "git", repo)
	for _, opt := range opts {
		if opt != nil {
			opt(&tools)
		}
	}
	tools.Client = buildGitClient(ctx, gitAuth, repoUrl, tools.RepositoryDirectory)

	tools.Writer, err = newRepositoryWriter(writer, repo)
	if err != nil {
		return nil, err
	}

	return &tools, nil
}

func buildGitClient(ctx context.Context, auth transport.AuthMethod, repoUrl string, repo string) *gitclient.GitClient {
	opts := []gitclient.Opt{
		gitclient.WithRepositoryUrl(repoUrl),
		gitclient.WithRepositoryDirectory(repo),
		gitclient.WithAuth(auth),
	}

	return gitclient.New(opts...)
}

func buildGithubProvider(ctx context.Context, githubToken string, config *v1alpha1.GithubProviderConfig) (git.ProviderClient, error) {
	auth := git.TokenAuth{Token: githubToken, Username: config.Owner}
	gogithubOpts := gogithub.Options{Auth: auth}
	githubProviderClient := gogithub.New(ctx, gogithubOpts)
	provider, err := github.New(githubProviderClient, config, auth)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

func newRepositoryWriter(writer filewriter.FileWriter, repository string) (filewriter.FileWriter, error) {
	localGitWriterPath := filepath.Join("git", repository)
	gitwriter, err := writer.WithDir(localGitWriterPath)
	if err != nil {
		return nil, fmt.Errorf("creating file writer: %v", err)
	}
	gitwriter.CleanUpTemp()
	return gitwriter, nil
}

func WithRepositoryDirectory(repoDir string) GitToolsOpt {
	return func(opts *GitTools) {
		opts.RepositoryDirectory = repoDir
	}
}

func getSSHAuthFromPrivateKey(privateKeyFile string, passphrase string, user string) (gogitssh.AuthMethod, error) {
	signer, err := getSignerFromPrivateKeyFile(privateKeyFile, passphrase)
	if err != nil {
		return nil, err
	}
	return &gogitssh.PublicKeys{
		Signer: signer,
		User:   user,
	}, nil
}

func getSignerFromPrivateKeyFile(privateKeyFile string, passphrase string) (ssh.Signer, error) {
	var signer ssh.Signer
	var err error

	sshKey, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return nil, err
	}

	if passphrase == "" {
		signer, err = ssh.ParsePrivateKey(sshKey)
		if err != nil {
			if _, ok := err.(*ssh.PassphraseMissingError); ok {
				return nil, fmt.Errorf("%s, please set the EKSA_GIT_SSH_KEY_PASSPHRASE environment variable", err)
			}
			return nil, err
		}
		return signer, nil
	}
	return ssh.ParsePrivateKeyWithPassphrase(sshKey, []byte(passphrase))
}
