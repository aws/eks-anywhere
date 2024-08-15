package cmd

/*
	what does this command do?

	this command is responsible for staging bundle release

	A PR is created originating from the forked repo targeting the upstream repo latest release branch

	changes are committed into branch depending on release type
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v62/github"
	"github.com/spf13/cobra"
)

var (
	bundleNumPath     = "release/triggers/bundle-release/development/BUNDLE_NUMBER"
	cliMaxVersionPath = "release/triggers/bundle-release/development/CLI_MAX_VERSION"
	cliMinVersionPath = "release/triggers/bundle-release/development/CLI_MIN_VERSION"
	//triggerFilePath         = "release/triggers/eks-a-releaser-trigger"
	usersForkedRepoAccount = getAuthenticatedUsername()
)

// stageBundleCmd represents the stageBundle command
var stageBundleCmd = &cobra.Command{
	Use:   "stage-bundle",
	Short: "creates a PR containing 3 commits, each updating the contents of a singular file intended for staging bundle release",
	Long: `Retrieves updated content for development : bundle number, cli max version, and cli min version. 
	Writes the updated changes to the 3 files and raises a PR with the 3 commits.`,

	Run: func(cmd *cobra.Command, args []string) {

		err := runAllStagebundle()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func runAllStagebundle() error {

	RELEASE_TYPE := os.Getenv("RELEASE_TYPE")

	commitSHA, err := updateFilesStageBundle(RELEASE_TYPE)
	if err != nil {
		return err
	}
	fmt.Print(commitSHA)

	err = createPullRequestStageBundleTwo(RELEASE_TYPE)
	if err != nil {
		return err
	}

	return nil
}

func updateFilesStageBundle(releaseType string) (string, error) {

	// create client
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	// env variables
	bundleNumber := os.Getenv("RELEASE_NUMBER")
	latestVersion := os.Getenv("LATEST_VERSION")
	latestRelease := os.Getenv("LATEST_RELEASE")

	// Get the latest commit SHA from the appropriate branch, patch vs minor
	ref, _, err := client.Git.GetRef(ctx, usersForkedRepoAccount, EKSAnyrepoName, "heads/"+getBranchName(releaseType, latestRelease))
	if err != nil {
		return "", fmt.Errorf("error getting ref %s", err)
	}
	latestCommitSha := ref.Object.GetSHA()

	entries := []*github.TreeEntry{}
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(bundleNumPath, "/")), Type: github.String("blob"), Content: github.String(string(bundleNumber)), Mode: github.String("100644")})
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(cliMaxVersionPath, "/")), Type: github.String("blob"), Content: github.String(string(latestVersion)), Mode: github.String("100644")})
	entries = append(entries, &github.TreeEntry{Path: github.String(strings.TrimPrefix(cliMinVersionPath, "/")), Type: github.String("blob"), Content: github.String(string(latestVersion)), Mode: github.String("100644")})

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
		Message: github.String("Update version files for bundle release"),
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

func createPullRequestStageBundleTwo(releaseType string) error {
	latestRelease := os.Getenv("LATEST_RELEASE")

	// Create client
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	base := latestRelease // Target branch for upstream repo
	head := fmt.Sprintf("%s:%s", usersForkedRepoAccount, getBranchName(releaseType, latestRelease))
	title := "Update version files to stage bundle release"
	body := "This pull request is responsible for updating the contents of 3 separate files in order to trigger the staging bundle release pipeline"

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

func getBranchName(releaseType, latestRelease string) string {
	if releaseType == "minor" {
		return latestRelease
	}
	return latestRelease + "-releaser-patch"
}

// non related to staging bundle release
// User represents the user's GitHub account information.
type User struct {
	Login             string `json:"login"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
	Name              string `json:"name"`
	Company           string `json:"company"`
	Blog              string `json:"blog"`
	Location          string `json:"location"`
	Email             string `json:"email"`
	Hireable          bool   `json:"hireable"`
	Bio               string `json:"bio"`
	TwitterUsername   string `json:"twitter_username"`
	PublicRepos       int    `json:"public_repos"`
	PublicGists       int    `json:"public_gists"`
	Followers         int    `json:"followers"`
	Following         int    `json:"following"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

func getAuthenticatedUsername() string {

	// username is fetched using gh PAT
	accessToken := os.Getenv("SECRET_PAT")
	// github PAT is retrieved from secrets manager / buildspec file

	// Create a new HTTP client
	client := &http.Client{}

	// Create a new HTTP request
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return "error creating HTTP request"
	}

	// Set the authorization header with the personal access token
	req.Header.Set("Authorization", "token "+accessToken)

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		return "error sending HTTP request"
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "error reading response body"
	}

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return "failed to retrieve user information"
	}

	// Unmarshal the response body into a User struct
	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return "error unmarshalling"
	}

	stringUser := user.Login
	return stringUser
}
