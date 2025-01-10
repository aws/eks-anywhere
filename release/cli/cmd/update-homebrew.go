package cmd

/*
	what does this command do?

	this command is responsible for updating the homebrew file

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
	homebrewPath = "release/triggers/brew-version-release/CLI_RELEASE_VERSION"
)

// updateHomebrewCmd represents the updateHomebrew command
var updateHomebrewCmd = &cobra.Command{
	Use:   "update-homebrew",
	Short: "Updates homebrew with latest version in eks-a-releaser branch, PR targets release branch",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,

	Run: func(cmd *cobra.Command, args []string) {
		runAllHomebrew()
	},
}

func runAllHomebrew() {
	
	_, err := updateHomebrew(RELEASE_TYPE)
	if err != nil {
		log.Panic(err)
	}

	err = createPullRequestHomebrew(RELEASE_TYPE)
	if err != nil {
		log.Panic(err)
	}
}

func updateHomebrew(releaseType string) (string, error) {

	// Create client
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	opts := &github.RepositoryContentGetOptions{
		Ref: "main", // Specific branch to check for homebrew file
	}

	// Access homebrew file
	FileContentBundleNumber, _, _, err := client.Repositories.GetContents(ctx, usersForkedRepoAccount, EKSAnyrepoName, homebrewPath, opts)
	if err != nil {
		return "", fmt.Errorf("error accessing homebrew file %s", err)
	}

	// Holds content of homebrew cli version file
	content, err := FileContentBundleNumber.GetContent()
	if err != nil {
		return "", fmt.Errorf("error fetching file contents %s", err)
	}

	// Update instances of previous release with new
	updatedFile := strings.ReplaceAll(content, content, latestVersion)

	// Get the latest commit SHA from the appropriate branch
	ref, _, err := client.Git.GetRef(ctx, usersForkedRepoAccount, EKSAnyrepoName, "heads/"+getBranchName(releaseType, latestRelease))
	if err != nil {
		return "", fmt.Errorf("error getting ref %s", err)
	}
	latestCommitSha := ref.Object.GetSHA()

	entries := []*github.TreeEntry{}
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(homebrewPath, "/")), Type: github.String("blob"), Content: github.String(string(updatedFile)), Mode: github.String("100644")})
	tree, _, err := client.Git.CreateTree(ctx, usersForkedRepoAccount, EKSAnyrepoName, *ref.Object.SHA, entries)
	if err != nil {
		return "", fmt.Errorf("error creating tree %s", err)
	}

	newTreeSHA := tree.GetSHA()

	// Create a new commit
	author := &github.CommitAuthor{
		Name:  github.String("eks-a-releaser"),
		Email: github.String("fake@wtv.com"),
	}

	commit := &github.Commit{
		Message: github.String("Update brew-version value to point to new release"),
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

func createPullRequestHomebrew(releaseType string) error {

	// Create client
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	base := latestRelease // Target branch for upstream repo
	head := fmt.Sprintf("%s:%s", usersForkedRepoAccount, getBranchName(releaseType, latestRelease))
	title := "Update homebrew cli version value to point to new release"
	body := "This pull request is responsible for updating the contents of the home brew cli version file"

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
