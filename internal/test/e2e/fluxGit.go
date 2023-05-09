package e2e

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere/internal/pkg/ssm"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/retrier"
	e2etests "github.com/aws/eks-anywhere/test/framework"
)

type s3Files struct {
	key, dstPath string
	permission   int
}

type fileFromBytes struct {
	dstPath    string
	permission int
	content    []byte
}

func (f *fileFromBytes) contentString() string {
	return string(f.content)
}

func (e *E2ESession) setupFluxGitEnv(testRegex string) error {
	re := regexp.MustCompile(`^.*GitFlux.*$`)
	if !re.MatchString(testRegex) {
		return nil
	}

	requiredEnvVars := e2etests.RequiredFluxGitCreateRepoEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}

	repo, err := e.setupGithubRepo()
	if err != nil {
		return fmt.Errorf("setting up github repo for test: %v", err)
	}
	// add the newly generated repository to the test
	e.testEnvVars[e2etests.GitRepoSshUrl] = gitRepoSshUrl(repo.Name, e.testEnvVars[e2etests.GithubUserVar])

	for _, file := range buildFluxGitFiles(e.testEnvVars) {
		if err := e.downloadFileInInstance(file); err != nil {
			return fmt.Errorf("downloading flux-git file to instance: %v", err)
		}
	}

	err = e.setUpSshAgent(e.testEnvVars[config.EksaGitPrivateKeyTokenEnv])
	if err != nil {
		return fmt.Errorf("setting up ssh agent on remote instance: %v", err)
	}

	return nil
}

func buildFluxGitFiles(envVars map[string]string) []s3Files {
	return []s3Files{
		{
			key:        "git-flux/known_hosts",
			dstPath:    envVars[config.EksaGitKnownHostsFileEnv],
			permission: 644,
		},
	}
}

func (e *E2ESession) decodeAndWriteFileToInstance(file fileFromBytes) error {
	e.logger.V(1).Info("Writing bytes to file in instance", "file", file.dstPath)

	command := fmt.Sprintf("echo '%s' | base64 -d >> %s && chmod %d %[2]s", file.contentString(), file.dstPath, file.permission)
	if err := ssm.Run(e.session, logr.Discard(), e.instanceId, command, ssmTimeout); err != nil {
		return fmt.Errorf("writing file in instance: %v", err)
	}
	e.logger.V(1).Info("Successfully decoded and wrote file", "file", file.dstPath)

	return nil
}

func (e *E2ESession) downloadFileInInstance(file s3Files) error {
	e.logger.V(1).Info("Downloading from s3 in instance", "file", file.key)

	command := fmt.Sprintf("aws s3 cp s3://%s/%s %s && chmod %d %[3]s", e.storageBucket, file.key, file.dstPath, file.permission)
	if err := ssm.Run(e.session, logr.Discard(), e.instanceId, command, ssmTimeout); err != nil {
		return fmt.Errorf("downloading file in instance: %v", err)
	}
	e.logger.V(1).Info("Successfully downloaded file", "file", file.key)

	return nil
}

func (e *E2ESession) setUpSshAgent(privateKeyFile string) error {
	command := fmt.Sprintf("eval $(ssh-agent -s) ssh-add %s", privateKeyFile)

	if err := ssm.Run(e.session, logr.Discard(), e.instanceId, command, ssmTimeout); err != nil {
		return fmt.Errorf("starting SSH agent on instance: %v", err)
	}
	e.logger.V(1).Info("Successfully started SSH agent on instance")

	return nil
}

func (e *E2ESession) setupGithubRepo() (*git.Repository, error) {
	e.logger.V(1).Info("setting up Github repo for test")
	owner := e.testEnvVars[e2etests.GithubUserVar]
	repo := strings.ReplaceAll(e.jobId, ":", "-") // Github API urls get funky if you use ":" in the repo name

	c := &v1alpha1.GithubProviderConfig{
		Owner:      owner,
		Repository: repo,
		Personal:   true,
	}

	ctx := context.Background()
	g, err := e.TestGithubClient(ctx, e.testEnvVars[e2etests.GithubTokenVar], c.Owner, c.Repository, c.Personal)
	if err != nil {
		return nil, fmt.Errorf("couldn't create Github client for test setup: %v", err)
	}

	// Create a new github repository for the tests to run on
	o := git.CreateRepoOpts{
		Name:        repo,
		Owner:       owner,
		Description: fmt.Sprintf("repository for use with E2E test job %v", e.jobId),
		Personal:    true,
		AutoInit:    true,
	}

	r, err := g.CreateRepo(ctx, o)
	if err != nil {
		return nil, fmt.Errorf("creating repository in Github for test: %v", err)
	}

	pk, pub, err := e.generateKeyPairForGitTest()
	if err != nil {
		return nil, fmt.Errorf("generating key pair for git tests: %v", err)
	}

	e.logger.Info("Create Deploy Key Configuration for Git Flux tests", "owner", owner, "repo", repo)
	// Add the newly generated public key to the newly created repository as a deploy key
	ko := git.AddDeployKeyOpts{
		Owner:      owner,
		Repository: repo,
		Key:        string(pub),
		Title:      fmt.Sprintf("Test key created for job %v", e.jobId),
		ReadOnly:   false,
	}

	// Newly generated repositories may take some time to show up in the GitHub API; retry a few times to get around this
	err = retrier.Retry(10, time.Second*10, func() error {
		err = g.AddDeployKeyToRepo(ctx, ko)
		if err != nil {
			return fmt.Errorf("couldn't add deploy key to repo: %v", err)
		}
		return nil
	})
	if err != nil {
		return r, err
	}

	encodedPK := encodePrivateKey(pk)
	// Generate a PEM file from the private key and write it instance at the user-provided path
	pkFile := fileFromBytes{
		dstPath:    e.testEnvVars[config.EksaGitPrivateKeyTokenEnv],
		permission: 600,
		content:    encodedPK,
	}

	err = e.decodeAndWriteFileToInstance(pkFile)
	if err != nil {
		return nil, fmt.Errorf("writing private key file to instance: %v", err)
	}

	return r, err
}

func encodePrivateKey(privateKey []byte) []byte {
	b64EncodedPK := make([]byte, base64.StdEncoding.EncodedLen(len(privateKey)))
	base64.StdEncoding.Encode(b64EncodedPK, privateKey)
	return b64EncodedPK
}

func (e *E2ESession) generateKeyPairForGitTest() (privateKeyBytes, publicKeyBytes []byte, err error) {
	k, err := generateKeyPairEcdsa()
	if err != nil {
		return nil, nil, err
	}

	privateKeyBytes, err = pemFromPrivateKeyEcdsa(k)
	if err != nil {
		return nil, nil, err
	}

	publicKeyBytes, err = pubFromPrivateKeyEcdsa(k)
	if err != nil {
		return nil, nil, err
	}

	log.Println("Public key generated")
	return privateKeyBytes, publicKeyBytes, nil
}

func gitRepoSshUrl(repo, owner string) string {
	t := "ssh://git@github.com/%s/%s.git"
	return fmt.Sprintf(t, owner, repo)
}
