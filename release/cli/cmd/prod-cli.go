package cmd

/*
	what does this command do?

	this command is responsible for staging prod cli release

	A PR is then created originating from the forked repo targeting the upstream repo latest release branch

	changes are committed into branch depending on release type
*/
import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/v62/github"
	"github.com/spf13/cobra"
)

var (
	prodCliReleaseNumPath = "release/triggers/eks-a-release/production/RELEASE_NUMBER"
	prodCliReleaseVerPath = "release/triggers/eks-a-release/production/RELEASE_VERSION"
)

// prodCliCmd represents the prodCli command
var prodCliCmd = &cobra.Command{
	Use:   "prod-cli",
	Short: "creates a PR containing a single commit updating the contents of 2 files intended for prod cli release",
	Long: `Retrieves updated content for production : release_number and release_version. 
	Writes the updated changes to the two files and raises a PR with a single commit.`,

	Run: func(cmd *cobra.Command, args []string) {
		updateAllProdCliFiles()
	},
}

// runs both updates functions
func updateAllProdCliFiles() {

	_, err := updateProdCliFiles(RELEASE_TYPE)
	if err != nil {
		log.Panic(err)
	}

	err = createProdCliPullRequest(RELEASE_TYPE)
	if err != nil {
		log.Panic(err)
	}
}

func updateProdCliFiles(releaseType string) (string, error) {

	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	// Get the latest commit SHA from the appropriate branch
	ref, _, err := client.Git.GetRef(ctx, usersForkedRepoAccount, EKSAnyrepoName, "heads/"+getBranchName(releaseType, latestRelease))
	if err != nil {
		return "", fmt.Errorf("error getting ref %s", err)
	}
	latestCommitSha := ref.Object.GetSHA()

	entries := []*github.TreeEntry{}
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(prodCliReleaseNumPath, "/")), Type: github.String("blob"), Content: github.String(string(releaseNumber)), Mode: github.String("100644")})
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(prodCliReleaseVerPath, "/")), Type: github.String("blob"), Content: github.String(string(latestVersion)), Mode: github.String("100644")})

	tree, _, err := client.Git.CreateTree(ctx, usersForkedRepoAccount, EKSAnyrepoName, *ref.Object.SHA, entries)
	if err != nil {
		return "", fmt.Errorf("error creating tree %s", err)
	}

	newTreeSHA := tree.GetSHA()

	// Create a new commit with all the changes
	author := &github.CommitAuthor{
		Name:  github.String("eks-a-releaser"),
		Email: github.String("fake@wtv.com"),
	}

	commit := &github.Commit{
		Message: github.String("Update version files for prod cli release"),
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

func createProdCliPullRequest(releaseType string) error {

	// Create client
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	base := latestRelease // Target branch for upstream repo
	head := fmt.Sprintf("%s:%s", usersForkedRepoAccount, getBranchName(releaseType, latestRelease))
	title := "Update version files to stage prod cli release"
	body := "This pull request is responsible for updating the contents of 2 separate files in order to trigger the prod cli release pipeline"

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
