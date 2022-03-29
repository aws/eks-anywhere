package pull_request

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/eks-anywhere/release/pkg/utils"

	"github.com/aws/eks-anywhere/release/pkg/pull_request/types"
)

var (
	bundleReleaseTriggerFiles = map[string]string{
		"BUNDLE_NUMBER":   "BundleNumber",
		"CLI_MIN_VERSION": "CliMinVersion",
		"CLI_MAX_VERSION": "CliMaxVersion",
	}
	eksAReleaseTriggerFiles = map[string]string{
		"RELEASE_NUMBER":  "ReleaseNumber",
		"RELEASE_VERSION": "ReleaseVersion",
	}
)

const (
	originHTTPSRemoteName   = "origin-https"
	upstreamHTTPSRemoteName = "upstream-https"
	upstreamHTTPSURL        = "https://github.com/aws/eks-anywhere.git"
)

func CreatePR(prc *types.PullRequestConfig) error {
	prBranchName := getPRBranchName(prc)
	err := updateTriggerFiles(prc)
	if err != nil {
		return err
	}

	err = updateLocalBranch(prc.GithubUser, prc.BaseBranch, prBranchName)
	if err != nil {
		return err
	}

	err = commitAndCreatePR(prc, prBranchName)
	if err != nil {
		return err
	}

	return nil
}

func updateLocalBranch(githubUser, baseBranch, prBranch string) error {
	originHTTPSURL := fmt.Sprintf("https://github.com/%s/eks-anywhere.git", githubUser)
	updateLocalBranchCommandSequence := fmt.Sprintf("git remote add %s %s; git remote add %s %s; git checkout -B %s; git stash; git fetch %s; git rebase -Xtheirs %s/%s; git stash pop", originHTTPSRemoteName, originHTTPSURL, upstreamHTTPSRemoteName, upstreamHTTPSURL, prBranch, upstreamHTTPSRemoteName, upstreamHTTPSRemoteName, baseBranch)

	cmd := exec.Command("bash", "-c", updateLocalBranchCommandSequence)
	out, err := utils.ExecCommand(cmd)
	if err != nil {
		return err
	}
	fmt.Println(out)

	return nil
}

func commitAndCreatePR(prc *types.PullRequestConfig, prBranchName string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	gitAddCommandSequence := ""
	triggerFiles := getTriggerFilesForRelease(prc.ReleaseType)
	triggerFileDir := filepath.Join("triggers", fmt.Sprintf("%s-release", prc.ReleaseType), prc.ReleaseEnvironment)
	for fileName := range triggerFiles {
		triggerFilePath := filepath.Join(pwd, triggerFileDir, fileName)
		gitAddCommandSequence += fmt.Sprintf("git add %s;", triggerFilePath)
	}
	commitMessage := fmt.Sprintf("Bump up trigger files for %s %s release", prc.ReleaseEnvironment, strings.ToTitle(prc.ReleaseType))

	addAndCommitFilesCommandSequence := fmt.Sprintf("%s git commit -m '%s'", gitAddCommandSequence, commitMessage)

	cmd := exec.Command("bash", "-c", addAndCommitFilesCommandSequence)
	out, err := utils.ExecCommand(cmd)
	if err != nil {
		return err
	}
	fmt.Println(out)

	if prc.DryRun {
		fmt.Println("Dry-run of PR creation successful")
		return nil
	}

	prTitle := commitMessage
	prBody := fmt.Sprintf(`
	%s
	
	By submitting this pull request, I confirm that you can use, modify, copy, and redistribute this contribution, under the terms of your choice.`, prTitle)
	ghPushCommand := fmt.Sprintf("git push -u %s %s -f", originHTTPSRemoteName, prBranchName)
	ghLoginCommand := fmt.Sprintf("echo -n %s | gh auth login --with-token", githubToken)
	ghPrCreateCommand := fmt.Sprintf("gh pr create --title '%s' --body '%s' --base '%s' --label 'do-not-merge/hold'", prTitle, prBody, prc.BaseBranch)
	createPRCommandSequence := fmt.Sprintf("%s; %s; %s", ghPushCommand, ghLoginCommand, ghPrCreateCommand)
	cmd = exec.Command("bash", "-c", createPRCommandSequence)
	out, err = utils.ExecCommand(cmd)
	if err != nil {
		return err
	}
	fmt.Println(out)

	return nil
}

func updateTriggerFiles(prc *types.PullRequestConfig) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	triggerFiles := getTriggerFilesForRelease(prc.ReleaseType)
	triggerFileDir := filepath.Join("triggers", fmt.Sprintf("%s-release", prc.ReleaseType), prc.ReleaseEnvironment)
	for fileName, fieldName := range triggerFiles {
		triggerFilePath := filepath.Join(pwd, triggerFileDir, fileName)
		fieldValue := getPropertyByName(prc, fieldName)
		err := updateFile(triggerFilePath, fieldValue.(string))
		if err != nil {
			return err
		}
	}

	return nil
}

func updateFile(filename string, value string) error {
	fd, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = fd.Write([]byte(value))
	if err != nil {
		return err
	}

	return nil
}

func getTriggerFilesForRelease(releaseType string) map[string]string {
	if releaseType == types.BundleKind {
		return bundleReleaseTriggerFiles
	} else if releaseType == types.EksAKind {
		return eksAReleaseTriggerFiles
	}

	return nil
}

func getPRBranchName(prc *types.PullRequestConfig) string {
	if prc.ReleaseType == types.BundleKind {
		return fmt.Sprintf("trigger-%s-release-%s", prc.ReleaseType, prc.BundleNumber)
	} else if prc.ReleaseType == types.EksAKind {
		return fmt.Sprintf("trigger-%s-release-%s", prc.ReleaseType, prc.ReleaseVersion)
	}

	return ""
}

func getPropertyByName(prc *types.PullRequestConfig, fieldName string) interface{} {
	prcMarshal, err := json.Marshal(prc)
	if err != nil {
		return err
	}

	var x map[string]interface{}
	err = json.Unmarshal(prcMarshal, &x)
	if err != nil {
		return err
	}

	return x[fieldName]
}
