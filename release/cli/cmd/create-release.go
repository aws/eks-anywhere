package cmd

/*
	what does this command do?

	this command is responsible for creating a release tag with the commit hash that triggered the prod CLI release

	depending on release type, either minor or patch branch will be checked to retrieve commit hash
*/

import (
	"context"
	"log"

	"github.com/google/go-github/v62/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

// createReleaseCmd represents the createRelease command
var createReleaseCmd = &cobra.Command{
	Use:   "create-release",
	Short: "this command is responsible for creating a release tag with the commit hash that triggered the prod CLI release",
	Long: `this command is responsible for creating a release tag with the commit hash that triggered the prod CLI release
	depending on release type, either minor or patch branch will be checked to retrieve commit hash.`,

	Run: func(cmd *cobra.Command, args []string) {
		runBothTag()
	},
}

func runBothTag() {

	//retrieve commit hash
	commitHash, HashErr := retrieveLatestProdCLIHash(RELEASE_TYPE)
	if HashErr != nil {
		log.Panic(HashErr)
	} 

	//create tag with commit hash
	tag, errOne := createTag(commitHash)
	if errOne != nil {
		log.Panic(errOne)
	}

	rel, errTwo := createGitHubRelease(tag)
	if errTwo != nil {
		log.Panic(errTwo)
	}

	//log release object
	log.Print(rel)
}

// creates tag using retrieved commit hash
func createTag(commitHash string) (*github.RepositoryRelease, error) {

	ctx := context.Background()
	// Create a new GitHub client instance with the token type set to "Bearer"
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	releaseName := latestVersion
	releaseDesc := latestVersion //"EKS-Anywhere " + latestVersionValue + " release"
	commitSHA := commitHash
	release := &github.RepositoryRelease{
		TagName:         github.String(releaseName),
		Name:            github.String(releaseName),
		Body:            github.String(releaseDesc),
		TargetCommitish: github.String(commitSHA),
	}

	rel, _, err := client.Repositories.CreateRelease(ctx, upStreamRepoOwner, EKSAnyrepoName, release)
	if err != nil {
		log.Printf("error creating release: %v", err)
	}

	log.Printf("Release tag %s created successfully!\n", rel.GetTagName())
	return rel, nil
}

func retrieveLatestProdCLIHash(releaseType string) (string, error) {

	if releaseType == "minor" {

		//create client
		ctx := context.Background()
		client := github.NewClient(nil).WithAuthToken(accessToken)

		opts := &github.CommitsListOptions{
			Path: prodCliReleaseVerPath, // file to check
			SHA:  latestRelease,         // branch to check - release-0.xx
		}

		commits, _, err := client.Repositories.ListCommits(ctx, usersForkedRepoAccount, EKSAnyrepoName, opts)
		if err != nil {
			return "error fetching commits list: ", err
		}

		if len(commits) > 0 {
			latestCommit := commits[0]
			return latestCommit.GetSHA(), nil
		}

		return "no commits found for file", nil

	}

	// else
	//create client
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	opts := &github.CommitsListOptions{
		Path: prodCliReleaseVerPath,             // file to check
		SHA:  latestRelease + "-releaser-patch", // branch to check - release-0.xx-releaser-patch
	}

	commits, _, err := client.Repositories.ListCommits(ctx, usersForkedRepoAccount, EKSAnyrepoName, opts)
	if err != nil {
		return "error fetching commits list: ", err
	}

	if len(commits) > 0 {
		latestCommit := commits[0]
		return latestCommit.GetSHA(), nil
	}

	return "no commits found for file", nil

}

func createGitHubRelease(releaseTag *github.RepositoryRelease) (*github.RepositoryRelease, error) {

	//create client
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	release, _, err := client.Repositories.GetReleaseByTag(ctx, upStreamRepoOwner, EKSAnyrepoName, latestVersion)
	if err == nil {
		log.Printf("Release %s already exists!\n", latestVersion)
		return release, nil
	}

	release = &github.RepositoryRelease{
		TagName: releaseTag.TagName,
		Name:    &latestVersion,
		Body:    releaseTag.Body,
	}

	rel, _, err := client.Repositories.CreateRelease(ctx, upStreamRepoOwner, EKSAnyrepoName, release)
	if err != nil {
		return nil, err
	}

	return rel, nil
}
