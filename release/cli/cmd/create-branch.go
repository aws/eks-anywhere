package cmd

/*
	what does this command do?

	if release type is "minor" then :
	creates a new release branch in upstream eks-a repo based off "main" & build tooling repo

	creates a new release branch in forked repo based off newly created release branch in upstream repo

	else :
	creates a new patch branch in users forked repo based off latest release branch upstream

*/

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v62/github"
	"github.com/spf13/cobra"
)

var (
	buildToolingRepoName = "eks-anywhere-build-tooling"
	upStreamRepoOwner    = "testerIbix" // will eventually be replaced by actual upstream owner, aws
	latestRelease        = os.Getenv("LATEST_RELEASE")
	RELEASE_TYPE         = os.Getenv("RELEASE_TYPE")
	latestVersion        = os.Getenv("LATEST_VERSION")
	releaseNumber        = os.Getenv("RELEASE_NUMBER")
	accessToken          = os.Getenv("SECRET_PAT")
)

// createBranchCmd represents the createBranch command
var createBranchCmd = &cobra.Command{
	Use:   "create-branch",
	Short: "Creates new release branch from updated trigger file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,

	Run: func(cmd *cobra.Command, args []string) {

		err := releaseDecision()
		if err != nil {
			log.Printf("error creating branch %s", err)
		}
	},
}

func releaseDecision() error {

	if RELEASE_TYPE == "minor" {
		err := createMinorBranches()
		if err != nil {
			log.Printf("error calling createMinorBranches %s", err)
			return err
		}
		return nil
	}
	// else
	err := createPatchBranch()
	if err != nil {
		log.Printf("error calling createPatchBranch %s", err)
		return err
	}
	return nil
}

func createMinorBranches() error {

	//create client
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	// create branch in upstream repo based off main branch
	ref := "refs/heads/" + latestRelease
	baseRef := "main"

	// Get the reference for the base branch
	baseRefObj, _, err := client.Git.GetRef(ctx, upStreamRepoOwner, EKSAnyrepoName, "heads/"+baseRef)
	if err != nil {
		return fmt.Errorf("error getting base branch reference one: %v", err)
	}

	// Create a new branch
	newBranchRef, _, err := client.Git.CreateRef(ctx, upStreamRepoOwner, EKSAnyrepoName, &github.Reference{
		Ref: &ref,
		Object: &github.GitObject{
			SHA: baseRefObj.Object.SHA,
		},
	})
	if err != nil {
		return fmt.Errorf("error creating branch one: %v", err)
	}

	// branch created upstream
	log.Printf("New release branch '%s' created upstream successfully\n", *newBranchRef.Ref)

	// create branch in forked repo based off upstream
	ref = "refs/heads/" + latestRelease
	baseRef = latestRelease

	// Get the reference for the base branch from the upstream repository
	baseRefObj, _, err = client.Git.GetRef(ctx, upStreamRepoOwner, EKSAnyrepoName, "heads/"+baseRef)
	if err != nil {
		return fmt.Errorf("error getting base branch reference two: %v", err)
	}

	// Create a new branch
	newBranchRef, _, err = client.Git.CreateRef(ctx, usersForkedRepoAccount, EKSAnyrepoName, &github.Reference{
		Ref: &ref,
		Object: &github.GitObject{
			SHA: baseRefObj.Object.SHA,
		},
	})
	if err != nil {
		return fmt.Errorf("error creating branch two: %v", err)
	}

	// branch created upstream
	log.Printf("New user fork branch '%s' created successfully\n", *newBranchRef.Ref)

	// create branch in upstream build tooling repo based off main branch
	ref = "refs/heads/" + latestRelease
	baseRef = "main"

	// Get the reference for the base branch
	baseRefObj, _, err = client.Git.GetRef(ctx, upStreamRepoOwner, buildToolingRepoName, "heads/"+baseRef)
	if err != nil {
		return fmt.Errorf("error getting base branch reference three: %v", err)
	}

	// Create a new branch
	newBranchRef, _, err = client.Git.CreateRef(ctx, upStreamRepoOwner, buildToolingRepoName, &github.Reference{
		Ref: &ref,
		Object: &github.GitObject{
			SHA: baseRefObj.Object.SHA,
		},
	})
	if err != nil {
		return fmt.Errorf("error creating branch three: %v", err)
	}

	// branch created upstream
	log.Printf("New build tooling branch '%s' created successfully\n", *newBranchRef.Ref)

	return nil
}

func createPatchBranch() error {

	//create client
	accessToken := os.Getenv("SECRET_PAT")
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(accessToken)

	// create branch in forked repo based off upstream
	ref := "refs/heads/" + latestRelease + "-releaser-patch"
	baseRef := latestRelease

	// Get the reference for the base branch from upstream
	baseRefObj, _, err := client.Git.GetRef(ctx, upStreamRepoOwner, EKSAnyrepoName, "heads/"+baseRef)
	if err != nil {
		return fmt.Errorf("error getting base branch reference: %v", err)
	}

	// Create a new branch in fork
	newBranchRef, _, err := client.Git.CreateRef(ctx, usersForkedRepoAccount, EKSAnyrepoName, &github.Reference{
		Ref: &ref,
		Object: &github.GitObject{
			SHA: baseRefObj.Object.SHA,
		},
	})
	if err != nil {
		return fmt.Errorf("error creating branch: %v", err)
	}

	// branch created upstream
	log.Printf("New branch '%s' created successfully\n", *newBranchRef.Ref)

	return nil
}
