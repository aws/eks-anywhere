package cmd

/*
	what does this command do?

	this command is responsible for accessing and updating the Makefile with the latest release value

	the updated makefile is committed to the latest release branch, forked repo

	and a pull request is raised targeting the upstream repo latest release branch
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
	EKSAnyrepoName = "eks-anywhere"
	makeFilePath   = "/Makefile"
)

// upMakeFileCmd represents the upMakeFile command
var updateMakefileCmd = &cobra.Command{
	Use:   "update-makefile",
	Short: "Updates BRANCH_NAME?= variable to match new release branch within the Makefile",
	Long:  `A longer description.`,

	Run: func(cmd *cobra.Command, args []string) {
		content := updateMakefile()
		fmt.Print(content)
	},
}

func updateMakefile() error {

	// create client
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)


	opts := &github.RepositoryContentGetOptions{
		Ref: "main", // branch that will be accessed
	}

	// access makefile in forked repo and retrieve entire file contents
	triggerFileContentBundleNumber, _, _, err := client.Repositories.GetContents(ctx, usersForkedRepoAccount, EKSAnyrepoName, makeFilePath, opts)
	if err != nil {
		return fmt.Errorf("error accessing file : %s", err)
	}
	// holds makefile
	content, err := triggerFileContentBundleNumber.GetContent()
	if err != nil {
		return fmt.Errorf("error fetching file content : %s", err)
	}

	// stores entire updated Makefile as a string
	updatedContent := returnUpdatedMakeFile(content, latestRelease)

	// get latest commit sha from latest release branch
	ref, _, err := client.Git.GetRef(ctx, usersForkedRepoAccount, EKSAnyrepoName, "heads/"+latestRelease)
	if err != nil {
		return fmt.Errorf("error getting ref %s", err)
	}
	latestCommitSha := ref.Object.GetSHA()

	entries := []*github.TreeEntry{}
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(makeFilePath, "/")), Type: github.String("blob"), Content: github.String(string(updatedContent)), Mode: github.String("100644")})
	tree, _, err := client.Git.CreateTree(ctx, usersForkedRepoAccount, EKSAnyrepoName, *ref.Object.SHA, entries)
	if err != nil {
		return fmt.Errorf("error creating tree %s", err)
	}

	//validate tree sha
	newTreeSHA := tree.GetSHA()

	// create new commit, update email address
	author := &github.CommitAuthor{
		Name:  github.String("eks-a-releaser"),
		Email: github.String("fake@wtv.com"),
	}

	commit := &github.Commit{
		Message: github.String("Update Makefile"),
		Tree:    &github.Tree{SHA: github.String(newTreeSHA)},
		Author:  author,
		Parents: []*github.Commit{{SHA: github.String(latestCommitSha)}},
	}

	commitOP := &github.CreateCommitOptions{}
	newCommit, _, err := client.Git.CreateCommit(ctx, usersForkedRepoAccount, EKSAnyrepoName, commit, commitOP)
	if err != nil {
		return fmt.Errorf("creating commit %s", err)
	}
	newCommitSHA := newCommit.GetSHA()

	// update branch reference
	ref.Object.SHA = github.String(newCommitSHA)

	_, _, err = client.Git.UpdateRef(ctx, usersForkedRepoAccount, EKSAnyrepoName, ref, false)
	if err != nil {
		return fmt.Errorf("error updating ref %s", err)
	}

	// create pull request
	base := latestRelease // branch PR will be merged into
	head := fmt.Sprintf("%s:%s", usersForkedRepoAccount, latestRelease)
	title := "Updates Makefile to point to new release"
	body := "This pull request is responsible for updating the contents of the Makefile"

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

// updates Makefile with new release, returns entire file updated
func returnUpdatedMakeFile(fileContent, newRelease string) string {
	snippetStartIdentifierB := "BRANCH_NAME?="
	lines := strings.Split(fileContent, "\n")
	var updatedLines []string

	for _, line := range lines {
		if strings.Contains(line, snippetStartIdentifierB) {
			parts := strings.Split(line, "=")
			varNamePart := parts[0] // holds "BRANCH_NAME?"
			updatedLine := varNamePart + "=" + newRelease
			updatedLines = append(updatedLines, updatedLine)
		} else {
			updatedLines = append(updatedLines, line)
		}
	}

	return strings.Join(updatedLines, "\n")

}
