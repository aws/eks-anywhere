package cmd

/*
	what does this command do?

	(1) creates a folder on user's Desktop
	(2) clones github prow repo (update account name)
	(3) Renames templater file and updates contents
	(4) executes make command
	(5) creates a branch on user's fork of prow repo
	(6) stages, commits, and pushes changes to newly created branch
	(7) creates PR targeting upstream "main" branch

	Only command to include local cloning of repo
*/

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aws/eks-anywhere-build-tooling/tools/version-tracker/pkg/util/command"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v62/github"
	"github.com/spf13/cobra"
)

var (
	prowRepo = "eks-anywhere-prow-jobs"
)

// upProwCmd represents the upProw command
var updateProwCmd = &cobra.Command{
	Use:   "update-prow",
	Short: "accesses prow-jobs repo and updates version files",
	Long:  `A`,
	Run: func(cmd *cobra.Command, args []string) {
		updateProw()
	},
}

func updateProw() {

	// Step 1: Create a folder on the user's desktop
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("error getting user home directory: %v", err)
		return
	}
	desktopPath := filepath.Join(homeDir, "Desktop")
	newFolderPath := filepath.Join(desktopPath, "ProwJobsRepo")
	err = os.Mkdir(newFolderPath, 0755)
	if err != nil {
		log.Println("Error creating folder:", err)
		return
	}
	fmt.Println("Folder created successfully at:", newFolderPath)

	//clones github repo into newly created folder
	clonedRepoDestination := filepath.Join(homeDir, "Desktop", "ProwJobsRepo")
	repo, err := cloneRepo("https://github.com/ibix16/eks-anywhere-prow-jobs", clonedRepoDestination)
	if err != nil {
		log.Printf("error cloning repo: %v", err)
		return
	}

	// Step 2: Rename the file with the latest version
	originalFilePath, err := retrieveFilePath(clonedRepoDestination + "/templater/jobs/periodic/eks-anywhere-build-tooling")
	if err != nil {
		log.Printf("error fetching path to file on cloned repo: %v", err)
	}
	newFilePath := clonedRepoDestination + "/templater/jobs/periodic/eks-anywhere-build-tooling/eks-anywhere-attribution-periodics-" + latestRelease + ".yaml"
	err = os.Rename(originalFilePath, newFilePath)
	if err != nil {
		log.Printf("error renaming file: %v", err)
		return
	}

	// Step 3: Update file contents
	convertedRelease := strings.Replace(latestRelease, ".", "-", 1)
	content, err := ioutil.ReadFile(newFilePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	releasePattern := regexp.MustCompile(`release-0\.\d+\d+`)
	jobNamePattern := regexp.MustCompile(`release-0-\d+\d+`)
	updatedContent := releasePattern.ReplaceAllString(string(content), latestRelease)
	updatedContent = jobNamePattern.ReplaceAllString(updatedContent, convertedRelease)
	err = ioutil.WriteFile(newFilePath, []byte(updatedContent), 0644)
	if err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
	log.Println("File updated successfully.")

	// Execute make command
	err = makeCommand()
	if err != nil {
		log.Printf("error running make command: %v", err)
		return
	}
	fmt.Println("Make command executed successfully.")

	// Create a branch in the user's forked repo
	err = createProwBranch(usersForkedRepoAccount, prowRepo)
	if err != nil {
		log.Printf("error creating branch: %v", err)
		return
	}

	// Commit and push changes to the branch
	err = commitAndPushChanges(repo, latestRelease+"-releaser")
	if err != nil {
		log.Printf("error pushing changes to branch: %v", err)
		return
	}
	fmt.Println("Changes pushed successfully.")

	// Create PR
	err = createProwPr()
	if err != nil {
		log.Printf("error creating PR: %v", err)
		return
	}

	// delete folder
	err = os.RemoveAll(clonedRepoDestination)
	if err != nil {
		log.Printf("error deleting folder: %s", err)
	}

	log.Println("Folder deleted successfully from desktop.")

}

// return full system file path to templater file on cloned repo
func retrieveFilePath(directory string) (string, error) {
	var filePath string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // Return the error immediately
		}
		if !info.IsDir() && strings.Contains(info.Name(), "release-0.") {
			filePath = path
			return nil // Stop walking after finding the first matching file
		}
		return nil
	})
	if err != nil {
		return "", err // Return the error if one occurred during the walk
	}
	if filePath == "" {
		return "", fmt.Errorf("no file found with 'release-0.' in its name")
	}
	return filePath, nil
}

// clones prow jobs repo on local machine destination
func cloneRepo(cloneURL, destination string) (*git.Repository, error) {
	repo, err := git.PlainClone(destination, false, &git.CloneOptions{
		URL:      cloneURL,
		Progress: os.Stdout,
	})
	if err != nil {
		if err == git.ErrRepositoryAlreadyExists {
			fmt.Printf("Repo already exists at %s\n", destination)
			repo, err = git.PlainOpen(destination)
			if err != nil {
				return nil, fmt.Errorf("opening repo from %s directory: %v", destination, err)
			}
		} else {
			return nil, fmt.Errorf("cloning repo %s to %s directory: %v", cloneURL, destination, err)
		}
	}
	return repo, nil
}

// function to execute make command located within templater dir
func makeCommand() error {
	desktopPath, err := os.UserHomeDir()
	if err != nil {
		log.Printf("error getting user home directory: %v", err)
		return nil
	}
	clonedRepoPath := filepath.Join(desktopPath, "Desktop", "ProwJobsRepo")
	templaterDirPath := filepath.Join(clonedRepoPath, "templater")
	updateProwJobsCommandSequence := fmt.Sprintf("make prowjobs -C %s", templaterDirPath)
	updateProwJobsCmd := exec.Command("bash", "-c", updateProwJobsCommandSequence)
	_, err = command.ExecCommand(updateProwJobsCmd)
	if err != nil {
		return fmt.Errorf("running make prowjobs command: %v", err)
	}
	return nil
}

func createProwBranch(owner, repo string) error {

	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	// Get the latest commit on the "main" branch
	mainBranch, _, err := client.Git.GetRef(ctx, owner, repo, "refs/heads/main")
	if err != nil {
		return fmt.Errorf("failed to get 'main' branch: %v", err)
	}
	latestCommit, _, err := client.Git.GetCommit(ctx, owner, repo, *mainBranch.Object.SHA)
	if err != nil {
		return fmt.Errorf("failed to get latest commit: %v", err)
	}
	// Create a new reference for the branch
	newRef := &github.Reference{
		Ref: github.String("refs/heads/" + latestRelease + "-releaser"),
		Object: &github.GitObject{
			SHA: latestCommit.SHA,
		},
	}
	// Create the new branch
	_, _, err = client.Git.CreateRef(ctx, owner, repo, newRef)
	if err != nil {
		return fmt.Errorf("failed to create 'releaser' branch: %v", err)
	}
	fmt.Println("Branch created successfully")
	return nil
}

// Commits and pushes the changes to the new branch
func commitAndPushChanges(repo *git.Repository, branchName string) error {

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("could not get worktree: %v", err)
	}
	// Stage all changes
	err = w.AddGlob(".")
	if err != nil {
		return fmt.Errorf("could not stage changes: %v", err)
	}
	// Commit changes
	_, err = w.Commit("Update prow jobs for "+branchName, &git.CommitOptions{
		Author: &object.Signature{
			Name: "your-github-username", // Update with your GitHub username
			When: time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("could not commit changes: %v", err)
	}
	// Push changes to  the new branch
	accessToken := os.Getenv("SECRET_PAT")
	err = repo.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth: &http.BasicAuth{
			Username: "your-github-username", // Update with your GitHub username
			Password: accessToken,            // GitHub personal access token
		},
		RefSpecs: []config.RefSpec{
			config.RefSpec("refs/heads/main" + ":refs/heads/" + branchName),
		},
	})
	if err != nil {
		return fmt.Errorf("could not push changes: %v", err)
	}
	return nil
}

// Function to create the PR
func createProwPr() error {
	
	// Create a GitHub client
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	// Create a PR from this branch
	branchName := latestRelease + "-releaser"
	base := "main"                                                   // target branch in upstream repo
	head := fmt.Sprintf("%s:%s", usersForkedRepoAccount, branchName) // PR originates from
	title := "Update Prow Jobs Templater file & execute make command"
	body := "This pull request contains the most recent commit from the release branch."
	newPR := &github.NewPullRequest{
		Title: &title,
		Head:  &head,
		Base:  &base,
		Body:  &body,
	}
	pr, _, err := client.PullRequests.Create(ctx, upStreamRepoOwner, prowRepo, newPR)
	if err != nil {
		return fmt.Errorf("error creating PR: %s", err)
	}
	log.Printf("Pull request created: %s\n", pr.GetHTMLURL())
	return nil
}
