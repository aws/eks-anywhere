package cmd

/*
	what does this command do?

	this command is responsible for staging cli release

	A PR is then created originating from the forked repo targeting the upstream repo latest release branch

	changes are committed into branch depending on release type
*/
import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v62/github"
	"github.com/spf13/cobra"
)

var (
	cliReleaseNumPath = "release/triggers/eks-a-release/development/RELEASE_NUMBER"
	cliReleaseVerPath = "release/triggers/eks-a-release/development/RELEASE_VERSION"
)

// stageCliCmd represents the stageCli command
var stageCliCmd = &cobra.Command{
	Use:   "stage-cli",
	Short: "creates a PR containing a single commit updating the contents of 2 files intended for staging cli release",
	Long: `Retrieves updated content for development : release_number and release_version. 
	Writes the updated changes to the two files and raises a PR with a single commit.`,

	Run: func(cmd *cobra.Command, args []string) {
		updateAllStageCliFiles()
	},
}

// runs both update functions
func updateAllStageCliFiles() {
	RELEASE_TYPE := os.Getenv("RELEASE_TYPE")

	commitSHA, err := updateFilesStageCli(RELEASE_TYPE)
	if err != nil {
		log.Panic(err)
	}
	fmt.Print(commitSHA)

	err = createPullRequestStageCli(RELEASE_TYPE)
	if err != nil {
		log.Panic(err)
	}
}

func updateFilesStageCli(releaseType string) (string, error) {
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	releaseNumber := os.Getenv("RELEASE_NUMBER")
	latestVersion := os.Getenv("LATEST_VERSION")
	latestRelease := os.Getenv("LATEST_RELEASE")

	// Get the latest commit SHA from the appropriate branch
	ref, _, err := client.Git.GetRef(ctx, usersForkedRepoAccount, EKSAnyrepoName, "heads/"+getBranchName(releaseType, latestRelease))
	if err != nil {
		return "", fmt.Errorf("error getting ref %s", err)
	}
	latestCommitSha := ref.Object.GetSHA()

	entries := []*github.TreeEntry{}
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(cliReleaseNumPath, "/")), Type: github.String("blob"), Content: github.String(string(releaseNumber)), Mode: github.String("100644")})
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(cliReleaseVerPath, "/")), Type: github.String("blob"), Content: github.String(string(latestVersion)), Mode: github.String("100644")})

	tree, _, err := client.Git.CreateTree(ctx, usersForkedRepoAccount, EKSAnyrepoName, *ref.Object.SHA, entries)
	if err != nil {
		return "", fmt.Errorf("error creating tree %s", err)
	}

	newTreeSHA := tree.GetSHA()

	// Create a new commit with all the changes
	author := &github.CommitAuthor{
		Name:  github.String("ibix16"),
		Email: github.String("fake@wtv.com"),
	}

	commit := &github.Commit{
		Message: github.String("Update version files for cli release"),
		Tree:    &github.Tree{SHA: github.String(newTreeSHA)},
		Author:  author,
		Parents: []*github.Commit{{SHA: github.String(latestCommitSha)}},
	}

	commitOP := &github.CreateCommitOptions{}
	newCommit, _, err := client.Git.CreateCommit(ctx, usersForkedRepoAccount, EKSAnyrepoName, commit, commitOP)
	if err != nil {
		return "", fmt.Errorf("creating commit %s", err)
	}
	newCommitSHA := newCommit.GetSHA()

	// Update the branch reference
	ref.Object.SHA = github.String(newCommitSHA)
	_, _, err = client.Git.UpdateRef(ctx, usersForkedRepoAccount, EKSAnyrepoName, ref, false)
	if err != nil {
		return "", fmt.Errorf("error updating ref %s", err)
	}

	return newCommitSHA, nil
}

func createPullRequestStageCli(releaseType string) error {
	latestRelease := os.Getenv("LATEST_RELEASE")

	// Create client
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	base := latestRelease // Target branch for upstream repo
	head := fmt.Sprintf("%s:%s", usersForkedRepoAccount, getBranchName(releaseType, latestRelease))
	title := "Update version files to stage cli release"
	body := "This pull request is responsible for updating the contents of 2 separate files in order to trigger the staging cli release pipeline"

	newPR := &github.NewPullRequest{
		Title: &title,
		Head:  &head,
		Base:  &base,
		Body:  &body,
	}

	pr, _, err := client.PullRequests.Create(ctx, upStreamRepoOwner, EKSAnyrepoName, newPR)
	if err != nil {
		return fmt.Errorf("error creating PR %s", err)
	}

	log.Printf("Pull request created: %s\n", pr.GetHTMLURL())
	return nil
}
