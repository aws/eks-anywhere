package cmd

/*
	what does this command do?

	this command is responsible for creating a release tag with the commit hash that triggered the prod CLI release

	depending on release type, either minor or patch branch will be checked to retrieve commit hash
*/

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v62/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

// createReleaseCmd represents the createRelease command
var createReleaseCmd = &cobra.Command{
	Use:   "create-release",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,

	Run: func(cmd *cobra.Command, args []string) {
		runBothTag()
	},
}

func runBothTag() {

	RELEASE_TYPE := os.Getenv("RELEASE_TYPE")

	//retrieve commit hash
	commitHash := retrieveLatestProdCLIHash(RELEASE_TYPE)

	//create tag with commit hash
	tag, errOne := createTag(commitHash)
	if errOne != nil {
		log.Panic(errOne)
	}

	rel, errTwo := createGitHubRelease(tag)
	if errTwo != nil {
		log.Panic(errTwo)
	}

	//print release object
	fmt.Print(rel)
}

// creates tag using retrieved commit hash
func createTag(commitHash string) (*github.RepositoryRelease, error) {

	// retrieve tag name "v0.0.00"
	latestVersionValue := os.Getenv("LATEST_VERSION")

	//create client
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()

	// Create a new GitHub client instance with the token type set to "Bearer"
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	releaseName := latestVersionValue
	releaseDesc := latestVersionValue //"EKS-Anywhere " + latestVersionValue + " release"
	commitSHA := commitHash
	release := &github.RepositoryRelease{
		TagName:         github.String(releaseName),
		Name:            github.String(releaseName),
		Body:            github.String(releaseDesc),
		TargetCommitish: github.String(commitSHA),
	}

	rel, _, err := client.Repositories.CreateRelease(ctx, upStreamRepoOwner, EKSAnyrepoName, release)
	if err != nil {
		fmt.Printf("error creating release: %v", err)
	}

	fmt.Printf("Release tag %s created successfully!\n", rel.GetTagName())
	return rel, nil
}

func retrieveLatestProdCLIHash(releaseType string) string {
	latestRelease := os.Getenv("LATEST_RELEASE")

	if releaseType == "minor" {

		//create client
		accessToken := os.Getenv("SECRET_PAT")
		ctx := context.Background()
		client := github.NewClient(nil).WithAuthToken(accessToken)

		opts := &github.CommitsListOptions{
			Path: prodCliReleaseVerPath, // file to check
			SHA:  latestRelease,         // branch to check - release-0.xx
		}

		commits, _, err := client.Repositories.ListCommits(ctx, usersForkedRepoAccount, EKSAnyrepoName, opts)
		if err != nil {
			return "error fetching commits list"
		}

		if len(commits) > 0 {
			latestCommit := commits[0]
			return latestCommit.GetSHA()
		}

		return "no commits found for file"

	}

	// else
	//create client
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	opts := &github.CommitsListOptions{
		Path: prodCliReleaseVerPath,             // file to check
		SHA:  latestRelease + "-releaser-patch", // branch to check - release-0.xx-releaser-patch
	}

	commits, _, err := client.Repositories.ListCommits(ctx, usersForkedRepoAccount, EKSAnyrepoName, opts)
	if err != nil {
		return "error fetching commits list"
	}

	if len(commits) > 0 {
		latestCommit := commits[0]
		return latestCommit.GetSHA()
	}

	return "no commits found for file"

}

func createGitHubRelease(releaseTag *github.RepositoryRelease) (*github.RepositoryRelease, error) {

	latestVersionValue := os.Getenv("LATEST_VERSION")

	//create client
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	release, _, err := client.Repositories.GetReleaseByTag(ctx, upStreamRepoOwner, EKSAnyrepoName, latestVersionValue)
	if err == nil {
		fmt.Printf("Release %s already exists!\n", latestVersionValue)
		return release, nil
	}

	release = &github.RepositoryRelease{
		TagName: releaseTag.TagName,
		Name:    &latestVersionValue,
		Body:    releaseTag.Body,
	}

	rel, _, err := client.Repositories.CreateRelease(ctx, upStreamRepoOwner, EKSAnyrepoName, release)
	if err != nil {
		return nil, err
	}

	return rel, nil
}
