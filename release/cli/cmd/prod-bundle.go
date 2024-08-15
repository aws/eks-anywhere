package cmd

/*
	what does this command do?

	this command is responsible for staging prod bundle release

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
	prodBundleNumPath     = "release/triggers/bundle-release/production/BUNDLE_NUMBER"
	prodCliMaxVersionPath = "release/triggers/bundle-release/production/CLI_MAX_VERSION"
	prodCliMinVersionPath = "release/triggers/bundle-release/production/CLI_MIN_VERSION"
)

// prodBundleCmd represents the prodBundle command
var prodBundleCmd = &cobra.Command{
	Use:   "prod-bundle",
	Short: "creates a PR containing a single commit updating the contents of 3 files intended for prod bundle release",
	Long: `Retrieves updated content for production : bundle number, cli max version, and cli min version. 
	Writes the updated changes to the 3 files and raises a PR with a single commit.`,

	Run: func(cmd *cobra.Command, args []string) {
		err := updateAllProdBundleFiles()
		if err != nil {
			log.Panic(err)
		}
	},
}

func updateAllProdBundleFiles() error {
	RELEASE_TYPE := os.Getenv("RELEASE_TYPE")

	_, err := updateProdBundleFiles(RELEASE_TYPE)
	if err != nil {
		return err
	}

	err = createProdBundlePullRequest(RELEASE_TYPE)
	if err != nil {
		return err
	}

	return nil
}

func updateProdBundleFiles(releaseType string) (string, error) {
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	bundleNumber := os.Getenv("RELEASE_NUMBER")
	latestVersion := os.Getenv("LATEST_VERSION")
	latestRelease := os.Getenv("LATEST_RELEASE")

	// Get the latest commit SHA from the appropriate branch
	ref, _, err := client.Git.GetRef(ctx, usersForkedRepoAccount, EKSAnyrepoName, "heads/"+getBranchName(releaseType, latestRelease))
	if err != nil {
		return "", fmt.Errorf("error getting ref %s", err)
	}
	latestCommitSha := ref.Object.GetSHA()

	entries := []*github.TreeEntry{}
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(prodBundleNumPath, "/")), Type: github.String("blob"), Content: github.String(string(bundleNumber)), Mode: github.String("100644")})
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(prodCliMaxVersionPath, "/")), Type: github.String("blob"), Content: github.String(string(latestVersion)), Mode: github.String("100644")})
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(prodCliMinVersionPath, "/")), Type: github.String("blob"), Content: github.String(string(latestVersion)), Mode: github.String("100644")})

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
		Message: github.String("Update version files for production bundle release"),
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

func createProdBundlePullRequest(releaseType string) error {
	latestRelease := os.Getenv("LATEST_RELEASE")

	// Create client
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	base := latestRelease // Target branch for upstream repo
	head := fmt.Sprintf("%s:%s", usersForkedRepoAccount, getBranchName(releaseType, latestRelease))
	title := "Update version files to stage production bundle release"
	body := "This pull request is responsible for updating the contents of 3 separate files in order to trigger the production bundle release pipeline"

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
